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

// Live SLURM routes: the cluster view of partitions, system accounts and conflicts, and conflict resolve.
// Every read is live; nothing is cached.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/drift"
	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/subscribers"
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// ConflictResolvedEvent is the audit event type written for every resolve.
const ConflictResolvedEvent = "SlurmConflictResolved"

// Snapshotter diffs one cluster's Custos state against SLURM, writing neither.
type Snapshotter interface {
	Snapshot(ctx context.Context, cluster models.ComputeCluster) (subscribers.Snapshot, error)
}

// Slurm is the part of slurmrestd the routes call directly.
type Slurm interface {
	ListAssociations(f client.AssocFilter) ([]client.Association, error)
	DeleteQoS(name string) error
	DeleteAssociation(f client.AssocFilter) error
}

type Handlers struct {
	core      service.CoreService
	snapshots Snapshotter
	slurm     Slurm
}

func NewHandlers(core service.CoreService, snapshots Snapshotter, slurm Slurm) *Handlers {
	return &Handlers{core: core, snapshots: snapshots, slurm: slurm}
}

// RegisterRoutes attaches the SLURM connector routes to router.
// Resolve checks the write privilege the conflict's kind needs.
func (h *Handlers) RegisterRoutes(router *identity.Router) {
	router.RequirePrivilege("GET /connectors/slurm/clusters/{id}", models.ClustersRead, h.cluster)
	router.RequireAuth("POST /connectors/slurm/clusters/{id}/conflicts/resolve", h.resolve)
}

// ClusterView is the cluster's live SLURM state, read in one snapshot.
type ClusterView struct {
	Partitions     []client.Partition `json:"partitions"`
	SystemAccounts []SystemAccount    `json:"system_accounts"`
	Conflicts      []drift.Conflict   `json:"conflicts"`
}

// SystemAccount is an account no allocation owns, with the users it holds.
type SystemAccount struct {
	Name   string       `json:"name"`
	Parent string       `json:"parent"`
	QoS    []string     `json:"qos"`
	Users  []SystemUser `json:"users"`
}

type SystemUser struct {
	Name           string   `json:"name"`
	DefaultAccount string   `json:"default_account"`
	QoS            []string `json:"qos"`
}

// ResolveRequest is the body of the resolve route. SeenSlurm is Slurm's value as the caller saw it.
type ResolveRequest struct {
	Key       string       `json:"key"`
	Choice    drift.Choice `json:"choice"`
	SeenSlurm any          `json:"seen_slurm"`
	Manual    any          `json:"manual"`
}

// @Summary	Read the cluster's live SLURM state
// @Description	The cluster's partitions, system accounts and conflicts, read live in one snapshot.
// @Tags	SLURM
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Success	200	{object}	ClusterView
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Failure	403	{object}	object{error=string}	"Caller lacks core:clusters:read"
// @Failure	404	{object}	object{error=string}	"Compute cluster not found"
// @Failure	503	{object}	object{error=string}	"slurmrestd cannot be read"
// @Router	/connectors/slurm/clusters/{id} [get]
func (h *Handlers) cluster(w http.ResponseWriter, r *http.Request) {
	_, snap, ok := h.snapshot(w, r)
	if !ok {
		return
	}
	common.WriteJSON(w, http.StatusOK, ClusterView{Partitions: snap.Have.Partitions, SystemAccounts: systemAccounts(snap), Conflicts: snap.Conflicts})
}

// snapshot loads the path's cluster and its fresh diff, writing the failure response itself.
func (h *Handlers) snapshot(w http.ResponseWriter, r *http.Request) (models.ComputeCluster, subscribers.Snapshot, bool) {
	cluster, err := h.core.GetComputeCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return models.ComputeCluster{}, subscribers.Snapshot{}, false
	}
	snap, err := h.snapshots.Snapshot(r.Context(), *cluster)
	if err != nil {
		common.WriteServiceError(w, err)
		return *cluster, snap, false
	}
	return *cluster, snap, true
}

// systemAccounts lists the live accounts no allocation owns, each with its users.
func systemAccounts(snap subscribers.Snapshot) []SystemAccount {
	byName := map[string]*SystemAccount{}
	account := func(name string) *SystemAccount {
		if byName[name] == nil {
			byName[name] = &SystemAccount{Name: name, QoS: []string{}, Users: []SystemUser{}}
		}
		return byName[name]
	}
	defaults := map[string]string{} // user -> default account
	for _, a := range snap.Have.Associations {
		if a.User != "" && a.IsDefault != nil && *a.IsDefault {
			defaults[a.User] = a.Account
		}
	}
	for _, a := range snap.Have.Associations {
		if !snap.Want.SystemAccounts[a.Account] || a.Partition != "" {
			continue
		}
		acc := account(a.Account)
		if a.User == "" {
			acc.Parent, acc.QoS = a.ParentAccount, nonNil(a.QoS)
			continue
		}
		acc.Users = append(acc.Users, SystemUser{Name: a.User, DefaultAccount: defaults[a.User], QoS: nonNil(a.QoS)})
	}
	out := make([]SystemAccount, 0, len(byName))
	for _, acc := range byName {
		sort.Slice(acc.Users, func(i, j int) bool { return acc.Users[i].Name < acc.Users[j].Name })
		out = append(out, *acc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// @Summary	Resolve a SLURM conflict
// @Description	Applies one decision to one conflict of a fresh diff: custos marks the owner pending, slurm saves SLURM's value without a mark, manual saves the given value and marks pending.
// @Description	A QoS definition conflict needs core:clusters:write; every other kind needs core:allocations:write.
// @Tags	SLURM
// @Security	BearerAuth
// @Accept	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Param	body	body	ResolveRequest	true	"The decision"
// @Success	204	"Resolved"
// @Failure	400	{object}	object{error=string}	"Malformed body or choice not offered"
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Failure	403	{object}	object{error=string}	"Caller lacks the write privilege for the conflict's kind"
// @Failure	404	{object}	object{error=string}	"Compute cluster not found"
// @Failure	409	{object}	object{error=string}	"Conflict is stale, or the change would break SLURM"
// @Failure	502	{object}	object{error=string}	"SLURM refused the delete"
// @Failure	503	{object}	object{error=string}	"slurmrestd cannot be read"
// @Router	/connectors/slurm/clusters/{id}/conflicts/resolve [post]
func (h *Handlers) resolve(w http.ResponseWriter, r *http.Request) {
	var req ResolveRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if req.SeenSlurm == nil {
		common.WriteError(w, http.StatusBadRequest, fmt.Errorf("seen_slurm is required"))
		return
	}
	// QoS definitions belong to the cluster; every other conflict belongs to an allocation.
	needed := models.AllocationsWrite
	if strings.HasPrefix(req.Key, "qos:") {
		needed = models.ClustersWrite
	}
	if !identity.HasPrivilege(r.Context(), needed) {
		common.WriteError(w, http.StatusForbidden, fmt.Errorf("caller lacks %s", needed))
		return
	}

	cluster, snap, ok := h.snapshot(w, r)
	if !ok {
		return
	}
	idx := slices.IndexFunc(snap.Conflicts, func(c drift.Conflict) bool { return c.Key == req.Key })
	if idx < 0 {
		common.WriteError(w, http.StatusConflict, fmt.Errorf("conflict %q is no longer present", req.Key))
		return
	}
	conflict := snap.Conflicts[idx]
	if !sameJSON(req.SeenSlurm, conflict.Slurm) {
		common.WriteError(w, http.StatusConflict, fmt.Errorf("conflict %q changed on SLURM since it was read", req.Key))
		return
	}
	if !slices.Contains(conflict.Choices, req.Choice) {
		common.WriteError(w, http.StatusBadRequest, fmt.Errorf("choice %q is not offered for %q", req.Choice, req.Key))
		return
	}

	ctx := audit.WithSource(r.Context(), "slurm")
	if err := h.apply(ctx, cluster, snap, conflict, req); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	h.recordResolved(ctx, cluster, req)
	w.WriteHeader(http.StatusNoContent)
}

// sameJSON reports whether the caller's seen value is the live one, compared as decoded JSON.
func sameJSON(seen, live any) bool {
	encoded, err := json.Marshal(live)
	if err != nil {
		return false
	}
	var want any
	if json.Unmarshal(encoded, &want) != nil {
		return false
	}
	return reflect.DeepEqual(want, seen)
}

func (h *Handlers) recordResolved(ctx context.Context, cluster models.ComputeCluster, req ResolveRequest) {
	details, _ := json.Marshal(map[string]string{"key": req.Key, "choice": string(req.Choice)})
	_, err := h.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  ConflictResolvedEvent,
		EntityType: "compute_cluster",
		EntityID:   cluster.ID,
		Details:    string(details),
		EventTime:  time.Now(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to record SLURM conflict resolution", "key", req.Key, "error", err)
	}
}

func (h *Handlers) apply(ctx context.Context, cluster models.ComputeCluster, snap subscribers.Snapshot, c drift.Conflict, req ResolveRequest) error {
	switch c.Kind {
	case drift.KindQoSDefinition:
		return h.resolveQoSDefinition(ctx, cluster, c, req)
	case drift.KindAllocationQoS, drift.KindMemberQoS:
		return h.resolveQoSList(ctx, cluster, snap, c, req)
	case drift.KindAllocationLimit, drift.KindMemberLimit:
		return h.resolveLimit(ctx, c, req)
	case drift.KindAccess:
		return h.resolveAccess(ctx, c, req)
	default: // drift.KindPartitionAssociation
		return h.deleteAssociation(c.Target)
	}
}

// --- QoS definitions ---

func (h *Handlers) resolveQoSDefinition(ctx context.Context, cluster models.ComputeCluster, c drift.Conflict, req ResolveRequest) error {
	definitions, err := h.core.ListComputeClusterQoS(ctx, cluster.ID, "")
	if err != nil {
		return err
	}
	saved := models.ComputeClusterQoS{ComputeClusterID: cluster.ID, Name: c.Name}
	i := slices.IndexFunc(definitions, func(q models.ComputeClusterQoS) bool { return q.Name == c.Name })
	if i >= 0 {
		saved = definitions[i]
	}

	var limits models.ComputeClusterQoS
	switch req.Choice {
	case drift.ChoiceCustos:
		if i >= 0 {
			return h.core.SetSlurmPending(ctx, service.SlurmPendingQoS, saved.ID, nil)
		}
		// Custos has no definition, so SLURM's is deleted directly; no row exists to mark.
		if err := h.slurmOnlyQoSInUse(ctx, c.Name); err != nil {
			return err
		}
		if err := h.slurm.DeleteQoS(c.Name); err != nil {
			return fmt.Errorf("%w: delete qos %q on SLURM: %v", service.ErrUpstream, c.Name, err)
		}
		return nil

	case drift.ChoiceSlurm:
		limits = c.Slurm.(models.ComputeClusterQoS)
		// SLURM already holds this value, so nothing is left to push.
		ctx = service.WithoutSlurmMark(ctx)

	default: // manual
		if i < 0 {
			return fmt.Errorf("%w: qos %q has no Custos definition to edit", service.ErrInvalidInput, c.Name)
		}
		if err := decodeManual(req.Manual, &limits); err != nil {
			return err
		}
		if limits.ID != "" || limits.ComputeClusterID != "" || limits.Name != "" || limits.IsDefault {
			return fmt.Errorf("%w: manual takes only the qos limits", service.ErrInvalidInput)
		}
	}
	saved.Priority, saved.MaxWallMinutes, saved.MaxJobsPerUser = limits.Priority, limits.MaxWallMinutes, limits.MaxJobsPerUser
	saved.MaxTRESPerUser, saved.Flags = limits.MaxTRESPerUser, limits.Flags
	if i < 0 {
		_, err = h.core.CreateComputeClusterQoS(ctx, &saved)
	} else {
		_, err = h.core.UpdateComputeClusterQoS(ctx, &saved)
	}
	return err
}

// slurmOnlyQoSInUse refuses to delete a QoS any association defaults to or any system account holds.
// slurmdbd keeps one QoS list for all clusters, so every cluster counts.
func (h *Handlers) slurmOnlyQoSInUse(ctx context.Context, name string) error {
	clusters, err := h.core.ListComputeClusters(ctx)
	if err != nil {
		return err
	}
	owned := map[string]bool{} // by cluster/account
	for _, cluster := range clusters {
		allocations, err := h.core.ListComputeAllocationsByCluster(ctx, cluster.ID)
		if err != nil {
			return err
		}
		for _, a := range allocations {
			owned[cluster.Name+"/"+subscribers.SlurmAccount(a.Name)] = true
		}
	}
	associations, err := h.slurm.ListAssociations(client.AssocFilter{})
	if err != nil {
		slog.ErrorContext(ctx, "read SLURM associations", "error", err)
		return fmt.Errorf("read SLURM associations: %w", service.ErrUnavailable)
	}
	for _, a := range associations {
		switch {
		case a.DefaultQoS == name:
			return fmt.Errorf("%w: qos %q is the default of account %q", service.ErrInUse, name, a.Account)
		case !owned[a.Cluster+"/"+a.Account] && slices.Contains(a.QoS, name):
			return fmt.Errorf("%w: qos %q is used by system account %q", service.ErrInUse, name, a.Account)
		}
	}
	return nil
}

func decodeManual(manual, dst any) error {
	if manual == nil {
		return fmt.Errorf("%w: manual is required for choice manual", service.ErrInvalidInput)
	}
	raw, err := json.Marshal(manual)
	if err != nil {
		return fmt.Errorf("%w: manual: %v", service.ErrInvalidInput, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: manual: %v", service.ErrInvalidInput, err)
	}
	return nil
}

// --- Allocation and member QoS lists ---

// markOwner marks the conflict's membership pending, or its allocation when it has none.
func (h *Handlers) markOwner(ctx context.Context, c drift.Conflict) error {
	if c.MembershipID != "" {
		return h.core.SetSlurmPending(ctx, service.SlurmPendingMembership, c.MembershipID, nil)
	}
	return h.core.SetSlurmPending(ctx, service.SlurmPendingAllocation, c.AllocationID, nil)
}

func (h *Handlers) resolveQoSList(ctx context.Context, cluster models.ComputeCluster, snap subscribers.Snapshot, c drift.Conflict, req ResolveRequest) error {
	// The cluster default is added to every list on push, so it is not stored per allocation.
	withoutDefault := func(names []string) []string {
		return slices.DeleteFunc(slices.Clone(names), func(n string) bool { return n == snap.Want.DefaultQoS })
	}

	switch req.Choice {
	case drift.ChoiceCustos:
		return h.markOwner(ctx, c)

	case drift.ChoiceSlurm:
		live := c.Slurm.(drift.QoSValue)
		names := withoutDefault(live.QoS)
		if c.Kind == drift.KindAllocationQoS {
			// Pending, so the members take the adopted list too.
			return h.core.SetComputeAllocationQoS(ctx, c.AllocationID, names)
		}
		// A member's list follows its account's, so SLURM's extra QoS join the allocation's. The wider
		// list is new to the account and the other members, so the allocation stays pending to push it.
		granted, err := subscribers.GrantedQoSNames(ctx, h.core, cluster.ID, c.AllocationID)
		if err != nil {
			return err
		}
		extra := slices.DeleteFunc(slices.Clone(names), func(n string) bool { return slices.Contains(granted, n) })
		if len(extra) == 0 {
			return fmt.Errorf("%w: SLURM's list for %q adds nothing to the allocation's; use Custos or Do manually", service.ErrInUse, c.Key)
		}
		return h.core.SetComputeAllocationQoS(ctx, c.AllocationID, append(granted, extra...))

	default: // manual
		var names []string
		if err := decodeManual(req.Manual, &names); err != nil {
			return err
		}
		return h.core.SetComputeAllocationQoS(ctx, c.AllocationID, withoutDefault(names))
	}
}

// --- Limits ---

// resolveLimit handles a conflicting limit. Only a TRES limit has a Custos value to save into:
// the mapping's amount (tres) or time (tres_mins), or the member's override of the same resource.
func (h *Handlers) resolveLimit(ctx context.Context, c drift.Conflict, req ResolveRequest) error {
	if req.Choice == drift.ChoiceCustos {
		return h.markOwner(ctx, c)
	}
	field, resourceType, _ := strings.Cut(c.LimitKey, ":")
	if field != "tres" && field != "tres_mins" {
		return fmt.Errorf("%w: limit %q has no Custos value to save into", service.ErrInvalidInput, c.LimitKey)
	}
	isTime, value := field == "tres_mins", int64(c.Slurm.(float64))
	// SLURM already holds this value, so nothing is left to push.
	ctx = service.WithoutSlurmMark(ctx)

	if c.Kind == drift.KindMemberLimit {
		overrides, err := h.core.ListOverridesForMembership(ctx, c.MembershipID)
		if err != nil {
			return err
		}
		o, err := onlyGrant(ctx, h.core, overrides, func(o models.ComputeAllocationMembershipResourceOverride) string {
			return o.ComputeAllocationResourceID
		}, resourceType)
		if err != nil {
			return err
		}
		if isTime {
			o.OverrideResourceTime = value
		} else {
			o.OverrideResourceAmount = value
		}
		_, err = h.core.UpdateComputeAllocationMembershipResourceOverride(ctx, &o)
		return err
	}

	mappings, err := h.core.ListResourceMappingsForAllocation(ctx, c.AllocationID)
	if err != nil {
		return err
	}
	m, err := onlyGrant(ctx, h.core, mappings, func(m models.ComputeAllocationResourceMapping) string { return m.ComputeAllocationResourceID }, resourceType)
	if err != nil {
		return err
	}
	if isTime {
		m.ResourceTime = value
	} else {
		m.ResourceAmount = value
	}
	_, err = h.core.UpdateAllocationResourceMapping(ctx, c.AllocationID, m.ComputeAllocationResourceID, m.ResourceAmount, m.ResourceTime)
	return err
}

// onlyGrant returns the one row whose resource has the given type; SLURM's single limit cannot be split across several.
func onlyGrant[T any](ctx context.Context, core service.CoreService, rows []T, resourceID func(T) string, resourceType string) (T, error) {
	var matches []T
	for _, row := range rows {
		resource, err := core.GetComputeAllocationResource(ctx, resourceID(row))
		if err != nil {
			return *new(T), err
		}
		if resource.ResourceType == resourceType {
			matches = append(matches, row)
		}
	}
	if len(matches) != 1 {
		return *new(T), fmt.Errorf("%w: %d Custos values grant %q; Custos cannot split SLURM's limit between them", service.ErrInUse, len(matches), resourceType)
	}
	return matches[0], nil
}

// --- Access and partitions ---

func (h *Handlers) resolveAccess(ctx context.Context, c drift.Conflict, req ResolveRequest) error {
	if req.Choice == drift.ChoiceSlurm {
		_, err := h.core.UpdateMembershipStatus(service.WithoutSlurmMark(ctx), c.MembershipID, models.ACTIVE)
		return err
	}
	if c.MembershipID != "" {
		return h.markOwner(ctx, c)
	}
	// SLURM holds access for a user with no membership, so there is no row to mark.
	return h.deleteAssociation(c.Target)
}

func (h *Handlers) deleteAssociation(f client.AssocFilter) error {
	if err := h.slurm.DeleteAssociation(f); err != nil {
		return fmt.Errorf("%w: delete association on SLURM: %v", service.ErrUpstream, err)
	}
	return nil
}
