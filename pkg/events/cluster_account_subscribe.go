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
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ClusterAccountHandler handles cluster account lifecycle events with a typed payload.
type ClusterAccountHandler func(a models.ClusterAccount)

// SubscribeClusterAccountCreated registers a typed handler invoked whenever a
// cluster_account::create event is published.
func (b *Bus) SubscribeClusterAccountCreated(handler ClusterAccountHandler) {
	b.subscribeClusterAccount(ClusterAccountCreateEvent, handler)
}

// SubscribeClusterAccountUpdated registers a typed handler invoked whenever a
// cluster_account::update event is published.
func (b *Bus) SubscribeClusterAccountUpdated(handler ClusterAccountHandler) {
	b.subscribeClusterAccount(ClusterAccountUpdateEvent, handler)
}

// SubscribeClusterAccountDeleted registers a typed handler invoked whenever a
// cluster_account::delete event is published.
func (b *Bus) SubscribeClusterAccountDeleted(handler ClusterAccountHandler) {
	b.subscribeClusterAccount(ClusterAccountDeleteEvent, handler)
}

func (b *Bus) subscribeClusterAccount(topic EventType, handler ClusterAccountHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch a := value.(type) {
		case models.ClusterAccount:
			handler(a)
		case *models.ClusterAccount:
			if a != nil {
				handler(*a)
			}
		default:
			slog.Warn("cluster account event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
