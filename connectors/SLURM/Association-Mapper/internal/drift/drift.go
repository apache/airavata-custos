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

// Pure diff between the SLURM state Custos wants and the state SLURM reports.
// The sweep pushes what Diff returns; the conflict routes show its conflicts.
package drift

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

type Desired struct {
	// DefaultQoS is added to every pushed QoS list; SLURM rejects lists without it.
	DefaultQoS string
	QoS        map[string]models.ComputeClusterQoS // by name
	// OtherClusterQoS names the QoS the connector's other clusters define. slurmdbd holds
	// one QoS namespace, so those are not SLURM-only here.
	OtherClusterQoS map[string]bool
	Accounts        map[string]DesiredAccount // by SLURM account name
	// SystemAccounts names the live accounts no allocation owns.
	SystemAccounts map[string]bool
}

type DesiredAccount struct {
	AllocationID string
	Active       bool
	PendingSince *time.Time
	// QoS is the complete list written to the account and every member.
	// Empty means the account takes its parent account's live list.
	QoS     []string
	Limits  client.AssocLimits
	Members map[string]DesiredMember // by local username
}

type DesiredMember struct {
	MembershipID string // "" when the user has no membership
	Active       bool
	PendingSince *time.Time
	Limits       client.AssocLimits
	Held         bool // provisioned too recently for the cluster to resolve; its writes wait
}

type Live struct {
	QoS          []client.QoS
	Associations []client.Association
	Partitions   []client.Partition // shown, never diffed
}

type Kind string

const (
	KindQoSDefinition        Kind = "qos_definition"
	KindAllocationQoS        Kind = "allocation_qos"
	KindMemberQoS            Kind = "member_qos"
	KindAllocationLimit      Kind = "allocation_limit"
	KindMemberLimit          Kind = "member_limit"
	KindPartitionAssociation Kind = "partition_association"
	KindAccess               Kind = "access"
)

type Choice string

const (
	ChoiceCustos Choice = "custos"
	ChoiceSlurm  Choice = "slurm"
	ChoiceManual Choice = "manual"
)

type Conflict struct {
	Key          string   `json:"key"`
	Kind         Kind     `json:"kind"`
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	User         string   `json:"user,omitempty"`
	AllocationID string   `json:"allocation_id,omitempty"`
	MembershipID string   `json:"membership_id,omitempty"`
	Custos       any      `json:"custos"`
	Slurm        any      `json:"slurm"`
	Choices      []Choice `json:"choices"`
	// Target is the association an access or partition conflict is about.
	Target client.AssocFilter `json:"-"`
	// LimitKey is a limit conflict's limit, such as tres:cpu.
	LimitKey string `json:"-"`
}

// QoSValue is each side of an allocation_qos or member_qos conflict.
type QoSValue struct {
	QoS     []string `json:"qos"`
	Default string   `json:"default,omitempty"`
}

// Mark is a satisfied pending mark. It clears only while its owner still carries Since,
// so a change made during the sweep keeps its own mark.
type Mark struct {
	ID    string
	Since time.Time
}

type Push struct {
	QoS                []client.QoS
	Associations       []client.Association // conflicting fields carry SLURM's value
	DeleteAssociations []client.AssocFilter
	ClearAllocations   []Mark
	ClearMemberships   []Mark
	ClearQoS           []Mark
}

var (
	choicesCustosSlurm       = []Choice{ChoiceCustos, ChoiceSlurm}
	choicesCustosSlurmManual = []Choice{ChoiceCustos, ChoiceSlurm, ChoiceManual}
	choicesCustos            = []Choice{ChoiceCustos}
)

func Diff(want Desired, have Live) (Push, []Conflict) {
	d := &differ{want: want, have: have}
	d.diffQoSDefinitions()
	d.diffAccounts()
	sort.Slice(d.conflicts, func(i, j int) bool { return d.conflicts[i].Key < d.conflicts[j].Key })
	return d.push, d.conflicts
}

type differ struct {
	want      Desired
	have      Live
	push      Push
	conflicts []Conflict
}

// --- QoS definitions ---

func (d *differ) diffQoSDefinitions() {
	live := map[string]client.QoS{}
	for _, q := range d.have.QoS {
		live[q.Name] = q
	}
	for _, name := range slices.Sorted(maps.Keys(d.want.QoS)) {
		definition := d.want.QoS[name]
		wantQoS := toClientQoS(definition)
		got, ok := live[name]
		switch {
		case ok && qosEqual(wantQoS, got):
			if definition.SlurmPendingSince != nil {
				d.push.ClearQoS = append(d.push.ClearQoS, Mark{definition.ID, *definition.SlurmPendingSince})
			}
		case !ok || definition.SlurmPendingSince != nil:
			for _, t := range got.MaxTRESPerUser {
				if _, kept := tresMap(wantQoS.MaxTRESPerUser)[t.Key()]; !kept {
					wantQoS.MaxTRESPerUser = append(wantQoS.MaxTRESPerUser, client.TRES{Type: t.Type, Name: t.Name, Count: -1})
				}
			}
			d.push.QoS = append(d.push.QoS, wantQoS)
		default:
			d.conflicts = append(d.conflicts, Conflict{Key: "qos:" + name, Kind: KindQoSDefinition, Name: name, Label: "QoS",
				Custos: definition, Slurm: fromClientQoS(got), Choices: choicesCustosSlurmManual})
		}
	}
	usedByPeople, usedBySystem := map[string]bool{}, map[string]bool{}
	for _, a := range d.have.Associations {
		used := usedByPeople
		if d.want.SystemAccounts[a.Account] {
			used = usedBySystem
		}
		for _, q := range a.QoS {
			used[q] = true
		}
	}
	// Known gap: clusters that define one QoS name differently overwrite each other on push.
	for _, name := range slices.Sorted(maps.Keys(live)) {
		if _, ok := d.want.QoS[name]; ok || d.want.OtherClusterQoS[name] || (usedBySystem[name] && !usedByPeople[name]) {
			continue
		}
		d.conflicts = append(d.conflicts, Conflict{Key: "qos:" + name, Kind: KindQoSDefinition, Name: name, Label: "QoS",
			Slurm: fromClientQoS(live[name]), Choices: choicesCustosSlurm})
	}
}

func toClientQoS(q models.ComputeClusterQoS) client.QoS {
	tres, _ := client.ParseTRES(q.MaxTRESPerUser) // a malformed string reads as no limit
	return client.QoS{Name: q.Name, Priority: q.Priority, MaxWallMinutes: q.MaxWallMinutes,
		MaxJobsPerUser: q.MaxJobsPerUser, MaxTRESPerUser: tres, Flags: q.Flags}
}

func fromClientQoS(q client.QoS) models.ComputeClusterQoS {
	return models.ComputeClusterQoS{Name: q.Name, Priority: q.Priority, MaxWallMinutes: q.MaxWallMinutes,
		MaxJobsPerUser: q.MaxJobsPerUser, MaxTRESPerUser: client.FormatTRES(q.MaxTRESPerUser), Flags: q.Flags}
}

func qosEqual(a, b client.QoS) bool {
	return a.Priority == b.Priority && ptrEqual(a.MaxWallMinutes, b.MaxWallMinutes) &&
		ptrEqual(a.MaxJobsPerUser, b.MaxJobsPerUser) &&
		maps.Equal(tresMap(a.MaxTRESPerUser), tresMap(b.MaxTRESPerUser)) &&
		slices.Equal(sortedSet(a.Flags), sortedSet(b.Flags))
}

// --- Associations ---

func (d *differ) diffAccounts() {
	byAccount := map[string][]client.Association{}
	for _, a := range d.have.Associations {
		byAccount[a.Account] = append(byAccount[a.Account], a)
	}
	for _, account := range slices.Sorted(maps.Keys(d.want.Accounts)) {
		if !d.want.SystemAccounts[account] {
			d.diffAccount(account, d.want.Accounts[account], byAccount[account])
		}
	}
}

func (d *differ) diffAccount(account string, want DesiredAccount, live []client.Association) {
	var accountAssoc *client.Association
	liveMembers := map[string]client.Association{}
	for _, a := range live {
		switch {
		case a.Partition != "":
			d.conflicts = append(d.conflicts, Conflict{
				Key:  fmt.Sprintf("partition:%s:%s:%s", account, a.User, a.Partition),
				Kind: KindPartitionAssociation, Name: a.Partition, Label: "Partition " + a.Partition, User: a.User,
				AllocationID: want.AllocationID, MembershipID: want.Members[a.User].MembershipID,
				Slurm: a.Partition, Choices: choicesCustos,
				Target: client.AssocFilter{Cluster: a.Cluster, Account: account, User: a.User, Partition: a.Partition}})
		case a.User == "":
			accountAssoc = &a
		default:
			liveMembers[a.User] = a
		}
	}

	wantQoS := d.accountQoS(want, accountAssoc)
	cluster := ""
	if accountAssoc != nil {
		cluster = accountAssoc.Cluster
	}

	accountPending := want.PendingSince != nil
	accountTarget := target{prefix: "account:" + account, qosKind: KindAllocationQoS, limitKind: KindAllocationLimit,
		allocationID: want.AllocationID, wantQoS: wantQoS, limits: want.Limits,
		qosPending: accountPending, limitPending: accountPending,
		base: client.Association{Account: account, Cluster: cluster}}
	accountEqual := d.compare(accountTarget, accountAssoc)
	// An account conflict is one conflict: members are compared once it resolves.
	memberQoS := wantQoS
	if accountEqual.qosConflict {
		memberQoS = nil
	}

	allMembersQoSEqual := true
	for _, user := range slices.Sorted(maps.Keys(want.Members)) {
		m := want.Members[user]
		got, present := liveMembers[user]
		if !m.Active {
			d.diffRevoked(account, user, want, m, got, present)
			continue
		}
		memberPending := m.PendingSince != nil
		t := target{prefix: "member:" + account + ":" + user, qosKind: KindMemberQoS, limitKind: KindMemberLimit,
			allocationID: want.AllocationID, membershipID: m.MembershipID, wantQoS: memberQoS, limits: m.Limits,
			qosPending: accountPending || memberPending, limitPending: memberPending,
			base: client.Association{Account: account, User: user, Cluster: cluster}}
		var gotPtr *client.Association
		if present {
			gotPtr = &got
		}
		memberEqual := d.compare(t, gotPtr)
		if memberPending && memberEqual.all() {
			d.push.ClearMemberships = append(d.push.ClearMemberships, Mark{m.MembershipID, *m.PendingSince})
		}
		allMembersQoSEqual = allMembersQoSEqual && memberEqual.qos
	}
	for _, user := range slices.Sorted(maps.Keys(liveMembers)) {
		if _, known := want.Members[user]; !known {
			d.diffRevoked(account, user, want, DesiredMember{}, liveMembers[user], true)
		}
	}
	if accountPending && accountEqual.all() && allMembersQoSEqual {
		d.push.ClearAllocations = append(d.push.ClearAllocations, Mark{want.AllocationID, *want.PendingSince})
	}
}

// diffRevoked handles a live member association whose membership is not active. An allocation
// that stopped being active revokes through its own mark.
func (d *differ) diffRevoked(account, user string, a DesiredAccount, m DesiredMember, live client.Association, present bool) {
	target := client.AssocFilter{Account: account, User: user, Cluster: live.Cluster}
	switch {
	case !present:
		if m.PendingSince != nil {
			d.push.ClearMemberships = append(d.push.ClearMemberships, Mark{m.MembershipID, *m.PendingSince})
		}
	// Access of a user Custos has no membership for is never removed unasked.
	case m.PendingSince != nil || (a.PendingSince != nil && m.MembershipID != ""):
		d.push.DeleteAssociations = append(d.push.DeleteAssociations, target)
	default:
		// Use SLURM activates the membership, so it needs one under an active allocation.
		offered := choicesCustosSlurm
		if m.MembershipID == "" || !a.Active {
			offered = choicesCustos
		}
		d.conflicts = append(d.conflicts, Conflict{Key: fmt.Sprintf("member:%s:%s:access", account, user),
			Kind: KindAccess, Name: "access", Label: "Access", User: user,
			AllocationID: a.AllocationID, MembershipID: m.MembershipID,
			Custos: "inactive", Slurm: "access", Choices: offered, Target: target})
	}
}

// accountQoS is the complete list the account and its members hold, or nil when it
// cannot be known: no own list and no live parent list to inherit.
func (d *differ) accountQoS(a DesiredAccount, live *client.Association) []string {
	list := a.QoS
	if len(list) == 0 {
		if live == nil {
			return nil
		}
		parent := d.liveAccountAssociation(live.ParentAccount, live.Cluster)
		if parent == nil || len(parent.QoS) == 0 {
			return nil
		}
		list = parent.QoS
	}
	if d.want.DefaultQoS != "" {
		list = append(slices.Clone(list), d.want.DefaultQoS)
	}
	return sortedSet(list)
}

func (d *differ) liveAccountAssociation(account, cluster string) *client.Association {
	for _, a := range d.have.Associations {
		if a.Account == account && a.User == "" && a.Partition == "" && a.Cluster == cluster {
			return &a
		}
	}
	return nil
}

type target struct {
	prefix                   string
	qosKind, limitKind       Kind
	allocationID             string
	membershipID             string
	wantQoS                  []string // nil: unknown, not compared
	limits                   client.AssocLimits
	qosPending, limitPending bool
	base                     client.Association
}

type equality struct {
	qos, limits bool
	qosConflict bool
}

func (e equality) all() bool { return e.qos && e.limits }

// compare diffs one association, queues its push and returns which parts already match.
func (d *differ) compare(t target, live *client.Association) equality {
	if live == nil {
		p := t.base
		p.QoS, p.Limits = t.wantQoS, t.limits
		if t.wantQoS != nil {
			p.DefaultQoS = d.want.DefaultQoS
		}
		d.push.Associations = append(d.push.Associations, p)
		return equality{}
	}
	out := *live
	eq := equality{qos: true, limits: true}
	push := false

	if t.wantQoS != nil {
		listDiffers := !slices.Equal(sortedSet(live.QoS), t.wantQoS)
		defaultDiffers := d.want.DefaultQoS != "" && live.DefaultQoS != d.want.DefaultQoS
		// A changed cluster default marks every allocation pending itself, so it drives the default push alone.
		defaultQoS, isDefined := d.want.QoS[d.want.DefaultQoS]
		defaultPending := t.qosPending || (isDefined && defaultQoS.SlurmPendingSince != nil)
		listPush := listDiffers && (t.qosPending || len(live.QoS) == 0)
		// A default outside SLURM's kept list would be rejected, so a list conflict holds the default back too.
		listConflict := listDiffers && !listPush
		defaultPush := defaultDiffers && !listConflict && (defaultPending || live.DefaultQoS == "")
		eq.qos = !listDiffers && !defaultDiffers
		if listPush {
			out.QoS, push = t.wantQoS, true
		}
		if defaultPush {
			out.DefaultQoS, push = d.want.DefaultQoS, true
		}
		eq.qosConflict = listConflict || (defaultDiffers && !defaultPush)
		if eq.qosConflict {
			// Custos stores a list per allocation but one default per cluster, so a
			// difference in the default alone has nothing to save SLURM's value into.
			offered := choicesCustosSlurmManual
			if !listDiffers {
				offered = choicesCustos
			}
			d.conflicts = append(d.conflicts, Conflict{Key: t.prefix + ":qos", Kind: t.qosKind,
				Name: "qos", Label: "QoS", User: t.base.User, AllocationID: t.allocationID, MembershipID: t.membershipID,
				Custos: QoSValue{QoS: t.wantQoS, Default: d.want.DefaultQoS},
				Slurm:  QoSValue{QoS: sortedSet(live.QoS), Default: live.DefaultQoS}, Choices: offered})
		}
	}

	wantLimits, haveLimits := limitMap(t.limits), limitMap(live.Limits)
	merged := maps.Clone(wantLimits)
	all := maps.Clone(haveLimits)
	maps.Copy(all, wantLimits)
	for _, key := range slices.Sorted(maps.Keys(all)) {
		w, wok := wantLimits[key]
		h, hok := haveLimits[key]
		if wok == hok && w == h {
			continue
		}
		eq.limits = false
		if t.limitPending || !hok {
			push = true
			continue
		}
		merged[key] = h
		d.conflicts = append(d.conflicts, Conflict{Key: t.prefix + ":limit:" + key, Kind: t.limitKind,
			Name: key[strings.LastIndex(key, ":")+1:], Label: key, User: t.base.User,
			AllocationID: t.allocationID, MembershipID: t.membershipID,
			Custos: valueOrNil(w, wok), Slurm: h, Choices: limitChoices(wok), LimitKey: key})
	}
	if push {
		out.Limits = limitsFromMap(merged)
		d.push.Associations = append(d.push.Associations, out)
	}
	return eq
}

// A limit SLURM has and Custos lacks has nowhere to be saved in Custos.
func limitChoices(custosHasIt bool) []Choice {
	if custosHasIt {
		return choicesCustosSlurm
	}
	return choicesCustos
}

func valueOrNil(v float64, ok bool) any {
	if !ok {
		return nil
	}
	return v
}

const (
	limitGrpJobs = "grp_jobs"
	limitMaxWall = "max_wall"
	limitTRES    = "tres:"
	limitMins    = "tres_mins:"
)

func limitMap(l client.AssocLimits) map[string]float64 {
	m := map[string]float64{}
	if l.GrpJobs != nil {
		m[limitGrpJobs] = float64(*l.GrpJobs)
	}
	if l.MaxWallPM != nil {
		m[limitMaxWall] = float64(*l.MaxWallPM)
	}
	for k, v := range tresMap(l.GrpTRES) {
		m[limitTRES+k] = v
	}
	for k, v := range tresMap(l.GrpTRESMins) {
		m[limitMins+k] = v
	}
	return m
}

func limitsFromMap(m map[string]float64) client.AssocLimits {
	var l client.AssocLimits
	for _, key := range slices.Sorted(maps.Keys(m)) {
		switch {
		case key == limitGrpJobs:
			v := int64(m[key])
			l.GrpJobs = &v
		case key == limitMaxWall:
			v := int64(m[key])
			l.MaxWallPM = &v
		case strings.HasPrefix(key, limitMins):
			l.GrpTRESMins = append(l.GrpTRESMins, client.TRESFromKey(strings.TrimPrefix(key, limitMins), m[key]))
		default:
			l.GrpTRES = append(l.GrpTRES, client.TRESFromKey(key[len(limitTRES):], m[key]))
		}
	}
	return l
}

func tresMap(ts []client.TRES) map[string]float64 {
	m := map[string]float64{}
	for _, t := range ts {
		m[t.Key()] = t.Count
	}
	return m
}

func ptrEqual(a, b *int64) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func sortedSet(s []string) []string {
	out := slices.Clone(s)
	slices.Sort(out)
	return slices.Compact(out)
}
