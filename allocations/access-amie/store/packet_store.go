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

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/jmoiron/sqlx"
)

type PacketStore interface {
	FindByAmieID(ctx context.Context, amieID int64) (*model.Packet, error)
	FindByID(ctx context.Context, id string) (*model.Packet, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Packet) error
	Update(ctx context.Context, tx *sql.Tx, p *model.Packet) error
}

type mariaDBPacketStore struct {
	db *sqlx.DB
}

func NewPacketStore(db *sqlx.DB) PacketStore {
	return &mariaDBPacketStore{db: db}
}

func (s *mariaDBPacketStore) FindByAmieID(ctx context.Context, amieID int64) (*model.Packet, error) {
	var p model.Packet
	err := s.db.GetContext(ctx, &p,
		`SELECT id, amie_id, type, status, raw_json, received_at, decoded_at, processed_at, retries, last_error
		 FROM amie_packets WHERE amie_id = ?`, amieID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mariaDBPacketStore) FindByID(ctx context.Context, id string) (*model.Packet, error) {
	var p model.Packet
	err := s.db.GetContext(ctx, &p,
		`SELECT id, amie_id, type, status, raw_json, received_at, decoded_at, processed_at, retries, last_error
		 FROM amie_packets WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mariaDBPacketStore) Save(ctx context.Context, tx *sql.Tx, p *model.Packet) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO amie_packets (id, amie_id, type, status, raw_json, received_at, decoded_at, processed_at, retries, last_error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.AmieID, p.Type, p.Status, p.RawJSON,
		p.ReceivedAt, p.DecodedAt, p.ProcessedAt,
		p.Retries, p.LastError)
	return err
}

func (s *mariaDBPacketStore) Update(ctx context.Context, tx *sql.Tx, p *model.Packet) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE amie_packets SET type = ?, status = ?, raw_json = ?, decoded_at = ?, processed_at = ?, retries = ?, last_error = ?
		 WHERE id = ?`,
		p.Type, p.Status, p.RawJSON, p.DecodedAt, p.ProcessedAt,
		p.Retries, p.LastError, p.ID)
	return err
}
