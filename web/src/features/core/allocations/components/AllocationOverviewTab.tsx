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

import { formatDate, formatNumber } from "@/shared/format";
import { useCluster } from "@/features/core/clusters/queries";
import { useEffectiveRate, useResources } from "@/features/core/resources/queries";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import {
  useAllocationResourceMapping,
  useAllocationResources,
  useAttachResource,
  useDetachResource,
  useUpdateResourceMapping,
} from "../queries";
import type {
  ComputeAllocation,
  ComputeAllocationResource,
} from "@/generated/core/types.gen";
import { Field, FormDialog, named, num, SelectField, text } from "@/shared/ui/FormDialog";
import { ProjectLink } from "./AllocationsList";

export type AllocationOverviewTabProps = {
  allocation: ComputeAllocation;
  canManage: boolean;
};

// The amount and wall-clock time a resource is granted for, as a mapping or a member override.
export function GrantFields({ amount, time }: { amount?: number; time?: number }) {
  return (
    <>
      <Field
        label="Resource amount"
        name="amount"
        type="number"
        min={0}
        required
        defaultValue={amount}
      />
      <Field
        label="Resource time (minutes)"
        name="time"
        type="number"
        min={0}
        required
        defaultValue={time}
      />
    </>
  );
}

function AllocationResourceRow({
  allocationId,
  resource,
  canManage,
}: {
  allocationId: string;
  resource: ComputeAllocationResource;
  canManage: boolean;
}) {
  const rateQuery = useEffectiveRate(resource.id);
  const update = useUpdateResourceMapping(allocationId);
  const detach = useDetachResource(allocationId);
  const mapping = useAllocationResourceMapping(allocationId, resource.id);
  const path = { id: allocationId, resourceId: resource.id ?? "" };
  return (
    <li className="flex items-center justify-between rounded-md border bg-card px-3 py-2 text-sm">
      <div>
        <span className="font-medium">{resource.name}</span>
        <span className="ml-2 text-xs uppercase tracking-wide text-muted-foreground">
          {resource.resource_type}
        </span>
        {rateQuery.data ? (
          <div className="mt-0.5 text-xs text-muted-foreground">
            {formatNumber(rateQuery.data.rate)} SU/unit
          </div>
        ) : null}
      </div>
      <div className="flex items-center gap-2">
        <span className="tabular-nums">
          {mapping.data
            ? `${formatNumber(mapping.data.resource_amount)} · ${formatNumber(mapping.data.resource_time)} min`
            : "—"}
        </span>
        {canManage ? (
          <>
            <FormDialog
              trigger={
                <Button size="sm" variant="outline">
                  Edit
                </Button>
              }
              title={`Edit ${resource.name}`}
              description="Sets the amount and wall-clock time this allocation is granted."
              submitLabel="Save"
              isPending={update.isPending}
              onSubmit={(form, close) =>
                update.mutate(
                  {
                    path,
                    body: { resource_amount: num(form, "amount"), resource_time: num(form, "time") },
                  },
                  toastOnSuccess("Resource updated", close),
                )
              }
            >
              {mapping.isLoading ? (
                <TableSkeleton rows={2} columns={1} />
              ) : (
                <GrantFields
                  amount={mapping.data?.resource_amount}
                  time={mapping.data?.resource_time}
                />
              )}
            </FormDialog>
            <Button
              size="sm"
              variant="destructive"
              disabled={detach.isPending}
              onClick={() =>
                confirmToast(`Detach ${resource.name} from this allocation?`, "Detach", () =>
                  detach.mutate({ path }, toastOnSuccess("Resource detached")),
                )
              }
            >
              Detach
            </Button>
          </>
        ) : null}
      </div>
    </li>
  );
}

function AttachResourceDialog({
  allocation,
  attached,
}: {
  allocation: ComputeAllocation;
  attached: ComputeAllocationResource[];
}) {
  const resources = useResources();
  const attach = useAttachResource(allocation.id ?? "");
  const options = named(
    resources.data?.filter(
      (r) =>
        r.compute_cluster_id === allocation.compute_cluster_id &&
        !attached.some((a) => a.id === r.id),
    ),
  );
  return (
    <FormDialog
      trigger={
        <Button size="sm" disabled={options.length === 0}>
          + Attach resource
        </Button>
      }
      title="Attach resource"
      description="Resources of this allocation's cluster that are not attached yet."
      submitLabel="Attach"
      isPending={attach.isPending}
      onSubmit={(form, close) =>
        attach.mutate(
          {
            path: { id: allocation.id ?? "" },
            body: {
              compute_allocation_resource_id: text(form, "resource"),
              resource_amount: num(form, "amount"),
              resource_time: num(form, "time"),
            },
          },
          toastOnSuccess("Resource attached", close),
        )
      }
    >
      <SelectField label="Resource" name="resource" options={options} />
      <GrantFields />
    </FormDialog>
  );
}

export function AllocationOverviewTab({ allocation, canManage }: AllocationOverviewTabProps) {
  const resourcesQuery = useAllocationResources(allocation.id);
  const cluster = useCluster(allocation.compute_cluster_id).data;

  return (
    <div className="space-y-6">
      <dl className="grid gap-x-8 gap-y-3 rounded-lg border border-border bg-muted/40 p-4 text-sm sm:grid-cols-[max-content_1fr]">
        <dt className="text-muted-foreground">Allocation ID</dt>
        <dd className="font-mono text-foreground before:font-sans before:content-[':_']">
          {allocation.id}
        </dd>

        <dt className="text-muted-foreground">Project</dt>
        <dd className="text-foreground before:content-[':_']">
          <ProjectLink id={allocation.project_id ?? ""} className="hover:underline" />
        </dd>

        <dt className="text-muted-foreground">Name</dt>
        <dd className="text-foreground before:content-[':_']">{allocation.name}</dd>

        <dt className="text-muted-foreground">Status</dt>
        <dd className="text-foreground before:content-[':_']">{allocation.status}</dd>

        <dt className="text-muted-foreground">Cluster</dt>
        <dd className="text-foreground before:content-[':_']">
          {cluster?.name ?? allocation.compute_cluster_id}
        </dd>

        <dt className="text-muted-foreground">Initial SUs</dt>
        <dd className="tabular-nums text-foreground before:content-[':_']">
          {formatNumber(allocation.initial_su_amount)}
        </dd>

        <dt className="text-muted-foreground">Start</dt>
        <dd className="text-foreground before:content-[':_']">{formatDate(allocation.start_time)}</dd>

        <dt className="text-muted-foreground">End</dt>
        <dd className="text-foreground before:content-[':_']">{formatDate(allocation.end_time)}</dd>
      </dl>

      <section className="space-y-2">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-foreground">Resources</h2>
          {canManage ? (
            <AttachResourceDialog allocation={allocation} attached={resourcesQuery.data ?? []} />
          ) : null}
        </div>
        {resourcesQuery.isLoading ? (
          <TableSkeleton rows={2} columns={3} />
        ) : resourcesQuery.error ? (
          <ErrorState
            message={resourcesQuery.error.message}
            onRetry={() => resourcesQuery.refetch()}
          />
        ) : !resourcesQuery.data || resourcesQuery.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">No resources attached.</p>
        ) : (
          <ul className="space-y-1">
            {resourcesQuery.data.map((resource) => (
              <AllocationResourceRow
                key={resource.id}
                allocationId={allocation.id ?? ""}
                resource={resource}
                canManage={canManage}
              />
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
