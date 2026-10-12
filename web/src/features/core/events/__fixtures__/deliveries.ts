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

import type { AuditEvent, DeliveryHistory, Subscription } from "@/generated/core/types.gen";

function attempt(
  deliveryId: string,
  n: number,
  at: string,
  error?: string,
  subscriber = "comanage-identity-provisioner",
): AuditEvent {
  return {
    id: crypto.randomUUID(),
    event_type: error ? "EVENT_DELIVERY_FAILED" : "EVENT_DELIVERY_SUCCEEDED",
    event_time: at,
    details: JSON.stringify({ subscriber, attempt: n, ...(error ? { error } : {}) }),
    source: subscriber,
    entity_id: deliveryId,
    entity_type: "event_delivery",
    trace_id: "4bf92f3577b34da6a3ce929d0e0e4736",
    span_id: "",
  };
}

const REGISTRY_ERROR = "email search: comanage: 401 unauthorized";

// A delivery that failed twice, was retried by an admin, then failed again.
const failed: DeliveryHistory = {
  id: "07000001-0000-4000-8000-000007000001",
  event_id: "08000015-0000-4000-8000-000008000015",
  subscriber: "comanage-identity-provisioner",
  status: "FAILED",
  attempts: 10,
  next_run_at: "2026-10-02T03:40:00Z",
  last_error: REGISTRY_ERROR,
  created_at: "2026-10-02T03:18:44Z",
  finished_at: "2026-10-02T09:12:00Z",
  event: {
    id: "08000015-0000-4000-8000-000008000015",
    event_type: "compute_cluster_user::approve",
    payload: {
      id: "03000001-0000-4000-8000-000003000001",
      user_id: "24000013-0000-4000-8000-000024000013",
      local_username: "jdoe",
    },
    source: "core",
    trace_id: "4bf92f3577b34da6a3ce929d0e0e4736",
    span_id: "",
    created_at: "2026-10-02T03:18:44Z",
  },
  history: [
    attempt("07000001-0000-4000-8000-000007000001", 1, "2026-10-02T03:18:52Z", REGISTRY_ERROR),
    attempt("07000001-0000-4000-8000-000007000001", 2, "2026-10-02T03:19:32Z", REGISTRY_ERROR),
    {
      id: "02000001-0000-4000-8000-000002000001",
      event_type: "EVENT_DELIVERY_RETRIED",
      event_time: "2026-10-02T03:21:00Z",
      details: JSON.stringify({
        subscriber: "comanage-identity-provisioner",
        actor_id: "24000005-0000-4000-8000-000024000005",
        previous_attempts: 2,
        last_error: REGISTRY_ERROR,
      }),
      source: "comanage-identity-provisioner",
      entity_id: "07000001-0000-4000-8000-000007000001",
      entity_type: "event_delivery",
      trace_id: "5c0a3f3577b34da6a3ce929d0e0e4999",
      span_id: "",
    },
    attempt("07000001-0000-4000-8000-000007000001", 1, "2026-10-02T03:21:10Z", REGISTRY_ERROR),
  ],
};

const slurmAttempt = (n: number, at: string) =>
  attempt(
    "07000003-0000-4000-8000-000007000003",
    n,
    at,
    "slurmrestd 500",
    "slurm-association-mapper",
  );

const retrying: DeliveryHistory = {
  id: "07000003-0000-4000-8000-000007000003",
  event_id: "08000014-0000-4000-8000-000008000014",
  subscriber: "slurm-association-mapper",
  status: "PENDING",
  attempts: 3,
  next_run_at: "2099-01-01T00:00:00Z",
  last_error: "slurmrestd 500: unable to connect to database",
  created_at: "2026-10-02T08:00:00Z",
  event: {
    id: "08000014-0000-4000-8000-000008000014",
    event_type: "compute_allocation::create",
    payload: { id: "01000005-0000-4000-8000-000001000005", name: "E2E-201" },
    source: "amie",
    trace_id: "",
    span_id: "",
    created_at: "2026-10-02T08:00:00Z",
  },
  history: [
    slurmAttempt(1, "2026-10-02T08:00:05Z"),
    slurmAttempt(2, "2026-10-02T08:00:40Z"),
    slurmAttempt(3, "2026-10-02T08:01:50Z"),
  ],
};

const succeeded: DeliveryHistory = {
  id: "07000002-0000-4000-8000-000007000002",
  event_id: "08000016-0000-4000-8000-000008000016",
  subscriber: "comanage-identity-provisioner",
  status: "SUCCEEDED",
  attempts: 1,
  next_run_at: "2026-10-02T07:00:00Z",
  created_at: "2026-10-02T07:00:00Z",
  finished_at: "2026-10-02T07:00:05Z",
  event: {
    id: "08000016-0000-4000-8000-000008000016",
    event_type: "user_identity::create",
    payload: { user_id: "24000013-0000-4000-8000-000024000013", source: "oidc" },
    source: "core",
    trace_id: "",
    span_id: "",
    created_at: "2026-10-02T07:00:00Z",
  },
  history: [attempt("07000002-0000-4000-8000-000007000002", 1, "2026-10-02T07:00:05Z")],
};

// Its connector is not loaded, so it has never been tried.
const notRunning: DeliveryHistory = {
  id: "07000004-0000-4000-8000-000007000004",
  event_id: "08000014-0000-4000-8000-000008000014",
  subscriber: "signer-principal-sync",
  status: "PENDING",
  attempts: 0,
  next_run_at: "2026-10-02T08:00:00Z",
  created_at: "2026-10-02T08:00:00Z",
  event: retrying.event,
  history: [],
};

export const deliveryFixtures: DeliveryHistory[] = [retrying, notRunning, succeeded, failed];

const subscription = (subscriber: string, event_type: string, loaded: boolean): Subscription => ({
  subscriber,
  event_type,
  loaded,
  created_at: "2026-10-01T00:00:00Z",
});

export const subscriptionFixtures: Subscription[] = [
  subscription("comanage-identity-provisioner", "compute_cluster_user::approve", true),
  subscription("slurm-association-mapper", "compute_allocation::create", true),
  subscription("signer-principal-sync", "compute_allocation::create", false),
];
