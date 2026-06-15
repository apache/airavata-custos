package pkg

import (
	"context"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/service"
	"github.com/jmoiron/sqlx"
	"net/http"
	"sync"
)

func LoadConnector(ctx context.Context, _ *sqlx.DB, eventBus *events.Bus, coreService *service.Service, wg *sync.WaitGroup, _ *http.ServeMux, connectorConfig *config.ConnectorConfig) error {

	return nil
}
