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

import { AbilityProvider } from "@/shared/casl/AbilityProvider";
import { ConfirmToaster, NotificationToaster } from "@/shared/ui/sonner";
import { TooltipProvider } from "@/shared/ui/tooltip";
import { SessionProvider } from "next-auth/react";
import { ThemeProvider } from "next-themes";
import type { ReactNode } from "react";
import { QueryProvider } from "./QueryProvider";

const THEME_KEY = "custos.theme";
const THEMES = ["light", "dark", "system"];

// An unknown stored theme would reach classList.add and throw; it is dropped before the provider reads it.
try {
  const stored = globalThis.localStorage?.getItem(THEME_KEY);
  if (stored && !THEMES.includes(stored)) localStorage.removeItem(THEME_KEY);
} catch {}

export function Providers({ children }: { children: ReactNode }) {
  return (
    <SessionProvider>
      <ThemeProvider
        attribute="class"
        defaultTheme="light"
        enableSystem
        storageKey={THEME_KEY}
        disableTransitionOnChange
      >
        <QueryProvider>
          <AbilityProvider>
            <TooltipProvider>
              {children}
              <NotificationToaster />
              <ConfirmToaster />
            </TooltipProvider>
          </AbilityProvider>
        </QueryProvider>
      </ThemeProvider>
    </SessionProvider>
  );
}
