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

// OrganizationHandler handles organization lifecycle events with a typed payload.
type OrganizationHandler func(ctx context.Context, organization models.Organization) error

// SubscribeOrganizationCreated registers a typed handler invoked whenever an
// organization::create event is published. Events with payloads that are not a
// models.Organization (or *models.Organization) are dropped with a warning log.
func (b *Bus) SubscribeOrganizationCreated(subscriber string, handler OrganizationHandler) {
	subscribeTyped(b, subscriber, OrganizationCreateEvent, handler)
}

// SubscribeOrganizationUpdated registers a typed handler invoked whenever an
// organization::update event is published.
func (b *Bus) SubscribeOrganizationUpdated(subscriber string, handler OrganizationHandler) {
	subscribeTyped(b, subscriber, OrganizationUpdateEvent, handler)
}

// SubscribeOrganizationDeleted registers a typed handler invoked whenever an
// organization::delete event is published.
func (b *Bus) SubscribeOrganizationDeleted(subscriber string, handler OrganizationHandler) {
	subscribeTyped(b, subscriber, OrganizationDeleteEvent, handler)
}
