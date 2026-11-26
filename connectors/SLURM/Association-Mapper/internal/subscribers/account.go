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

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// createAccount creates the allocation's Slurm account, owned by the PI's organization.
func (a *AssociationSubscriber) createAccount(ctx context.Context, cluster models.ComputeCluster, allocationID string) error {
	allocation, err := a.coreService.GetComputeAllocation(ctx, allocationID)
	if err != nil {
		return fmt.Errorf("get compute allocation: %w", err)
	}
	project, err := a.coreService.GetProject(ctx, allocation.ProjectID)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}
	pi, err := a.coreService.GetUser(ctx, project.ProjectPIID)
	if err != nil {
		return fmt.Errorf("get project PI: %w", err)
	}
	organization, err := a.coreService.GetOrganization(ctx, pi.OrganizationID)
	if err != nil {
		return fmt.Errorf("get organization: %w", err)
	}
	return a.slurmClient.CreateAccount(client.Account{
		Name:         SlurmAccount(allocation.Name),
		Description:  allocation.Name,
		Organization: organization.Name,
	}, cluster.Name)
}

// syncAllocation syncs the allocation's cluster. A deleted allocation's SLURM account stays for job history;
// the delete guards have already revoked access.
func (a *AssociationSubscriber) syncAllocation(ctx context.Context, allocation models.ComputeAllocation) error {
	return a.syncClusterByID(ctx, allocation.ComputeClusterID)
}

func (a *AssociationSubscriber) syncMapping(ctx context.Context, mapping models.ComputeAllocationResourceMapping) error {
	resource, err := a.coreService.GetComputeAllocationResource(ctx, mapping.ComputeAllocationResourceID)
	if err != nil {
		return fmt.Errorf("get compute allocation resource: %w", err)
	}
	if resource.ResourceType == "" && (mapping.ResourceAmount > 0 || mapping.ResourceTime > 0) {
		return errEmptyResourceType
	}
	return a.syncAllocationCluster(ctx, mapping.ComputeAllocationID)
}
