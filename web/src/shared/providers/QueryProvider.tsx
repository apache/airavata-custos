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

import { ApiError, type KnownError, errorCopy } from "@/shared/api/client";
import { useAuthErrorHandler } from "@/shared/auth/useAuthErrorHandler";
import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { toast } from "sonner";

declare module "@tanstack/react-query" {
  interface Register {
    // inlineError: the caller renders the failure itself, so no toast. errors: the causes master
    // returns for this mutation, worded; anything else toasts errorCopy's generic copy.
    mutationMeta: { inlineError?: boolean; errors?: readonly KnownError[] };
  }
}

export function QueryProvider({ children }: { children: ReactNode }) {
  const handleAuthError = useAuthErrorHandler();
  const [client] = useState(
    () =>
      new QueryClient({
        queryCache: new QueryCache({ onError: handleAuthError }),
        // Every mutation failure toasts here; call sites add only their success toast.
        mutationCache: new MutationCache({
          onError: (error, _variables, _context, mutation) => {
            handleAuthError(error);
            if (!mutation.meta?.inlineError) toast.error(errorCopy(error, mutation.meta?.errors));
          },
        }),
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            gcTime: 5 * 60_000,
            refetchOnWindowFocus: false,
            // A client error (missing record, no privilege, disabled account) fails the same way twice.
            retry: (failures, error) =>
              failures < 1 &&
              !(error instanceof ApiError && error.status >= 400 && error.status < 500),
          },
          mutations: { retry: 0 },
        },
      }),
  );
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
