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

// ProjectHandler handles project lifecycle events with a typed payload.
type ProjectHandler func(ctx context.Context, project models.Project)

// SubscribeProjectCreated registers a typed handler invoked whenever a
// project::create event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectCreated(handler ProjectHandler) {
	b.subscribeProject(ProjectCreateEvent, handler)
}

// SubscribeProjectUpdated registers a typed handler invoked whenever a
// project::update event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectUpdated(handler ProjectHandler) {
	b.subscribeProject(ProjectUpdateEvent, handler)
}

// SubscribeProjectDeleted registers a typed handler invoked whenever a
// project::delete event is published. Events with payloads that are not a
// models.Project (or *models.Project) are dropped with a warning log.
func (b *Bus) SubscribeProjectDeleted(handler ProjectHandler) {
	b.subscribeProject(ProjectDeleteEvent, handler)
}

func (b *Bus) subscribeProject(topic EventType, handler ProjectHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch p := value.(type) {
		case models.Project:
			handler(ctx, p)
		case *models.Project:
			if p != nil {
				handler(ctx, *p)
			}
		default:
			slog.Warn("project event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
