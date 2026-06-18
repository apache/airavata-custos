import "server-only";
import { auth } from "./auth";

export type PortalSession = { accessToken?: string | null; userId?: string | null } | null;

export async function getPortalSession(): Promise<PortalSession> {
  const session = await auth();
  if (!session) return null;
  return {
    accessToken: session.accessToken ?? null,
    userId: session.user?.id ?? null,
  };
}
