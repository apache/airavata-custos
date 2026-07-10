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

// ComputeAllocationMembershipHandler handles compute allocation membership lifecycle events with a typed payload.
type ComputeAllocationMembershipHandler func(ctx context.Context, membership models.ComputeAllocationMembership)

// SubscribeComputeAllocationMembershipCreated registers a typed handler invoked whenever a
// compute_allocation_membership::create event is published. Events with payloads that are
// not a models.ComputeAllocationMembership (or *models.ComputeAllocationMembership) are
// dropped with a warning log.
func (b *Bus) SubscribeComputeAllocationMembershipCreated(handler ComputeAllocationMembershipHandler) {
	b.subscribeComputeAllocationMembership(ComputeAllocationMembershipCreateEvent, handler)
}

// SubscribeComputeAllocationMembershipUpdated registers a typed handler invoked whenever a
// compute_allocation_membership::update event is published.
func (b *Bus) SubscribeComputeAllocationMembershipUpdated(handler ComputeAllocationMembershipHandler) {
	b.subscribeComputeAllocationMembership(ComputeAllocationMembershipUpdateEvent, handler)
}

// SubscribeComputeAllocationMembershipDeleted registers a typed handler invoked whenever a
// compute_allocation_membership::delete event is published.
func (b *Bus) SubscribeComputeAllocationMembershipDeleted(handler ComputeAllocationMembershipHandler) {
	b.subscribeComputeAllocationMembership(ComputeAllocationMembershipDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocationMembership(topic EventType, handler ComputeAllocationMembershipHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch m := value.(type) {
		case models.ComputeAllocationMembership:
			handler(ctx, m)
		case *models.ComputeAllocationMembership:
			if m != nil {
				handler(ctx, *m)
			}
		default:
			slog.Warn("compute allocation membership event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
