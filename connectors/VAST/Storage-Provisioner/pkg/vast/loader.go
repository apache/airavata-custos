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

// Package vast is the VAST Storage-Provisioner entry point. Wired from
// internal/connectors/loader.go.
package vast

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/VAST/Storage-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/VAST/Storage-Provisioner/internal/subscribers"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

// Type is the connector type in the config file and the subscriber name on the bus.
const Type = "vast-storage-provisioner"

func LoadConnector(ctx context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, _ *identity.Router, connectorConfig *config.ConnectorConfig) error {
	vms, _ := connectorConfig.GetNestedConfig("vms")
	vmsURL, _ := vms["url"].(string)
	username, _ := vms["username"].(string)
	password, _ := vms["password"].(string)
	tenantID, _ := vms["tenant_id"].(int)
	clusterID, _ := connectorConfig.GetStringField("custos_cluster_id")
	storage, _ := connectorConfig.Config["storage"].([]interface{})
	if vmsURL == "" || username == "" || password == "" || tenantID == 0 || clusterID == "" || len(storage) == 0 {
		slog.Info("vast provisioner: required config not set; skipping")
		return nil
	}
	mounts, err := parseMounts(storage)
	if err != nil {
		return err
	}
	subscriber := subscribers.NewStorageSubscriber(client.New(vmsURL, username, password, tenantID), eventBus, coreService, clusterID, mounts)
	subscriber.RegisterSubscribers(Type)
	slog.Info("vast provisioner: subscriber registered", "vms", vmsURL, "cluster_id", clusterID)
	wg.Add(1)
	go func() {
		defer wg.Done()
		subscriber.StartReconciler(ctx)
	}()
	return nil
}

func parseMounts(storage []interface{}) ([]subscribers.Mount, error) {
	var mounts []subscribers.Mount
	for i, e := range storage {
		m, _ := e.(map[string]interface{})
		str := func(k string) string { v, _ := m[k].(string); return v }
		hardLimit, _ := m["hard_limit"].(int)
		hardLimitInodes, _ := m["hard_limit_inodes"].(int)
		mode, err := strconv.ParseInt(str("mode"), 8, 32)
		if err != nil {
			return nil, fmt.Errorf("vast provisioner: storage[%d].mode must be octal, such as \"0700\": %w", i, err)
		}
		mounts = append(mounts, subscribers.Mount{Path: str("path"), Owner: str("owner"), Group: str("group"),
			Mode: int(mode), HardLimit: int64(hardLimit), HardLimitInodes: int64(hardLimitInodes)})
	}
	return mounts, nil
}
