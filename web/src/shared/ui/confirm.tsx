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

import { Button } from "@/shared/ui/button";
import {
  type FocusOrigin,
  captureOrigin,
  focusIsLost,
  keepFocus,
  originTarget,
} from "@/shared/ui/focus";
import { useEffect, useId, useRef, useState } from "react";
import { toast } from "sonner";

// A confirmation raised while no dialog or drawer is open waits mid-screen in this toaster.
export const CONFIRM_TOASTER = "confirm";

type Confirm = {
  id: number;
  message: string;
  actionLabel: string;
  onConfirm: () => void;
  origin: FocusOrigin;
  focusAfter?: () => HTMLElement | null;
};
type Host = (confirm: Confirm | null) => void;

// The open dialogs and drawers, oldest first; the newest shows a confirmation raised while it is open.
const hosts: Host[] = [];

// An in-app stand-in for window.confirm: onConfirm runs only if the action is clicked. One stands
// at a time, so a newer confirmation replaces an older one. Cancel and Escape return focus to the
// raiser; after the action it goes to `focusAfter` when the caller names one, else to the raiser,
// or to its region once the action removed it.
let standing: { host?: Host; toastId?: string; origin: FocusOrigin } | undefined;
let confirmSeq = 0;

export function confirmToast(
  message: string,
  actionLabel: string,
  onConfirm: () => void,
  focusAfter?: () => HTMLElement | null,
) {
  const origin = captureOrigin();
  dismissConfirm();
  const id = ++confirmSeq;
  const host = hosts.at(-1);
  if (host) {
    standing = { host, origin };
    host({ id, message, actionLabel, onConfirm, origin, focusAfter });
    return;
  }
  // A fresh id each time: re-using a dismissed toast's id would show the new one already removed.
  standing = { toastId: `${CONFIRM_TOASTER}-${id}`, origin };
  toast.warning(message, {
    id: standing.toastId,
    toasterId: CONFIRM_TOASTER,
    duration: Number.POSITIVE_INFINITY,
    action: { label: actionLabel, onClick: () => act(onConfirm, origin, focusAfter) },
    cancel: { label: "Cancel", onClick: dismissConfirm },
  });
}

// Withdraws a standing confirmation whose action no longer matches what the user is doing; true
// when one stood.
export function dismissConfirm(): boolean {
  if (!standing) return false;
  const { host, toastId, origin } = standing;
  standing = undefined;
  host?.(null);
  if (toastId) {
    toast.dismiss(toastId);
    if (document.activeElement?.closest("[data-sonner-toaster]")) originTarget(origin)?.focus();
  }
  return true;
}

function act(
  onConfirm: () => void,
  origin: FocusOrigin,
  focusAfter: () => HTMLElement | null = () => originTarget(origin),
) {
  standing = undefined;
  onConfirm();
  keepFocus(focusAfter);
  focusAfter()?.focus();
}

function ConfirmBar({ confirm }: { confirm: Confirm }) {
  const cancel = useRef<HTMLButtonElement>(null);
  const acted = useRef(false);
  const messageId = useId();
  useEffect(() => {
    cancel.current?.focus();
    return () => {
      if (!acted.current && focusIsLost()) originTarget(confirm.origin)?.focus();
    };
  }, [confirm.origin]);
  return (
    <div
      role="group"
      aria-labelledby={messageId}
      data-slot="confirm-bar"
      className="sticky bottom-0 z-10 space-y-3 rounded-md border border-[color:var(--tone-warn-bg)] bg-popover p-3 text-sm text-popover-foreground shadow-lg"
    >
      <p id={messageId} className="[overflow-wrap:anywhere]">
        {confirm.message}
      </p>
      <div className="flex justify-end gap-2">
        <Button ref={cancel} type="button" variant="outline" size="sm" onClick={dismissConfirm}>
          Cancel
        </Button>
        <Button
          type="button"
          size="sm"
          onClick={() => {
            if (acted.current) return;
            acted.current = true;
            dismissConfirm();
            act(confirm.onConfirm, confirm.origin, confirm.focusAfter);
          }}
        >
          {confirm.actionLabel}
        </Button>
      </div>
    </div>
  );
}

// Makes the calling dialog or drawer host the confirmations raised while it is the newest open
// container; returns the bar to render inside it, pinned to its bottom.
export function useConfirmBar(): React.ReactNode {
  const [confirm, setConfirm] = useState<Confirm | null>(null);
  useEffect(() => {
    hosts.push(setConfirm);
    return () => void hosts.splice(hosts.indexOf(setConfirm), 1);
  }, []);
  return confirm ? <ConfirmBar key={confirm.id} confirm={confirm} /> : null;
}
