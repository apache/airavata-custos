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

import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState } from "react";
import {
  accessRequestKeys,
  useAccessEvent,
  useCreateAccessRequest,
  useMyAccessRequest,
} from "@/features/core/access-requests/queries";
import type { AccessRequest } from "@/features/core/access-requests/schemas";
import { ApiError } from "@/shared/api/client";
import { useDebounce } from "@/shared/hooks/useDebounce";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

// Four requester states driven by GET /access-requests/me:
// no request yet → entry, then the form; PENDING → received; DENIED →
// declined with re-request. APPROVED shows a re-sign-in hint because the
// session's privileges are minted at sign-in.
export function NoAccessBody() {
  const { data: request, isLoading, error } = useMyAccessRequest();
  const [formOpen, setFormOpen] = useState(false);

  let content: React.ReactNode;
  if (isLoading) {
    content = <p className="text-sm text-muted-foreground">Loading…</p>;
  } else if (error) {
    content = (
      <Card heading="No portal access">
        Your account has no privileges yet. Ask an administrator to grant you a role.
      </Card>
    );
  } else if (request?.status === "PENDING") {
    content = <PendingCard request={request} />;
  } else if (request?.status === "APPROVED") {
    content = (
      <Card heading="Request approved">
        Your trial access is ready. Sign out and sign back in to enter the portal.
      </Card>
    );
  } else if (request?.status === "DENIED" && !formOpen) {
    content = <DeclinedCard request={request} onReapply={() => setFormOpen(true)} />;
  } else if (formOpen) {
    content = <RequestForm />;
  } else {
    content = (
      <Card heading="No portal access">
        <span className="block">
          Your identity isn&apos;t linked to a portal user yet. If you are attending a supported
          event, you can request temporary trial access with your event code.
        </span>
        <Button variant="brand" className="mt-4" onClick={() => setFormOpen(true)}>
          Request trial access
        </Button>
      </Card>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-8 shadow-sm">
        {content}
        <Link href="/sign-in" className="mt-6 inline-block text-sm font-medium text-brand">
          Sign in as a different user
        </Link>
      </div>
    </div>
  );
}

function Card({ heading, children }: { heading: string; children: React.ReactNode }) {
  return (
    <div>
      <h1 className="text-xl font-semibold tracking-tight">{heading}</h1>
      <div className="mt-2 text-sm text-muted-foreground">{children}</div>
    </div>
  );
}

function PendingCard({ request }: { request: AccessRequest }) {
  const { data: event } = useAccessEvent(request.event_code);
  return (
    <Card heading="Request received">
      <span className="block">
        Your request for {event?.name ?? request.event_code} is waiting for review. Submitted{" "}
        {formatDate(request.timestamp)}.
      </span>
      <span className="mt-2 block">
        Once it's approved you'll get an email with your account details, so check your inbox.
      </span>
    </Card>
  );
}

function DeclinedCard({ request, onReapply }: { request: AccessRequest; onReapply: () => void }) {
  return (
    <Card heading="Request declined">
      <span className="block">Your access request was declined.</span>
      {request.deny_reason ? (
        <span className="mt-2 block rounded-md bg-muted px-3 py-2 text-foreground">
          {request.deny_reason}
        </span>
      ) : null}
      <Button variant="brand" className="mt-4" onClick={onReapply}>
        Submit a new request
      </Button>
    </Card>
  );
}

function RequestForm() {
  const { data: session } = useSession();
  const params = useSearchParams();
  const queryClient = useQueryClient();
  const create = useCreateAccessRequest();

  const [institution, setInstitution] = useState("");
  const [code, setCode] = useState(params.get("event") ?? "");
  const [reason, setReason] = useState("");

  const trimmedCode = code.trim();
  const debouncedCode = useDebounce(trimmedCode, 400);
  const eventQuery = useAccessEvent(debouncedCode);
  const settled = debouncedCode === trimmedCode && trimmedCode.length > 0 && !eventQuery.isFetching;
  const resolved = settled ? (eventQuery.data ?? null) : null;
  const unknownCode = settled && eventQuery.data === null;

  const canSubmit = institution.trim().length > 0 && Boolean(resolved) && !create.isPending;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    create.mutate(
      {
        institution: institution.trim(),
        event_code: trimmedCode,
        reason: reason.trim() || undefined,
      },
      {
        onError: (err) => {
          // A pending request already exists; refetch /me to show it.
          if (err instanceof ApiError && err.status === 409) {
            void queryClient.invalidateQueries({ queryKey: accessRequestKeys.mine() });
          }
        },
      },
    );
  }

  const submitError =
    create.error && !(create.error instanceof ApiError && create.error.status === 409)
      ? create.error.message
      : null;

  return (
    <form onSubmit={handleSubmit} noValidate>
      <h1 className="text-xl font-semibold tracking-tight">Request trial access</h1>
      <div className="mt-4 flex flex-col gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="ar-name">Name</Label>
          <Input id="ar-name" value={session?.user?.name ?? ""} readOnly disabled />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="ar-email">Email</Label>
          <Input id="ar-email" value={session?.user?.email ?? ""} readOnly disabled />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="ar-institution">Institution</Label>
          <Input
            id="ar-institution"
            required
            value={institution}
            onChange={(e) => setInstitution(e.target.value)}
            placeholder="Your university or organization"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="ar-event">Event code</Label>
          <Input
            id="ar-event"
            required
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="e.g. PEARC26"
            aria-invalid={unknownCode || undefined}
          />
          {unknownCode ? (
            <p className="text-xs text-destructive">Unknown event code.</p>
          ) : resolved ? (
            <p className="text-xs text-muted-foreground">Event: {resolved.name}</p>
          ) : null}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="ar-reason">Reason (optional)</Label>
          <textarea
            id="ar-reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="What do you plan to use the trial access for?"
            className="min-h-20 w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
          />
        </div>
        {submitError ? <p className="text-xs text-destructive">{submitError}</p> : null}
        <Button type="submit" variant="brand" disabled={!canSubmit}>
          {create.isPending ? "Submitting…" : "Submit request"}
        </Button>
      </div>
    </form>
  );
}
