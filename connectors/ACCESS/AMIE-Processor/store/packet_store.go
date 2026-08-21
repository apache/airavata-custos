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
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/jmoiron/sqlx"
)

// PacketListFilter narrows the packet listing returned by ListPackets.
type PacketListFilter struct {
	Status string
	Type   string
	Query  string
	From   time.Time
	To     time.Time
	Limit  int
	Offset int
}

// StatBucket is one row from the per-day (status, type) packet aggregation.
type StatBucket struct {
	Date   string `db:"date"`
	Status string `db:"status"`
	Type   string `db:"type"`
	Count  int64  `db:"count"`
}

// PacketStore is the data-access surface for the amie_packets table.
// FindByAmieID/FindByID/Save/Update serve the worker pipeline.
// ListPackets/ListPacketEvents/GetStats serve read endpoints.
type PacketStore interface {
	FindByAmieID(ctx context.Context, amieID int64) (*model.Packet, error)
	FindByID(ctx context.Context, id string) (*model.Packet, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Packet) error
	Update(ctx context.Context, tx *sql.Tx, p *model.Packet) error

	ListPackets(ctx context.Context, f PacketListFilter) ([]model.Packet, int, error)
	ListPacketEvents(ctx context.Context, packetID string) ([]model.ProcessingEvent, error)
	GetStats(ctx context.Context, window time.Duration) ([]StatBucket, error)
}

type pgPacketStore struct {
	db *sqlx.DB
}

func NewPacketStore(db *sqlx.DB) PacketStore {
	return &pgPacketStore{db: db}
}

const packetColumns = `id, amie_id, type, status, raw_json, received_at, decoded_at, processed_at, retries, last_error`

func (s *pgPacketStore) FindByAmieID(ctx context.Context, amieID int64) (*model.Packet, error) {
	var p model.Packet
	err := s.db.GetContext(ctx, &p,
		`SELECT `+packetColumns+` FROM amie_packets WHERE amie_id = $1`, amieID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *pgPacketStore) FindByID(ctx context.Context, id string) (*model.Packet, error) {
	var p model.Packet
	err := s.db.GetContext(ctx, &p,
		`SELECT `+packetColumns+` FROM amie_packets WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *pgPacketStore) Save(ctx context.Context, tx *sql.Tx, p *model.Packet) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO amie_packets (id, amie_id, type, status, raw_json, received_at, decoded_at, processed_at, retries, last_error)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.ID, p.AmieID, p.Type, p.Status, p.RawJSON,
		p.ReceivedAt, p.DecodedAt, p.ProcessedAt,
		p.Retries, p.LastError)
	return err
}

func (s *pgPacketStore) Update(ctx context.Context, tx *sql.Tx, p *model.Packet) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE amie_packets SET type = $1, status = $2, raw_json = $3, decoded_at = $4, processed_at = $5, retries = $6, last_error = $7
		 WHERE id = $8`,
		p.Type, p.Status, p.RawJSON, p.DecodedAt, p.ProcessedAt,
		p.Retries, p.LastError, p.ID)
	return err
}

func (s *pgPacketStore) ListPackets(ctx context.Context, f PacketListFilter) ([]model.Packet, int, error) {
	where, args := buildPacketFilter(f)

	var total int
	if err := s.db.GetContext(ctx, &total, s.db.Rebind(`SELECT COUNT(*) FROM amie_packets`+where), args...); err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	q := fmt.Sprintf(`SELECT %s FROM amie_packets%s ORDER BY received_at DESC LIMIT %d OFFSET %d`,
		packetColumns, where, limit, offset)

	var rows []model.Packet
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(q), args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func buildPacketFilter(f PacketListFilter) (string, []any) {
	var clauses []string
	var args []any
	if f.Status != "" && !strings.EqualFold(f.Status, "all") {
		clauses = append(clauses, "status = ?")
		args = append(args, f.Status)
	}
	if f.Type != "" {
		clauses = append(clauses, "type = ?")
		args = append(args, f.Type)
	}
	if f.Query != "" {
		clauses = append(clauses, "(id ILIKE ? OR CAST(amie_id AS TEXT) ILIKE ?)")
		like := "%" + f.Query + "%"
		args = append(args, like, like)
	}
	if !f.From.IsZero() {
		clauses = append(clauses, "received_at >= ?")
		args = append(args, f.From)
	}
	if !f.To.IsZero() {
		clauses = append(clauses, "received_at <= ?")
		args = append(args, f.To)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (s *pgPacketStore) ListPacketEvents(ctx context.Context, packetID string) ([]model.ProcessingEvent, error) {
	var rows []model.ProcessingEvent
	err := s.db.SelectContext(ctx, &rows,
		`SELECT id, packet_id, type, status, attempts, created_at, started_at, finished_at, last_error, next_retry_at
		 FROM amie_processing_events WHERE packet_id = $1 ORDER BY created_at ASC`, packetID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgPacketStore) GetStats(ctx context.Context, window time.Duration) ([]StatBucket, error) {
	if window <= 0 {
		window = 30 * 24 * time.Hour
	}
	since := time.Now().UTC().Add(-window)
	var rows []StatBucket
	err := s.db.SelectContext(ctx, &rows,
		`SELECT TO_CHAR(received_at, 'YYYY-MM-DD') AS date, status, type, COUNT(*) AS count
		 FROM amie_packets
		 WHERE received_at >= $1
		 GROUP BY TO_CHAR(received_at, 'YYYY-MM-DD'), status, type
		 ORDER BY date ASC`, since)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
