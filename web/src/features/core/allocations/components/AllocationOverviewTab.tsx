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

import { useCluster } from "@/features/core/clusters/queries";
import { useProject, useProjects } from "@/features/core/projects/queries";
import { useEffectiveRate, useResources } from "@/features/core/resources/queries";
import type { ComputeAllocation, ComputeAllocationResource } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatRate, formatSU } from "@/shared/format";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Field, FormDialog, SelectField, named, num, text } from "@/shared/ui/FormDialog";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Link } from "@/shared/ui/link";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import * as React from "react";
import { useAllocationResources, useAttachResource, useDetachResource } from "../queries";
import { AllocationStatusLabel, knownDay } from "./AllocationStatusLabel";

export type AllocationOverviewTabProps = {
  allocation: ComputeAllocation;
  canManage: boolean;
  // Rates and the resource catalog require the allocations read privilege.
  canRead: boolean;
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
        max={Number.MAX_SAFE_INTEGER}
        required
        defaultValue={amount}
      />
      <Field
        label="Resource time (minutes)"
        name="time"
        type="number"
        min={0}
        max={Number.MAX_SAFE_INTEGER}
        required
        defaultValue={time}
      />
    </>
  );
}

function ProjectLink({ id, title }: { id: string; title: string | undefined }) {
  if (!title) return <>—</>;
  return (
    <Link href={`/projects/${id}`} className="hover:underline">
      {title}
    </Link>
  );
}

function ReadableProject({ id }: { id: string }) {
  const project = useProject(id);
  if (project.isLoading) return <>—</>;
  return <ProjectLink id={id} title={project.data?.title} />;
}

// Without projects:read only the caller's own projects are readable, and master lists those whole.
function ParticipantProject({ id }: { id: string }) {
  const projects = useProjects({ limit: 200 }, false);
  if (projects.isLoading) return <>—</>;
  return <ProjectLink id={id} title={projects.data?.items?.find((p) => p.id === id)?.title} />;
}

// Rendered for cluster readers only, whom the cluster's name is never withheld from.
function ClusterName({ id }: { id: string }) {
  return <>{useCluster(id).data?.name ?? "—"}</>;
}

function AllocationResourceRow({
  allocationId,
  resource,
  detach,
  canRead,
}: {
  allocationId: string;
  resource: ComputeAllocationResource;
  // Undefined for a caller who cannot detach.
  detach: ReturnType<typeof useDetachResource> | undefined;
  canRead: boolean;
}) {
  const rateQuery = useEffectiveRate(canRead ? resource.id : undefined);
  return (
    <li className="flex flex-wrap items-center justify-between gap-3 rounded-md border bg-card px-3 py-2 text-sm">
      <div className="min-w-0 [overflow-wrap:anywhere]">
        <span className="font-medium">{resource.name}</span>
        {resource.resource_type ? (
          <span className="ml-2 text-xs text-muted-foreground">{resource.resource_type}</span>
        ) : null}
        {rateQuery.data ? (
          <div className="mt-0.5 text-xs text-muted-foreground">
            {formatRate(rateQuery.data.rate)} SU/unit
          </div>
        ) : null}
      </div>
      {detach ? (
        <Button
          size="sm"
          variant="destructive"
          aria-label={`Detach ${resource.name}`}
          pending={detach.isPending && detach.variables?.path.resourceId === resource.id}
          onClick={() =>
            confirmToast(`Detach ${resource.name} from this allocation?`, "Detach", () =>
              toastOnSuccess(
                detach.mutateAsync({ path: { id: allocationId, resourceId: resource.id } }),
                "Resource detached",
              ),
            )
          }
        >
          Detach
        </Button>
      ) : null}
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
  const attach = useAttachResource(allocation.id);
  const ofCluster = resources.data?.filter(
    (r) => r.compute_cluster_id === allocation.compute_cluster_id,
  );
  const options = named(ofCluster?.filter((r) => !attached.some((a) => a.id === r.id)));
  if (resources.error) {
    return (
      <p role="alert" className="text-xs text-destructive-text">
        Couldn't load resources.{" "}
        <Button type="button" variant="link" size="xs" onClick={() => resources.refetch()}>
          Try again
        </Button>
      </p>
    );
  }
  return (
    <div className="flex flex-wrap items-center justify-end gap-2">
      {ofCluster && options.length === 0 ? (
        <p className="text-xs text-muted-foreground">
          {ofCluster.length === 0
            ? "This allocation's cluster has no resources."
            : "Every resource of this allocation's cluster is attached."}
        </p>
      ) : null}
      <FormDialog
        trigger={
          <Button size="sm" disabled={options.length === 0} focusableWhenDisabled>
            + Attach resource
          </Button>
        }
        title="Attach resource"
        description="Resources of this allocation's cluster that are not attached yet."
        submitLabel="Attach"
        isPending={attach.isPending}
        onSubmit={(form, close) =>
          toastOnSuccess(
            attach.mutateAsync({
              path: { id: allocation.id },
              body: {
                compute_allocation_resource_id: text(form, "resource"),
                resource_amount: num(form, "amount"),
                resource_time: num(form, "time"),
              },
            }),
            "Resource attached",
            close,
          )
        }
      >
        <SelectField label="Resource" name="resource" options={options} />
        <GrantFields />
      </FormDialog>
    </div>
  );
}

export function AllocationOverviewTab({
  allocation,
  canManage,
  canRead,
}: AllocationOverviewTabProps) {
  const resourcesQuery = useAllocationResources(allocation.id);
  const detach = useDetachResource(allocation.id);
  const ability = useAbility();
  const canReadProject = ability.can("read", "Project");
  const headingId = React.useId();

  return (
    <div className="space-y-6">
      <dl className="grid gap-x-8 gap-y-3 rounded-lg border border-border bg-muted/40 p-4 text-sm sm:grid-cols-[max-content_minmax(0,1fr)]">
        <dt className="text-muted-foreground">Allocation ID</dt>
        <dd className="font-mono text-foreground before:font-sans sm:before:content-[':_']">
          {allocation.id}
        </dd>

        <dt className="text-muted-foreground">Project</dt>
        <dd className="text-foreground [overflow-wrap:anywhere] sm:before:content-[':_']">
          {canReadProject ? (
            <ReadableProject id={allocation.project_id} />
          ) : (
            <ParticipantProject id={allocation.project_id} />
          )}
        </dd>

        <dt className="text-muted-foreground">Name</dt>
        <dd className="text-foreground [overflow-wrap:anywhere] sm:before:content-[':_']">
          {allocation.name}
        </dd>

        <dt className="text-muted-foreground">Status</dt>
        <dd className="text-foreground sm:before:content-[':_']">
          <AllocationStatusLabel
            status={allocation.status}
            startTime={allocation.start_time}
            endTime={allocation.end_time}
          />
        </dd>

        {ability.can("read", "Cluster") ? (
          <>
            <dt className="text-muted-foreground">Cluster</dt>
            <dd className="text-foreground sm:before:content-[':_']">
              <ClusterName id={allocation.compute_cluster_id} />
            </dd>
          </>
        ) : null}

        <dt className="text-muted-foreground">Initial SUs</dt>
        <dd className="tabular-nums text-foreground sm:before:content-[':_']">
          {formatSU(allocation.initial_su_amount)}
        </dd>

        <dt className="text-muted-foreground">Start</dt>
        <dd className="text-foreground sm:before:content-[':_']">
          {knownDay(allocation.start_time)}
        </dd>

        <dt className="text-muted-foreground">End</dt>
        <dd className="text-foreground sm:before:content-[':_']">
          {knownDay(allocation.end_time)}
        </dd>
      </dl>

      <section
        aria-labelledby={headingId}
        className="space-y-2 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 id={headingId} className="text-sm font-semibold text-foreground">
            Resources
          </h2>
          {canManage && canRead ? (
            <AttachResourceDialog allocation={allocation} attached={resourcesQuery.data ?? []} />
          ) : null}
        </div>
        {resourcesQuery.isLoading ? (
          <TableSkeleton rows={2} columns={3} />
        ) : resourcesQuery.error ? (
          <ErrorState onRetry={() => resourcesQuery.refetch()} />
        ) : !resourcesQuery.data?.length ? (
          <p className="text-sm text-muted-foreground">No resources attached.</p>
        ) : (
          <ul className="space-y-1">
            {resourcesQuery.data.map((resource) => (
              <AllocationResourceRow
                key={resource.id}
                allocationId={allocation.id}
                resource={resource}
                detach={canManage ? detach : undefined}
                canRead={canRead}
              />
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
