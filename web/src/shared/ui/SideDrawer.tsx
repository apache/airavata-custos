"use client";

import { Dialog as DialogPrimitive } from "@base-ui/react/dialog";
import { XIcon } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";

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
  // Non-modal drawers skip the dimming backdrop and leave the rest of the
  // page interactive — e.g. clicking another table row while the drawer is
  // open should swap its contents rather than being blocked by an overlay.
  modal?: boolean;
  // Base UI closes on any outside pointer press by default, which races with
  // a caller's own click handling (e.g. a row toggling the same drawer) —
  // the outside-press listener fires before React's onClick, so the caller's
  // "close if already open" check always sees a fresh `null` state. Set this
  // when the caller owns open/close entirely (row clicks, toggle buttons).
  disablePointerDismissal?: boolean;
};

export function SideDrawer({
  open,
  onOpenChange,
  title,
  description,
  children,
  width = "md",
  modal = true,
  disablePointerDismissal = false,
}: SideDrawerProps) {
  return (
    <DialogPrimitive.Root
      open={open}
      onOpenChange={onOpenChange}
      modal={modal}
      disablePointerDismissal={disablePointerDismissal}
    >
      <DialogPrimitive.Portal>
        {modal ? (
          <DialogPrimitive.Backdrop className="fixed inset-0 z-50 bg-black/30 data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0" />
        ) : null}
        <DialogPrimitive.Popup
          className={cn(
            "fixed inset-y-0 right-0 z-50 flex w-full flex-col bg-popover text-popover-foreground shadow-xl outline-none",
            "data-open:animate-in data-open:slide-in-from-right-12 data-closed:animate-out data-closed:slide-out-to-right-12",
            widthClass[width],
          )}
        >
          <header className="flex items-start justify-between border-b border-border px-6 py-4">
            <div className="space-y-1">
              {title ? (
                <DialogPrimitive.Title className="font-display text-lg font-semibold text-foreground">
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
          </header>
          <div className="flex-1 overflow-y-auto px-6 py-4">{children}</div>
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
