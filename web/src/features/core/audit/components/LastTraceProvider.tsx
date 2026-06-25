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

import * as React from "react";
import { getLastTraceId, subscribeLastTraceId } from "@/shared/api/last-trace-id";

// Surfaces the latest X-Trace-Id captured by apiFetch via React context so
// toast deep-links and global "View this trace" affordances can read the
// value without coupling to the singleton.
const LastTraceContext = React.createContext<string | null | undefined>(undefined);

export function LastTraceProvider({ children }: { children: React.ReactNode }) {
  const [traceId, setTraceId] = React.useState<string | null>(() => getLastTraceId());

  React.useEffect(() => {
    setTraceId(getLastTraceId());
    return subscribeLastTraceId(setTraceId);
  }, []);

  return <LastTraceContext.Provider value={traceId}>{children}</LastTraceContext.Provider>;
}

// Returns `undefined` outside the provider; queries.useLastTraceId falls back
// to the singleton in that case to preserve the public API.
export function useLastTraceContext(): string | null | undefined {
  return React.useContext(LastTraceContext);
}
