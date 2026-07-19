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

package smonitor

import (
	"context"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type fakeJobLister struct {
	jobs []client.JobInfo
}

func (f *fakeJobLister) ListJobs(filter client.JobFilter) ([]client.JobInfo, error) {
	return f.jobs, nil
}

func fixtureJob(jobID int64, user, partition string) client.JobInfo {
	return client.JobInfo{
		JobID:     jobID,
		Account:   "acct",
		User:      user,
		Partition: partition,
		Time:      client.JobTime{Start: 1000, End: 4600}, // exactly 3600 seconds
		Tres: client.JobTresInfo{
			Allocated: []client.TRES{
				{Type: "cpu", Count: 2},
				{Type: "node", Count: 1},
			},
		},
	}
}

func newMockCore(rates map[string]float64, knownUsers map[string]bool) *service.CoreServiceMock {
	return &service.CoreServiceMock{
		GetComputeClusterFunc: func(ctx context.Context, id string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: "cl-1", Name: "test"}, nil
		},
		ListComputeAllocationsByClusterFunc: func(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error) {
			return []models.ComputeAllocation{{ID: "alloc-1", Name: "acct", ComputeClusterID: clusterID}}, nil
		},
		// The interface takes (clusterID, localUsername); rejecting a wrong
		// cluster id catches swapped-argument calls.
		GetComputeClusterUserByClusterAndLocalUsernameFunc: func(ctx context.Context, clusterID, localUsername string) (*models.ComputeClusterUser, error) {
			if clusterID != "cl-1" || !knownUsers[localUsername] {
				return nil, service.ErrNotFound
			}
			return &models.ComputeClusterUser{ID: "ccu-" + localUsername, UserID: "user-" + localUsername, LocalUsername: localUsername}, nil
		},
		GetComputeAllocationResourceByNameAndClusterFunc: func(ctx context.Context, name, clusterID string) (*models.ComputeAllocationResource, error) {
			return &models.ComputeAllocationResource{ID: "res-" + name, Name: name, ResourceType: "cpu"}, nil
		},
		GetComputeAllocationUsageByComputeAllocationIDAndJobIDFunc: func(ctx context.Context, allocationID, jobID string) (*models.ComputeAllocationUsage, error) {
			return nil, service.ErrNotFound
		},
		GetEffectiveRateForResourceFunc: func(ctx context.Context, resourceID string, at time.Time) (*models.ComputeAllocationResourceRate, error) {
			rate, ok := rates[resourceID]
			if !ok {
				return nil, service.ErrNotFound
			}
			return &models.ComputeAllocationResourceRate{ID: "rate-1", ComputeAllocationResourceID: resourceID, Rate: rate}, nil
		},
		CreateComputeAllocationUsageFunc: func(ctx context.Context, u *models.ComputeAllocationUsage) (*models.ComputeAllocationUsage, error) {
			return u, nil
		},
	}
}

func newTestMonitor(core *service.CoreServiceMock, jobs ...client.JobInfo) *SlurmMonitor {
	return &SlurmMonitor{
		slurmClient:     &fakeJobLister{jobs: jobs},
		coreService:     core,
		clusterId:       "cl-1",
		pollOverlap:     defaultPollOverlap,
		lastMonitorTime: 1,
	}
}

func TestPollRecordsRawAndSUExactly(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 8.0}, map[string]bool{"alice": true})
	newTestMonitor(core, fixtureJob(42, "alice", "debug")).poll()

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 usage row, got %d", len(calls))
	}
	u := calls[0].U
	if u.UsedRawAmount != 2.0 {
		t.Errorf("expected used_raw == 2.0 (2 cpu x 1 node x 3600s / 3600), got %v", u.UsedRawAmount)
	}
	if u.UsedSUAmount != 16.0 {
		t.Errorf("expected used_su == 16.0 (2.0 raw x 8.0 rate), got %v", u.UsedSUAmount)
	}
	if u.JobID != "42" || u.ComputeAllocationID != "alloc-1" || u.ComputeAllocationResourceID != "res-debug" {
		t.Errorf("unexpected usage row identity: %+v", u)
	}
	// Attribution must carry the portal user, not the cluster-user mapping row.
	if u.UserID != "user-alice" {
		t.Errorf("expected user_id user-alice, got %q", u.UserID)
	}
}

func TestPollSkipsJobWithoutEffectiveRate(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 8.0}, map[string]bool{"alice": true})
	newTestMonitor(core,
		fixtureJob(1, "alice", "unrated"), // res-unrated has no rate row
		fixtureJob(2, "alice", "debug"),
	).poll()

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected exactly 1 usage row (unrated job skipped), got %d", len(calls))
	}
	if calls[0].U.JobID != "2" {
		t.Errorf("expected the rated job 2 to be recorded, got job %s", calls[0].U.JobID)
	}
}

func TestPollContinuesPastUnknownClusterUser(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 8.0}, map[string]bool{"alice": true})
	newTestMonitor(core,
		fixtureJob(1, "ghost", "debug"), // unknown cluster user
		fixtureJob(2, "alice", "debug"),
	).poll()

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected exactly 1 usage row (unknown-user job skipped), got %d", len(calls))
	}
	if calls[0].U.JobID != "2" {
		t.Errorf("expected job 2 to be recorded after skipping job 1, got job %s", calls[0].U.JobID)
	}
}
