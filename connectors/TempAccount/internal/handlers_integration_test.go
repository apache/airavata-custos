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
