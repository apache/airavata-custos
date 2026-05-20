package subscribers

import (
	"log/slog"

	"context"
	"time"

	client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
	"github.com/apache/airavata-custos/pkg/models"
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationCreation(computeAccount models.ComputeAllocation) {
	slog.Info("Received compute allocation creation event", "account", computeAccount)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cluster, err := a.coreService.GetComputeCluster(ctx, computeAccount.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster for allocation creation", "error", err)
		return
	}

	project, err := a.coreService.GetProject(ctx, computeAccount.ProjectID)
	if err != nil {
		slog.Error("Failed to get project for allocation creation", "error", err)
		return
	}

	pi, err := a.coreService.GetUser(ctx, project.ProjectPIID)
	if err != nil {
		slog.Error("Failed to get project PI for allocation creation", "error", err)
		return
	}

	organization, err := a.coreService.GetOrganization(ctx, pi.OrganizationID)
	if err != nil {
		slog.Error("Failed to get organization for allocation creation", "error", err)
		return
	}

	slurmAccount := client.Account{
		Name:         computeAccount.Name,
		Description:  computeAccount.Name,
		Organization: organization.Name,
	}

	err = a.slurmClient.CreateAccount(slurmAccount, cluster.Name) // TODO: where to get cluster name from?
	if err != nil {
		slog.Error("Failed to create SLURM account", "error", err)
	}
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationDeletion(computeAccount models.ComputeAllocation) {
	slog.Info("Received compute allocation deletion event", "account", computeAccount)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationUpdate(computeAccount models.ComputeAllocation) {
	slog.Info("Received compute allocation update event", "account", computeAccount)
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationResourceMappingCreation(mapping models.ComputeAllocationResourceMapping) {
	slog.Info("Received compute allocation resource mapping creation event", "mapping", mapping)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allocation, err := a.coreService.GetComputeAllocation(ctx, mapping.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation for resource mapping creation", "error", err)
		return
	}

	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster for resource mapping creation", "error", err)
		return
	}

	resource, err := a.coreService.GetComputeAllocationResource(ctx, mapping.ComputeAllocationResourceID)
	if err != nil {
		slog.Error("Failed to get compute allocation resource for resource mapping creation", "error", err)
		return
	}

	grpTres := []client.TRES{}

	if mapping.ResourceAmount > 0 {
		grpTres = append(grpTres, client.TRES{
			Type:  resource.ResourceType,
			Count: mapping.ResourceAmount,
		})
	}

	grpTresMins := []client.TRES{}
	if mapping.ResourceTime > 0 {
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
	} else {
		slog.Info("Successfully upserted association for membership resource override creation", "association", association)
	}
}
