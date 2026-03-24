// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistered(t *testing.T) {
	// Just verify the metrics exist and can be collected
	SignRequestsTotal.WithLabelValues("test", "success").Inc()
	SignDurationSeconds.WithLabelValues("test").Observe(0.1)
	RevokeRequestsTotal.WithLabelValues("test", "success").Inc()
	AuthFailuresTotal.WithLabelValues("unauthorized").Inc()
	VaultOperationsTotal.WithLabelValues("get_ca_key", "success").Inc()
	DBQueryDurationSeconds.WithLabelValues("select").Observe(0.01)

	// Verify they can be gathered without panicking
	gathering, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	if len(gathering) == 0 {
		t.Error("expected some metric families")
	}

	// Check that our custom metrics exist
	found := make(map[string]bool)
	for _, mf := range gathering {
		found[mf.GetName()] = true
	}

	expectedMetrics := []string{
		"signer_sign_requests_total",
		"signer_sign_duration_seconds",
		"signer_revoke_requests_total",
		"signer_auth_failures_total",
		"signer_vault_operations_total",
		"signer_db_query_duration_seconds",
	}

	for _, name := range expectedMetrics {
		if !found[name] {
			t.Errorf("metric %s not found in gathered metrics", name)
		}
	}
}
