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

var (
	terminalMarkersMu sync.RWMutex
	terminalMarkers   = map[string][]string{}
)

// RegisterTerminalMarkers declares the event names that close out a trace for
// the given source. Connectors call this at boot so the core stays unaware of
// connector-specific event names.
func RegisterTerminalMarkers(source string, markers ...string) {
	terminalMarkersMu.Lock()
	defer terminalMarkersMu.Unlock()
	terminalMarkers[source] = append(terminalMarkers[source], markers...)
}

type TraceEventStatus struct {
	Source    string
	EventType string
}

func setTerminalMarkersForTest(t interface{ Cleanup(func()) }, source string, markers []string) {
	terminalMarkersMu.Lock()
	prev, hadPrev := terminalMarkers[source]
	terminalMarkers[source] = markers
	terminalMarkersMu.Unlock()
	t.Cleanup(func() {
		terminalMarkersMu.Lock()
		defer terminalMarkersMu.Unlock()
		if hadPrev {
			terminalMarkers[source] = prev
		} else {
			delete(terminalMarkers, source)
		}
	})
}

// TraceStatus is "error" if any event errored, "ok" if a registered terminal
// marker is present, else "in_progress".
func TraceStatus(events []TraceEventStatus) string {
	terminalMarkersMu.RLock()
	snapshot := make(map[string][]string, len(terminalMarkers))
	for k, v := range terminalMarkers {
		snapshot[k] = v
	}
	terminalMarkersMu.RUnlock()

	hasError := false
	hasTerminal := false
	for _, e := range events {
		if EventStatus(e.EventType) == StatusError {
			hasError = true
		}
		for _, m := range snapshot[e.Source] {
			if e.EventType == m {
				hasTerminal = true
			}
		}
	}
	if hasError {
		return StatusError
	}
	if hasTerminal {
		return StatusOk
	}
	return StatusInProgress
}
