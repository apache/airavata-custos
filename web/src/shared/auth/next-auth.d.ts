import type { Privilege } from "@/features/core/identity/types";

declare module "next-auth" {
  interface Session {
    accessToken?: string | null;
    privileges?: Privilege[];
  }
  interface User {
    privileges?: Privilege[];
  }
}
