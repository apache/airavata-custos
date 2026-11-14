// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

// StorageHandler handles a request about a storage with a typed payload.
type StorageHandler func(ctx context.Context, storage models.Storage) error

// SubscribeStorageCreated registers a typed handler for a storage to make.
func (b *Bus) SubscribeStorageCreated(subscriber string, handler StorageHandler) {
	subscribeTyped(b, subscriber, StorageCreateEvent, handler)
}

// SubscribeStorageUpdated registers a typed handler for a change to a storage's state.
func (b *Bus) SubscribeStorageUpdated(subscriber string, handler StorageHandler) {
	subscribeTyped(b, subscriber, StorageUpdateEvent, handler)
}

// SubscribeStorageDeleted registers a typed handler for an administrator deleting a pending storage.
func (b *Bus) SubscribeStorageDeleted(subscriber string, handler StorageHandler) {
	subscribeTyped(b, subscriber, StorageDeleteEvent, handler)
}
