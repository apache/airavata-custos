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

func TestMarkComputeClusterUserProvisioned_RoundTrip(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{
		Name: "provisioned-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	org, err := svc.CreateOrganization(ctx(), &models.Organization{
		OriginatedID: uuid.NewString(),
		Name:         "provisioned-test-org",
	})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	user, err := svc.CreateUser(ctx(), &models.User{
		OrganizationID: org.ID,
		FirstName:      "Prov",
		LastName:       "Target",
		Email:          fmt.Sprintf("prov-%s@example.invalid", uuid.NewString()),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	ccu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           user.ID,
		LocalUsername:    "prov-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create compute cluster user: %v", err)
	}
	if ccu.ProvisionedAt != nil {
		t.Fatalf("new mapping should not be provisioned yet")
	}

	if err := svc.MarkComputeClusterUserProvisioned(ctx(), ccu.ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}

	got, err := svc.GetComputeClusterUser(ctx(), ccu.ID)
	if err != nil {
		t.Fatalf("get compute cluster user: %v", err)
	}
	if got.ProvisionedAt == nil {
		t.Fatalf("provisioned_at should be set after MarkComputeClusterUserProvisioned")
	}

	if err := svc.MarkComputeClusterUserProvisioned(ctx(), "missing-"+uuid.NewString()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for unknown id, got %v", err)
	}
}
