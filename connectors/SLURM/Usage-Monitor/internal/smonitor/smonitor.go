package smonitor

import (
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
)

type SlurmMonitor struct {
	slurmClient *client.Client
	eventBus    *events.Bus
	coreService service.CoreService
}

func NewSlurmMonitor(slurmClient *client.Client, eventBus *events.Bus, coreService service.CoreService) *SlurmMonitor {
	return &SlurmMonitor{
		slurmClient: slurmClient,
		eventBus:    eventBus,
		coreService: coreService,
	}
}

func (m *SlurmMonitor) StartMonitor() {

}
