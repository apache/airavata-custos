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

package tracing

import (
	"slices"
	"strings"
	"sync"
)

const (
	StatusOk         = "ok"
	StatusError      = "error"
	StatusInProgress = "in_progress"
)

var errorMarkers = []string{"failed", "error", "rejected"}

func EventStatus(eventType string) string {
	lower := strings.ToLower(eventType)
	for _, marker := range errorMarkers {
		if strings.Contains(lower, marker) {
			return StatusError
		}
	}
	return StatusOk
}

type traceMarkers struct {
	start     string
	terminals []string
}

var (
	markersMu sync.RWMutex
	markers   = map[string]traceMarkers{}
)

// RegisterMarkers declares the event that opens a source's work in a trace and
// the events that close it. Connectors call this at boot so the core stays
// unaware of connector-specific event names.
func RegisterMarkers(source, start string, terminals ...string) {
	markersMu.Lock()
	defer markersMu.Unlock()
	markers[source] = traceMarkers{start, terminals}
}

type TraceEventStatus struct {
	Source    string
	EventType string
}

// TraceStatus is "error" if any event errored, "in_progress" while a source has
// written its start marker and none of its terminals, else "ok". A source with
// no registered markers never holds a trace in progress.
func TraceStatus(events []TraceEventStatus) string {
	markersMu.RLock()
	defer markersMu.RUnlock()
	started, done := map[string]bool{}, map[string]bool{}
	for _, e := range events {
		if EventStatus(e.EventType) == StatusError {
			return StatusError
		}
		m, ok := markers[e.Source]
		started[e.Source] = started[e.Source] || ok && e.EventType == m.start
		done[e.Source] = done[e.Source] || slices.Contains(m.terminals, e.EventType)
	}
	for s := range started {
		if started[s] && !done[s] {
			return StatusInProgress
		}
	}
	return StatusOk
}
