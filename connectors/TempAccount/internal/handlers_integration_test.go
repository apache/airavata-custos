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

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

func TestHandlers_UpdateAllocationKeepsVirtualUser(t *testing.T) {
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	ctx := tracingtest.Context()
	bus, err := events.New(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(database, bus)
	org, _ := svc.CreateOrganization(ctx, &models.Organization{Name: "Example University"})
	cluster, _ := svc.CreateComputeCluster(ctx, &models.ComputeCluster{Name: uuid.NewString()})
	virtual, _ := svc.CreateUser(ctx, &models.User{OrganizationID: org.ID, Email: uuid.NewString() + "@example.edu", Type: models.UserTypeVirtual})
	project, _ := svc.CreateProject(ctx, &models.Project{Title: "Example Project", ProjectPIID: virtual.ID})
	alloc, _ := svc.CreateComputeAllocation(ctx, &models.ComputeAllocation{ProjectID: project.ID, Name: "example", ComputeClusterID: cluster.ID})
	m, _ := svc.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{ComputeAllocationID: alloc.ID, UserID: virtual.ID})

	rec := httptest.NewRecorder()
	NewHandlers(svc).updateAllocationToTempAccount(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"id": "`+m.ID+`", "user_id": "`+uuid.NewString()+`"}`)))
	if stored, _ := svc.GetComputeAllocationMembership(ctx, m.ID); rec.Code != http.StatusOK || stored.UserID != virtual.ID {
		t.Fatalf("status %d, stored user %s", rec.Code, stored.UserID)
	}
}

// Make sure removing a temp account keeps the user INACTIVE, deactivates its approved account and deletes its pending one.
func TestHandlers_RemoveDeactivatesApprovedAndDeletesPending(t *testing.T) {
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	ctx := tracingtest.Context()
	bus, err := events.New(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	bus.SubscribeComputeClusterUserDeleted("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	svc := service.New(database, bus)
	org, _ := svc.CreateOrganization(ctx, &models.Organization{Name: "Example University"})
	cluster, _ := svc.CreateComputeCluster(ctx, &models.ComputeCluster{Name: uuid.NewString()})
	virtual, _ := svc.CreateUser(ctx, &models.User{OrganizationID: org.ID, Email: uuid.NewString() + "@example.edu", Type: models.UserTypeVirtual})
	account, _ := svc.CreateComputeClusterUser(ctx, &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: virtual.ID, LocalUsername: "temp-" + uuid.NewString()[:8]})
	other, _ := svc.CreateComputeCluster(ctx, &models.ComputeCluster{Name: uuid.NewString()})
	approved, _ := svc.CreateComputeClusterUser(ctx, &models.ComputeClusterUser{ComputeClusterID: other.ID, UserID: virtual.ID, LocalUsername: "temp-" + uuid.NewString()[:8],
		ApprovalStatus: models.ClusterAccountApproved, ReviewedBy: &virtual.ID})

	req := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/", nil)
	req.SetPathValue("user_id", virtual.ID)
	rec := httptest.NewRecorder()
	NewHandlers(svc).removeTempAccount(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	if u, err := svc.GetUser(ctx, virtual.ID); err != nil || u.Status != models.UserInactive {
		t.Fatalf("user: got %+v, %v, want INACTIVE", u, err)
	}
	if accounts, _ := svc.ListComputeClusterUsersByUser(ctx, virtual.ID); len(accounts) != 1 || accounts[0].ID != approved.ID || accounts[0].Status != models.INACTIVE {
		t.Fatalf("cluster accounts left: %+v, want only the approved one, INACTIVE", accounts)
	}
	var deletes int
	if err := database.Get(&deletes, "SELECT COUNT(*) FROM events WHERE event_type = $1 AND payload->>'id' = $2", events.ComputeClusterUserDeleteEvent, account.ID); err != nil || deletes != 1 {
		t.Fatalf("delete events for the account: got %d, %v, want 1", deletes, err)
	}
}
