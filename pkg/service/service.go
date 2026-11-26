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

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// Service is a high-level façade over the underlying stores. It wraps each
// mutating operation in a transaction so callers do not need to manage
// *sql.Tx themselves.
type Service struct {
	db                  *sqlx.DB
	eventBus            *events.Bus
	reservedUsernames   map[string]bool
	orgs                store.OrganizationStore
	users               store.UserStore
	projs               store.ProjectStore
	clusters            store.ComputeClusterStore
	clusterUsers        store.ComputeClusterUserStore
	clusterQoS          store.ComputeClusterQoSStore
	allocs              store.ComputeAllocationStore
	resources           store.ComputeAllocationResourceStore
	resourceMappings    store.ComputeAllocationResourceMappingStore
	resourceRates       store.ComputeAllocationResourceRateStore
	allocDiffs          store.ComputeAllocationDiffStore
	changeRequests      store.ComputeAllocationChangeRequestStore
	changeEvents        store.ComputeAllocationChangeRequestEventStore
	memberships         store.ComputeAllocationMembershipStore
	projMemberships     store.ProjectMembershipStore
	membershipOverrides store.ComputeAllocationMembershipResourceOverrideStore
	usages              store.ComputeAllocationUsageStore
	userIdentities      store.UserIdentityStore
	auditEvents         store.AuditEventStore
	privileges          store.UserPrivilegeStore
	roles               store.RoleStore
	userRoles           store.UserRoleStore
	agreements          store.UserAgreementStore
	acceptances         store.UserAgreementAcceptanceStore
	auditTraces         store.AuditTraceStore
	serviceAccounts     store.ServiceAccountStore
	storages            store.StorageStore
	storageTargets      map[string]map[models.StorageType]models.StorageTarget
	claimedClusters     map[string]StorageUsage
	storageHolders      map[models.StorageType]StorageHolderOf
	claimedMu           sync.RWMutex // connectors claim clusters while others already run
	identityCache       *identityCache
}

// New constructs a Service backed by the supplied database handle.
// Stores are instantiated internally using the default PostgreSQL implementations.
func New(database *sqlx.DB, eventBus *events.Bus) *Service {
	return &Service{
		db:                  database,
		eventBus:            eventBus,
		reservedUsernames:   map[string]bool{},
		orgs:                store.NewOrganizationStore(database),
		users:               store.NewUserStore(database),
		projs:               store.NewProjectStore(database),
		clusters:            store.NewComputeClusterStore(database),
		clusterUsers:        store.NewComputeClusterUserStore(database),
		clusterQoS:          store.NewComputeClusterQoSStore(database),
		allocs:              store.NewComputeAllocationStore(database),
		resources:           store.NewComputeAllocationResourceStore(database),
		resourceMappings:    store.NewComputeAllocationResourceMappingStore(database),
		resourceRates:       store.NewComputeAllocationResourceRateStore(database),
		allocDiffs:          store.NewComputeAllocationDiffStore(database),
		changeRequests:      store.NewComputeAllocationChangeRequestStore(database),
		changeEvents:        store.NewComputeAllocationChangeRequestEventStore(database),
		memberships:         store.NewComputeAllocationMembershipStore(database),
		projMemberships:     store.NewProjectMembershipStore(database),
		membershipOverrides: store.NewComputeAllocationMembershipResourceOverrideStore(database),
		usages:              store.NewComputeAllocationUsageStore(database),
		userIdentities:      store.NewUserIdentityStore(database),
		auditEvents:         store.NewAuditEventStore(database),
		privileges:          store.NewUserPrivilegeStore(database),
		roles:               store.NewRoleStore(database),
		userRoles:           store.NewUserRoleStore(database),
		agreements:          store.NewUserAgreementStore(database),
		acceptances:         store.NewUserAgreementAcceptanceStore(database),
		auditTraces:         store.NewAuditTraceStore(database),
		serviceAccounts:     store.NewServiceAccountStore(database),
		storages:            store.NewStorageStore(database),
		claimedClusters:     map[string]StorageUsage{},
		storageHolders:      map[models.StorageType]StorageHolderOf{},
		identityCache:       newIdentityCache(),
	}
}

// NewWithStores constructs a Service from explicit stores. Useful for tests
// within this module — stores are an internal type and cannot be supplied by
// external callers.
func NewWithStores(
	database *sqlx.DB,
	eventBus *events.Bus,
	orgs store.OrganizationStore,
	users store.UserStore,
	projs store.ProjectStore,
	clusters store.ComputeClusterStore,
	clusterUsers store.ComputeClusterUserStore,
	allocs store.ComputeAllocationStore,
	resources store.ComputeAllocationResourceStore,
	resourceMappings store.ComputeAllocationResourceMappingStore,
	resourceRates store.ComputeAllocationResourceRateStore,
	allocDiffs store.ComputeAllocationDiffStore,
	changeRequests store.ComputeAllocationChangeRequestStore,
	changeEvents store.ComputeAllocationChangeRequestEventStore,
	membershipOverrides store.ComputeAllocationMembershipResourceOverrideStore,
	memberships store.ComputeAllocationMembershipStore,
	projMemberships store.ProjectMembershipStore,
	usages store.ComputeAllocationUsageStore,
	userIdentities store.UserIdentityStore,
	auditEvents store.AuditEventStore,
	privileges store.UserPrivilegeStore,
	roles store.RoleStore,
	userRoles store.UserRoleStore,
	agreements store.UserAgreementStore,
	acceptances store.UserAgreementAcceptanceStore,
) *Service {
	return &Service{
		db:                  database,
		eventBus:            eventBus,
		orgs:                orgs,
		users:               users,
		projs:               projs,
		clusters:            clusters,
		clusterUsers:        clusterUsers,
		allocs:              allocs,
		resources:           resources,
		resourceMappings:    resourceMappings,
		resourceRates:       resourceRates,
		allocDiffs:          allocDiffs,
		changeRequests:      changeRequests,
		changeEvents:        changeEvents,
		membershipOverrides: membershipOverrides,
		memberships:         memberships,
		projMemberships:     projMemberships,
		usages:              usages,
		userIdentities:      userIdentities,
		auditEvents:         auditEvents,
		privileges:          privileges,
		roles:               roles,
		userRoles:           userRoles,
		agreements:          agreements,
		acceptances:         acceptances,
		auditTraces:         store.NewAuditTraceStore(database),
		serviceAccounts:     store.NewServiceAccountStore(database),
		storages:            store.NewStorageStore(database),
		claimedClusters:     map[string]StorageUsage{},
		storageHolders:      map[models.StorageType]StorageHolderOf{},
		identityCache:       newIdentityCache(),
	}
}

func (s *Service) AuditTraces() store.AuditTraceStore {
	return s.auditTraces
}

func (s *Service) EventBus() *events.Bus {
	return s.eventBus
}

func (s *Service) SetAuditTraces(ts store.AuditTraceStore) {
	s.auditTraces = ts
}

// inTx runs fn inside a database transaction managed by the Service.
func (s *Service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return db.TxFn(ctx, s.db, fn)
}

func (s *Service) writeAuditTx(ctx context.Context, tx *sql.Tx, entityType, eventType, entityID string, details map[string]any) error {
	// Record the service account too when one made the request.
	if caller := identity.CallerFromContext(ctx); caller != nil && caller.IsServiceAccount() {
		details["service_account_key"] = caller.ServiceAccountKey
	}
	payload, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}
	return s.auditEvents.Create(ctx, tx, &models.AuditEvent{
		ID:         newID(),
		EventType:  eventType,
		EventTime:  nowUTC(),
		EntityID:   entityID,
		EntityType: entityType,
		Details:    string(payload),
	})
}

// newID returns a new identifier for a freshly created entity.
func newID() string {
	return uuid.NewString()
}

// nowUTC returns the current time in UTC.
func nowUTC() time.Time {
	return time.Now().UTC()
}

// SlurmPendingKind names the table a pending mark sits on.
type SlurmPendingKind = store.SlurmPendingTable

const (
	SlurmPendingAllocation = store.SlurmPendingAllocations
	SlurmPendingMembership = store.SlurmPendingMemberships
	SlurmPendingQoS        = store.SlurmPendingQoS
)

type withoutSlurmMarkKey struct{}

// WithoutSlurmMark returns a context whose writes set no mark, for saving what SLURM already holds.
func WithoutSlurmMark(ctx context.Context) context.Context {
	return context.WithValue(ctx, withoutSlurmMarkKey{}, true)
}

func slurmMarkWanted(ctx context.Context) bool {
	return ctx.Value(withoutSlurmMarkKey{}) == nil
}

// markSlurmPendingTx marks the row pending, unless ctx comes from WithoutSlurmMark.
func (s *Service) markSlurmPendingTx(ctx context.Context, tx *sql.Tx, kind SlurmPendingKind, id string) error {
	if !slurmMarkWanted(ctx) {
		return nil
	}
	_, err := store.MarkSlurmPending(ctx, tx, kind, `id = $1`, id)
	return err
}

// SetSlurmPending marks the row pending when clearSince is nil, else clears its mark unless a change marked it after clearSince.
func (s *Service) SetSlurmPending(ctx context.Context, kind SlurmPendingKind, id string, clearSince *time.Time) error {
	found := true
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if clearSince != nil {
			return store.ClearSlurmPending(ctx, tx, kind, id, *clearSince)
		}
		found, err = store.MarkSlurmPending(ctx, tx, kind, `id = $1`, id)
		return err
	}); err != nil {
		return fmt.Errorf("set slurm pending %s: %w", kind, err)
	}
	if !found {
		return ErrNotFound
	}
	return nil
}

// refuseDeleteHeldBySlurm refuses deleting a row SLURM still holds or has a change pending for.
func refuseDeleteHeldBySlurm(what, id string, status models.AllocationStatus, pendingSince *time.Time) error {
	if status == models.ACTIVE {
		return fmt.Errorf("%w: %s %q is active, set it INACTIVE first", ErrInUse, what, id)
	}
	if pendingSince != nil {
		return fmt.Errorf("%w: %s %q has a change SLURM has not applied yet", ErrInUse, what, id)
	}
	return nil
}
