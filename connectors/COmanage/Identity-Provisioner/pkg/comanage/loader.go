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
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/subscribers"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

func init() {
	tracing.RegisterTerminalMarkers("comanage", "ComanageClusterAccountAttached")
}

// LoadConnector wires the subscriber to the event bus. Reads YAML config first
// and falls back to environment variables. If neither yields a complete
// config, it logs and returns nil without registering.
func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *http.ServeMux, connectorConfig *config.ConnectorConfig) error {
	cfg, ok := loadConfigFromConnectorConfig(connectorConfig)
	if !ok {
		cfg, ok = loadConfigFromEnv()
		if !ok {
			slog.Info("comanage provisioner: required config not set; skipping")
			return nil
		}
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

func loadConfigFromConnectorConfig(connectorConfig *config.ConnectorConfig) (client.Config, bool) {
	if connectorConfig == nil {
		return client.Config{}, false
	}

	var registryURL, apiUser, apiKey, personIDType, custosCluster, defaultShell, homedirPrefix string
	var coID, unixClusterID int
	timeout := 30 * time.Second

	// Load registry config
	if registry, err := connectorConfig.GetNestedConfig("registry"); err == nil {
		if url, ok := registry["url"].(string); ok {
			registryURL = url
		}
		if u, ok := registry["api_user"].(string); ok {
			apiUser = u
		}
		if k, ok := registry["api_key"].(string); ok {
			apiKey = k
		}
		coID = asInt(registry["co_id"])
	}

	// Load unix_cluster config
	if unixCluster, err := connectorConfig.GetNestedConfig("unix_cluster"); err == nil {
		unixClusterID = asInt(unixCluster["id"])
		if pType, ok := unixCluster["person_id_type"].(string); ok {
			personIDType = pType
		}
	}

	// Load provisioning config
	if provisioning, err := connectorConfig.GetNestedConfig("provisioning"); err == nil {
		if id, ok := provisioning["custos_cluster_id"].(string); ok {
			custosCluster = id
		}
		if shell, ok := provisioning["default_shell"].(string); ok {
			defaultShell = shell
		}
		if prefix, ok := provisioning["homedir_prefix"].(string); ok {
			homedirPrefix = prefix
		}
		if timeoutStr, ok := provisioning["http_timeout"].(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				timeout = d
			}
		}
	}

	if registryURL == "" || coID == 0 || apiUser == "" || apiKey == "" || personIDType == "" || unixClusterID == 0 || custosCluster == "" {
		return client.Config{}, false
	}

	if defaultShell == "" {
		defaultShell = "/bin/bash"
	}
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

// asInt handles the int / int64 / float64 shapes yaml.v3 may produce.
func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
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
