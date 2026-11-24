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
import {
  createServiceAccount,
  deleteServiceAccount,
  disableServiceAccount,
  enableServiceAccount,
  issueServiceAccountSecret,
  listServiceAccountPrivileges,
  listServiceAccounts,
  updateServiceAccount,
} from "./api";

const serviceAccountKeys = {
  list: ["service-accounts", "list"] as const,
  privileges: ["service-accounts", "privileges"] as const,
};

export function useServiceAccounts() {
  return useQuery({ queryKey: serviceAccountKeys.list, queryFn: listServiceAccounts });
}

export function useServiceAccountPrivileges() {
  return useQuery({
    queryKey: serviceAccountKeys.privileges,
    queryFn: listServiceAccountPrivileges,
  });
}

function useServiceAccountMutation<Input, Output>(
  mutationFn: (input: Input) => Promise<Output>,
  queryKey: readonly string[] = serviceAccountKeys.list,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey }),
  });
}

export const useCreateServiceAccount = () => useServiceAccountMutation(createServiceAccount);
export const useUpdateServiceAccount = () => useServiceAccountMutation(updateServiceAccount);
export const useDisableServiceAccount = () => useServiceAccountMutation(disableServiceAccount);
export const useEnableServiceAccount = () => useServiceAccountMutation(enableServiceAccount);
export const useDeleteServiceAccount = () => useServiceAccountMutation(deleteServiceAccount);
export const useIssueServiceAccountSecret = () =>
  useServiceAccountMutation(issueServiceAccountSecret);
