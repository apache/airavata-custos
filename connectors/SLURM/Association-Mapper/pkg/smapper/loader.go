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

package smapper

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/subscribers"
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *identity.Router, connectorConfig *config.ConnectorConfig) error {

	// Read url, username, and password from config or environment variables
	var apiUrl, user, token, apiVersion string

	if connectorConfig != nil {
		slurmAPI, err := connectorConfig.GetNestedConfig("slurm_api")
		if err == nil {
			if url, ok := slurmAPI["url"].(string); ok {
				apiUrl = url
			}
			if u, ok := slurmAPI["username"].(string); ok {
				user = u
			}
			if t, ok := slurmAPI["token"].(string); ok {
				token = t
			}
			if v, ok := slurmAPI["version"].(string); ok {
				apiVersion = v
			}
		}
	}

	// Fall back to environment variables
	if apiUrl == "" {
		apiUrl = os.Getenv("SLURM_API")
	}
	if user == "" {
		user = os.Getenv("SLURM_USER")
	}
	if token == "" {
		token = os.Getenv("SLURM_TOKEN")
	}
	if apiVersion == "" {
		apiVersion = os.Getenv("SLURM_API_VERSION")
	}

	if apiUrl == "" || user == "" || token == "" || apiVersion == "" {
		slog.Info("SLURM API credentials not fully provided, skipping SLURM Association Mapper connector")
		slog.Info("SLURM API credentials", "apiUrl", apiUrl, "user", user, "token", token, "apiVersion", apiVersion)
		return nil
	}

	slurmClient := client.New(apiUrl, user, token, apiVersion)
	subscribers.NewAssociationSubscriber(slurmClient, eventBus, coreService).RegisterSubscribers()
	return nil
}
