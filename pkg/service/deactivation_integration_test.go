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

func addMember(t *testing.T, svc *Service, allocID, userID string) string {
	t.Helper()
	m, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: allocID, UserID: userID, MembershipStatus: models.ACTIVE})
	if err != nil {
		t.Fatalf("create membership: %v", err)
	}
	return m.ID
}

func membershipStatus(t *testing.T, svc *Service, id string) models.AllocationStatus {
	t.Helper()
	m, err := svc.GetComputeAllocationMembership(ctx(), id)
	if err != nil {
		t.Fatalf("get membership %s: %v", id, err)
	}
	return m.MembershipStatus
}

// Make sure an active membership cannot be deleted, deactivating it twice makes it INACTIVE, and then deleting removes it.
func TestDeactivateMembership_ThenDelete(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	_, allocID := newClusterAllocation(t, svc, pi)
	m := addMember(t, svc, allocID, seedUser(t, database, fmt.Sprintf("m-%s@example.edu", uuid.NewString())))

	if err := svc.DeleteComputeAllocationMembership(ctx(), m); !errors.Is(err, ErrInUse) {
		t.Fatalf("deleting an active membership: %v, want ErrInUse", err)
	}
	for range 2 {
		if err := svc.DeactivateComputeAllocationMembership(ctx(), m); err != nil || membershipStatus(t, svc, m) != models.INACTIVE {
			t.Fatalf("deactivate (err %v) left the membership %s, want INACTIVE", err, membershipStatus(t, svc, m))
		}
	}
	if err := svc.DeleteComputeAllocationMembership(ctx(), m); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetComputeAllocationMembership(ctx(), m); !errors.Is(err, ErrNotFound) {
		t.Fatalf("membership still there after the delete: %v", err)
	}
}

// Make sure deactivating an allocation makes it and all its memberships INACTIVE, and deleting it then removes it.
func TestDeactivateAllocation_DeactivatesItsMemberships(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	_, allocID := newClusterAllocation(t, svc, pi)
	members := []string{
		addMember(t, svc, allocID, seedUser(t, database, fmt.Sprintf("a-%s@example.edu", uuid.NewString()))),
		addMember(t, svc, allocID, seedUser(t, database, fmt.Sprintf("b-%s@example.edu", uuid.NewString()))),
	}

	if err := svc.DeactivateComputeAllocation(ctx(), allocID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if a, err := svc.GetComputeAllocation(ctx(), allocID); err != nil || a.Status != models.INACTIVE {
		t.Fatalf("allocation %+v (err %v), want INACTIVE", a, err)
	}
	for _, m := range members {
		if got := membershipStatus(t, svc, m); got != models.INACTIVE {
			t.Fatalf("membership %s is %s under an inactive allocation, want INACTIVE", m, got)
		}
	}
	if err := svc.DeleteComputeAllocation(ctx(), allocID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetComputeAllocation(ctx(), allocID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("allocation still there after the delete: %v", err)
	}
}

// Make sure deactivating an approved cluster account makes it and the user's memberships on that cluster INACTIVE, and
// leaves their memberships on other clusters alone.
func TestDeactivateClusterAccount_DeactivatesItsMembershipsOnTheCluster(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	user := seedUser(t, database, fmt.Sprintf("u-%s@example.edu", uuid.NewString()))
	clusterID, allocID := newClusterAllocation(t, svc, pi)
	_, elsewhereID := newClusterAllocation(t, svc, pi)
	here, elsewhere := addMember(t, svc, allocID, user), addMember(t, svc, elsewhereID, user)
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: clusterID, UserID: user,
		LocalUsername: "u-" + uuid.NewString()[:8], ApprovalStatus: models.ClusterAccountApproved, ReviewedBy: &pi})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}

	if err := svc.DeactivateComputeClusterUser(ctx(), cu.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if got, err := svc.GetComputeClusterUser(ctx(), cu.ID); err != nil || got.Status != models.INACTIVE {
		t.Fatalf("cluster account %+v (err %v), want INACTIVE", got, err)
	}
	if got := membershipStatus(t, svc, here); got != models.INACTIVE {
		t.Fatalf("membership on the account's cluster is %s, want INACTIVE", got)
	}
	if got := membershipStatus(t, svc, elsewhere); got != models.ACTIVE {
		t.Fatalf("membership on another cluster is %s, want ACTIVE", got)
	}
}

// Make sure a pending request cannot be deactivated but is deleted directly.
func TestDeletePendingClusterAccount_NeedsNoDeactivation(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	clusterID, _ := newClusterAllocation(t, svc, seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString())))
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: clusterID,
		UserID: seedUser(t, database, fmt.Sprintf("u-%s@example.edu", uuid.NewString())), LocalUsername: "u-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}

	if err := svc.DeactivateComputeClusterUser(ctx(), cu.ID); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("deactivating a pending request: %v, want ErrInvalidInput", err)
	}
	if err := svc.DeleteComputeClusterUser(ctx(), cu.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

// deactivatedAccount gives a new user an approved, provisioned account on a new cluster and deactivates it.
func deactivatedAccount(t *testing.T, svc *Service, database *sqlx.DB) *models.ComputeClusterUser {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "held-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID,
		UserID: seedUser(t, database, fmt.Sprintf("u-%s@example.edu", uuid.NewString())), LocalUsername: "held-" + uuid.NewString()[:8],
		ApprovalStatus: models.ClusterAccountApproved, ReviewedBy: &admin})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
		t.Fatalf("provision: %v", err)
	}
	if err := svc.DeactivateComputeClusterUser(ctx(), cu.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	return cu
}

// Make sure a returning user's request reuses their inactive account and its username as a new, unprovisioned request.
func TestCreateClusterAccount_ReusesTheInactiveRow(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	old := deactivatedAccount(t, svc, database)

	back, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: old.ComputeClusterID, UserID: old.UserID, LocalUsername: "other-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create again: %v", err)
	}
	got, err := svc.GetComputeClusterUser(ctx(), old.ID)
	if err != nil || back.ID != old.ID || got.LocalUsername != old.LocalUsername || got.Status != models.ACTIVE ||
		got.ApprovalStatus != models.ClusterAccountPending || got.ProvisionedAt != nil || got.ReviewedBy != nil {
		t.Fatalf("got %+v (err %v), want row %s reopened as a pending request named %s", got, err, old.ID, old.LocalUsername)
	}
}

// Make sure an inactive account keeps its username from anyone else, and deleting it frees the name.
func TestDeleteInactiveClusterAccount_FreesTheUsername(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	old := deactivatedAccount(t, svc, database)
	other := &models.ComputeClusterUser{ComputeClusterID: old.ComputeClusterID,
		UserID: seedUser(t, database, fmt.Sprintf("b-%s@example.edu", uuid.NewString())), LocalUsername: old.LocalUsername}

	if _, err := svc.CreateComputeClusterUser(ctx(), other); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("another user took an inactive account's username: %v", err)
	}
	if err := svc.DeleteComputeClusterUser(ctx(), old.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.CreateComputeClusterUser(ctx(), other); err != nil {
		t.Fatalf("freed username: %v", err)
	}
}

// Make sure deactivating a user makes them, their cluster accounts and those accounts' memberships INACTIVE, and
// deleting them is refused until their cluster accounts are deleted.
func TestDeactivateUser_DeactivatesTheirAccountsAndMemberships(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	user := seedUser(t, database, fmt.Sprintf("u-%s@example.edu", uuid.NewString()))
	clusterID, allocID := newClusterAllocation(t, svc, pi)
	m := addMember(t, svc, allocID, user)
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: clusterID, UserID: user,
		LocalUsername: "u-" + uuid.NewString()[:8], ApprovalStatus: models.ClusterAccountApproved, ReviewedBy: &pi})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}

	if err := svc.DeactivateUser(ctx(), user); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if u, err := svc.GetUser(ctx(), user); err != nil || u.Status != models.UserInactive {
		t.Fatalf("user %+v (err %v), want INACTIVE", u, err)
	}
	if got, err := svc.GetComputeClusterUser(ctx(), cu.ID); err != nil || got.Status != models.INACTIVE {
		t.Fatalf("cluster account %+v (err %v), want INACTIVE", got, err)
	}
	if got := membershipStatus(t, svc, m); got != models.INACTIVE {
		t.Fatalf("membership is %s under an inactive user, want INACTIVE", got)
	}
	if err := svc.DeleteUser(ctx(), user); !errors.Is(err, ErrInUse) {
		t.Fatalf("delete with a cluster account left: %v, want ErrInUse", err)
	}
	if err := svc.DeleteComputeClusterUser(ctx(), cu.ID); err != nil {
		t.Fatalf("delete cluster account: %v", err)
	}
	if err := svc.DeleteUser(ctx(), user); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetUser(ctx(), user); !errors.Is(err, ErrNotFound) {
		t.Fatalf("user still there after the delete: %v", err)
	}
}

// Make sure a temp account's VIRTUAL user, holding only a membership, is deactivated and then deleted.
func TestDeactivateThenDeleteVirtualUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	_, allocID := newClusterAllocation(t, svc, seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString())))
	user := seedUser(t, database, fmt.Sprintf("temp-%s@example.edu", uuid.NewString()))
	if _, err := database.Exec("UPDATE users SET type = $1 WHERE id = $2", models.UserTypeVirtual, user); err != nil {
		t.Fatalf("make virtual: %v", err)
	}
	addMember(t, svc, allocID, user)

	if err := svc.DeactivateUser(ctx(), user); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := svc.DeleteUser(ctx(), user); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetUser(ctx(), user); !errors.Is(err, ErrNotFound) {
		t.Fatalf("user still there after the delete: %v", err)
	}
}

// Make sure the retiring user of a merge can be deleted afterwards, leaving the survivor and its accounts.
func TestDeleteMergedUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	survivor := seedUser(t, database, fmt.Sprintf("b-%s@example.edu", uuid.NewString()))
	retiring := seedUser(t, database, fmt.Sprintf("a-%s@example.edu", uuid.NewString()))
	clusterID, _ := newClusterAllocation(t, svc, survivor)
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: clusterID, UserID: retiring, LocalUsername: "a-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatalf("merge: %v", err)
	}

	if err := svc.DeleteUser(ctx(), retiring); err != nil {
		t.Fatalf("delete merged user: %v", err)
	}
	if u, err := svc.GetUser(ctx(), survivor); err != nil || u.Status != models.UserActive {
		t.Fatalf("survivor %+v (err %v), want it active", u, err)
	}
	if got, err := svc.GetComputeClusterUser(ctx(), cu.ID); err != nil || got.UserID != survivor {
		t.Fatalf("cluster account %+v (err %v), want it moved to the survivor", got, err)
	}
}

// Make sure a user is kept while a connector has yet to handle the deletion of one of their cluster accounts.
func TestDeleteUser_WaitsForPendingAccountEvents(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserDeleted("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	clusterID, _ := newClusterAllocation(t, svc, seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString())))
	user := seedUser(t, database, fmt.Sprintf("u-%s@example.edu", uuid.NewString()))
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: clusterID, UserID: user, LocalUsername: "u-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
	if err := svc.DeleteComputeClusterUser(ctx(), cu.ID); err != nil {
		t.Fatalf("delete cluster account: %v", err)
	}
	if err := svc.DeactivateUser(ctx(), user); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	if err := svc.DeleteUser(ctx(), user); !errors.Is(err, ErrInUse) {
		t.Fatalf("delete with a pending account event: %v, want ErrInUse", err)
	}
	if _, err := database.Exec("UPDATE event_deliveries SET status = 'SUCCEEDED'"); err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if err := svc.DeleteUser(ctx(), user); err != nil {
		t.Fatalf("delete once delivered: %v", err)
	}
}
