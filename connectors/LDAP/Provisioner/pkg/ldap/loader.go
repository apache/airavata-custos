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

// Package ldap is the LDAP Provisioner entry point. A site runs it or the
// COmanage connector, not both.
package ldap

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	ldapdb "github.com/apache/airavata-custos/connectors/LDAP/Provisioner/db"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/store"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/subscribers"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "ldap-identity-provisioner"

const connectorName = "ldap"

// LoadConnector wires the subscriber to the event bus. Without a complete
// config it logs and returns nil without registering, so a dev server
// boots without LDAP credentials.
func LoadConnector(_ context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *identity.Router, connectorConfig *config.ConnectorConfig) error {
	cfg, ok := loadConfigFromConnectorConfig(connectorConfig)
	if !ok {
		slog.Info("ldap provisioner: required config not set; skipping")
		return nil
	}

	if err := db.MigrateConnectorFS(database, ldapdb.MigrationFS(), "migrations", connectorName); err != nil {
		return err
	}

	ldapClient, err := client.New(cfg)
	if err != nil {
		return err
	}

	floor := uidFloor(ldapClient, cfg)
	subscribers.NewClusterUserSubscriber(ldapClient, store.NewAccounts(database), floor, eventBus, coreService, cfg.CustosClusterID).RegisterSubscribers(Type)
	slog.Info("ldap provisioner: subscriber registered",
		"url", cfg.URL, "base_dn", cfg.BaseDN, "cluster_id", cfg.CustosClusterID, "uid_floor", floor)
	// Custos is the source of truth for provisioning: entries created by this
	// connector must not be edited directly in LDAP. There is no drift
	// reconciliation; out-of-band edits will be invisible to Custos.
	return nil
}

// uidFloor is the lowest uidNumber to assign: above every uidNumber already
// in the directory, and at least min_uid. If the directory cannot be read
// at boot, min_uid is used so the connector still starts.
func uidFloor(ldapClient *client.Client, cfg client.Config) int64 {
	minUID := cfg.MinUID
	if minUID <= 0 {
		minUID = client.DefaultMinUID
	}
	floor, err := ldapClient.AllocateNextUID(minUID)
	if err != nil {
		slog.Warn("ldap provisioner: could not scan LDAP for existing max uidNumber; falling back to min_uid",
			"min_uid", minUID, "err", err)
		return minUID
	}
	return floor
}

func loadConfigFromConnectorConfig(connectorConfig *config.ConnectorConfig) (client.Config, bool) {
	if connectorConfig == nil {
		return client.Config{}, false
	}

	var url, bindDN, bindPassword, baseDN, custosCluster, defaultShell, homedirPrefix string
	verifySSL := true
	timeout := 30 * time.Second
	minUID := client.DefaultMinUID

	var groupBaseDN string

	if directory, err := connectorConfig.GetNestedConfig("directory"); err == nil {
		if v, ok := directory["url"].(string); ok {
			url = v
		}
		if v, ok := directory["bind_dn"].(string); ok {
			bindDN = v
		}
		if v, ok := directory["bind_password"].(string); ok {
			bindPassword = v
		}
		if v, ok := directory["base_dn"].(string); ok {
			baseDN = v
		}
		if v, ok := directory["group_base_dn"].(string); ok {
			groupBaseDN = v
		}
		if v, ok := directory["verify_ssl"]; ok {
			if b, ok := v.(bool); ok {
				verifySSL = b
			} else {
				slog.Warn("ldap provisioner: directory.verify_ssl is not a boolean; using the default",
					"value", v, "default", verifySSL)
			}
		}
	}

	if provisioning, err := connectorConfig.GetNestedConfig("provisioning"); err == nil {
		if v, ok := provisioning["custos_cluster_id"].(string); ok {
			custosCluster = v
		}
		if v, ok := provisioning["default_shell"].(string); ok {
			defaultShell = v
		}
		if v, ok := provisioning["homedir_prefix"].(string); ok {
			homedirPrefix = v
		}
		if v, ok := provisioning["timeout"].(string); ok {
			if d, err := time.ParseDuration(v); err == nil {
				timeout = d
			} else {
				slog.Warn("ldap provisioner: provisioning.timeout is not a duration; using the default",
					"value", v, "default", timeout, "err", err)
			}
		}
		if v, ok := provisioning["min_uid"]; ok {
			if n := asInt64(v); n > 0 {
				minUID = n
			} else {
				slog.Warn("ldap provisioner: provisioning.min_uid is not a positive integer; using the default",
					"value", v, "default", minUID)
			}
		}
	}

	if url == "" || bindDN == "" || bindPassword == "" || baseDN == "" || custosCluster == "" {
		return client.Config{}, false
	}

	if defaultShell == "" {
		defaultShell = "/bin/bash"
	}
	if homedirPrefix == "" {
		homedirPrefix = "/home/"
	}

	return client.Config{
		URL:             url,
		BindDN:          bindDN,
		BindPassword:    bindPassword,
		BaseDN:          baseDN,
		GroupBaseDN:     groupBaseDN,
		VerifySSL:       verifySSL,
		CustosClusterID: custosCluster,
		DefaultShell:    defaultShell,
		HomedirPrefix:   homedirPrefix,
		Timeout:         timeout,
		MinUID:          minUID,
	}, true
}

// asInt64 tolerates the int / int64 / float64 shapes yaml.v3 may produce.
func asInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}
