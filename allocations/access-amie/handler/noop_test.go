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
	"testing"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoOpHandler_SupportsType(t *testing.T) {
	h := NewNoOpHandler()
	assert.Equal(t, "*", h.SupportsType())
}

func TestNoOpHandler(t *testing.T) {
	tests := []struct {
		name       string
		packetType string
	}{
		{
			name:       "unknown packet type returns nil",
			packetType: "unknown_packet_type",
		},
		{
			name:       "any packet type returns nil without error",
			packetType: "some_other_type",
		},
		{
			name:       "empty packet type returns nil",
			packetType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewNoOpHandler()
			packet := &model.Packet{
				ID:     "test-id",
				AmieID: 12345,
				Type:   tt.packetType,
			}

			err := h.Handle(context.Background(), nil, map[string]any{}, packet, "event-1")
			require.NoError(t, err)
		})
	}
}

func TestNoOpHandler_DoesNotPanic_WithNilPacketJSON(t *testing.T) {
	h := NewNoOpHandler()
	packet := &model.Packet{ID: "test-id", Type: "unknown"}
	// The NoOp handler ignores packetJSON entirely.
	require.NotPanics(t, func() {
		_ = h.Handle(context.Background(), nil, nil, packet, "event-1")
	})
}
