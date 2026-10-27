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

import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  type AcceptanceParams,
  acceptAgreement,
  checkUsername,
  completeOnboarding,
  createAgreement,
  deleteAgreement,
  getAgreement,
  getGateProfile,
  getSetupSteps,
  listAcceptances,
  listAgreements,
  publishAgreement,
  removeAgreement,
  updateAgreement,
} from "./api";

export const agreementKeys = {
  all: ["agreements"] as const,
  list: () => [...agreementKeys.all, "list"] as const,
  detail: (id: string) => [...agreementKeys.all, "detail", id] as const,
  acceptances: (id: string, params: AcceptanceParams) =>
    [...agreementKeys.detail(id), "acceptances", params] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useAgreements() {
  return useQuery({ queryKey: agreementKeys.list(), queryFn: listAgreements, ...DEFAULTS });
}

export function useAgreement(id: string) {
  return useQuery({
    queryKey: agreementKeys.detail(id),
    queryFn: () => getAgreement(id),
    ...DEFAULTS,
  });
}

export function useAgreementAcceptances(id: string, params: AcceptanceParams) {
  return useQuery({
    queryKey: agreementKeys.acceptances(id, params),
    queryFn: () => listAcceptances(id, params),
    placeholderData: keepPreviousData,
    ...DEFAULTS,
  });
}

export function useCreateAgreement() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: createAgreement,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.all }),
  });
}

export function useUpdateAgreement() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: updateAgreement,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.all }),
  });
}

export function usePublishAgreement() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: publishAgreement,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.all }),
  });
}

export function useRemoveAgreement() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: removeAgreement,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.all }),
  });
}

// Polls every 3s until every step is done.
export function useSetupSteps() {
  return useQuery({
    queryKey: [...agreementKeys.all, "setup"],
    queryFn: getSetupSteps,
    refetchInterval: (query) => (query.state.data?.every((s) => s.done_at) ? false : 3000),
  });
}

export function useGateProfile() {
  return useQuery({ queryKey: [...agreementKeys.all, "me"], queryFn: getGateProfile, ...DEFAULTS });
}

export function useUsernameCheck(username: string, enabled: boolean) {
  return useQuery({
    queryKey: [...agreementKeys.all, "username", username],
    queryFn: () => checkUsername(username),
    enabled: enabled && username.length > 0,
    ...DEFAULTS,
  });
}

export function useCompleteOnboarding() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: completeOnboarding,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.all }),
  });
}

export function useAcceptAgreement() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: acceptAgreement,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.all }),
  });
}

// Only the list is refreshed; refetching the deleted detail would 404 before the page leaves.
export function useDeleteAgreement() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteAgreement,
    onSuccess: () => client.invalidateQueries({ queryKey: agreementKeys.list() }),
  });
}
