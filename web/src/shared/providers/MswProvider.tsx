"use client";

import { type ReactNode, useEffect, useState } from "react";

export function MswProvider({ children }: { children: ReactNode }) {
  const enabled = process.env.NEXT_PUBLIC_PORTAL_USE_MSW === "true";
  const [ready, setReady] = useState(!enabled);

  useEffect(() => {
    if (!enabled) return;
    let cancelled = false;
    (async () => {
      const { worker } = await import("@/mocks/browser");
      await worker.start({
        onUnhandledRequest: "bypass",
        serviceWorker: { url: "/mockServiceWorker.js" },
      });
      if (!cancelled) setReady(true);
    })().catch((err) => {
      console.error("MSW worker failed to start", err);
      if (!cancelled) setReady(true);
    });
    return () => {
      cancelled = true;
    };
  }, [enabled]);

  if (!ready) return null;
  return children;
}
