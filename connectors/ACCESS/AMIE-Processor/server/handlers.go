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

// Package server exposes the AMIE connector's REST read endpoints.
package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
)

type Handlers struct {
	audits  store.PacketAuditStore
	packets store.PacketStore
}

func NewHandlers(audits store.PacketAuditStore, packets store.PacketStore) *Handlers {
	return &Handlers{audits: audits, packets: packets}
}

// RegisterRoutes attaches the AMIE connector's HTTP endpoints via router, gated
// by scope-specific read/write privileges.
func (h *Handlers) RegisterRoutes(router *identity.Router) {
	router.RequirePrivilege("GET /connectors/amie/packets", PacketsRead, h.listPackets)
	router.RequirePrivilege("GET /connectors/amie/packets/{id}", PacketsRead, h.getPacket)
	router.RequirePrivilege("GET /connectors/amie/packets/{id}/events", PacketsRead, h.listPacketEvents)
	router.RequirePrivilege("GET /connectors/amie/packets/{packet_id}/audits", PacketsRead, h.listPacketAudits)
	router.RequirePrivilege("GET /connectors/amie/stats", PacketsRead, h.getStats)
	router.RequirePrivilege("GET /connectors/amie/replies", RepliesRead, h.listReplies)
	router.RequirePrivilege("GET /connectors/amie/unmapped", UnmappedRead, h.listUnmapped)
	router.RequirePrivilege("POST /connectors/amie/packets/{id}/retry", PacketsWrite, h.retryPacket)
	router.RequirePrivilege("POST /connectors/amie/packets/{id}/resolve", PacketsWrite, h.resolvePacket)
	router.RequirePrivilege("POST /connectors/amie/replies/{id}/retry", RepliesWrite, h.retryReply)
	router.RequirePrivilege("POST /connectors/amie/unmapped/{id}/link", UnmappedWrite, h.linkUnmapped)
}

// @Summary	List audit events for an AMIE packet
// @Tags	AMIE Audit
// @Security	BearerAuth
// @Produce	json
// @Param	packet_id	path	string	true	"AMIE packet ID"
// @Success	200	{object}	object{packet_id=string,events=[]object{span_id=string,parent_span_id=string,source=string,event_type=string,entity_type=string,entity_id=string,description=string,status=string,created_at=string}}
// @Failure	400	{object}	object{error=string}	"packet_id is required"
// @Failure	500	{object}	object{error=string}	"Store lookup failed"
// @Failure	503	{object}	object{error=string}	"AMIE packet audit store not configured"
// @Router	/connectors/amie/packets/{packet_id}/audits [get]
func (h *Handlers) listPacketAudits(w http.ResponseWriter, r *http.Request) {
	if h.audits == nil {
		common.WriteError(w, http.StatusServiceUnavailable, errors.New("amie packet audit store not configured"))
		return
	}
	packetID := strings.TrimSpace(r.PathValue("packet_id"))
	if packetID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("packet_id is required"))
		return
	}
	events, err := h.audits.ListAuditsForPacket(r.Context(), packetID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"packet_id": packetID,
		"events":    events,
	})
}

// @Summary	List AMIE packets
// @Tags	AMIE Packets
// @Security	BearerAuth
// @Produce	json
// @Param	status	query	string	false	"Filter by status (NEW, DECODED, PROCESSED, FAILED, all)"
// @Param	type	query	string	false	"Filter by packet type"
// @Param	q	query	string	false	"Search by packet id or AMIE id"
// @Param	from	query	string	false	"RFC3339 lower bound on received_at"
// @Param	to	query	string	false	"RFC3339 upper bound on received_at"
// @Param	limit	query	int	false	"Page size (default 50, max 200)"
// @Param	offset	query	int	false	"Pagination offset"
// @Success	200	{object}	PacketListResponse
// @Failure	400	{object}	object{error=string}	"Invalid query parameter"
// @Failure	500	{object}	object{error=string}	"Store lookup failed"
// @Router	/connectors/amie/packets [get]
func (h *Handlers) listPackets(w http.ResponseWriter, r *http.Request) {
	if h.packets == nil {
		common.WriteJSON(w, http.StatusOK, emptyPacketPage(parseLimit(r), parseOffset(r)))
		return
	}
	f, err := parsePacketFilter(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	rows, total, err := h.packets.ListPackets(r.Context(), f)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]PacketResponse, 0, len(rows))
	for _, p := range rows {
		items = append(items, packetResponseFrom(p))
	}
	common.WriteJSON(w, http.StatusOK, PacketListResponse{
		Packets: items,
		Total:   total,
		Limit:   effectiveLimit(f.Limit),
		Offset:  f.Offset,
	})
}

// @Summary	Get an AMIE packet by ID
// @Tags	AMIE Packets
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"AMIE packet ID"
// @Success	200	{object}	PacketResponse
// @Failure	404	{object}	object{error=string}	"Packet not found"
// @Failure	500	{object}	object{error=string}	"Store lookup failed"
// @Router	/connectors/amie/packets/{id} [get]
func (h *Handlers) getPacket(w http.ResponseWriter, r *http.Request) {
	if h.packets == nil {
		common.WriteError(w, http.StatusNotFound, errors.New("not found"))
		return
	}
	p, err := h.packets.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if p == nil {
		common.WriteError(w, http.StatusNotFound, errors.New("not found"))
		return
	}
	common.WriteJSON(w, http.StatusOK, packetResponseFrom(*p))
}

// @Summary	List processing events for an AMIE packet
// @Tags	AMIE Packets
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"AMIE packet ID"
// @Success	200	{array}	PacketEventResponse
// @Failure	500	{object}	object{error=string}	"Store lookup failed"
// @Router	/connectors/amie/packets/{id}/events [get]
func (h *Handlers) listPacketEvents(w http.ResponseWriter, r *http.Request) {
	if h.packets == nil {
		common.WriteJSON(w, http.StatusOK, []PacketEventResponse{})
		return
	}
	rows, err := h.packets.ListPacketEvents(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	out := make([]PacketEventResponse, 0, len(rows))
	for _, e := range rows {
		out = append(out, packetEventResponseFrom(e))
	}
	common.WriteJSON(w, http.StatusOK, out)
}

// @Summary	Per-day packet stats grouped by status and type
// @Tags	AMIE Stats
// @Security	BearerAuth
// @Produce	json
// @Param	window	query	string	false	"Lookback window (e.g. 30d, 24h); default 30d"
// @Success	200	{object}	PacketStatsResponse
// @Failure	500	{object}	object{error=string}	"Store lookup failed"
// @Router	/connectors/amie/stats [get]
func (h *Handlers) getStats(w http.ResponseWriter, r *http.Request) {
	if h.packets == nil {
		common.WriteJSON(w, http.StatusOK, PacketStatsResponse{ByDay: []PacketStatBucketResponse{}})
		return
	}
	window := parseWindow(r.URL.Query().Get("window"))
	buckets, err := h.packets.GetStats(r.Context(), window)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, packetStatsResponseFrom(buckets))
}

// ReplyListResponse is the paginated list envelope for connector replies.
type ReplyListResponse struct {
	Replies []any `json:"replies"`
	Total   int   `json:"total"`
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
}

// @Summary	List replies sent to AMIE
// @Tags	AMIE Replies
// @Security	BearerAuth
// @Produce	json
// @Param	limit	query	int	false	"Page size (default 50, max 200)"
// @Param	offset	query	int	false	"Pagination offset"
// @Success	200	{object}	ReplyListResponse
// @Router	/connectors/amie/replies [get]
func (h *Handlers) listReplies(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, http.StatusOK, ReplyListResponse{
		Replies: []any{},
		Total:   0,
		Limit:   effectiveLimit(parseLimit(r)),
		Offset:  parseOffset(r),
	})
}

// @Summary	List AMIE packets that could not be mapped to a Custos entity
// @Tags	AMIE Unmapped
// @Security	BearerAuth
// @Produce	json
// @Param	limit	query	int	false	"Page size (default 50, max 200)"
// @Param	offset	query	int	false	"Pagination offset"
// @Success	200	{object}	PacketListResponse
// @Router	/connectors/amie/unmapped [get]
func (h *Handlers) listUnmapped(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, http.StatusOK, emptyPacketPage(parseLimit(r), parseOffset(r)))
}

// @Summary	Retry an AMIE packet (not yet implemented)
// @Tags	AMIE Packets
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"AMIE packet ID"
// @Failure	501	{object}	object{error=string,message=string}
// @Router	/connectors/amie/packets/{id}/retry [post]
func (h *Handlers) retryPacket(w http.ResponseWriter, _ *http.Request) {
	writeNotImplemented(w, "packet retry")
}

// @Summary	Resolve an AMIE packet (not yet implemented)
// @Tags	AMIE Packets
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"AMIE packet ID"
// @Failure	501	{object}	object{error=string,message=string}
// @Router	/connectors/amie/packets/{id}/resolve [post]
func (h *Handlers) resolvePacket(w http.ResponseWriter, _ *http.Request) {
	writeNotImplemented(w, "packet resolve")
}

// @Summary	Retry an AMIE reply (not yet implemented)
// @Tags	AMIE Replies
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"AMIE reply ID"
// @Failure	501	{object}	object{error=string,message=string}
// @Router	/connectors/amie/replies/{id}/retry [post]
func (h *Handlers) retryReply(w http.ResponseWriter, _ *http.Request) {
	writeNotImplemented(w, "reply retry")
}

// @Summary	Link an unmapped packet to a Custos entity (not yet implemented)
// @Tags	AMIE Unmapped
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"AMIE packet ID"
// @Failure	501	{object}	object{error=string,message=string}
// @Router	/connectors/amie/unmapped/{id}/link [post]
func (h *Handlers) linkUnmapped(w http.ResponseWriter, _ *http.Request) {
	writeNotImplemented(w, "unmapped link")
}

func writeNotImplemented(w http.ResponseWriter, op string) {
	common.WriteJSON(w, http.StatusNotImplemented, map[string]string{
		"error":   "not_implemented",
		"message": op + " not supported on this branch",
	})
}

func emptyPacketPage(limit, offset int) PacketListResponse {
	return PacketListResponse{
		Packets: []PacketResponse{},
		Total:   0,
		Limit:   effectiveLimit(limit),
		Offset:  offset,
	}
}

func parsePacketFilter(r *http.Request) (store.PacketListFilter, error) {
	q := r.URL.Query()
	f := store.PacketListFilter{
		Status: q.Get("status"),
		Type:   q.Get("type"),
		Query:  q.Get("q"),
		Limit:  parseLimit(r),
		Offset: parseOffset(r),
	}
	if v := q.Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return f, errors.New("invalid 'from' query parameter; expected RFC 3339")
		}
		f.From = t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return f, errors.New("invalid 'to' query parameter; expected RFC 3339")
		}
		f.To = t
	}
	return f, nil
}

func parseLimit(r *http.Request) int {
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 0
}

func parseOffset(r *http.Request) int {
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return 0
}

func effectiveLimit(l int) int {
	if l <= 0 || l > 200 {
		return 50
	}
	return l
}

func parseWindow(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 30 * 24 * time.Hour
	}
	if strings.HasSuffix(v, "d") {
		if n, err := strconv.Atoi(strings.TrimSuffix(v, "d")); err == nil && n > 0 {
			return time.Duration(n) * 24 * time.Hour
		}
	}
	if d, err := time.ParseDuration(v); err == nil && d > 0 {
		return d
	}
	return 30 * 24 * time.Hour
}
