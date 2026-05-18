package subscribers

import client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
import "github.com/apache/airavata-custos/pkg/events"

type AssociationSubscriber struct {
	slurmClient *client.Client
	eventBus    *events.Bus
}

func NewAssociationSubscriber(slurmClient *client.Client, eventBus *events.Bus) *AssociationSubscriber {
	return &AssociationSubscriber{
		slurmClient: slurmClient,
		eventBus:    eventBus,
	}
}

func (a *AssociationSubscriber) RegisterSubscribers() {
	a.eventBus.SubscribeComputeAllocationCreated(a.SubscribeToComputeAccountCreation)
	a.eventBus.SubscribeComputeAllocationDeleted(a.SubscribeToComputeAccountDeletion)
	a.eventBus.SubscribeComputeAllocationUpdated(a.SubscribeToComputeAccountUpdate)
}
