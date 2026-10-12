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

import {
  DEFAULT_FILTERS,
  type ListFilters,
  hasActiveFilters,
  parseFilters,
  serializeFilters,
} from "@/features/core/audit/components/traceListUrlState";
import { describe, expect, it } from "vitest";

describe("traceListUrlState", () => {
  // Make sure a pasted or edited URL cannot widen the query: unknown status,
  // window, page or page size values (including a partly numeric one) fall back to the defaults.
  it("drops values the list does not know", () => {
    const params = new URLSearchParams(
      "status=error&status=bogus&window=90d&pageSize=100abc&page=2abc&source=Amie",
    );
    expect(parseFilters(params, ["amie"])).toEqual({
      ...DEFAULT_FILTERS,
      status: ["error"],
    });
  });

  // Make sure every source GET /audit/sources lists filters the list, new connectors included.
  it("keeps any source the server lists", () => {
    const params = new URLSearchParams("source=email-notifier&source=core&source=gone");
    expect(parseFilters(params, ["core", "email-notifier"]).source).toEqual([
      "email-notifier",
      "core",
    ]);
    expect(parseFilters(params).source).toEqual(["email-notifier", "core", "gone"]);
  });

  it("round-trips every filter and leaves defaults out of the URL", () => {
    const filters: ListFilters = {
      status: ["in_progress", "error"],
      source: ["core"],
      window: "30d",
      q: "jsmith",
      page: 2,
      pageSize: 100,
    };
    const url = serializeFilters(filters);
    expect(url.toString()).toBe(
      "status=error&status=in_progress&source=core&window=30d&q=jsmith&page=2&pageSize=100",
    );
    expect(parseFilters(url)).toEqual({ ...filters, status: ["error", "in_progress"] });
    expect(serializeFilters(DEFAULT_FILTERS).toString()).toBe("");
    expect(hasActiveFilters(DEFAULT_FILTERS)).toBe(false);
    expect(hasActiveFilters(filters)).toBe(true);
  });
});
