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
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

func newUser(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	return seedUser(t, database, uuid.NewString()+"@example.edu")
}

// addAccount gives the user a pending request or, with a reviewer, an approved account.
func addAccount(t *testing.T, svc *Service, clusterID, userID string, reviewer *string) *models.ComputeClusterUser {
	t.Helper()
	cu := &models.ComputeClusterUser{ComputeClusterID: clusterID, UserID: userID, LocalUsername: "u-" + uuid.NewString()[:8]}
	if reviewer != nil {
		cu.ApprovalStatus, cu.ReviewedBy = models.ClusterAccountApproved, reviewer
	}
	cu, err := svc.CreateComputeClusterUser(ctx(), cu)
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
	return cu
}

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
	pi := newUser(t, database)
	_, allocID := newClusterAllocation(t, svc, pi)
	m := addMember(t, svc, allocID, newUser(t, database))

	if err := svc.DeleteComputeAllocationMembership(ctx(), m); !errors.Is(err, ErrInUse) {
		t.Fatalf("deleting an active membership: %v, want ErrInUse", err)
	}
	for range 2 {
		if _, err := svc.UpdateMembershipStatus(ctx(), m, models.INACTIVE); err != nil || membershipStatus(t, svc, m) != models.INACTIVE {
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

// Make sure deactivating an allocation makes it and all its memberships INACTIVE.
func TestDeactivateAllocation_DeactivatesItsMemberships(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := newUser(t, database)
	_, allocID := newClusterAllocation(t, svc, pi)
	members := []string{
		addMember(t, svc, allocID, newUser(t, database)),
		addMember(t, svc, allocID, newUser(t, database)),
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
}

// Make sure deactivating an approved cluster account makes it and the user's memberships on that cluster INACTIVE, and
// leaves their memberships on other clusters alone.
func TestDeactivateClusterAccount_DeactivatesItsMembershipsOnTheCluster(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := newUser(t, database)
	user := newUser(t, database)
	clusterID, allocID := newClusterAllocation(t, svc, pi)
	_, elsewhereID := newClusterAllocation(t, svc, pi)
	here, elsewhere := addMember(t, svc, allocID, user), addMember(t, svc, elsewhereID, user)
	cu := addAccount(t, svc, clusterID, user, &pi)

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
	clusterID, _ := newClusterAllocation(t, svc, newUser(t, database))
	cu := addAccount(t, svc, clusterID, newUser(t, database), nil)

	if err := svc.DeactivateComputeClusterUser(ctx(), cu.ID); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("deactivating a pending request: %v, want ErrInvalidInput", err)
	}
	if err := svc.DeleteComputeClusterUser(ctx(), cu.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func deactivatedAccount(t *testing.T, svc *Service, database *sqlx.DB) *models.ComputeClusterUser {
	t.Helper()
	admin := newUser(t, database)
	clusterID, _ := newClusterAllocation(t, svc, admin)
	cu := addAccount(t, svc, clusterID, newUser(t, database), &admin)
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
		UserID: newUser(t, database), LocalUsername: old.LocalUsername}

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
	pi := newUser(t, database)
	user := newUser(t, database)
	clusterID, allocID := newClusterAllocation(t, svc, pi)
	m := addMember(t, svc, allocID, user)
	cu := addAccount(t, svc, clusterID, user, &pi)

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
}

// Make sure the retiring user of a merge can be deleted afterwards, leaving the survivor and its accounts.
func TestDeleteMergedUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	survivor := newUser(t, database)
	retiring := newUser(t, database)
	clusterID, _ := newClusterAllocation(t, svc, survivor)
	cu := addAccount(t, svc, clusterID, retiring, nil)
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
	clusterID, _ := newClusterAllocation(t, svc, newUser(t, database))
	user := newUser(t, database)
	cu := addAccount(t, svc, clusterID, user, nil)
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

// Make sure a user who accepted an agreement stays, inactive, since the acceptance must outlive them.
func TestDeleteUser_KeepsAUserWithAnAgreementAcceptance(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := newUser(t, database)
	admin := newUser(t, database)
	agreement := uuid.NewString()
	database.MustExec(`INSERT INTO user_agreements (id, title, version_label, body_md, body_hash, created_by) VALUES ($1, 'terms', 'v1', 'body', 'hash', $2)`, agreement, admin)
	database.MustExec(`INSERT INTO user_agreement_acceptances (id, user_id, agreement_id, body_hash) VALUES ($1, $2, $3, 'hash')`, uuid.NewString(), user, agreement)
	if err := svc.DeactivateUser(ctx(), user); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if err := svc.DeleteUser(ctx(), user); !errors.Is(err, ErrInUse) {
		t.Fatalf("delete: %v, want ErrInUse", err)
	}
}

// Make sure a merge never swaps a live account for the survivor's inactive one on the same cluster.
func TestMergeUsers_RefusesAnActiveAccountOverAnInactiveOne(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	held := deactivatedAccount(t, svc, database)
	retiring := newUser(t, database)
	addAccount(t, svc, held.ComputeClusterID, retiring, nil)
	if _, err := svc.MergeUsers(ctx(), held.UserID, retiring); !errors.Is(err, ErrInUse) {
		t.Fatalf("merge: %v, want ErrInUse", err)
	}
}

// Make sure a pending account cannot be approved for an inactive user.
func TestApproveClusterAccount_RefusesAnInactiveUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := newUser(t, database)
	admin := newUser(t, database)
	clusterID, _ := newClusterAllocation(t, svc, admin)
	cu := addAccount(t, svc, clusterID, user, nil)
	if err := svc.DeactivateUser(ctx(), user); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if _, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin); !errors.Is(err, ErrInUse) {
		t.Fatalf("approve: %v, want ErrInUse", err)
	}
}

// Make sure the status endpoint's path to INACTIVE takes the user's accounts with it.
func TestUpdateUserStatus_InactiveDeactivatesTheirAccounts(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := newUser(t, database)
	admin := newUser(t, database)
	clusterID, _ := newClusterAllocation(t, svc, admin)
	cu := addAccount(t, svc, clusterID, user, &admin)
	if _, err := svc.UpdateUserStatus(ctx(), user, models.UserInactive); err != nil {
		t.Fatalf("set inactive: %v", err)
	}
	if got, err := svc.GetComputeClusterUser(ctx(), cu.ID); err != nil || got.Status != models.INACTIVE {
		t.Fatalf("cluster account %+v (err %v), want INACTIVE", got, err)
	}
}
