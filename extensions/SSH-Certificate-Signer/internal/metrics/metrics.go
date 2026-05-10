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

// Package metrics defines Prometheus metric collectors.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SignRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signer_sign_requests_total",
		Help: "Total number of signing requests",
	}, []string{"tenant_id", "status"})

	SignDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "signer_sign_duration_seconds",
		Help:    "Duration of signing requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"tenant_id"})

	RevokeRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signer_revoke_requests_total",
		Help: "Total number of revocation requests",
	}, []string{"tenant_id", "status"})

	AuthFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signer_auth_failures_total",
		Help: "Total number of authentication failures",
	}, []string{"reason"})

	VaultOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signer_vault_operations_total",
		Help: "Total number of Vault operations",
	}, []string{"operation", "status"})

	DBQueryDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "signer_db_query_duration_seconds",
		Help:    "Duration of database queries in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"query"})
)
