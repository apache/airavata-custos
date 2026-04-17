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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	packetsReceived    *prometheus.CounterVec
	packetsProcessed   *prometheus.CounterVec
	eventsRetry        prometheus.Counter
	processingDuration *prometheus.HistogramVec
	pollerFetched      prometheus.Counter
}

func New() *Metrics {
	return NewWithRegistry(prometheus.DefaultRegisterer)
}

// NewWithRegistry accepts a custom Registerer so tests can use an isolated
// registry and avoid double-registration panics.
func NewWithRegistry(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		packetsReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "amie_packets_received_total",
			Help: "Total number of AMIE packets received from the API",
		}, []string{"type"}),

		packetsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "amie_packets_processed_total",
			Help: "Total number of AMIE packets that completed processing",
		}, []string{"type", "outcome"}),

		eventsRetry: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "amie_events_retry_total",
			Help: "Total number of AMIE processing event retry attempts",
		}),

		processingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "amie_packet_processing_duration_seconds",
			Help:    "Time taken to process an AMIE packet by handler type",
			Buckets: prometheus.DefBuckets,
		}, []string{"handler"}),

		pollerFetched: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "amie_poller_packets_fetched",
			Help: "Total number of AMIE packets fetched by the poller",
		}),
	}

	reg.MustRegister(
		m.packetsReceived,
		m.packetsProcessed,
		m.eventsRetry,
		m.processingDuration,
		m.pollerFetched,
	)

	return m
}

func (m *Metrics) RecordPacketReceived(packetType string) {
	m.packetsReceived.WithLabelValues(packetType).Inc()
}

func (m *Metrics) RecordPacketProcessed(packetType, outcome string) {
	m.packetsProcessed.WithLabelValues(packetType, outcome).Inc()
}

func (m *Metrics) RecordRetry() {
	m.eventsRetry.Inc()
}

// StartProcessingTimer returns a stop function that, when called, records the
// elapsed duration as a histogram observation for the given handler type.
//
// Usage:
//
//	stop := metrics.StartProcessingTimer()
//	defer stop("MyHandler")
func (m *Metrics) StartProcessingTimer() func(handlerType string) {
	start := time.Now()
	return func(handlerType string) {
		duration := time.Since(start).Seconds()
		m.processingDuration.WithLabelValues(handlerType).Observe(duration)
	}
}

func (m *Metrics) RecordPollerFetch(count int) {
	m.pollerFetched.Add(float64(count))
}
