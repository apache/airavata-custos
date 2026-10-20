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
import { deleteAllocation } from "@/features/core/allocations/api";
import { allocationKeys } from "@/features/core/allocations/queries";
import { clusterAccountKeys } from "@/features/core/cluster-accounts/queries";
import type { DeletionRequest } from "@/generated/core/zod.gen";
import { type DeletionReview, listDeletionRequests, reviewDeletion } from "./api";

export const deletionRequestKeys = { all: ["deletion-requests"] as const };

export function useDeletionRequests(enabled = true) {
  return useQuery({
    queryKey: deletionRequestKeys.all,
    queryFn: listDeletionRequests,
    enabled,
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}

export function useRequestAllocationDeletion() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteAllocation,
    onSuccess: () => client.invalidateQueries({ queryKey: deletionRequestKeys.all }),
  });
}

export function useReviewDeletion() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ row, action }: { row: DeletionRequest; action: DeletionReview }) =>
      reviewDeletion(row, action),
    onSuccess: () =>
      Promise.all(
        [deletionRequestKeys.all, allocationKeys.all, clusterAccountKeys.all].map((queryKey) =>
          client.invalidateQueries({ queryKey }),
        ),
      ),
  });
}
