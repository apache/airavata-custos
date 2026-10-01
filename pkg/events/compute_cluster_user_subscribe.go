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

package events

import (
	"context"

	"github.com/apache/airavata-custos/pkg/models"
)

// ComputeClusterUserHandler handles compute-cluster user lifecycle events
// with a typed payload.
type ComputeClusterUserHandler func(ctx context.Context, user models.ComputeClusterUser) error

// SubscribeComputeClusterUserCreated registers a typed handler invoked
// whenever a compute_cluster_user::create event is published.
func (b *Bus) SubscribeComputeClusterUserCreated(subscriber string, handler ComputeClusterUserHandler) {
	subscribeTyped(b, subscriber, ComputeClusterUserCreateEvent, handler)
}

// SubscribeComputeClusterUserApproved registers a typed handler invoked
// whenever a compute_cluster_user::approve event is published, which is when
// an admin approves the account, and it can be created on the cluster.
func (b *Bus) SubscribeComputeClusterUserApproved(subscriber string, handler ComputeClusterUserHandler) {
	subscribeTyped(b, subscriber, ComputeClusterUserApproveEvent, handler)
}

// SubscribeComputeClusterUserDenied registers a typed handler invoked
// whenever a compute_cluster_user::deny event is published. The payload
// carries the admin's reason in ReviewNote.
func (b *Bus) SubscribeComputeClusterUserDenied(subscriber string, handler ComputeClusterUserHandler) {
	subscribeTyped(b, subscriber, ComputeClusterUserDenyEvent, handler)
}

// SubscribeComputeClusterUserUpdated registers a typed handler invoked
// whenever a compute_cluster_user::update event is published.
func (b *Bus) SubscribeComputeClusterUserUpdated(subscriber string, handler ComputeClusterUserHandler) {
	subscribeTyped(b, subscriber, ComputeClusterUserUpdateEvent, handler)
}

// SubscribeComputeClusterUserDeleted registers a typed handler invoked
// whenever a compute_cluster_user::delete event is published.
func (b *Bus) SubscribeComputeClusterUserDeleted(subscriber string, handler ComputeClusterUserHandler) {
	subscribeTyped(b, subscriber, ComputeClusterUserDeleteEvent, handler)
}
