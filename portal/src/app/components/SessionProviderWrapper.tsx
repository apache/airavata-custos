"use client";

import type React from "react";
import { SessionProvider } from "next-auth/react";

/**
 * Client-component bridge that mounts NextAuth's <SessionProvider> from the
 * server-rendered root layout. NextAuth's provider is a client component
 * and would otherwise pull the entire layout out of server rendering; this
 * thin wrapper isolates the "use client" boundary to a single subtree.
 */
export function SessionProviderWrapper({
  children,
}: {
  children: React.ReactNode;
}) {
  return <SessionProvider>{children}</SessionProvider>;
}
