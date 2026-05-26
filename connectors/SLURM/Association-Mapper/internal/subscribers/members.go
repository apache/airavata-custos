package subscribers

import (
	"log/slog"

	"context"
	"time"

	client "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations"
	"github.com/apache/airavata-custos/pkg/models"
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipCreation(
	membership models.ComputeAllocationMembership) {

	slog.Info("Received compute allocation membership creation event", "membership", membership)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allocation, err := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipCreationFailed", membership.ID, "Failed to get compute allocation. Error: "+err.Error())
		return
	}

	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipCreationFailed", membership.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}

	user, err := a.coreService.GetUser(ctx, membership.UserID)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipCreationFailed", membership.ID, "Failed to get user. Error: "+err.Error())
		return
	}

	csu, err := a.coreService.GetComputeClusterUserByPair(ctx, cluster.ID, user.ID) // TODO: use this to get the local username for the association instead of assuming it's the same as the Airavata Custos username
	if err != nil {
		slog.Error("Failed to get compute cluster user by pair", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipCreationFailed", membership.ID, "Failed to get compute cluster user by pair. Error: "+err.Error())
		return
	}

	resources, err := a.coreService.ListResourcesForAllocation(ctx, allocation.ID) // TODO: use this to get the partition for the association instead of hardcoding it to "default"
	if err != nil {
		slog.Error("Failed to list resources for allocation", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipCreationFailed", membership.ID, "Failed to list resources for allocation. Error: "+err.Error())
		return
	}

	if len(resources) == 0 {
		slog.Warn("No resources found for allocation, defaulting to partition 'default'", "allocation_id", allocation.ID)
	}

	association := client.Association{
		Account:   allocation.Name,
		Cluster:   cluster.Name,
		User:      csu.LocalUsername,
		Partition: resources[0].Name, // TODO: do for each resource
		QoS:       []string{"normal"},
	}

	err = a.slurmClient.UpsertAssociation(association)
	if err != nil {
		slog.Error("Failed to upsert association", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipCreationFailed", membership.ID, "Failed to upsert association. Error: "+err.Error())
	} else {
		slog.Info("Successfully upserted association", "association", association)
		a.recordAuditEvent("ComputeAllocationMembershipCreationSucceeded", membership.ID, "Successfully upserted association.")
	}
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipResourceOverrideCreation(
	override models.ComputeAllocationMembershipResourceOverride) {

	slog.Info("Received compute allocation membership resource override creation event", "override", override)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// TODO: read per-resource override via
	membership, err := a.coreService.GetComputeAllocationMembership(ctx, override.ComputeAllocationMembershipID)
	if err != nil {
		slog.Error("Failed to get compute allocation membership for resource override creation", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to get compute allocation membership. Error: "+err.Error())
		return
	}

	allocationResource, err := a.coreService.GetComputeAllocationResource(ctx, override.ComputeAllocationResourceID)
	if err != nil {
		slog.Error("Failed to get compute allocation resource for resource override creation", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to get compute allocation resource. Error: "+err.Error())
		return
	}

	slog.Info("Received compute allocation membership resource override creation event",
		"membership", membership, "resource", allocationResource, "override", override)

	allocation, err := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to get compute allocation. Error: "+err.Error())
		return
	}

	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}

	user, err := a.coreService.GetUser(ctx, membership.UserID)
	if err != nil {
		slog.Error("Failed to get user", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to get user. Error: "+err.Error())
		return
	}

	csu, err := a.coreService.GetComputeClusterUserByPair(ctx, cluster.ID, user.ID) // TODO: use this to get the local username for the association instead of assuming it's the same as the Airavata Custos username
	if err != nil {
		slog.Error("Failed to get compute cluster user by pair", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to get compute cluster user by pair. Error: "+err.Error())
		return
	}

	grpTres := []client.TRES{}

	if override.OverrideResourceAmount > 0 {
		grpTres = append(grpTres, client.TRES{
			Type:  allocationResource.ResourceType,
			Count: override.OverrideResourceAmount, // override.OverrideResourceAmount is the resource amount granted to the user for this resource (e.g., number of CPUs, GPUs).
		})
	}

	grpTresMins := []client.TRES{}

	if override.OverrideResourceTime > 0 {
		grpTresMins = append(grpTresMins, client.TRES{
			Type:  allocationResource.ResourceType,
			Count: override.OverrideResourceTime, // override.OverrideResourceTime is the wall-clock time in minutes that the resource amount is granted for.
		})
	}

	limits := client.AssocLimits{
		GrpTRES:     grpTres,
		GrpTRESMins: grpTresMins,
	}

	association := client.Association{
		Account:   allocation.Name,
		Cluster:   cluster.Name,
		User:      csu.LocalUsername,
		Partition: allocationResource.Name,
		QoS:       []string{"normal"},
		Limits:    limits,
	}

	err = a.slurmClient.UpsertAssociation(association)
	if err != nil {
		slog.Error("Failed to upsert association for membership resource override creation", "error", err)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationFailed", override.ID, "Failed to upsert association. Error: "+err.Error())
	} else {
		slog.Info("Successfully upserted association for membership resource override creation", "association", association)
		a.recordAuditEvent("ComputeAllocationMembershipResourceOverrideCreationSucceeded", override.ID, "Successfully upserted association.")
	}
}
