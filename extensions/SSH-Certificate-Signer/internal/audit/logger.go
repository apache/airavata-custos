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
	SerialNumber         int64
	KeyID                string
	Principal            string
	UserEmail            string
	UserSub              string
	RequestedBy          string
	PublicKeyFingerprint string
	CAFingerprint        string
	ValidAfter           time.Time
	ValidBefore          time.Time
	GrantedExtensions    []string
	ForceCommand         *string
	SourceAddress        *string
}

func (l *Logger) LogIssuance(ctx context.Context, entry *IssuanceEntry) error {
	dbLog := &store.IssuanceLog{
		SerialNumber:         entry.SerialNumber,
		KeyID:                entry.KeyID,
		Principal:            entry.Principal,
		UserEmail:            entry.UserEmail,
		UserSub:              entry.UserSub,
		RequestedBy:          entry.RequestedBy,
		PublicKeyFingerprint: entry.PublicKeyFingerprint,
		CAFingerprint:        entry.CAFingerprint,
		ValidAfter:           entry.ValidAfter,
		ValidBefore:          entry.ValidBefore,
		GrantedExtensions:    entry.GrantedExtensions,
		ForceCommand:         entry.ForceCommand,
		SourceAddress:        entry.SourceAddress,
	}

	if err := l.db.InsertIssuanceLog(ctx, dbLog); err != nil {
		return err
	}

	l.slogger.InfoContext(ctx, "certificate issued",
		"event_type", "certificate_issuance",
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"serial_number", entry.SerialNumber,
		"key_id", entry.KeyID,
		"principal", entry.Principal,
		"user_email", entry.UserEmail,
		"user_sub", entry.UserSub,
		"requested_by", entry.RequestedBy,
		"public_key_fingerprint", entry.PublicKeyFingerprint,
		"ca_fingerprint", entry.CAFingerprint,
		"valid_after", entry.ValidAfter.Unix(),
		"valid_before", entry.ValidBefore.Unix(),
		"granted_extensions", entry.GrantedExtensions,
		"force_command", entry.ForceCommand,
		"source_address", entry.SourceAddress,
	)

	return nil
}

func (l *Logger) LogCACreated(ctx context.Context, caFingerprint string) {
	l.slogger.WarnContext(ctx, "CA created",
		"event_type", "ca_created",
		"timestamp", time.Now().UTC().Format(time.RFC3339),
		"ca_fingerprint", caFingerprint,
	)
}
