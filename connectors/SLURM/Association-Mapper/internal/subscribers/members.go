package subscribers

import (
	"log/slog"

	"context"
	client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
	"github.com/apache/airavata-custos/pkg/models"
	"time"
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipCreation(membership models.ComputeAllocationMembership) {
	slog.Info("Received compute allocation membership creation event", "membership", membership)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allocation, err := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation", "error", err)
		return
	}

	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster", "error", err)
		return
	}

	user, err := a.coreService.GetUser(ctx, membership.UserID)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		return
	}

	csu, err := a.coreService.GetComputeClusterUserByPair(ctx, cluster.ID, user.ID) // TODO: use this to get the local username for the association instead of assuming it's the same as the Airavata Custos username
	if err != nil {
		slog.Error("Failed to get compute cluster user by pair", "error", err)
		return
	}

	resources, err := a.coreService.ListResourcesForAllocation(ctx, allocation.ID) // TODO: use this to get the partition for the association instead of hardcoding it to "default"
	if err != nil {
		slog.Error("Failed to list resources for allocation", "error", err)
		return
	}

	if len(resources) == 0 {
		slog.Warn("No resources found for allocation, defaulting to partition 'default'", "allocation_id", allocation.ID)
	}
	association := client.Association{
		Account:   allocation.Name,
		Cluster:   cluster.Name,
		User:      csu.LocalUsername,
		Partition: resources[0].Name,
	}

	err = a.slurmClient.UpsertAssociation(association)
	if err != nil {
		slog.Error("Failed to upsert association", "error", err)
	} else {
		slog.Info("Successfully upserted association", "association", association)
	}
}
