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

type SlurmMonitor struct {
	slurmClient     *client.Client
	eventBus        *events.Bus
	coreService     service.CoreService
	clusterId       string
	lastMonitorTime int64
}

func NewSlurmMonitor(slurmClient *client.Client, eventBus *events.Bus, coreService service.CoreService, clusterId string) *SlurmMonitor {
	return &SlurmMonitor{
		slurmClient:     slurmClient,
		eventBus:        eventBus,
		coreService:     coreService,
		clusterId:       clusterId,
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

	jobFilter := client.JobFilter{
		StartTime: m.lastMonitorTime,
		EndTime:   time.Now().Unix(),
	}

	jobs, err := m.slurmClient.ListJobs(jobFilter)
	if err != nil {
		slog.Error("failed to list SLURM jobs", "error", err)
		return
	}
	m.lastMonitorTime = jobFilter.EndTime

	for _, job := range jobs {
		//slog.Debug("processing SLURM job", "job_id", job.JobID, "job_name", job.Name)
		//m.coreService.GetComputeAllocationResource()
		slog.Info("Job object", "job", job)
		targetAccount := job.Account
		for _, alloc := range allocations {
			if alloc.Name == targetAccount {
				slog.Info("found matching compute allocation for SLURM job", "job_id", job.JobID, "allocation_id", alloc.ID)

				user, err := m.coreService.GetComputeClusterUserByLocalUsernameAndCluster(context, job.User, cluster.ID)
				if err != nil {
					slog.Error("failed to get compute cluster user", "error", err)
					return
				}

				resource, err := m.coreService.GetComputeAllocationResourceByNameAndCluster(context, job.Partition, cluster.ID)

				if err != nil {
					slog.Error("failed to get compute allocation resource", "error", err)
					return
				}

				usageModel := &models.ComputeAllocationUsage{
					ComputeAllocationID:         alloc.ID,
					UsedRawAmount:               (job.Time.End - job.Time.Start),     // This is a simplification, adjust as needed based on how you want to calculate usage
					UsedSUAmount:                (job.Time.End - job.Time.Start) * 1, // Assuming 1 SU per second for simplicity, adjust as needed based on your SU calculation logic
					CalculatedTime:              time.Now(),
					UserID:                      user.ID,
					JobID:                       strconv.FormatInt(job.JobID, 10),
					ComputeAllocationResourceID: resource.ID,
				}
				m.coreService.CreateComputeAllocationUsage(context, usageModel)
				break
			}
		}
	}

	slog.Info("successfully polled SLURM usage", "num_allocations", len(allocations), "num_jobs", len(jobs))

}
