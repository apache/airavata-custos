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
import { Button, buttonVariants } from "@/shared/ui/button";
import { formatDate } from "../format";
import type { Actor, AgreementDetail } from "../schemas";
import type { AgreementAction } from "./AgreementActionDialog";
import { AgreementStatusBadge } from "./AgreementStatusBadge";

export type AgreementDetailHeaderProps = {
  agreement: AgreementDetail;
  canManage: boolean;
  onAction: (action: AgreementAction) => void;
};

function ActorLine({
  label,
  actor,
  at,
}: {
  label: string;
  actor?: Actor | null | undefined;
  at?: string | undefined;
}) {
  if (!at) return null;
  const who = actor ? [actor.name || actor.id, actor.email].filter(Boolean) : ["Unknown user"];
  return (
    <p className="text-sm text-muted-foreground">
      {label} {[...who, formatDate(at)].join(" · ")}
    </p>
  );
}

export function AgreementDetailHeader({
  agreement,
  canManage,
  onAction,
}: AgreementDetailHeaderProps) {
  const { status } = agreement;
  return (
    <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div className="space-y-2">
        <h1 className="font-display text-[28px] font-bold leading-tight">{agreement.title}</h1>
        <div className="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
          <span>Version {agreement.version_label}</span>
          <AgreementStatusBadge status={status} />
          <span>
            Published {agreement.published_at ? formatDate(agreement.published_at) : "Not yet"}
          </span>
        </div>
        <ActorLine label="Created by" actor={agreement.created_by} at={agreement.created_at} />
        <ActorLine
          label="Published by"
          actor={agreement.published_by}
          at={agreement.published_at}
        />
        <ActorLine label="Removed by" actor={agreement.removed_by} at={agreement.removed_at} />
      </div>

      {canManage && status === "DRAFT" ? (
        <div className="flex flex-wrap gap-2">
          <Link
            href={`/admin/agreements/${agreement.id}/edit`}
            className={buttonVariants({ variant: "outline" })}
          >
            Edit
          </Link>
          <Button variant="destructive" onClick={() => onAction("delete")}>
            Delete draft
          </Button>
          <Button onClick={() => onAction("publish")}>Publish</Button>
        </div>
      ) : null}
      {canManage && status === "PUBLISHED" ? (
        <Button variant="destructive" onClick={() => onAction("remove")}>
          Remove
        </Button>
      ) : null}
    </header>
  );
}
