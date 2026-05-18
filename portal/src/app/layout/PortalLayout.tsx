"use client";

import type React from "react";
import { useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { signIn, signOut, useSession } from "next-auth/react";
import {
  Bell,
  ChevronLeft,
  ChevronRight,
  Headphones,
  LogOut,
} from "lucide-react";
import { navItems } from "../nav";
import { PORTAL_NAME, SUPPORT_EMAIL } from "../../lib/config";

export function PortalLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { data: session, status } = useSession();
  const loading = status === "loading";
  const user = session?.user;
  const displayName = user?.name ?? user?.email ?? "Signed out";
  const displayEmail = user?.email ?? "no session";
  const avatarInitial = (user?.email ?? displayName).slice(0, 1).toUpperCase();

  useEffect(() => {
    if (status === "unauthenticated") {
      // No provider id: NextAuth picks the registered provider (OIDC in prod,
      // dev Credentials fallback when OIDC_ISSUER_URL is unset).
      signIn();
    }
  }, [status]);

  return (
    <div className="flex min-h-screen bg-white text-neutral-950">
      <aside className="flex w-[230px] shrink-0 flex-col bg-[#f1f1f1]">
        <div className="px-8 pb-6 pt-6">
          <Link
            href="/signer/certificates"
            className="text-2xl font-extrabold uppercase tracking-normal text-blue-800"
          >
            {PORTAL_NAME}
          </Link>
        </div>

        <nav className="space-y-0">
          {navItems.map((item) => {
            const active = pathname === item.href;
            const Icon = item.icon;

            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={active ? "page" : undefined}
                className={`relative flex min-h-12 items-center gap-3 px-7 py-3 text-base transition ${
                  active
                    ? "bg-white text-neutral-950"
                    : "text-neutral-600 hover:bg-white/70 hover:text-neutral-950"
                }`}
              >
                <Icon className="h-5 w-5 stroke-[1.8]" />
                {item.label}
                {active && (
                  <span className="absolute left-0 top-2 h-8 w-1.5 rounded-r-full bg-blue-600" />
                )}
              </Link>
            );
          })}
        </nav>

        <div className="mx-4 mb-5 mt-auto rounded-xl border border-neutral-200 bg-white px-5 py-4 text-center shadow-sm">
          <div className="mx-auto flex h-9 w-9 items-center justify-center rounded-md bg-neutral-100">
            <Headphones className="h-6 w-6" />
          </div>
          <div className="mt-2 text-sm font-bold">Need Help?</div>
          <p className="mt-1 text-sm leading-snug text-neutral-500">
            Get help in our Help Center or contact support.
          </p>
          <a
            className="mt-4 inline-flex h-9 items-center justify-center rounded-lg bg-black px-6 text-sm font-semibold text-white shadow-sm transition hover:bg-neutral-800"
            href={`mailto:${SUPPORT_EMAIL}`}
          >
            Get Support
          </a>
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-20 items-center justify-between border-b border-neutral-200 bg-white px-8">
          <div className="flex items-center gap-3">
            <button
              aria-label="Go back"
              className="flex h-9 w-9 items-center justify-center rounded-full text-neutral-950 hover:bg-neutral-100"
              type="button"
              onClick={() => window.history.back()}
            >
              <ChevronLeft className="h-6 w-6" />
            </button>
            <button
              aria-label="Go forward"
              className="flex h-9 w-9 items-center justify-center rounded-full text-neutral-300 hover:bg-neutral-100 hover:text-neutral-500"
              type="button"
              onClick={() => window.history.forward()}
            >
              <ChevronRight className="h-6 w-6" />
            </button>
            <Link
              href="/signer/certificates"
              className="ml-2 inline-flex h-9 items-center gap-2 rounded-md bg-blue-50 px-3 text-sm font-medium text-blue-700"
            >
              SSH Certificates
              <ChevronRight className="h-4 w-4" />
            </Link>
          </div>

          <div className="flex items-center gap-3">
            <Bell className="h-6 w-6 text-neutral-500" />
            <div className="h-12 w-px bg-neutral-200" />
            <div className="text-left text-xs">
              <div className="font-medium">
                {loading ? "Loading user..." : displayName}
              </div>
              <div className="text-neutral-500">{displayEmail}</div>
            </div>

            <div className="flex h-11 w-11 items-center justify-center rounded-full border-[3px] border-blue-700 bg-neutral-200 text-sm font-semibold text-neutral-700">
              {avatarInitial}
            </div>

            <button
              className="hidden items-center gap-2 rounded-lg border px-3 py-2 text-sm text-neutral-700 hover:bg-neutral-50 xl:inline-flex"
              onClick={() => signOut()}
              type="button"
            >
              <LogOut className="h-4 w-4" />
              Sign out
            </button>
          </div>
        </header>

        <main className="flex-1 px-12 py-8">{children}</main>
      </div>
    </div>
  );
}
