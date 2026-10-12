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

import {
  replaceShallowSearchParams,
  useDeclaredParams,
} from "@/shared/hooks/useShallowSearchParams";
import { decodeSegment, normalizeRouteId } from "@/shared/layout/nav";
import * as React from "react";

// For a page shell: the query keys its view reads; useDeclaredParams drops the rest.
export function DeclaredParams({ keys }: { keys: readonly string[] }) {
  useDeclaredParams(keys);
  return null;
}

// Rendered in place of a page's view (not permitted, unavailable, not found): the view's tab,
// drawer and filter params have nothing to apply to, so the address bar drops them. A route id the
// page never got to normalise (its gate refused first) is normalised here.
export function ViewParamsDropped() {
  React.useEffect(() => {
    const { pathname, search } = window.location;
    const canonical = (segment: string) => {
      const id = normalizeRouteId(segment) ?? "";
      return segment && id !== decodeSegment(segment) ? encodeURIComponent(id) : segment;
    };
    const path = pathname.split("/").map(canonical).join("/");
    if (path !== pathname) window.history.replaceState(window.history.state, "", path + search);
    if (search) replaceShallowSearchParams(new URLSearchParams());
  }, []);
  return null;
}
