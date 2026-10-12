/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { describe, expect, it, vi } from "vitest";

vi.mock("@/shared/auth/auth", () => ({ auth: async () => ({ privileges: [] }) }));
vi.mock("@/shared/casl/abilities", () => ({ defineAbilitiesFor: () => ({ can: () => true }) }));
const redirect = vi.hoisted(() =>
  vi.fn(() => {
    throw new Error("NEXT_REDIRECT");
  }),
);
vi.mock("next/navigation", () => ({ redirect }));

import AmieIndex from "../page";

describe("/admin/amie", () => {
  // Make sure a deep link such as ?packet= survives the bounce to the first tab.
  it("redirects to the first readable tab keeping the query", async () => {
    await expect(AmieIndex({ searchParams: Promise.resolve({ packet: "p-1" }) })).rejects.toThrow();
    expect(redirect).toHaveBeenCalledWith("/admin/amie/packets?packet=p-1");
  });
});
