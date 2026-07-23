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
	"log/slog"
	"strconv"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const monitorInterval = 30 * time.Second

// defaultPollOverlap is the look-back used when usage_lookback is not set.
// The look-back makes each poll re-scan a short way before where the last one
// stopped. SLURM saves a finished job to its accounting database a little
// after the job ends, so a late-arriving record can land after the poll
// window it belongs to has already passed. The look-back re-covers that gap,
// and a repeat of the same job id replaces its earlier row. Its right value
// depends on the cluster's slurmdbd commit lag, so it is configurable.
const defaultPollOverlap = 15 * time.Minute

type jobLister interface {
	ListJobs(filter client.JobFilter) ([]client.JobInfo, error)
}

type SlurmMonitor struct {
	slurmClient     jobLister
	eventBus        *events.Bus
	coreService     service.CoreService
	clusterId       string
	pollOverlap     time.Duration
	lastMonitorTime int64
}

func NewSlurmMonitor(slurmClient *client.Client, eventBus *events.Bus, coreService service.CoreService, clusterId string, pollOverlap time.Duration) *SlurmMonitor {
	if pollOverlap <= 0 {
		pollOverlap = defaultPollOverlap
	}
	return &SlurmMonitor{
		slurmClient:     slurmClient,
		eventBus:        eventBus,
		coreService:     coreService,
		clusterId:       clusterId,
		pollOverlap:     pollOverlap,
		lastMonitorTime: 1, // initialize to 1 to avoid issues with zero value
	}
}

func (m *SlurmMonitor) StartMonitor(ctx context.Context) {
	ticker := time.NewTicker(monitorInterval)
	defer ticker.Stop()

	slog.Info("Starting SLURM usage monitor", "interval", monitorInterval)
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping SLURM usage monitor", "reason", ctx.Err())
			return
		case <-ticker.C:
			m.poll()
		}
	}
}

func (m *SlurmMonitor) poll() {
	slog.Debug("polling SLURM usage")
	context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cluster, err := m.coreService.GetComputeCluster(context, m.clusterId)
	if err != nil {
		slog.Error("failed to get compute cluster", "error", err)
		return
	}

	allocations, err := m.coreService.ListComputeAllocationsByCluster(context, cluster.ID)
	if err != nil {
		slog.Error("failed to list compute allocations", "error", err)
		return
	}

	windowStart := m.lastMonitorTime - int64(m.pollOverlap.Seconds())
	if windowStart < 1 {
		windowStart = 1
	}
	jobFilter := client.JobFilter{
		StartTime: windowStart,
		EndTime:   time.Now().Unix(),
	}

	jobs, err := m.slurmClient.ListJobs(jobFilter)
	if err != nil {
		slog.Error("failed to list SLURM jobs", "error", err)
		return
	}
	m.lastMonitorTime = jobFilter.EndTime

	for _, job := range jobs {
		m.recordJob(context, job, cluster, allocations)
	}

	slog.Info("successfully polled SLURM usage", "num_allocations", len(allocations), "num_jobs", len(jobs))

}

// recordJob writes one usage row for a matched job; returning skips only this
// job so one bad lookup cannot abort the rest of the poll cycle.
func (m *SlurmMonitor) recordJob(ctx context.Context, job client.JobInfo, cluster *models.ComputeCluster, allocations []models.ComputeAllocation) {
	slog.Info("Job object", "job", job)
	targetAccount := job.Account
	for _, alloc := range allocations {
		if alloc.Name != targetAccount {
			continue
		}
		slog.Info("found matching compute allocation for SLURM job", "job_id", job.JobID, "allocation_id", alloc.ID)

		user, err := m.coreService.GetComputeClusterUserByClusterAndLocalUsername(ctx, cluster.ID, job.User)
		if err != nil {
			if err == service.ErrNotFound {
				slog.Warn("compute cluster user not found for SLURM job, skipping usage recording", "local_username", job.User, "cluster_id", cluster.ID)
			} else {
				slog.Error("failed to get compute cluster user", "error", err)
			}
			return
		}

		resource, err := m.coreService.GetComputeAllocationResourceByNameAndCluster(ctx, job.Partition, cluster.ID)
		if err != nil {
			if err == service.ErrNotFound {
				slog.Warn("compute allocation resource not found for SLURM job, skipping usage recording", "resource_name", job.Partition, "cluster_id", cluster.ID)
			} else {
				slog.Error("failed to get compute allocation resource", "error", err)
			}
			return
		}

		jobId := strconv.FormatInt(job.JobID, 10)
		existing, err := m.coreService.GetComputeAllocationUsageByComputeAllocationIDAndJobID(ctx, alloc.ID, jobId)
		if err != nil && err != service.ErrNotFound {
			slog.Error("failed to check for existing compute allocation usage", "error", err)
			return
		}

		jobDurationSec := job.Time.End - job.Time.Start
		if jobDurationSec <= 0 {
			slog.Warn("SLURM job has non-positive duration, skipping usage recording", "job_id", job.JobID, "duration_seconds", jobDurationSec)
			return
		}

		tresType := resource.ResourceType

		resourceAmount := int64(0)
		for _, tres := range job.Tres.Allocated {
			// Example tres entry Allocated:[{Type:cpu Name: Count:1} {Type:mem Name: Count:8000} {Type:energy Name: Count:-2} {Type:node Name: Count:1} {Type:billing Name: Count:1}]
			if tres.Type == tresType {
				resourceAmount = tres.Count
			}
		}

		// tres.Count already is per-node amount x node count (the whole-job total),
		// so raw = tres.Count x hours.
		calculatedRawAmount := float64(resourceAmount) * float64(jobDurationSec) / 3600

		rate, err := m.coreService.GetEffectiveRateForResource(ctx, resource.ID, time.Unix(job.Time.End, 0))
		if err != nil {
			if err == service.ErrNotFound {
				slog.Warn("no rate covers the job end time, skipping usage recording", "job_id", job.JobID, "resource_id", resource.ID)
			} else {
				slog.Error("failed to get effective rate for resource", "error", err, "job_id", job.JobID, "resource_id", resource.ID)
			}
			return
		}

		usageModel := &models.ComputeAllocationUsage{
			ComputeAllocationID:         alloc.ID,
			UsedRawAmount:               calculatedRawAmount,
			UsedSUAmount:                calculatedRawAmount * rate.Rate,
			CalculatedTime:              time.Now(),
			UserID:                      user.UserID,
			JobID:                       jobId,
			ComputeAllocationResourceID: resource.ID,
		}

		if existing != nil {
			m.coreService.DeleteComputeAllocationUsage(ctx, existing.ID)
			slog.Info("deleted existing compute allocation usage for SLURM job", "job_id", job.JobID, "existing_usage_id", existing.ID)
		}
		m.coreService.CreateComputeAllocationUsage(ctx, usageModel)
		return
	}
}
