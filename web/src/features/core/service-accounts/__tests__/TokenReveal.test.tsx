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
import { Dialog, DialogContent } from "@/shared/ui/dialog";
import { TokenReveal } from "../components/TokenReveal";

describe("TokenReveal", () => {
  // Make sure the dialog cannot be finished before the admin says the one-time token is stored.
  it("keeps Done disabled until the token is marked as stored", () => {
    const onDone = vi.fn();
    render(
      <Dialog open>
        <DialogContent>
          <TokenReveal accountName="Notebook Hub" token="csa_notebook-hub.SECRET" onDone={onDone} />
        </DialogContent>
      </Dialog>,
    );

    const done = screen.getByRole("button", { name: "Done" });
    expect(done).toBeDisabled();

    fireEvent.click(screen.getByRole("checkbox", { name: "I have stored the token" }));
    expect(done).toBeEnabled();

    fireEvent.click(done);
    expect(onDone).toHaveBeenCalledOnce();
  });
});
