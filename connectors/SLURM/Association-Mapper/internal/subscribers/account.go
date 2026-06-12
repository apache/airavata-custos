package subscribers

import (
	"log/slog"

	"context"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationCreation(computeAllocation models.ComputeAllocation) {
	slog.Info("Received compute allocation creation event", "account", computeAllocation)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cluster, err := a.coreService.GetComputeCluster(ctx, computeAllocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster for allocation creation", "error", err)
		a.recordAuditEvent("ComputeAllocationCreationFailed", computeAllocation.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}

	project, err := a.coreService.GetProject(ctx, computeAllocation.ProjectID)
	if err != nil {
		slog.Error("Failed to get project for allocation creation", "error", err)
		a.recordAuditEvent("ComputeAllocationCreationFailed", computeAllocation.ID, "Failed to get project. Error: "+err.Error())
		return
	}

	pi, err := a.coreService.GetUser(ctx, project.ProjectPIID)
	if err != nil {
		slog.Error("Failed to get project PI for allocation creation", "error", err)
		a.recordAuditEvent("ComputeAllocationCreationFailed", computeAllocation.ID, "Failed to get project PI. Error: "+err.Error())
		return
	}

	organization, err := a.coreService.GetOrganization(ctx, pi.OrganizationID)
	if err != nil {
		slog.Error("Failed to get organization for allocation creation", "error", err)
		a.recordAuditEvent("ComputeAllocationCreationFailed", computeAllocation.ID, "Failed to get organization. Error: "+err.Error())
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
		a.recordAuditEvent("ComputeAllocationCreationFailed", computeAllocation.ID, "Failed to create SLURM account. Error: "+err.Error())
		return
	}

	a.recordAuditEvent("ComputeAllocationCreationSucceeded", computeAllocation.ID, "Successfully created SLURM account for compute allocation.")
	slog.Info("Successfully created SLURM account for compute allocation", "account", slurmAccount)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationDeletion(computeAllocation models.ComputeAllocation) {
	slog.Info("Received compute allocation deletion event", "account", computeAllocation)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationUpdate(computeAllocation models.ComputeAllocation) {
	slog.Info("Received compute allocation update event", "account", computeAllocation)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationResourceMappingCreation(mapping models.ComputeAllocationResourceMapping) {
	slog.Info("Received compute allocation resource mapping creation event", "mapping", mapping)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allocation, err := a.coreService.GetComputeAllocation(ctx, mapping.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation for resource mapping creation", "error", err)
		a.recordAuditEvent("ComputeAllocationResourceMappingCreationFailed", mapping.ID, "Failed to get compute allocation. Error: "+err.Error())
		return
	}

	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster for resource mapping creation", "error", err)
		a.recordAuditEvent("ComputeAllocationResourceMappingCreationFailed", mapping.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}

	resource, err := a.coreService.GetComputeAllocationResource(ctx, mapping.ComputeAllocationResourceID)
	if err != nil {
		slog.Error("Failed to get compute allocation resource for resource mapping creation", "error", err)
		a.recordAuditEvent("ComputeAllocationResourceMappingCreationFailed", mapping.ID, "Failed to get compute allocation resource. Error: "+err.Error())
		return
	}

	grpTres := []client.TRES{}

	if mapping.ResourceAmount > 0 {

		if resource.ResourceType == "" {
			slog.Error("Resource type is empty for resource mapping creation", "mapping", mapping)
			a.recordAuditEvent("ComputeAllocationResourceMappingCreationFailed", mapping.ID, "Resource type is empty for resource mapping creation")
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
			a.recordAuditEvent("ComputeAllocationResourceMappingCreationFailed", mapping.ID, "Resource type is empty for resource mapping creation")
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
		a.recordAuditEvent("ComputeAllocationResourceMappingCreationFailed", mapping.ID, "Failed to upsert association. Error: "+err.Error())
		return
	} else {
		slog.Info("Successfully upserted association for membership resource override creation", "association", association)
		a.recordAuditEvent("ComputeAllocationResourceMappingCreationSucceeded", mapping.ID, "Successfully upserted association for compute allocation resource mapping creation")
	}
}
