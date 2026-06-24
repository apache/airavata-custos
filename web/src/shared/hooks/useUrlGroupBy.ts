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

"use client";

import { useRouter, useSearchParams } from "next/navigation";
import * as React from "react";

// `?gb=a,b,c` <-> string[] page-state. Returns the full list because one page
// composes multiple GroupByChips; each chip's onChange is spliced into its slot.
export function useUrlGroupBy() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const groupBy = React.useMemo<string[]>(() => {
    const raw = searchParams.get("gb");
    if (!raw) return [];
    return raw
      .split(",")
      .map((v) => v.trim())
      .filter((v) => v.length > 0);
  }, [searchParams]);

  const setGroupBy = React.useCallback(
    (next: string[]) => {
      const params = new URLSearchParams(searchParams.toString());
      const cleaned = next.map((v) => v.trim()).filter((v) => v.length > 0);
      if (cleaned.length === 0) {
        params.delete("gb");
      } else {
        params.set("gb", cleaned.join(","));
      }
      const query = params.toString();
      router.replace(query ? `?${query}` : "?", { scroll: false });
    },
    [router, searchParams],
  );

  return { groupBy, setGroupBy };
}
