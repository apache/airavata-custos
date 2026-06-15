package internal

import (
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
	"net/http"
)

type Handlers struct {
	coreService *service.Service
}

func NewHandlers() *Handlers {
	return &Handlers{}
}

// RegisterRoutes attaches the TempAccount connector's HTTP endpoints to mux.
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/connectors/temp-account/create", h.createTempAccount)
}

func (h *Handlers) createTempAccount(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := common.DecodeJSON(r, &u); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := h.coreService.CreateUser(r.Context(), &u); err != nil {
		common.WriteServiceError(w, err)
		return
	}
}
