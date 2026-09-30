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

// UserHandler handles user lifecycle events with a typed payload.
type UserHandler func(ctx context.Context, user models.User) error

// SubscribeUserCreated registers a typed handler invoked whenever a
// user::create event is published. Events with payloads that are not a
// models.User (or *models.User) are dropped with a warning log.
func (b *Bus) SubscribeUserCreated(subscriber string, handler UserHandler) {
	subscribeTyped(b, subscriber, UserCreateEvent, handler)
}

// SubscribeUserUpdated registers a typed handler invoked whenever a
// user::update event is published.
func (b *Bus) SubscribeUserUpdated(subscriber string, handler UserHandler) {
	subscribeTyped(b, subscriber, UserUpdateEvent, handler)
}

// SubscribeUserDeleted registers a typed handler invoked whenever a
// user::delete event is published.
func (b *Bus) SubscribeUserDeleted(subscriber string, handler UserHandler) {
	subscribeTyped(b, subscriber, UserDeleteEvent, handler)
}
