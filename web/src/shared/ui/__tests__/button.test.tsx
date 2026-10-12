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
import { Button } from "../button";

function Editor({ write, onCancel }: { write: () => Promise<unknown>; onCancel: () => void }) {
  const mutation = useMutation({ mutationFn: write });
  return (
    <>
      <Button pending={mutation.isPending} onClick={() => mutation.mutate()}>
        Save
      </Button>
      <Button onClick={onCancel}>Cancel</Button>
    </>
  );
}

describe("<Button />", () => {
  // Make sure a double press sends one write, and that only the pressed button waits for it.
  it("runs its write once per press, leaving other buttons free", async () => {
    let settle = () => {};
    const write = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          settle = resolve;
        }),
    );
    const onCancel = vi.fn();
    render(
      <QueryClientProvider client={new QueryClient()}>
        <Editor write={write} onCancel={onCancel} />
      </QueryClientProvider>,
    );
    const save = screen.getByRole("button", { name: "Save" });
    fireEvent.click(save);
    fireEvent.click(save);
    await waitFor(() => expect(save).toHaveAttribute("aria-disabled", "true"));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledTimes(1);
    expect(write).toHaveBeenCalledTimes(1);
    settle();
    await waitFor(() => expect(save).not.toHaveAttribute("aria-disabled"));
    fireEvent.click(save);
    await waitFor(() => expect(write).toHaveBeenCalledTimes(2));
  });
});
