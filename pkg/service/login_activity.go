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
	"fmt"
	"time"

	"github.com/apache/airavata-custos/internal/store"
)

// loginEventFutureSkew is the clock drift tolerated between the issuer and
// this server before an event timestamp is rejected as implausible.
const loginEventFutureSkew = 5 * time.Minute

// RecordLoginEventInput is trusted evidence derived from verified OIDC claims.
// It never carries client-supplied identity or timestamps.
type RecordLoginEventInput struct {
	EventKey   string
	OccurredAt time.Time
	Provider   string
	SessionID  string
}

// RecordLoginEventResult reports whether this login was new. A repeated
// callback for the same token is accepted but changes nothing.
type RecordLoginEventResult struct {
	Recorded bool `json:"recorded"`
}

// RecordLoginEvent folds one verified sign-in into the caller's activity
// history, bucketed by the user's own calendar day.
func (s *Service) RecordLoginEvent(ctx context.Context, userID string, input RecordLoginEventInput) (*RecordLoginEventResult, error) {
	if userID == "" || input.EventKey == "" {
		return nil, fmt.Errorf("%w: login event requires a user and an event key", ErrInvalidInput)
	}
	if input.OccurredAt.IsZero() {
		return nil, fmt.Errorf("%w: login event requires an occurrence time", ErrInvalidInput)
	}
	if input.OccurredAt.After(nowUTC().Add(loginEventFutureSkew)) {
		return nil, fmt.Errorf("%w: login event occurs too far in the future", ErrInvalidInput)
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user for login event: %w", err)
	}
	if user == nil {
		return nil, ErrNotFound
	}
	timezone := ""
	if user.Timezone != nil {
		timezone = *user.Timezone
	}
	zone, localDate := resolveLoginLocalDate(input.OccurredAt, timezone)

	var recorded bool
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		var txErr error
		recorded, txErr = s.loginActivity.Record(ctx, tx, store.LoginEvent{
			ID:         newID(),
			EventKey:   input.EventKey,
			UserID:     userID,
			OccurredAt: input.OccurredAt.UTC(),
			LocalDate:  localDate,
			Timezone:   zone,
			Provider:   input.Provider,
			SessionID:  input.SessionID,
		})
		return txErr
	}); err != nil {
		return nil, fmt.Errorf("record login event: %w", err)
	}
	return &RecordLoginEventResult{Recorded: recorded}, nil
}

// resolveLoginLocalDate buckets an instant into the user's calendar day,
// falling back to UTC when the stored zone is missing or unusable.
func resolveLoginLocalDate(occurredAt time.Time, timezone string) (string, string) {
	location := time.UTC
	if timezone != "" {
		if loaded, err := time.LoadLocation(timezone); err == nil {
			location = loaded
		} else {
			timezone = ""
		}
	}
	if timezone == "" {
		timezone = "UTC"
	}
	return timezone, occurredAt.In(location).Format(store.LoginDateLayout)
}
