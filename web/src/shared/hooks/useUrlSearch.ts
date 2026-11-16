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

// A list page's search box. The text renders as typed; the trimmed query
// settles after a pause and is kept in the URL as `q`, so reloads and shared
// links keep it.

"use client";

import * as React from "react";
import { useDebounce } from "./useDebounce";
import { replaceShallowSearchParam, useShallowSearchParams } from "./useShallowSearchParams";

export function useUrlSearch() {
  const params = useShallowSearchParams();
  const [text, setText] = React.useState(() => params.get("q") ?? "");
  const settledQuery = useDebounce(text.trim(), 300);

  React.useEffect(() => {
    if ((new URLSearchParams(window.location.search).get("q") ?? "") !== settledQuery) {
      replaceShallowSearchParam("q", settledQuery || null);
    }
  }, [settledQuery]);

  return { text, setText, settledQuery };
}
