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

import "testing"

func TestEventStatus(t *testing.T) {
	cases := map[string]string{
		"CREATE_PERSON":                  StatusOk,
		"ComanageClusterAccountAttached": StatusOk,
		"TRANSACTION_COMPLETE":           StatusOk,
		"ComanageProvisioningFailed":     StatusError,
		"REQUEST_REJECTED":               StatusError,
		"some.error.happened":            StatusError,
		"":                               StatusOk,
	}
	for in, want := range cases {
		if got := EventStatus(in); got != want {
			t.Errorf("EventStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTraceStatus(t *testing.T) {
	RegisterMarkers("status-test", "PACKET_RECEIVED", "TRANSACTION_COMPLETE", "REPLY_HELD")
	cases := []struct {
		name   string
		events []string
		want   string
	}{
		{"start without terminal", []string{"PACKET_RECEIVED", "CREATE_PERSON"}, StatusInProgress},
		{"no start", []string{"CREATE_PERSON"}, StatusOk},
		{"start and terminal", []string{"PACKET_RECEIVED", "REPLY_HELD"}, StatusOk},
		{"error wins", []string{"PACKET_RECEIVED", "REQUEST_REJECTED", "TRANSACTION_COMPLETE"}, StatusError},
	}
	for _, c := range cases {
		events := make([]TraceEventStatus, len(c.events))
		for i, et := range c.events {
			events[i] = TraceEventStatus{Source: "status-test", EventType: et}
		}
		if got := TraceStatus(events); got != c.want {
			t.Errorf("%s: TraceStatus = %q, want %q", c.name, got, c.want)
		}
	}
}
