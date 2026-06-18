import { apiFetch } from "@/shared/api/client";
import { auth } from "@/shared/auth/auth";
import { privilegesResponseSchema } from "./schemas";
import type { CurrentUser, Privilege } from "./types";

export async function getCurrentUser(): Promise<CurrentUser | null> {
  const session = await auth();
  if (!session?.user) return null;
  return {
    id: session.user.id ?? session.user.email ?? "",
    email: session.user.email ?? "",
    name: session.user.name ?? session.user.email ?? "",
    privileges: session.privileges ?? [],
  };
}

export async function getPrivileges(): Promise<Privilege[]> {
  const raw = await apiFetch("/user/privileges");
  return privilegesResponseSchema.parse(raw);
}
