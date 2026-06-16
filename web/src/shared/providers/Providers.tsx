"use client";

import { SessionProvider } from "next-auth/react";
import { ThemeProvider } from "next-themes";
import type { ReactNode } from "react";
import { LastTraceProvider } from "@/features/core/audit/components/LastTraceProvider";
import { AbilityProvider } from "@/shared/casl/AbilityProvider";
import { Toaster } from "@/shared/ui/sonner";
import { TooltipProvider } from "@/shared/ui/tooltip";
import { MswProvider } from "./MswProvider";
import { QueryProvider } from "./QueryProvider";

export function Providers({ children }: { children: ReactNode }) {
  return (
    <SessionProvider>
      <ThemeProvider
        attribute="class"
        defaultTheme="light"
        enableSystem={false}
        storageKey="custos.theme"
        disableTransitionOnChange
      >
        <MswProvider>
          <QueryProvider>
            <AbilityProvider>
              <LastTraceProvider>
                <TooltipProvider>
                  {children}
                  <Toaster position="top-right" richColors />
                </TooltipProvider>
              </LastTraceProvider>
            </AbilityProvider>
          </QueryProvider>
        </MswProvider>
      </ThemeProvider>
    </SessionProvider>
  );
}
