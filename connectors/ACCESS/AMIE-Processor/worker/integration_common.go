//go:build integration

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

package worker

import (
	"context"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amietest"
	"github.com/apache/airavata-custos/internal/db/dbtest"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	// Worker tests touch connector tables only.
	return amietest.OpenDB(t,
		"amie_audit_extras",
		"audit_events",
		"amie_processing_errors",
		"amie_processing_events",
		"amie_packets",
		"amie_user_dns",
	)
}

// stubAmieClient returns canned responses/errors per call so tests can drive
// multi-call sequences deterministically.
type stubAmieClient struct {
	mu        sync.Mutex
	responses [][]map[string]any
	errors    []error
	calls     int
}

func (s *stubAmieClient) FetchInProgressPackets(_ context.Context) ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.calls
	s.calls++
	if idx < len(s.errors) && s.errors[idx] != nil {
		return nil, s.errors[idx]
	}
	if idx < len(s.responses) {
		return s.responses[idx], nil
	}
	return nil, nil
}

func (s *stubAmieClient) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

// stubMetrics counts calls instead of touching Prometheus.
type stubMetrics struct {
	mu                sync.Mutex
	packetsReceived   []string
	fetchCounts       []int
	processedOutcomes []string
	retries           int
	timerStops        int
}

func (m *stubMetrics) RecordPacketReceived(packetType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.packetsReceived = append(m.packetsReceived, packetType)
}

func (m *stubMetrics) RecordPollerFetch(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fetchCounts = append(m.fetchCounts, count)
}

func (m *stubMetrics) RecordPacketProcessed(packetType, outcome string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processedOutcomes = append(m.processedOutcomes, outcome)
}

func (m *stubMetrics) RecordRetry() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retries++
}

func (m *stubMetrics) StartProcessingTimer() func(string) {
	return func(string) {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.timerStops++
	}
}

var countRows = dbtest.CountRows
