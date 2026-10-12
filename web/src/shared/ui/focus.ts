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

// Where focus goes when the control that raised a dialog, drawer or confirmation is gone. Every
// press is recorded before focus moves (capture phase): its row's key and opener. Focus later
// returns to that opener while it is connected, else to the opener of the row with that key, else
// to the list region it sat in, never to <body>. A drawer opened with no recent press (a deep
// link) returns to the page's list region.

"use client";

import { useRouter } from "next/navigation";
import * as React from "react";

const CONTROL = "button, a[href], input, select, textarea";
const OPENER = "[data-row-opener]";
const OPEN_CONTAINER = '[role="dialog"], [role="alertdialog"]';
const FOCUS_REGION = '[role="region"], section[aria-label], section[aria-labelledby]';
const LOST = `#main-content, [role="tabpanel"], [data-slot="drawer-body"], ${OPEN_CONTAINER}, [data-base-ui-focus-guard]`;
// A press older than this raised nothing the opening drawer or dialog shows.
const PRESS_MS = 1500;

// `kind`: the opener's data-row-opener value, for a row with more than one (a name and its Rates).
export type FocusOrigin = {
  opener: HTMLElement | null;
  rowKey?: string;
  kind?: string;
  region: HTMLElement | null;
};

type Press = { origin: FocusOrigin; inside: boolean; at: number };

let press: Press | null = null;
let acted = false;

// Focus with no meaningful holder: <body>, the page's main, a tab panel, an open dialog or drawer's
// popup or its own scroll region (where Base UI parks focus whose holder unmounted), or a focus guard.
export function focusIsLost(): boolean {
  const active = document.activeElement;
  return !active || active === document.body || active.matches(LOST);
}

function originOf(el: Element | null): FocusOrigin {
  const row = el?.closest<HTMLElement>("[data-row-key]");
  const opener =
    el?.closest<HTMLElement>(`${OPENER}, ${CONTROL}`) ??
    row?.querySelector<HTMLElement>(OPENER) ??
    row?.querySelector<HTMLElement>(CONTROL) ??
    null;
  return {
    opener,
    rowKey: row?.dataset.rowKey,
    kind: opener?.dataset.rowOpener,
    region: (opener ?? el)?.closest<HTMLElement>(FOCUS_REGION) ?? null,
  };
}

function record(el: Element | null) {
  acted = true;
  press = el && {
    origin: originOf(el),
    inside: Boolean(el.closest(OPEN_CONTAINER)),
    at: Date.now(),
  };
}

if (typeof document !== "undefined") {
  document.addEventListener("pointerdown", (e) => record(e.target as Element), true);
  document.addEventListener(
    "keydown",
    () => record(focusIsLost() ? null : document.activeElement),
    true,
  );
}

const recentPress = () => (press && Date.now() - press.at < PRESS_MS ? press : null);

// True once the user pressed or typed in this document; a fresh load claims no focus.
export const userHasActed = () => acted;

// The origin of what is being raised now: the latest press when it named a control or row (a
// menu item names neither), else the focused control.
export function captureOrigin(): FocusOrigin {
  const fresh = recentPress()?.origin;
  if (fresh?.opener || fresh?.rowKey) return fresh;
  return originOf(focusIsLost() ? null : document.activeElement);
}

const focusable = (el: HTMLElement) => {
  if (!el.hasAttribute("tabindex") && el.tabIndex < 0) el.tabIndex = -1;
  return el;
};

function live(origin: FocusOrigin | null | undefined): HTMLElement | null {
  if (!origin) return null;
  if (origin.opener?.isConnected) return origin.opener;
  const region = origin.region?.isConnected ? origin.region : null;
  const row = origin.rowKey
    ? (region ?? document).querySelector<HTMLElement>(
        `[data-row-key="${CSS.escape(origin.rowKey)}"]`,
      )
    : null;
  const kind =
    origin.kind === undefined ? OPENER : `[data-row-opener="${CSS.escape(origin.kind)}"]`;
  return (
    row?.querySelector<HTMLElement>(kind) ??
    row?.querySelector<HTMLElement>(OPENER) ??
    row?.querySelector<HTMLElement>(CONTROL) ??
    (region && focusable(region))
  );
}

// The raiser while connected, else its row's opener, else its region, else the page's h1.
export function originTarget(origin: FocusOrigin | null | undefined): HTMLElement | null {
  const heading = document.querySelector<HTMLElement>("#main-content h1");
  return live(origin) ?? (heading && focusable(heading));
}

// The page's labelled list region, else its h1; none while the page still shows a loading state.
export function landingTarget(): HTMLElement | null {
  const main = document.getElementById("main-content");
  if (main?.querySelector('output[aria-label="Loading"]')) return null;
  const el =
    main?.querySelector<HTMLElement>("[data-table-region]") ??
    main?.querySelector<HTMLElement>("h1");
  return el ? focusable(el) : null;
}

// Spread onto a scroll region: it is a Tab stop only while its content overflows it.
export function useScrollStop<T extends HTMLElement>(): {
  ref: (el: T | null) => void;
  tabIndex: 0 | undefined;
} {
  const [el, setEl] = React.useState<T | null>(null);
  const [overflows, setOverflows] = React.useState(false);
  React.useEffect(() => {
    if (!el) return;
    const check = () =>
      setOverflows(el.scrollHeight > el.clientHeight || el.scrollWidth > el.clientWidth);
    check();
    if (typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(check);
    for (const box of [el, ...el.children]) observer.observe(box);
    return () => observer.disconnect();
  }, [el]);
  return { ref: setEl, tabIndex: overflows ? 0 : undefined };
}

const TABBABLE =
  'a[href], button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled), summary, [contenteditable]:not([contenteditable="false"]), [tabindex]:not([tabindex="-1"])';

// A non-modal drawer's onKeyDown: Tab wraps within it, so focus never falls past it to <body>.
// Base UI's trap-focus would do the same but hides the page from assistive tech.
export function wrapTab(e: React.KeyboardEvent<HTMLElement>) {
  const box = e.currentTarget;
  const active = document.activeElement;
  if (e.key !== "Tab" || !active || !box.contains(active)) return;
  const items = [...box.querySelectorAll<HTMLElement>(TABBABLE)].filter(
    (el) => !el.hasAttribute("data-base-ui-focus-guard") && el.checkVisibility(),
  );
  const first = items[0];
  const last = items.at(-1);
  if (!first || !last) return;
  if (e.shiftKey ? active === first || active === box : active === last) {
    e.preventDefault();
    (e.shiftKey ? last : first).focus();
  }
}

// A dialog or drawer's finalFocus, focusing the origin of the press that opened it, replaced by a press
// outside it that swaps the record it shows (`key`); after such a swap focus moves into the
// drawer, onto its Close. With no origin (a deep link) focus goes to the list region, else the h1.
export function useFocusOrigin(open: boolean, key?: string): () => false {
  const origin = React.useRef<FocusOrigin | null>(null);
  const shown = React.useRef<{ open: boolean; key?: string }>({ open: false });
  const swapped = React.useRef(false);
  if (
    open &&
    (!shown.current.open || key !== shown.current.key) &&
    typeof document !== "undefined"
  ) {
    const fresh = recentPress();
    if (!shown.current.open) origin.current = captureOrigin();
    else if (fresh && !fresh.inside) {
      origin.current = fresh.origin;
      swapped.current = true;
    }
  }
  shown.current = { open, key };
  // biome-ignore lint/correctness/useExhaustiveDependencies: runs after the commit of each swap
  React.useEffect(() => {
    if (!swapped.current) return;
    swapped.current = false;
    keepFocus(
      () =>
        [...document.querySelectorAll(OPEN_CONTAINER)]
          .at(-1)
          ?.querySelector<HTMLElement>('button[aria-label="Close"]') ?? null,
      { ms: 2000, lost: () => !document.activeElement?.closest(OPEN_CONTAINER) },
    );
  }, [key]);
  // Focus is placed here, not by Base UI: it skips its return once focus has left a non-modal
  // drawer (a swap press), and it would pass a region on to the region's first tabbable child.
  return React.useCallback(() => {
    const target = live(origin.current) ?? landingTarget();
    if (!target) keepFocus(landingTarget);
    queueMicrotask(() => {
      if (!shown.current.open && target?.isConnected && focusIsLost()) target.focus();
    });
    return false;
  }, []);
}

// For a while after an action, focus that is lost (its holder removed, a page swapped in) moves
// to target(). It ends on the user's next press or key, after ms, or once the
// address leaves `routes`, and it never pulls focus out of an open dialog or drawer.
export function keepFocus(
  target: () => HTMLElement | null,
  {
    ms = 10_000,
    routes = [window.location.pathname],
    lost = focusIsLost,
  }: { ms?: number; routes?: string[]; lost?: () => boolean } = {},
) {
  const place = () => {
    if (!routes.includes(window.location.pathname)) return stop();
    if (!lost()) return;
    const el = target();
    const open = [...document.querySelectorAll(OPEN_CONTAINER)];
    if (el && el !== document.activeElement && (!open.length || open.some((c) => c.contains(el))))
      focusable(el).focus();
  };
  const observer = new MutationObserver(place);
  const stop = () => {
    observer.disconnect();
    clearTimeout(timer);
    document.removeEventListener("pointerdown", stop, true);
    document.removeEventListener("keydown", stop, true);
  };
  const timer = setTimeout(stop, ms);
  observer.observe(document.body, { childList: true, subtree: true });
  document.addEventListener("pointerdown", stop, true);
  document.addEventListener("keydown", stop, true);
  place();
}

// Routes that only redirect, by the route they land on.
const REDIRECTS: Record<string, string> = { "/": "/allocations" };

function focusOnArrival(href: string, target: () => HTMLElement | null) {
  const asked = new URL(href, window.location.href).pathname;
  const path = REDIRECTS[asked] ?? asked;
  keepFocus(() => (window.location.pathname === path ? target() : null), {
    routes: [window.location.pathname, asked, path],
  });
}

// Focuses the h1 of the page at href once it renders, for a navigation about to start.
export function focusHeadingOf(href: string) {
  focusOnArrival(href, () => document.querySelector<HTMLElement>("#main-content h1"));
}

// Focuses the list region (else h1) of the page at href once it renders: a drawer with no live
// origin closing by navigation.
export function focusLandingOf(href: string) {
  focusOnArrival(href, landingTarget);
}

// Navigates away (after a delete) and focuses the destination page's h1.
export function useLeaveTo(): (href: string) => void {
  const router = useRouter();
  return React.useCallback(
    (href: string) => {
      focusHeadingOf(href);
      router.push(href);
    },
    [router],
  );
}
