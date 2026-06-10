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
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/internal/store"
)

func (s *Server) requireAmiePacketAuditStore(w http.ResponseWriter) (store.AmiePacketAuditStore, bool) {
	if s.admin == nil || s.admin.AmiePacketAudits == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("amie packet audit store not configured"))
		return nil, false
	}
	return s.admin.AmiePacketAudits, true
}

// handleListAmiePacketAudits returns every audit_events row written for one
// AMIE packet, joined through amie_audit_extras. Flat list, ordered by
// event_time; clients build any tree they want.
func (s *Server) handleListAmiePacketAudits(w http.ResponseWriter, r *http.Request) {
	st, ok := s.requireAmiePacketAuditStore(w)
	if !ok {
		return
	}
	packetID := strings.TrimSpace(r.PathValue("packet_id"))
	if packetID == "" {
		writeError(w, http.StatusBadRequest, errors.New("packet_id is required"))
		return
	}
	events, err := st.ListAuditsForPacket(r.Context(), packetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"packet_id": packetID,
		"events":    events,
	})
}
