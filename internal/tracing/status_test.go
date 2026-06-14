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

func TestTraceStatusOk(t *testing.T) {
	setTerminalMarkersForTest(t, "amie", []string{"TRANSACTION_COMPLETE"})
	events := []TraceEventStatus{
		{Source: "amie", EventType: "CREATE_PERSON"},
		{Source: "amie", EventType: "CREATE_ACCOUNT"},
		{Source: "amie", EventType: "TRANSACTION_COMPLETE"},
	}
	if got := TraceStatus(events); got != StatusOk {
		t.Errorf("TraceStatus(ok flow) = %q, want %q", got, StatusOk)
	}
}

func TestTraceStatusError(t *testing.T) {
	setTerminalMarkersForTest(t, "amie", []string{"TRANSACTION_COMPLETE"})
	events := []TraceEventStatus{
		{Source: "amie", EventType: "CREATE_PERSON"},
		{Source: "comanage", EventType: "ComanageProvisioningFailed"},
		{Source: "amie", EventType: "TRANSACTION_COMPLETE"},
	}
	if got := TraceStatus(events); got != StatusError {
		t.Errorf("TraceStatus(error wins over terminal) = %q, want %q", got, StatusError)
	}
}

func TestTraceStatusInProgress(t *testing.T) {
	events := []TraceEventStatus{
		{Source: "amie", EventType: "CREATE_PERSON"},
		{Source: "amie", EventType: "CREATE_ACCOUNT"},
	}
	if got := TraceStatus(events); got != StatusInProgress {
		t.Errorf("TraceStatus(no terminal) = %q, want %q", got, StatusInProgress)
	}
}

func TestTraceStatusEmpty(t *testing.T) {
	if got := TraceStatus(nil); got != StatusInProgress {
		t.Errorf("TraceStatus(nil) = %q, want %q", got, StatusInProgress)
	}
}

func TestTraceStatusComanageTerminal(t *testing.T) {
	setTerminalMarkersForTest(t, "comanage", []string{"ComanageClusterAccountAttached"})
	events := []TraceEventStatus{
		{Source: "comanage", EventType: "ComanageClusterAccountAttached"},
	}
	if got := TraceStatus(events); got != StatusOk {
		t.Errorf("TraceStatus(comanage terminal) = %q, want %q", got, StatusOk)
	}
}
