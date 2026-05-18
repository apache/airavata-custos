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
