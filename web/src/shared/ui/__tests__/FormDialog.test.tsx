import { ApiError } from "@/shared/api/client";
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
import { QueryClient, QueryClientProvider, skipToken, useQuery } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { FormDialog, edited } from "../FormDialog";

describe("<FormDialog />", () => {
  it("says the record no longer exists, without a retry, and refreshes its parent when its refetch on open 404s", () => {
    const onGone = vi.fn();
    const record = {
      refetch: vi.fn(),
      dataUpdatedAt: 0,
      errorUpdatedAt: Number.MAX_SAFE_INTEGER,
      error: new ApiError(404, "/x", null, "override not found"),
    };
    render(
      <FormDialog
        open
        title="Edit override"
        submitLabel="Save"
        isPending={false}
        onSubmit={vi.fn()}
        record={record}
        gone="This override no longer exists"
        onGone={onGone}
      >
        fields
      </FormDialog>,
    );
    expect(screen.getByText("This override no longer exists")).toBeVisible();
    expect(screen.queryByText("fields")).toBeNull();
    expect(screen.queryByRole("button", { name: "Try again" })).toBeNull();
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
    expect(onGone).toHaveBeenCalledTimes(1);
  });

  it("refetches a record whose query turns enabled in the render that opens it, then focuses its first field", async () => {
    const client = new QueryClient();
    client.setQueryData(["record"], "cached");
    const queryFn = vi.fn(async () => "fresh");
    function Editor({ open }: { open: boolean }) {
      const record = useQuery({
        queryKey: ["record"],
        queryFn: open ? queryFn : skipToken,
        staleTime: Number.POSITIVE_INFINITY,
      });
      return (
        <FormDialog
          open={open}
          title="Edit"
          submitLabel="Save"
          isPending={false}
          onSubmit={vi.fn()}
          record={open ? record : undefined}
        >
          <input aria-label="Name" defaultValue={String(record.data)} />
        </FormDialog>
      );
    }
    const editor = (open: boolean) => (
      <QueryClientProvider client={client}>
        <Editor open={open} />
      </QueryClientProvider>
    );
    const { rerender } = render(editor(false));
    rerender(editor(true));
    const field = await screen.findByDisplayValue("fresh");
    await waitFor(() => expect(field).toHaveFocus());
    expect(queryFn).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Save" })).toBeEnabled();
  });
});

describe("edited", () => {
  const record = { id: "r-1", name: "gpu", amount: 4, note: "" };

  it("sends the full record with the form's fields over it", () => {
    expect(edited(record, { name: "gpu-2", note: "" })).toEqual({ ...record, name: "gpu-2" });
  });

  it("is null when no field differs, a blank field matching an absent one", () => {
    expect(edited({ id: "r-1", name: "gpu" }, { name: "gpu", note: "" })).toBeNull();
  });
});
