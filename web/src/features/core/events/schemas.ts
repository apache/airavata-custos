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

import {
  zAuditEvent,
  zEvent,
  zGetEventsDeliveriesByIdResponse,
  zGetEventsDeliveriesResponse,
  zGetEventsSubscriptionsResponse,
  zPendingDelivery,
  zSubscription,
  zUser,
} from "@/generated/core/zod.gen";
import { z } from "zod";

// The trace id column is fixed width, so a missing trace comes back as spaces.
const traceIdSchema = z.string().transform((s) => s.trim());

// The payload is raw JSON on the backend, whatever the spec calls it.
export const busEventSchema = zEvent.extend({
  payload: z.unknown(),
  trace_id: traceIdSchema,
});

export const deliverySchema = zPendingDelivery.extend({ event: busEventSchema });

export const deliveryListSchema = zGetEventsDeliveriesResponse.extend({
  items: z.array(deliverySchema),
});

export const historyEntrySchema = zAuditEvent
  .pick({ id: true, event_type: true, event_time: true, details: true, source: true })
  .extend({ trace_id: traceIdSchema });

export const userNameSchema = zUser.pick({ first_name: true, last_name: true, email: true });

export const deliveryDetailSchema = zGetEventsDeliveriesByIdResponse.extend({
  event: busEventSchema,
  history: z
    .array(historyEntrySchema)
    .nullable()
    .transform((history) => history ?? []),
});

export const subscriptionListSchema = zGetEventsSubscriptionsResponse.extend({
  items: z.array(zSubscription.pick({ subscriber: true, event_type: true, loaded: true })),
});
