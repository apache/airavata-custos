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
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure setting resources attaches the new ones, detaches the dropped ones, and refuses another cluster's resource.
func TestSetAllocationResources_ReplacesTheSet(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("grant-%s@example.edu", uuid.NewString()))
	clusterID, allocID := newClusterAllocation(t, svc, pi)
	otherCluster, _ := newClusterAllocation(t, svc, pi)

	resource := func(cluster string) string {
		res, err := svc.CreateComputeAllocationResource(ctx(), &models.ComputeAllocationResource{
			Name: "r-" + uuid.NewString()[:8], ResourceType: "cpu", ComputeClusterID: cluster,
		})
		if err != nil {
			t.Fatalf("create resource: %v", err)
		}
		return res.ID
	}
	a, b, c, foreign := resource(clusterID), resource(clusterID), resource(clusterID), resource(otherCluster)

	if err := svc.SetAllocationResources(ctx(), allocID, []string{a, b}); err != nil {
		t.Fatalf("first set: %v", err)
	}
	if err := svc.SetAllocationResources(ctx(), allocID, []string{b, c}); err != nil {
		t.Fatalf("second set: %v", err)
	}
	if err := svc.SetAllocationResources(ctx(), allocID, []string{foreign}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("another cluster's resource: got %v, want ErrInvalidInput", err)
	}
	attached, err := svc.ListResourcesForAllocation(ctx(), allocID)
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}
	got := map[string]bool{}
	for _, r := range attached {
		got[r.ID] = true
	}
	if len(got) != 2 || !got[b] || !got[c] {
		t.Errorf("attached: got %v, want %s and %s", got, b, c)
	}
}
