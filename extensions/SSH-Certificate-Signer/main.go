// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/apache/airavata-custos/signer/internal/audit"
	"github.com/apache/airavata-custos/signer/internal/cert"
	"github.com/apache/airavata-custos/signer/internal/config"
	"github.com/apache/airavata-custos/signer/internal/handler"
	"github.com/apache/airavata-custos/signer/internal/policy"
	"github.com/apache/airavata-custos/signer/internal/server"
	signerservice "github.com/apache/airavata-custos/signer/internal/service"
	"github.com/apache/airavata-custos/signer/internal/store"
	"github.com/apache/airavata-custos/signer/internal/validation"
	"github.com/apache/airavata-custos/signer/internal/vault"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const usage = `Custos Signer Service - SSH certificate signing service.

Usage:
  custos-signer <command> [options]

Commands:
  serve      Start the HTTP server (default)
  migrate    Apply pending database migrations and exit

Options:
  --config string    Path to configuration file (default "config.yaml")
  --auto-migrate     Automatically apply migrations on startup (serve only)
`

func main() {
	cmd := "serve"
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	}

	configPath := "config.yaml"
	autoMigrate := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config":
			if i+1 < len(args) {
				configPath = args[i+1]
				i++
			}
		case "--auto-migrate":
			autoMigrate = true
		case "--help", "-h":
			fmt.Print(usage)
			os.Exit(0)
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Logging)
	slog.SetDefault(logger)

	switch cmd {
	case "migrate":
		runMigrate(cfg, logger)
		logger.Info("migrations complete")
		os.Exit(0)
	case "serve":
		runServer(cfg, logger, autoMigrate)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}
}

// Applies pending migrations and exits.
// Invoked by the "migrate" subcommand and also reused by --auto-migrate in serve mode.
func runMigrate(cfg *config.Config, logger *slog.Logger) {
	dsn := fmt.Sprintf("pgx5://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.Username, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		logger.Error("failed to create migrate instance", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("database migrations: no change")
		} else {
			logger.Error("failed to run database migrations", "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("database migrations applied successfully")
	}
}

func runServer(cfg *config.Config, logger *slog.Logger, autoMigrate bool) {
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("connected to database")

	if autoMigrate {
		logger.Info("auto-migrate enabled, applying pending migrations")
		runMigrate(cfg, logger)
	}

	vaultClient, err := vault.NewClient(cfg.Vault, cfg.Signer.CA.Rotation)
	if err != nil {
		logger.Error("failed to create vault client", "error", err)
		os.Exit(1)
	}
	logger.Info("vault client initialized")

	policyEnforcer := policy.NewEnforcer(cfg.Signer.Policy.MaxTTLSeconds, cfg.Signer.Policy.AllowedKeyTypes)
	auditLogger := audit.NewLogger(db, logger)

	caFingerprint, err := loadCA(context.Background(), vaultClient, cfg.Signer.CA.CreateIfMissing, auditLogger)
	if err != nil {
		logger.Error("failed to load CA", "error", err)
		os.Exit(1)
	}
	logger.Info("CA loaded", "ca_fingerprint", caFingerprint)

	cacheTTL := time.Duration(cfg.Signer.Validation.CacheTTLSeconds) * time.Second
	ldapConnector := validation.NewDefaultLDAPConnector()
	principalValidator := validation.NewValidatorDispatcher(
		vaultClient, ldapConnector,
		cfg.Signer.Validation.PrincipalValidator,
		cacheTTL, logger,
	)
	logger.Info("principal validation dispatcher initialized",
		"source", cfg.Signer.Validation.PrincipalValidator,
		"cache_ttl_seconds", cfg.Signer.Validation.CacheTTLSeconds)

	signHandler := handler.NewSignHandler(policyEnforcer, principalValidator, vaultClient, auditLogger, handler.SignSettings{
		ReservedUsernames: cfg.Signer.ReservedUsernames,
		DeniedExtensions:  cfg.Signer.Policy.DeniedExtensions,
		LoginHost:         cfg.Signer.LoginHost,
		LoginPort:         cfg.Signer.LoginPort,
	}, logger)
	caPublicKeyHandler := handler.NewCAPublicKeyHandler(vaultClient, logger)
	healthHandler := handler.NewHealthHandler(db, vaultClient)
	certificatesHandler := handler.NewCertificatesHandler(db, signerservice.NewRevocationService(db), logger)

	handlers := server.Handlers{
		Sign:              signHandler.Handle,
		CAPublicKey:       caPublicKeyHandler.Handle,
		Health:            healthHandler.Handle,
		Certificates:      certificatesHandler.HandleList,
		CertificateDetail: certificatesHandler.HandleGet,
		CertificateRevoke: certificatesHandler.HandleRevoke,
	}

	router := server.NewRouter(cfg, handlers)

	srv := server.New(cfg.Server, router, logger)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

// loadCA makes sure the CA exists before the signer serves requests and returns its fingerprint.
func loadCA(ctx context.Context, vaultClient *vault.Client, createIfMissing bool, auditLogger *audit.Logger) (string, error) {
	kp, err := vaultClient.GetCurrentCAKey(ctx)
	if errors.Is(err, vault.ErrCANotFound) {
		if !createIfMissing {
			return "", fmt.Errorf("%w: store the CA key, or set signer.ca.create_if_missing to create one", err)
		}
		if kp, err = vaultClient.GenerateCAKey(ctx); err != nil {
			return "", err
		}
		fingerprint, err := cert.CAFingerprint(kp.PublicKey)
		auditLogger.LogCACreated(ctx, fingerprint)
		return fingerprint, err
	}
	if err != nil {
		return "", err
	}
	return cert.CAFingerprint(kp.PublicKey)
}

func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
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
	if strings.ToLower(cfg.Format) == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(h)
}
