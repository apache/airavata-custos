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
	"cmp"
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/Storage/VAST-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Mount is a directory created for each account, each provisioned allocation's
// group, or each member of one. Path, Owner and Group are templates over
// {user}, the account's username, and {allocation}, the allocation's group;
// the ones it uses decide which. Its quota, named after the path, is created
// when HardLimit is set; a directory without one counts against its parent's.
// When View is set, the directory also becomes a view under the cluster's view policy.
type Mount struct {
	Path, Owner, Group         string
	Mode                       int
	HardLimit, HardLimitInodes int64
	View                       bool
}

func (m Mount) uses(placeholder string) bool {
	return strings.Contains(m.Path+m.Owner+m.Group, placeholder)
}

type Cluster struct {
	ID            string
	VMS           *client.Client
	Mounts        []Mount
	ViewPolicyID  int
	ViewProtocols []string
	Folders       Registry
}

// Registry records the folders ensure makes and marks a deleted owner's for an admin to review.
type Registry interface {
	Record(ctx context.Context, clusterID, path, ownerKind, ownerName, unixOwner string, a *models.ComputeAllocation) (int64, error)
	MarkPending(ctx context.Context, clusterID, ownerKind, ownerName string) error
}

type StorageSubscriber struct {
	bus      *events.Bus
	core     service.CoreService
	clusters map[string]Cluster
}

func NewStorageSubscriber(bus *events.Bus, core service.CoreService, clusters map[string]Cluster) *StorageSubscriber {
	return &StorageSubscriber{bus: bus, core: core, clusters: clusters}
}

// Each event creates only its own user's or allocation's mounts, so a failure
// names the right user and can be retried alone. The reconciler does the rest.
func (s *StorageSubscriber) RegisterSubscribers(subscriber string) {
	s.bus.SubscribeComputeClusterUserProvisioned(subscriber, s.ensureUser)
	s.bus.SubscribeComputeAllocationGroupProvisioned(subscriber, s.ensureAllocation)
	s.bus.SubscribeComputeAllocationMembershipCreated(subscriber, s.ensureMember)
	s.bus.SubscribeComputeAllocationMembershipUpdated(subscriber, s.ensureMember)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.markUser)
	s.bus.SubscribeComputeAllocationDeleted(subscriber, s.markAllocation)
}

func (s *StorageSubscriber) markUser(ctx context.Context, cu models.ComputeClusterUser) error {
	return s.mark(ctx, cu.ComputeClusterID, "cluster_user", cu.LocalUsername)
}

func (s *StorageSubscriber) markAllocation(ctx context.Context, a models.ComputeAllocation) error {
	if a.PosixGroup == nil {
		return nil
	}
	return s.mark(ctx, a.ComputeClusterID, "allocation", *a.PosixGroup)
}

func (s *StorageSubscriber) mark(ctx context.Context, clusterID, ownerKind, ownerName string) error {
	if c, ok := s.clusters[clusterID]; ok {
		return c.Folders.MarkPending(ctx, clusterID, ownerKind, ownerName)
	}
	return nil
}

// ensureUser needs no approval check: COmanage provisions approved accounts only.
// After the account's own mounts, it creates the account's mount in each of its allocations.
func (s *StorageSubscriber) ensureUser(ctx context.Context, cu models.ComputeClusterUser) error {
	c, ok := s.clusters[cu.ComputeClusterID]
	if !ok || cu.ProvisionedAt == nil {
		return nil
	}
	if err := s.ensure(ctx, c, cu.LocalUsername, nil); err != nil {
		return err
	}
	memberships, err := s.core.ListAllocationsForUser(ctx, cu.UserID)
	if err != nil {
		return err
	}
	for _, m := range memberships {
		if m.MembershipStatus != models.ACTIVE {
			continue
		}
		a, _, err := s.activeAllocation(ctx, m.ComputeAllocationID)
		if err != nil {
			return err
		}
		if a == nil || a.ComputeClusterID != cu.ComputeClusterID {
			continue
		}
		if err := s.ensureMemberMounts(ctx, c, cu.LocalUsername, a); err != nil {
			return err
		}
	}
	return nil
}

// ensureMember creates the member's mount in the allocation. A member without a
// provisioned account is skipped, its provision event creates the mount later.
func (s *StorageSubscriber) ensureMember(ctx context.Context, m models.ComputeAllocationMembership) error {
	if m.MembershipStatus != models.ACTIVE {
		return nil
	}
	a, c, err := s.activeAllocation(ctx, m.ComputeAllocationID)
	if err != nil || a == nil {
		return err
	}
	members, err := s.core.ListMembersForAllocation(ctx, a.ID)
	if err != nil {
		return err
	}
	for _, mw := range members {
		if mw.UserID == m.UserID && mw.ProvisionedAt != nil {
			return s.ensureMemberMounts(ctx, c, mw.LocalUsername, a)
		}
	}
	return nil
}

// cluster returns the allocation's cluster when the allocation is active, its group is on the cluster, and the cluster is configured.
func (s *StorageSubscriber) cluster(a models.ComputeAllocation) (Cluster, bool) {
	c, ok := s.clusters[a.ComputeClusterID]
	return c, ok && a.Status == models.ACTIVE && a.PosixGroup != nil && a.GroupProvisionedAt != nil
}

// activeAllocation loads the allocation and its cluster, nil when there is nothing to create for it.
func (s *StorageSubscriber) activeAllocation(ctx context.Context, id string) (*models.ComputeAllocation, Cluster, error) {
	a, err := s.core.GetComputeAllocation(ctx, id)
	if errors.Is(err, service.ErrNotFound) {
		return nil, Cluster{}, nil
	}
	if err != nil {
		return nil, Cluster{}, err
	}
	c, ok := s.cluster(*a)
	if !ok {
		return nil, Cluster{}, nil
	}
	return a, c, nil
}

// ensureMemberMounts creates the group's mounts, then the user's mount inside them.
func (s *StorageSubscriber) ensureMemberMounts(ctx context.Context, c Cluster, user string, a *models.ComputeAllocation) error {
	if err := s.ensure(ctx, c, "", a); err != nil {
		return err
	}
	return s.ensure(ctx, c, user, a)
}

// ensureAllocation creates an allocation's mounts once its group is on the cluster, then each provisioned member's.
func (s *StorageSubscriber) ensureAllocation(ctx context.Context, a models.ComputeAllocation) error {
	c, ok := s.cluster(a)
	if !ok {
		return nil
	}
	if err := s.ensure(ctx, c, "", &a); err != nil {
		return err
	}
	members, err := s.core.ListMembersForAllocation(ctx, a.ID)
	if err != nil {
		return err
	}
	var errs []error
	for _, m := range members {
		if m.MembershipStatus != models.ACTIVE || m.ProvisionedAt == nil {
			continue
		}
		errs = append(errs, s.ensure(ctx, c, m.LocalUsername, &a))
	}
	return errors.Join(errs...)
}

// ensure creates the mounts that use exactly the placeholders given; a is nil for the user's own mounts.
func (s *StorageSubscriber) ensure(ctx context.Context, c Cluster, user string, a *models.ComputeAllocation) error {
	group := ""
	if a != nil {
		group = *a.PosixGroup
	}
	r := strings.NewReplacer("{user}", user, "{allocation}", group)
	kind, owner := "allocation", group
	if user != "" {
		kind, owner = "cluster_user", user
	}
	for _, m := range c.Mounts {
		if m.uses("{user}") != (user != "") || m.uses("{allocation}") != (group != "") {
			continue
		}
		path := r.Replace(m.Path)
		if err := c.VMS.CreateFolder(ctx, path, r.Replace(m.Owner), r.Replace(m.Group), m.Mode); err != nil {
			return err
		}
		override, err := c.Folders.Record(ctx, c.ID, path, kind, owner, r.Replace(m.Owner)+":"+r.Replace(m.Group), a)
		if err != nil {
			return err
		}
		if m.View {
			if err := c.VMS.CreateView(ctx, path, c.ViewPolicyID, c.ViewProtocols); err != nil {
				return err
			}
		}
		if m.HardLimit == 0 {
			continue
		}
		if err := c.VMS.CreateQuota(ctx, path, cmp.Or(override, m.HardLimit), m.HardLimitInodes); err != nil {
			return err
		}
	}
	return nil
}

func (s *StorageSubscriber) StartReconciler(ctx context.Context, interval time.Duration) {
	for {
		for id, c := range s.clusters {
			s.reconcile(ctx, id, c)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

func (s *StorageSubscriber) reconcile(ctx context.Context, id string, c Cluster) {
	ctx, span := tracing.Start(ctx, "vast.reconcile")
	defer span.End()
	users, err := s.core.ListComputeClusterUsersByCluster(ctx, id)
	errs := []error{err}
	for _, cu := range users {
		if cu.ProvisionedAt != nil {
			errs = append(errs, s.ensure(ctx, c, cu.LocalUsername, nil))
		}
	}
	allocs, err := s.core.ListComputeAllocationsByCluster(ctx, id)
	errs = append(errs, err)
	for _, a := range allocs {
		errs = append(errs, s.ensureAllocation(ctx, a))
	}
	if err := errors.Join(errs...); err != nil {
		slog.ErrorContext(ctx, "VAST reconciler", "cluster_id", id, "err", err)
	}
}
