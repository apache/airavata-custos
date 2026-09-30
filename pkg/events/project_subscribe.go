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

// ProjectHandler handles project lifecycle events with a typed payload.
type ProjectHandler func(ctx context.Context, project models.Project) error

// SubscribeProjectCreated registers a typed handler invoked whenever a
// project::create event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectCreated(subscriber string, handler ProjectHandler) {
	subscribeTyped(b, subscriber, ProjectCreateEvent, handler)
}

// SubscribeProjectUpdated registers a typed handler invoked whenever a
// project::update event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectUpdated(subscriber string, handler ProjectHandler) {
	subscribeTyped(b, subscriber, ProjectUpdateEvent, handler)
}

// SubscribeProjectDeleted registers a typed handler invoked whenever a
// project::delete event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectDeleted(subscriber string, handler ProjectHandler) {
	subscribeTyped(b, subscriber, ProjectDeleteEvent, handler)
}
