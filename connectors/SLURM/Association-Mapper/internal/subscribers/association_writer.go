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
	"fmt"
	"log/slog"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
)

// errEmptyResourceType means the resource row has an empty resource_type. A
// retry cannot fix the row, so the delivery is not retried.
var errEmptyResourceType = fmt.Errorf("%w: resource type is empty", events.ErrPermanent)

// noAllocation is every user's default account. It runs no jobs, so a job names the allocation it charges,
// and no allocation is anyone's default, which SLURM would refuse to remove.
const noAllocation = "none"

// provisionNoAllocationDefault makes noAllocation the user's default account, creating the account and the user's association with it.
func (a *AssociationSubscriber) provisionNoAllocationDefault(cluster, user string) error {
	noJobs := int64(0)
	account := client.Account{Name: noAllocation, Description: "Default account; runs no jobs", GrpJobs: &noJobs}
	if err := a.slurmClient.CreateAccount(account, cluster); err != nil {
		return fmt.Errorf("create %s account: %w", noAllocation, err)
	}
	if err := a.slurmClient.SetDefaultAccount(cluster, user, noAllocation); err != nil {
		return fmt.Errorf("set default account: %w", err)
	}
	return nil
}

// resourceLimit is one resource's amount and time, from a mapping or an override.
type resourceLimit struct {
	resourceID   string
	amount, time int64
}

// limitsOf sums resource limits per TRES type, since one association covers every partition.
// A resource with no type is skipped.
func (a *AssociationSubscriber) limitsOf(ctx context.Context, rows []resourceLimit) (client.AssocLimits, error) {
	var limits client.AssocLimits
	for _, r := range rows {
		resource, err := a.coreService.GetComputeAllocationResource(ctx, r.resourceID)
		if err != nil {
			return limits, fmt.Errorf("get compute allocation resource: %w", err)
		}
		if resource.ResourceType == "" {
			slog.WarnContext(ctx, "Skipping a resource limit with an empty resource type", "resource_id", r.resourceID)
			continue
		}
		if r.amount > 0 {
			limits.GrpTRES = addTRES(limits.GrpTRES, client.TRESFromKey(resource.ResourceType, float64(r.amount)))
		}
		if r.time > 0 {
			limits.GrpTRESMins = addTRES(limits.GrpTRESMins, client.TRESFromKey(resource.ResourceType, float64(r.time)))
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
