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
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ComputeAllocationMembershipResourceOverrideHandler handles lifecycle
// events for membership resource overrides with a typed payload.
type ComputeAllocationMembershipResourceOverrideHandler func(ctx context.Context, o models.ComputeAllocationMembershipResourceOverride)

// SubscribeComputeAllocationMembershipResourceOverrideCreated registers a
// typed handler invoked whenever a
// compute_allocation_membership_resource_override::create event is published.
func (b *Bus) SubscribeComputeAllocationMembershipResourceOverrideCreated(handler ComputeAllocationMembershipResourceOverrideHandler) {
	b.subscribeMembershipResourceOverride(ComputeAllocationMembershipResourceOverrideCreateEvent, handler)
}

// SubscribeComputeAllocationMembershipResourceOverrideUpdated registers a
// typed handler invoked whenever a
// compute_allocation_membership_resource_override::update event is published.
func (b *Bus) SubscribeComputeAllocationMembershipResourceOverrideUpdated(handler ComputeAllocationMembershipResourceOverrideHandler) {
	b.subscribeMembershipResourceOverride(ComputeAllocationMembershipResourceOverrideUpdateEvent, handler)
}

// SubscribeComputeAllocationMembershipResourceOverrideDeleted registers a
// typed handler invoked whenever a
// compute_allocation_membership_resource_override::delete event is published.
func (b *Bus) SubscribeComputeAllocationMembershipResourceOverrideDeleted(handler ComputeAllocationMembershipResourceOverrideHandler) {
	b.subscribeMembershipResourceOverride(ComputeAllocationMembershipResourceOverrideDeleteEvent, handler)
}

func (b *Bus) subscribeMembershipResourceOverride(topic EventType, handler ComputeAllocationMembershipResourceOverrideHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch o := value.(type) {
		case models.ComputeAllocationMembershipResourceOverride:
			handler(ctx, o)
		case *models.ComputeAllocationMembershipResourceOverride:
			if o != nil {
				handler(ctx, *o)
			}
		default:
			slog.Warn("compute allocation membership resource override event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
