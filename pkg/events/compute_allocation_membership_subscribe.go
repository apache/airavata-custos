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

// ComputeAllocationMembershipHandler handles compute allocation membership lifecycle events with a typed payload.
type ComputeAllocationMembershipHandler func(ctx context.Context, membership models.ComputeAllocationMembership) error

// SubscribeComputeAllocationMembershipReady registers a typed handler invoked whenever a
// compute_allocation_membership::ready event is published, which is when core decides the membership grants access.
func (b *Bus) SubscribeComputeAllocationMembershipReady(subscriber string, handler ComputeAllocationMembershipHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationMembershipReadyEvent, handler)
}

// SubscribeComputeAllocationMembershipDeactivated registers a typed handler invoked whenever a
// compute_allocation_membership::deactivate event is published, which is when a ready membership stops granting access.
func (b *Bus) SubscribeComputeAllocationMembershipDeactivated(subscriber string, handler ComputeAllocationMembershipHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationMembershipDeactivateEvent, handler)
}

// SubscribeComputeAllocationMembershipStorageRequested registers a typed handler invoked whenever a
// compute_allocation_membership::storage_request event is published, which is when a ready member's allocation group is provisioned.
func (b *Bus) SubscribeComputeAllocationMembershipStorageRequested(subscriber string, handler ComputeAllocationMembershipHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationMembershipStorageRequestEvent, handler)
}
