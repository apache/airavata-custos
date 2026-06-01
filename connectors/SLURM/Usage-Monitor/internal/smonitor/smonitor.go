package smonitor

import (
	"context"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	//"github.com/apache/airavata-custos/pkg/models"
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

	slog.Info("starting SLURM usage monitor", "interval", monitorInterval)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping SLURM usage monitor", "reason", ctx.Err())
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

	/*
		for _, job := range jobs {
			//slog.Debug("processing SLURM job", "job_id", job.JobID, "job_name", job.Name)
			//m.coreService.GetComputeAllocationResource()

			usageModel := &models.ComputeAllocationUsage{
				ComputeAllocationID:         "",
				UsedRawAmount:               23,
				UsedSUAmount:                23,
				CalculatedTime:              time.Now(),
				UserID:                      "32",
				JobID:                       "",
				ComputeAllocationResourceID: "",
			}

			m.coreService.CreateComputeAllocationUsage(context, usageModel)
		}*/

	slog.Info("successfully polled SLURM usage", "num_allocations", len(allocations), "num_jobs", len(jobs))

}
