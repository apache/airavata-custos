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

// ComputeAllocationDiffHandler handles compute allocation diff lifecycle events with a typed payload.
type ComputeAllocationDiffHandler func(ctx context.Context, diff models.ComputeAllocationDiff)

// SubscribeComputeAllocationDiffCreated registers a typed handler invoked whenever a
// compute_allocation_diff::create event is published. Events with payloads that are
// not a models.ComputeAllocationDiff (or *models.ComputeAllocationDiff) are dropped
// with a warning log.
func (b *Bus) SubscribeComputeAllocationDiffCreated(handler ComputeAllocationDiffHandler) {
	b.subscribeComputeAllocationDiff(ComputeAllocationDiffCreateEvent, handler)
}

// SubscribeComputeAllocationDiffUpdated registers a typed handler invoked whenever a
// compute_allocation_diff::update event is published.
func (b *Bus) SubscribeComputeAllocationDiffUpdated(handler ComputeAllocationDiffHandler) {
	b.subscribeComputeAllocationDiff(ComputeAllocationDiffUpdateEvent, handler)
}

// SubscribeComputeAllocationDiffDeleted registers a typed handler invoked whenever a
// compute_allocation_diff::delete event is published.
func (b *Bus) SubscribeComputeAllocationDiffDeleted(handler ComputeAllocationDiffHandler) {
	b.subscribeComputeAllocationDiff(ComputeAllocationDiffDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocationDiff(topic EventType, handler ComputeAllocationDiffHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch d := value.(type) {
		case models.ComputeAllocationDiff:
			handler(ctx, d)
		case *models.ComputeAllocationDiff:
			if d != nil {
				handler(ctx, *d)
			}
		default:
			slog.Warn("compute allocation diff event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
