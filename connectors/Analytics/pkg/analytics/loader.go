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

// Package analytics wires the analytics connector into the core binary. It
// reads usage that core and other connectors produce and exposes read-only,
// membership-scoped views. The system runs without it.
package analytics

//go:generate go run github.com/swaggo/swag/cmd/swag init -g loader.go -d . -o ../../api --outputTypes yaml --parseDependency --useStructName
//go:generate mv ../../api/swagger.yaml ../../api/analytics.openapi.yaml

import (
	"context"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

func LoadConnector(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *coreservice.Service, wg *sync.WaitGroup, router *identity.Router, connectorConfig *config.ConnectorConfig) error {
	svc := NewService(coreService, database)
	NewHandlers(svc).RegisterRoutes(router)
	return nil
}
