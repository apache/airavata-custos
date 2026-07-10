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

import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";

// Shadows the browser history stack so the topbar chevrons can render disabled
// when there's nowhere to go. Next.js doesn't surface a usable history cursor,
// so we track a path stack in sessionStorage and bind it to popstate.
export function useNavHistory(): { canGoBack: boolean; canGoForward: boolean } {
  const pathname = usePathname();
  const popped = useRef(false);
  const [state, setState] = useState({ canGoBack: false, canGoForward: false });

  useEffect(() => {
    if (typeof window === "undefined") return;
    const onPop = () => {
      popped.current = true;
    };
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  useEffect(() => {
    if (typeof window === "undefined" || pathname == null) return;
    const KEY_STACK = "custos.nav.stack";
    const KEY_CURSOR = "custos.nav.cursor";
    const raw = sessionStorage.getItem(KEY_STACK);
    let stack: string[] = raw ? (JSON.parse(raw) as string[]) : [];
    let cursor = Number(sessionStorage.getItem(KEY_CURSOR) ?? "-1");

    if (popped.current) {
      popped.current = false;
      if (stack[cursor - 1] === pathname) cursor -= 1;
      else if (stack[cursor + 1] === pathname) cursor += 1;
      else {
        const idx = stack.indexOf(pathname);
        if (idx >= 0) cursor = idx;
      }
    } else if (stack[cursor] !== pathname) {
      stack = stack.slice(0, cursor + 1);
      stack.push(pathname);
      cursor = stack.length - 1;
    }

    sessionStorage.setItem(KEY_STACK, JSON.stringify(stack));
    sessionStorage.setItem(KEY_CURSOR, String(cursor));
    setState({ canGoBack: cursor > 0, canGoForward: cursor < stack.length - 1 });
  }, [pathname]);

  return state;
}
