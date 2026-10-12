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

import { QueryClient, QueryClientProvider, useMutation } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("sonner", () => ({ toast: { success: vi.fn() } }));

import { toast } from "sonner";
import { toastOnSuccess } from "../sonner";

let finish: (() => void) | undefined;

function DetachButton() {
  const detach = useMutation({
    mutationFn: () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
  });
  return (
    <button type="button" onClick={() => toastOnSuccess(detach.mutateAsync(), "Resource detached")}>
      Detach
    </button>
  );
}

describe("toastOnSuccess", () => {
  it("still toasts when the caller unmounts before the mutation settles", async () => {
    const { unmount } = render(
      <QueryClientProvider client={new QueryClient()}>
        <DetachButton />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Detach" }));
    await waitFor(() => expect(finish).toBeDefined());
    unmount();
    finish?.();
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith("Resource detached"));
  });
});
