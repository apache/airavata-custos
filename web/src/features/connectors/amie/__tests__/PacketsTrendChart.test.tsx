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
import { fillDays } from "../components/PacketsTrendChart";

describe("fillDays", () => {
  it("zero-fills every day of the window and drops buckets outside it", () => {
    const rows = fillDays(
      [
        { date: "2026-06-08", status: "FAILED", type: "request_account_create", count: 2 },
        { date: "2026-05-01", status: "FAILED", type: "request_account_create", count: 9 },
      ],
      Date.parse("2026-06-08T12:00:00Z"),
    );
    expect(rows).toHaveLength(30);
    expect(rows[0]).toMatchObject({ date: "2026-05-10", FAILED: 0 });
    expect(rows.at(-1)).toMatchObject({ date: "2026-06-08", FAILED: 2 });
  });
});
