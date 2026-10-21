// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import * as React from "react";
import { toast } from "sonner";
import { ApiError } from "@/shared/api/client";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { Button, buttonVariants } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { CardSkeleton } from "@/shared/ui/Loading";
import { useAgreement, useAgreements, useCreateAgreement, useUpdateAgreement } from "../queries";
import type { AgreementDraft } from "../schemas";
import { AgreementBody } from "./AgreementBody";

const TEXTAREA =
  "w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50";

const EMPTY: AgreementDraft = {
  title: "",
  version_label: "",
  summary: "",
  body_md: "",
  change_note: "",
};

function Field({
  id,
  label,
  help,
  error,
  children,
}: {
  id: string;
  label: string;
  help?: string;
  error?: string | null;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      {children}
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
      {help ? <p className="text-xs text-muted-foreground">{help}</p> : null}
    </div>
  );
}

function EditorForm({ id, initial }: { id?: string; initial: AgreementDraft }) {
  const router = useRouter();
  const create = useCreateAgreement();
  const update = useUpdateAgreement();
  const [form, setForm] = React.useState(initial);
  const [versionError, setVersionError] = React.useState<string | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const isPending = create.isPending || update.isPending;
  const canSave = Boolean(form.title.trim() && form.version_label.trim() && form.body_md.trim());
  const cancelHref = id ? `/admin/agreements/${id}` : "/admin/agreements";

  const set = (key: keyof AgreementDraft) => (e: { target: { value: string } }) =>
    setForm((prev) => ({ ...prev, [key]: e.target.value }));

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setVersionError(null);
    setError(null);
    try {
      const saved = id ? await update.mutateAsync({ id, ...form }) : await create.mutateAsync(form);
      toast.success(`Draft ${saved.version_label} saved.`);
      router.push(`/admin/agreements/${saved.id}`);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setVersionError(`Version ${form.version_label} already exists.`);
      } else {
        setError(err instanceof Error ? err.message : "Request failed");
      }
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="font-display text-[28px] font-bold">{id ? "Edit draft" : "New version"}</h1>
        <div className="flex gap-2">
          <Link href={cancelHref} className={buttonVariants({ variant: "outline" })}>
            Cancel
          </Link>
          <Button type="submit" disabled={!canSave || isPending}>
            Save draft
          </Button>
        </div>
      </header>

      <section className="grid gap-4 rounded-lg border border-border bg-card p-6 md:grid-cols-2">
        <Field id="agr-title" label="Title">
          <Input id="agr-title" value={form.title} onChange={set("title")} />
        </Field>
        <Field id="agr-version" label="Version label" help="Must be new." error={versionError}>
          <Input
            id="agr-version"
            value={form.version_label}
            onChange={set("version_label")}
            aria-invalid={Boolean(versionError)}
          />
        </Field>
        <Field
          id="agr-summary"
          label="Summary"
          help="Shown in the admin list and the SSH login prompt. Not shown on the consent screen."
        >
          <textarea
            id="agr-summary"
            rows={3}
            className={TEXTAREA}
            value={form.summary}
            onChange={set("summary")}
          />
        </Field>
        <Field
          id="agr-change-note"
          label="Change note (optional)"
          help="Shown on the consent screen to users who accepted an earlier version."
        >
          <textarea
            id="agr-change-note"
            rows={3}
            className={TEXTAREA}
            value={form.change_note}
            onChange={set("change_note")}
          />
        </Field>
      </section>

      <section className="grid gap-4 lg:grid-cols-2">
        <Field id="agr-body" label="Body (Markdown)">
          <textarea
            id="agr-body"
            rows={24}
            className={`${TEXTAREA} font-mono`}
            value={form.body_md}
            onChange={set("body_md")}
          />
        </Field>
        <div className="space-y-2">
          <h2 className="text-sm font-medium">Preview</h2>
          <div className="rounded-lg border border-border bg-card p-6">
            <AgreementBody markdown={form.body_md} />
          </div>
        </div>
      </section>

      {error ? <p className="text-sm text-destructive">{error}</p> : null}
    </form>
  );
}

function EditDraft({ id }: { id: string }) {
  const query = useAgreement(id);
  const agreement = query.data;
  useBreadcrumbLabel(id, agreement ? `${agreement.title} ${agreement.version_label}` : undefined);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error)
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  if (!agreement) return null;
  if (agreement.status !== "DRAFT") return <ErrorState message="Only drafts can be edited." />;
  const { title, version_label, summary, body_md, change_note } = agreement;
  return <EditorForm id={id} initial={{ title, version_label, summary, body_md, change_note }} />;
}

// A new version starts from the published text so admins edit rather than retype it.
function NewVersion() {
  const list = useAgreements();
  if (list.isLoading) return <CardSkeleton />;
  const published = list.data?.find((a) => a.status === "PUBLISHED");
  return (
    <EditorForm
      initial={{ ...EMPTY, title: published?.title ?? "", body_md: published?.body_md ?? "" }}
    />
  );
}

export function AgreementEditor({ id }: { id?: string }) {
  return id ? <EditDraft id={id} /> : <NewVersion />;
}
