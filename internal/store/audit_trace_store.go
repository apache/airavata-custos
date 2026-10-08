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
	"slices"
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

	traceIDs []string // set by ListTraces once a status filter has picked the traces
}

type AuditTraceStore interface {
	ListTraces(ctx context.Context, filter TraceFilter) ([]models.TraceSummary, int, error)
	GetTraceTree(ctx context.Context, traceID string) (*models.TraceNode, bool, error)
	ListEvents(ctx context.Context, traceID, spanID string) ([]models.TraceEvent, error)
	ListSources(ctx context.Context) ([]string, error)
}

// auditTraceSelect projects audit_events into the trace-view shape. Every
// subsystem (core and each connector) writes its audit rows here, tagged by
// `source`. Connector-specific references (e.g. AMIE's packet_id / event_id
// on amie_audit_extras) live on connector-owned tables and are joined
// only by the connector-specific endpoints.
const auditTraceSelect = `
SELECT
    trace_id,
    span_id,
    parent_span_id,
    source,
    event_type,
    entity_type,
    entity_id,
    details      AS description,
    event_time   AS created_at
FROM audit_events
`

type pgAuditTraceStore struct {
	db *sqlx.DB
}

func NewAuditTraceStore(db *sqlx.DB) AuditTraceStore {
	return &pgAuditTraceStore{db: db}
}

type rowEvent struct {
	TraceID      string    `db:"trace_id"`
	SpanID       string    `db:"span_id"`
	ParentSpanID *string   `db:"parent_span_id"`
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

// ListTraces pages traces in SQL and summarises the page in one more query.
// Status is computed in Go, so a status filter first loads every matching
// trace's rows in one query and narrows the trace set before counting and paging.
func (s *pgAuditTraceStore) ListTraces(ctx context.Context, f TraceFilter) ([]models.TraceSummary, int, error) {
	if len(f.Statuses) > 0 {
		whereSQL, args := buildTraceWhere(f)
		byTrace, err := s.traceRows(ctx, `trace_id IN (SELECT trace_id FROM (`+auditTraceSelect+`) u `+whereSQL+`)`, args...)
		if err != nil {
			return nil, 0, err
		}
		f.traceIDs = []string{}
		for id, rows := range byTrace {
			if _, _, status := summarise(rows); slices.Contains(f.Statuses, status) {
				f.traceIDs = append(f.traceIDs, id)
			}
		}
	}
	whereSQL, args := buildTraceWhere(f)

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	countQuery := `SELECT COUNT(*) FROM (
		SELECT trace_id FROM (` + auditTraceSelect + `) u ` + whereSQL + `
		GROUP BY trace_id
	) g`
	var total int
	if err := s.db.GetContext(ctx, &total, s.db.Rebind(countQuery), args...); err != nil {
		return nil, 0, fmt.Errorf("audit_trace_store: count: %w", err)
	}

	listQuery := `SELECT trace_id,
	        MIN(created_at) AS started_at,
	        MAX(created_at) AS ended_at,
	        COUNT(*)        AS event_count
	  FROM (` + auditTraceSelect + `) u ` + whereSQL + `
	  GROUP BY trace_id
	  ORDER BY MAX(created_at) DESC
	  LIMIT ? OFFSET ?`
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, limit, f.Offset)

	type listRow struct {
		TraceID    string    `db:"trace_id"`
		StartedAt  time.Time `db:"started_at"`
		EndedAt    time.Time `db:"ended_at"`
		EventCount int       `db:"event_count"`
	}
	var raw []listRow
	if err := s.db.SelectContext(ctx, &raw, s.db.Rebind(listQuery), listArgs...); err != nil {
		return nil, 0, fmt.Errorf("audit_trace_store: list: %w", err)
	}

	ids := make([]string, len(raw))
	for i, r := range raw {
		ids[i] = r.TraceID
	}
	byTrace, err := s.traceRows(ctx, `trace_id = ANY(?)`, ids)
	if err != nil {
		return nil, 0, err
	}
	out := make([]models.TraceSummary, 0, len(raw))
	for _, r := range raw {
		rootOp, src, status := summarise(byTrace[r.TraceID])
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

// traceRows loads the rows of every trace matching where, oldest first, keyed by trace.
func (s *pgAuditTraceStore) traceRows(ctx context.Context, where string, args ...any) (map[string][]rowEvent, error) {
	q := `SELECT trace_id, span_id, parent_span_id, source, event_type, entity_type, entity_id, description, created_at
	  FROM (` + auditTraceSelect + `) u
	  WHERE ` + where + `
	  ORDER BY created_at ASC, span_id ASC`
	var rows []rowEvent
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(q), args...); err != nil {
		return nil, fmt.Errorf("audit_trace_store: rows: %w", err)
	}
	byTrace := map[string][]rowEvent{}
	for _, r := range rows {
		byTrace[r.TraceID] = append(byTrace[r.TraceID], r)
	}
	return byTrace, nil
}

// summarise names a trace by its root row, the first whose parent span wrote no row.
func summarise(rows []rowEvent) (rootOp, source, status string) {
	spans := make(map[string]bool, len(rows))
	statuses := make([]tracing.TraceEventStatus, len(rows))
	for i, r := range rows {
		spans[r.SpanID] = true
		statuses[i] = tracing.TraceEventStatus{Source: r.Source, EventType: r.EventType}
	}
	root := rows[0]
	for _, r := range rows {
		if r.ParentSpanID == nil || !spans[*r.ParentSpanID] {
			root = r
			break
		}
	}
	return root.EventType, root.Source, tracing.TraceStatus(statuses)
}

func (s *pgAuditTraceStore) GetTraceTree(ctx context.Context, traceID string) (*models.TraceNode, bool, error) {
	byTrace, err := s.traceRows(ctx, `trace_id = ?`, traceID)
	rows := byTrace[traceID]
	if err != nil || len(rows) == 0 {
		return nil, false, err
	}
	_, _, status := summarise(rows)
	truncated := len(rows) > TreeRowLimit
	tree := buildTree(rows[:min(len(rows), TreeRowLimit)])
	tree.Status = status
	return tree, truncated, nil
}

// buildTree makes one node per row and hangs children off their parent span's
// first row. Rows whose parent isn't in the table become top-level siblings.
func buildTree(rows []rowEvent) *models.TraceNode {
	nodes := make([]*models.TraceNode, len(rows))
	bySpan := make(map[string]*models.TraceNode, len(rows))
	for i, r := range rows {
		nodes[i] = &models.TraceNode{TraceEvent: r.toTraceEvent()}
		if _, ok := bySpan[r.SpanID]; !ok {
			bySpan[r.SpanID] = nodes[i]
		}
	}
	root := &models.TraceNode{}
	for i, r := range rows {
		node := nodes[i]
		if r.ParentSpanID != nil {
			if parent, ok := bySpan[*r.ParentSpanID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		root.Children = append(root.Children, node)
	}
	return root
}

func (s *pgAuditTraceStore) ListEvents(ctx context.Context, traceID, spanID string) ([]models.TraceEvent, error) {
	q := `SELECT trace_id, span_id, parent_span_id, source, event_type, entity_type, entity_id, description, created_at
	  FROM (` + auditTraceSelect + `) u
	  WHERE trace_id = ?`
	args := []any{traceID}
	if spanID != "" {
		q += ` AND span_id = ?`
		args = append(args, spanID)
	}
	q += ` ORDER BY created_at ASC, span_id ASC`

	var rows []rowEvent
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(q), args...); err != nil {
		return nil, fmt.Errorf("audit_trace_store: events: %w", err)
	}
	out := make([]models.TraceEvent, len(rows))
	for i, r := range rows {
		out[i] = r.toTraceEvent()
	}
	return out, nil
}

func (s *pgAuditTraceStore) ListSources(_ context.Context) ([]string, error) {
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
		clauses = append(clauses, "(u.trace_id ILIKE ? OR u.event_type ILIKE ?)")
		args = append(args, strings.ToLower(f.Q)+"%", "%"+f.Q+"%")
	}

	if f.traceIDs != nil {
		clauses = append(clauses, "u.trace_id = ANY(?)")
		args = append(args, f.traceIDs)
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}
