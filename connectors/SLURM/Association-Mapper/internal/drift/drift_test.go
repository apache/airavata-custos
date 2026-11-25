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

package drift

import (
	"slices"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

func i64(n int64) *int64 { return &n }

var pend = &time.Time{}

func ids(marks []Mark) []string {
	var out []string
	for _, m := range marks {
		out = append(out, m.ID)
	}
	return out
}

func assoc(account, user string, qos ...string) client.Association {
	return client.Association{Account: account, User: user, Cluster: "c1", ParentAccount: "root", QoS: qos, DefaultQoS: "normal"}
}

func allocation(qos ...string) DesiredAccount {
	return DesiredAccount{AllocationID: "al1", Active: true, QoS: qos, Members: map[string]DesiredMember{
		"jdoe": {MembershipID: "m1", Active: true},
	}}
}

func wantAccount(a DesiredAccount) Desired {
	return Desired{DefaultQoS: "normal", Accounts: map[string]DesiredAccount{"bio1": a}}
}

func diff(want Desired, assocs []client.Association, qos ...client.QoS) (Push, []Conflict) {
	want.SystemAccounts = map[string]bool{"sys1": true}
	return Diff(want, Live{QoS: qos, Associations: assocs})
}

func syncedAssocs() []client.Association {
	return []client.Association{{Account: "root", Cluster: "c1", QoS: []string{"normal"}, DefaultQoS: "normal"}, assoc("bio1", "", "normal", "openfold"), assoc("bio1", "jdoe", "normal", "openfold")}
}

func conflictKinds(cs []Conflict) []Kind {
	var out []Kind
	for _, c := range cs {
		out = append(out, c.Kind)
	}
	return out
}

func find(t *testing.T, cs []Conflict, key string) Conflict {
	t.Helper()
	for _, c := range cs {
		if c.Key == key {
			return c
		}
	}
	t.Fatalf("no conflict %q in %+v", key, cs)
	return Conflict{}
}

func choices(t *testing.T, cs []Conflict, key string, want ...Choice) {
	t.Helper()
	if got := find(t, cs, key).Choices; !slices.Equal(got, want) {
		t.Fatalf("%s choices = %v, want %v", key, got, want)
	}
}

func pushed(p Push, account, user string) *client.Association {
	for i := range p.Associations {
		if p.Associations[i].Account == account && p.Associations[i].User == user {
			return &p.Associations[i]
		}
	}
	return nil
}

// A QoS only SLURM defines is reported, not deleted.
func TestSlurmOnlyQoSConflicts(t *testing.T) {
	p, cs := diff(Desired{}, nil, client.QoS{Name: "openfold", Priority: 10})
	choices(t, cs, "qos:openfold", ChoiceCustos, ChoiceSlurm)
	if len(p.QoS) != 0 || find(t, cs, "qos:openfold").Kind != KindQoSDefinition {
		t.Fatalf("push %+v conflicts %+v", p, cs)
	}
}

// A TRES SLURM has but Custos removed is pushed at -1, which clears it.
func TestRemovedTRESPushedCleared(t *testing.T) {
	want := Desired{QoS: map[string]models.ComputeClusterQoS{"gold": {Name: "gold", MaxTRESPerUser: "cpu=16", SlurmPendingSince: pend}}}
	p, _ := diff(want, nil, client.QoS{Name: "gold", MaxTRESPerUser: []client.TRES{{Type: "mem", Count: 64}}})
	if len(p.QoS) != 1 || !slices.Equal(p.QoS[0].MaxTRESPerUser, []client.TRES{{Type: "cpu", Count: 16}, {Type: "mem", Count: -1}}) {
		t.Fatalf("pushed %+v", p.QoS)
	}
}

func TestSystemAccountSkipped(t *testing.T) {
	want := Desired{Accounts: map[string]DesiredAccount{"sys1": allocation("openfold")}}
	p, cs := diff(want, []client.Association{assoc("sys1", "", "x"), assoc("sys1", "stranger", "x")})
	if len(p.Associations) != 0 || len(cs) != 0 {
		t.Fatalf("push %+v conflicts %+v", p, cs)
	}
}

// An unmarked difference is a conflict and is not pushed.
func TestUnmarkedDifferenceConflicts(t *testing.T) {
	p, cs := diff(wantAccount(allocation("openfold")), []client.Association{assoc("bio1", "", "normal", "other"), assoc("bio1", "jdoe", "normal", "other")})
	choices(t, cs, "account:bio1:qos", ChoiceCustos, ChoiceSlurm, ChoiceManual)
	if len(p.Associations) != 0 {
		t.Fatalf("pushed %+v", p.Associations)
	}
}

// A marked change pushes to the account and its members.
func TestPendingAllocationPushes(t *testing.T) {
	a := allocation("openfold")
	a.PendingSince = pend
	p, cs := diff(wantAccount(a), []client.Association{assoc("bio1", "", "normal"), assoc("bio1", "jdoe", "normal")})
	if len(cs) != 0 {
		t.Fatalf("%+v", cs)
	}
	for _, user := range []string{"", "jdoe"} {
		if got := pushed(p, "bio1", user); got == nil || !slices.Equal(got.QoS, []string{"normal", "openfold"}) {
			t.Fatalf("user %q pushed %+v", user, got)
		}
	}
}

// A push of Custos-owned limits keeps the SLURM value of a conflicting field.
func TestConflictingFieldKeptOnWrite(t *testing.T) {
	a := allocation("openfold")
	a.Limits = client.AssocLimits{GrpJobs: i64(3)}
	l := syncedAssocs()
	l[1].QoS, l[2].QoS = []string{"normal", "other"}, []string{"normal", "other"}
	p, cs := diff(wantAccount(a), l)
	got := pushed(p, "bio1", "")
	if got == nil || got.Limits.GrpJobs == nil || *got.Limits.GrpJobs != 3 || !slices.Equal(got.QoS, []string{"normal", "other"}) {
		t.Fatalf("pushed %+v", got)
	}
	if !slices.Equal(conflictKinds(cs), []Kind{KindAllocationQoS}) {
		t.Fatalf("%+v", cs)
	}
}

func TestPartitionAssociationConflicts(t *testing.T) {
	part := assoc("bio1", "jdoe", "normal", "openfold")
	part.Partition = "gpu"
	p, cs := diff(wantAccount(allocation("openfold")), append(syncedAssocs(), part))
	choices(t, cs, "partition:bio1:jdoe:gpu", ChoiceCustos)
	if c := find(t, cs, "partition:bio1:jdoe:gpu"); c.Kind != KindPartitionAssociation || c.MembershipID != "m1" || len(p.DeleteAssociations) != 0 {
		t.Fatalf("conflict %+v deletes %+v", c, p.DeleteAssociations)
	}
}

// A SLURM association without a Custos membership is reported, not deleted, even under a pending allocation.
func TestUnknownMemberConflicts(t *testing.T) {
	a := allocation("openfold")
	a.PendingSince = pend
	delete(a.Members, "jdoe")
	p, cs := diff(wantAccount(a), syncedAssocs())
	choices(t, cs, "member:bio1:jdoe:access", ChoiceCustos)
	c := find(t, cs, "member:bio1:jdoe:access")
	if len(p.DeleteAssociations) != 0 || c.Custos != "inactive" || c.Name != "access" || c.Label != "Access" || c.User != "jdoe" ||
		c.Target.Account != "bio1" || c.Target.User != "jdoe" {
		t.Fatalf("conflict %+v deletes %+v", c, p.DeleteAssociations)
	}
}

func TestPendingRevocationDeletes(t *testing.T) {
	a := allocation("openfold")
	a.Members["jdoe"] = DesiredMember{MembershipID: "m1", PendingSince: pend}
	p, cs := diff(wantAccount(a), syncedAssocs())
	if len(cs) != 0 || len(p.DeleteAssociations) != 1 || p.DeleteAssociations[0].User != "jdoe" || p.DeleteAssociations[0].Account != "bio1" {
		t.Fatalf("push %+v conflicts %+v", p, cs)
	}
}

func TestPendingClearedWhenEqual(t *testing.T) {
	a := allocation("openfold")
	a.PendingSince = pend
	a.Members["jdoe"] = DesiredMember{MembershipID: "m1", Active: true, PendingSince: pend}
	p, cs := diff(wantAccount(a), syncedAssocs())
	if len(p.Associations) != 0 || len(cs) != 0 ||
		!slices.Equal(ids(p.ClearAllocations), []string{"al1"}) || !slices.Equal(ids(p.ClearMemberships), []string{"m1"}) {
		t.Fatalf("push %+v conflicts %+v", p, cs)
	}
}
