// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package audit provides dual audit logging (database + structured JSON log).
package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/signer/internal/store"
)

// Logger writes audit entries to both the database and structured slog output.
type Logger struct {
	db      *store.DB
	slogger *slog.Logger
}

func NewLogger(db *store.DB, baseLogger *slog.Logger) *Logger {
	return &Logger{
		db:      db,
		slogger: baseLogger.With("logger", "audit"),
	}
}

type IssuanceEntry struct {
	TenantID             string
	ClientID             string
	SerialNumber         int64
	KeyID                string
	Principal            string
	UserEmail            string
	PublicKeyFingerprint string
	CAFingerprint        string
	ValidAfter           time.Time
	ValidBefore          time.Time
	SourceIP             string
	UserAccessTokenHash  string
	CorrelationID        string
}

func (l *Logger) LogIssuance(ctx context.Context, entry *IssuanceEntry) error {
	dbLog := &store.IssuanceLog{
		TenantID:             entry.TenantID,
		ClientID:             entry.ClientID,
		SerialNumber:         entry.SerialNumber,
		KeyID:                entry.KeyID,
		Principal:            entry.Principal,
		UserEmail:            entry.UserEmail,
		PublicKeyFingerprint: entry.PublicKeyFingerprint,
		CAFingerprint:        entry.CAFingerprint,
		ValidAfter:           entry.ValidAfter,
		ValidBefore:          entry.ValidBefore,
		SourceIP:             entry.SourceIP,
		UserAccessTokenHash:  entry.UserAccessTokenHash,
	}

	if err := l.db.InsertIssuanceLog(ctx, dbLog); err != nil {
		return err
	}

	l.slogger.InfoContext(ctx, "certificate issued",
		"event_type", "certificate_issuance",
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"tenant_id", entry.TenantID,
		"client_id", entry.ClientID,
		"serial_number", entry.SerialNumber,
		"key_id", entry.KeyID,
		"principal", entry.Principal,
		"user_email", entry.UserEmail,
		"public_key_fingerprint", entry.PublicKeyFingerprint,
		"ca_fingerprint", entry.CAFingerprint,
		"valid_after", entry.ValidAfter.Unix(),
		"valid_before", entry.ValidBefore.Unix(),
		"source_ip", entry.SourceIP,
		"user_token_hash", entry.UserAccessTokenHash,
		"correlation_id", entry.CorrelationID,
	)

	return nil
}

type RevocationEntry struct {
	TenantID      string
	ClientID      string
	SerialNumber  *int64
	KeyID         *string
	CAFingerprint *string
	Reason        string
	RevokedBy     string
	CorrelationID string
}

func (l *Logger) LogRevocation(ctx context.Context, entry *RevocationEntry) error {
	ev := &store.RevocationEvent{
		TenantID:      entry.TenantID,
		ClientID:      entry.ClientID,
		SerialNumber:  entry.SerialNumber,
		KeyID:         entry.KeyID,
		CAFingerprint: entry.CAFingerprint,
		Reason:        entry.Reason,
		RevokedBy:     entry.RevokedBy,
	}

	if err := l.db.InsertRevocationEvent(ctx, ev); err != nil {
		return err
	}

	attrs := []any{
		"event_type", "certificate_revocation",
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"tenant_id", entry.TenantID,
		"client_id", entry.ClientID,
		"reason", entry.Reason,
		"revoked_by", entry.RevokedBy,
		"correlation_id", entry.CorrelationID,
	}
	if entry.SerialNumber != nil {
		attrs = append(attrs, "serial_number", *entry.SerialNumber)
	}
	if entry.KeyID != nil {
		attrs = append(attrs, "key_id", *entry.KeyID)
	}
	if entry.CAFingerprint != nil {
		attrs = append(attrs, "ca_fingerprint", *entry.CAFingerprint)
	}

	l.slogger.InfoContext(ctx, "certificate revoked", attrs...)

	return nil
}
