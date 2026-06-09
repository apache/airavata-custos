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

package subscribers

import (
	"context"
	"log/slog"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/operations"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// ClusterUserSubscriber listens for ComputeClusterUserCreateEvent and drives
// the orchestrator. Events whose ComputeClusterID does not match
// CustosClusterID are dropped.
type ClusterUserSubscriber struct {
	ops             *operations.Orchestrator
	bus             *events.Bus
	core            *service.Service
	custosClusterID string
}

func NewClusterUserSubscriber(c *client.Client, bus *events.Bus, core *service.Service, custosClusterID string) *ClusterUserSubscriber {
	return &ClusterUserSubscriber{
		ops:             operations.New(c, core),
		bus:             bus,
		core:            core,
		custosClusterID: custosClusterID,
	}
}

func (s *ClusterUserSubscriber) RegisterSubscribers() {
	s.bus.Subscribe(events.ComputeClusterUserCreateEvent, s.handleClusterUserCreate)
}

func (s *ClusterUserSubscriber) handleClusterUserCreate(_ events.Event, payload interface{}) {
	cu, ok := payload.(*models.ComputeClusterUser)
	if !ok {
		slog.Error("comanage subscriber: payload is not *ComputeClusterUser", "type", payload)
		return
	}
	if cu.ComputeClusterID != s.custosClusterID {
		return
	}
	ctx := context.Background()
	// TODO: move to a transactional scope. In-process delivery loses events
	// if the process crashes between the core commit and subscriber pickup.
	if err := s.ops.EnsurePOSIXAccount(ctx, cu); err != nil {
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed",
			"compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
	}
}
