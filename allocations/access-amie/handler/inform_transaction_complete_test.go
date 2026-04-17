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
// Mocks
// ---------------------------------------------------------------------------

type mockITCAuditService struct{ mock.Mock }

func (m *mockITCAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestInformTransactionCompleteHandler_SupportsType(t *testing.T) {
	h := NewInformTransactionCompleteHandler(&mockITCAuditService{})
	assert.Equal(t, "inform_transaction_complete", h.SupportsType())
}

func TestInformTransactionCompleteHandler(t *testing.T) {
	validFixture := loadTestData(t, "inform_transaction_complete/incoming-inform.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(aud *mockITCAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet logs transaction complete",
			input: validFixture,
			setupMocks: func(aud *mockITCAuditService) {
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					model.AuditTransactionComplete, "transaction", "", mock.Anything).Return(nil)
			},
		},
		{
			name:    "missing body returns error",
			input:   map[string]any{},
			wantErr: "packet missing 'body'",
		},
		{
			name: "missing StatusCode uses Unknown as default",
			input: map[string]any{"body": map[string]any{
				"Message": "done",
			}},
			setupMocks: func(aud *mockITCAuditService) {
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					model.AuditTransactionComplete, "transaction", "",
					mock.MatchedBy(func(summary string) bool {
						return len(summary) > 0
					})).Return(nil)
			},
		},
		{
			name: "no reply is sent (terminal packet)",
			input: map[string]any{"body": map[string]any{
				"StatusCode": "Success",
				"Message":    "All good",
			}},
			setupMocks: func(aud *mockITCAuditService) {
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					model.AuditTransactionComplete, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aud := &mockITCAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(aud)
			}

			h := NewInformTransactionCompleteHandler(aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497913, Type: "inform_transaction_complete"}

			err := h.Handle(context.Background(), nil, tt.input, packet, "event-1")

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)

			// Verify the audit log was called for terminal packet cases
			if tt.setupMocks != nil {
				aud.AssertExpectations(t)
			}
		})
	}
}
