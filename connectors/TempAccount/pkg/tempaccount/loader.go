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

package tempaccount

import (
	"context"
	"sync"

	"github.com/apache/airavata-custos/connectors/TempAccount/internal"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
	"github.com/jmoiron/sqlx"
)

func LoadConnector(ctx context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, router *identity.Router, connectorConfig *config.ConnectorConfig) error {
	handlers := internal.NewHandlers(coreService)
	handlers.RegisterRoutes(router)
	return nil
}
