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
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const dismissConfirm = vi.hoisted(() => vi.fn());
vi.mock("../sonner", () => ({ dismissConfirm }));

import { SideDrawer } from "../SideDrawer";

describe("<SideDrawer />", () => {
  it("withdraws a standing confirmation when it opens and one raised in it when it closes", () => {
    const drawer = (open: boolean) => (
      <SideDrawer open={open} onOpenChange={() => {}} title="Override">
        body
      </SideDrawer>
    );
    const { rerender } = render(drawer(true));
    expect(dismissConfirm).toHaveBeenCalledTimes(1);
    rerender(drawer(false));
    expect(dismissConfirm).toHaveBeenCalledTimes(2);
  });

  it("closes on a press outside it, except on a row or toggle that opens drawers", () => {
    const onOpenChange = vi.fn();
    render(
      <>
        <table>
          <tbody>
            <tr data-drawer-trigger="">
              <td>Pat</td>
            </tr>
          </tbody>
        </table>
        <p>Page</p>
        <SideDrawer open onOpenChange={onOpenChange} title="Pat" modal={false}>
          body
        </SideDrawer>
      </>,
    );
    const press = (el: Element) => {
      fireEvent.pointerDown(el);
      fireEvent.mouseDown(el);
      fireEvent.click(el);
    };
    press(screen.getByText("Pat", { selector: "td" }));
    expect(onOpenChange).not.toHaveBeenCalled();
    press(screen.getByText("Page"));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("returns focus to the labelled section a removed trigger sat in", async () => {
    function Timeline({ open, withTrigger }: { open: boolean; withTrigger: boolean }) {
      return (
        <>
          <section aria-label="Timeline" tabIndex={-1}>
            {withTrigger ? <button type="button">Open event</button> : null}
          </section>
          <SideDrawer open={open} onOpenChange={() => {}} title="Event">
            body
          </SideDrawer>
        </>
      );
    }
    const { rerender } = render(<Timeline open={false} withTrigger />);
    screen.getByRole("button", { name: "Open event" }).focus();
    rerender(<Timeline open withTrigger />);
    rerender(<Timeline open withTrigger={false} />);
    rerender(<Timeline open={false} withTrigger={false} />);
    await waitFor(() => expect(screen.getByRole("region", { name: "Timeline" })).toHaveFocus());
  });
});
