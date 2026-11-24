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

import { Download } from "lucide-react";
import { Button } from "@/shared/ui/button";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { formatDate } from "../format";
import type { PendingAgreement } from "../schemas";
import { AgreementBody } from "./AgreementBody";

export function versionLine(agreement: PendingAgreement) {
  const effective = agreement.published_at
    ? ` · Effective ${formatDate(agreement.published_at)}`
    : "";
  return `Version ${agreement.version_label}${effective}`;
}

function download(agreement: PendingAgreement) {
  const url = URL.createObjectURL(new Blob([agreement.body_md], { type: "text/plain" }));
  const link = document.createElement("a");
  link.href = url;
  link.download = `${agreement.title}-${agreement.version_label}.txt`
    .replace(/\s+/g, "-")
    .toLowerCase();
  link.click();
  URL.revokeObjectURL(url);
}

export function AgreementBlock({
  agreement,
  checked,
  onCheckedChange,
}: {
  agreement: PendingAgreement;
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
}) {
  return (
    <section className="rounded-xl border border-border bg-card shadow-xs">
      <div className="space-y-3 p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 flex-col gap-0.5">
            <h2 className="text-base font-semibold">{agreement.title}</h2>
            <span className="text-xs text-muted-foreground">{versionLine(agreement)}</span>
          </div>
          {checked ? (
            <StatusBadge variant="approved" label="Accepted" />
          ) : (
            <StatusBadge variant="inactive" label="Not yet accepted" />
          )}
        </div>
        <section
          aria-label="Agreement text"
          // biome-ignore lint/a11y/noNoninteractiveTabindex: keyboard users need focus to scroll the text
          tabIndex={0}
          className="max-h-80 overflow-y-auto rounded-lg border border-border bg-muted/50 p-4 outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
        >
          <AgreementBody markdown={agreement.body_md} />
        </section>
        <Button
          type="button"
          variant="link"
          size="sm"
          className="px-0"
          onClick={() => download(agreement)}
        >
          <Download aria-hidden="true" />
          Download as text
        </Button>
      </div>
      <div className="flex min-h-11 items-center gap-3 border-t border-border px-5 py-1">
        <input
          id="agreement-accept"
          type="checkbox"
          className="size-5 shrink-0 rounded border-input accent-primary"
          checked={checked}
          onChange={(e) => onCheckedChange(e.target.checked)}
        />
        <label htmlFor="agreement-accept" className="cursor-pointer py-2.5 leading-snug">
          I have read and agree to the {agreement.title}, version {agreement.version_label}
        </label>
      </div>
    </section>
  );
}
