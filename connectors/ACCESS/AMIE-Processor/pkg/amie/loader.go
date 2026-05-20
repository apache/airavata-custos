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

// Package amie wires the ACCESS-AMIE connector into the core binary.
package amie

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amieclient"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/config"
	amiedb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/handler"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/metrics"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/worker"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/events"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

const connectorName = "amie"

// LoadConnector applies the connector's migrations and starts its background
// workers. It returns nil and skips when AMIE_BASE_URL / AMIE_SITE_CODE /
// AMIE_API_KEY are not all set.
func LoadConnector(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *coreservice.Service) error {
	cfg := loadConfig()
	if cfg.AMIE.APIKey == "" || cfg.AMIE.BaseURL == "" || cfg.AMIE.SiteCode == "" {
		slog.Warn("AMIE credentials not fully provided, skipping AMIE connector")
		return nil
	}

	migFS := amiedb.MigrationFS()
	if err := db.MigrateConnectorFS(database, migFS, "migrations", connectorName); err != nil {
		return err
	}

	packetStore := store.NewPacketStore(database)
	eventStore := store.NewEventStore(database)
	errorStore := store.NewProcessingErrorStore(database)
	auditStore := store.NewAuditStore(database)
	auditSvc := service.NewAuditService(auditStore)

	// One AMIE site is tied to one downstream cluster by protocol, so cluster
	// identity is per-deployment configuration rather than per-packet.
	clusterID := os.Getenv("AMIE_CLUSTER_ID")
	if clusterID == "" {
		slog.Warn("AMIE_CLUSTER_ID not set; account and project create handlers will fail")
	}

	amie := amieclient.New(cfg.AMIE)

	router := handler.NewRouter(
		handler.NewRequestProjectCreateHandler(coreService, clusterID, amie, auditSvc),
		handler.NewRequestAccountCreateHandler(coreService, clusterID, amie, auditSvc),
		handler.NewRequestProjectInactivateHandler(coreService, amie, auditSvc),
		handler.NewRequestProjectReactivateHandler(coreService, amie, auditSvc),
		handler.NewRequestAccountInactivateHandler(coreService, amie, auditSvc),
		handler.NewRequestAccountReactivateHandler(coreService, amie, auditSvc),
		handler.NewRequestPersonMergeHandler(coreService, amie, auditSvc),
		handler.NewRequestUserModifyHandler(coreService, amie, auditSvc),
		handler.NewDataProjectCreateHandler(coreService, amie, auditSvc),
		handler.NewDataAccountCreateHandler(coreService, amie, auditSvc),
		handler.NewInformTransactionCompleteHandler(auditSvc),
		handler.NewNoOpHandler(),
	)

	met := metrics.New()
	poller := worker.NewPoller(amie, packetStore, eventStore, met, database, cfg.AMIE)
	processor := worker.NewProcessor(eventStore, packetStore, errorStore, router, met, database, cfg.AMIE)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); poller.Run(ctx) }()
	go func() { defer wg.Done(); processor.Run(ctx) }()

	go func() {
		<-ctx.Done()
		wg.Wait()
		slog.Info("AMIE connector stopped")
	}()

	slog.Info("AMIE connector started", "site", cfg.AMIE.SiteCode, "baseURL", cfg.AMIE.BaseURL)
	return nil
}

func loadConfig() *config.Config {
	cfg := &config.Config{}
	cfg.AMIE.BaseURL = os.Getenv("AMIE_BASE_URL")
	cfg.AMIE.SiteCode = os.Getenv("AMIE_SITE_CODE")
	cfg.AMIE.APIKey = os.Getenv("AMIE_API_KEY")
	cfg.AMIE.PollInterval = durationEnv("AMIE_POLL_INTERVAL", 30*time.Second)
	cfg.AMIE.WorkerInterval = durationEnv("AMIE_WORKER_INTERVAL", 5*time.Second)
	cfg.AMIE.ConnectTimeout = durationEnv("AMIE_CONNECT_TIMEOUT", 5*time.Second)
	cfg.AMIE.ReadTimeout = durationEnv("AMIE_READ_TIMEOUT", 20*time.Second)
	cfg.AMIE.PollerEnabled = boolEnv("AMIE_POLLER_ENABLED", true)
	return cfg
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func boolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
