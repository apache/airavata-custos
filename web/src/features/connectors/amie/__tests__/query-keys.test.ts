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

import { describe, expect, it } from "vitest";
import { amieKeys } from "../queries";

describe("amieKeys factory", () => {
  it("all is stable", () => {
    expect(amieKeys.all).toEqual(["amie"]);
  });

  it("packets is parameterized by filter object", () => {
    expect(amieKeys.packets({ status: "FAILED" })).toEqual([
      "amie",
      "packets",
      "list",
      { status: "FAILED" },
    ]);
  });

  it("packet detail keys are namespaced by id", () => {
    expect(amieKeys.packet("pkt-1")).toEqual(["amie", "packets", "detail", "pkt-1"]);
  });

  it("events keys are namespaced by packet id", () => {
    expect(amieKeys.events("pkt-1")).toEqual(["amie", "packets", "events", "pkt-1"]);
  });

  it("stats keys carry the window param", () => {
    expect(amieKeys.stats({ window: "30d" })).toEqual(["amie", "stats", { window: "30d" }]);
  });
});
