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

// Package events defines message types emitted by the service for
// downstream consumers (audit, projections, integrations).
package events

import (
	"sync"
)

// EventType identifies the kind of event carried on the bus.
type EventType string

// Project lifecycle message types.
const (
	ProjectCreateEvent EventType = "project::create"
	ProjectUpdateEvent EventType = "project::update"
	ProjectDeleteEvent EventType = "project::delete"
)

// User lifecycle message types.
const (
	UserCreateEvent EventType = "user::create"
	UserUpdateEvent EventType = "user::update"
	UserDeleteEvent EventType = "user::delete"
)

// Organization lifecycle message types.
const (
	OrganizationCreateEvent EventType = "organization::create"
	OrganizationUpdateEvent EventType = "organization::update"
	OrganizationDeleteEvent EventType = "organization::delete"
)

// ComputeCluster lifecycle message types.
const (
	ComputeClusterCreateEvent EventType = "compute_cluster::create"
	ComputeClusterUpdateEvent EventType = "compute_cluster::update"
	ComputeClusterDeleteEvent EventType = "compute_cluster::delete"
)

// ComputeClusterUser lifecycle message types.
const (
	ComputeClusterUserCreateEvent EventType = "compute_cluster_user::create"
	ComputeClusterUserUpdateEvent EventType = "compute_cluster_user::update"
	ComputeClusterUserDeleteEvent EventType = "compute_cluster_user::delete"
)

// ClusterAccount lifecycle message types.
const (
	ClusterAccountCreateEvent EventType = "cluster_account::create"
	ClusterAccountUpdateEvent EventType = "cluster_account::update"
	ClusterAccountDeleteEvent EventType = "cluster_account::delete"
)

// ComputeAllocation lifecycle message types.
const (
	ComputeAllocationCreateEvent EventType = "compute_allocation::create"
	ComputeAllocationUpdateEvent EventType = "compute_allocation::update"
	ComputeAllocationDeleteEvent EventType = "compute_allocation::delete"
)

// CreateComputeAllocationDiff lifecycle message types.
const (
	ComputeAllocationDiffCreateEvent EventType = "compute_allocation_diff::create"
	ComputeAllocationDiffUpdateEvent EventType = "compute_allocation_diff::update"
	ComputeAllocationDiffDeleteEvent EventType = "compute_allocation_diff::delete"
)

// ComputeAllocationResource lifecycle message types.
const (
	ComputeAllocationResourceCreateEvent EventType = "compute_allocation_resource::create"
	ComputeAllocationResourceUpdateEvent EventType = "compute_allocation_resource::update"
	ComputeAllocationResourceDeleteEvent EventType = "compute_allocation_resource::delete"
)

// ComputeAllocationMembership lifecycle message types.
const (
	ComputeAllocationMembershipCreateEvent EventType = "compute_allocation_membership::create"
	ComputeAllocationMembershipUpdateEvent EventType = "compute_allocation_membership::update"
	ComputeAllocationMembershipDeleteEvent EventType = "compute_allocation_membership::delete"
)

// ComputeAllocationResourceMapping lifecycle message types.
const (
	ComputeAllocationResourceMappingCreateEvent EventType = "compute_allocation_resource_mapping::create"
	ComputeAllocationResourceMappingUpdateEvent EventType = "compute_allocation_resource_mapping::update"
	ComputeAllocationResourceMappingDeleteEvent EventType = "compute_allocation_resource_mapping::delete"
)

// Event represents a change in the system that downstream consumers may be interested in.
// The payload is the full record after the change (e.g. the
// new state of a project after an update).
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// EventSubscriberFunc is a function type that can be registered to receive events from the bus.
type EventSubscriberFunc func(event Event, value interface{})

// Bus is a lightweight, in-memory, topic-based pub/sub event bus.
// Modules publish and subscribe by topic without knowing about each other.
type Bus struct {
	mu   sync.RWMutex
	subs map[string][]EventSubscriberFunc
}
