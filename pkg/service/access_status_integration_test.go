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
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

func setupAccessStatusEnv(t *testing.T) (*accessTestEnv, string) {
	t.Helper()
	env := setupAccessEnv(t)
	for _, tbl := range []string{"access_check_events", "access_checks"} {
		if _, err := env.db.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("clear %s: %v", tbl, err)
		}
	}
	member, err := env.svc.CreateUser(ctx(), &models.User{
		OrganizationID: env.orgID,
		FirstName:      "Check",
		LastName:       "Member",
		Email:          "member-" + uuid.NewString()[:8] + "@example.invalid",
		Status:         models.UserActive,
	})
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	return env, member.ID
}

func checkEventTypes(t *testing.T, env *accessTestEnv, allocID, userID string, checkType models.AccessCheckType) []string {
	t.Helper()
	var types []string
	if err := env.db.Select(&types,
		`SELECT e.event_type FROM access_check_events e
		 JOIN access_checks c ON c.id = e.access_check_id
		 WHERE c.compute_allocation_id = ? AND c.user_id = ? AND c.check_type = ?
		 ORDER BY e.occurred_at, e.event_type`, allocID, userID, checkType); err != nil {
		t.Fatalf("select events: %v", err)
	}
	return types
}

func findCheck(t *testing.T, env *accessTestEnv, allocID, userID string, checkType models.AccessCheckType) models.AccessCheck {
	t.Helper()
	status, err := env.svc.AccessStatusForUser(ctx(), allocID, userID)
	if err != nil {
		t.Fatalf("access status: %v", err)
	}
	for _, c := range status.Checks {
		if c.CheckType == checkType {
			return c
		}
	}
	t.Fatalf("check %s missing from status", checkType)
	return models.AccessCheck{}
}

func TestRecordAccessCheck_FirstOKWritesStartedAndOnline(t *testing.T) {
	env, member := setupAccessStatusEnv(t)
	if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, true, "", false); err != nil {
		t.Fatalf("record: %v", err)
	}
	got := checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	want := []string{models.AccessCheckEventStarted, models.AccessCheckEventOnline}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("events: got %v, want %v", got, want)
	}
	c := findCheck(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	if c.Status != models.AccessCheckOK || c.LastOKAt == nil || c.FailingSince != nil {
		t.Fatalf("check: got %+v, want OK", c)
	}
}

func TestRecordAccessCheck_FirstFailureStaysPending(t *testing.T) {
	env, member := setupAccessStatusEnv(t)
	if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, false, "association missing", false); err != nil {
		t.Fatalf("record: %v", err)
	}
	got := checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	if len(got) != 1 || got[0] != models.AccessCheckEventStarted {
		t.Fatalf("events: got %v, want [STARTED] only", got)
	}
	c := findCheck(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	if c.Status != models.AccessCheckPending || c.FailingSince != nil {
		t.Fatalf("never-passed check must stay PENDING without failing_since, got %+v", c)
	}

	// Coming online later writes ONLINE exactly once.
	if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, true, "", false); err != nil {
		t.Fatalf("record ok: %v", err)
	}
	if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, true, "", false); err != nil {
		t.Fatalf("record ok again: %v", err)
	}
	got = checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	if len(got) != 2 || got[1] != models.AccessCheckEventOnline {
		t.Fatalf("events after coming online twice: got %v, want STARTED,ONLINE", got)
	}
}

func TestRecordAccessCheck_FailAndRecover(t *testing.T) {
	env, member := setupAccessStatusEnv(t)
	rec := func(ok bool, detail string) {
		t.Helper()
		if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckSignIn, ok, detail, false); err != nil {
			t.Fatalf("record: %v", err)
		}
	}
	rec(true, "")
	rec(false, "registry unreachable")
	c := findCheck(t, env, env.allocID, member, models.AccessCheckSignIn)
	if c.Status != models.AccessCheckFailing || c.FailingSince == nil {
		t.Fatalf("after failure: got %+v, want FAILING with failing_since", c)
	}
	rec(false, "registry unreachable")
	rec(true, "")
	c = findCheck(t, env, env.allocID, member, models.AccessCheckSignIn)
	if c.Status != models.AccessCheckOK || c.FailingSince != nil {
		t.Fatalf("after recovery: got %+v, want OK with nil failing_since", c)
	}
	got := checkEventTypes(t, env, env.allocID, member, models.AccessCheckSignIn)
	want := []string{
		models.AccessCheckEventStarted, models.AccessCheckEventOnline,
		models.AccessCheckEventFailed, models.AccessCheckEventRecovered,
	}
	if len(got) != len(want) {
		t.Fatalf("events: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("events: got %v, want %v", got, want)
		}
	}
}

func TestRecordAccessCheck_StuckWrittenOncePerEpisode(t *testing.T) {
	env, member := setupAccessStatusEnv(t)
	env.svc.SetAccessCheckStuckAfter(time.Nanosecond)
	rec := func(ok bool) {
		t.Helper()
		if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, ok, "x", false); err != nil {
			t.Fatalf("record: %v", err)
		}
	}
	rec(true)
	rec(false) // OK -> FAILING
	rec(false) // past threshold -> STUCK
	rec(false) // must not repeat STUCK
	got := checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	stuck := 0
	for _, e := range got {
		if e == models.AccessCheckEventStuck {
			stuck++
		}
	}
	if stuck != 1 {
		t.Fatalf("STUCK count: got %d in %v, want exactly 1", stuck, got)
	}

	// A new failure episode after recovery earns its own STUCK.
	rec(true)
	rec(false)
	rec(false)
	got = checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	stuck = 0
	for _, e := range got {
		if e == models.AccessCheckEventStuck {
			stuck++
		}
	}
	if stuck != 2 {
		t.Fatalf("STUCK per episode: got %d in %v, want 2", stuck, got)
	}
}

func TestRecordAccessCheck_InfrastructureStaysCalm(t *testing.T) {
	env, member := setupAccessStatusEnv(t)
	env.svc.SetAccessCheckStuckAfter(time.Nanosecond)
	rec := func(ok, infra bool, detail string) {
		t.Helper()
		if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, ok, detail, infra); err != nil {
			t.Fatalf("record: %v", err)
		}
	}
	// An infrastructure failure on a member with no row must not create one:
	// the synthesized state is the better answer.
	rec(false, true, "cluster API unreachable")
	status, err := env.svc.AccessStatusForUser(ctx(), env.allocID, member)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	for _, c := range status.Checks {
		if c.ID != "" {
			t.Fatalf("infrastructure first result created a row: %+v", c)
		}
	}

	// Past the threshold an infrastructure episode never earns a STUCK.
	rec(true, false, "")
	rec(false, true, "cluster API unreachable")
	rec(false, true, "cluster API unreachable")
	for _, e := range checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission) {
		if e == models.AccessCheckEventStuck {
			t.Fatalf("infrastructure episode wrote STUCK")
		}
	}

	// A member-specific failure resuming the episode escalates normally.
	rec(false, false, "association missing")
	stuck := 0
	for _, e := range checkEventTypes(t, env, env.allocID, member, models.AccessCheckJobSubmission) {
		if e == models.AccessCheckEventStuck {
			stuck++
		}
	}
	if stuck != 1 {
		t.Fatalf("member-specific failure after infra episode: stuck=%d, want 1", stuck)
	}
}

func TestAccessStatusForUser_DerivedDefaults(t *testing.T) {
	env, member := setupAccessStatusEnv(t)

	// No probe rows, no cluster account: both checks synthesized PENDING.
	status, err := env.svc.AccessStatusForUser(ctx(), env.allocID, member)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(status.Checks) != 2 || len(status.Events) != 0 {
		t.Fatalf("derived: got %d checks %d events, want 2/0", len(status.Checks), len(status.Events))
	}
	for _, c := range status.Checks {
		if c.Status != models.AccessCheckPending || c.ID != "" {
			t.Fatalf("derived check must be synthetic PENDING, got %+v", c)
		}
	}

	// Provisioned cluster account derives sign-in OK as of provisioning.
	cu, err := env.svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: env.clusterID, UserID: member, LocalUsername: "chk" + uuid.NewString()[:6],
	})
	if err != nil {
		t.Fatalf("create cluster user: %v", err)
	}
	if err := env.svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	c := findCheck(t, env, env.allocID, member, models.AccessCheckSignIn)
	if c.Status != models.AccessCheckOK || c.LastOKAt == nil {
		t.Fatalf("derived sign-in: got %+v, want OK from provisioned_at", c)
	}
	jc := findCheck(t, env, env.allocID, member, models.AccessCheckJobSubmission)
	if jc.Status != models.AccessCheckPending {
		t.Fatalf("derived job submission: got %+v, want PENDING", jc)
	}

	if _, err := env.svc.AccessStatusForUser(ctx(), uuid.NewString(), member); err != ErrNotFound {
		t.Fatalf("missing allocation: got %v, want ErrNotFound", err)
	}
}

func TestAccessStatusForAllocation_ActiveMembersOnly(t *testing.T) {
	env, member := setupAccessStatusEnv(t)
	former, err := env.svc.CreateUser(ctx(), &models.User{
		OrganizationID: env.orgID, FirstName: "Former", LastName: "Member",
		Email: "former-" + uuid.NewString()[:8] + "@example.invalid", Status: models.UserActive,
	})
	if err != nil {
		t.Fatalf("create former: %v", err)
	}
	mustMembership := func(userID string, st models.AllocationStatus) {
		t.Helper()
		if _, err := env.svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{
			ComputeAllocationID: env.allocID, UserID: userID,
			StartTime: nowUTC(), EndTime: nowUTC().Add(24 * time.Hour), MembershipStatus: st,
		}); err != nil {
			t.Fatalf("create membership: %v", err)
		}
	}
	mustMembership(member, models.ACTIVE)
	mustMembership(former.ID, models.INACTIVE)

	if err := env.svc.RecordAccessCheckResult(ctx(), env.allocID, member, models.AccessCheckJobSubmission, true, "", false); err != nil {
		t.Fatalf("record: %v", err)
	}

	rows, err := env.svc.AccessStatusForAllocation(ctx(), env.allocID)
	if err != nil {
		t.Fatalf("for allocation: %v", err)
	}
	if len(rows) != 1 || rows[0].UserID != member {
		t.Fatalf("members: got %+v, want only the ACTIVE member", rows)
	}
	if len(rows[0].Checks) != 2 {
		t.Fatalf("member checks: got %d, want 2 (real + derived)", len(rows[0].Checks))
	}
	var foundOK bool
	for _, c := range rows[0].Checks {
		if c.CheckType == models.AccessCheckJobSubmission && c.Status == models.AccessCheckOK {
			foundOK = true
		}
	}
	if !foundOK {
		t.Fatalf("job submission check not OK in %+v", rows[0].Checks)
	}
}
