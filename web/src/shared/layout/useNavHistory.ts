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

import { useSyncExternalStore } from "react";

// Each history entry carries its index (custosIdx) in history.state, stamped by wrapping pushState
// and replaceState (Next's replaceState drops custom state); the highest index still ahead of the
// current entry lives in sessionStorage, so the chevrons stay right across reloads and documents.
// An entry entered from outside the portal is index 0: Back from it would leave the portal. An
// unstamped entry reached by Back (a page the portal replaced) sits just behind the one it left.
const TOP = "custos.nav.top";
const LAST = "custos.nav.last";
const listeners = new Set<() => void>();
let top = 0;
let installed = false;

function index(): number | undefined {
  const state: unknown = window.history.state;
  return typeof state === "object" && state !== null && "custosIdx" in state
    ? Number(state.custosIdx)
    : undefined;
}

function changed() {
  sessionStorage.setItem(TOP, String(top));
  sessionStorage.setItem(LAST, String(index() ?? 0));
  for (const listener of listeners) listener();
}

function install() {
  installed = true;
  const push = window.history.pushState.bind(window.history);
  const replace = window.history.replaceState.bind(window.history);
  top = Number(sessionStorage.getItem(TOP) ?? 0);
  if (index() === undefined) {
    const last = sessionStorage.getItem(LAST);
    const fromPortal =
      last !== null &&
      document.referrer !== "" &&
      new URL(document.referrer).origin === window.location.origin;
    const nav = performance.getEntriesByType("navigation")[0];
    const returned = nav instanceof PerformanceNavigationTiming && nav.type === "back_forward";
    const idx = returned ? Math.max(0, Number(last ?? 0) - 1) : fromPortal ? Number(last) + 1 : 0;
    if (!returned) top = idx;
    replace(Object.assign({}, window.history.state, { custosIdx: idx }), "");
  }
  window.history.pushState = (data: unknown, unused: string, url?: string | URL | null) => {
    top = (index() ?? 0) + 1;
    push(Object.assign({}, data, { custosIdx: top }), unused, url);
    changed();
  };
  window.history.replaceState = (data: unknown, unused: string, url?: string | URL | null) => {
    replace(Object.assign({}, data, { custosIdx: index() ?? 0 }), unused, url);
  };
  window.addEventListener("popstate", changed);
  changed();
}

function subscribe(listener: () => void) {
  if (!installed) install();
  listeners.add(listener);
  return () => void listeners.delete(listener);
}

const canGoBack = () => (index() ?? 0) > 0;
const canGoForward = () => (index() ?? 0) < top;
const never = () => false;

export function useNavHistory(): { canGoBack: boolean; canGoForward: boolean } {
  return {
    canGoBack: useSyncExternalStore(subscribe, canGoBack, never),
    canGoForward: useSyncExternalStore(subscribe, canGoForward, never),
  };
}
