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

// Package ldap is the LDAP Provisioner entry point. Wired from
// internal/connectors/loader.go alongside the COmanage connector — a site
// runs one or the other depending on whether it uses COmanage as a
// managed layer in front of the directory.
package ldap

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/store"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/subscribers"
	ldapdb "github.com/apache/airavata-custos/connectors/LDAP/Provisioner/db"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

const connectorName = "ldap"

// init registers the event types that close out an audit trace for this
// connector, so the audit-trace UI marks a provisioning run as complete
// instead of leaving it at in_progress. Mirrors the equivalent block in
// the COmanage and AMIE loaders.
func init() {
	tracing.RegisterTerminalMarkers("ldap",
		"LDAPAccountCreated",
		"LDAPAccountUpdated",
		"LDAPGroupCreated",
	)
}

// LoadConnector wires the subscriber to the event bus. Reads YAML config
// first and falls back to environment variables. If neither yields a
// complete config, it logs and returns nil without registering — same
// skip-with-log pattern the other connectors use so a dev server boots
// without LDAP credentials.
func LoadConnector(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup, _ *identity.Router, connectorConfig *config.ConnectorConfig) error {
	cfg, ok := loadConfigFromConnectorConfig(connectorConfig)
	if !ok {
		cfg, ok = loadConfigFromEnv()
		if !ok {
			slog.Info("ldap provisioner: required config not set; skipping")
			return nil
		}
	}

	if err := db.MigrateConnectorFS(database, ldapdb.MigrationFS(), "migrations", connectorName); err != nil {
		return err
	}

	ldapClient, err := client.New(cfg)
	if err != nil {
		return err
	}

	uidSeq := store.NewUIDSequence(database)
	if err := seedUIDCounter(ctx, ldapClient, uidSeq, cfg); err != nil {
		return err
	}

	subscribers.NewClusterUserSubscriber(ldapClient, uidSeq, eventBus, coreService, cfg.CustosClusterID).RegisterSubscribers()
	slog.Info("ldap provisioner: subscriber registered",
		"url", cfg.URL, "base_dn", cfg.BaseDN, "cluster_id", cfg.CustosClusterID)
	// Custos is the source of truth for provisioning: entries created by this
	// connector must not be edited directly in LDAP. There is no drift
	// reconciliation; out-of-band edits will be invisible to Custos.
	return nil
}

// seedUIDCounter initialises the ldap_uid_sequence row for this
// cluster. Runs one LDAP scan to find max(existing uidNumber) so the
// counter starts above any entries provisioned out-of-band before
// this connector ran. Idempotent: on subsequent boots the store's
// GREATEST() upsert preserves whatever value the counter has grown
// to, so a re-scan cannot regress the sequence.
func seedUIDCounter(ctx context.Context, ldapClient *client.Client, seq *store.UIDSequence, cfg client.Config) error {
	minUID := cfg.MinUID
	if minUID <= 0 {
		minUID = client.DefaultMinUID
	}

	// One-time LDAP scan to find the current head of the uidNumber
	// range. AllocateNextUID(minUID) returns max(uidNumber)+1, floored
	// at minUID — exactly the value we want as the counter's next_uid.
	// If the scan fails (e.g. LDAP unreachable at boot), fall back to
	// the floor so the connector still starts.
	initial, err := ldapClient.AllocateNextUID(minUID)
	if err != nil {
		slog.Warn("ldap provisioner: could not scan LDAP for existing max uidNumber during seed; falling back to floor",
			"floor", minUID, "err", err)
		initial = minUID
	}

	if err := seq.Seed(ctx, cfg.CustosClusterID, initial); err != nil {
		return err
	}
	slog.Info("ldap provisioner: uid sequence seeded",
		"cluster_id", cfg.CustosClusterID, "initial_next_uid", initial)
	return nil
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
		if v, ok := directory["verify_ssl"].(bool); ok {
			verifySSL = v
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
		if v, ok := provisioning["http_timeout"].(string); ok {
			if d, err := time.ParseDuration(v); err == nil {
				timeout = d
			}
		}
		if n := asInt64(provisioning["min_uid"]); n > 0 {
			minUID = n
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

func loadConfigFromEnv() (client.Config, bool) {
	url := os.Getenv("LDAP_URL")
	bindDN := os.Getenv("LDAP_BIND_DN")
	bindPassword := os.Getenv("LDAP_BIND_PASSWORD")
	baseDN := os.Getenv("LDAP_BASE_DN")
	groupBaseDN := os.Getenv("LDAP_GROUP_BASE_DN")
	custosCluster := os.Getenv("CUSTOS_CLUSTER_ID")

	if url == "" || bindDN == "" || bindPassword == "" || baseDN == "" || custosCluster == "" {
		return client.Config{}, false
	}

	verifySSL := true
	if v := os.Getenv("LDAP_VERIFY_SSL"); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			verifySSL = parsed
		}
	}

	timeout := 30 * time.Second
	if v := os.Getenv("LDAP_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}

	defaultShell := os.Getenv("LDAP_DEFAULT_SHELL")
	if defaultShell == "" {
		defaultShell = "/bin/bash"
	}
	homedirPrefix := os.Getenv("LDAP_HOMEDIR_PREFIX")
	if homedirPrefix == "" {
		homedirPrefix = "/home/"
	}

	minUID := client.DefaultMinUID
	if v := os.Getenv("LDAP_MIN_UID"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			minUID = n
		}
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
