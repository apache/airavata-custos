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

// Package health keeps the last known result of each connector's own work
// against an outside system, such as the SLURM mapper reading a cluster. It is
// in memory, so a target is unknown after a restart until its next check.
package health

import (
	"slices"
	"strings"
	"sync"
	"time"
)

// ConnectorCheck is one connector's last result against one target.
type ConnectorCheck struct {
	Target    string    `json:"target"`
	OK        bool      `json:"ok"`
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

var (
	mu     sync.RWMutex
	checks = map[string]map[string]ConnectorCheck{}
)

// Record stores the outcome of connectorType's latest call to target.
func Record(connectorType, target string, err error) {
	c := ConnectorCheck{Target: target, OK: err == nil, CheckedAt: time.Now().UTC()}
	if err != nil {
		c.Error = err.Error()
	}
	mu.Lock()
	defer mu.Unlock()
	if checks[connectorType] == nil {
		checks[connectorType] = map[string]ConnectorCheck{}
	}
	checks[connectorType][target] = c
}

// For returns connectorType's last results, ordered by target.
func For(connectorType string) []ConnectorCheck {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]ConnectorCheck, 0, len(checks[connectorType]))
	for _, c := range checks[connectorType] {
		out = append(out, c)
	}
	slices.SortFunc(out, func(a, b ConnectorCheck) int { return strings.Compare(a.Target, b.Target) })
	return out
}
