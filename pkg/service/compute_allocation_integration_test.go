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
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure an admin status change records one diff, and setting the same status records none.
func TestUpdateComputeAllocationStatus_RecordsDiff(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	_, allocID := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))
	for range 2 {
		if _, err := svc.UpdateComputeAllocationStatus(ctx(), allocID, models.INACTIVE); err != nil {
			t.Fatalf("update status: %v", err)
		}
	}
	diffs, err := svc.ListDiffsForAllocation(ctx(), allocID)
	if err != nil || len(diffs) != 1 || diffs[0].DiffType != "ALLOCATION_STATUS_CHANGE" || diffs[0].Status != models.INACTIVE {
		t.Fatalf("diffs: %+v, %v; want one INACTIVE ALLOCATION_STATUS_CHANGE", diffs, err)
	}
	if _, err := svc.UpdateComputeAllocationStatus(ctx(), allocID, models.DELETED); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("status DELETED: %v, want ErrInvalidInput", err)
	}
}
