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
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

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

// newClusterUser creates a cluster, a user, and a cluster user with no approval
// given, the way an allocation source does, and returns the cluster user.
func newClusterUser(t *testing.T, svc *Service, database *sqlx.DB) *models.ComputeClusterUser {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "appr-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	userID := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           userID,
		LocalUsername:    "appr-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create compute cluster user: %v", err)
	}
	return cu
}

func countDeliveries(t *testing.T, svc *Service) int {
	t.Helper()
	rows, err := svc.EventBus().ListDeliveries(ctx(), "", 10)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	return len(rows)
}

// Make sure a cluster user created without an approval waits for an admin
// instead of being created on the cluster.
func TestCreateComputeClusterUser_WaitsForApproval(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	cu := newClusterUser(t, svc, database)
	if cu.ApprovalStatus != models.ClusterAccountPending || cu.ReviewedBy != nil {
		t.Fatalf("expected a pending row with no reviewer, got %+v", cu)
	}
}

// Make sure a cluster user an admin adds directly is approved and provisioning
// starts at once, without a separate approval step.
func TestCreateComputeClusterUser_ApprovedByAdminStartsProvisioning(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "appr-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	member := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))

	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           member,
		LocalUsername:    "appr-" + uuid.NewString()[:8],
		ApprovalStatus:   models.ClusterAccountApproved,
		ReviewedBy:       &admin,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if cu.ApprovalStatus != models.ClusterAccountApproved || cu.ReviewedAt == nil {
		t.Fatalf("expected an approved row with a review time, got %+v", cu)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the approve event, got %d", n)
	}
}

// Make sure an approval stores who approved, leaves an audit row, and publishes
// the approve event so the account gets created on the cluster.
func TestApproveComputeClusterUser_PublishesApproveEvent(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	got, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountApproved || got.ReviewedBy == nil || *got.ReviewedBy != admin || got.ReviewedAt == nil {
		t.Fatalf("expected an approved row reviewed by %s, got %+v", admin, got)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the approve event, got %d", n)
	}
	if n := countAuditEventsOfType(t, database, clusterUserAuditApproved, cu.ID); n != 1 {
		t.Fatalf("expected one approval audit row, got %d", n)
	}
}

// Make sure a denial stores who denied and the note, leaves an audit row, and
// publishes only the deny event, so the account is never created on the cluster.
func TestDenyComputeClusterUser_PublishesDenyEvent(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	noop := func(context.Context, models.ComputeClusterUser) error { return nil }
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", noop)
	svc.EventBus().SubscribeComputeClusterUserDenied("test-subscriber", noop)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	got, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "not on the collaborator list")
	if err != nil {
		t.Fatalf("deny: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountDenied || got.ReviewNote == nil || *got.ReviewNote != "not on the collaborator list" {
		t.Fatalf("expected a denied row with the note, got %+v", got)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the deny event, got %d", n)
	}
	if n := countAuditEventsOfType(t, database, clusterUserAuditDenied, cu.ID); n != 1 {
		t.Fatalf("expected one denial audit row, got %d", n)
	}
}

// Make sure the cluster account denial reason is mandatory
func TestDenyComputeClusterUser_RequiresReason(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	if _, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, ""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for a denial without a reason, got %v", err)
	}
	got, err := svc.GetComputeClusterUser(ctx(), cu.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountPending {
		t.Fatalf("expected the account to stay pending, got %s", got.ApprovalStatus)
	}
}

// Make sure an approved account cannot be approved again or denied, since it
// may already exist on the cluster, while a denied one can still be approved.
func TestReviewComputeClusterUser_ApprovalIsFinal(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	if _, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "not on the collaborator list"); err != nil {
		t.Fatalf("deny: %v", err)
	}
	if _, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin); err != nil {
		t.Fatalf("approve after deny: %v", err)
	}
	if _, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists on a second approval, got %v", err)
	}
	if _, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "not on the collaborator list"); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists denying an approved account, got %v", err)
	}
}

// Make sure a cluster user created through onboarding is approved by the
// onboarding admin and provisioning starts at once, since adding the user by hand is the approval.
func TestOnboardUser_ClusterAccountIsApproved(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	in := onboardInput(fmt.Sprintf("researcher-%s@example.edu", uuid.NewString()))
	in.Username = "onb-" + uuid.NewString()[:8]
	in.ComputeClusterID = cluster.ID
	in.OnboardedBy = onboarder

	created, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard: %v", err)
	}
	cu, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, created.ID)
	if err != nil {
		t.Fatalf("cluster user: %v", err)
	}
	if cu.ApprovalStatus != models.ClusterAccountApproved || cu.ReviewedBy == nil || *cu.ReviewedBy != onboarder {
		t.Fatalf("expected approval by the onboarding admin, got %+v", cu)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the approve event, got %d", n)
	}
}
