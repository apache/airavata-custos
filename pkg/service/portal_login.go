// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with this work
// for additional information regarding copyright ownership. The ASF licenses
// this file to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
)

// PortalLogin is trusted, verified OIDC evidence sent to the activity connector.
type PortalLogin struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PublishPortalLogin queues a sign-in in the durable bus transaction. Only the
// explicit sign-in callback calls this; ordinary authenticated requests do not.
func (s *Service) PublishPortalLogin(ctx context.Context, userID string, claims *identity.Claims) error {
	if userID == "" || claims == nil || claims.Sub == "" || claims.Issuer == "" {
		return fmt.Errorf("%w: portal login identity missing", ErrInvalidInput)
	}
	var evidence string
	var occurredAt int64
	if claims.TokenID != "" && claims.IssuedAt > 0 {
		evidence = "token\x00" + claims.TokenID
		occurredAt = claims.IssuedAt
	} else if claims.SessionID != "" && claims.AuthTime > 0 {
		evidence = "session\x00" + claims.SessionID
		occurredAt = claims.AuthTime
	} else {
		return fmt.Errorf("%w: token has no stable login identifier", ErrInvalidInput)
	}
	when := time.Unix(occurredAt, 0).UTC()
	if when.After(time.Now().Add(5 * time.Minute)) {
		return fmt.Errorf("%w: login time is in the future", ErrInvalidInput)
	}
	hash := sha256.Sum256([]byte("v1\x00" + claims.Issuer + "\x00" + claims.Sub + "\x00" + evidence + fmt.Sprintf("\x00%d", occurredAt)))
	login := PortalLogin{ID: "portal-login:" + hex.EncodeToString(hash[:]), UserID: userID, OccurredAt: when}
	return s.inTx(ctx, func(tx *sql.Tx) error { return s.eventBus.Publish(ctx, tx, events.PortalLoginEvent, login) })
}
