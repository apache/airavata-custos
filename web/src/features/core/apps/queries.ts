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
  authorizeApp,
  createApp,
  disableApp,
  enableApp,
  issueAppSecret,
  listAppAuthorizations,
  listAppConnectors,
  listApps,
  removeAppAuthorization,
  updateApp,
} from "./api";

const appKeys = {
  list: ["apps", "list"] as const,
  connectors: ["apps", "connectors"] as const,
  authorizations: ["apps", "authorizations"] as const,
};

export function useAppAuthorizations() {
  return useQuery({ queryKey: appKeys.authorizations, queryFn: listAppAuthorizations });
}

export const useAuthorizeApp = () => useAppMutation(authorizeApp, appKeys.authorizations);
export const useRemoveAppAuthorization = () =>
  useAppMutation(removeAppAuthorization, appKeys.authorizations);

export function useApps() {
  return useQuery({ queryKey: appKeys.list, queryFn: listApps });
}

export function useAppConnectors() {
  return useQuery({ queryKey: appKeys.connectors, queryFn: listAppConnectors });
}

function useAppMutation<Input, Output>(
  mutationFn: (input: Input) => Promise<Output>,
  queryKey: readonly string[] = appKeys.list,
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey }),
  });
}

export const useCreateApp = () => useAppMutation(createApp);
export const useUpdateApp = () => useAppMutation(updateApp);
export const useDisableApp = () => useAppMutation(disableApp);
export const useEnableApp = () => useAppMutation(enableApp);
export const useIssueAppSecret = () => useAppMutation(issueAppSecret);
