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

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { type ReactNode, createElement } from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/generated/core/sdk.gen", () => ({
  deleteProjectsById: () => Promise.reject(new ApiError(404, "/x", null, "project not found")),
  getProjectsByIdMembers: () => Promise.reject(new ApiError(404, "/x", null, "project not found")),
}));

import type { ProjectResponse } from "@/generated/core/types.gen";
import { ApiError } from "@/shared/api/client";
import { participantPage, projectKeys, useDeleteProject, useProjectMembers } from "../queries";

describe("project queries", () => {
  it("drops the cached project when a delete finds it gone, so the page shows it unavailable", async () => {
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    client.setQueryData(projectKeys.detail("5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b"), {
      id: "5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b",
    });
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, children);
    const { result } = renderHook(() => useDeleteProject(), { wrapper });
    result.current.mutate({ path: { id: "5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b" } });
    await waitFor(() =>
      expect(
        client.getQueryData(projectKeys.detail("5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b")),
      ).toBeUndefined(),
    );
  });

  it("drops the cached project when its members 404, so the page shows it unavailable", async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    client.setQueryData(projectKeys.detail("5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b"), {
      id: "5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b",
    });
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, children);
    renderHook(() => useProjectMembers("5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b"), { wrapper });
    await waitFor(() =>
      expect(
        client.getQueryData(projectKeys.detail("5b0f6a7e-2c1d-4e8f-9a3b-1c2d3e4f5a6b")),
      ).toBeUndefined(),
    );
  });

  it("searches, filters and pages a participant's projects locally", () => {
    const project = (i: number, status: ProjectResponse["status"]): ProjectResponse => ({
      id: `p-${i}`,
      originated_id: `BIO${i}`,
      title: i % 2 ? "Genomics" : "Climate",
      origination: "ACCESS",
      project_pi_id: "u-1",
      status,
      created_time: "2026-04-01T00:00:00.000Z",
    });
    const items = [1, 2, 3, 4, 5].map((i) => project(i, i === 5 ? "INACTIVE" : "ACTIVE"));
    const page = participantPage(
      { items, total: items.length },
      { q: "GENOM", status: "ACTIVE", limit: 1, offset: 1 },
    );
    expect(page).toEqual({ items: [items[2]], total: 2 });
  });
});
