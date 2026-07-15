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
import { RevokeCertificateDialog } from "../components/RevokeCertificateDialog";

function renderDialog(overrides: Partial<React.ComponentProps<typeof RevokeCertificateDialog>> = {}) {
  const props = {
    serialNumber: 42,
    open: true,
    onOpenChange: vi.fn(),
    onSubmit: vi.fn(),
    isPending: false,
    ...overrides,
  };
  return { ...render(<RevokeCertificateDialog {...props} />), props };
}

describe("<RevokeCertificateDialog />", () => {
  it("disables confirm until a reason is entered, then submits the trimmed reason", () => {
    const { props } = renderDialog();
    const confirm = screen.getByRole("button", { name: /confirm revoke/i });
    expect(confirm).toBeDisabled();

    fireEvent.change(screen.getByLabelText(/reason/i), { target: { value: "  compromised  " } });
    expect(confirm).toBeEnabled();

    fireEvent.click(confirm);
    expect(props.onSubmit).toHaveBeenCalledWith("compromised");
  });

  it("surfaces an error message", () => {
    renderDialog({ error: "You can only revoke your own certificates" });
    expect(screen.getByText(/only revoke your own/i)).toBeInTheDocument();
  });

  it("shows a pending label while revoking", () => {
    renderDialog({ isPending: true });
    expect(screen.getByRole("button", { name: /revoking/i })).toBeDisabled();
  });
});
