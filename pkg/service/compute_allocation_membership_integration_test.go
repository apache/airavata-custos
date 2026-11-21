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

	"github.com/apache/airavata-custos/pkg/models"
)

func TestDeleteComputeAllocationMembership_RefusedWhileActive(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	_, allocID := newClusterAllocation(t, svc, seedUser(t, database, "pi@example.edu"))
	m, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: allocID, UserID: seedUser(t, database, "member@example.edu"), MembershipStatus: models.ACTIVE})
	if err != nil {
		t.Fatalf("create membership: %v", err)
	}
	if err := svc.DeleteComputeAllocationMembership(ctx(), m.ID); !errors.Is(err, ErrInUse) {
		t.Fatalf("delete active membership: %v, want ErrInUse", err)
	}
}
