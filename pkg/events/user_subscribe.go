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

// UserHandler handles user lifecycle events with a typed payload.
type UserHandler func(ctx context.Context, user models.User)

// SubscribeUserCreated registers a typed handler invoked whenever a
// user::create event is published. Events with payloads that are not a
// models.User (or *models.User) are dropped with a warning log.
func (b *Bus) SubscribeUserCreated(handler UserHandler) {
	b.subscribeUser(UserCreateEvent, handler)
}

// SubscribeUserUpdated registers a typed handler invoked whenever a
// user::update event is published.
func (b *Bus) SubscribeUserUpdated(handler UserHandler) {
	b.subscribeUser(UserUpdateEvent, handler)
}

// SubscribeUserDeleted registers a typed handler invoked whenever a
// user::delete event is published.
func (b *Bus) SubscribeUserDeleted(handler UserHandler) {
	b.subscribeUser(UserDeleteEvent, handler)
}

func (b *Bus) subscribeUser(topic EventType, handler UserHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch u := value.(type) {
		case models.User:
			handler(ctx, u)
		case *models.User:
			if u != nil {
				handler(ctx, *u)
			}
		default:
			slog.Warn("user event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
