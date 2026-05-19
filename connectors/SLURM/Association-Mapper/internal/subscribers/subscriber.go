package subscribers

import client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
import "github.com/apache/airavata-custos/pkg/events"
import "github.com/apache/airavata-custos/pkg/service"

type AssociationSubscriber struct {
	slurmClient *client.Client
	eventBus    *events.Bus
	coreService *service.Service
}

func NewAssociationSubscriber(slurmClient *client.Client, eventBus *events.Bus, coreService *service.Service) *AssociationSubscriber {
	return &AssociationSubscriber{
		slurmClient: slurmClient,
		eventBus:    eventBus,
		coreService: coreService,
	}
}

func (a *AssociationSubscriber) RegisterSubscribers() {
	a.eventBus.SubscribeComputeAllocationCreated(a.SubscribeToComputeAllocationCreation)
	a.eventBus.SubscribeComputeAllocationDeleted(a.SubscribeToComputeAllocationDeletion)
	a.eventBus.SubscribeComputeAllocationUpdated(a.SubscribeToComputeAllocationUpdate)
	a.eventBus.SubscribeComputeAllocationMembershipCreated(a.SubscribeToComputeAllocationMembershipCreation)
	a.eventBus.SubscribeComputeAllocationMembershipResourceOverrideCreated(a.SubscribeToComputeAllocationMembershipResourceOverrideCreation)
}
