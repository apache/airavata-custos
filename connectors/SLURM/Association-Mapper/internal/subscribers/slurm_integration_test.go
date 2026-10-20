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
	"cmp"
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// TestSlurmLifecycle runs against a slurmrestd such as dev-ops/local-slurm, named by TEST_SLURM_*.
func TestSlurmLifecycle(t *testing.T) {
	if !client.IsLocalSlurmConfigAvailable() {
		t.Skip("TEST_SLURM_* is not set")
	}
	ctx, cluster := context.Background(), cmp.Or(os.Getenv("TEST_SLURM_CLUSTER"), "artisan")
	slurm := client.New(os.Getenv("TEST_SLURM_API"), os.Getenv("TEST_SLURM_USER"), os.Getenv("TEST_SLURM_TOKEN"), os.Getenv("TEST_SLURM_API_VERSION"))
	t.Cleanup(func() {
		_ = slurm.DeleteUser("testuser")
		_ = slurm.DeleteAccount("test-alloc")
		_ = slurm.DeleteAccount("new-name")
	})
	sub := func(o mockOpts) *AssociationSubscriber {
		core := coreMock(o)
		core.GetComputeClusterFunc = func(_ context.Context, id string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: id, Name: cluster}, nil
		}
		return NewAssociationSubscriber(slurm, nil, core, fakeMappings(o.mappings), 0, time.Millisecond)
	}
	check := func(step string, want ...string) {
		got, err := slurm.ListAssociations(client.AssocFilter{Cluster: cluster, User: "testuser"})
		run(t, err)
		var accounts []string
		for _, a := range got {
			accounts = append(accounts, a.Account)
		}
		if slices.Sort(accounts); !slices.Equal(accounts, want) {
			t.Fatalf("after %s testuser holds %v, want %v", step, accounts, want)
		}
	}
	active := mockOpts{provisionedAt: ago(time.Hour), memberships: member, mappings: cpu8}
	left := mockOpts{provisionedAt: ago(time.Hour), mappings: cpu8}

	run(t, sub(active).SubscribeToComputeAllocationChange(ctx, models.ComputeAllocation{ID: "alloc-1"}))
	check("allocation change", "none", "test-alloc")
	run(t, sub(left).SubscribeToComputeAllocationMembershipChange(ctx, testMembership()))
	check("membership removal", "none")
	run(t, sub(active).SubscribeToComputeAllocationMembershipChange(ctx, testMembership()))
	sub(mockOpts{allocationName: "new-name", provisionedAt: ago(time.Hour), memberships: member, mappings: cpu8}).reconcile(ctx)
	check("rename", "new-name", "none")
	run(t, sub(mockOpts{}).SubscribeToComputeClusterUserChange(ctx, cu))
	check("account deletion")
}
