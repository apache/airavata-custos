// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// TreeRowLimit caps tree size; the handler marks oversize trees truncated.
const TreeRowLimit = 500

type TraceFilter struct {
	Sources  []string
	Statuses []string
	From     time.Time
	To       time.Time
	Q        string
	Limit    int
	Offset   int
}

type AuditTraceStore interface {
	ListTraces(ctx context.Context, filter TraceFilter) ([]models.TraceSummary, int, error)
	GetTraceTree(ctx context.Context, traceID []byte) (*models.TraceNode, bool, error)
	ListEvents(ctx context.Context, traceID, spanID []byte) ([]models.TraceEvent, error)
	ListSources(ctx context.Context) ([]string, error)
}

// unionAuditSelect normalises audit_events and amie_audit_log to one shape.
const unionAuditSelect = `
SELECT
    trace_id,
    span_id,
    parent_span_id,
    source       AS source,
    event_type   AS event_type,
    ''           AS entity_type,
    entity_id    AS entity_id,
    details      AS description,
    event_time   AS created_at
FROM audit_events
WHERE trace_id IS NOT NULL
UNION ALL
SELECT
    trace_id,
    span_id,
    parent_span_id,
    'amie'                 AS source,
    action                 AS event_type,
    entity_type            AS entity_type,
    COALESCE(entity_id,'') AS entity_id,
    COALESCE(summary,'')   AS description,
    created_at             AS created_at
FROM amie_audit_log
WHERE trace_id IS NOT NULL
`

type mysqlAuditTraceStore struct {
	db *sqlx.DB
}

func NewAuditTraceStore(db *sqlx.DB) AuditTraceStore {
	return &mysqlAuditTraceStore{db: db}
}

type rowEvent struct {
	TraceID      []byte    `db:"trace_id"`
	SpanID       []byte    `db:"span_id"`
	ParentSpanID []byte    `db:"parent_span_id"`
	Source       string    `db:"source"`
	EventType    string    `db:"event_type"`
	EntityType   string    `db:"entity_type"`
	EntityID     string    `db:"entity_id"`
	Description  string    `db:"description"`
	CreatedAt    time.Time `db:"created_at"`
}

func (r rowEvent) toTraceEvent() models.TraceEvent {
	return models.TraceEvent{
		SpanID:       r.SpanID,
		ParentSpanID: r.ParentSpanID,
		Source:       r.Source,
		EventType:    r.EventType,
		EntityType:   r.EntityType,
		EntityID:     r.EntityID,
		Description:  r.Description,
		Status:       tracing.EventStatus(r.EventType),
		CreatedAt:    r.CreatedAt,
	}
}

// ListTraces aggregates by trace_id then runs one summary probe per row.
// Acceptable while page sizes stay <= 200.
func (s *mysqlAuditTraceStore) ListTraces(ctx context.Context, f TraceFilter) ([]models.TraceSummary, int, error) {
	whereSQL, args := buildTraceWhere(f)

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	countQuery := `SELECT COUNT(*) FROM (
		SELECT trace_id FROM (` + unionAuditSelect + `) u ` + whereSQL + `
		GROUP BY trace_id
	) g`
	var total int
	if err := s.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("audit_trace_store: count: %w", err)
	}

	listQuery := `SELECT trace_id,
	        MIN(created_at) AS started_at,
	        MAX(created_at) AS ended_at,
	        COUNT(*)        AS event_count
	  FROM (` + unionAuditSelect + `) u ` + whereSQL + `
	  GROUP BY trace_id
	  ORDER BY MAX(created_at) DESC
	  LIMIT ? OFFSET ?`
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, f.Offset)

	type listRow struct {
		TraceID    []byte    `db:"trace_id"`
		StartedAt  time.Time `db:"started_at"`
		EndedAt    time.Time `db:"ended_at"`
		EventCount int       `db:"event_count"`
	}
	var raw []listRow
	if err := s.db.SelectContext(ctx, &raw, listQuery, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("audit_trace_store: list: %w", err)
	}

	out := make([]models.TraceSummary, 0, len(raw))
	for _, r := range raw {
		rootOp, src, status, err := s.summariseTrace(ctx, r.TraceID)
		if err != nil {
			return nil, 0, err
		}
		if !matchesStatusFilter(status, f.Statuses) {
			continue
		}
		out = append(out, models.TraceSummary{
			TraceID:       r.TraceID,
			RootOperation: rootOp,
			Source:        src,
			Status:        status,
			StartedAt:     r.StartedAt,
			EndedAt:       r.EndedAt,
			EventCount:    r.EventCount,
		})
	}
	return out, total, nil
}

func (s *mysqlAuditTraceStore) summariseTrace(ctx context.Context, traceID []byte) (string, string, string, error) {
	q := `SELECT trace_id, span_id, parent_span_id, source, event_type, entity_type, entity_id, description, created_at
	  FROM (` + unionAuditSelect + `) u
	  WHERE trace_id = ?
	  ORDER BY created_at ASC, span_id ASC`
	var rows []rowEvent
	if err := s.db.SelectContext(ctx, &rows, q, traceID); err != nil {
		return "", "", "", fmt.Errorf("audit_trace_store: summarise: %w", err)
	}
	if len(rows) == 0 {
		return "", "", tracing.StatusInProgress, nil
	}

	var rootOp string
	for _, r := range rows {
		if len(r.ParentSpanID) == 0 {
			rootOp = r.EventType
			break
		}
	}
	if rootOp == "" {
		rootOp = rows[0].EventType
	}

	statuses := make([]tracing.TraceEventStatus, 0, len(rows))
	for _, r := range rows {
		statuses = append(statuses, tracing.TraceEventStatus{Source: r.Source, EventType: r.EventType})
	}
	src := dominantSource(rows)
	return rootOp, src, tracing.TraceStatus(statuses), nil
}

func dominantSource(rows []rowEvent) string {
	for _, r := range rows {
		if len(r.ParentSpanID) == 0 {
			return r.Source
		}
	}
	return rows[0].Source
}

func (s *mysqlAuditTraceStore) GetTraceTree(ctx context.Context, traceID []byte) (*models.TraceNode, bool, error) {
	q := `SELECT trace_id, span_id, parent_span_id, source, event_type, entity_type, entity_id, description, created_at
	  FROM (` + unionAuditSelect + `) u
	  WHERE trace_id = ?
	  ORDER BY created_at ASC, span_id ASC
	  LIMIT ?`
	var rows []rowEvent
	if err := s.db.SelectContext(ctx, &rows, q, traceID, TreeRowLimit+1); err != nil {
		return nil, false, fmt.Errorf("audit_trace_store: tree: %w", err)
	}
	if len(rows) == 0 {
		return nil, false, nil
	}
	truncated := len(rows) > TreeRowLimit
	if truncated {
		rows = rows[:TreeRowLimit]
	}

	return buildTree(rows), truncated, nil
}

// buildTree links rows by span_id. Rows whose parent isn't in the table
// become top-level siblings.
func buildTree(rows []rowEvent) *models.TraceNode {
	bySpan := make(map[string]*models.TraceNode, len(rows))
	for _, r := range rows {
		key := string(r.SpanID)
		bySpan[key] = &models.TraceNode{TraceEvent: r.toTraceEvent()}
	}
	root := &models.TraceNode{}
	for _, r := range rows {
		node := bySpan[string(r.SpanID)]
		if len(r.ParentSpanID) > 0 {
			if parent, ok := bySpan[string(r.ParentSpanID)]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		root.Children = append(root.Children, node)
	}
	return root
}

func (s *mysqlAuditTraceStore) ListEvents(ctx context.Context, traceID, spanID []byte) ([]models.TraceEvent, error) {
	q := `SELECT trace_id, span_id, parent_span_id, source, event_type, entity_type, entity_id, description, created_at
	  FROM (` + unionAuditSelect + `) u
	  WHERE trace_id = ?`
	args := []any{traceID}
	if len(spanID) > 0 {
		q += ` AND span_id = ?`
		args = append(args, spanID)
	}
	q += ` ORDER BY created_at ASC, span_id ASC`

	var rows []rowEvent
	if err := s.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("audit_trace_store: events: %w", err)
	}
	out := make([]models.TraceEvent, len(rows))
	for i, r := range rows {
		out[i] = r.toTraceEvent()
	}
	return out, nil
}

func (s *mysqlAuditTraceStore) ListSources(_ context.Context) ([]string, error) {
	return []string{"amie", "comanage", "core", "slurm"}, nil
}

func buildTraceWhere(f TraceFilter) (string, []any) {
	var clauses []string
	var args []any

	if len(f.Sources) > 0 {
		placeholders := make([]string, len(f.Sources))
		for i, src := range f.Sources {
			placeholders[i] = "?"
			args = append(args, src)
		}
		clauses = append(clauses, "u.source IN ("+strings.Join(placeholders, ",")+")")
	}
	if !f.From.IsZero() {
		clauses = append(clauses, "u.created_at >= ?")
		args = append(args, f.From)
	}
	if !f.To.IsZero() {
		clauses = append(clauses, "u.created_at <= ?")
		args = append(args, f.To)
	}
	if f.Q != "" {
		clauses = append(clauses, "(HEX(u.trace_id) LIKE ? OR u.event_type LIKE ?)")
		args = append(args, strings.ToUpper(f.Q)+"%", "%"+f.Q+"%")
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func matchesStatusFilter(status string, want []string) bool {
	if len(want) == 0 {
		return true
	}
	for _, w := range want {
		if w == status {
			return true
		}
	}
	return false
}
