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
	custosconfig "github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/events"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

const connectorName = "amie"

// LoadConnector skips silently when AMIE_BASE_URL / AMIE_SITE_CODE /
// AMIE_API_KEY are not all set.
func LoadConnector(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *coreservice.Service, wg *sync.WaitGroup, connectorConfig *custosconfig.ConnectorConfig) error {
	cfg := loadConfig(connectorConfig)
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
	userDNStore := store.NewUserDNStore(database)
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
		handler.NewRequestPersonMergeHandler(coreService, userDNStore, amie, auditSvc),
		handler.NewRequestUserModifyHandler(coreService, userDNStore, amie, auditSvc),
		handler.NewDataProjectCreateHandler(coreService, userDNStore, amie, auditSvc),
		handler.NewDataAccountCreateHandler(coreService, userDNStore, amie, auditSvc),
		handler.NewInformTransactionCompleteHandler(auditSvc),
		handler.NewNoOpHandler(),
	)

	met := metrics.New()
	poller := worker.NewPoller(amie, packetStore, eventStore, met, database, cfg.AMIE)
	processor := worker.NewProcessor(eventStore, packetStore, errorStore, router, met, database, cfg.AMIE)

	wg.Add(2)
	go func() {
		defer wg.Done()
		poller.Run(ctx)
		slog.Info("AMIE poller stopped")
	}()
	go func() {
		defer wg.Done()
		processor.Run(ctx)
		slog.Info("AMIE processor stopped")
	}()

	slog.Info("AMIE connector started", "site", cfg.AMIE.SiteCode, "baseURL", cfg.AMIE.BaseURL)
	return nil
}

func loadConfig(connectorConfig *custosconfig.ConnectorConfig) *config.Config {
	cfg := &config.Config{}

	// Load from connector config if available
	if connectorConfig != nil {
		if credentials, err := connectorConfig.GetNestedConfig("credentials"); err == nil {
			if url, ok := credentials["base_url"].(string); ok {
				cfg.AMIE.BaseURL = url
			}
			if code, ok := credentials["site_code"].(string); ok {
				cfg.AMIE.SiteCode = code
			}
			if key, ok := credentials["api_key"].(string); ok {
				cfg.AMIE.APIKey = key
			}
		}

		if polling, err := connectorConfig.GetNestedConfig("polling"); err == nil {
			if interval, ok := polling["poll_interval"].(string); ok {
				if d, err := time.ParseDuration(interval); err == nil {
					cfg.AMIE.PollInterval = d
				}
			}
			if interval, ok := polling["worker_interval"].(string); ok {
				if d, err := time.ParseDuration(interval); err == nil {
					cfg.AMIE.WorkerInterval = d
				}
			}
			if enabled, ok := polling["poller_enabled"].(bool); ok {
				cfg.AMIE.PollerEnabled = enabled
			}
		}

		if timeouts, err := connectorConfig.GetNestedConfig("timeouts"); err == nil {
			if timeout, ok := timeouts["connect_timeout"].(string); ok {
				if d, err := time.ParseDuration(timeout); err == nil {
					cfg.AMIE.ConnectTimeout = d
				}
			}
		}
	}

	// Fall back to environment variables
	if cfg.AMIE.BaseURL == "" {
		cfg.AMIE.BaseURL = os.Getenv("AMIE_BASE_URL")
	}
	if cfg.AMIE.SiteCode == "" {
		cfg.AMIE.SiteCode = os.Getenv("AMIE_SITE_CODE")
	}
	if cfg.AMIE.APIKey == "" {
		cfg.AMIE.APIKey = os.Getenv("AMIE_API_KEY")
	}
	if cfg.AMIE.PollInterval == 0 {
		cfg.AMIE.PollInterval = durationEnv("AMIE_POLL_INTERVAL", 30*time.Second)
	}
	if cfg.AMIE.WorkerInterval == 0 {
		cfg.AMIE.WorkerInterval = durationEnv("AMIE_WORKER_INTERVAL", 5*time.Second)
	}
	if cfg.AMIE.ConnectTimeout == 0 {
		cfg.AMIE.ConnectTimeout = durationEnv("AMIE_CONNECT_TIMEOUT", 5*time.Second)
	}
	if cfg.AMIE.ReadTimeout == 0 {
		cfg.AMIE.ReadTimeout = durationEnv("AMIE_READ_TIMEOUT", 20*time.Second)
	}
	cfg.AMIE.PollerEnabled = boolEnv("AMIE_POLLER_ENABLED", cfg.AMIE.PollerEnabled)

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
