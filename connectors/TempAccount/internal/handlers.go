package internal

import (
	"errors"
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
	mux.HandleFunc("/connectors/temp-account/assign-allocation", h.assignAllocationToTempAccount)
}

func (h *Handlers) createTempAccount(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := common.DecodeJSON(r, &u); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if u.Type != models.UserTypeVirtual {
		common.WriteError(w, http.StatusBadRequest, errors.New("User type must be VIRTUAL"))
		return
	}

	created, err := h.coreService.CreateUser(r.Context(), &u)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handlers) assignAllocationToTempAccount(w http.ResponseWriter, r *http.Request) {
	var m models.ComputeAllocationMembership
	if err := common.DecodeJSON(r, &m); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if m.UserID == "" || m.ComputeAllocationID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("UserID and ComputeAllocationID must be provided"))
		return
	}

	// In future, we may want to check that the user already has a membership to an allocation
	// If that's the case, prevent user from getting membershup to multiple allocations, which TempAccount connector doesn't support.

	user, err := h.coreService.GetUser(r.Context(), m.UserID)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}

	if user.Type != models.UserTypeVirtual {
		common.WriteError(w, http.StatusBadRequest, errors.New("User type must be VIRTUAL"))
		return
	}

	assigned, err := h.coreService.CreateComputeAllocationMembership(r.Context(), &m)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, assigned)
}
