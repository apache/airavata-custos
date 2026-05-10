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

package metrics_test

import (
	"testing"
	"time"

	"github.com/apache/airavata-custos/allocations/access-amie/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestMetrics builds a Metrics instance backed by a fresh isolated
// registry so test runs do not conflict with each other or the default
// global registry.
func newTestMetrics(reg prometheus.Registerer) *metrics.Metrics {
	return metrics.NewWithRegistry(reg)
}

// counterValue reads the current value of a counter registered in reg.
func counterValue(t *testing.T, reg prometheus.Gatherer, name string, labels map[string]string) float64 {
	t.Helper()
	mfs, err := reg.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if labelsMatch(m.GetLabel(), labels) {
				return m.GetCounter().GetValue()
			}
		}
	}
	t.Fatalf("metric %q with labels %v not found", name, labels)
	return 0
}

func labelsMatch(pairs []*dto.LabelPair, want map[string]string) bool {
	if len(pairs) != len(want) {
		return false
	}
	for _, lp := range pairs {
		if v, ok := want[lp.GetName()]; !ok || v != lp.GetValue() {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRecordPacketReceived_IncrementsCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := newTestMetrics(reg)

	m.RecordPacketReceived("request_account")
	m.RecordPacketReceived("request_account")
	m.RecordPacketReceived("notify_person_modify")

	val := counterValue(t, reg, "amie_packets_received_total", map[string]string{"type": "request_account"})
	assert.Equal(t, float64(2), val)

	val2 := counterValue(t, reg, "amie_packets_received_total", map[string]string{"type": "notify_person_modify"})
	assert.Equal(t, float64(1), val2)
}

func TestRecordPacketProcessed_DifferentOutcomes(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := newTestMetrics(reg)

	m.RecordPacketProcessed("request_account", "succeeded")
	m.RecordPacketProcessed("request_account", "succeeded")
	m.RecordPacketProcessed("request_account", "retry_scheduled")
	m.RecordPacketProcessed("notify_project_activate", "permanently_failed")

	success := counterValue(t, reg, "amie_packets_processed_total", map[string]string{
		"type": "request_account", "outcome": "succeeded",
	})
	assert.Equal(t, float64(2), success)

	retry := counterValue(t, reg, "amie_packets_processed_total", map[string]string{
		"type": "request_account", "outcome": "retry_scheduled",
	})
	assert.Equal(t, float64(1), retry)

	failed := counterValue(t, reg, "amie_packets_processed_total", map[string]string{
		"type": "notify_project_activate", "outcome": "permanently_failed",
	})
	assert.Equal(t, float64(1), failed)
}

func TestRecordRetry_IncrementsCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := newTestMetrics(reg)

	m.RecordRetry()
	m.RecordRetry()
	m.RecordRetry()

	mfs, err := reg.Gather()
	require.NoError(t, err)

	var retryVal float64
	for _, mf := range mfs {
		if mf.GetName() == "amie_events_retry_total" {
			for _, metric := range mf.GetMetric() {
				retryVal = metric.GetCounter().GetValue()
			}
		}
	}
	assert.Equal(t, float64(3), retryVal)
}

func TestStartProcessingTimer_ReturnsWorkingStopFunction(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := newTestMetrics(reg)

	stop := m.StartProcessingTimer()
	time.Sleep(5 * time.Millisecond)
	stop("request_account")

	// Verify a histogram observation was recorded.
	mfs, err := reg.Gather()
	require.NoError(t, err)

	var found bool
	for _, mf := range mfs {
		if mf.GetName() == "amie_packet_processing_duration_seconds" {
			for _, metric := range mf.GetMetric() {
				for _, lp := range metric.GetLabel() {
					if lp.GetName() == "handler" && lp.GetValue() == "request_account" {
						found = true
						assert.Equal(t, uint64(1), metric.GetHistogram().GetSampleCount())
					}
				}
			}
		}
	}
	assert.True(t, found, "histogram observation for 'request_account' handler not found")
}

func TestRecordPollerFetch_AddsCount(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := newTestMetrics(reg)

	m.RecordPollerFetch(5)
	m.RecordPollerFetch(3)

	mfs, err := reg.Gather()
	require.NoError(t, err)

	var total float64
	for _, mf := range mfs {
		if mf.GetName() == "amie_poller_packets_fetched" {
			for _, metric := range mf.GetMetric() {
				total = metric.GetCounter().GetValue()
			}
		}
	}
	assert.Equal(t, float64(8), total)
}
