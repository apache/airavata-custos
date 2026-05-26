package subscribers

import (
	"context"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type AssociationSubscriber struct {
	slurmClient *client.Client
	eventBus    *events.Bus
	coreService service.CoreService
}

func NewAssociationSubscriber(slurmClient *client.Client, eventBus *events.Bus, coreService service.CoreService) *AssociationSubscriber {
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
	a.eventBus.SubscribeComputeAllocationResourceMappingCreated(a.SubscribeToComputeAllocationResourceMappingCreation)
}

func (a *AssociationSubscriber) recordAuditEvent(eventType, entityId, message string) {
	auditEvent := &models.AuditEvent{
		EventType: eventType,
		EntityID:  entityId,
		Details:   message,
		EventTime: time.Now(),
	}
	a.coreService.CreateAuditEvent(context.Background(), auditEvent)
}
