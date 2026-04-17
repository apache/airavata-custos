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
	"testing"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock handler used only by router tests
// ---------------------------------------------------------------------------

type mockPacketHandler struct {
	mock.Mock
}

func (m *mockPacketHandler) SupportsType() string {
	return m.Called().String(0)
}

func (m *mockPacketHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	args := m.Called(ctx, tx, packetJSON, packet, eventID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRouter(t *testing.T) {
	tests := []struct {
		name            string
		packetType      string
		wantHandler1    bool
		wantHandler2    bool
		wantNoOpInvoked bool
		wantErr         bool
	}{
		{
			name:         "routes to matching handler (handler1)",
			packetType:   "request_project_create",
			wantHandler1: true,
			wantHandler2: false,
		},
		{
			name:         "routes to different matching handler (handler2)",
			packetType:   "request_account_create",
			wantHandler1: false,
			wantHandler2: true,
		},
		{
			name:            "unknown type falls back to NoOp",
			packetType:      "unknown_packet_type",
			wantHandler1:    false,
			wantHandler2:    false,
			wantNoOpInvoked: true,
		},
		{
			name:         "case-insensitive matching routes correctly",
			packetType:   "REQUEST_PROJECT_CREATE",
			wantHandler1: true,
			wantHandler2: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := &mockPacketHandler{}
			h2 := &mockPacketHandler{}
			noOp := NewNoOpHandler()

			h1.On("SupportsType").Return("request_project_create")
			h2.On("SupportsType").Return("request_account_create")

			if tt.wantHandler1 {
				h1.On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}
			if tt.wantHandler2 {
				h2.On("Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			router := NewRouter(h1, h2, noOp)
			packet := &model.Packet{ID: "test-id", AmieID: 12345, Type: tt.packetType}

			err := router.Route(context.Background(), nil, map[string]any{}, packet, "event-1")

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.wantHandler1 {
				h1.AssertCalled(t, "Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			} else {
				h1.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}

			if tt.wantHandler2 {
				h2.AssertCalled(t, "Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			} else {
				h2.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestRouter_NoHandlerAndNoNoOp_ReturnsError(t *testing.T) {
	router := NewRouter() // no handlers at all, no noop
	packet := &model.Packet{ID: "test-id", Type: "unknown_type"}
	err := router.Route(context.Background(), nil, map[string]any{}, packet, "e1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler for packet type")
}
