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

// Package subscribers creates the configured directories and quotas on VAST
// once COmanage has provisioned the account, since VAST resolves owners by name
// from the LDAP that COmanage feeds. Mounts are created in config order, so
// parents come first.
package subscribers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/VAST/Storage-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Mount is a directory created for each account. Path, Owner and Group are
// templates over {user}, the account's username; its quota is named after the path.
type Mount struct {
	Path, Owner, Group         string
	Mode                       int
	HardLimit, HardLimitInodes int64
}

type StorageSubscriber struct {
	vms             *client.Client
	bus             *events.Bus
	core            service.CoreService
	custosClusterID string
	mounts          []Mount
}

func NewStorageSubscriber(vms *client.Client, bus *events.Bus, core service.CoreService, custosClusterID string, mounts []Mount) *StorageSubscriber {
	return &StorageSubscriber{vms: vms, bus: bus, core: core, custosClusterID: custosClusterID, mounts: mounts}
}

// RegisterSubscribers waits for the "provisioned" update COmanage publishes, which alone releases the work.
func (s *StorageSubscriber) RegisterSubscribers(subscriber string) {
	s.bus.SubscribeComputeClusterUserUpdated(subscriber, s.ensureUser)
}

// ensureUser needs no approval check: COmanage provisions approved accounts only.
func (s *StorageSubscriber) ensureUser(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.custosClusterID || cu.ProvisionedAt == nil {
		return nil
	}
	r := strings.NewReplacer("{user}", cu.LocalUsername)
	for _, m := range s.mounts {
		path := r.Replace(m.Path)
		if err := s.vms.CreateFolder(ctx, path, r.Replace(m.Owner), r.Replace(m.Group), m.Mode); err != nil {
			return err
		}
		if err := s.vms.CreateQuota(ctx, path, m.HardLimit, m.HardLimitInodes); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorageSubscriber) StartReconciler(ctx context.Context) {
	for {
		s.reconcile(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(24 * time.Hour):
		}
	}
}

func (s *StorageSubscriber) reconcile(ctx context.Context) {
	users, err := s.core.ListComputeClusterUsersByCluster(ctx, s.custosClusterID)
	errs := []error{err}
	for _, cu := range users {
		errs = append(errs, s.ensureUser(ctx, cu))
	}
	if err := errors.Join(errs...); err != nil {
		slog.Error("VAST reconciler", "err", err)
	}
}
