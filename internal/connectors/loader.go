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
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/pkg/amie"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/pkg/smapper"
	"github.com/apache/airavata-custos/connectors/SLURM/Usage-Monitor/pkg/monitor"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

func LoadConnectors(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup) error {
	slog.Info("loading connectors")

	slog.Info("loading SLURM Association Mapper connector")
	if err := smapper.LoadConnector(ctx, database, eventBus, coreService, wg); err != nil {
		slog.Error("failed to load SLURM Association Mapper connector", "error", err)
		return err
	}

	slog.Info("loading AMIE connector")
	if err := amie.LoadConnector(ctx, database, eventBus, coreService, wg); err != nil {
		slog.Error("failed to load AMIE connector", "error", err)
		return err
	}

	slog.Info("loading SLURM Usage Monitor connector")
	if err := monitor.LoadConnector(ctx, database, eventBus, coreService, wg); err != nil {
		slog.Error("failed to load SLURM Usage Monitor connector", "error", err)
		return err
	}

	slog.Info("finished loading connectors")
	return nil
}
