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

import "strings"

const (
	StatusOk         = "ok"
	StatusError      = "error"
	StatusInProgress = "in_progress"
)

var errorMarkers = []string{"failed", "error", "rejected"}

// EventStatus reads a step's outcome from its name, since a step records no
// status. The trace's status is decided in the audit trace store from its
// steps and its deliveries.
func EventStatus(eventType string) string {
	lower := strings.ToLower(eventType)
	for _, marker := range errorMarkers {
		if strings.Contains(lower, marker) {
			return StatusError
		}
	}
	return StatusOk
}
