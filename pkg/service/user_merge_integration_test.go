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
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure a merge publishes the delete event only for the retiring user's
// cluster account that it drops, which is the one on a cluster where the
// survivor already has an account. An account that moves to the survivor is
// not deleted, so it gets no event.
func TestMergeUsers_PublishesDeleteForTheDroppedClusterAccount(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserDeleted("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })

	survivor := seedUser(t, database, fmt.Sprintf("survivor-%s@example.edu", uuid.NewString()))
	retiring := seedUser(t, database, fmt.Sprintf("retiring-%s@example.edu", uuid.NewString()))
	account := func(userID string, cluster *models.ComputeCluster) *models.ComputeClusterUser {
		cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
			ComputeClusterID: cluster.ID,
			UserID:           userID,
			LocalUsername:    "merge-" + uuid.NewString()[:8],
		})
		if err != nil {
			t.Fatalf("create compute cluster user: %v", err)
		}
		return cu
	}
	cluster := func() *models.ComputeCluster {
		c, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "merge-" + uuid.NewString()[:8]})
		if err != nil {
			t.Fatalf("create cluster: %v", err)
		}
		return c
	}
	shared, onlyRetiring := cluster(), cluster()
	account(survivor, shared)
	dropped := account(retiring, shared)
	moved := account(retiring, onlyRetiring)

	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatalf("merge: %v", err)
	}

	deliveries, err := svc.EventBus().ListDeliveries(ctx(), "", 10)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if len(deliveries) != 1 || deliveries[0].Event.EventType != string(events.ComputeClusterUserDeleteEvent) {
		t.Fatalf("expected one delivery of the delete event, got %+v", deliveries)
	}
	var got models.ComputeClusterUser
	if err := json.Unmarshal(deliveries[0].Event.Payload, &got); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if got.ID != dropped.ID || got.LocalUsername != dropped.LocalUsername {
		t.Errorf("delete event is for %s (%s), want the dropped account %s (%s)", got.ID, got.LocalUsername, dropped.ID, dropped.LocalUsername)
	}

	after, err := svc.GetComputeClusterUser(ctx(), moved.ID)
	if err != nil {
		t.Fatalf("get moved account: %v", err)
	}
	if after.UserID != survivor {
		t.Errorf("moved account belongs to %s, want the survivor %s", after.UserID, survivor)
	}
}

// Make sure a merge keeps the higher tag per project and that the survivor
// holds no tag on a project it leads.
func TestMergeUsers_KeepsHigherTagAndPIHoldsNoTag(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := func(name string) string {
		return seedUser(t, database, fmt.Sprintf("%s-%s@example.edu", name, uuid.NewString()))
	}
	survivor, retiring, other := user("survivor"), user("retiring"), user("other")
	project := func(pi string, tags map[string]string) string {
		p, err := svc.CreateProject(ctx(), &models.Project{Title: "merge-" + uuid.NewString()[:8], ProjectPIID: pi})
		if err != nil {
			t.Fatalf("create project: %v", err)
		}
		for u, role := range tags {
			if err := svc.EnsureProjectMembership(ctx(), p.ID, u, role); err != nil {
				t.Fatalf("tag %s: %v", role, err)
			}
		}
		return p.ID
	}
	led := project(retiring, map[string]string{survivor: "CO_PI"})
	own := project(survivor, map[string]string{retiring: "ALLOCATION_MANAGER"})
	shared := project(other, map[string]string{retiring: "CO_PI", survivor: "ALLOCATION_MANAGER"})

	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatalf("merge: %v", err)
	}

	if p, err := svc.GetProject(ctx(), led); err != nil || p.ProjectPIID != survivor {
		t.Errorf("led project: %+v (%v), want PI %s", p, err, survivor)
	}
	for p, want := range map[string]string{led: "", own: "", shared: survivor + ":CO_PI"} {
		tags, err := svc.ListProjectMemberships(ctx(), p)
		got := ""
		for _, m := range tags {
			got += m.UserID + ":" + string(m.Role)
		}
		if err != nil || got != want {
			t.Errorf("project %s tags: %q (%v), want %q", p, got, err, want)
		}
	}
}
