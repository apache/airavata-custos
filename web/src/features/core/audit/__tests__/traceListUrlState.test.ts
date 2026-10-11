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
  statusFiltersToApi,
} from "@/features/core/audit/components/traceListUrlState";
import { describe, expect, it } from "vitest";

describe("traceListUrlState", () => {
  // Make sure a pasted or edited URL cannot widen the query: unknown status,
  // window, or page size values fall back to the defaults.
  it("drops values the list does not know", () => {
    const params = new URLSearchParams(
      "status=failed&status=bogus&window=90d&pageSize=7&page=0&source=Amie",
    );
    expect(parseFilters(params)).toEqual({
      ...DEFAULT_FILTERS,
      status: ["failed"],
    });
  });

  it("round-trips every filter and leaves defaults out of the URL", () => {
    const filters: ListFilters = {
      status: ["retrying", "failed"],
      source: ["core"],
      window: "30d",
      q: "jsmith",
      page: 2,
      pageSize: 100,
    };
    const url = serializeFilters(filters);
    expect(url.toString()).toBe(
      "status=failed&status=retrying&source=core&window=30d&q=jsmith&page=2&pageSize=100",
    );
    expect(parseFilters(url)).toEqual({ ...filters, status: ["failed", "retrying"] });
    expect(serializeFilters(DEFAULT_FILTERS).toString()).toBe("");
    expect(hasActiveFilters(DEFAULT_FILTERS)).toBe(false);
    expect(hasActiveFilters(filters)).toBe(true);
  });

  // Make sure Retrying and Waiting both ask the backend for in_progress and
  // the page keeps only the one that was picked.
  it("maps the status words onto the wire statuses", () => {
    const { apiStatus, keep } = statusFiltersToApi(["retrying", "done"]);
    expect(apiStatus).toEqual(["in_progress", "ok"]);
    expect(keep("retrying")).toBe(true);
    expect(keep("waiting")).toBe(false);
    expect(statusFiltersToApi([]).keep("waiting")).toBe(true);
  });
});
