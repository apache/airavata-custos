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
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// The Slurm account for an allocation. Slurm wants lowercase names in its accounting database.
func slurmAccount(allocationName string) string {
	return strings.ToLower(allocationName)
}

// SlurmClient is the subset of the REST client the subscribers use.
type SlurmClient interface {
	CreateAccount(a client.Account, cluster string) error
	UpsertAssociation(a client.Association) error
	SetDefaultAccount(cluster, user, account string) error
	ListAssociations(f client.AssocFilter) ([]client.Association, error)
	DeleteAssociations(f client.AssocFilter) error
}

type AssociationSubscriber struct {
	slurmClient       SlurmClient
	eventBus          *events.Bus
	coreService       service.CoreService
	reconcileInterval time.Duration
	provisionGrace    time.Duration
	name              string
}

func NewAssociationSubscriber(slurmClient SlurmClient, eventBus *events.Bus, coreService service.CoreService, reconcileInterval, provisionGrace time.Duration) *AssociationSubscriber {
	return &AssociationSubscriber{
		slurmClient:       slurmClient,
		eventBus:          eventBus,
		coreService:       coreService,
		reconcileInterval: reconcileInterval,
		provisionGrace:    provisionGrace,
	}
}

// RegisterSubscribers subscribes the handlers under the given subscriber name.
func (a *AssociationSubscriber) RegisterSubscribers(subscriber string) {
	a.name = subscriber
	a.eventBus.SubscribeComputeAllocationCreated(subscriber, a.handleComputeAllocationCreated)
	a.eventBus.SubscribeComputeAllocationDeleted(subscriber, a.handleComputeAllocationDeleted)
	a.eventBus.SubscribeComputeAllocationUpdated(subscriber, a.handleComputeAllocationUpdated)
	a.eventBus.SubscribeComputeAllocationMembershipCreated(subscriber, a.handleComputeAllocationMembershipCreated)
	a.eventBus.SubscribeComputeAllocationMembershipUpdated(subscriber, a.handleComputeAllocationMembershipUpdated)
	a.eventBus.SubscribeComputeAllocationMembershipDeleted(subscriber, a.handleComputeAllocationMembershipDeleted)
	a.eventBus.SubscribeComputeAllocationMembershipResourceOverrideCreated(subscriber, a.handleComputeAllocationMembershipResourceOverrideCreated)
	a.eventBus.SubscribeComputeAllocationResourceMappingCreated(subscriber, a.handleComputeAllocationResourceMappingCreated)
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
