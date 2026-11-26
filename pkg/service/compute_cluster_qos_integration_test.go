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
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestComputeClusterQoS_CreateMarksPending(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	clusterID, _ := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))

	got, err := svc.CreateComputeClusterQoS(ctx(), &models.ComputeClusterQoS{
		ComputeClusterID: clusterID, Name: "normal", Priority: 10, Flags: []string{"DenyOnLimit"},
	})
	if err != nil {
		t.Fatalf("create qos: %v", err)
	}
	if pendingSince(t, svc, SlurmPendingQoS, got.ID) == nil {
		t.Fatalf("new QoS should be pending a write to SLURM")
	}
	list, err := svc.ListComputeClusterQoS(ctx(), clusterID, "")
	if err != nil || len(list) != 1 {
		t.Fatalf("list qos: %v %v", list, err)
	}
	if list[0].Priority != 10 || len(list[0].Flags) != 1 || list[0].Flags[0] != "DenyOnLimit" {
		t.Fatalf("stored QoS differs: %+v", list[0])
	}
}

// A rename marks the QoS's grantees; a default change marks every allocation on the cluster.
func TestComputeClusterQoS_DefaultAndRenameMarkAllocations(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, uuid.NewString()+"@example.edu")
	clusterID, granting := newClusterAllocation(t, svc, pi)
	other := newAllocation(t, svc, clusterID, pi)
	normal := newQoS(t, svc, clusterID, "normal")
	gold := newQoS(t, svc, clusterID, "gold")
	if err := svc.SetComputeAllocationQoS(ctx(), granting, []string{"gold"}); err != nil {
		t.Fatal(err)
	}
	pending := func() []bool {
		return []bool{pendingSince(t, svc, SlurmPendingAllocation, granting) != nil, pendingSince(t, svc, SlurmPendingAllocation, other) != nil}
	}
	clearBoth := func() {
		clearPending(t, svc, SlurmPendingAllocation, granting)
		clearPending(t, svc, SlurmPendingAllocation, other)
	}

	clearBoth()
	gold.Name = "platinum"
	if _, err := svc.UpdateComputeClusterQoS(ctx(), gold); err != nil {
		t.Fatal(err)
	}
	if got := pending(); !slices.Equal(got, []bool{true, false}) {
		t.Fatalf("rename: pending = %v, want only the granting allocation", got)
	}

	clearBoth()
	normal.IsDefault = true
	if _, err := svc.UpdateComputeClusterQoS(ctx(), normal); err != nil {
		t.Fatal(err)
	}
	if got := pending(); !slices.Equal(got, []bool{true, true}) {
		t.Fatalf("new default: pending = %v, want every allocation", got)
	}

	clearBoth()
	if _, err := svc.CreateComputeClusterQoS(ctx(), &models.ComputeClusterQoS{ComputeClusterID: clusterID, Name: "high", IsDefault: true}); err != nil {
		t.Fatal(err)
	}
	if got := pending(); !slices.Equal(got, []bool{true, true}) {
		t.Fatalf("created default: pending = %v, want every allocation", got)
	}
}

func newQoS(t *testing.T, svc *Service, clusterID, name string) *models.ComputeClusterQoS {
	t.Helper()
	q, err := svc.CreateComputeClusterQoS(ctx(), &models.ComputeClusterQoS{ComputeClusterID: clusterID, Name: name})
	if err != nil {
		t.Fatalf("create qos %q: %v", name, err)
	}
	return q
}

// pendingSince returns the row's pending mark, or nil when it carries none.
func pendingSince(t *testing.T, svc *Service, kind SlurmPendingKind, id string) *time.Time {
	t.Helper()
	var since *time.Time
	if err := svc.db.Get(&since, `SELECT slurm_pending_since FROM `+string(kind)+` WHERE id = $1`, id); err != nil {
		t.Fatalf("read %s mark: %v", kind, err)
	}
	return since
}

// clearPending clears the mark the row carries now, as the sweep does after reading it.
func clearPending(t *testing.T, svc *Service, kind SlurmPendingKind, id string) {
	t.Helper()
	if since := pendingSince(t, svc, kind, id); since != nil {
		if err := svc.SetSlurmPending(ctx(), kind, id, since); err != nil {
			t.Fatalf("clear %s mark: %v", kind, err)
		}
	}
}

// changeAllocationStatus records the status change as a diff, which applies it.
func changeAllocationStatus(svc *Service, id string, status models.AllocationStatus) error {
	_, err := svc.CreateComputeAllocationDiff(ctx(), &models.ComputeAllocationDiff{ComputeAllocationID: id, DiffType: "ALLOCATION_STATUS_CHANGE", Status: status})
	return err
}

// requireAllocationPending clears the mark, runs write, and fails unless write sets it again.
func requireAllocationPending(t *testing.T, svc *Service, id, what string, write func() error) {
	t.Helper()
	clearPending(t, svc, SlurmPendingAllocation, id)
	if err := write(); err != nil {
		t.Fatalf("%s: %v", what, err)
	}
	if pendingSince(t, svc, SlurmPendingAllocation, id) == nil {
		t.Errorf("%s: should be pending", what)
	}
}

func grantedQoSNames(t *testing.T, svc *Service, clusterID, allocID string) []string {
	t.Helper()
	granted, err := svc.ListComputeClusterQoS(ctx(), clusterID, allocID)
	if err != nil {
		t.Fatalf("list granted qos: %v", err)
	}
	names := []string{}
	for _, q := range granted {
		names = append(names, q.Name)
	}
	return names
}

func newClusterResource(t *testing.T, svc *Service, clusterID string) string {
	t.Helper()
	res, err := svc.CreateComputeAllocationResource(ctx(), &models.ComputeAllocationResource{
		Name: "r-" + uuid.NewString()[:8], ResourceType: "cpu", ComputeClusterID: clusterID,
	})
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}
	return res.ID
}

func newMembership(t *testing.T, svc *Service, allocationID, userID string) *models.ComputeAllocationMembership {
	t.Helper()
	m, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: allocationID, UserID: userID})
	if err != nil {
		t.Fatalf("create membership: %v", err)
	}
	return m
}

func TestSetComputeAllocationQoS_MarksAllocationPending(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	clusterID, allocID := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))
	newQoS(t, svc, clusterID, "normal")
	newQoS(t, svc, clusterID, "fast")
	newQoS(t, svc, clusterID, "unused")

	requireAllocationPending(t, svc, allocID, "set qos", func() error {
		return svc.SetComputeAllocationQoS(ctx(), allocID, []string{"normal", "fast"})
	})
	if got := grantedQoSNames(t, svc, clusterID, allocID); !slices.Equal(got, []string{"fast", "normal"}) {
		t.Fatalf("qos names: %v; want [fast normal]", got)
	}

	requireAllocationPending(t, svc, allocID, "replace qos", func() error {
		return svc.SetComputeAllocationQoS(ctx(), allocID, []string{"normal"})
	})
	if got := grantedQoSNames(t, svc, clusterID, allocID); !slices.Equal(got, []string{"normal"}) {
		t.Fatalf("qos names after replace: %v, want [normal]", got)
	}
}

func TestDeleteComputeClusterQoS_RefusedWhileUsed(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	clusterID, allocID := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))
	q := newQoS(t, svc, clusterID, "normal")
	if err := svc.SetComputeAllocationQoS(ctx(), allocID, []string{"normal"}); err != nil {
		t.Fatalf("set qos: %v", err)
	}

	if err := svc.DeleteComputeClusterQoS(ctx(), clusterID, q.ID); !errors.Is(err, ErrInUse) {
		t.Fatalf("delete a used qos: %v, want ErrInUse", err)
	}
	if err := svc.SetComputeAllocationQoS(ctx(), allocID, nil); err != nil {
		t.Fatalf("clear qos: %v", err)
	}
	if err := svc.DeleteComputeClusterQoS(ctx(), clusterID, q.ID); err != nil {
		t.Fatalf("delete an unused qos: %v", err)
	}
}

func TestClearPending_KeepsAMarkSetAfterTheRead(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	clusterID, allocID := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))
	m := newMembership(t, svc, allocID, seedUser(t, database, uuid.NewString()+"@example.edu"))
	q := newQoS(t, svc, clusterID, "normal")
	if err := changeAllocationStatus(svc, allocID, models.INACTIVE); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateMembershipStatus(ctx(), m.ID, models.INACTIVE); err != nil {
		t.Fatal(err)
	}
	marks := map[SlurmPendingKind]string{SlurmPendingAllocation: allocID, SlurmPendingMembership: m.ID, SlurmPendingQoS: q.ID}
	read := map[SlurmPendingKind]*time.Time{}
	for kind, id := range marks {
		if read[kind] = pendingSince(t, svc, kind, id); read[kind] == nil {
			t.Fatalf("%s: should be pending", kind)
		}
	}

	if err := changeAllocationStatus(svc, allocID, models.ACTIVE); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateMembershipStatus(ctx(), m.ID, models.ACTIVE); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateComputeClusterQoS(ctx(), q); err != nil {
		t.Fatal(err)
	}
	for kind, id := range marks {
		if err := svc.SetSlurmPending(ctx(), kind, id, read[kind]); err != nil {
			t.Fatal(err)
		}
		if pendingSince(t, svc, kind, id) == nil {
			t.Fatalf("%s: a mark set after the read must survive the clear", kind)
		}
		if clearPending(t, svc, kind, id); pendingSince(t, svc, kind, id) != nil {
			t.Fatalf("%s: the current mark should clear", kind)
		}
	}
}

func TestDelete_RefusedWhileRevokePending(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	_, allocID := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))
	m := newMembership(t, svc, allocID, seedUser(t, database, uuid.NewString()+"@example.edu"))
	for _, tc := range []struct {
		kind       SlurmPendingKind
		id         string
		deactivate func() error
		delete     func() error
	}{
		{SlurmPendingMembership, m.ID,
			func() error { _, err := svc.UpdateMembershipStatus(ctx(), m.ID, models.INACTIVE); return err },
			func() error { return svc.DeleteComputeAllocationMembership(ctx(), m.ID) }},
		{SlurmPendingAllocation, allocID,
			func() error { return changeAllocationStatus(svc, allocID, models.INACTIVE) },
			func() error { return svc.DeleteComputeAllocation(ctx(), allocID) }},
	} {
		if err := tc.delete(); !errors.Is(err, ErrInUse) {
			t.Fatalf("%s: delete while active: %v, want ErrInUse", tc.kind, err)
		}
		if err := tc.deactivate(); err != nil {
			t.Fatalf("%s: deactivate: %v", tc.kind, err)
		}
		if err := tc.delete(); !errors.Is(err, ErrInUse) {
			t.Fatalf("%s: delete with the revoke pending: %v, want ErrInUse", tc.kind, err)
		}
		clearPending(t, svc, tc.kind, tc.id)
		if err := tc.delete(); err != nil {
			t.Fatalf("%s: delete once applied: %v", tc.kind, err)
		}
	}
}

// A save of what SLURM already holds sets no mark, so the sync its update event triggers finds no drift.
// The event still goes out for the other connectors.
func TestWithoutSlurmMark_SetsNoMarkAndPublishesUpdateEvents(t *testing.T) {
	database := setupTestDB(t)
	bus := newTestBus(database)
	svc := New(database, bus)
	// An event is stored only when something subscribes to it.
	bus.SubscribeComputeAllocationResourceMappingUpdated("test", func(context.Context, models.ComputeAllocationResourceMapping) error { return nil })
	bus.SubscribeComputeAllocationMembershipResourceOverrideUpdated("test", func(context.Context, models.ComputeAllocationMembershipResourceOverride) error { return nil })
	bus.SubscribeComputeAllocationMembershipUpdated("test", func(context.Context, models.ComputeAllocationMembership) error { return nil })
	clusterID, allocID := newClusterAllocation(t, svc, seedUser(t, database, uuid.NewString()+"@example.edu"))
	m := newMembership(t, svc, allocID, seedUser(t, database, uuid.NewString()+"@example.edu"))
	resourceID := newClusterResource(t, svc, clusterID)
	if _, err := svc.AttachResourceToAllocation(ctx(), allocID, resourceID, 1, 1); err != nil {
		t.Fatal(err)
	}
	o, err := svc.CreateComputeAllocationMembershipResourceOverride(ctx(), &models.ComputeAllocationMembershipResourceOverride{
		ComputeAllocationMembershipID: m.ID, ComputeAllocationResourceID: resourceID, OverrideResourceAmount: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	q := newQoS(t, svc, clusterID, "normal")
	if _, err := svc.UpdateMembershipStatus(ctx(), m.ID, models.INACTIVE); err != nil {
		t.Fatal(err)
	}
	clearPending(t, svc, SlurmPendingAllocation, allocID)
	clearPending(t, svc, SlurmPendingMembership, m.ID)
	clearPending(t, svc, SlurmPendingQoS, q.ID)
	const events = "SELECT COUNT(*) FROM events"
	before := countRows(t, database, events)
	fromSlurm := WithoutSlurmMark(ctx())

	o.OverrideResourceAmount = 7
	if _, err := svc.UpdateComputeAllocationMembershipResourceOverride(fromSlurm, o); err != nil {
		t.Fatalf("save override: %v", err)
	}
	if _, err := svc.UpdateAllocationResourceMapping(fromSlurm, allocID, resourceID, 9, 3); err != nil {
		t.Fatalf("save mapping: %v", err)
	}
	if _, err := svc.UpdateMembershipStatus(fromSlurm, m.ID, models.ACTIVE); err != nil {
		t.Fatalf("save status: %v", err)
	}
	q.Priority = 4
	if _, err := svc.UpdateComputeClusterQoS(fromSlurm, q); err != nil {
		t.Fatalf("save qos: %v", err)
	}
	if _, err := svc.CreateComputeClusterQoS(fromSlurm, &models.ComputeClusterQoS{ComputeClusterID: clusterID, Name: "fast"}); err != nil {
		t.Fatalf("save new qos: %v", err)
	}

	if pendingSince(t, svc, SlurmPendingAllocation, allocID) != nil || pendingSince(t, svc, SlurmPendingMembership, m.ID) != nil {
		t.Errorf("a save from SLURM must leave no mark")
	}
	list, _ := svc.ListComputeClusterQoS(ctx(), clusterID, "")
	for _, got := range list {
		if got.SlurmPendingSince != nil {
			t.Errorf("qos %s is pending", got.Name)
		}
	}
	if mappings, _ := svc.ListResourceMappingsForAllocation(ctx(), allocID); len(mappings) != 1 || mappings[0].ResourceAmount != 9 || mappings[0].ResourceTime != 3 {
		t.Errorf("mappings = %+v", mappings)
	}
	// Mapping, override and membership updates each publish their normal event; the QoS saves have none.
	if after := countRows(t, database, events); after != before+3 {
		t.Errorf("events %d -> %d; want the 3 update events", before, after)
	}
}
