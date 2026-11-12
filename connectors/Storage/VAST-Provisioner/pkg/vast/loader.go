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

// Package vast is the VAST-Provisioner entry point. Wired from
// internal/connectors/loader.go.
package vast

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/subscribers"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "vast-storage-provisioner"

func LoadConnector(ctx context.Context, database *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, router *identity.Router, connectorConfig *config.ConnectorConfig) error {
	if _, err := database.ExecContext(ctx, pullFolders); err != nil {
		return fmt.Errorf("vast provisioner: pull vast_folders into storages: %w", err)
	}
	entries, _ := connectorConfig.Config["clusters"].([]interface{})
	clusters := map[string]subscribers.Cluster{}
	for _, e := range entries {
		cluster, _ := e.(map[string]interface{})
		vms, _ := cluster["vast"].(map[string]interface{})
		vmsURL, username, password := setting(vms, "url"), setting(vms, "username"), setting(vms, "password")
		tenantID, _ := vms["tenant_id"].(int)
		clusterID := setting(cluster, "id")
		if vmsURL == "" || username == "" || password == "" || tenantID == 0 || clusterID == "" {
			return fmt.Errorf("vast provisioner: cluster %q needs id, vast.url, vast.username, vast.password and vast.tenant_id", clusterID)
		}
		if cluster["targets"] != nil {
			return fmt.Errorf("vast provisioner: cluster %s: targets moved to core; create them with POST /storage-policies", clusterID)
		}
		listed, _ := vms["storages"].([]interface{})
		views := map[models.StorageType]bool{}
		for _, e := range listed {
			entry, _ := e.(map[string]interface{})
			t := models.StorageType(setting(entry, "type"))
			if _, _, ok := t.Takes(); !ok {
				return fmt.Errorf("vast provisioner: cluster %s: vast.storages: %q is not a storage type", clusterID, t)
			}
			if view, _ := entry["view"].(bool); view {
				views[t] = true
			}
		}
		c := subscribers.Cluster{VMS: client.New(vmsURL, username, password, tenantID), Views: views}
		// Verify that view policy is set otherwise stop the startup instead of failing every create request.
		if len(views) > 0 {
			policy := setting(vms, "view_policy")
			for p := range strings.SplitSeq(setting(vms, "view_protocols"), ",") {
				if p = strings.TrimSpace(p); p != "" {
					c.ViewProtocols = append(c.ViewProtocols, p)
				}
			}
			if policy == "" || len(c.ViewProtocols) == 0 {
				return fmt.Errorf("vast provisioner: cluster %s has a vast.storages view, so vast.view_policy and vast.view_protocols are required", clusterID)
			}
			var err error
			if c.ViewPolicyID, err = c.VMS.ViewPolicyID(ctx, policy); err != nil {
				return fmt.Errorf("vast provisioner: cluster %s: %w", clusterID, err)
			}
		}
		clusters[clusterID] = c
		slog.Info("vast provisioner: cluster registered", "vms", vmsURL, "cluster_id", clusterID)
	}
	if len(clusters) == 0 {
		return fmt.Errorf("vast provisioner: enabled with no clusters")
	}
	interval := 5 * time.Minute
	if s := setting(connectorConfig.Config, "reconcile_interval"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil || d <= 0 {
			return fmt.Errorf("vast provisioner: reconcile_interval must be a positive duration, such as \"5m\", not %q", s)
		}
		interval = d
	}
	subscriber := subscribers.NewStorageSubscriber(eventBus, coreService, clusters)
	subscriber.RegisterSubscribers(Type)
	wg.Add(1)
	go func() {
		defer wg.Done()
		subscriber.StartReconciler(ctx, interval)
	}()
	return nil
}

func setting(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	return s
}

// pullFolders moves the folders this connector tracked before storages moved to core, once.
const pullFolders = `DO $$ BEGIN
PERFORM pg_advisory_xact_lock(hashtext('vast_folders'));
IF to_regclass('vast_folders') IS NOT NULL THEN
	INSERT INTO storages (id, compute_cluster_id, type, path, username, allocation_id, allocation_name, start_time, end_time,
		owner, "group", mode, quota_bytes, quota_until, status)
	SELECT id, compute_cluster_id,
		CASE WHEN owner_kind = 'allocation' AND path LIKE '%/shared' THEN 'project_shared'
			WHEN owner_kind = 'allocation' THEN 'project' WHEN allocation_id IS NULL THEN 'user' ELSE 'project_user' END,
		path, CASE owner_kind WHEN 'cluster_user' THEN owner_name ELSE '' END, allocation_id,
		allocation_name, start_time, end_time, split_part(unix_owner, ':', 1), split_part(unix_owner, ':', 2), '',
		COALESCE(quota_override, 0), quota_until, status
	FROM vast_folders WHERE compute_cluster_id IN (SELECT id FROM compute_clusters)
	ON CONFLICT (compute_cluster_id, path) DO NOTHING;
	DROP TABLE IF EXISTS vast_folders, schema_migrations_vast;
END IF;
END $$`
