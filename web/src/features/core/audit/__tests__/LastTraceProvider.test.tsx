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

import { act, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { recordTraceId } from "@/shared/api/last-trace-id";

beforeEach(() => recordTraceId(null));
afterEach(() => recordTraceId(null));

describe("LastTraceProvider", () => {
  it("exposes the singleton value via context to its descendants", async () => {
    recordTraceId("a3b1c92d3f4e5a6b7c8d9e0f12345678");
    const { LastTraceProvider, useLastTraceContext } = await import(
      "../components/LastTraceProvider"
    );
    function Consumer() {
      const id = useLastTraceContext();
      return <span data-testid="ctx">{id ?? "none"}</span>;
    }
    render(
      <LastTraceProvider>
        <Consumer />
      </LastTraceProvider>,
    );
    await waitFor(() => {
      expect(screen.getByTestId("ctx").textContent).toBe("a3b1c92d3f4e5a6b7c8d9e0f12345678");
    });
  });

  it("useLastTraceId inside the provider reflects post-mount updates", async () => {
    const { LastTraceProvider } = await import("../components/LastTraceProvider");
    const { useLastTraceId } = await import("../queries");
    function Consumer() {
      const id = useLastTraceId();
      return <span data-testid="hook">{id ?? "none"}</span>;
    }
    render(
      <LastTraceProvider>
        <Consumer />
      </LastTraceProvider>,
    );
    expect(screen.getByTestId("hook").textContent).toBe("none");
    act(() => {
      recordTraceId("b4c2da3e405f6b7c8d9e0f1234567890");
    });
    await waitFor(() => {
      expect(screen.getByTestId("hook").textContent).toBe("b4c2da3e405f6b7c8d9e0f1234567890");
    });
  });
});
