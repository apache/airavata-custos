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

// Package amie wires the ACCESS-AMIE connector into the core binary as a
// subsystem. It is selected at build time by importing the package blank in
// cmd/server/connectors.go; its init() registers a connectors.Connector that
// the host invokes during startup.
package amie

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amieclient"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/config"
	amiedb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/handler"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/metrics"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/worker"
	"github.com/apache/airavata-custos/pkg/connectors"
)

const connectorName = "amie"

type amieConnector struct{}

func (amieConnector) Name() string { return connectorName }

func (amieConnector) Migrations() (fs.FS, string) {
	return amiedb.MigrationFS(), "migrations"
}

func (amieConnector) Start(ctx context.Context, deps connectors.Deps) error {
	cfg := loadConfig()
	if cfg.AMIE.APIKey == "" || cfg.AMIE.BaseURL == "" || cfg.AMIE.SiteCode == "" {
		slog.Warn("amie connector disabled: AMIE_API_KEY, AMIE_BASE_URL, AMIE_SITE_CODE required")
		return nil
	}

	packetStore := store.NewPacketStore(deps.DB)
	eventStore := store.NewEventStore(deps.DB)
	errorStore := store.NewProcessingErrorStore(deps.DB)
	auditStore := store.NewAuditStore(deps.DB)
	auditSvc := service.NewAuditService(auditStore)

	// One AMIE site is tied to one downstream cluster by protocol, so the
	// cluster id is a per-deployment config value rather than a per-packet
	// lookup. Organizations are resolved per-packet from the *OrgCode and
	// *Organization fields the AMIE packet carries.
	clusterID := os.Getenv("AMIE_CLUSTER_ID")
	if clusterID == "" {
		slog.Warn("AMIE_CLUSTER_ID not set; request_project_create and request_account_create will fail when provisioning allocations/accounts")
	}

	amie := amieclient.New(cfg.AMIE)

	router := handler.NewRouter(
		handler.NewRequestProjectCreateHandler(deps.Service, clusterID, amie, auditSvc),
		handler.NewRequestAccountCreateHandler(deps.Service, clusterID, amie, auditSvc),
		handler.NewRequestProjectInactivateHandler(deps.Service, amie, auditSvc),
		handler.NewRequestProjectReactivateHandler(deps.Service, amie, auditSvc),
		handler.NewRequestAccountInactivateHandler(deps.Service, amie, auditSvc),
		handler.NewRequestAccountReactivateHandler(deps.Service, amie, auditSvc),
		handler.NewRequestPersonMergeHandler(deps.Service, amie, auditSvc),
		handler.NewRequestUserModifyHandler(deps.Service, amie, auditSvc),
		handler.NewDataProjectCreateHandler(deps.Service, amie, auditSvc),
		handler.NewDataAccountCreateHandler(deps.Service, amie, auditSvc),
		handler.NewInformTransactionCompleteHandler(auditSvc),
		handler.NewNoOpHandler(),
	)

	met := metrics.New()
	poller := worker.NewPoller(amie, packetStore, eventStore, met, deps.DB, cfg.AMIE)
	processor := worker.NewProcessor(eventStore, packetStore, errorStore, router, met, deps.DB, cfg.AMIE)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); poller.Run(ctx) }()
	go func() { defer wg.Done(); processor.Run(ctx) }()

	go func() {
		<-ctx.Done()
		wg.Wait()
		slog.Info("amie connector stopped")
	}()

	slog.Info("amie connector started", "site", cfg.AMIE.SiteCode, "baseURL", cfg.AMIE.BaseURL)
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

func init() {
	connectors.Register(amieConnector{})
}
