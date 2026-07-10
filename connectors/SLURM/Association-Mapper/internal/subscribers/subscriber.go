// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

func (a *AssociationSubscriber) recordAuditEvent(ctx context.Context, eventType, entityType, entityId, message string) {
	auditEvent := &models.AuditEvent{
		EventType:  eventType,
		EntityID:   entityId,
		EntityType: entityType,
		Details:    message,
		EventTime:  time.Now(),
	}
	a.coreService.CreateAuditEvent(ctx, auditEvent)
}
