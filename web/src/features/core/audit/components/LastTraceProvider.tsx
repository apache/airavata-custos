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
