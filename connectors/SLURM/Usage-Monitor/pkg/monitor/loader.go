package monitor

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/connectors/SLURM/Usage-Monitor/internal/smonitor"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
	"github.com/jmoiron/sqlx"
)

func LoadConnector(ctx context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup) error {

	// Read url, username, and password from environment variables
	apiUrl := os.Getenv("SLURM_API")
	user := os.Getenv("SLURM_USER")
	token := os.Getenv("SLURM_TOKEN")
	apiVersion := os.Getenv("SLURM_API_VERSION")
	monitorClusterID := os.Getenv("SLURM_MONITOR_CLUSTER_ID")
	if monitorClusterID == "" {
		slog.Warn("SLURM_MONITOR_CLUSTER_ID not set, defaulting to 'slurm-cluster'")
		monitorClusterID = "slurm-cluster"
	}
	if apiUrl == "" || user == "" || token == "" || apiVersion == "" {
		slog.Warn("SLURM API credentials not fully provided, skipping SLURM Usage Monitor connector")
		slog.Warn("SLURM API credentials", "apiUrl", apiUrl, "user", user, "token", token, "apiVersion", apiVersion)
		return nil
	}

	slurmClient := client.New(apiUrl, user, token, apiVersion)
	monitor := smonitor.NewSlurmMonitor(slurmClient, eventBus, coreService, monitorClusterID)
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitor.StartMonitor(ctx)
	}()
	return nil
}
