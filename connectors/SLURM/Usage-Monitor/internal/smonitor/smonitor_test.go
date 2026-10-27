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
	"errors"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type fakeJobLister struct {
	jobs    []client.JobInfo
	err     error              // returned by every query when set
	filters []client.JobFilter // every query it was asked for, in order
}

func (f *fakeJobLister) ListJobs(filter client.JobFilter) ([]client.JobInfo, error) {
	f.filters = append(f.filters, filter)
	return f.jobs, f.err
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
				// billing differs from every other count so a wrong read
				// cannot pass by coincidence.
				{Type: "cpu", Count: 2},
				{Type: "node", Count: 1},
				{Type: "billing", Count: 5},
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
		lastMonitorTime: time.Now().Unix(),
	}
}

// maxQuerySeconds is the longest time range one job query may cover.
const maxQuerySeconds = int64((maxPollStep + defaultPollOverlap) / time.Second)

// Make sure the first poll after a start asks only for the last day of jobs.
func TestPollFirstQueryAfterAStartCoversOneStep(t *testing.T) {
	lister := &fakeJobLister{}
	m := NewSlurmMonitor(nil, nil, newMockCore(nil, nil), "cl-1", 0)
	m.slurmClient = lister

	m.poll(context.Background())

	query := lister.filters[0]
	if seconds := query.EndTime - query.StartTime; seconds > maxQuerySeconds {
		t.Errorf("first query covers %d seconds, want at most %d", seconds, maxQuerySeconds)
	}
}

// Make sure the first poll after a start picks up from the last recorded usage.
func TestPollResumesFromTheLastRecordedUsage(t *testing.T) {
	lastUsage := time.Now().Add(-3 * time.Hour).Truncate(time.Second)
	core := newMockCore(nil, nil)
	core.LatestUsageTimeForClusterFunc = func(ctx context.Context, clusterID string) (*time.Time, error) {
		return &lastUsage, nil
	}
	lister := &fakeJobLister{}
	m := NewSlurmMonitor(nil, nil, core, "cl-1", 0)
	m.slurmClient = lister

	m.resumeFromLastUsage(context.Background())
	m.poll(context.Background())

	if got, want := lister.filters[0].StartTime, lastUsage.Add(-defaultPollOverlap).Unix(); got != want {
		t.Errorf("first query starts at %d, want %d", got, want)
	}
}

// Make sure a cluster with no recorded usage starts one day back.
func TestPollStartsOneDayBackWithoutRecordedUsage(t *testing.T) {
	core := newMockCore(nil, nil)
	core.LatestUsageTimeForClusterFunc = func(ctx context.Context, clusterID string) (*time.Time, error) {
		return nil, nil
	}
	lister := &fakeJobLister{}
	before := time.Now().Add(-startupCatchUp - defaultPollOverlap).Unix()
	m := NewSlurmMonitor(nil, nil, core, "cl-1", 0)
	m.slurmClient = lister

	m.resumeFromLastUsage(context.Background())
	m.poll(context.Background())

	if got := lister.filters[0].StartTime; got < before || got > before+1 {
		t.Errorf("first query starts at %d, want one day and the look-back ago (%d)", got, before)
	}
}

// Make sure a long gap is polled one step at a time, with no time skipped.
func TestPollWalksALongGapInSteps(t *testing.T) {
	lister := &fakeJobLister{}
	m := newTestMonitor(newMockCore(nil, nil))
	m.slurmClient = lister
	m.lastMonitorTime = time.Now().Add(-60 * time.Hour).Unix()

	before := time.Now().Unix()
	for range 3 {
		m.poll(context.Background())
	}

	for i, query := range lister.filters {
		if seconds := query.EndTime - query.StartTime; seconds > maxQuerySeconds {
			t.Errorf("query %d covers %d seconds, want at most %d", i, seconds, maxQuerySeconds)
		}
		if i > 0 && query.StartTime > lister.filters[i-1].EndTime {
			t.Errorf("query %d starts at %d, after query %d ended at %d", i, query.StartTime, i-1, lister.filters[i-1].EndTime)
		}
	}
	if last := lister.filters[2].EndTime; last < before {
		t.Errorf("after three polls the queries reach %d, want now (%d)", last, before)
	}
}

// Make sure a failed query does not move the window forward.
func TestPollKeepsTheWindowAfterAFailedQuery(t *testing.T) {
	lister := &fakeJobLister{err: errors.New("cluster unreachable")}
	m := newTestMonitor(newMockCore(nil, nil))
	m.slurmClient = lister

	m.poll(context.Background())
	m.poll(context.Background())

	if first, second := lister.filters[0].StartTime, lister.filters[1].StartTime; second != first {
		t.Errorf("second query starts at %d, want the same start as the failed one (%d)", second, first)
	}
}

func TestPollRecordsRawAndSUExactly(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 8.0}, map[string]bool{"alice": true})
	newTestMonitor(core, fixtureJob(42, "alice", "debug")).poll(context.Background())

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 usage row, got %d", len(calls))
	}
	u := calls[0].U
	if u.UsedRawAmount != 5.0 {
		t.Errorf("expected used_raw == 5.0 (5 billing x 3600s / 3600), got %v", u.UsedRawAmount)
	}
	if u.UsedSUAmount != 40.0 {
		t.Errorf("expected used_su == 40.0 (5.0 raw x 8.0 rate), got %v", u.UsedSUAmount)
	}
	if u.JobID != "42" || u.ComputeAllocationID != "alloc-1" || u.ComputeAllocationResourceID != "res-debug" {
		t.Errorf("unexpected usage row identity: %+v", u)
	}
	// Attribution must carry the portal user, not the cluster-user mapping row.
	if u.UserID != "user-alice" {
		t.Errorf("expected user_id user-alice, got %q", u.UserID)
	}
}

func TestPollMultiNodeDoesNotOvercount(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 1.0}, map[string]bool{"alice": true})
	job := client.JobInfo{
		JobID:     43,
		Account:   "acct",
		User:      "alice",
		Partition: "debug",
		Time:      client.JobTime{Start: 1000, End: 4600}, // 3600 seconds
		Tres: client.JobTresInfo{
			Allocated: []client.TRES{
				{Type: "cpu", Count: 8},
				{Type: "node", Count: 2},
				{Type: "billing", Count: 12}, // total across both nodes
			},
		},
	}
	newTestMonitor(core, job).poll(context.Background())

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 usage row, got %d", len(calls))
	}
	if got := calls[0].U.UsedRawAmount; got != 12.0 {
		t.Errorf("expected used_raw == 12.0 (12 billing x 3600s / 3600, NOT x 2 nodes), got %v", got)
	}
}

// Charging the resource the partition is named for would bill 1 here, but the
// job holds a whole node worth 4.
func TestPollUsesBillingTresNotThePartitionResourceType(t *testing.T) {
	core := newMockCore(map[string]float64{"res-gpu": 2.0}, map[string]bool{"alice": true})
	core.GetComputeAllocationResourceByNameAndClusterFunc = func(ctx context.Context, name, clusterID string) (*models.ComputeAllocationResource, error) {
		return &models.ComputeAllocationResource{ID: "res-gpu", Name: name, ResourceType: "gres/gpu"}, nil
	}
	job := client.JobInfo{
		JobID:     44,
		Account:   "acct",
		User:      "alice",
		Partition: "gpu",
		Time:      client.JobTime{Start: 1000, End: 4600}, // 3600 seconds
		Tres: client.JobTresInfo{
			Allocated: []client.TRES{
				{Type: "cpu", Count: 128}, // every core on the node
				{Type: "gres", Name: "gpu", Count: 1},
				{Type: "node", Count: 1},
				{Type: "billing", Count: 4},
			},
		},
	}
	newTestMonitor(core, job).poll(context.Background())

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 usage row, got %d", len(calls))
	}
	if got := calls[0].U.UsedRawAmount; got != 4.0 {
		t.Errorf("expected used_raw == 4.0 (4 billing x 3600s / 3600), got %v", got)
	}
	if got := calls[0].U.UsedSUAmount; got != 8.0 {
		t.Errorf("expected used_su == 8.0 (4.0 raw x 2.0 rate), got %v", got)
	}
}

func TestPollSkipsJobWithoutBillingTres(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 8.0}, map[string]bool{"alice": true})
	job := fixtureJob(45, "alice", "debug")
	job.Tres.Allocated = []client.TRES{
		{Type: "cpu", Count: 2},
		{Type: "node", Count: 1},
	}
	newTestMonitor(core, job, fixtureJob(46, "alice", "debug")).poll(context.Background())

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected exactly 1 usage row (job without billing skipped), got %d", len(calls))
	}
	if calls[0].U.JobID != "46" {
		t.Errorf("expected job 46 to be recorded after skipping job 45, got job %s", calls[0].U.JobID)
	}
}

func TestPollSkipsJobWithoutEffectiveRate(t *testing.T) {
	core := newMockCore(map[string]float64{"res-debug": 8.0}, map[string]bool{"alice": true})
	newTestMonitor(core,
		fixtureJob(1, "alice", "unrated"), // res-unrated has no rate row
		fixtureJob(2, "alice", "debug"),
	).poll(context.Background())

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
	).poll(context.Background())

	calls := core.CreateComputeAllocationUsageCalls()
	if len(calls) != 1 {
		t.Fatalf("expected exactly 1 usage row (unknown-user job skipped), got %d", len(calls))
	}
	if calls[0].U.JobID != "2" {
		t.Errorf("expected job 2 to be recorded after skipping job 1, got job %s", calls[0].U.JobID)
	}
}
