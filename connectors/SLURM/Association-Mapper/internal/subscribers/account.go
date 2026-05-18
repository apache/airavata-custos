package subscribers

import "github.com/apache/airavata-custos/pkg/models"
import "log/slog"

func (a *AssociationSubscriber) SubscribeToComputeAccountCreation(computeAccount models.ComputeAllocation) {
	slog.Info("Received compute account creation event", "account", computeAccount)
}

func (a *AssociationSubscriber) SubscribeToComputeAccountDeletion(computeAccount models.ComputeAllocation) {
	slog.Info("Received compute account deletion event", "account", computeAccount)
}

func (a *AssociationSubscriber) SubscribeToComputeAccountUpdate(computeAccount models.ComputeAllocation) {
	slog.Info("Received compute account update event", "account", computeAccount)
}
