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

import { useSearchParams as useNextSearchParams } from "next/navigation";
import * as React from "react";

// Custom event name used to notify all useShallowSearchParams subscribers
// in the same window when one of them writes via replaceShallowSearchParams.
const SHALLOW_EVENT = "shallow-search-params-change";

let hydrated = false;

function readCurrent(): URLSearchParams {
  if (typeof window === "undefined") return new URLSearchParams();
  return new URLSearchParams(window.location.search);
}

/**
 * Drop-in for `useSearchParams()` that mirrors URL state into React via
 * `window.history.replaceState` instead of `router.replace`. Avoids Next App
 * Router's RSC roundtrip on every URL change. Until the page has hydrated the
 * initial value is seeded from Next's hook so SSR + hydration agree; later
 * mounts read the address bar, which Next's hook lags behind after a shallow
 * write. Post-mount updates come from the shallow setter via a custom event,
 * and from Next's hook when a same-route <Link> navigates.
 */
export function useShallowSearchParams(): URLSearchParams {
  const nextParams = useNextSearchParams();
  const [params, setParams] = React.useState<URLSearchParams>(() =>
    hydrated ? readCurrent() : new URLSearchParams(nextParams?.toString() ?? ""),
  );
  const nextQuery = nextParams?.toString() ?? "";
  React.useEffect(() => setParams(new URLSearchParams(nextQuery)), [nextQuery]);
  React.useEffect(() => {
    hydrated = true;
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

// Sets or, for an empty value, removes one key and writes the URL shallowly.
export function setSearchParam(params: URLSearchParams, key: string, value?: string | null): void {
  const next = new URLSearchParams(params);
  if (value) next.set(key, value);
  else next.delete(key);
  replaceShallowSearchParams(next);
}

// Rewrites the keys a list owns to the filters it applied (replace, no history entry), so an ignored or
// unnormalised value never lingers in the address bar. Keys left out of applied are not touched.
export function useCanonicalParams(applied: Record<string, string | string[] | undefined>): void {
  const target = JSON.stringify(applied, (_key, value) => value ?? null);
  const seen = useShallowSearchParams().toString();
  React.useEffect(() => {
    const current = readCurrent();
    // applied was parsed from a query this render had not caught up with yet; the sync re-runs this.
    if (seen !== current.toString()) return;
    const wanted = JSON.parse(target) as Record<string, string | string[] | null>;
    const values = (key: string) => [wanted[key] ?? []].flat().filter(Boolean);
    const keys = Object.keys(wanted);
    // Compared as lists, so an empty ?q= differs from no q and is removed.
    const same = (k: string) => JSON.stringify(current.getAll(k)) === JSON.stringify(values(k));
    if (keys.every(same)) return;
    const next = new URLSearchParams(current);
    for (const key of keys) {
      next.delete(key);
      for (const v of values(key)) next.append(key, v);
    }
    replaceShallowSearchParams(next);
  }, [target, seen]);
}

// A page's declared query keys: every other key, and every empty value, leaves the address bar.
export function useDeclaredParams(keys: readonly string[]): void {
  const declared = keys.join(",");
  const seen = useShallowSearchParams().toString();
  // biome-ignore lint/correctness/useExhaustiveDependencies: re-checks on every query change
  React.useEffect(() => {
    const allowed = new Set(declared.split(","));
    const current = readCurrent();
    const next = new URLSearchParams([...current].filter(([k, v]) => allowed.has(k) && v));
    if (next.toString() !== current.toString()) replaceShallowSearchParams(next);
  }, [declared, seen]);
}
