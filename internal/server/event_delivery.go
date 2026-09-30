// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	defaultDeliveryLimit = 50
	maxDeliveryLimit     = 200
)

// @Summary	List event deliveries
// @Description	Newest first, every status unless one is given. A pending delivery with no attempts means its subscriber is not loaded, so it waits until that connector starts.
// @Tags	Events
// @Security	BearerAuth
// @Produce	json
// @Param	status	query	string	false	"PENDING | SUCCEEDED | FAILED"
// @Param	limit	query	integer	false	"Page size (default 50, max 200)"
// @Success	200	{object}	object{items=[]models.PendingDelivery}
// @Failure	400	{object}	object{error=string}	"Invalid status or limit"
// @Router	/events/deliveries [get]
func (s *Server) listEventDeliveries(w http.ResponseWriter, r *http.Request) {
	status := models.EventDeliveryStatus(r.URL.Query().Get("status"))
	switch status {
	case "", models.EventDeliveryPending, models.EventDeliverySucceeded, models.EventDeliveryFailed:
	default:
		common.WriteError(w, http.StatusBadRequest, fmt.Errorf("unknown status %q", status))
		return
	}
	limit := atoiOr(r.URL.Query().Get("limit"), defaultDeliveryLimit)
	if limit < 1 || limit > maxDeliveryLimit {
		common.WriteError(w, http.StatusBadRequest, fmt.Errorf("limit must be between 1 and %d", maxDeliveryLimit))
		return
	}
	rows, err := s.svc.EventBus().ListDeliveries(r.Context(), status, limit)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"items": rows})
}

// @Summary	Retry a failed event delivery
// @Description	Makes the delivery due again with a fresh attempt count. The worker picks it up on its next pass.
// @Tags	Events
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Delivery id"
// @Success	204
// @Failure	404	{object}	object{error=string}	"Delivery not found"
// @Failure	409	{object}	object{error=string}	"Delivery has not failed"
// @Router	/events/deliveries/{id}/retry [post]
func (s *Server) retryEventDelivery(w http.ResponseWriter, r *http.Request) {
	err := s.svc.EventBus().RetryDelivery(r.Context(), r.PathValue("id"))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, events.ErrDeliveryNotFound):
		common.WriteError(w, http.StatusNotFound, err)
	case errors.Is(err, events.ErrDeliveryNotFailed):
		common.WriteError(w, http.StatusConflict, err)
	default:
		common.WriteServiceError(w, err)
	}
}

// @Summary	List event subscriptions
// @Description	Every saved subscription and whether its subscriber is loaded in this process.
// @Tags	Events
// @Security	BearerAuth
// @Produce	json
// @Success	200	{object}	object{items=[]events.Subscription}
// @Router	/events/subscriptions [get]
func (s *Server) listEventSubscriptions(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.EventBus().ListSubscriptions(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"items": rows})
}
