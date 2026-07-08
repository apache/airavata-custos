import type { ReactNode } from "react";
import { UsersAdminProvider } from "./UsersAdminContext";

export default function AdminUsersLayout({ children }: { children: ReactNode }) {
  return (
    <UsersAdminProvider>
      <div className="space-y-4">
        <header className="space-y-1 pb-4">
          <h1 className="font-display text-[28px] font-bold leading-tight">
            Users & Permissions
          </h1>
          <p className="text-sm text-muted-foreground">
            Manage user accounts, roles, and the permissions each role grants.
          </p>
        </header>
        {children}
      </div>
    </UsersAdminProvider>
  );
}
