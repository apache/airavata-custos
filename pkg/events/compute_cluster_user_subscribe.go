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

// ComputeClusterUserHandler handles compute-cluster user lifecycle events
// with a typed payload.
type ComputeClusterUserHandler func(ctx context.Context, user models.ComputeClusterUser)

// SubscribeComputeClusterUserCreated registers a typed handler invoked
// whenever a compute_cluster_user::create event is published.
func (b *Bus) SubscribeComputeClusterUserCreated(handler ComputeClusterUserHandler) {
	b.subscribeComputeClusterUser(ComputeClusterUserCreateEvent, handler)
}

// SubscribeComputeClusterUserUpdated registers a typed handler invoked
// whenever a compute_cluster_user::update event is published.
func (b *Bus) SubscribeComputeClusterUserUpdated(handler ComputeClusterUserHandler) {
	b.subscribeComputeClusterUser(ComputeClusterUserUpdateEvent, handler)
}

// SubscribeComputeClusterUserDeleted registers a typed handler invoked
// whenever a compute_cluster_user::delete event is published.
func (b *Bus) SubscribeComputeClusterUserDeleted(handler ComputeClusterUserHandler) {
	b.subscribeComputeClusterUser(ComputeClusterUserDeleteEvent, handler)
}

func (b *Bus) subscribeComputeClusterUser(topic EventType, handler ComputeClusterUserHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch u := value.(type) {
		case models.ComputeClusterUser:
			handler(ctx, u)
		case *models.ComputeClusterUser:
			if u != nil {
				handler(ctx, *u)
			}
		default:
			slog.Warn("compute cluster user event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
