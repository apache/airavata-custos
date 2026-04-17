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

package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type Router struct {
	handlers map[string]PacketHandler
	noop     PacketHandler
}

func NewRouter(handlers ...PacketHandler) *Router {
	r := &Router{handlers: make(map[string]PacketHandler)}
	for _, h := range handlers {
		t := strings.ToLower(h.SupportsType())
		if t == "*" {
			r.noop = h
		} else {
			r.handlers[t] = h
		}
	}
	return r
}

func (r *Router) Route(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	h, ok := r.handlers[strings.ToLower(packet.Type)]
	if !ok {
		if r.noop != nil {
			h = r.noop
		} else {
			return fmt.Errorf("no handler for packet type: %s", packet.Type)
		}
	}
	slog.InfoContext(ctx, "routing packet to handler", "handler", fmt.Sprintf("%T", h), "packetType", packet.Type)
	return h.Handle(ctx, tx, packetJSON, packet, eventID)
}
