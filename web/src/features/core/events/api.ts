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

import { apiFetch } from "@/shared/api/client";
import {
  deliveryDetailSchema,
  deliveryListSchema,
  subscriptionListSchema,
  userNameSchema,
} from "./schemas";
import type { Delivery, DeliveryDetail, DeliveryStatus, Subscription } from "./types";

// The backend caps one page at 200 rows and has no paging yet.
export const DELIVERY_PAGE_SIZE = 200;

export async function listDeliveries(status?: DeliveryStatus): Promise<Delivery[]> {
  const params = new URLSearchParams({ limit: String(DELIVERY_PAGE_SIZE) });
  if (status) params.set("status", status);
  const raw = await apiFetch(`/events/deliveries?${params.toString()}`);
  return deliveryListSchema.parse(raw).items;
}

export async function getDelivery(id: string): Promise<DeliveryDetail> {
  const raw = await apiFetch(`/events/deliveries/${encodeURIComponent(id)}`);
  return deliveryDetailSchema.parse(raw);
}

export async function retryDelivery(id: string): Promise<void> {
  await apiFetch(`/events/deliveries/${encodeURIComponent(id)}/retry`, { method: "POST" });
}

// The name of the admin who retried a delivery. Falls back to the id when the
// caller cannot read users.
export async function getUserName(id: string): Promise<string> {
  try {
    const user = userNameSchema.parse(await apiFetch(`/users/${encodeURIComponent(id)}`));
    const name = [user.first_name, user.last_name].filter(Boolean).join(" ");
    return name || user.email || id;
  } catch {
    return id;
  }
}

export async function listSubscriptions(): Promise<Subscription[]> {
  const raw = await apiFetch("/events/subscriptions");
  return subscriptionListSchema.parse(raw).items;
}
