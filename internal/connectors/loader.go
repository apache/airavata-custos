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

package connectors

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/pkg/amie"
	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/pkg/comanage"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/pkg/smapper"
	"github.com/apache/airavata-custos/connectors/SLURM/Usage-Monitor/pkg/monitor"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

// LoadConnectors brings every connector up at boot. Connectors register their
// own HTTP routes on mux; core does not.
func LoadConnectors(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, mux *http.ServeMux) error {
	slog.Info("loading connectors")

	slog.Info("loading SLURM Association Mapper connector")
	if err := smapper.LoadConnector(ctx, database, eventBus, coreService, wg, mux, nil); err != nil {
		slog.Error("failed to load SLURM Association Mapper connector", "error", err)
		return err
	}

	slog.Info("loading AMIE connector")
	if err := amie.LoadConnector(ctx, database, eventBus, coreService, wg, mux, nil); err != nil {
		slog.Error("failed to load AMIE connector", "error", err)
		return err
	}
	slog.Info("loading COmanage Identity-Provisioner connector")
	if err := comanage.LoadConnector(ctx, database, eventBus, coreService, wg, mux, nil); err != nil {
		slog.Error("failed to load COmanage Identity-Provisioner connector", "error", err)
		return err
	}

	slog.Info("loading SLURM Usage Monitor connector")
	if err := monitor.LoadConnector(ctx, database, eventBus, coreService, wg, mux, nil); err != nil {
		slog.Error("failed to load SLURM Usage Monitor connector", "error", err)
		return err
	}

	slog.Info("finished loading connectors")
	return nil
}

func LoadConnectorsFromConfig(ctx context.Context, cfg *config.Config, database *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, mux *http.ServeMux) error {
	slog.Info("loading connectors from config")

	connectorLoaders := map[string]func(context.Context, *sqlx.DB, *events.Bus, *service.Service, *sync.WaitGroup, *http.ServeMux, *config.ConnectorConfig) error{
		"slurm-association-mapper":      smapper.LoadConnector,
		"amie-processor":                amie.LoadConnector,
		"comanage-identity-provisioner": comanage.LoadConnector,
		"slurm-usage-monitor":           monitor.LoadConnector,
	}

	for connectorName, connectorCfg := range cfg.Connectors {
		if connectorCfg == nil {
			slog.Debug("connector config is nil", "name", connectorName)
			continue
		}

		if !connectorCfg.Enabled {
			slog.Info("connector is disabled", "name", connectorName, "type", connectorCfg.Type)
			continue
		}

		loader, ok := connectorLoaders[connectorCfg.Type]
		if !ok {
			slog.Warn("unknown connector type", "name", connectorName, "type", connectorCfg.Type)
			continue
		}

		slog.Info("loading connector", "name", connectorName, "type", connectorCfg.Type)
		if err := loader(ctx, database, eventBus, coreService, wg, mux, connectorCfg); err != nil {
			slog.Error("failed to load connector", "name", connectorName, "type", connectorCfg.Type, "error", err)
			return err
		}
	}

	slog.Info("finished loading connectors from config")
	return nil
}
