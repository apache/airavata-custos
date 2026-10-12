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

import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const url = vi.hoisted(() => ({ path: "/projects/p-1", query: "tab=members" }));
vi.mock("next/navigation", () => ({
  usePathname: () => url.path,
  useSearchParams: () => new URLSearchParams(url.query),
}));
vi.mock("next-themes", () => ({ useTheme: () => ({}) }));
vi.mock("sonner", () => ({
  Toaster: () => null,
  toast: { dismiss: vi.fn(), warning: vi.fn() },
}));

import { toast } from "sonner";
import { Dialog, DialogContent, DialogFooter, DialogTitle } from "../dialog";
import { ConfirmToaster, confirmToast } from "../sonner";

const raisedId = () => vi.mocked(toast.warning).mock.lastCall?.[1]?.id;

describe("<ConfirmToaster />", () => {
  it("withdraws a standing confirmation when only the query changes", () => {
    const { rerender } = render(<ConfirmToaster />);
    confirmToast("Remove?", "Remove", vi.fn());
    const id = raisedId();
    vi.mocked(toast.dismiss).mockClear();
    url.query = "tab=usage";
    rerender(<ConfirmToaster />);
    expect(toast.dismiss).toHaveBeenCalledWith(id);
  });

  it("withdraws a standing confirmation on Escape when no dialog is open", () => {
    render(<ConfirmToaster />);
    confirmToast("Detach?", "Detach", vi.fn());
    const id = raisedId();
    vi.mocked(toast.dismiss).mockClear();
    fireEvent.keyDown(document, { key: "Escape" });
    expect(toast.dismiss).toHaveBeenCalledWith(id);
  });

  // Make sure a confirmation raised right after another one is withdrawn gets its own id, so the
  // toaster does not show it already removed.
  it("gives each confirmation a fresh id", () => {
    confirmToast("Mark deleted?", "Mark deleted", vi.fn());
    const first = raisedId();
    confirmToast("Mark deleted?", "Mark deleted", vi.fn());
    expect(raisedId()).not.toBe(first);
    expect(toast.dismiss).toHaveBeenCalledWith(first);
  });

  // Make sure a confirmation raised in a dialog stands inside it, takes focus on Cancel, acts once,
  // and that Escape withdraws it before a second Escape closes the dialog.
  it("hosts a confirmation inside the open dialog", () => {
    const onOpenChange = vi.fn();
    const onConfirm = vi.fn();
    render(
      <Dialog open onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogTitle>Edit role</DialogTitle>
          <DialogFooter>
            <button type="button">Delete role</button>
          </DialogFooter>
        </DialogContent>
      </Dialog>,
    );
    act(() => confirmToast("Delete role?", "Delete", onConfirm));
    const dialog = screen.getByRole("dialog");
    const bar = within(dialog).getByRole("group", { name: "Delete role?" });
    expect(toast.warning).not.toHaveBeenCalledWith("Delete role?", expect.anything());
    expect(within(bar).getByRole("button", { name: "Cancel" })).toHaveFocus();
    fireEvent.keyDown(dialog, { key: "Escape" });
    expect(screen.queryByRole("group", { name: "Delete role?" })).toBeNull();
    expect(onOpenChange).not.toHaveBeenCalled();
    fireEvent.keyDown(dialog, { key: "Escape" });
    expect(onOpenChange).toHaveBeenCalledWith(false, expect.anything());

    act(() => confirmToast("Delete role?", "Delete", onConfirm));
    const action = screen.getByRole("button", { name: "Delete" });
    fireEvent.click(action);
    fireEvent.click(action);
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  // Make sure focus never falls to <body> after a page confirm: it returns to the raiser, and goes
  // to the raiser's list once the action removed it.
  it("returns focus to the raiser, else to its list", async () => {
    const { rerender } = render(
      <section aria-label="Members">
        <button type="button">Remove</button>
      </section>,
    );
    const raiser = screen.getByRole("button", { name: "Remove" });
    raiser.focus();
    confirmToast("Remove member?", "Remove", () => rerender(<section aria-label="Members" />));
    const action = vi.mocked(toast.warning).mock.lastCall?.[1]?.action as unknown as {
      onClick: () => void;
    };
    act(() => action.onClick());
    await act(() => Promise.resolve());
    expect(screen.getByRole("region", { name: "Members" })).toHaveFocus();
  });
});
