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
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/tracing"
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

// maxPollStep is the value of the next poll step keeping the gap not too wide.
const maxPollStep = 24 * time.Hour

// lookupTimeout bounds the cluster lookups and, separately, each job's writes,
// so a slow job query or a long catch-up batch cannot use up the budget of the
// jobs after it.
const lookupTimeout = 10 * time.Second

// startupCatchUp is the value how far back the poller scans for jobs when no
// usage has been recorded for the cluster yet.
const startupCatchUp = 24 * time.Hour

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
		lastMonitorTime: time.Now().Add(-startupCatchUp).Unix(),
	}
}

// resumeFromLastUsage starts polling from the end of the last recorded job.
func (m *SlurmMonitor) resumeFromLastUsage(ctx context.Context) {
	last, err := m.coreService.LatestUsageTimeForCluster(ctx, m.clusterId)
	if err != nil {
		slog.Warn("could not read the last recorded usage, starting from the default", "error", err)
		return
	}
	if last != nil {
		m.lastMonitorTime = last.Unix()
	}
}

func (m *SlurmMonitor) StartMonitor(ctx context.Context) {
	m.resumeFromLastUsage(ctx)
	ticker := time.NewTicker(monitorInterval)
	defer ticker.Stop()

	slog.Info("Starting SLURM usage monitor", "interval", monitorInterval, "first_poll_from", time.Unix(m.lastMonitorTime, 0).UTC())
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping SLURM usage monitor", "reason", ctx.Err())
			return
		case <-ticker.C:
			m.poll(ctx)
		}
	}
}

func (m *SlurmMonitor) poll(ctx context.Context) {
	slog.Debug("polling SLURM usage")
	ctx, span := tracing.Start(ctx, "slurm.usage_poll")
	defer span.End()
	lookupCtx, cancel := context.WithTimeout(ctx, lookupTimeout)
	defer cancel()
	cluster, err := m.coreService.GetComputeCluster(lookupCtx, m.clusterId)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get compute cluster", "error", err)
		return
	}

	allocations, err := m.coreService.ListComputeAllocationsByCluster(lookupCtx, cluster.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list compute allocations", "error", err)
		return
	}

	jobFilter := client.JobFilter{
		StartTime: m.lastMonitorTime - int64(m.pollOverlap.Seconds()),
		EndTime:   min(time.Now().Unix(), m.lastMonitorTime+int64(maxPollStep.Seconds())),
	}

	jobs, err := m.slurmClient.ListJobs(jobFilter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list SLURM jobs", "error", err)
		return
	}
	failed := 0
	for _, job := range jobs {
		jobCtx, cancelJob := context.WithTimeout(ctx, lookupTimeout)
		if err := m.recordJob(jobCtx, job, cluster, allocations); err != nil {
			failed++
		}
		cancelJob()
	}
	// Moving on would drop the jobs that failed: the next window only reaches
	// back by the look-back. Polling the same window again is safe because a
	// repeat of a job id replaces its earlier row.
	if failed > 0 {
		slog.WarnContext(ctx, "some SLURM jobs were not recorded, polling the same window again", "failed_jobs", failed, "num_jobs", len(jobs))
		return
	}
	m.lastMonitorTime = jobFilter.EndTime

	slog.InfoContext(ctx, "successfully polled SLURM usage", "num_allocations", len(allocations), "num_jobs", len(jobs))

}

// recordJob writes one usage row for a matched job. A job it skips on purpose
// returns nil. A failed lookup or write returns the error, so the poll can try
// the job again. Either way it touches only this job.
func (m *SlurmMonitor) recordJob(ctx context.Context, job client.JobInfo, cluster *models.ComputeCluster, allocations []models.ComputeAllocation) error {
	slog.InfoContext(ctx, "Job object", "job", job)
	// Accounts are created in lowercase
	for _, alloc := range allocations {
		if !strings.EqualFold(alloc.Name, job.Account) {
			continue
		}
		slog.InfoContext(ctx, "found matching compute allocation for SLURM job", "job_id", job.JobID, "allocation_id", alloc.ID)

		user, err := m.coreService.GetComputeClusterUserByClusterAndLocalUsername(ctx, cluster.ID, job.User)
		if err != nil {
			if err == service.ErrNotFound {
				slog.WarnContext(ctx, "compute cluster user not found for SLURM job, skipping usage recording", "local_username", job.User, "cluster_id", cluster.ID)
				return nil
			}
			slog.ErrorContext(ctx, "failed to get compute cluster user", "error", err)
			return err
		}

		resource, err := m.coreService.GetComputeAllocationResourceByNameAndCluster(ctx, job.Partition, cluster.ID)
		if err != nil {
			if err == service.ErrNotFound {
				slog.WarnContext(ctx, "compute allocation resource not found for SLURM job, skipping usage recording", "resource_name", job.Partition, "cluster_id", cluster.ID)
				return nil
			}
			slog.ErrorContext(ctx, "failed to get compute allocation resource", "error", err)
			return err
		}

		jobId := strconv.FormatInt(job.JobID, 10)
		existing, err := m.coreService.GetComputeAllocationUsageByComputeAllocationIDAndJobID(ctx, alloc.ID, jobId)
		if err != nil && err != service.ErrNotFound {
			slog.ErrorContext(ctx, "failed to check for existing compute allocation usage", "error", err)
			return err
		}

		jobDurationSec := job.Time.End - job.Time.Start
		if jobDurationSec <= 0 {
			slog.WarnContext(ctx, "SLURM job has non-positive duration, skipping usage recording", "job_id", job.JobID, "duration_seconds", jobDurationSec)
			return nil
		}

		// Without TRESBillingWeights and PriorityFlags=MAX_TRES on the
		// partition, billing is just the cpu count and this charge is wrong.
		billingAmount := float64(0)
		for _, tres := range job.Tres.Allocated {
			if tres.Type == "billing" {
				billingAmount = tres.Count
			}
		}
		if billingAmount <= 0 {
			slog.WarnContext(ctx, "SLURM job has no billing charge, skipping usage recording", "job_id", job.JobID, "billing", billingAmount)
			return nil
		}

		// billing is already the whole-job total, so no node multiplier.
		calculatedRawAmount := billingAmount * float64(jobDurationSec) / 3600

		rate, err := m.coreService.GetEffectiveRateForResource(ctx, resource.ID, time.Unix(job.Time.End, 0))
		if err != nil {
			if err == service.ErrNotFound {
				slog.WarnContext(ctx, "no rate covers the job end time, skipping usage recording", "job_id", job.JobID, "resource_id", resource.ID)
				return nil
			}
			slog.ErrorContext(ctx, "failed to get effective rate for resource", "error", err, "job_id", job.JobID, "resource_id", resource.ID)
			return err
		}

		usageModel := &models.ComputeAllocationUsage{
			ComputeAllocationID:         alloc.ID,
			UsedRawAmount:               calculatedRawAmount,
			UsedSUAmount:                calculatedRawAmount * rate.Rate,
			CalculatedTime:              time.Unix(job.Time.End, 0),
			UserID:                      user.UserID,
			JobID:                       jobId,
			ComputeAllocationResourceID: resource.ID,
		}

		if existing != nil {
			if err := m.coreService.DeleteComputeAllocationUsage(ctx, existing.ID); err != nil {
				slog.ErrorContext(ctx, "failed to delete existing compute allocation usage", "error", err, "job_id", job.JobID, "existing_usage_id", existing.ID)
				return err
			}
			slog.InfoContext(ctx, "deleted existing compute allocation usage for SLURM job", "job_id", job.JobID, "existing_usage_id", existing.ID)
		}
		if _, err := m.coreService.CreateComputeAllocationUsage(ctx, usageModel); err != nil {
			slog.ErrorContext(ctx, "failed to record compute allocation usage", "error", err, "job_id", job.JobID)
			// An invalid row fails the same way on every poll, so retrying it
			// would only hold the window back.
			if errors.Is(err, service.ErrInvalidInput) {
				return nil
			}
			return err
		}
		return nil
	}
	return nil
}
