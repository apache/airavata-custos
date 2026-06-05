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

// UserIdentityHandler handles user-identity lifecycle events with a typed payload.
type UserIdentityHandler func(ctx context.Context, identity models.UserIdentity)

// SubscribeUserIdentityCreated registers a typed handler invoked whenever a
// user_identity::create event is published.
func (b *Bus) SubscribeUserIdentityCreated(handler UserIdentityHandler) {
	b.subscribeUserIdentity(UserIdentityCreateEvent, handler)
}

// SubscribeUserIdentityUpdated registers a typed handler invoked whenever a
// user_identity::update event is published.
func (b *Bus) SubscribeUserIdentityUpdated(handler UserIdentityHandler) {
	b.subscribeUserIdentity(UserIdentityUpdateEvent, handler)
}

// SubscribeUserIdentityDeleted registers a typed handler invoked whenever a
// user_identity::delete event is published.
func (b *Bus) SubscribeUserIdentityDeleted(handler UserIdentityHandler) {
	b.subscribeUserIdentity(UserIdentityDeleteEvent, handler)
}

func (b *Bus) subscribeUserIdentity(topic EventType, handler UserIdentityHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch e := value.(type) {
		case models.UserIdentity:
			handler(ctx, e)
		case *models.UserIdentity:
			if e != nil {
				handler(ctx, *e)
			}
		default:
			slog.Warn("user identity event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
