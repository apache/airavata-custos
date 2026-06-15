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
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	traceIDHexLen     = 32
	spanIDHexLen      = 16
	maxTraceListLimit = 200
	defaultTraceLimit = 50
	maxTraceOffset    = 1_000_000
	maxAuditWindow    = 365 * 24 * time.Hour
)

func (s *Server) requireAuditStore(w http.ResponseWriter) (store.AuditTraceStore, bool) {
	if s.admin == nil || s.admin.AuditTraces == nil {
		common.WriteError(w, http.StatusServiceUnavailable, errors.New("audit trace store not configured"))
		return nil, false
	}
	return s.admin.AuditTraces, true
}

// @Summary	List audit traces with filters
// @Description	One row per distinct trace_id. `q` matches trace_id hex prefix or root-operation substring.
// @Tags	Audit
// @Security	CustosUserHeader
// @Produce	json
// @Param	source	query	[]string	false	"Restrict to listed sources (repeatable)"	collectionFormat(multi)
// @Param	status	query	[]string	false	"ok | error | in_progress (repeatable)"	collectionFormat(multi)
// @Param	from	query	string	false	"RFC3339 lower bound on started_at"
// @Param	to	query	string	false	"RFC3339 upper bound on started_at"
// @Param	q	query	string	false	"Substring match on trace_id hex or root operation"
// @Param	limit	query	integer	false	"Page size (default 50, max 200)"
// @Param	offset	query	integer	false	"Page offset (default 0)"
// @Success	200	{object}	object{traces=[]models.TraceSummary,total=integer,limit=integer,offset=integer}
// @Failure	400	{object}	object{error=string}	"Invalid filter value"
// @Failure	503	{object}	object{error=string}	"Audit trace store not configured"
// @Router	/audit/traces [get]
func (s *Server) handleListTraces(w http.ResponseWriter, r *http.Request) {
	ts, ok := s.requireAuditStore(w)
	if !ok {
		return
	}
	f, err := parseTraceFilter(r)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	rows, total, err := ts.ListTraces(r.Context(), f)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"traces": rows,
		"total":  total,
		"limit":  f.Limit,
		"offset": f.Offset,
	})
}

// @Summary	Get a trace's nested span tree
// @Tags	Audit
// @Security	CustosUserHeader
// @Produce	json
// @Param	trace_id	path	string	true	"32-char lowercase hex"
// @Success	200	{object}	object{trace_id=string,tree=[]models.TraceNode,truncated=boolean}
// @Failure	400	{object}	object{error=string}	"Malformed trace_id"
// @Failure	404	{object}	object{error=string}	"Trace not found"
// @Failure	503	{object}	object{error=string}	"Audit trace store not configured"
// @Router	/audit/traces/{trace_id} [get]
func (s *Server) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	ts, ok := s.requireAuditStore(w)
	if !ok {
		return
	}
	traceID, err := validateHexID(r.PathValue("trace_id"), traceIDHexLen)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	tree, truncated, err := ts.GetTraceTree(r.Context(), traceID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if tree == nil {
		common.WriteError(w, http.StatusNotFound, errors.New("trace not found"))
		return
	}

	resp := map[string]any{
		"trace_id":  traceID,
		"tree":      tree.Children,
		"truncated": truncated,
	}
	common.WriteJSON(w, http.StatusOK, resp)
}

// @Summary	List audit events for a trace (optionally one span)
// @Tags	Audit
// @Security	CustosUserHeader
// @Produce	json
// @Param	trace_id	query	string	true	"32-char lowercase hex"
// @Param	span_id	query	string	false	"16-char lowercase hex"
// @Success	200	{object}	object{events=[]models.TraceEvent}
// @Failure	400	{object}	object{error=string}	"Malformed trace_id or span_id"
// @Failure	503	{object}	object{error=string}	"Audit trace store not configured"
// @Router	/audit/events [get]
func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	ts, ok := s.requireAuditStore(w)
	if !ok {
		return
	}
	q := r.URL.Query()
	traceID, err := validateHexID(q.Get("trace_id"), traceIDHexLen)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	var spanID string
	if raw := q.Get("span_id"); raw != "" {
		spanID, err = validateHexID(raw, spanIDHexLen)
		if err != nil {
			common.WriteError(w, http.StatusBadRequest, err)
			return
		}
	}
	events, err := ts.ListEvents(r.Context(), traceID, spanID)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if events == nil {
		events = []models.TraceEvent{}
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"events": events})
}

// @Summary	List distinct audit-event sources
// @Tags	Audit
// @Security	CustosUserHeader
// @Produce	json
// @Success	200	{object}	object{sources=[]string}
// @Failure	503	{object}	object{error=string}	"Audit trace store not configured"
// @Router	/audit/sources [get]
func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	ts, ok := s.requireAuditStore(w)
	if !ok {
		return
	}
	sources, err := ts.ListSources(r.Context())
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"sources": sources})
}

func parseTraceFilter(r *http.Request) (store.TraceFilter, error) {
	q := r.URL.Query()
	f := store.TraceFilter{Q: strings.TrimSpace(q.Get("q"))}
	if vals := q["source"]; len(vals) > 0 {
		f.Sources = splitMulti(vals)
	}
	for _, raw := range splitMulti(q["status"]) {
		if !isValidStatus(raw) {
			return f, errors.New("invalid status: " + raw)
		}
		f.Statuses = append(f.Statuses, raw)
	}
	if raw := q.Get("from"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return f, errors.New("invalid from: expected RFC3339")
		}
		f.From = t
	}
	if raw := q.Get("to"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return f, errors.New("invalid to: expected RFC3339")
		}
		f.To = t
	}
	if !f.From.IsZero() && !f.To.IsZero() && f.To.Sub(f.From) > maxAuditWindow {
		return f, errors.New("window exceeds maximum of 365d")
	}
	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return f, errors.New("invalid limit")
		}
		if n > maxTraceListLimit {
			n = maxTraceListLimit
		}
		f.Limit = n
	}
	if f.Limit == 0 {
		f.Limit = defaultTraceLimit
	}
	if raw := q.Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return f, errors.New("invalid offset")
		}
		if n > maxTraceOffset {
			return f, errors.New("offset exceeds maximum")
		}
		f.Offset = n
	}
	return f, nil
}

func isValidStatus(raw string) bool {
	switch raw {
	case "ok", "error", "in_progress":
		return true
	}
	return false
}

// splitMulti handles repeated AND comma-separated values for one param.
func splitMulti(vals []string) []string {
	var out []string
	for _, v := range vals {
		for _, part := range strings.Split(v, ",") {
			if p := strings.TrimSpace(part); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func validateHexID(raw string, wantLen int) (string, error) {
	if len(raw) != wantLen {
		return "", errors.New("expected " + strconv.Itoa(wantLen) + "-char lowercase hex id, got len " + strconv.Itoa(len(raw)))
	}
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return "", errors.New("invalid hex id")
		}
	}
	return raw, nil
}
