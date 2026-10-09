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

import { UserName } from "@/features/core/users/components/UserPicker";
import { AllocationName } from "./AllocationName";
import { formatDateTime, formatNumber } from "@/shared/format";
import { useRouter } from "next/navigation";
import * as React from "react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { CardSkeleton } from "@/shared/ui/Loading";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { StatusBadge, statusBadgeVariantFromChangeRequest } from "@/shared/ui/StatusBadge";
import { text } from "@/shared/ui/FormDialog";
import { useCurrentUser } from "@/features/core/identity/queries";
import {
  useChangeRequest,
  useChangeRequestEvent,
  useChangeRequestEvents,
  useCreateChangeRequestEvent,
  useCustosManaged,
  useDecideChangeRequest,
  useDeleteChangeRequest,
  useDeleteChangeRequestEvent,
  useLatestChangeRequestEvent,
} from "../queries";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";

export type ChangeRequestDetailProps = {
  changeRequestId: string;
};

function AddEventForm({ changeRequestId }: { changeRequestId: string }) {
  const create = useCreateChangeRequestEvent(changeRequestId);
  return (
    <form
      className="flex items-center gap-2"
      onSubmit={(e) => {
        e.preventDefault();
        const formEl = e.currentTarget;
        const form = new FormData(formEl);
        create.mutate(
          {
            body: {
              compute_allocation_change_request_id: changeRequestId,
              event_type: text(form, "type"),
              description: text(form, "description"),
            },
          },
          toastOnSuccess("Event added", () => formEl.reset()),
        );
      }}
    >
      <Input name="type" placeholder="Event type, e.g. COMMENT" aria-label="Event type" required />
      <Input name="description" placeholder="Description" aria-label="Event description" />
      <Button type="submit" size="sm" disabled={create.isPending}>
        Add event
      </Button>
    </form>
  );
}

export function ChangeRequestDetail({ changeRequestId }: ChangeRequestDetailProps) {
  const router = useRouter();
  const { user } = useCurrentUser();
  const ability = useAbility();
  const query = useChangeRequest(changeRequestId);
  const eventsQuery = useChangeRequestEvents(changeRequestId);
  const latestQuery = useLatestChangeRequestEvent(changeRequestId);
  const decideMutation = useDecideChangeRequest();
  const deleteMutation = useDeleteChangeRequest();
  const [eventId, setEventId] = React.useState<string>();
  const event = useChangeRequestEvent(eventId);
  const removeEvent = useDeleteChangeRequestEvent(changeRequestId);
  const closeEvent = () => setEventId(undefined);
  const managed = useCustosManaged(query.data?.compute_allocation_id);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const request = query.data;
  if (!request) return null;

  const canManage = ability.can("write", "Allocation") && managed;
  const canApprove = canManage && request.change_status === "PENDING";
  const latest = latestQuery.data;

  function decide(changeStatus: "APPROVED" | "REJECTED") {
    if (!user?.id) return;
    decideMutation.mutate(
      {
        path: { id: changeRequestId },
        body: { ...request, change_status: changeStatus, approver_id: user.id },
      },
      toastOnSuccess(changeStatus === "APPROVED" ? "Approved" : "Rejected"),
    );
  }

  return (
    <article className="space-y-6">
      <header className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h1 className="font-display text-[24px] font-bold leading-tight">
              Change request {request.id}
            </h1>
            <p className="text-sm text-muted-foreground">
              For allocation{" "}
              <AllocationName id={request.compute_allocation_id} />
            </p>
          </div>
          <StatusBadge
            variant={statusBadgeVariantFromChangeRequest(request.change_status)}
            label={request.change_status}
          />
        </div>
        {latest ? (
          <p className="text-xs text-muted-foreground">
            Latest event: {latest.event_type} · {formatDateTime(latest.timestamp)}
          </p>
        ) : null}
        {canManage ? (
          <div className="flex justify-end gap-2">
            <Button
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() =>
                confirmToast(`Delete change request ${request.id}?`, "Delete", () =>
                  deleteMutation.mutate(changeRequestId, toastOnSuccess("Change request deleted", () => router.push("/change-requests"))),
                )
              }
            >
              Delete
            </Button>
            {canApprove ? (
              <>
                <Button
                  variant="outline"
                  disabled={decideMutation.isPending || !user?.id}
                  onClick={() => decide("REJECTED")}
                >
                  Reject
                </Button>
                <Button
                  disabled={decideMutation.isPending || !user?.id}
                  onClick={() => decide("APPROVED")}
                >
                  Approve
                </Button>
              </>
            ) : null}
          </div>
        ) : null}
      </header>

      <section>
        <dl className="grid gap-x-8 gap-y-3 sm:grid-cols-[max-content_1fr] text-sm">
          <dt className="text-muted-foreground">Requester</dt>
          <dd className="text-foreground">
            <UserName id={request.requester_id} />
          </dd>

          <dt className="text-muted-foreground">Submitted</dt>
          <dd className="text-foreground">{formatDateTime(request.timestamp)}</dd>

          <dt className="text-muted-foreground">Requested SUs</dt>
          <dd className="tabular-nums text-foreground">{formatNumber(request.requested_su_amount)}</dd>

          <dt className="text-muted-foreground">Requested status</dt>
          <dd className="text-foreground">{request.requested_status}</dd>

          {request.approver_id ? (
            <>
              <dt className="text-muted-foreground">Reviewer</dt>
              <dd className="text-foreground">
                <UserName id={request.approver_id} />
              </dd>
            </>
          ) : null}

          <dt className="text-muted-foreground">Reason</dt>
          <dd className="whitespace-pre-wrap text-foreground">{request.reason}</dd>
        </dl>
      </section>

      <section className="space-y-2">
        <h2 className="text-sm font-semibold text-foreground">Timeline</h2>
        {canManage ? <AddEventForm changeRequestId={changeRequestId} /> : null}
        {eventsQuery.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading events…</p>
        ) : eventsQuery.error ? (
          <ErrorState message={eventsQuery.error.message} onRetry={() => eventsQuery.refetch()} />
        ) : !eventsQuery.data || eventsQuery.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events recorded.</p>
        ) : (
          <ol className="space-y-2">
            {eventsQuery.data.map((e) => (
              <li key={e.id}>
                <button
                  type="button"
                  onClick={() => setEventId(e.id)}
                  className="w-full rounded-md border bg-card p-3 text-left text-sm hover:bg-muted/40"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{e.event_type}</span>
                    <span className="text-xs text-muted-foreground">{formatDateTime(e.timestamp)}</span>
                  </div>
                  {e.description ? (
                    <p className="mt-1 text-xs text-muted-foreground">{e.description}</p>
                  ) : null}
                </button>
              </li>
            ))}
          </ol>
        )}
      </section>
      <RecordDrawer
        title="Change request event"
        id={eventId}
        onClose={closeEvent}
        query={event}
        fields={(e) => ({
          "Event ID": <Mono>{e.id}</Mono>,
          Type: e.event_type,
          When: formatDateTime(e.timestamp),
          Description: <span className="whitespace-pre-wrap">{e.description}</span>,
        })}
        remove={
          canManage
            ? {
                label: "Delete event",
                confirm: "Delete this event from the timeline?",
                isPending: removeEvent.isPending,
                onConfirm: () =>
                  removeEvent.mutate(
                    { path: { id: eventId ?? "" } },
                    toastOnSuccess("Event deleted", closeEvent),
                  ),
              }
            : undefined
        }
      />
    </article>
  );
}
