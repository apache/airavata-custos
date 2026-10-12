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
vi.mock("@/features/core/audit/components/ViewTraceLink", () => ({ ViewTraceLink: () => null }));

import { deliveryFixtures } from "../__fixtures__/deliveries";
import { DeliveryDetailDrawer } from "../components/DeliveryDetailDrawer";

describe("<DeliveryDetailDrawer />", () => {
  // Make sure the retry confirm stands in the drawer's shared bar, Escape withdraws it before closing
  // the drawer, and the action retries once and moves focus to the history.
  it("confirms a retry in the drawer's confirm bar", async () => {
    const failed = deliveryFixtures.find((d) => d.status === "FAILED");
    const onRetry = vi.fn();
    const onOpenChange = vi.fn();
    render(
      <DeliveryDetailDrawer
        deliveryId={failed?.id}
        onOpenChange={onOpenChange}
        delivery={failed}
        isLoading={false}
        error={null}
        canRetry
        confirmRetryOnOpen={false}
        isRetrying={false}
        onRetry={onRetry}
        onRefresh={vi.fn()}
      />,
    );
    fireEvent.click(await screen.findByRole("button", { name: "Retry" }));
    const cancel = await screen.findByRole("button", { name: "Cancel" });
    expect(cancel).toHaveFocus();
    fireEvent.keyDown(cancel, { key: "Escape" });
    expect(screen.queryByRole("button", { name: "Retry now" })).toBeNull();
    expect(onOpenChange).not.toHaveBeenCalled();

    // A press that only raised a confirm frees the button once it settles.
    await new Promise((settle) => setTimeout(settle, 60));
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    const action = await screen.findByRole("button", { name: "Retry now" });
    fireEvent.click(action);
    fireEvent.click(action);
    expect(onRetry).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("region", { name: "Try history" })).toHaveFocus();
  });
});
