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
	"errors"

	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

func (s *ClusterUserSubscriber) handleComputeAllocationChanged(ctx context.Context, a models.ComputeAllocation) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.compute_allocation_changed")
	defer span.End()
	return s.syncAllocation(ctx, a.ID)
}

func (s *ClusterUserSubscriber) handleMembershipChanged(ctx context.Context, m models.ComputeAllocationMembership) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.membership_changed")
	defer span.End()
	return s.syncAllocation(ctx, m.ComputeAllocationID)
}

func (s *ClusterUserSubscriber) handleComputeAllocationDeleted(ctx context.Context, a models.ComputeAllocation) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.compute_allocation_deleted")
	defer span.End()
	if a.ComputeClusterID != s.custosClusterID {
		return nil
	}
	return s.ops.RemoveAllocationGroup(ctx, &a)
}

// syncAllocation reloads the allocation, since a stale payload could undo a
// newer change, and ensures its group.
func (s *ClusterUserSubscriber) syncAllocation(ctx context.Context, allocationID string) error {
	a, err := s.core.GetComputeAllocation(ctx, allocationID)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil || a.ComputeClusterID != s.custosClusterID {
		return err
	}
	return s.ops.EnsureAllocationGroup(ctx, a)
}

func (s *ClusterUserSubscriber) syncUserAllocations(ctx context.Context, userID string) error {
	memberships, err := s.core.ListAllocationsForUser(ctx, userID)
	if err != nil {
		return err
	}
	var errs []error
	for _, m := range memberships {
		errs = append(errs, s.syncAllocation(ctx, m.ComputeAllocationID))
	}
	return errors.Join(errs...)
}
