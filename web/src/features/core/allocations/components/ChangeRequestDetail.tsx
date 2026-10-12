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

import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import { useCurrentUser } from "@/features/core/identity/queries";
import { isGone, useOnGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { enumLabel, formatSU } from "@/shared/format";
import {
  setSearchParam,
  useCanonicalParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { ROUTE_LABELS, queryUuid } from "@/shared/layout/nav";
import { ErrorState, QueryErrorState } from "@/shared/ui/ErrorState";
import { statusLabel, text } from "@/shared/ui/FormDialog";
import { DetailLoading } from "@/shared/ui/Loading";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { StatusBadge, statusBadgeVariantFromChangeRequest } from "@/shared/ui/StatusBadge";
import { Button } from "@/shared/ui/button";
import { keepFocus, useLeaveTo } from "@/shared/ui/focus";
import { Input } from "@/shared/ui/input";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import * as React from "react";
import {
  useAllocation,
  useChangeRequest,
  useChangeRequestEvent,
  useChangeRequestEvents,
  useCreateChangeRequestEvent,
  useDecideChangeRequest,
  useDeleteChangeRequest,
  useDeleteChangeRequestEvent,
  useLatestChangeRequestEvent,
  useRefreshChangeRequests,
} from "../queries";
import { AllocationName } from "./AllocationName";
import { decisionToast } from "./ChangeRequestApproverQueue";
import { ChangeRequestSubmitDrawer } from "./ChangeRequestSubmitDrawer";
import { MemberName, useNameable } from "./MemberName";

// Lifecycle events the server records, kept apart from the ones people add.
const SYSTEM_EVENTS = ["CREATED", "UPDATED", "PENDING", "APPROVED", "REJECTED", "DELETED"];
const isSystemEvent = (type: string) => SYSTEM_EVENTS.includes(type.trim().toUpperCase());
// The server words a decision "Change request transitioned from PENDING to APPROVED".
const eventText = (description?: string) =>
  description?.replace(
    /^Change request transitioned from (\w+) to (\w+)$/,
    (_, from: string, to: string) => `${enumLabel(from)} → ${enumLabel(to)}`,
  );

// A person's row, left out when this caller can never name them.
function PersonRow({
  label,
  id,
  allocationId,
}: { label: string; id: string; allocationId: string }) {
  if (!useNameable(id, allocationId)) return null;
  return (
    <>
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="text-foreground">
        <MemberName id={id} allocationId={allocationId} />
      </dd>
    </>
  );
}

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
        toastOnSuccess(
          create.mutateAsync({
            body: {
              compute_allocation_change_request_id: changeRequestId,
              event_type: text(form, "type"),
              description: text(form, "description"),
            },
          }),
          "Event added",
          () => formEl.reset(),
        );
      }}
    >
      <Input
        name="type"
        placeholder="Event type, e.g. COMMENT"
        aria-label="Event type"
        maxLength={64}
        required
        onChange={(e) => {
          const type = e.currentTarget.value;
          e.currentTarget.setCustomValidity(
            isSystemEvent(type) ? `${type.trim().toUpperCase()} is recorded by the system.` : "",
          );
        }}
      />
      <Input name="description" placeholder="Description" aria-label="Event description" />
      <Button type="submit" size="sm" pending={create.isPending}>
        Add event
      </Button>
    </form>
  );
}

export function ChangeRequestDetail({ changeRequestId }: ChangeRequestDetailProps) {
  const leaveTo = useLeaveTo();
  const { user } = useCurrentUser();
  const ability = useAbility();
  const query = useChangeRequest(changeRequestId);
  useOnGone(isGone(query.error), useRefreshChangeRequests());
  // Sub-reads wait for the request, so a hidden one does not fan out into more 404s.
  const loadedId = query.data ? changeRequestId : undefined;
  const eventsQuery = useChangeRequestEvents(loadedId);
  const latestQuery = useLatestChangeRequestEvent(loadedId);
  const decideMutation = useDecideChangeRequest();
  const deleteMutation = useDeleteChangeRequest();
  const params = useShallowSearchParams();
  const eventId = queryUuid(params.get("event"));
  useCanonicalParams({ event: eventId });
  const setEventId = (id?: string) => setSearchParam(params, "event", id);
  const event = useChangeRequestEvent(eventId);
  const removeEvent = useDeleteChangeRequestEvent(changeRequestId);
  const closeEvent = () => setEventId();
  const [editing, setEditing] = React.useState(false);
  const allocation = useAllocation(query.data?.compute_allocation_id);
  useBreadcrumbLabel(changeRequestId, allocation.data?.name ?? "Change request");
  const pending = ability.can("write", "Allocation") && query.data?.change_status === "PENDING";
  const header = React.useRef<HTMLElement>(null);
  const headingId = React.useId();
  // A decision made here or elsewhere, or found under an open edit, removes the controls that held
  // focus; the header takes it.
  const wasPending = React.useRef(pending);
  React.useEffect(() => {
    if (wasPending.current && !pending) keepFocus(() => header.current);
    wasPending.current = pending;
    if (!pending) setEditing(false);
  }, [pending]);

  if (query.isLoading) return <DetailLoading title={ROUTE_LABELS["/change-requests"]} />;
  if (query.error) {
    return (
      <QueryErrorState
        error={query.error}
        what="change request"
        onRetry={() => query.refetch()}
        as="h1"
        page
      />
    );
  }
  const request = query.data;
  if (!request) return null;

  const canManage = ability.can("write", "Allocation");
  const latest = latestQuery.data;

  function decide(changeStatus: "APPROVED" | "REJECTED") {
    if (!user?.id) return;
    toastOnSuccess(
      decideMutation.mutateAsync({
        id: changeRequestId,
        change_status: changeStatus,
        approver_id: user.id,
      }),
      decisionToast(changeStatus),
    );
  }

  return (
    <article className="space-y-6">
      <section
        ref={header}
        aria-labelledby={headingId}
        tabIndex={-1}
        className="space-y-3 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <div className="flex items-start justify-between gap-4">
          <h1 id={headingId} className="min-w-0 font-display text-[24px] font-bold leading-tight">
            Change request for{" "}
            <AllocationName id={request.compute_allocation_id} name={allocation.data?.name} />
          </h1>
          <StatusBadge
            variant={statusBadgeVariantFromChangeRequest(request.change_status)}
            label={statusLabel(request.change_status)}
          />
        </div>
        {latest ? (
          <p className="text-xs text-muted-foreground [overflow-wrap:anywhere]">
            Latest event: {enumLabel(latest.event_type)} · <Timestamp iso={latest.timestamp} />
          </p>
        ) : null}
        {canManage ? (
          <div className="flex items-center justify-end gap-2">
            {pending ? (
              <Button
                variant="outline"
                disabled={!allocation.data}
                focusableWhenDisabled
                onClick={() => setEditing(true)}
              >
                Edit
              </Button>
            ) : null}
            <Button
              variant="destructive"
              pending={deleteMutation.isPending}
              onClick={() =>
                confirmToast("Delete this change request and its timeline?", "Delete", () =>
                  toastOnSuccess(
                    deleteMutation.mutateAsync(changeRequestId),
                    "Change request deleted",
                    () => leaveTo("/change-requests"),
                  ),
                )
              }
            >
              Delete
            </Button>
            {pending ? (
              <>
                <Button
                  variant="outline"
                  pending={decideMutation.isPending}
                  disabled={!user?.id}
                  focusableWhenDisabled
                  onClick={() => decide("REJECTED")}
                >
                  Reject
                </Button>
                <Button
                  pending={decideMutation.isPending}
                  disabled={!user?.id}
                  focusableWhenDisabled
                  onClick={() => decide("APPROVED")}
                >
                  Approve
                </Button>
              </>
            ) : null}
          </div>
        ) : null}
      </section>

      <section>
        <dl className="grid gap-x-8 gap-y-3 sm:grid-cols-[max-content_1fr] text-sm">
          <PersonRow
            label="Requester"
            id={request.requester_id}
            allocationId={request.compute_allocation_id}
          />

          <dt className="text-muted-foreground">Submitted</dt>
          <dd className="text-foreground">
            <Timestamp iso={request.timestamp} />
          </dd>

          <dt className="text-muted-foreground">Requested SUs</dt>
          <dd className="tabular-nums text-foreground">{formatSU(request.requested_su_amount)}</dd>

          <dt className="text-muted-foreground">Requested status</dt>
          <dd className="text-foreground">{statusLabel(request.requested_status)}</dd>

          {request.approver_id ? (
            <PersonRow
              label="Reviewer"
              id={request.approver_id}
              allocationId={request.compute_allocation_id}
            />
          ) : null}

          <dt className="text-muted-foreground">Reason</dt>
          <dd className="whitespace-pre-wrap text-foreground [overflow-wrap:anywhere]">
            {request.reason}
          </dd>
        </dl>
      </section>

      <section className="space-y-2">
        <h2 className="text-sm font-semibold text-foreground">Timeline</h2>
        {canManage ? <AddEventForm changeRequestId={changeRequestId} /> : null}
        <section
          aria-label="Timeline"
          tabIndex={-1}
          className="rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {eventsQuery.isLoading ? (
            <output aria-label="Loading" className="block text-sm text-muted-foreground">
              Loading events…
            </output>
          ) : eventsQuery.error ? (
            <ErrorState onRetry={() => eventsQuery.refetch()} />
          ) : !eventsQuery.data?.length ? (
            <p className="text-sm text-muted-foreground">No events recorded.</p>
          ) : (
            <ol className="space-y-2">
              {[...eventsQuery.data]
                .sort((a, b) => Date.parse(b.timestamp) - Date.parse(a.timestamp))
                .map((e) => (
                  <li key={e.id} data-row-key={e.id}>
                    <button
                      type="button"
                      data-drawer-trigger=""
                      data-row-opener=""
                      onClick={() => setEventId(e.id)}
                      className="w-full rounded-md border bg-card p-3 text-left text-sm hover:bg-muted/40"
                    >
                      <div className="flex flex-wrap items-center justify-between gap-x-2">
                        <span className="min-w-0 font-medium [overflow-wrap:anywhere]">
                          {enumLabel(e.event_type)}
                        </span>
                        <Timestamp iso={e.timestamp} className="text-xs text-muted-foreground" />
                      </div>
                      {e.description ? (
                        <p className="mt-1 text-xs text-muted-foreground [overflow-wrap:anywhere]">
                          {eventText(e.description)}
                        </p>
                      ) : null}
                    </button>
                  </li>
                ))}
            </ol>
          )}
        </section>
      </section>
      {canManage && allocation.data ? (
        <ChangeRequestSubmitDrawer
          open={editing && pending}
          onOpenChange={setEditing}
          allocation={allocation.data}
          request={request}
        />
      ) : null}
      <RecordDrawer
        title="Change request event"
        id={eventId}
        onClose={closeEvent}
        query={event}
        owned={(e) => e.compute_allocation_change_request_id === changeRequestId}
        fields={(e) => ({
          "Event ID": <Mono>{e.id}</Mono>,
          Type: enumLabel(e.event_type),
          When: <Timestamp iso={e.timestamp} />,
          Description: <span className="whitespace-pre-wrap">{eventText(e.description)}</span>,
        })}
        remove={
          canManage && event.data && !isSystemEvent(event.data.event_type)
            ? {
                label: "Delete event",
                confirm: "Delete this event from the timeline?",
                isPending: removeEvent.isPending,
                onConfirm: () =>
                  toastOnSuccess(
                    removeEvent.mutateAsync({ path: { id: eventId ?? "" } }),
                    "Event deleted",
                    closeEvent,
                  ),
              }
            : undefined
        }
      />
    </article>
  );
}
