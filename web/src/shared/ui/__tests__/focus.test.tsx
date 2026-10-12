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
import { describe, expect, it } from "vitest";
import { keepFocus, useFocusOrigin, wrapTab } from "../focus";

describe("keepFocus", () => {
  it("treats focus on the open dialog's popup itself as lost", () => {
    render(
      <div role="dialog" tabIndex={-1} data-testid="popup">
        <button type="button">Accounts</button>
      </div>,
    );
    const accounts = screen.getByRole("button", { name: "Accounts" });
    screen.getByTestId("popup").focus();
    keepFocus(() => accounts);
    expect(accounts).toHaveFocus();
  });

  it("treats focus on the open drawer's own scroll region as lost", () => {
    render(
      <div role="dialog">
        <button type="button">Status</button>
        <div role="region" aria-label="Cluster account" data-slot="drawer-body" tabIndex={-1} />
      </div>,
    );
    const status = screen.getByRole("button", { name: "Status" });
    screen.getByRole("region", { name: "Cluster account" }).focus();
    keepFocus(() => status);
    expect(status).toHaveFocus();
  });
});

describe("wrapTab", () => {
  it("never counts content inside a closed <details> as the last stop", () => {
    // Chromium's answer: a closed <details> shows only its summary.
    Element.prototype.checkVisibility = function (this: Element) {
      const details = this.closest("details");
      return !details || details.open || this.tagName === "SUMMARY";
    };
    render(
      <div role="dialog" onKeyDown={wrapTab}>
        <button type="button">Close</button>
        <details>
          <summary>Payload</summary>
          {/* biome-ignore lint/a11y/noNoninteractiveTabindex: JsonBlock's scrolling region takes focus */}
          <section tabIndex={0} aria-label="Payload JSON" />
        </details>
      </div>,
    );
    screen.getByText("Payload").focus();
    fireEvent.keyDown(screen.getByText("Payload"), { key: "Tab" });
    expect(screen.getByRole("button", { name: "Close" })).toHaveFocus();
  });
});

describe("useFocusOrigin", () => {
  it("returns to the opener of the row a press swapped to, once that row re-rendered", async () => {
    let returnTo: () => false = () => false;
    function List({ open, record, version }: { open: boolean; record?: string; version: number }) {
      returnTo = useFocusOrigin(open, record);
      return (
        <table>
          <tbody>
            {["a", "b"].map((key) => (
              <tr key={`${key}-${version}`} data-row-key={key}>
                <td>
                  <input type="checkbox" aria-label={`Select ${key}`} />
                </td>
                <td>
                  <button type="button" data-row-opener>
                    Open {key}
                  </button>
                </td>
                <td>Row {key}</td>
              </tr>
            ))}
          </tbody>
        </table>
      );
    }
    const { rerender } = render(<List open={false} version={0} />);
    const openA = screen.getByRole("button", { name: "Open a" });
    openA.focus();
    fireEvent.keyDown(openA, { key: "Enter" });
    rerender(<List open record="a" version={0} />);
    fireEvent.pointerDown(screen.getByText("Row b"));
    rerender(<List open record="b" version={0} />);
    rerender(<List open record="b" version={1} />);
    rerender(<List open={false} version={1} />);
    (document.activeElement as HTMLElement).blur();
    returnTo();
    await Promise.resolve();
    expect(screen.getByRole("button", { name: "Open b" })).toHaveFocus();
  });
});
