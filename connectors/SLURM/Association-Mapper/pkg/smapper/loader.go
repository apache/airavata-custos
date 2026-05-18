package smapper

import "github.com/apache/airavata-custos/pkg/events"
import "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/subscribers"
import client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
import "os"
import "log/slog"
import "github.com/apache/airavata-custos/pkg/service"

func LoadConnector(eventBus *events.Bus, coreService *service.Service) error {

	// Read url, username, and password from environment variables
	apiUrl := os.Getenv("SLURM_API")
	user := os.Getenv("SLURM_USER")
	token := os.Getenv("SLURM_TOKEN")
	apiVersion := os.Getenv("SLURM_API_VERSION")
	if apiUrl == "" || user == "" || token == "" || apiVersion == "" {
		slog.Info("SLURM API credentials not fully provided, skipping SLURM Association Mapper connector")
		// print valualues of the env vars for debugging
		slog.Info("SLURM API credentials", "apiUrl", apiUrl, "user", user, "token", token, "apiVersion", apiVersion)
		return nil // skip loading if any of the required env vars are missing
	}

	slurmClient := client.New(apiUrl, user, token, apiVersion)
	subscribers.NewAssociationSubscriber(slurmClient, eventBus, coreService).RegisterSubscribers()
	return nil
}
