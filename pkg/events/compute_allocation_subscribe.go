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

// ComputeAllocationHandler handles compute allocation lifecycle events with a typed payload.
type ComputeAllocationHandler func(ctx context.Context, allocation models.ComputeAllocation) error

// SubscribeComputeAllocationCreated registers a typed handler invoked whenever a
// compute_allocation::create event is published. Events with payloads that are
// not a models.ComputeAllocation (or *models.ComputeAllocation) are dropped
// with a warning log.
func (b *Bus) SubscribeComputeAllocationCreated(subscriber string, handler ComputeAllocationHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationCreateEvent, handler)
}

// SubscribeComputeAllocationUpdated registers a typed handler invoked whenever a
// compute_allocation::update event is published.
func (b *Bus) SubscribeComputeAllocationUpdated(subscriber string, handler ComputeAllocationHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationUpdateEvent, handler)
}

// SubscribeComputeAllocationDeleted registers a typed handler invoked whenever a
// compute_allocation::delete event is published.
func (b *Bus) SubscribeComputeAllocationDeleted(subscriber string, handler ComputeAllocationHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationDeleteEvent, handler)
}
