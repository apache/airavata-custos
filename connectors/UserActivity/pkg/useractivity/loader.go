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

package useractivity

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
	"github.com/jmoiron/sqlx"
)

const Type = "user-activity"

func LoadConnector(_ context.Context, database *sqlx.DB, bus *events.Bus, _ *service.Service, _ *sync.WaitGroup, router *identity.Router, _ *config.ConnectorConfig) error {
	connector := &Connector{db: database}
	bus.Subscribe(Type, events.PortalLoginEvent, connector.recordLogin)
	connector.registerRoutes(router)
	return nil
}

type Connector struct{ db *sqlx.DB }

func (c *Connector) recordLogin(ctx context.Context, _ events.Event, payload any) error {
	var login service.PortalLogin
	switch value := payload.(type) {
	case json.RawMessage:
		if err := json.Unmarshal(value, &login); err != nil {
			return fmt.Errorf("decode portal login: %w", err)
		}
	case service.PortalLogin:
		login = value
	default:
		return fmt.Errorf("unexpected portal login payload %T", payload)
	}
	if login.ID == "" || login.UserID == "" || login.OccurredAt.IsZero() {
		return fmt.Errorf("incomplete portal login event")
	}
	traceID, spanID, err := tracing.IDsFromContext(ctx)
	if err != nil {
		return err
	}
	// The deterministic ID makes redelivery and repeated sign-in callbacks safe.
	_, err = c.db.ExecContext(ctx, `INSERT INTO audit_events
		(id, event_type, event_time, entity_id, entity_type, details, source, trace_id, span_id)
		VALUES ($1, 'PORTAL_LOGIN', $2, $3, 'user', '{}', 'user-activity', $4, $5)
		ON CONFLICT (id) DO NOTHING`, login.ID, login.OccurredAt, login.UserID, traceID, spanID)
	return err
}
