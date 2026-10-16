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

import { errorCopy } from "@/shared/api/client";
import { useShallowSearchParams } from "@/shared/hooks/useShallowSearchParams";
import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import { usePathname } from "next/navigation";
import { Suspense, useEffect, useSyncExternalStore } from "react";
import { Toaster as Sonner, type ToasterProps, toast } from "sonner";
import { CONFIRM_TOASTER, confirmToast, dismissConfirm } from "./confirm";

// Typed toasts take the themed tone pair; the tint is layered over the popover so dark toasts stay opaque.
const tone = (type: string, name: string) => ({
  [`--${type}-bg`]: `linear-gradient(var(--tone-${name}-bg), var(--tone-${name}-bg)), var(--popover)`,
  [`--${type}-border`]: `var(--tone-${name}-bg)`,
  [`--${type}-text`]: `var(--tone-${name}-fg)`,
});

const Toaster = ({ ...props }: React.ComponentProps<typeof Sonner>) => {
  const { theme = "system" } = useTheme();

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      icons={{
        success: <CircleCheckIcon className="size-4" />,
        info: <InfoIcon className="size-4" />,
        warning: <TriangleAlertIcon className="size-4" />,
        error: <OctagonXIcon className="size-4" />,
        loading: <Loader2Icon className="size-4 animate-spin" />,
      }}
      richColors
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
          "--border-radius": "var(--radius)",
          ...tone("success", "ok"),
          ...tone("info", "info"),
          ...tone("warning", "warn"),
          ...tone("error", "error"),
        } as React.CSSProperties
      }
      toastOptions={{
        classNames: {
          toast: "cn-toast",
        },
      }}
      {...props}
    />
  );
};

// A confirmation belongs to the view that raised it; any path or query change withdraws it.
const DismissOnNavigate = () => {
  const view = `${usePathname()}?${useShallowSearchParams().toString()}`;
  // biome-ignore lint/correctness/useExhaustiveDependencies: runs on every view change
  useEffect(() => {
    dismissConfirm();
  }, [view]);
  return null;
};

// Escape works like Cancel; inside a dialog it closes the dialog, which withdraws the confirmation.
const dismissOnEscape = (e: KeyboardEvent) => {
  if (e.key === "Escape" && !document.querySelector('[role="dialog"]')) dismissConfirm();
};

// alt+T belongs to the notifications toaster: no key matches this one, and its label drops the hint.
const NO_HOTKEY = ["none"];
const dropHotkeyLabel = (el: HTMLElement | null) => el?.setAttribute("aria-label", "Confirmations");

const MID_SCREEN = { top: "45vh" };

const ConfirmToaster = () => {
  useEffect(() => {
    document.addEventListener("keydown", dismissOnEscape);
    return () => document.removeEventListener("keydown", dismissOnEscape);
  }, []);
  return (
    <>
      <Suspense>
        <DismissOnNavigate />
      </Suspense>
      <Toaster
        id={CONFIRM_TOASTER}
        containerAriaLabel="Confirmations"
        position="top-center"
        offset={MID_SCREEN}
        mobileOffset={MID_SCREEN}
        hotkey={NO_HOTKEY}
        ref={dropHotkeyLabel}
      />
    </>
  );
};

const OPEN_CONTAINER = '[role="dialog"], [role="alertdialog"]';
const watchContainers = (changed: () => void) => {
  const observer = new MutationObserver(changed);
  observer.observe(document.body, { childList: true, subtree: true });
  window.addEventListener("resize", changed);
  return () => {
    observer.disconnect();
    window.removeEventListener("resize", changed);
  };
};

// Room a toast stack needs beside a container (sonner's 356px width and 24px offsets), and below
// or above it.
const SIDE_ROOM = 356 + 2 * 24;
const ROW_ROOM = 112;

// "<position>", or "top-center@<px>" / "bottom-center@<px>" as offsets from that edge. With no
// container open, top-right; else beside, below or above the open containers, or failing room for
// that, above a full-height drawer's confirm bar (its first rows under the header hold the controls)
// or just under a dialog's header, so toasts never cover a header, Close, confirm bar or footer.
const placement = (): string => {
  const open = [...document.querySelectorAll<HTMLElement>(OPEN_CONTAINER)];
  const top = open.at(-1);
  if (!top) return "top-right";
  const rects = open.map((el) => el.getBoundingClientRect());
  const left = Math.min(...rects.map((r) => r.left));
  const right = Math.max(...rects.map((r) => r.right));
  const upper = Math.min(...rects.map((r) => r.top));
  const lower = Math.max(...rects.map((r) => r.bottom));
  if (left >= SIDE_ROOM) return "bottom-left";
  if (window.innerWidth - right >= SIDE_ROOM) return "bottom-right";
  if (window.innerHeight - lower >= ROW_ROOM) return "bottom-center";
  if (upper >= ROW_ROOM) return "top-center";
  const bar = top.querySelector('[data-slot="confirm-bar"]')?.getBoundingClientRect();
  if (top.matches('[data-slot="drawer-content"]'))
    return `bottom-center@${Math.round(window.innerHeight - (bar?.top ?? window.innerHeight) + 16)}`;
  const header = top.querySelector('[data-slot="drawer-header"], [data-slot="dialog-header"]');
  const below = (header ?? top).getBoundingClientRect()[header ? "bottom" : "top"];
  // A short body leaves its confirm bar just under the header; the toasts then go under the bar.
  return `top-center@${Math.round((bar && bar.top < below + ROW_ROOM ? bar.bottom : below) + 8)}`;
};

const POSITIONS = [
  "top-right",
  "bottom-left",
  "bottom-right",
  "bottom-center",
  "top-center",
] as const;

const NotificationToaster = () => {
  const [name, at] = useSyncExternalStore(watchContainers, placement, () => "top-right").split("@");
  const position = POSITIONS.find((p) => p === name) ?? "top-right";
  const offset = at ? { [name === "bottom-center" ? "bottom" : "top"]: Number(at) } : undefined;
  return <Toaster position={position} offset={offset} mobileOffset={offset} />;
};

const toastError = (err: unknown) => toast.error(errorCopy(err));

// Toasts once the awaited mutation (mutateAsync) succeeds; mutate's own callbacks are dropped when
// the caller unmounts or mutates again first. Failures already toast from the query client.
const toastOnSuccess = <T,>(
  pending: Promise<T>,
  message: string | ((result: T) => string),
  then?: (result: T) => void,
) =>
  pending.then(
    (result) => {
      toast.success(typeof message === "string" ? message : message(result));
      then?.(result);
    },
    () => {},
  );

export {
  ConfirmToaster,
  NotificationToaster,
  confirmToast,
  dismissConfirm,
  toastError,
  toastOnSuccess,
};
