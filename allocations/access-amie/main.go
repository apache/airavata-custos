// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/airavata-custos/allocations/access-amie/amieclient"
	"github.com/apache/airavata-custos/allocations/access-amie/config"
	"github.com/apache/airavata-custos/allocations/access-amie/db"
	"github.com/apache/airavata-custos/allocations/access-amie/handler"
	"github.com/apache/airavata-custos/allocations/access-amie/metrics"
	"github.com/apache/airavata-custos/allocations/access-amie/server"
	"github.com/apache/airavata-custos/allocations/access-amie/service"
	"github.com/apache/airavata-custos/allocations/access-amie/store"
	"github.com/apache/airavata-custos/allocations/access-amie/worker"
	domainstore "github.com/apache/airavata-custos/allocations/domain/store"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogging(cfg.Log)
	slog.Info("starting access-amie service", "port", cfg.Server.Port)

	database, err := db.Open(cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	slog.Info("connected to database")

	if err := db.Migrate(database, "db/migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	personStore := domainstore.NewPersonStore(database)
	personDNStore := domainstore.NewPersonDNStore(database)
	accountStore := domainstore.NewClusterAccountStore(database)
	projectStore := domainstore.NewProjectStore(database)
	membershipStore := domainstore.NewMembershipStore(database)
	packetStore := store.NewPacketStore(database)
	eventStore := store.NewEventStore(database)
	errorStore := store.NewProcessingErrorStore(database)
	auditStore := store.NewAuditStore(database)

	personSvc := service.NewPersonService(personStore, personDNStore, accountStore)
	accountSvc := service.NewUserAccountService(accountStore)
	projectSvc := service.NewProjectService(projectStore)
	membershipSvc := service.NewProjectMembershipService(membershipStore, projectStore, accountStore)
	auditSvc := service.NewAuditService(auditStore)

	amie := amieclient.New(cfg.AMIE)

	met := metrics.New()

	router := handler.NewRouter(
		handler.NewRequestProjectCreateHandler(personSvc, accountSvc, projectSvc, membershipSvc, amie, auditSvc),
		handler.NewRequestAccountCreateHandler(personSvc, accountSvc, projectSvc, membershipSvc, amie, auditSvc),
		handler.NewRequestProjectInactivateHandler(projectSvc, membershipSvc, amie, auditSvc),
		handler.NewRequestProjectReactivateHandler(projectSvc, membershipSvc, amie, auditSvc),
		handler.NewRequestAccountInactivateHandler(membershipSvc, amie, auditSvc),
		handler.NewRequestAccountReactivateHandler(membershipSvc, amie, auditSvc),
		handler.NewRequestPersonMergeHandler(personSvc, amie, auditSvc),
		handler.NewRequestUserModifyHandler(personSvc, amie, auditSvc),
		handler.NewDataProjectCreateHandler(personSvc, amie, auditSvc),
		handler.NewDataAccountCreateHandler(personSvc, amie, auditSvc),
		handler.NewInformTransactionCompleteHandler(auditSvc),
		handler.NewNoOpHandler(),
	)

	poller := worker.NewPoller(amie, packetStore, eventStore, met, database, cfg.AMIE)
	processor := worker.NewProcessor(eventStore, packetStore, errorStore, router, met, database, cfg.AMIE)

	srv := server.New(cfg.Server.Port, database, amie)
	go func() {
		slog.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("HTTP server stopped", "error", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go poller.Run(ctx)
	go processor.Run(ctx)

	slog.Info("access-amie service started successfully")
	<-ctx.Done()

	slog.Info("shutting down...")
	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}
	slog.Info("access-amie service stopped")
}

func setupLogging(cfg config.LogConfig) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if cfg.Format == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(h))
}
