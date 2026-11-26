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

package subscribers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// fakeOps records the accounts it was asked to provision.
type fakeOps struct {
	provisioned []string
	admins      []string
}

func (f *fakeOps) EnsureClusterAccount(_ context.Context, cu *models.ComputeClusterUser) error {
	f.provisioned = append(f.provisioned, cu.ID)
	return nil
}
func (f *fakeOps) RemoveClusterAccount(context.Context, *models.ComputeClusterUser) error { return nil }
func (f *fakeOps) EnsureClusterAdminMembership(_ context.Context, cu *models.ComputeClusterUser) error {
	f.admins = append(f.admins, cu.ID)
	return nil
}
func (f *fakeOps) EnsureAllocationGroup(context.Context, *models.ComputeAllocation) error {
	return nil
}
func (f *fakeOps) RemoveAllocationGroup(context.Context, *models.ComputeAllocation) error {
	return nil
}

func approveWith(t *testing.T, status models.UserStatus) (*fakeOps, []models.AuditEvent) {
	t.Helper()
	var audits []models.AuditEvent
	core := &service.CoreServiceMock{
		GetUserFunc: func(context.Context, string) (*models.User, error) {
			return &models.User{ID: "u1", Status: status}, nil
		},
		CreateAuditEventFunc: func(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
			audits = append(audits, *e)
			return e, nil
		},
		ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) { return nil, nil },
	}
	ops := &fakeOps{}
	s := &Subscriber{ops: ops, core: core, custosClusterID: "c1"}
	cu := models.ComputeClusterUser{ID: "cu1", UserID: "u1", ComputeClusterID: "c1", ApprovalStatus: models.ComputeClusterUserApproved}
	if err := s.handleComputeClusterUserApproved(context.Background(), cu); err != nil {
		t.Fatal(err)
	}
	return ops, audits
}

func TestClusterUserApproved_SkipsAUserNotActive(t *testing.T) {
	ops, audits := approveWith(t, models.UserPending)
	if len(ops.provisioned) != 0 {
		t.Fatalf("provisioned a user not active: %v", ops.provisioned)
	}
	var deferred bool
	for _, a := range audits {
		if a.EventType == "ComanageProvisioningDeferred" && strings.Contains(a.Details, "reason=user_not_active") && a.EntityID == "cu1" {
			deferred = true
		}
	}
	if !deferred {
		t.Fatalf("no ComanageProvisioningDeferred audit: %+v", audits)
	}
}

func TestClusterUserApproved_RunsForAnActiveUser(t *testing.T) {
	ops, _ := approveWith(t, models.UserActive)
	if len(ops.provisioned) != 1 || ops.provisioned[0] != "cu1" {
		t.Fatalf("provisioned = %v, want [cu1]", ops.provisioned)
	}
}

func TestUserOnboarded_JoinsAdminGroupOnlyForApprovedAdminAccounts(t *testing.T) {
	signIn := func(level models.ClusterAccessLevel, approval models.ComputeClusterUserApproval) *fakeOps {
		core := &service.CoreServiceMock{
			GetComputeClusterUserByPairFunc: func(context.Context, string, string) (*models.ComputeClusterUser, error) {
				return &models.ComputeClusterUser{ID: "cu1", UserID: "u1", ComputeClusterID: "c1", AccessLevel: level, ApprovalStatus: approval}, nil
			},
			ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) { return nil, nil },
		}
		ops := &fakeOps{}
		s := &Subscriber{ops: ops, core: core, custosClusterID: "c1"}
		if err := s.handleUserOnboarded(context.Background(), models.User{ID: "u1", Status: models.UserActive}); err != nil {
			t.Fatal(err)
		}
		return ops
	}

	// An admin account joins the cluster admin group on first sign-in.
	if ops := signIn(models.ClusterAccessAdmin, models.ComputeClusterUserApproved); len(ops.admins) != 1 || ops.admins[0] != "cu1" {
		t.Fatalf("admin account: admin group calls = %v, want [cu1]", ops.admins)
	}
	// A normal account does not.
	if ops := signIn(models.ClusterAccessUser, models.ComputeClusterUserApproved); len(ops.admins) != 0 {
		t.Fatalf("normal account: admin group calls = %v, want none", ops.admins)
	}
	// An admin account still awaiting approval does not.
	if ops := signIn(models.ClusterAccessAdmin, models.ComputeClusterUserPending); len(ops.admins) != 0 {
		t.Fatalf("pending admin account: admin group calls = %v, want none", ops.admins)
	}
}

func TestUserUpdated_ProvisionsAnApprovedAccountOfAnActiveUser(t *testing.T) {
	cases := []struct {
		provisionedAt *time.Time
		approval      models.ComputeClusterUserApproval
		status        models.UserStatus
		sent          bool
	}{
		{nil, models.ComputeClusterUserApproved, models.UserActive, true},
		{nil, models.ComputeClusterUserApproved, models.UserPending, false},
		{nil, models.ComputeClusterUserPending, models.UserActive, false},
		{new(time.Time), models.ComputeClusterUserApproved, models.UserInactive, true},
	}
	for _, c := range cases {
		core := &service.CoreServiceMock{
			GetComputeClusterUserByPairFunc: func(context.Context, string, string) (*models.ComputeClusterUser, error) {
				return &models.ComputeClusterUser{ID: "cu1", UserID: "u1", ComputeClusterID: "c1", ApprovalStatus: c.approval, ProvisionedAt: c.provisionedAt}, nil
			},
			GetUserFunc: func(context.Context, string) (*models.User, error) {
				return &models.User{ID: "u1", Status: c.status}, nil
			},
			CreateAuditEventFunc:       func(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) { return e, nil },
			ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) { return nil, nil },
		}
		ops := &fakeOps{}
		s := &Subscriber{ops: ops, core: core, custosClusterID: "c1"}
		if err := s.handleUserUpdated(context.Background(), models.User{ID: "u1", Status: c.status}); err != nil {
			t.Fatal(err)
		}
		if sent := len(ops.provisioned) == 1; sent != c.sent {
			t.Errorf("%+v: provision calls = %v", c, ops.provisioned)
		}
	}
}
