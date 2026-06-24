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

// ComputeAllocationResourceMappingHandler handles compute allocation resource mapping lifecycle events with a typed payload.
type ComputeAllocationResourceMappingHandler func(ctx context.Context, mapping models.ComputeAllocationResourceMapping)

// SubscribeComputeAllocationResourceMappingCreated registers a typed handler invoked whenever a
// compute_allocation_resource_mapping::create event is published. Events with payloads that are
// not a models.ComputeAllocationResourceMapping (or *models.ComputeAllocationResourceMapping) are
// dropped with a warning log.
func (b *Bus) SubscribeComputeAllocationResourceMappingCreated(handler ComputeAllocationResourceMappingHandler) {
	b.subscribeComputeAllocationResourceMapping(ComputeAllocationResourceMappingCreateEvent, handler)
}

// SubscribeComputeAllocationResourceMappingUpdated registers a typed handler invoked whenever a
// compute_allocation_resource_mapping::update event is published.
func (b *Bus) SubscribeComputeAllocationResourceMappingUpdated(handler ComputeAllocationResourceMappingHandler) {
	b.subscribeComputeAllocationResourceMapping(ComputeAllocationResourceMappingUpdateEvent, handler)
}

// SubscribeComputeAllocationResourceMappingDeleted registers a typed handler invoked whenever a
// compute_allocation_resource_mapping::delete event is published.
func (b *Bus) SubscribeComputeAllocationResourceMappingDeleted(handler ComputeAllocationResourceMappingHandler) {
	b.subscribeComputeAllocationResourceMapping(ComputeAllocationResourceMappingDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocationResourceMapping(topic EventType, handler ComputeAllocationResourceMappingHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch m := value.(type) {
		case models.ComputeAllocationResourceMapping:
			handler(ctx, m)
		case *models.ComputeAllocationResourceMapping:
			if m != nil {
				handler(ctx, *m)
			}
		default:
			slog.Warn("compute allocation resource mapping event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
