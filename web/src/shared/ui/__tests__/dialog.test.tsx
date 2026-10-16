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
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const dismissConfirm = vi.hoisted(() => vi.fn());
vi.mock("@/shared/ui/sonner", () => ({ dismissConfirm }));

import { Dialog, DialogContent, DialogTitle } from "../dialog";

describe("<Dialog />", () => {
  it("withdraws a standing confirmation when it opens, on Escape and when its owner closes it", () => {
    const onOpenChange = vi.fn();
    const dialog = (open: boolean) => (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogTitle>Merge</DialogTitle>
        </DialogContent>
      </Dialog>
    );
    const { rerender } = render(dialog(true));
    expect(dismissConfirm).toHaveBeenCalledTimes(1);
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    expect(onOpenChange).toHaveBeenCalledWith(false, expect.anything());
    expect(dismissConfirm).toHaveBeenCalledTimes(2);
    rerender(dialog(false));
    expect(dismissConfirm).toHaveBeenCalledTimes(3);
  });
});
