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

import { cn } from "@/lib/utils";
import { ApiError } from "@/shared/api/client";
import { useBreadcrumbGone } from "@/shared/layout/BreadcrumbLabelsProvider";
import { decodeSegment, normalizeRouteId } from "@/shared/layout/nav";
import { Button } from "@/shared/ui/button";
import { focusIsLost, userHasActed } from "@/shared/ui/focus";
import { AlertCircleIcon } from "lucide-react";
import { usePathname } from "next/navigation";
import * as React from "react";
import { ViewParamsDropped } from "./ViewParamsDropped";

export type ErrorStateProps = {
  // A failed load: worded by its status, never by the server's text.
  error?: unknown;
  heading?: string;
  // The portal's own copy, never a server message.
  message?: React.ReactNode;
  onRetry?: () => void;
  retryLabel?: string;
  className?: string;
  // Its level under the surrounding headings; a page-level notice is the page's only heading.
  as?: "h1" | "h2" | "h3";
  // Swapped in for a view whose record vanished: the heading takes focus the swap dropped, never
  // on a fresh document load.
  claimFocus?: boolean;
};

const NOT_PERMITTED = "Your account lacks the privilege this page needs.";

// The copy a failed load shows: unavailable for a missing record, not permitted for a refusal,
// else a retry prompt.
function loadErrorCopy(error: unknown): { heading: string; message: string; retry: boolean } {
  const status = error instanceof ApiError ? error.status : undefined;
  if (status === 404 || status === 400)
    return {
      heading: "Unavailable",
      message: "Not found, or you don't have access.",
      retry: false,
    };
  if (status === 403) return { heading: "Not permitted", message: NOT_PERMITTED, retry: false };
  return {
    heading: "Something went wrong",
    message: "Couldn't load this. Try again.",
    retry: true,
  };
}

export function ErrorState({
  error,
  heading,
  message,
  onRetry,
  retryLabel = "Try again",
  className,
  as: Heading = "h2",
  claimFocus = false,
}: ErrorStateProps) {
  const copy = error === undefined ? undefined : loadErrorCopy(error);
  const retry = copy && !copy.retry ? undefined : onRetry;
  const headingRef = React.useRef<HTMLHeadingElement>(null);
  React.useEffect(() => {
    if (claimFocus && userHasActed() && focusIsLost()) headingRef.current?.focus();
  }, [claimFocus]);
  return (
    <div
      role="alert"
      className={cn(
        // Soft-destructive tint; text stays neutral so contrast on the tint hits AA.
        "flex flex-col items-center justify-center gap-3 rounded-md border border-[color:var(--banner-error-border)] bg-[color:var(--banner-error-bg)] px-6 py-10 text-center",
        className,
      )}
    >
      <AlertCircleIcon
        className="size-8 text-[color:var(--banner-error-icon)]"
        aria-hidden="true"
      />
      <Heading
        ref={headingRef}
        tabIndex={claimFocus ? -1 : undefined}
        className="font-heading text-base font-medium text-foreground"
      >
        {heading ?? copy?.heading ?? "Something went wrong"}
      </Heading>
      {(message ?? copy?.message) ? (
        <p className="max-w-md text-sm text-muted-foreground">{message ?? copy?.message}</p>
      ) : null}
      {retry ? (
        <Button variant="outline" onClick={retry}>
          {retryLabel}
        </Button>
      ) : null}
    </div>
  );
}

// Every page renders this when the caller lacks the privilege it needs, under the page's own
// heading; a title renders that heading for pages whose header sits behind the gate.
export function NotPermitted({ title }: { title?: string }) {
  const card = (
    <>
      <ViewParamsDropped />
      <ErrorState heading="Not permitted" message={NOT_PERMITTED} />
    </>
  );
  return title ? (
    <div className="space-y-6">
      <h1 className="font-display text-[28px] font-bold leading-tight">{title}</h1>
      {card}
    </div>
  ) : (
    card
  );
}

// The API answers 404 for deleted records and records outside the caller's scope, 400 for a malformed ID.
// A detail page passes page and as="h1": the notice replaces its view, whose params are dropped.
// Elsewhere (a drawer, a dialog) the record is gone and the notice offers nothing to do.
export function QueryErrorState({
  error,
  what,
  onRetry,
  as = "h2",
  page = false,
}: {
  error: unknown;
  what: string;
  onRetry: () => void;
  as?: ErrorStateProps["as"];
  page?: boolean;
}) {
  return error instanceof ApiError && (error.status === 404 || error.status === 400) ? (
    <>
      {page ? (
        <>
          <ViewParamsDropped />
          <GoneCrumb />
        </>
      ) : null}
      <ErrorState
        heading={page ? `This ${what} is unavailable` : `This ${what} no longer exists`}
        message={page ? "Not found, or you don't have access." : undefined}
        as={as}
        claimFocus
      />
    </>
  ) : (
    <ErrorState error={error} onRetry={onRetry} as={as} />
  );
}

function GoneCrumb() {
  const last = usePathname()?.split("/").at(-1) ?? "";
  useBreadcrumbGone(normalizeRouteId(last) ?? decodeSegment(last));
  return null;
}
