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

// Command server starts the Custos HTTP API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/connectors"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/server"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := envDefault("CONFIG_PATH", "config/custos.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Warn("config file not found, falling back to environment variables", "config_path", configPath, "error", err)
		return runLegacy()
	}

	slog.Info("loaded config", "path", configPath)

	dsn := cfg.Core.Database.URL
	if dsn == "" {
		return errors.New("database.url in config is required")
	}

	port := cfg.Core.API.Port
	if port == 0 {
		port = 8080
	}
	addr := ":" + strconv.Itoa(port)

	maxOpen := envInt("DB_MAX_OPEN_CONNS", 25)
	maxIdle := envInt("DB_MAX_IDLE_CONNS", 5)

	database, err := db.Open(db.Config{
		DSN:          dsn,
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	// Core schema must run before connector migrations because connector
	// schemas may FK into core tables.
	if err := db.MigrateEmbedded(database); err != nil {
		return err
	}

	// Create a new event bus instance to async messaging between service and connectors
	eventBus := events.New()
	svc := service.New(database, eventBus)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tryBootstrap(ctx, svc)

	// Tracks every background goroutine spawned by connectors so we can wait
	// for them to drain on shutdown instead of killing them mid-flight.
	var connectorsWG sync.WaitGroup
	if err := connectors.LoadConnectorsFromConfig(ctx, cfg, database, eventBus, svc, &connectorsWG); err != nil {
		return err
	}

	handler := server.LoggingMiddleware(server.New(svc))

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("waiting for connectors to drain")
	connectorsDone := make(chan struct{})
	go func() {
		connectorsWG.Wait()
		close(connectorsDone)
	}()
	select {
	case <-connectorsDone:
		slog.Info("connectors drained cleanly")
	case <-time.After(30 * time.Second):
		slog.Warn("connector drain timed out; some workers may have leaked")
	}

	slog.Info("server stopped cleanly")
	return nil
}

func runLegacy() error {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return errors.New("DATABASE_DSN environment variable is required " +
			"(e.g. user:pass@tcp(localhost:3306)/custos?parseTime=true&charset=utf8mb4)")
	}

	addr := envDefault("HTTP_ADDR", ":8080")
	maxOpen := envInt("DB_MAX_OPEN_CONNS", 25)
	maxIdle := envInt("DB_MAX_IDLE_CONNS", 5)

	database, err := db.Open(db.Config{
		DSN:          dsn,
		MaxOpenConns: maxOpen,
		MaxIdleConns: maxIdle,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	// Core schema must run before connector migrations because connector
	// schemas may FK into core tables.
	if err := db.MigrateEmbedded(database); err != nil {
		return err
	}

	// Create a new event bus instance to async messaging between service and connectors
	eventBus := events.New()
	svc := service.New(database, eventBus)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tryBootstrap(ctx, svc)

	// Tracks every background goroutine spawned by connectors so we can wait
	// for them to drain on shutdown instead of killing them mid-flight.
	var connectorsWG sync.WaitGroup
	if err := connectors.LoadConnectors(ctx, database, eventBus, svc, &connectorsWG); err != nil {
		return err
	}

	handler := server.LoggingMiddleware(server.New(svc))

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	slog.Info("waiting for connectors to drain")
	connectorsDone := make(chan struct{})
	go func() {
		connectorsWG.Wait()
		close(connectorsDone)
	}()
	select {
	case <-connectorsDone:
		slog.Info("connectors drained cleanly")
	case <-time.After(30 * time.Second):
		slog.Warn("connector drain timed out; some workers may have leaked")
	}

	slog.Info("server stopped cleanly")
	return nil
}

func envDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid integer env var, using default", "key", key, "value", v, "default", fallback)
		return fallback
	}
	return n
}
