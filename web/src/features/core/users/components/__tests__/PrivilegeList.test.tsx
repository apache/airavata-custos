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

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { groupPrivileges, PrivilegeList } from "../PrivilegeList";

describe("PrivilegeList", () => {
  it("groups real read/write keys without inventing write", () => {
    expect(groupPrivileges(["core:traces:read", "core:users:read", "core:users:write"])).toEqual([
      { kind: "rw", key: "core:traces", read: true, write: false },
      { kind: "rw", key: "core:users", read: true, write: true },
    ]);
  });

  it("preserves meta and unknown privileges verbatim", () => {
    render(<PrivilegeList privileges={["core:roles:manage", "connector:custom:inspect"]} />);
    expect(screen.getByText("core:roles:manage")).toBeInTheDocument();
    expect(screen.getByText("connector:custom:inspect")).toBeInTheDocument();
  });

  it("deduplicates repeated keys", () => {
    render(<PrivilegeList privileges={["core:roles:manage", "core:roles:manage"]} />);
    expect(screen.getAllByText("core:roles:manage")).toHaveLength(1);
  });
});
