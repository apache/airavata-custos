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

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/apache/airavata-custos/signer/internal/store"
)

var (
	ErrForbidden            = errors.New("certificate revocation is forbidden")
	ErrCertificateNotActive = errors.New("certificate is not active")
)

type Transactor interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

type RevokeCommand struct {
	SerialNumber  int64
	Reason        string
	RevokedBy     string
	OwnerEmail    string
	CanRevokeAny  bool
	RequireActive bool
}

type RevokedCertificate struct {
	SerialNumber   int64
	Reason         string
	RevokedAt      time.Time
	AlreadyRevoked bool
}

type RevocationService struct {
	db  Transactor
	now func() time.Time
}

func NewRevocationService(db Transactor) *RevocationService {
	return &RevocationService{db: db, now: func() time.Time { return time.Now().UTC() }}
}

func (s *RevocationService) Revoke(ctx context.Context, command RevokeCommand) (*RevokedCertificate, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	certificate, err := store.GetCertificateForRevocation(ctx, tx, command.SerialNumber)
	if err != nil {
		return nil, err
	}
	if command.OwnerEmail != "" && certificate.UserEmail != command.OwnerEmail && !command.CanRevokeAny {
		return nil, ErrForbidden
	}

	existing, err := store.GetLatestRevocation(ctx, tx, command.SerialNumber)
	switch {
	case err == nil:
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("committing transaction: %w", err)
		}
		return &RevokedCertificate{
			SerialNumber: command.SerialNumber, Reason: existing.Reason,
			RevokedAt: existing.RevokedAt, AlreadyRevoked: true,
		}, nil
	case errors.Is(err, sql.ErrNoRows):
		// Continue with first-time revocation.
	default:
		return nil, err
	}

	now := s.now().UTC()
	if command.RequireActive && (now.Before(certificate.ValidAfter) || !now.Before(certificate.ValidBefore)) {
		return nil, ErrCertificateNotActive
	}

	serial := command.SerialNumber
	keyID := certificate.KeyID
	caFingerprint := certificate.CAFingerprint
	event := &store.RevocationEvent{
		TenantID: certificate.TenantID, ClientID: certificate.ClientID,
		SerialNumber: &serial, KeyID: &keyID, CAFingerprint: &caFingerprint,
		Reason: command.Reason, RevokedBy: command.RevokedBy,
	}
	if err := store.InsertRevocationEventAt(ctx, tx, event, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}
	return &RevokedCertificate{
		SerialNumber: command.SerialNumber, Reason: command.Reason,
		RevokedAt: now, AlreadyRevoked: false,
	}, nil
}
