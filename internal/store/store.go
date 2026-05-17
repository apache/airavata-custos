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
	// Delete removes a project by ID within the provided transaction.
	Delete(ctx context.Context, tx *sql.Tx, id string) error
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
}

// ComputeAllocationResourceStore defines persistence operations for compute
// allocation resources (CPU, GPU, etc.).
type ComputeAllocationResourceStore interface {
	// FindByID returns the resource with the given ID, or nil if not found.
	FindByID(ctx context.Context, id string) (*models.ComputeAllocationResource, error)
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
	// FindByPair returns the mapping for a (allocation, resource) pair, or nil if absent.
	FindByPair(ctx context.Context, allocationID, resourceID string) (*models.ComputeAllocationResourceMapping, error)
	// FindResourcesByAllocation returns every resource attached to the given allocation.
	FindResourcesByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error)
	// FindAllocationsByResource returns every allocation that has the given resource attached.
	FindAllocationsByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocation, error)
	// Create inserts a new mapping within the provided transaction.
	Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error
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
