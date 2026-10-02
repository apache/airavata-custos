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
// once COmanage has provisioned the account, the allocation's group, or both, since VAST resolves owners by name
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

// Mount is a directory created for each account, each provisioned allocation's
// group, or each member of one. Path, Owner and Group are templates over
// {user}, the account's username, and {allocation}, the allocation's group;
// the ones it uses decide which. Its quota, named after the path, is created
// when HardLimit is set; a directory without one counts against its parent's.
type Mount struct {
	Path, Owner, Group         string
	Mode                       int
	HardLimit, HardLimitInodes int64
}

func (m Mount) uses(placeholder string) bool {
	return strings.Contains(m.Path+m.Owner+m.Group, placeholder)
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

func (s *StorageSubscriber) RegisterSubscribers(subscriber string) {
	s.bus.SubscribeComputeClusterUserUpdated(subscriber, s.ensureUser)
	s.bus.SubscribeComputeAllocationCreated(subscriber, s.ensureAllocation)
	s.bus.SubscribeComputeAllocationUpdated(subscriber, s.ensureAllocation)
	byMembership := func(ctx context.Context, m models.ComputeAllocationMembership) error {
		return s.ensureAllocationByID(ctx, m.ComputeAllocationID)
	}
	s.bus.SubscribeComputeAllocationMembershipCreated(subscriber, byMembership)
	s.bus.SubscribeComputeAllocationMembershipUpdated(subscriber, byMembership)
}

// ensureUser needs no approval check: COmanage provisions approved accounts only.
// After the account's mounts it ensures each of the account's allocations.
func (s *StorageSubscriber) ensureUser(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.custosClusterID || cu.ProvisionedAt == nil {
		return nil
	}
	if err := s.ensure(ctx, cu.LocalUsername, ""); err != nil {
		return err
	}
	memberships, err := s.core.ListAllocationsForUser(ctx, cu.UserID)
	if err != nil {
		return err
	}
	var errs []error
	for _, m := range memberships {
		errs = append(errs, s.ensureAllocationByID(ctx, m.ComputeAllocationID))
	}
	return errors.Join(errs...)
}

func (s *StorageSubscriber) ensureAllocationByID(ctx context.Context, id string) error {
	a, err := s.core.GetComputeAllocation(ctx, id)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.ensureAllocation(ctx, *a)
}

// ensureAllocation creates an active allocation's mounts, then those of each
// member its POSIX group holds.
func (s *StorageSubscriber) ensureAllocation(ctx context.Context, a models.ComputeAllocation) error {
	if a.ComputeClusterID != s.custosClusterID || a.Status != models.ACTIVE || a.PosixGroup == nil {
		return nil
	}
	if err := s.ensure(ctx, "", *a.PosixGroup); err != nil {
		return err
	}
	members, err := s.core.ListMembersForAllocation(ctx, a.ID)
	if err != nil {
		return err
	}
	for _, m := range members {
		if m.MembershipStatus != models.ACTIVE || m.ProvisionedAt == nil {
			continue
		}
		if err := s.ensure(ctx, m.LocalUsername, *a.PosixGroup); err != nil {
			return err
		}
	}
	return nil
}

// ensure creates the mounts that use exactly the placeholders given.
func (s *StorageSubscriber) ensure(ctx context.Context, user, group string) error {
	r := strings.NewReplacer("{user}", user, "{allocation}", group)
	for _, m := range s.mounts {
		if m.uses("{user}") != (user != "") || m.uses("{allocation}") != (group != "") {
			continue
		}
		path := r.Replace(m.Path)
		if err := s.vms.CreateFolder(ctx, path, r.Replace(m.Owner), r.Replace(m.Group), m.Mode); err != nil {
			return err
		}
		if m.HardLimit == 0 {
			continue
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
		if cu.ProvisionedAt != nil {
			errs = append(errs, s.ensure(ctx, cu.LocalUsername, ""))
		}
	}
	allocs, err := s.core.ListComputeAllocationsByCluster(ctx, s.custosClusterID)
	errs = append(errs, err)
	for _, a := range allocs {
		errs = append(errs, s.ensureAllocation(ctx, a))
	}
	if err := errors.Join(errs...); err != nil {
		slog.Error("VAST reconciler", "err", err)
	}
}
