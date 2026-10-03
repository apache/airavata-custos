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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getDelivery, getUserName, listDeliveries, listSubscriptions, retryDelivery } from "./api";
import type { DeliveryStatus } from "./types";

export const eventKeys = {
  all: ["events"] as const,
  deliveries: (status?: DeliveryStatus) =>
    [...eventKeys.all, "deliveries", status ?? "ALL"] as const,
  delivery: (id: string) => [...eventKeys.all, "delivery", id] as const,
  subscriptions: () => [...eventKeys.all, "subscriptions"] as const,
  userName: (id: string) => [...eventKeys.all, "user-name", id] as const,
};

export function useUserName(id: string) {
  return useQuery({
    queryKey: eventKeys.userName(id),
    queryFn: () => getUserName(id),
    enabled: Boolean(id),
    staleTime: Number.POSITIVE_INFINITY,
  });
}

export function useDeliveries(status?: DeliveryStatus) {
  return useQuery({
    queryKey: eventKeys.deliveries(status),
    queryFn: () => listDeliveries(status),
    refetchOnWindowFocus: false,
  });
}

export function useDelivery(id: string | undefined) {
  return useQuery({
    queryKey: eventKeys.delivery(id ?? "none"),
    queryFn: () => getDelivery(id as string),
    enabled: Boolean(id),
    refetchOnWindowFocus: false,
  });
}

export function useSubscriptions() {
  return useQuery({
    queryKey: eventKeys.subscriptions(),
    queryFn: listSubscriptions,
    refetchOnWindowFocus: false,
  });
}

export function useRetryDelivery() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: retryDelivery,
    onSuccess: () => client.invalidateQueries({ queryKey: eventKeys.all }),
  });
}
