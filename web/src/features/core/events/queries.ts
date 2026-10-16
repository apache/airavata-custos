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
import { ApiError, type KnownError } from "@/shared/api/client";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// The worker picks up a retried delivery asynchronously, so deliveries re-read every 2 s for 30 s after one.
let retriedAt = 0;
const retryPoll = () => (Date.now() < retriedAt + 30_000 ? 2_000 : false);

// The backend caps one page at 200 rows and has no paging.
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
    refetchInterval: retryPoll,
  });
}

export function useDelivery(id: string | undefined) {
  return useQuery({
    queryKey: eventKeys.delivery(id ?? "none"),
    queryFn: id ? () => getEventsDeliveriesById({ path: { id } }) : skipToken,
    refetchInterval: retryPoll,
  });
}

export function useSubscriptions() {
  return useQuery({
    queryKey: eventKeys.subscriptions(),
    queryFn: () => getEventsSubscriptions(),
  });
}

const STALE: Record<Exclude<EventDeliveryStatus, "FAILED">, string> = {
  PENDING: "This delivery is already waiting for its next try.",
  SUCCEEDED: "This delivery already succeeded.",
};

// A retry re-reads the delivery and refuses one no longer failed; a 409 means another retry or the
// worker settled it between the read and the post.
async function retryFailed(options: Parameters<typeof postEventsDeliveriesByIdRetry<true>>[0]) {
  const current = await getEventsDeliveriesById({ path: options.path });
  if (current.status !== "FAILED") throw new Error(STALE[current.status]);
  return postEventsDeliveriesByIdRetry<true>(options);
}

const RETRY_ERRORS: KnownError[] = [
  { status: 409, copy: "This delivery is no longer failed." },
  { status: 404, copy: "This delivery no longer exists." },
];

export function useRetryDelivery() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: retryFailed,
    meta: { errors: RETRY_ERRORS },
    onError: (error, { path }) => {
      if (error instanceof ApiError && error.status === 404)
        return client.resetQueries({ queryKey: eventKeys.delivery(path.id) });
    },
    // A refused retry polls too: the delivery it re-read may already be waiting for its next try.
    onSettled: () => {
      retriedAt = Date.now();
      return client.invalidateQueries({ queryKey: eventKeys.all });
    },
  });
}
