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
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/subscribers"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "vast-storage-provisioner"

func LoadConnector(ctx context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, _ *identity.Router, connectorConfig *config.ConnectorConfig) error {
	entries, _ := connectorConfig.Config["clusters"].([]interface{})
	clusters := map[string]subscribers.Cluster{}
	for _, e := range entries {
		cluster, _ := e.(map[string]interface{})
		vms, _ := cluster["vast"].(map[string]interface{})
		vmsURL, username, password := setting(vms, "url"), setting(vms, "username"), setting(vms, "password")
		tenantID, _ := vms["tenant_id"].(int)
		clusterID := setting(cluster, "id")
		targets, _ := cluster["targets"].([]interface{})
		if vmsURL == "" || username == "" || password == "" || tenantID == 0 || clusterID == "" || len(targets) == 0 {
			return fmt.Errorf("vast provisioner: cluster %q needs id, vast.url, vast.username, vast.password, vast.tenant_id and targets", clusterID)
		}
		mounts, err := parseMounts(targets)
		if err != nil {
			return err
		}
		c := subscribers.Cluster{VMS: client.New(vmsURL, username, password, tenantID), Mounts: mounts}
		// Verify that view policy is set otherwise stop the startup instead of failing every create request.
		if slices.ContainsFunc(mounts, func(m subscribers.Mount) bool { return m.View }) {
			policy := setting(vms, "view_policy")
			for p := range strings.SplitSeq(setting(vms, "view_protocols"), ",") {
				if p = strings.TrimSpace(p); p != "" {
					c.ViewProtocols = append(c.ViewProtocols, p)
				}
			}
			if policy == "" || len(c.ViewProtocols) == 0 {
				return fmt.Errorf("vast provisioner: cluster %s has view targets, so vast.view_policy and vast.view_protocols are required", clusterID)
			}
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
	subscriber := subscribers.NewStorageSubscriber(eventBus, coreService, clusters)
	subscriber.RegisterSubscribers(Type)
	wg.Add(1)
	go func() {
		defer wg.Done()
		subscriber.StartReconciler(ctx)
	}()
	return nil
}

// setting returns "" when the env var is not set, since the config keeps ${VAR} as is.
func setting(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	if strings.Contains(s, "${") {
		return ""
	}
	return s
}

func parseMounts(storage []interface{}) ([]subscribers.Mount, error) {
	var mounts []subscribers.Mount
	for i, e := range storage {
		m, _ := e.(map[string]interface{})
		str := func(k string) string { v, _ := m[k].(string); return v }
		hardLimit, _ := m["hard_limit_size"].(int)
		hardLimitInodes, _ := m["hard_limit_inodes"].(int)
		view, _ := m["view"].(bool)
		mode, err := strconv.ParseInt(str("mode"), 8, 32)
		if err != nil {
			return nil, fmt.Errorf("vast provisioner: targets[%d].mode must be octal, such as \"0700\": %w", i, err)
		}
		mounts = append(mounts, subscribers.Mount{Path: str("path"), Owner: str("owner"), Group: str("group"),
			Mode: int(mode), HardLimit: int64(hardLimit), HardLimitInodes: int64(hardLimitInodes), View: view})
	}
	return mounts, nil
}
