import type React from "react";

type Props = {
  title: string;
  description?: string;
  children?: React.ReactNode;
};

// Shared stub for nav entries that don't yet have a real backend or UI.
// Kept intentionally minimal — no fake datasets or invented behavior.
export function PlaceholderPage({ title, description, children }: Props) {
  return (
    <section className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-[30px] font-semibold leading-tight tracking-normal">
          {title}
        </h1>
        {description && (
          <p className="mt-4 text-lg leading-snug text-neutral-700">
            {description}
          </p>
        )}
      </div>

      <div className="h-px bg-neutral-300" />

      <div className="rounded-lg border border-dashed border-neutral-300 bg-white p-8 text-center text-sm text-neutral-500">
        {children ?? (
          <>
            This area is a placeholder. SSH certificate management is available
            under <span className="font-medium text-neutral-700">SSH Certificates</span>.
          </>
        )}
      </div>
    </section>
  );
}
