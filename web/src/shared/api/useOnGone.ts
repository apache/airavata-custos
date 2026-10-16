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

import * as React from "react";
import { ApiError } from "./client";

// A record read the server answered 404: the record is gone, and retrying cannot help.
export const isGone = (error: unknown) => error instanceof ApiError && error.status === 404;

// Calls onGone once each time a record read turns up gone, so the list that showed the record
// refreshes; detail pages and useFreshOnOpen both refresh their parent list through it.
export function useOnGone(gone: boolean, onGone?: () => void) {
  // biome-ignore lint/correctness/useExhaustiveDependencies: fires once per disappearance
  React.useEffect(() => {
    if (gone) onGone?.();
  }, [gone]);
}
