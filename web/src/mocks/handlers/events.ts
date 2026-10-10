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

import { http, HttpResponse } from "msw";
import {
  deliveryFixtures,
  subscriptionFixtures,
} from "@/features/core/events/__fixtures__/deliveries";
import type { DeliveryHistory } from "@/generated/core/types.gen";

// A copy, so a retry in one browser session does not change the fixtures.
const deliveries: DeliveryHistory[] = structuredClone(deliveryFixtures);

export const eventsHandlers = [
  http.get("/api/v1/events/deliveries", ({ request }) => {
    const status = new URL(request.url).searchParams.get("status");
    const items = deliveries
      .filter((d) => !status || d.status === status)
      .map(({ history: _history, ...row }) => row);
    return HttpResponse.json({ items });
  }),

  http.get("/api/v1/events/deliveries/:id", ({ params }) => {
    const delivery = deliveries.find((d) => d.id === params.id);
    if (!delivery) return HttpResponse.json({ error: "event delivery not found" }, { status: 404 });
    return HttpResponse.json(delivery);
  }),

  http.post("/api/v1/events/deliveries/:id/retry", ({ params }) => {
    const delivery = deliveries.find((d) => d.id === params.id);
    if (!delivery) return HttpResponse.json({ error: "event delivery not found" }, { status: 404 });
    if (delivery.status !== "FAILED") {
      return HttpResponse.json({ error: "event delivery has not failed" }, { status: 409 });
    }
    const now = new Date().toISOString();
    delivery.history ??= [];
    delivery.history.push({
      id: `audit-${delivery.id}-retry-${now}`,
      event_type: "EVENT_DELIVERY_RETRIED",
      event_time: now,
      details: JSON.stringify({
        subscriber: delivery.subscriber,
        actor_id: "user-admin",
        previous_attempts: delivery.attempts,
        last_error: delivery.last_error,
      }),
      source: delivery.subscriber,
      entity_id: delivery.id,
      entity_type: "event_delivery",
      trace_id: "",
      span_id: "",
    });
    delivery.status = "PENDING";
    delivery.attempts = 0;
    delivery.next_run_at = now;
    delivery.finished_at = undefined;
    return new HttpResponse(null, { status: 204 });
  }),

  http.get("/api/v1/events/subscriptions", () =>
    HttpResponse.json({ items: subscriptionFixtures }),
  ),
];
