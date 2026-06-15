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

//go:generate go run github.com/swaggo/swag/cmd/swag init -g loader.go -d .,../../server,../../store -o ../../api --outputTypes yaml --parseDependency
//go:generate mv ../../api/swagger.yaml ../../api/amie.openapi.yaml

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amieclient"
	amiedb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/handler"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/metrics"
	amieserver "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/server"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/worker"
	custosconfig "github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/db"
	corestore "github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

func init() {
	tracing.RegisterTerminalMarkers("amie", "TRANSACTION_COMPLETE")
}

const connectorName = "amie"

// LoadConnector skips silently when AMIE_BASE_URL / AMIE_SITE_CODE /
// AMIE_API_KEY are not all set. When enabled, it attaches /connectors/amie/*
// endpoints to mux.
//
// @title	AMIE Connector API
// @version	0.1.0
// @description	REST endpoints for the ACCESS-CI AMIE connector, all under /connectors/amie/.
// @host	localhost:8080
// @BasePath	/
// @securityDefinitions.apikey	CustosUserHeader
// @in	header
// @name	X-Custos-User-Id
func LoadConnector(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *coreservice.Service, wg *sync.WaitGroup, mux *http.ServeMux, connectorConfig *custosconfig.ConnectorConfig) error {
	cfg := loadConfig(connectorConfig)
	if cfg.APIKey == "" || cfg.BaseURL == "" || cfg.SiteCode == "" {
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
	auditExtras := store.NewAuditExtrasStore(database)
	userDNStore := store.NewUserDNStore(database)
	auditSvc := service.NewAuditService(corestore.NewAuditEventStore(database), auditExtras)
	packetAuditStore := store.NewPacketAuditStore(database)
	if mux != nil {
		amieserver.NewHandlers(packetAuditStore).RegisterRoutes(mux)
	}

	// One AMIE site is tied to one downstream cluster by protocol, so cluster
	// identity is per-deployment configuration rather than per-packet.
	clusterID := loadClusterID(connectorConfig)
	if clusterID == "" {
		slog.Warn("AMIE cluster id not set (cluster.id in YAML or AMIE_CLUSTER_ID env); account and project create handlers will fail")
	}

	amie := amieclient.New(cfg)

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
	poller := worker.NewPoller(amie, packetStore, eventStore, met, database, cfg)
	processor := worker.NewProcessor(eventStore, packetStore, errorStore, router, met, auditSvc, database, cfg)

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

	slog.Info("AMIE connector started", "site", cfg.SiteCode, "baseURL", cfg.BaseURL)
	return nil
}

func loadConfig(connectorConfig *custosconfig.ConnectorConfig) amieclient.Config {
	cfg := amieclient.Config{}
	pollerEnabledSet := false

	if connectorConfig != nil {
		if credentials, err := connectorConfig.GetNestedConfig("credentials"); err == nil {
			if url, ok := credentials["base_url"].(string); ok {
				cfg.BaseURL = url
			}
			if code, ok := credentials["site_code"].(string); ok {
				cfg.SiteCode = code
			}
			if key, ok := credentials["api_key"].(string); ok {
				cfg.APIKey = key
			}
		}

		if polling, err := connectorConfig.GetNestedConfig("polling"); err == nil {
			if interval, ok := polling["poll_interval"].(string); ok {
				if d, err := time.ParseDuration(interval); err == nil {
					cfg.PollInterval = d
				}
			}
			if interval, ok := polling["worker_interval"].(string); ok {
				if d, err := time.ParseDuration(interval); err == nil {
					cfg.WorkerInterval = d
				}
			}
			if enabled, ok := polling["poller_enabled"].(bool); ok {
				cfg.PollerEnabled = enabled
				pollerEnabledSet = true
			}
		}

		if timeouts, err := connectorConfig.GetNestedConfig("timeouts"); err == nil {
			if timeout, ok := timeouts["connect_timeout"].(string); ok {
				if d, err := time.ParseDuration(timeout); err == nil {
					cfg.ConnectTimeout = d
				}
			}
			if timeout, ok := timeouts["read_timeout"].(string); ok {
				if d, err := time.ParseDuration(timeout); err == nil {
					cfg.ReadTimeout = d
				}
			}
		}
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = os.Getenv("AMIE_BASE_URL")
	}
	if cfg.SiteCode == "" {
		cfg.SiteCode = os.Getenv("AMIE_SITE_CODE")
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("AMIE_API_KEY")
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = durationEnv("AMIE_POLL_INTERVAL", 30*time.Second)
	}
	if cfg.WorkerInterval == 0 {
		cfg.WorkerInterval = durationEnv("AMIE_WORKER_INTERVAL", 5*time.Second)
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = durationEnv("AMIE_CONNECT_TIMEOUT", 5*time.Second)
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = durationEnv("AMIE_READ_TIMEOUT", 20*time.Second)
	}
	if !pollerEnabledSet {
		cfg.PollerEnabled = boolEnv("AMIE_POLLER_ENABLED", true)
	}

	return cfg
}

func loadClusterID(connectorConfig *custosconfig.ConnectorConfig) string {
	if connectorConfig != nil {
		if cluster, err := connectorConfig.GetNestedConfig("cluster"); err == nil {
			if id, ok := cluster["id"].(string); ok && id != "" {
				return id
			}
		}
	}
	return os.Getenv("AMIE_CLUSTER_ID")
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
