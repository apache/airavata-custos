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
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// tresFor builds a SLURM TRES from a stored resource type and count. A GRES
// resource is stored joined as "gres/gpu" but SLURM wants it split into type
// and name; a plain resource like "cpu" has no name.
func tresFor(resourceType string, count int64) client.TRES {
	if typ, name, ok := strings.Cut(resourceType, "/"); ok {
		return client.TRES{Type: typ, Name: name, Count: float64(count)}
	}
	return client.TRES{Type: resourceType, Count: float64(count)}
}

// errNotProvisioned means the account does not exist on the cluster yet, so
// writing the association now would make slurmctld cache a failed uid lookup
// and reject the user's jobs. The reconciler retries these.
var errNotProvisioned = errors.New("cluster account not provisioned yet")

// errEmptyResourceType means the resource row has an empty resource_type. A
// retry cannot fix the row, so the delivery is not retried.
var errEmptyResourceType = fmt.Errorf("%w: resource type is empty", events.ErrPermanent)

// assocKey identifies an association the way Slurm does. Cluster is not part
// of it because callers scope their lookups to one cluster already.
type assocKey struct {
	account   string
	user      string
	partition string
}

func keyOf(a client.Association) assocKey {
	return assocKey{account: a.Account, user: a.User, partition: a.Partition}
}

// sameLimits reports whether two associations already carry the same limits.
// Only the limits are compared: the key fields match by construction.
func sameLimits(want, got client.Association) bool {
	return sameTRES(want.Limits.GrpTRES, got.Limits.GrpTRES) &&
		sameTRES(want.Limits.GrpTRESMins, got.Limits.GrpTRESMins)
}

func sameTRES(want, got []client.TRES) bool {
	if len(want) != len(got) {
		return false
	}
	counts := make(map[string]float64, len(want))
	for _, t := range want {
		counts[t.Type] = t.Count
	}
	for _, t := range got {
		if c, ok := counts[t.Type]; !ok || c != t.Count {
			return false
		}
	}
	return true
}

// noAllocation is every user's default account. It runs no jobs, so a job names the allocation it charges,
// and no allocation is anyone's default, which SLURM would refuse to remove.
const noAllocation = "none"

// custos marks Custos-created accounts for the reconciler.
const custos = "custos"

// writeAssociation upserts a user's association after making noAllocation their default.
func (a *AssociationSubscriber) writeAssociation(assoc client.Association) error {
	if assoc.User == "" {
		if err := a.slurmClient.CreateAccount(client.Account{Name: assoc.Account, Organization: custos}, assoc.Cluster); err != nil {
			return err
		}
		return a.slurmClient.UpsertAssociation(assoc)
	}
	noJobs := int64(0)
	account := client.Account{Name: noAllocation, Description: "Default account; runs no jobs", GrpJobs: &noJobs}
	if err := a.slurmClient.CreateAccount(account, assoc.Cluster); err != nil {
		return fmt.Errorf("create %s account: %w", noAllocation, err)
	}
	if err := a.slurmClient.SetDefaultAccount(assoc.Cluster, assoc.User, noAllocation); err != nil {
		return fmt.Errorf("set default account: %w", err)
	}
	return a.slurmClient.UpsertAssociation(assoc)
}

func (a *AssociationSubscriber) removeAssociation(ctx context.Context, assoc client.Association) error {
	err := a.slurmClient.DeleteAssociation(client.AssocFilter{Cluster: assoc.Cluster, Account: assoc.Account, User: assoc.User, Partition: assoc.Partition})
	if err == nil {
		slog.InfoContext(ctx, "Removed association", "cluster", assoc.Cluster, "account", assoc.Account, "user", assoc.User)
		return nil
	}
	slog.WarnContext(ctx, "Blocked association while jobs run", "cluster", assoc.Cluster, "account", assoc.Account, "user", assoc.User, "error", err)
	none := int64(0)
	assoc.Limits.GrpJobs = &none
	return a.slurmClient.UpsertAssociation(assoc)
}

// syncAssociations blocks a removal Slurm refuses while jobs run; the reconciler retries it.
func (a *AssociationSubscriber) syncAssociations(ctx context.Context, scope client.AssocFilter, wants []client.Association) error {
	existing, err := a.slurmClient.ListAssociations(scope)
	if err != nil {
		return err
	}
	have := make(map[assocKey]client.Association, len(existing))
	for _, e := range existing {
		have[keyOf(e)] = e
	}
	var errs []error
	for _, w := range wants {
		got, ok := have[keyOf(w)]
		delete(have, keyOf(w))
		if ok && got.Limits.GrpJobs == nil && sameLimits(w, got) {
			continue
		}
		if err := a.writeAssociation(w); err != nil {
			errs = append(errs, fmt.Errorf("upsert association: %w", err))
			continue
		}
		slog.InfoContext(ctx, "Upserted association", "association", w)
	}
	for _, e := range have {
		if e.User != "" && e.Account != noAllocation {
			errs = append(errs, a.removeAssociation(ctx, e))
		}
	}
	return errors.Join(errs...)
}

// Builds the member's association without writing it. It has no partition,
// so the cluster's own partition settings decide where the member can run.
func (a *AssociationSubscriber) userAssociation(ctx context.Context, cluster string, allocation models.ComputeAllocation, user, membershipID string) (*client.Association, error) {
	overrides, err := a.coreService.ListOverridesForMembership(ctx, membershipID)
	if err != nil {
		return nil, fmt.Errorf("list overrides for membership: %w", err)
	}
	limits, err := a.limitsFor(ctx, overrides)
	if err != nil {
		return nil, err
	}
	return &client.Association{
		Account: slurmAccount(allocation.Name),
		Cluster: cluster,
		User:    user,
		QoS:     []string{"normal"},
		Limits:  limits,
	}, nil
}

// Builds the member's limits from their overrides. The association covers
// every partition, so overrides on two GPU resources become one GPU limit,
// their sum. CPU and GPU stay separate.
func (a *AssociationSubscriber) limitsFor(ctx context.Context, overrides []models.ComputeAllocationMembershipResourceOverride) (client.AssocLimits, error) {
	var limits client.AssocLimits
	for _, o := range overrides {
		resource, err := a.coreService.GetComputeAllocationResource(ctx, o.ComputeAllocationResourceID)
		if err != nil {
			return limits, fmt.Errorf("get compute allocation resource: %w", err)
		}
		if resource.ResourceType == "" {
			return limits, errEmptyResourceType
		}
		if o.OverrideResourceAmount > 0 {
			limits.GrpTRES = addTRES(limits.GrpTRES, tresFor(resource.ResourceType, o.OverrideResourceAmount))
		}
		if o.OverrideResourceTime > 0 {
			limits.GrpTRESMins = addTRES(limits.GrpTRESMins, tresFor(resource.ResourceType, o.OverrideResourceTime))
		}
	}
	return limits, nil
}

func addTRES(list []client.TRES, t client.TRES) []client.TRES {
	for i := range list {
		if list[i].Type == t.Type && list[i].Name == t.Name {
			list[i].Count += t.Count
			return list
		}
	}
	return append(list, t)
}
