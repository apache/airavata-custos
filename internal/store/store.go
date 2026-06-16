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

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// UserStore defines persistence operations for users.
type UserStore interface {
	// FindByID returns the user with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.User, error)
	// FindByEmail returns the user with the given email, or nil if not found.
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// FindByOrganization returns all users belonging to the given organization.
	FindByOrganization(ctx context.Context, organizationID string) ([]models.User, error)
	// Create inserts a new user within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, u *models.User) error
	// Update replaces mutable fields of an existing user within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, u *models.User) error
	// UpdateStatus sets the lifecycle status of an existing user within the provided transaction.
	UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.UserStatus) error
	// Delete removes a user by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// OrganizationStore defines persistence operations for organizations.
type OrganizationStore interface {
	// FindByID returns the organization with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.Organization, error)
	// FindByOriginatedID returns the organization matching the external originated ID, or nil if not found.
	FindByOriginatedID(ctx context.Context, originatedID string) (*models.Organization, error)
	// Create inserts a new organization within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, o *models.Organization) error
	// Update replaces mutable fields of an existing organization within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, o *models.Organization) error
	// Delete removes an organization by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ComputeClusterStore defines persistence operations for compute clusters.
type ComputeClusterStore interface {
	// FindByID returns the cluster with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.ComputeCluster, error)
	// FindByName returns the cluster with the given name, or nil if not found.
	FindByName(ctx context.Context, name string) (*models.ComputeCluster, error)
	// List returns all compute clusters.
	List(ctx context.Context) ([]models.ComputeCluster, error)
	// Create inserts a new cluster within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, c *models.ComputeCluster) error
	// Update replaces mutable fields of an existing cluster within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, c *models.ComputeCluster) error
	// Delete removes a cluster by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ComputeClusterUserStore defines persistence operations for the mapping of
// a Custos user to their local account on a compute cluster.
type ComputeClusterUserStore interface {
	// FindByID returns the mapping with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.ComputeClusterUser, error)
	// FindByPair returns the mapping for a (compute_cluster_id, user_id) pair, or nil if absent.
	FindByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error)
	// FindByLocalUsernameAndCluster returns the mapping for a (local_username, compute_cluster_id) pair, or nil if absent.
	FindByLocalUsernameAndCluster(ctx context.Context, clusterID, localUsername string) (*models.ComputeClusterUser, error)
	// FindByCluster returns every user mapping for the given compute cluster.
	FindByCluster(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error)
	// FindByUser returns every cluster mapping held by the given Custos user.
	FindByUser(ctx context.Context, userID string) ([]models.ComputeClusterUser, error)
	// Create inserts a new mapping within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error
	// Update replaces mutable fields of an existing mapping within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error
	// ReassignUser moves every mapping owned by fromUserID over to toUserID,
	// dropping fromUserID's rows on clusters where toUserID already has one.
	ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error
	// Delete removes a mapping by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// UserIdentityStore defines persistence operations for the bindings between
// Custos users and identifiers issued by external systems (ACCESS, NAIRR,
// CILogon, etc.).
type UserIdentityStore interface {
	// FindByID returns the user identity with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.UserIdentity, error)
	// FindBySourceAndExternalID returns the binding for the given (source, external_id) pair, or nil if absent.
	FindBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.UserIdentity, error)
	// FindByOIDCSub returns the first binding matching the given OIDC subject, or nil if none.
	FindByOIDCSub(ctx context.Context, oidcSub string) (*models.UserIdentity, error)
	// FindByUser returns every user identity bound to the given user, ordered by created_at.
	FindByUser(ctx context.Context, userID string) ([]models.UserIdentity, error)
	// Create inserts a new user identity within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, e *models.UserIdentity) error
	// Update replaces mutable fields of an existing user identity within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, e *models.UserIdentity) error
	// ReassignUser moves every user identity owned by fromUserID over to toUserID.
	ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error
	// Delete removes a user identity by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ProjectStore defines persistence operations for projects.
type ProjectStore interface {
	// FindByID returns the project with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.Project, error)
	// FindByOriginatedID returns the project matching the external originated ID, or nil if not found.
	FindByOriginatedID(ctx context.Context, originatedID string) (*models.Project, error)
	// FindByPI returns all projects whose PI matches the given user ID.
	FindByPI(ctx context.Context, piUserID string) ([]models.Project, error)
	// Create inserts a new project within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, p *models.Project) error
	// Update replaces mutable fields of an existing project within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, p *models.Project) error
	// UpdateStatus sets the lifecycle status of an existing project within the provided transaction.
	UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.ProjectStatus) error
	// ReassignPI changes project_pi_id from fromUserID to toUserID for every project.
	ReassignPI(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error
	// Delete removes a project by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	// List returns a paginated, filtered slice of projects plus the total
	// count matching the filter (ignoring limit/offset).
	List(ctx context.Context, f ProjectListFilter) ([]models.Project, int, error)
	// FindByIDWithPI is FindByID joined with the PI user, so a single query
	// returns everything the portal needs to render a project header.
	FindByIDWithPI(ctx context.Context, id string) (*ProjectWithPI, error)
	// ListWithPI is List joined with the PI user, replacing the per-row
	// GetUser fan-out the handler would otherwise need.
	ListWithPI(ctx context.Context, f ProjectListFilter) ([]ProjectWithPI, int, error)
}

// ProjectListFilter selects which projects ProjectStore.List returns.
type ProjectListFilter struct {
	PIID   string
	Status string
	Query  string
	Limit  int
	Offset int
}

// ComputeAllocationStore defines persistence operations for compute allocations.
type ComputeAllocationStore interface {
	// FindByID returns the allocation with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocation, error)
	// FindByProject returns all allocations attached to the given project.
	FindByProject(ctx context.Context, projectID string) ([]models.ComputeAllocation, error)
	// FindByCluster returns all allocations attached to the given compute cluster.
	FindByCluster(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error)
	// Create inserts a new allocation within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error
	// Update replaces mutable fields of an existing allocation within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error
	// Delete removes an allocation by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	// List returns a paginated, filtered slice of allocations plus the total
	// count matching the filter (ignoring limit/offset).
	List(ctx context.Context, f AllocationListFilter) ([]models.ComputeAllocation, int, error)
}

// AllocationListFilter selects which allocations ComputeAllocationStore.List returns.
type AllocationListFilter struct {
	ProjectID string
	Status    string
	Query     string
	Limit     int
	Offset    int
}

// ComputeAllocationResourceStore defines persistence operations for compute
// allocation resources (CPU, GPU, etc.).
type ComputeAllocationResourceStore interface {
	// FindByID returns the resource with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationResource, error)
	// FindByNameAndCluster returns the resource with the given name on the
	// given compute cluster, or nil if not found.
	FindByNameAndCluster(ctx context.Context, name, clusterID string) (*models.ComputeAllocationResource, error)
	// FindByTypeAndCluster returns all resources of the given type on the
	// given compute cluster.
	FindByTypeAndCluster(ctx context.Context, resourceType, clusterID string) ([]models.ComputeAllocationResource, error)
	// List returns all compute allocation resources.
	List(ctx context.Context) ([]models.ComputeAllocationResource, error)
	// Create inserts a new resource within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResource) error
	// Update replaces mutable fields of an existing resource within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResource) error
	// Delete removes a resource by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ComputeAllocationResourceMappingStore defines persistence operations for
// the join table linking compute allocations and compute allocation resources.
type ComputeAllocationResourceMappingStore interface {
	// FindByID returns the mapping with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationResourceMapping, error)
	// FindByPair returns the mapping for a (allocation, resource) pair, or nil if absent.
	FindByPair(ctx context.Context, allocationID, resourceID string) (*models.ComputeAllocationResourceMapping, error)
	// FindResourcesByAllocation returns every resource attached to the given allocation.
	FindResourcesByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error)
	// FindAllocationsByResource returns every allocation that has the given resource attached.
	FindAllocationsByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocation, error)
	// Create inserts a new mapping within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error
	// Update replaces mutable fields of an existing mapping within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error
	// DeleteByPair removes the mapping for a (allocation, resource) pair within the provided transaction.
	DeleteByPair(ctx context.Context, tx *sql.Tx, allocationID, resourceID string) error
}

// ComputeAllocationResourceRateStore defines persistence operations for
// the time-windowed rate at which a compute allocation resource is charged
// in Service Units.
type ComputeAllocationResourceRateStore interface {
	// FindByID returns the rate with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationResourceRate, error)
	// FindByResource returns every rate ever defined for the given resource,
	// ordered by start_time ascending.
	FindByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationResourceRate, error)
	// FindEffective returns the rate effective for the given resource at the
	// supplied instant (start_time <= at < end_time), or nil if none applies.
	FindEffective(ctx context.Context, resourceID string, at time.Time) (*models.ComputeAllocationResourceRate, error)
	// Create inserts a new rate within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResourceRate) error
	// Update replaces mutable fields of an existing rate within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResourceRate) error
	// Delete removes a rate by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ComputeAllocationDiffStore defines persistence operations for the
// append-only log of changes (SU updates, status changes, etc.) applied to a
// compute allocation.
type ComputeAllocationDiffStore interface {
	// FindByID returns the diff with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationDiff, error)
	// FindByAllocation returns every diff ever recorded for the given allocation,
	// ordered by timestamp ascending.
	FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationDiff, error)
	// FindLatestByAllocation returns the most recent diff for the given
	// allocation (highest timestamp), or nil if none exist.
	FindLatestByAllocation(ctx context.Context, allocationID string) (*models.ComputeAllocationDiff, error)
	// Create inserts a new diff within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, d *models.ComputeAllocationDiff) error
	// Delete removes a diff by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ComputeAllocationChangeRequestStore defines persistence operations for
// user/admin requests to change a compute allocation (e.g. asking for more
// SUs or to change its status).
type ComputeAllocationChangeRequestStore interface {
	// FindByID returns the change request with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationChangeRequest, error)
	// FindByAllocation returns every change request ever recorded against the
	// given allocation, ordered by timestamp ascending.
	FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationChangeRequest, error)
	// FindByRequester returns every change request made by the given user,
	// ordered by timestamp ascending.
	FindByRequester(ctx context.Context, requesterID string) ([]models.ComputeAllocationChangeRequest, error)
	// Create inserts a new change request within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, c *models.ComputeAllocationChangeRequest) error
	// Update replaces mutable fields of an existing change request within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, c *models.ComputeAllocationChangeRequest) error
	// Delete removes a change request by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	// List returns change requests filtered by the supplied criteria.
	List(ctx context.Context, f ChangeRequestListFilter) ([]models.ComputeAllocationChangeRequest, error)
}

// ChangeRequestListFilter selects which change request
// ComputeAllocationChangeRequestStore.List returns.
type ChangeRequestListFilter struct {
	Status string
	Limit  int
}

// ComputeAllocationChangeRequestEventStore defines persistence operations
// for the append-only audit trail of state transitions applied to a
// ComputeAllocationChangeRequest.
type ComputeAllocationChangeRequestEventStore interface {
	// FindByID returns the event with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationChangeRequestEvent, error)
	// FindByChangeRequest returns every event recorded against the given
	// change request, ordered by timestamp ascending.
	FindByChangeRequest(ctx context.Context, changeRequestID string) ([]models.ComputeAllocationChangeRequestEvent, error)
	// FindLatestByChangeRequest returns the most recent event for the given
	// change request, or nil if none exist.
	FindLatestByChangeRequest(ctx context.Context, changeRequestID string) (*models.ComputeAllocationChangeRequestEvent, error)
	// Create inserts a new event within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, e *models.ComputeAllocationChangeRequestEvent) error
	// Delete removes an event by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ProjectMembershipStore defines persistence operations for project-level
// governance roles (PI / CO_PI / ALLOCATION_MANAGER). MEMBER is derived from
// compute_allocation_memberships and not stored here.
type ProjectMembershipStore interface {
	// FindByPair returns the (project, user) row, or nil if absent.
	FindByPair(ctx context.Context, projectID, userID string) (*models.ProjectMembership, error)
	// FindByProject returns every project_memberships row for the project.
	FindByProject(ctx context.Context, projectID string) ([]models.ProjectMembership, error)
	// FindPIByProject returns the PI row, or nil if the project has no PI yet.
	FindPIByProject(ctx context.Context, projectID string) (*models.ProjectMembership, error)
	// Create inserts a new row within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, pm *models.ProjectMembership) error
	// UpdateRole changes the role of an existing (project, user) row.
	UpdateRole(ctx context.Context, tx *sql.Tx, projectID, userID string, role models.ProjectRole) error
	// Delete removes the (project, user) row.
	Delete(ctx context.Context, tx *sql.Tx, projectID, userID string) error
	// ReassignUser moves every project_memberships row owned by fromUserID
	// over to toUserID, dropping fromUserID's rows on projects where toUserID
	// already has one.
	ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error
}

// ComputeAllocationMembershipStore defines persistence operations for the
// per-user membership of a compute allocation, including the SU sub-allocation
// granted to that user and the membership lifecycle.
type ComputeAllocationMembershipStore interface {
	// FindByID returns the membership with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationMembership, error)
	// FindByPair returns the membership for a (allocation, user) pair, or nil if absent.
	FindByPair(ctx context.Context, allocationID, userID string) (*models.ComputeAllocationMembership, error)
	// FindByAllocation returns every membership recorded against the given
	// allocation, ordered by start_time ascending.
	FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationMembership, error)
	// FindByUser returns every membership held by the given user, ordered by
	// start_time ascending.
	FindByUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error)
	// FindByAllocationWithUser is FindByAllocation joined with user +
	// project_memberships so each row carries display_name, email, and the
	// project-level role (defaulted to MEMBER).
	FindByAllocationWithUser(ctx context.Context, allocationID string) ([]MembershipWithUser, error)
	// FindByProjectWithUser returns memberships across every allocation in
	// the project joined with users, allocation name, and project-level role.
	// The caller aggregates per user (collapsing into the response's
	// allocations list).
	FindByProjectWithUser(ctx context.Context, projectID string) ([]MembershipWithUser, error)
	// Create inserts a new membership within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error
	// Update replaces mutable fields of an existing membership within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error
	// ReassignUser moves every membership owned by fromUserID over to toUserID,
	// dropping fromUserID's rows on allocations where toUserID already has one.
	ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error
	// Delete removes a membership by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// ComputeAllocationMembershipResourceOverrideStore defines persistence
// operations for per-resource overrides of the SU amount granted to a
// compute allocation membership.
type ComputeAllocationMembershipResourceOverrideStore interface {
	// FindByID returns the override with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationMembershipResourceOverride, error)
	// FindByPair returns the override for a (membership, resource) pair, or nil if absent.
	FindByPair(ctx context.Context, membershipID, resourceID string) (*models.ComputeAllocationMembershipResourceOverride, error)
	// FindByMembership returns every override recorded against the given membership.
	FindByMembership(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error)
	// FindByResource returns every override referencing the given resource.
	FindByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationMembershipResourceOverride, error)
	// Create inserts a new override within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, o *models.ComputeAllocationMembershipResourceOverride) error
	// Update replaces mutable fields of an existing override within the provided transaction.
	Update(ctx context.Context, tx *sql.Tx, o *models.ComputeAllocationMembershipResourceOverride) error
	// Delete removes an override by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// AuditEventStore defines persistence operations for the append-only log
// of audit events emitted as domain entities are created, updated, and
// deleted.
type AuditEventStore interface {
	// FindByID returns the audit event with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.AuditEvent, error)
	// FindByEntity returns every audit event recorded against the given entity,
	// ordered by event_time ascending.
	FindByEntity(ctx context.Context, entityID string) ([]models.AuditEvent, error)
	// FindByEventType returns every audit event of the given type, ordered by
	// event_time ascending.
	FindByEventType(ctx context.Context, eventType string) ([]models.AuditEvent, error)
	// ListAll returns every audit event ordered by event_time ascending.
	ListAll(ctx context.Context) ([]*models.AuditEvent, error)
	// Create inserts a new audit event within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, e *models.AuditEvent) error
	// Delete removes an audit event by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

// RoleStore covers role definitions and the privilege bundle each carries.
type RoleStore interface {
	FindByID(ctx context.Context, id string) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	List(ctx context.Context) ([]models.Role, error)
	Create(ctx context.Context, tx *sql.Tx, r *models.Role) error
	Update(ctx context.Context, tx *sql.Tx, r *models.Role) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	ListPrivileges(ctx context.Context, roleID string) ([]models.PrivilegeKey, error)
	AddPrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) error
	RemovePrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) error
	HasPrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) (bool, error)
	CountRolesGrantingPrivilege(ctx context.Context, tx *sql.Tx, privilege models.PrivilegeKey) (int, error)
}

// UserRoleStore covers role assignments. Revoke is DELETE; history lives in audit_events.
type UserRoleStore interface {
	Find(ctx context.Context, userID, roleID string) (*models.UserRole, error)
	FindForUpdate(ctx context.Context, tx *sql.Tx, userID, roleID string) (*models.UserRole, error)
	ListByUser(ctx context.Context, userID string) ([]models.UserRole, error)
	ListByRole(ctx context.Context, roleID string) ([]models.UserRole, error)
	ListUserIDsByRole(ctx context.Context, roleID string) ([]string, error)
	Create(ctx context.Context, tx *sql.Tx, r *models.UserRole) error
	Delete(ctx context.Context, tx *sql.Tx, userID, roleID string) error
	PrivilegesForUser(ctx context.Context, userID string) ([]models.PrivilegeKey, error)
	UsersHoldingPrivilege(ctx context.Context, privilege models.PrivilegeKey) ([]string, error)
}

// UserPrivilegeStore defines persistence operations for fine-grained admin
// privileges. Only active grants live in the table; revoke is DELETE. The
// full grant/revoke history is in audit_events.
type UserPrivilegeStore interface {
	// Find returns the active grant for (userID, privilege), or nil.
	Find(ctx context.Context, userID string, privilege models.PrivilegeKey) (*models.UserPrivilege, error)
	// FindForUpdate returns the active grant inside a tx with SELECT FOR
	// UPDATE so the caller can serialize grant / revoke decisions.
	FindForUpdate(ctx context.Context, tx *sql.Tx, userID string, privilege models.PrivilegeKey) (*models.UserPrivilege, error)
	// ListByUser returns every active grant held by the user.
	ListByUser(ctx context.Context, userID string) ([]models.UserPrivilege, error)
	// ListByPrivilege returns every active holder of the given privilege.
	ListByPrivilege(ctx context.Context, privilege models.PrivilegeKey) ([]models.UserPrivilege, error)
	// CountByPrivilege returns the number of active holders inside a tx.
	// Used to enforce the last-meta-holder guard when revoking PrivilegeGrant.
	CountByPrivilege(ctx context.Context, tx *sql.Tx, privilege models.PrivilegeKey) (int, error)
	// Create inserts a new grant inside the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, r *models.UserPrivilege) error
	// Delete removes the grant for (userID, privilege) inside the provided tx.
	Delete(ctx context.Context, tx *sql.Tx, userID string, privilege models.PrivilegeKey) error
}

// ComputeAllocationUsageStore defines persistence operations for the
// append-only log of resource consumption events charged against a compute
// allocation.
type ComputeAllocationUsageStore interface {
	// FindByID returns the usage with the given ID, or nil if it does not exist.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationUsage, error)
	// FindByAllocation returns every usage event recorded against the given
	// allocation, ordered by calculated_time ascending.
	FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationUsage, error)
	// FindByUser returns every usage event attributed to the given user,
	// ordered by calculated_time ascending.
	FindByUser(ctx context.Context, userID string) ([]models.ComputeAllocationUsage, error)
	// FindByComputeAllocationIDAndJobID returns the usage event for the given
	// compute allocation ID and job ID, or nil if it does not exist.
	FindByComputeAllocationIDAndJobID(ctx context.Context, allocationID, jobID string) (*models.ComputeAllocationUsage, error)
	// SumSUForAllocation returns the total SUs consumed against the given
	// allocation across all usage events.
	SumSUForAllocation(ctx context.Context, allocationID string) (int64, error)
	// SumSUForUserInAllocation returns the total SUs consumed by the given
	// user against the given allocation.
	SumSUForUserInAllocation(ctx context.Context, allocationID, userID string) (int64, error)
	// Create inserts a new usage event within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, u *models.ComputeAllocationUsage) error
	// Delete removes a usage event by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}
