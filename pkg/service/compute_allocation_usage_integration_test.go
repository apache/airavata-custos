//go:build integration

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

package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

// newClusterAllocation creates a cluster and an allocation on it, and returns both ids.
func newClusterAllocation(t *testing.T, svc *Service, piID string) (clusterID, allocationID string) {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "usage-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	project, err := svc.CreateProject(ctx(), &models.Project{Title: "usage-" + uuid.NewString()[:8], ProjectPIID: piID})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		ProjectID:        project.ID,
		Name:             "usage-" + uuid.NewString()[:8],
		ComputeClusterID: cluster.ID,
		InitialSUAmount:  1000,
		StartTime:        time.Now().UTC(),
		EndTime:          time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	return cluster.ID, alloc.ID
}

// Make sure the latest usage time only looks at the given cluster's allocations, and is nil for a cluster with none.
func TestLatestUsageTimeForCluster_OnlyThatCluster(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("usage-%s@example.edu", uuid.NewString()))

	clusterA, allocA := newClusterAllocation(t, svc, pi)
	clusterB, allocB := newClusterAllocation(t, svc, pi)
	clusterEmpty, _ := newClusterAllocation(t, svc, pi)

	older := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	newer := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	for _, u := range []*models.ComputeAllocationUsage{
		{ComputeAllocationID: allocA, JobID: "1", CalculatedTime: older},
		{ComputeAllocationID: allocB, JobID: "2", CalculatedTime: newer},
	} {
		if _, err := svc.CreateComputeAllocationUsage(ctx(), u); err != nil {
			t.Fatalf("create usage: %v", err)
		}
	}

	got, err := svc.LatestUsageTimeForCluster(ctx(), clusterA)
	if err != nil {
		t.Fatalf("latest for cluster A: %v", err)
	}
	if got == nil || !got.Equal(older) {
		t.Errorf("cluster A: got %v, want %v (cluster B's newer usage must not count)", got, older)
	}
	if got, err := svc.LatestUsageTimeForCluster(ctx(), clusterB); err != nil || got == nil || !got.Equal(newer) {
		t.Errorf("cluster B: got %v, %v, want %v", got, err, newer)
	}
	if got, err := svc.LatestUsageTimeForCluster(ctx(), clusterEmpty); err != nil || got != nil {
		t.Errorf("cluster with no usage: got %v, %v, want nil", got, err)
	}
}
