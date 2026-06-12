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

// Package server exposes the AMIE connector's HTTP read surface.
package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
)

type Handlers struct {
	audits store.PacketAuditStore
}

func NewHandlers(audits store.PacketAuditStore) *Handlers {
	return &Handlers{audits: audits}
}

// RegisterRoutes attaches the AMIE connector's HTTP endpoints to mux.
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /connectors/amie/packets/{packet_id}/audits", h.listPacketAudits)
}

// listPacketAudits returns every audit_events row written for one AMIE packet,
// joined through amie_audit_extras.
func (h *Handlers) listPacketAudits(w http.ResponseWriter, r *http.Request) {
	if h.audits == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("amie packet audit store not configured"))
		return
	}
	packetID := strings.TrimSpace(r.PathValue("packet_id"))
	if packetID == "" {
		writeError(w, http.StatusBadRequest, errors.New("packet_id is required"))
		return
	}
	events, err := h.audits.ListAuditsForPacket(r.Context(), packetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"packet_id": packetID,
		"events":    events,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
