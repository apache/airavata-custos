package smapper

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"

	client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/subscribers"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

func LoadConnector(_ context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, _ *sync.WaitGroup) error {

	// Read url, username, and password from environment variables
	apiUrl := os.Getenv("SLURM_API")
	user := os.Getenv("SLURM_USER")
	token := os.Getenv("SLURM_TOKEN")
	apiVersion := os.Getenv("SLURM_API_VERSION")
	if apiUrl == "" || user == "" || token == "" || apiVersion == "" {
		slog.Info("SLURM API credentials not fully provided, skipping SLURM Association Mapper connector")
		slog.Info("SLURM API credentials", "apiUrl", apiUrl, "user", user, "token", token, "apiVersion", apiVersion)
		return nil
	}

	slurmClient := client.New(apiUrl, user, token, apiVersion)
	subscribers.NewAssociationSubscriber(slurmClient, eventBus, coreService).RegisterSubscribers()
	return nil
}
