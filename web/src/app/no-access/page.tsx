import Link from "next/link";

export const metadata = { title: "No access · Custos Portal" };

export default function NoAccessPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-sm">
        <h1 className="text-xl font-semibold tracking-tight">No portal access</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Your account has no privileges yet. Ask an administrator to grant you a role.
        </p>
        <Link href="/sign-in" className="mt-6 inline-block text-sm font-medium text-brand">
          Sign in as a different user
        </Link>
      </div>
    </div>
  );
}
