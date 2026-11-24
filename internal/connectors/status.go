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

package connectors

import (
	"net/http"
	"slices"
	"strings"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// ConnectorStatus is a configured connector and its last known health.
type ConnectorStatus struct {
	Name             string                  `json:"name"`
	Type             string                  `json:"type"`
	Enabled          bool                    `json:"enabled"`
	FailedDeliveries int                     `json:"failed_deliveries"`
	Checks           []health.ConnectorCheck `json:"checks"`
}

// ConnectorListResponse wraps the connector list.
type ConnectorListResponse struct {
	Items []ConnectorStatus `json:"items"`
}

// @Summary	List configured connectors with their last known health
// @Description	Checks are the cached outcome of each connector's own work; nothing is probed on request.
// @Tags	Connectors
// @Security	BearerAuth
// @Produce	json
// @Success	200	{object}	ConnectorListResponse
// @Router	/connectors [get]
func listStatus(cfg map[string]*config.ConnectorConfig, bus *events.Bus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		failed, err := bus.CountDeliveriesBySubscriber(r.Context(), models.EventDeliveryFailed)
		if err != nil {
			common.WriteServiceError(w, err)
			return
		}
		items := []ConnectorStatus{}
		for name, c := range cfg {
			if c == nil {
				continue
			}
			items = append(items, ConnectorStatus{
				Name: name, Type: c.Type, Enabled: c.Enabled,
				FailedDeliveries: failed[c.Type], Checks: health.For(c.Type),
			})
		}
		slices.SortFunc(items, func(a, b ConnectorStatus) int { return strings.Compare(a.Name, b.Name) })
		common.WriteJSON(w, http.StatusOK, ConnectorListResponse{Items: items})
	}
}
