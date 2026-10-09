//go:build integration

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

package service

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure approving with only change_status and approver_id keeps the
// requested amount, status and reason.
func TestApproveChangeRequestKeepsRequestedFields(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("cr-%s@example.edu", uuid.NewString()))
	_, allocID := newClusterAllocation(t, svc, pi)

	cr, err := svc.CreateComputeAllocationChangeRequest(ctx(), &models.ComputeAllocationChangeRequest{
		ComputeAllocationID: allocID, RequesterID: pi, RequestedSUAmount: 1500, RequestedStatus: models.ACTIVE, Reason: "more compute",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := svc.UpdateComputeAllocationChangeRequest(ctx(), &models.ComputeAllocationChangeRequest{ID: cr.ID, ChangeStatus: "APPROVED", ApproverID: pi})
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got.RequestedSUAmount != 1500 || got.RequestedStatus != models.ACTIVE || got.Reason != "more compute" {
		t.Fatalf("approved request lost its fields: %+v", got)
	}
}
