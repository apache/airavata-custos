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

// ComputeAllocationResourceHandler handles compute allocation resource lifecycle events with a typed payload.
type ComputeAllocationResourceHandler func(ctx context.Context, resource models.ComputeAllocationResource) error

// SubscribeComputeAllocationResourceCreated registers a typed handler invoked whenever a
// compute_allocation_resource::create event is published. Events with payloads that are
// not a models.ComputeAllocationResource (or *models.ComputeAllocationResource) are dropped
// with a warning log.
func (b *Bus) SubscribeComputeAllocationResourceCreated(subscriber string, handler ComputeAllocationResourceHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationResourceCreateEvent, handler)
}

// SubscribeComputeAllocationResourceUpdated registers a typed handler invoked whenever a
// compute_allocation_resource::update event is published.
func (b *Bus) SubscribeComputeAllocationResourceUpdated(subscriber string, handler ComputeAllocationResourceHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationResourceUpdateEvent, handler)
}

// SubscribeComputeAllocationResourceDeleted registers a typed handler invoked whenever a
// compute_allocation_resource::delete event is published.
func (b *Bus) SubscribeComputeAllocationResourceDeleted(subscriber string, handler ComputeAllocationResourceHandler) {
	subscribeTyped(b, subscriber, ComputeAllocationResourceDeleteEvent, handler)
}
