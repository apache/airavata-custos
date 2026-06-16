import type { ReactNode } from "react";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

export function PortalLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex h-screen overflow-hidden bg-background text-foreground">
      <a
        href="#main-content"
        className="-translate-y-16 focus:translate-y-0 absolute z-50 m-2 rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground transition-transform focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        Skip to main content
      </a>
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <Topbar />
        <main id="main-content" className="flex-1 overflow-y-auto px-10 py-8" tabIndex={-1}>
          {children}
        </main>
      </div>
    </div>
  );
}
