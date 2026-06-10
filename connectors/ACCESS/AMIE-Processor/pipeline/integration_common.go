//go:build integration

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

package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amieclient"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/config"
	amiedb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/handler"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/metrics"
	amieservice "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/worker"
	"github.com/apache/airavata-custos/internal/db"
	corestore "github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

const testClusterID = "00000000-0000-0000-0000-000000000001"

var (
	sharedDB         *sqlx.DB
	sharedDBOnce     sync.Once
	sharedDBErr      error
	tracingInitOnce  sync.Once
	tracingInitError error
)

func isLocalAMIEConfigAvailable() bool {
	for _, key := range []string{
		"DATABASE_DSN",
		"AMIE_BASE_URL",
		"AMIE_SITE_CODE",
		"AMIE_API_KEY",
		"AMIE_CLUSTER_ID",
	} {
		if os.Getenv(key) == "" {
			return false
		}
	}
	return true
}

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	if !isLocalAMIEConfigAvailable() {
		t.Skip("integration env not set; run via scripts/run-amie-integration-tests.sh")
	}
	sharedDBOnce.Do(func() {
		database, err := db.Open(db.Config{
			DSN:          os.Getenv("DATABASE_DSN"),
			MaxOpenConns: 5,
			MaxIdleConns: 2,
		})
		if err != nil {
			sharedDBErr = err
			return
		}
		if err := db.MigrateEmbedded(database); err != nil {
			sharedDBErr = err
			return
		}
		if err := db.MigrateConnectorFS(database, amiedb.MigrationFS(), "migrations", "amie"); err != nil {
			sharedDBErr = err
			return
		}
		sharedDB = database
	})
	if sharedDBErr != nil {
		t.Fatalf("setup db: %v", sharedDBErr)
	}
	tracingInitOnce.Do(func() {
		_, tracingInitError = tracing.Init(tracing.InitConfig{
			Mode:        tracing.ModeProduction,
			Logger:      slog.Default(),
			ServiceName: "custos",
		})
	})
	if tracingInitError != nil {
		t.Fatalf("tracing init: %v", tracingInitError)
	}
	truncateAll(t, sharedDB)
	seedCluster(t, sharedDB)
	return sharedDB
}

func truncateAll(t *testing.T, database *sqlx.DB) {
	t.Helper()
	tables := []string{
		"amie_audit_extras",
		"audit_events",
		"amie_processing_errors",
		"amie_processing_events",
		"amie_packets",
		"amie_user_dns",
		"compute_allocation_membership_resource_overrides",
		"compute_allocation_usages",
		"compute_allocation_memberships",
		"compute_allocation_change_request_events",
		"compute_allocation_change_requests",
		"compute_allocation_diffs",
		"compute_allocation_resource_rates",
		"compute_allocation_resource_mappings",
		"compute_allocation_resources",
		"compute_allocations",
		"compute_cluster_users",
		"compute_clusters",
		"projects",
		"user_identities",
		"users",
		"organizations",
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK: %v", err)
	}
	for _, tbl := range tables {
		if _, err := database.Exec("TRUNCATE TABLE " + tbl); err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("re-enable FK: %v", err)
	}
}

func seedCluster(t *testing.T, database *sqlx.DB) {
	t.Helper()
	if _, err := database.Exec(
		"INSERT INTO compute_clusters (id, name) VALUES (?, ?)",
		testClusterID, "default-cluster",
	); err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
}

// testPipeline runs the AMIE connector in-process against the local mock
// server.
type testPipeline struct {
	db       *sqlx.DB
	baseURL  string
	siteCode string
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// newTestPipeline mirrors pkg/amie.LoadConnector but with short poll/worker
// intervals so tests don't wait the production default of 30s/5s.
func newTestPipeline(t *testing.T) *testPipeline {
	t.Helper()
	database := setupTestDB(t)

	cfg := config.AMIEConfig{
		BaseURL:        os.Getenv("AMIE_BASE_URL"),
		SiteCode:       os.Getenv("AMIE_SITE_CODE"),
		APIKey:         os.Getenv("AMIE_API_KEY"),
		PollInterval:   200 * time.Millisecond,
		WorkerInterval: 100 * time.Millisecond,
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    20 * time.Second,
		PollerEnabled:  true,
	}

	amieClient := amieclient.New(cfg)
	eventBus := events.New()
	coreSvc := coreservice.New(database, eventBus)

	packetStore := store.NewPacketStore(database)
	eventStore := store.NewEventStore(database)
	errorStore := store.NewProcessingErrorStore(database)
	auditExtras := store.NewAuditExtrasStore(database)
	userDNStore := store.NewUserDNStore(database)
	auditSvc := amieservice.NewAuditService(corestore.NewAuditEventStore(database), auditExtras)

	router := handler.NewRouter(
		handler.NewRequestProjectCreateHandler(coreSvc, testClusterID, amieClient, auditSvc),
		handler.NewRequestAccountCreateHandler(coreSvc, testClusterID, amieClient, auditSvc),
		handler.NewRequestProjectInactivateHandler(coreSvc, amieClient, auditSvc),
		handler.NewRequestProjectReactivateHandler(coreSvc, amieClient, auditSvc),
		handler.NewRequestAccountInactivateHandler(coreSvc, amieClient, auditSvc),
		handler.NewRequestAccountReactivateHandler(coreSvc, amieClient, auditSvc),
		handler.NewRequestPersonMergeHandler(coreSvc, userDNStore, amieClient, auditSvc),
		handler.NewRequestUserModifyHandler(coreSvc, userDNStore, amieClient, auditSvc),
		handler.NewDataProjectCreateHandler(coreSvc, userDNStore, amieClient, auditSvc),
		handler.NewDataAccountCreateHandler(coreSvc, userDNStore, amieClient, auditSvc),
		handler.NewInformTransactionCompleteHandler(auditSvc),
		handler.NewNoOpHandler(),
	)

	met := metrics.NewWithRegistry(prometheus.NewRegistry())
	poller := worker.NewPoller(amieClient, packetStore, eventStore, met, database, cfg)
	processor := worker.NewProcessor(eventStore, packetStore, errorStore, router, met, auditSvc, database, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	pipe := &testPipeline{
		db:       database,
		baseURL:  cfg.BaseURL,
		siteCode: cfg.SiteCode,
		cancel:   cancel,
	}
	pipe.wg.Add(2)
	go func() {
		defer pipe.wg.Done()
		poller.Run(ctx)
	}()
	go func() {
		defer pipe.wg.Done()
		processor.Run(ctx)
	}()
	return pipe
}

func (p *testPipeline) stop() {
	p.cancel()
	p.wg.Wait()
	flushTracing()
}

func flushTracing() {
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.ForceFlush(ctx)
	}
}

func (p *testPipeline) fireScenario(t *testing.T, scenarioType string) {
	t.Helper()
	url := fmt.Sprintf("%s/test/%s/scenarios?type=%s", p.baseURL, p.siteCode, scenarioType)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		t.Fatalf("scenario %s returned status %d", scenarioType, resp.StatusCode)
	}
}

// waitForDrain blocks until at least expectPackets have been ingested AND
// no rows are pending. The "at least N ingested" guard avoids a race where
// the poller hasn't fired yet so a naive drain check returns immediately.
func (p *testPipeline) waitForDrain(t *testing.T, expectPackets int, deadline time.Duration) int {
	t.Helper()
	end := time.Now().Add(deadline)
	for {
		var total, pendingDecode, pendingProcess int
		if err := p.db.Get(&total, "SELECT COUNT(*) FROM amie_packets"); err != nil {
			t.Fatalf("count packets: %v", err)
		}
		if err := p.db.Get(&pendingDecode,
			"SELECT COUNT(*) FROM amie_packets WHERE status = 'NEW'",
		); err != nil {
			t.Fatalf("count pending packets: %v", err)
		}
		if err := p.db.Get(&pendingProcess,
			"SELECT COUNT(*) FROM amie_processing_events WHERE status NOT IN ('SUCCEEDED','FAILED','PERMANENTLY_FAILED')",
		); err != nil {
			t.Fatalf("count pending events: %v", err)
		}
		if total >= expectPackets && pendingDecode == 0 && pendingProcess == 0 {
			break
		}
		if time.Now().After(end) {
			t.Fatalf("drain timeout: total=%d (want >=%d), pendingDecode=%d, pendingProcess=%d",
				total, expectPackets, pendingDecode, pendingProcess)
		}
		time.Sleep(250 * time.Millisecond)
	}
	var decoded int
	if err := p.db.Get(&decoded,
		"SELECT COUNT(*) FROM amie_packets WHERE status = 'DECODED'",
	); err != nil {
		t.Fatalf("count decoded: %v", err)
	}
	return decoded
}

func countRows(t *testing.T, database *sqlx.DB, table string) int {
	t.Helper()
	var n int
	if err := database.Get(&n, "SELECT COUNT(*) FROM "+table); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}
