"use client";

import { useSearchParams as useNextSearchParams } from "next/navigation";
import * as React from "react";

// Custom event name used to notify all useShallowSearchParams subscribers
// in the same window when one of them writes via replaceShallowSearchParams.
const SHALLOW_EVENT = "shallow-search-params-change";

function readCurrent(): URLSearchParams {
  if (typeof window === "undefined") return new URLSearchParams();
  return new URLSearchParams(window.location.search);
}

/**
 * Drop-in for `useSearchParams()` that mirrors URL state into React via
 * `window.history.replaceState` instead of `router.replace`. Avoids Next App
 * Router's RSC roundtrip on every URL change. The initial value is seeded from
 * Next's hook so SSR + hydration agree; post-mount updates come from the
 * shallow setter via a custom event.
 */
export function useShallowSearchParams(): URLSearchParams {
  const nextParams = useNextSearchParams();
  const [params, setParams] = React.useState<URLSearchParams>(
    () => new URLSearchParams(nextParams?.toString() ?? ""),
  );
  React.useEffect(() => {
    const sync = () => setParams(readCurrent());
    sync();
    window.addEventListener(SHALLOW_EVENT, sync);
    window.addEventListener("popstate", sync);
    return () => {
      window.removeEventListener(SHALLOW_EVENT, sync);
      window.removeEventListener("popstate", sync);
    };
  }, []);
  return params;
}

/**
 * Replace the current URL's search params without triggering a Next route
 * refresh. Subscribers of useShallowSearchParams re-render immediately.
 */
export function replaceShallowSearchParams(next: URLSearchParams): void {
  if (typeof window === "undefined") return;
  const query = next.toString();
  const url = query ? `${window.location.pathname}?${query}` : window.location.pathname;
  window.history.replaceState(window.history.state, "", url);
  window.dispatchEvent(new Event(SHALLOW_EVENT));
}
