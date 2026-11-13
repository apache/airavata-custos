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

package signer

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/Signer/internal"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

const Type = internal.Type

func LoadConnector(ctx context.Context, _ *sqlx.DB, _ *events.Bus, _ *service.Service, _ *sync.WaitGroup, router *identity.Router, cfg *config.ConnectorConfig) error {
	baseURL, _ := cfg.GetStringField("base_url")
	if baseURL == "" {
		slog.Info("signer: base_url not set; skipping")
		health.Record(Type, "config", errors.New("required setting base_url not set"))
		return nil
	}
	timeout, err := cfg.GetIntField("timeout_seconds")
	if err != nil {
		timeout = 10
	}
	handlers := internal.NewHandlers(baseURL, time.Duration(timeout)*time.Second)
	handlers.RegisterRoutes(router)
	// In the background so a down signer does not hold up startup.
	go handlers.CheckHealth(ctx)
	slog.Info("signer: routes registered", "base_url", baseURL)
	return nil
}
