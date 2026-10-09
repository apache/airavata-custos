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

"use client";

import {
  getEventsDeliveries,
  getEventsDeliveriesById,
  getEventsSubscriptions,
  postEventsDeliveriesByIdRetry,
} from "@/generated/core/sdk.gen";
import type { EventDeliveryStatus } from "@/generated/core/types.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useQuery } from "@tanstack/react-query";

// The backend caps one page at 200 rows and has no paging yet.
export const DELIVERY_PAGE_SIZE = 200;

export const eventKeys = {
  all: ["events"] as const,
  deliveries: (status?: EventDeliveryStatus) =>
    [...eventKeys.all, "deliveries", status ?? "ALL"] as const,
  delivery: (id: string) => [...eventKeys.all, "delivery", id] as const,
  subscriptions: () => [...eventKeys.all, "subscriptions"] as const,
};

export function useDeliveries(status?: EventDeliveryStatus) {
  return useQuery({
    queryKey: eventKeys.deliveries(status),
    queryFn: () => getEventsDeliveries({ query: { status, limit: DELIVERY_PAGE_SIZE } }),
  });
}

export function useDelivery(id: string | undefined) {
  return useQuery({
    queryKey: eventKeys.delivery(id ?? "none"),
    queryFn: id ? () => getEventsDeliveriesById({ path: { id } }) : skipToken,
  });
}

export function useSubscriptions() {
  return useQuery({
    queryKey: eventKeys.subscriptions(),
    queryFn: () => getEventsSubscriptions(),
  });
}

export function useRetryDelivery() {
  return useInvalidating(postEventsDeliveriesByIdRetry<true>, eventKeys.all);
}
