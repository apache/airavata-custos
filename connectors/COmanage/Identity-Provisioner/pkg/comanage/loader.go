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

// Package comanage is the COmanage Identity-Provisioner entry point. Wired
// from internal/connectors/loader.go.
package comanage

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/subscribers"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

func init() {
	tracing.RegisterTerminalMarkers("comanage", "ComanageClusterAccountAttached")
}

// LoadConnector wires the subscriber to the event bus. If any required env
// var is missing, the loader logs and returns nil without registering.
func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup) error {
	cfg, ok := loadConfigFromEnv()
	if !ok {
		slog.Info("comanage provisioner: required env vars not set; skipping")
		return nil
	}
	httpClient := client.New(cfg)
	subscribers.NewClusterUserSubscriber(httpClient, eventBus, coreService, cfg.CustosClusterID).RegisterSubscribers()
	slog.Info("comanage provisioner: subscriber registered",
		"registry", cfg.RegistryURL, "co_id", cfg.COID, "cluster_id", cfg.CustosClusterID)
	// Custos is the source of truth: CoPerson and UnixClusterAccount records
	// for users provisioned via this connector must not be edited directly in
	// COmanage. There is no drift reconciliation; out-of-band edits will be
	// invisible to Custos.
	return nil
}

func loadConfigFromEnv() (client.Config, bool) {
	registryURL := os.Getenv("COMANAGE_REGISTRY_URL")
	coIDStr := os.Getenv("COMANAGE_CO_ID")
	apiUser := os.Getenv("COMANAGE_API_USER")
	apiKey := os.Getenv("COMANAGE_API_KEY")
	personIDType := os.Getenv("COMANAGE_PERSON_ID_TYPE")
	unixClusterStr := os.Getenv("COMANAGE_UNIX_CLUSTER_ID")
	custosCluster := os.Getenv("CUSTOS_CLUSTER_ID")

	if registryURL == "" || coIDStr == "" || apiUser == "" || apiKey == "" || personIDType == "" || unixClusterStr == "" || custosCluster == "" {
		return client.Config{}, false
	}
	coID, err := strconv.Atoi(coIDStr)
	if err != nil {
		slog.Error("comanage provisioner: COMANAGE_CO_ID is not an int", "value", coIDStr, "err", err)
		return client.Config{}, false
	}
	unixClusterID, err := strconv.Atoi(unixClusterStr)
	if err != nil {
		slog.Error("comanage provisioner: COMANAGE_UNIX_CLUSTER_ID is not an int", "value", unixClusterStr, "err", err)
		return client.Config{}, false
	}

	timeout := 30 * time.Second
	if v := os.Getenv("COMANAGE_HTTP_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}

	defaultShell := os.Getenv("COMANAGE_DEFAULT_SHELL")
	if defaultShell == "" {
		defaultShell = "/bin/bash"
	}
	homedirPrefix := os.Getenv("COMANAGE_HOMEDIR_PREFIX")
	if homedirPrefix == "" {
		homedirPrefix = "/home/"
	}

	return client.Config{
		RegistryURL:     registryURL,
		COID:            coID,
		APIUser:         apiUser,
		APIKey:          apiKey,
		PersonIDType:    personIDType,
		UnixClusterID:   unixClusterID,
		CustosClusterID: custosCluster,
		DefaultShell:    defaultShell,
		HomedirPrefix:   homedirPrefix,
		HTTPTimeout:     timeout,
	}, true
}
