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

import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { XIcon } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { useConfirmBar } from "@/shared/ui/confirm";
import { focusHeadingOf, useFocusOrigin, useScrollStop, wrapTab } from "@/shared/ui/focus";
import { dismissConfirm } from "@/shared/ui/sonner";

export type SideDrawerWidth = "sm" | "md" | "lg";

const widthClass: Record<SideDrawerWidth, string> = {
  sm: "max-w-sm",
  md: "max-w-[480px]",
  lg: "max-w-2xl",
};

export type SideDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: React.ReactNode;
  description?: React.ReactNode;
  children: React.ReactNode;
  width?: SideDrawerWidth;
  // A modal drawer dims and blocks the page; only one raised from a button outside any list is.
  modal?: boolean;
  // Where focus goes on open, as Base UI's Popup initialFocus (a form drawer: its first field).
  initialFocus?: DialogPrimitive.Popup.Props["initialFocus"];
  // The shown record's id: a swap to another record re-reads where focus returns on close.
  recordKey?: string;
};

export function SideDrawer({
  open,
  onOpenChange,
  title,
  description,
  children,
  width = "md",
  modal = false,
  initialFocus,
  recordKey,
}: SideDrawerProps) {
  const returnTo = useFocusOrigin(open, recordKey);
  const scrollStop = useScrollStop<HTMLDivElement>();
  // Opening the drawer withdraws a standing confirmation, and one raised in it ends with it.
  React.useEffect(() => {
    if (!open) return;
    dismissConfirm();
    return () => void dismissConfirm();
  }, [open]);
  return (
    <DialogPrimitive.Root
      open={open}
      // Only an explicit action closes it: focus leaving because its trigger unmounted does not. A
      // press on a row or toggle that opens drawers is left to it, so it can swap or toggle this one.
      // Escape withdraws a standing confirmation first; a second Escape closes the drawer.
      onOpenChange={(next, details) => {
        if (next) return onOpenChange(next);
        if (details.reason === "focus-out") return;
        const target = details.event?.target;
        if (
          details.reason === "outside-press" &&
          target instanceof Element &&
          target.closest("[data-drawer-trigger]")
        )
          return details.cancel();
        if (dismissConfirm() && details.reason === "escape-key") return details.cancel();
        onOpenChange(next);
      }}
      modal={modal}
    >
      <DialogPrimitive.Portal>
        {modal ? (
          <DialogPrimitive.Backdrop className="fixed inset-0 z-50 bg-black/30 data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0" />
        ) : null}
        <DialogPrimitive.Popup
          data-slot="drawer-content"
          initialFocus={initialFocus}
          finalFocus={returnTo}
          // A link in the drawer leaves the page with it; focus goes to the destination's h1.
          onClick={(e) => {
            const link = e.target instanceof Element ? e.target.closest("a[href]") : null;
            if (link instanceof HTMLAnchorElement && link.origin === window.location.origin)
              focusHeadingOf(link.href);
          }}
          onKeyDown={wrapTab}
          className={cn(
            "fixed inset-y-0 right-0 z-50 flex w-full flex-col bg-popover text-popover-foreground shadow-xl outline-none",
            "data-open:animate-in data-open:slide-in-from-right-12 data-closed:animate-out data-closed:slide-out-to-right-12",
            widthClass[width],
          )}
        >
          <div
            data-slot="drawer-header"
            className="flex items-start justify-between gap-3 border-b border-border px-6 py-4"
          >
            <div className="min-w-0 space-y-1">
              {title ? (
                <DialogPrimitive.Title className="font-display text-lg font-semibold text-foreground [overflow-wrap:anywhere]">
                  {title}
                </DialogPrimitive.Title>
              ) : null}
              {description ? (
                <DialogPrimitive.Description className="text-sm text-muted-foreground">
                  {description}
                </DialogPrimitive.Description>
              ) : null}
            </div>
            <DialogPrimitive.Close
              render={
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label="Close"
                  className="h-8 w-8 rounded-md"
                />
              }
            >
              <XIcon className="size-4" />
            </DialogPrimitive.Close>
          </div>
          <div
            {...scrollStop}
            data-slot="drawer-body"
            role="region"
            aria-label={typeof title === "string" ? title : "Details"}
            className="flex-1 overflow-y-auto px-6 py-4 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
          >
            <div>{children}</div>
          </div>
          <DrawerConfirm />
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}

// A confirmation raised in the drawer sits below its scrolling body, clear of the header and Close.
function DrawerConfirm() {
  const bar = useConfirmBar();
  return bar ? <div className="border-t border-border px-6 py-3">{bar}</div> : null;
}
