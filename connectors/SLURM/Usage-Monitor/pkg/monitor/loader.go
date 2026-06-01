package monitor

import (
	"context"
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/connectors/SLURM/Usage-Monitor/internal/smonitor"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
	"github.com/jmoiron/sqlx"
	"log/slog"
	"os"
	"sync"
)

func LoadConnector(ctx context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup) error {

	// Read url, username, and password from environment variables
	apiUrl := os.Getenv("SLURM_API")
	user := os.Getenv("SLURM_USER")
	token := os.Getenv("SLURM_TOKEN")
	apiVersion := os.Getenv("SLURM_API_VERSION")
	if apiUrl == "" || user == "" || token == "" || apiVersion == "" {
		slog.Info("SLURM API credentials not fully provided, skipping SLURM Usage Monitor connector")
		slog.Info("SLURM API credentials", "apiUrl", apiUrl, "user", user, "token", token, "apiVersion", apiVersion)
		return nil
	}

	slurmClient := client.New(apiUrl, user, token, apiVersion)
	monitor := smonitor.NewSlurmMonitor(slurmClient, eventBus, coreService)
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitor.StartMonitor(ctx)
	}()
	return nil
}
