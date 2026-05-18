package connectors

import "github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/pkg/smapper"
import "github.com/apache/airavata-custos/pkg/events"
import "log/slog"

func LoadConnectors(eventBus *events.Bus) error {

	slog.Info("loading connectors")

	slog.Info("loading SLURM Association Mapper connector")
	err := smapper.LoadConnector(eventBus)
	if err != nil {
		slog.Error("failed to load SLURM Association Mapper connector", "error", err)
		return err
	}

	slog.Info("finished loading connectors")
	return err
}
