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

//go:generate go run github.com/matryer/moq@latest -out mock.go -rm . CoreService

package service

import (
	"context"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// This file defines per-domain interfaces over the concrete *Service so that
// callers (HTTP handlers, subscribers, CLI commands, etc.) can depend on a
// narrow surface and be tested with hand-rolled or generated mocks instead
// of standing up the full database.
//
// The aggregate CoreService interface at the bottom of the file is the
// mockable surface used by tests. After changing any interface here, run
// `go generate ./pkg/service/...` to regenerate mock.go (CoreServiceMock).
//
// *Service satisfies every interface declared here implicitly. The
// compile-time assertion at the bottom of the file enforces that the
// aggregate interface stays in sync as new methods are added.

// OrganizationService exposes organization CRUD.
type OrganizationService interface {
	CreateOrganization(ctx context.Context, org *models.Organization) (*models.Organization, error)
	GetOrganization(ctx context.Context, id string) (*models.Organization, error)
	GetOrganizationByOriginatedID(ctx context.Context, originatedID string) (*models.Organization, error)
	UpdateOrganization(ctx context.Context, org *models.Organization) error
	DeleteOrganization(ctx context.Context, id string) error
}

// UserService exposes user CRUD and merge.
type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByUserIdentity(ctx context.Context, source, externalID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ListUsersByOrganization(ctx context.Context, organizationID string) ([]models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserStatus(ctx context.Context, id string, status models.UserStatus) (*models.User, error)
	DeleteUser(ctx context.Context, id string) error
	MergeUsers(ctx context.Context, survivingID, retiringID string) (*models.User, error)
}

// UserIdentityService exposes external-identity linking.
type UserIdentityService interface {
	CreateUserIdentity(ctx context.Context, e *models.UserIdentity) (*models.UserIdentity, error)
	GetUserIdentity(ctx context.Context, id string) (*models.UserIdentity, error)
	GetUserIdentityBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.UserIdentity, error)
	GetUserIdentityByOIDCSub(ctx context.Context, oidcSub string) (*models.UserIdentity, error)
	ListUserIdentitiesForUser(ctx context.Context, userID string) ([]models.UserIdentity, error)
	UpdateUserIdentity(ctx context.Context, e *models.UserIdentity) error
	DeleteUserIdentity(ctx context.Context, id string) error
}

// ProjectService exposes project CRUD and status transitions.
type ProjectService interface {
	CreateProject(ctx context.Context, project *models.Project) (*models.Project, error)
	GetProject(ctx context.Context, id string) (*models.Project, error)
	GetProjectByOriginatedID(ctx context.Context, originatedID string) (*models.Project, error)
	ListProjectsByPI(ctx context.Context, piUserID string) ([]models.Project, error)
	UpdateProject(ctx context.Context, project *models.Project) error
	UpdateProjectStatus(ctx context.Context, id string, status models.ProjectStatus) (*models.Project, error)
	DeleteProject(ctx context.Context, id string) error
}

// ComputeClusterService exposes compute-cluster CRUD.
type ComputeClusterService interface {
	CreateComputeCluster(ctx context.Context, cluster *models.ComputeCluster) (*models.ComputeCluster, error)
	GetComputeCluster(ctx context.Context, id string) (*models.ComputeCluster, error)
	GetComputeClusterByName(ctx context.Context, name string) (*models.ComputeCluster, error)
	ListComputeClusters(ctx context.Context) ([]models.ComputeCluster, error)
	UpdateComputeCluster(ctx context.Context, cluster *models.ComputeCluster) error
	DeleteComputeCluster(ctx context.Context, id string) error
}

// ComputeClusterUserService exposes per-cluster user records.
type ComputeClusterUserService interface {
	CreateComputeClusterUser(ctx context.Context, cu *models.ComputeClusterUser) (*models.ComputeClusterUser, error)
	GetComputeClusterUser(ctx context.Context, id string) (*models.ComputeClusterUser, error)
	GetComputeClusterUserByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error)
	ListComputeClusterUsersByCluster(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error)
	ListComputeClusterUsersByUser(ctx context.Context, userID string) ([]models.ComputeClusterUser, error)
	UpdateComputeClusterUser(ctx context.Context, cu *models.ComputeClusterUser) error
	DeleteComputeClusterUser(ctx context.Context, id string) error
}

// ComputeAllocationService exposes compute-allocation CRUD.
type ComputeAllocationService interface {
	CreateComputeAllocation(ctx context.Context, alloc *models.ComputeAllocation) (*models.ComputeAllocation, error)
	GetComputeAllocation(ctx context.Context, id string) (*models.ComputeAllocation, error)
	ListComputeAllocationsByProject(ctx context.Context, projectID string) ([]models.ComputeAllocation, error)
	ListComputeAllocationsByCluster(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error)
	UpdateComputeAllocation(ctx context.Context, alloc *models.ComputeAllocation) error
	DeleteComputeAllocation(ctx context.Context, id string) error
}

// ComputeAllocationResourceService exposes resource-definition CRUD.
type ComputeAllocationResourceService interface {
	CreateComputeAllocationResource(ctx context.Context, resource *models.ComputeAllocationResource) (*models.ComputeAllocationResource, error)
	GetComputeAllocationResource(ctx context.Context, id string) (*models.ComputeAllocationResource, error)
	GetComputeAllocationResourceByNameAndCluster(ctx context.Context, name, clusterID string) (*models.ComputeAllocationResource, error)
	ListComputeAllocationResources(ctx context.Context) ([]models.ComputeAllocationResource, error)
	UpdateComputeAllocationResource(ctx context.Context, resource *models.ComputeAllocationResource) error
	DeleteComputeAllocationResource(ctx context.Context, id string) error
}

// ComputeAllocationResourceMappingService exposes attach/detach for resources on allocations.
type ComputeAllocationResourceMappingService interface {
	AttachResourceToAllocation(ctx context.Context, allocationID, resourceID string, resourceAmount, resourceTime int64) (*models.ComputeAllocationResourceMapping, error)
	UpdateAllocationResourceMapping(ctx context.Context, allocationID, resourceID string, resourceAmount, resourceTime int64) (*models.ComputeAllocationResourceMapping, error)
	DetachResourceFromAllocation(ctx context.Context, allocationID, resourceID string) error
	ListResourcesForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error)
	ListAllocationsForResource(ctx context.Context, resourceID string) ([]models.ComputeAllocation, error)
}

// ComputeAllocationResourceRateService exposes resource-rate CRUD.
type ComputeAllocationResourceRateService interface {
	CreateComputeAllocationResourceRate(ctx context.Context, rate *models.ComputeAllocationResourceRate) (*models.ComputeAllocationResourceRate, error)
	GetComputeAllocationResourceRate(ctx context.Context, id string) (*models.ComputeAllocationResourceRate, error)
	ListRatesForResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationResourceRate, error)
	GetEffectiveRateForResource(ctx context.Context, resourceID string, at time.Time) (*models.ComputeAllocationResourceRate, error)
	UpdateComputeAllocationResourceRate(ctx context.Context, rate *models.ComputeAllocationResourceRate) error
	DeleteComputeAllocationResourceRate(ctx context.Context, id string) error
}

// ComputeAllocationDiffService exposes allocation-diff records.
type ComputeAllocationDiffService interface {
	CreateComputeAllocationDiff(ctx context.Context, diff *models.ComputeAllocationDiff) (*models.ComputeAllocationDiff, error)
	GetComputeAllocationDiff(ctx context.Context, id string) (*models.ComputeAllocationDiff, error)
	ListDiffsForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationDiff, error)
	GetLatestDiffForAllocation(ctx context.Context, allocationID string) (*models.ComputeAllocationDiff, error)
	DeleteComputeAllocationDiff(ctx context.Context, id string) error
}

// ComputeAllocationChangeRequestService exposes change-request CRUD.
type ComputeAllocationChangeRequestService interface {
	CreateComputeAllocationChangeRequest(ctx context.Context, req *models.ComputeAllocationChangeRequest) (*models.ComputeAllocationChangeRequest, error)
	GetComputeAllocationChangeRequest(ctx context.Context, id string) (*models.ComputeAllocationChangeRequest, error)
	ListChangeRequestsForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationChangeRequest, error)
	ListChangeRequestsByRequester(ctx context.Context, requesterID string) ([]models.ComputeAllocationChangeRequest, error)
	UpdateComputeAllocationChangeRequest(ctx context.Context, req *models.ComputeAllocationChangeRequest) (*models.ComputeAllocationChangeRequest, error)
	DeleteComputeAllocationChangeRequest(ctx context.Context, id string) error
}

// ComputeAllocationChangeRequestEventService exposes change-request events.
type ComputeAllocationChangeRequestEventService interface {
	CreateComputeAllocationChangeRequestEvent(ctx context.Context, evt *models.ComputeAllocationChangeRequestEvent) (*models.ComputeAllocationChangeRequestEvent, error)
	GetComputeAllocationChangeRequestEvent(ctx context.Context, id string) (*models.ComputeAllocationChangeRequestEvent, error)
	ListEventsForChangeRequest(ctx context.Context, changeRequestID string) ([]models.ComputeAllocationChangeRequestEvent, error)
	GetLatestEventForChangeRequest(ctx context.Context, changeRequestID string) (*models.ComputeAllocationChangeRequestEvent, error)
	DeleteComputeAllocationChangeRequestEvent(ctx context.Context, id string) error
}

// ComputeAllocationMembershipService exposes allocation memberships.
type ComputeAllocationMembershipService interface {
	CreateComputeAllocationMembership(ctx context.Context, m *models.ComputeAllocationMembership) (*models.ComputeAllocationMembership, error)
	GetComputeAllocationMembership(ctx context.Context, id string) (*models.ComputeAllocationMembership, error)
	ListMembersForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationMembership, error)
	ListAllocationsForUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error)
	UpdateComputeAllocationMembership(ctx context.Context, m *models.ComputeAllocationMembership) (*models.ComputeAllocationMembership, error)
	UpdateMembershipStatus(ctx context.Context, id string, status models.AllocationStatus) (*models.ComputeAllocationMembership, error)
	DeleteComputeAllocationMembership(ctx context.Context, id string) error
}

// ComputeAllocationMembershipResourceOverrideService exposes per-membership
// resource overrides.
type ComputeAllocationMembershipResourceOverrideService interface {
	CreateComputeAllocationMembershipResourceOverride(ctx context.Context, o *models.ComputeAllocationMembershipResourceOverride) (*models.ComputeAllocationMembershipResourceOverride, error)
	GetComputeAllocationMembershipResourceOverride(ctx context.Context, id string) (*models.ComputeAllocationMembershipResourceOverride, error)
	GetComputeAllocationMembershipResourceOverrideByPair(ctx context.Context, membershipID, resourceID string) (*models.ComputeAllocationMembershipResourceOverride, error)
	ListOverridesForMembership(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error)
	ListOverridesForResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationMembershipResourceOverride, error)
	UpdateComputeAllocationMembershipResourceOverride(ctx context.Context, o *models.ComputeAllocationMembershipResourceOverride) (*models.ComputeAllocationMembershipResourceOverride, error)
	DeleteComputeAllocationMembershipResourceOverride(ctx context.Context, id string) error
}

// ComputeAllocationUsageService exposes usage accounting.
type ComputeAllocationUsageService interface {
	CreateComputeAllocationUsage(ctx context.Context, u *models.ComputeAllocationUsage) (*models.ComputeAllocationUsage, error)
	GetComputeAllocationUsage(ctx context.Context, id string) (*models.ComputeAllocationUsage, error)
	ListUsagesForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationUsage, error)
	ListUsagesByUser(ctx context.Context, userID string) ([]models.ComputeAllocationUsage, error)
	GetTotalSUUsageForAllocation(ctx context.Context, allocationID string) (int64, error)
	GetTotalSUUsageForUserInAllocation(ctx context.Context, allocationID, userID string) (int64, error)
	DeleteComputeAllocationUsage(ctx context.Context, id string) error
}

// AuditEventService exposes the append-only audit-event log.
type AuditEventService interface {
	CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error)
	GetAuditEvent(ctx context.Context, id string) (*models.AuditEvent, error)
	ListAuditEventsByEntity(ctx context.Context, entityID string) ([]models.AuditEvent, error)
	ListAuditEventsByEventType(ctx context.Context, eventType string) ([]models.AuditEvent, error)
	ListAllAuditEvents(ctx context.Context) ([]*models.AuditEvent, error)
	DeleteAuditEvent(ctx context.Context, id string) error
}

// UserPrivilegeService exposes direct privilege grants. HasPrivilege
// returns true when the user holds the key directly or via any role.
type UserPrivilegeService interface {
	GrantPrivilege(ctx context.Context, userID string, privilege models.PrivilegeKey, granterID, reason string) (*models.UserPrivilege, error)
	RevokePrivilege(ctx context.Context, userID string, privilege models.PrivilegeKey, revokerID, reason string) error
	HasPrivilege(ctx context.Context, userID string, privilege models.PrivilegeKey) (bool, error)
	ListUserPrivileges(ctx context.Context, userID string) ([]models.UserPrivilege, error)
	ListPrivilegeHolders(ctx context.Context, privilege models.PrivilegeKey) ([]models.UserPrivilege, error)
	EffectivePrivileges(ctx context.Context, userID string) ([]models.PrivilegeKey, error)
	PrivilegeCatalog() []models.PrivilegeKey
}

// RoleService manages role definitions and their privilege bundles. All
// mutating calls require roles:manage.
type RoleService interface {
	CreateRole(ctx context.Context, name, description, actorID string) (*models.Role, error)
	UpdateRole(ctx context.Context, roleID, name, description, actorID string) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID, actorID string) error
	GetRole(ctx context.Context, roleID string) (*models.Role, error)
	ListRoles(ctx context.Context) ([]models.Role, error)
	ListRolePrivileges(ctx context.Context, roleID string) ([]models.PrivilegeKey, error)
	AddPrivilegeToRole(ctx context.Context, roleID string, privilege models.PrivilegeKey, actorID string) error
	RemovePrivilegeFromRole(ctx context.Context, roleID string, privilege models.PrivilegeKey, actorID string) error
}

// UserRoleService manages role assignments. Granting and revoking require
// roles:manage.
type UserRoleService interface {
	GrantRoleToUser(ctx context.Context, userID, roleID, granterID, reason string) (*models.UserRole, error)
	RevokeRoleFromUser(ctx context.Context, userID, roleID, revokerID, reason string) error
	ListUserRoles(ctx context.Context, userID string) ([]models.UserRole, error)
	ListRoleHolders(ctx context.Context, roleID string) ([]models.UserRole, error)
	BootstrapSuperAdmin(ctx context.Context, email, source string) error
}

// CoreService is the aggregate of every domain interface this package exposes.
// Most callers should depend on this, or — when they need only a slice of
// the API — on the narrower per-domain interfaces above.
type CoreService interface {
	OrganizationService
	UserService
	UserIdentityService
	ProjectService
	ComputeClusterService
	ComputeClusterUserService
	ComputeAllocationService
	ComputeAllocationResourceService
	ComputeAllocationResourceMappingService
	ComputeAllocationResourceRateService
	ComputeAllocationDiffService
	ComputeAllocationChangeRequestService
	ComputeAllocationChangeRequestEventService
	ComputeAllocationMembershipService
	ComputeAllocationMembershipResourceOverrideService
	ComputeAllocationUsageService
	AuditEventService
	UserPrivilegeService
	RoleService
	UserRoleService
}

// Compile-time assertion that *Service satisfies the aggregate CoreService.
// If a new public method is added on *Service, add it to the appropriate
// per-domain interface above; if a method signature drifts, this assertion
// will fail and pinpoint the offending file at build time.
var _ CoreService = (*Service)(nil)
