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

import { useRouter } from "next/navigation";
import * as React from "react";
import { toast } from "sonner";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import {
  useAgreement,
  useDeleteAgreement,
  usePublishAgreement,
  useRemoveAgreement,
} from "../queries";
import type { AgreementDetail } from "../schemas";
import { AgreementAcceptances } from "./AgreementAcceptances";
import { ACTION_COPY, type AgreementAction, AgreementActionDialog } from "./AgreementActionDialog";
import { AgreementBody } from "./AgreementBody";
import { AgreementDetailHeader } from "./AgreementDetailHeader";

function Overview({ agreement }: { agreement: AgreementDetail }) {
  return (
    <div className="space-y-6">
      {agreement.summary ? (
        <section>
          <h2 className="mb-1 text-sm font-semibold">Summary</h2>
          <p className="text-sm text-muted-foreground">{agreement.summary}</p>
        </section>
      ) : null}
      {agreement.change_note ? (
        <section>
          <h2 className="mb-1 text-sm font-semibold">What changed</h2>
          <p className="text-sm text-muted-foreground">{agreement.change_note}</p>
        </section>
      ) : null}
      <section className="rounded-lg border border-border bg-card p-6">
        <AgreementBody markdown={agreement.body_md} />
      </section>
    </div>
  );
}

export function AgreementDetailView({ id }: { id: string }) {
  const router = useRouter();
  const ability = useAbility();
  const query = useAgreement(id);
  const mutations = {
    publish: usePublishAgreement(),
    remove: useRemoveAgreement(),
    delete: useDeleteAgreement(),
  };
  const [action, setAction] = React.useState<AgreementAction | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const agreement = query.data;
  useBreadcrumbLabel(id, agreement ? `${agreement.title} ${agreement.version_label}` : undefined);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error)
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  if (!agreement) return null;

  async function confirm() {
    if (!action || !agreement) return;
    setError(null);
    try {
      await mutations[action].mutateAsync(agreement.id);
      toast.success(ACTION_COPY[action](agreement.version_label).done);
      setAction(null);
      if (action === "delete") router.push("/admin/agreements");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    }
  }

  return (
    <section className="space-y-6">
      <AgreementDetailHeader
        agreement={agreement}
        canManage={ability.can("manage", "Agreement")}
        onAction={(next) => {
          setError(null);
          setAction(next);
        }}
      />
      <TabsRouter
        defaultValue="overview"
        tabs={[
          { value: "overview", label: "Overview", content: <Overview agreement={agreement} /> },
          {
            value: "acceptances",
            label: "Acceptances",
            content: <AgreementAcceptances agreement={agreement} />,
          },
        ]}
      />
      <AgreementActionDialog
        action={action}
        version={agreement.version_label}
        isPending={action ? mutations[action].isPending : false}
        error={error}
        onCancel={() => setAction(null)}
        onConfirm={confirm}
      />
    </section>
  );
}
