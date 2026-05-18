package smapper

import "github.com/apache/airavata-custos/pkg/events"
import "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/subscribers"
import client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"

func LoadConnector(eventBus *events.Bus) error {
	slurmClient := client.New("localhost:8080", "", "") // Replace with actual SLURM client initialization.
	subscribers.NewAssociationSubscriber(slurmClient, eventBus).RegisterSubscribers()
	return nil
}
