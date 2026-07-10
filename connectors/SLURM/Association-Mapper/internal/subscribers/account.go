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
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationCreation(ctx context.Context, computeAllocation models.ComputeAllocation) {
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm.compute_allocation_create")
	defer span.End()
	span.SetAttributes(
		attribute.String("slurm.allocation_id", computeAllocation.ID),
		attribute.String("slurm.cluster_id", computeAllocation.ComputeClusterID),
	)

	slog.Info("Received compute allocation creation event", "account", computeAllocation)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cluster, err := a.coreService.GetComputeCluster(ctx, computeAllocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster for allocation creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationCreationFailed", "compute_allocation", computeAllocation.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}

	project, err := a.coreService.GetProject(ctx, computeAllocation.ProjectID)
	if err != nil {
		slog.Error("Failed to get project for allocation creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationCreationFailed", "compute_allocation", computeAllocation.ID, "Failed to get project. Error: "+err.Error())
		return
	}

	pi, err := a.coreService.GetUser(ctx, project.ProjectPIID)
	if err != nil {
		slog.Error("Failed to get project PI for allocation creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationCreationFailed", "compute_allocation", computeAllocation.ID, "Failed to get project PI. Error: "+err.Error())
		return
	}

	organization, err := a.coreService.GetOrganization(ctx, pi.OrganizationID)
	if err != nil {
		slog.Error("Failed to get organization for allocation creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationCreationFailed", "compute_allocation", computeAllocation.ID, "Failed to get organization. Error: "+err.Error())
		return
	}

	slurmAccount := client.Account{
		Name:         computeAllocation.Name,
		Description:  computeAllocation.Name,
		Organization: organization.Name,
	}

	err = a.slurmClient.CreateAccount(slurmAccount, cluster.Name) // TODO: where to get cluster name from?
	if err != nil {
		slog.Error("Failed to create SLURM account", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationCreationFailed", "compute_allocation", computeAllocation.ID, "Failed to create SLURM account. Error: "+err.Error())
		return
	}

	a.recordAuditEvent(ctx, "ComputeAllocationCreationSucceeded", "compute_allocation", computeAllocation.ID, "Successfully created SLURM account for compute allocation.")
	slog.Info("Successfully created SLURM account for compute allocation", "account", slurmAccount)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationDeletion(ctx context.Context, computeAllocation models.ComputeAllocation) {
	slog.Info("Received compute allocation deletion event", "account", computeAllocation)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationUpdate(ctx context.Context, computeAllocation models.ComputeAllocation) {
	slog.Info("Received compute allocation update event", "account", computeAllocation)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationResourceMappingCreation(ctx context.Context, mapping models.ComputeAllocationResourceMapping) {
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm.compute_allocation_resource_mapping_create")
	defer span.End()
	span.SetAttributes(attribute.String("slurm.allocation_id", mapping.ComputeAllocationID))

	slog.Info("Received compute allocation resource mapping creation event", "mapping", mapping)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	allocation, err := a.coreService.GetComputeAllocation(ctx, mapping.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation for resource mapping creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationFailed", "compute_allocation_resource_mapping", mapping.ID, "Failed to get compute allocation. Error: "+err.Error())
		return
	}

	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster for resource mapping creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationFailed", "compute_allocation_resource_mapping", mapping.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}
	span.SetAttributes(attribute.String("slurm.cluster_id", cluster.ID))

	resource, err := a.coreService.GetComputeAllocationResource(ctx, mapping.ComputeAllocationResourceID)
	if err != nil {
		slog.Error("Failed to get compute allocation resource for resource mapping creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationFailed", "compute_allocation_resource_mapping", mapping.ID, "Failed to get compute allocation resource. Error: "+err.Error())
		return
	}

	grpTres := []client.TRES{}

	if mapping.ResourceAmount > 0 {

		if resource.ResourceType == "" {
			slog.Error("Resource type is empty for resource mapping creation", "mapping", mapping)
			span.SetStatus(codes.Error, "resource type is empty")
			a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationFailed", "compute_allocation_resource_mapping", mapping.ID, "Resource type is empty for resource mapping creation")
			return
		}

		grpTres = append(grpTres, client.TRES{
			Type:  resource.ResourceType,
			Count: mapping.ResourceAmount,
		})
	}

	grpTresMins := []client.TRES{}
	if mapping.ResourceTime > 0 {

		if resource.ResourceType == "" {
			slog.Error("Resource type is empty for resource mapping creation", "mapping", mapping)
			span.SetStatus(codes.Error, "resource type is empty")
			a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationFailed", "compute_allocation_resource_mapping", mapping.ID, "Resource type is empty for resource mapping creation")
			return
		}

		grpTresMins = append(grpTresMins, client.TRES{
			Type:  resource.ResourceType,
			Count: mapping.ResourceTime,
		})
	}

	limits := client.AssocLimits{
		GrpTRES:     grpTres,
		GrpTRESMins: grpTresMins,
	}

	association := client.Association{
		Account: allocation.Name,
		Cluster: cluster.Name,
		Limits:  limits,
	}

	err = a.slurmClient.UpsertAssociation(association)
	if err != nil {
		slog.Error("Failed to upsert association for membership resource override creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationFailed", "compute_allocation_resource_mapping", mapping.ID, "Failed to upsert association. Error: "+err.Error())
		return
	} else {
		slog.Info("Successfully upserted association for membership resource override creation", "association", association)
		a.recordAuditEvent(ctx, "ComputeAllocationResourceMappingCreationSucceeded", "compute_allocation_resource_mapping", mapping.ID, "Successfully upserted association for compute allocation resource mapping creation")
	}
}
