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
	"sync"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// SlurmAccount is the allocation's SLURM account name, lowercased for slurmdbd.
func SlurmAccount(allocationName string) string {
	return strings.ToLower(allocationName)
}

// SlurmClient is the subset of the REST client the subscribers use.
type SlurmClient interface {
	CreateAccount(a client.Account, cluster string) error
	UpsertAssociation(a client.Association) error
	SetDefaultAccount(cluster, user, account string) error
	ListAssociations(f client.AssocFilter) ([]client.Association, error)
	DeleteAssociation(f client.AssocFilter) error
	DeleteUser(name string) error
	ListQoS() ([]client.QoS, error)
	PostQoS(q client.QoS) error
	ListPartitions(cluster string) ([]client.Partition, error)
}

type AssociationSubscriber struct {
	slurmClient       SlurmClient
	eventBus          *events.Bus
	coreService       service.CoreService
	reconcileInterval time.Duration
	provisionGrace    time.Duration
	name              string
	clusterLocks      sync.Map // cluster ID -> *sync.Mutex
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
	b := a.eventBus
	b.SubscribeComputeAllocationCreated(subscriber, a.handleComputeAllocationCreated)
	b.SubscribeComputeAllocationUpdated(subscriber, a.handleComputeAllocationUpdated)
	b.SubscribeComputeAllocationDeleted(subscriber, a.handleComputeAllocationDeleted)
	b.SubscribeComputeAllocationMembershipCreated(subscriber, a.handleComputeAllocationMembershipCreated)
	b.SubscribeComputeAllocationMembershipUpdated(subscriber, a.handleComputeAllocationMembershipUpdated)
	b.SubscribeComputeAllocationMembershipDeleted(subscriber, a.handleComputeAllocationMembershipDeleted)
	b.SubscribeComputeAllocationMembershipResourceOverrideCreated(subscriber, a.handleComputeAllocationMembershipResourceOverrideCreated)
	b.SubscribeComputeAllocationResourceMappingCreated(subscriber, a.handleComputeAllocationResourceMappingCreated)
	b.SubscribeComputeClusterUserDeleted(subscriber, a.handleComputeClusterUserDeleted)
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

func (a *AssociationSubscriber) handleComputeAllocationCreated(ctx context.Context, x models.ComputeAllocation) error {
	return allocationCreated.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationUpdated(ctx context.Context, x models.ComputeAllocation) error {
	return allocationUpdated.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationDeleted(ctx context.Context, x models.ComputeAllocation) error {
	return allocationDeleted.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationMembershipCreated(ctx context.Context, x models.ComputeAllocationMembership) error {
	return membershipCreated.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationMembershipUpdated(ctx context.Context, x models.ComputeAllocationMembership) error {
	return membershipUpdated.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationMembershipDeleted(ctx context.Context, x models.ComputeAllocationMembership) error {
	return membershipDeleted.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationMembershipResourceOverrideCreated(ctx context.Context, x models.ComputeAllocationMembershipResourceOverride) error {
	return overrideCreated.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeAllocationResourceMappingCreated(ctx context.Context, x models.ComputeAllocationResourceMapping) error {
	return mappingCreated.run(ctx, a, x)
}

func (a *AssociationSubscriber) handleComputeClusterUserDeleted(ctx context.Context, x models.ComputeClusterUser) error {
	return clusterUserDeleted.run(ctx, a, x)
}
