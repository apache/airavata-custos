import "server-only";
import { auth } from "./auth";
import type { Privilege } from "@/features/core/identity/types";

export type PortalSession = {
  accessToken?: string | null;
  privileges: Privilege[];
} | null;

export async function getPortalSession(): Promise<PortalSession> {
  const session = await auth();
  if (!session) return null;
  return {
    accessToken: session.accessToken ?? null,
    privileges: session.privileges ?? [],
  };
}
