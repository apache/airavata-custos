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

// Package subscribers syncs Slurm from Custos events; the reconciler is the fallback.
package subscribers

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// The Slurm account for an allocation. Slurm wants lowercase names in its accounting database.
func slurmAccount(allocationName string) string {
	return strings.ToLower(allocationName)
}

// SlurmClient is the subset of the REST client the subscribers use.
type SlurmClient interface {
	CreateAccount(a client.Account, cluster string) error
	ListAccounts() ([]client.Account, error)
	CreateUser(cluster, user string) error
	DeleteUser(name string) error
	ListAssociations(f client.AssocFilter) ([]client.Association, error)
	UpsertAssociation(a client.Association) error
	DeleteAssociation(f client.AssocFilter) error
	ListClusters() ([]string, error)
}

// MappingStore reads mapping amounts, which the core service does not expose.
type MappingStore interface {
	FindByPair(ctx context.Context, allocationID, resourceID string) (*models.ComputeAllocationResourceMapping, error)
}

type AssociationSubscriber struct {
	slurmClient       SlurmClient
	mappings          MappingStore
	eventBus          *events.Bus
	coreService       service.CoreService
	reconcileInterval time.Duration
	provisionGrace    time.Duration
	name              string
}

func NewAssociationSubscriber(slurmClient SlurmClient, eventBus *events.Bus, coreService service.CoreService, mappings MappingStore, reconcileInterval, provisionGrace time.Duration) *AssociationSubscriber {
	return &AssociationSubscriber{
		slurmClient:       slurmClient,
		mappings:          mappings,
		eventBus:          eventBus,
		coreService:       coreService,
		reconcileInterval: reconcileInterval,
		provisionGrace:    provisionGrace,
	}
}

// RegisterSubscribers subscribes the handlers under the given subscriber name.
func (a *AssociationSubscriber) RegisterSubscribers(subscriber string) {
	a.name = subscriber
	a.eventBus.SubscribeComputeAllocationCreated(subscriber, a.SubscribeToComputeAllocationChange)
	a.eventBus.SubscribeComputeAllocationDeleted(subscriber, a.SubscribeToComputeAllocationChange)
	a.eventBus.SubscribeComputeAllocationUpdated(subscriber, a.SubscribeToComputeAllocationChange)
	a.eventBus.SubscribeComputeAllocationMembershipCreated(subscriber, a.SubscribeToComputeAllocationMembershipChange)
	a.eventBus.SubscribeComputeAllocationMembershipUpdated(subscriber, a.SubscribeToComputeAllocationMembershipChange)
	a.eventBus.SubscribeComputeAllocationMembershipDeleted(subscriber, a.SubscribeToComputeAllocationMembershipChange)
	a.eventBus.SubscribeComputeAllocationMembershipResourceOverrideCreated(subscriber, a.SubscribeToComputeAllocationMembershipResourceOverrideChange)
	a.eventBus.SubscribeComputeAllocationMembershipResourceOverrideUpdated(subscriber, a.SubscribeToComputeAllocationMembershipResourceOverrideChange)
	a.eventBus.SubscribeComputeAllocationMembershipResourceOverrideDeleted(subscriber, a.SubscribeToComputeAllocationMembershipResourceOverrideChange)
	a.eventBus.SubscribeComputeAllocationResourceMappingCreated(subscriber, a.SubscribeToComputeAllocationResourceMappingChange)
	a.eventBus.SubscribeComputeAllocationResourceMappingUpdated(subscriber, a.SubscribeToComputeAllocationResourceMappingChange)
	a.eventBus.SubscribeComputeAllocationResourceMappingDeleted(subscriber, a.SubscribeToComputeAllocationResourceMappingChange)
	a.eventBus.SubscribeComputeClusterUserProvisioned(subscriber, a.SubscribeToComputeClusterUserChange)
	a.eventBus.SubscribeComputeClusterUserUpdated(subscriber, a.SubscribeToComputeClusterUserChange)
	a.eventBus.SubscribeComputeClusterUserDeleted(subscriber, a.SubscribeToComputeClusterUserChange)
}

// handleEvent audits the outcome; errNotProvisioned is left for the bus to retry.
func (a *AssociationSubscriber) handleEvent(ctx context.Context, name, id string, work func(context.Context) error) error {
	entity := name[:strings.LastIndex(name, "_")]
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm."+name)
	defer span.End()
	span.SetAttributes(attribute.String("slurm."+entity+"_id", id))

	err := work(ctx)
	if errors.Is(err, errNotProvisioned) {
		return err
	}
	if errors.Is(err, service.ErrNotFound) {
		slog.InfoContext(ctx, "Nothing to do on the cluster", "event", name, "id", id, "reason", err)
		err = nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "Slurm sync failed", "event", name, "id", id, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, camelCase(name)+"Failed", entity, id, err.Error())
		return err
	}
	a.recordAuditEvent(ctx, camelCase(name)+"Succeeded", entity, id, "")
	return nil
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

func camelCase(name string) string {
	words := strings.Split(name, "_")
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, "")
}

func (a *AssociationSubscriber) slurmCluster(ctx context.Context, id string) (string, error) {
	cluster, err := a.coreService.GetComputeCluster(ctx, id)
	if err != nil {
		return "", err
	}
	hosted, err := a.slurmClient.ListClusters()
	if err != nil || !slices.Contains(hosted, cluster.Name) {
		return "", err
	}
	return cluster.Name, nil
}
