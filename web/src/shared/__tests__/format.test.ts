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
import { formatDay, formatRate, formatSU, formatTimestamp, formatUtc } from "../format";
import { dateInput, isoDate } from "../ui/FormDialog";

describe("date-only fields", () => {
  it("send and show the picked UTC calendar day west of UTC", () => {
    process.env.TZ = "America/Los_Angeles";
    const form = new FormData();
    form.set("end", "2026-08-01");
    const sent = isoDate(form, "end");
    expect(sent).toBe("2026-08-01T00:00:00Z");
    expect(dateInput(sent)).toBe("2026-08-01");
    expect(formatDay(sent)).toBe(
      new Date(Date.UTC(2026, 7, 1, 12)).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      }),
    );
  });
});

describe("formatSU", () => {
  it("keeps SU amounts to six decimals", () => {
    expect(formatSU(0.004)).toBe((0.004).toLocaleString());
    expect(formatSU(1234.5678919)).toBe(
      (1234.567892).toLocaleString(undefined, { maximumFractionDigits: 6 }),
    );
  });
});

describe("formatRate", () => {
  it("keeps sub-cent rates and whole parts, and reads only tiny ones in exponent form", () => {
    expect(formatRate(0.001)).toBe((0.001).toLocaleString());
    expect(formatRate(1e-300)).toBe("1E-300");
    expect(formatRate(999999.99)).toBe(
      (999999.99).toLocaleString(undefined, { maximumFractionDigits: 2 }),
    );
  });
});

describe("event timestamps", () => {
  it("show the viewer's local time with the UTC instant behind it", () => {
    process.env.TZ = "America/Los_Angeles";
    const at = "2026-08-01T03:04:05.123Z";
    expect(formatTimestamp(at)).toMatch(/Jul 31, 2026.*8:04:05/);
    expect(formatUtc(at)).toBe("2026-08-01 03:04:05.123 UTC");
  });
});
