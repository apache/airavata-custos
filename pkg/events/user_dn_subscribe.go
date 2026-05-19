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

// UserDNHandler handles user DN lifecycle events with a typed payload.
type UserDNHandler func(d models.UserDN)

// SubscribeUserDNCreated registers a typed handler invoked whenever a
// user_dn::create event is published.
func (b *Bus) SubscribeUserDNCreated(handler UserDNHandler) {
	b.subscribeUserDN(UserDNCreateEvent, handler)
}

// SubscribeUserDNDeleted registers a typed handler invoked whenever a
// user_dn::delete event is published.
func (b *Bus) SubscribeUserDNDeleted(handler UserDNHandler) {
	b.subscribeUserDN(UserDNDeleteEvent, handler)
}

func (b *Bus) subscribeUserDN(topic EventType, handler UserDNHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch d := value.(type) {
		case models.UserDN:
			handler(d)
		case *models.UserDN:
			if d != nil {
				handler(*d)
			}
		default:
			slog.Warn("user dn event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
