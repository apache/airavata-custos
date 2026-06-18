"use client";

import Link from "next/link";
import * as React from "react";
import { toast } from "sonner";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import {
  StatusBadge,
  statusBadgeVariantFromChangeRequest,
} from "@/shared/ui/StatusBadge";
import { useCurrentUser } from "@/features/core/identity/queries";
import {
  useApproveChangeRequest,
  useChangeRequest,
  useChangeRequestEvents,
  useRejectChangeRequest,
} from "../queries";

export type ChangeRequestDetailProps = {
  changeRequestId: string;
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function formatSU(n: number): string {
  return new Intl.NumberFormat().format(n);
}

export function ChangeRequestDetail({ changeRequestId }: ChangeRequestDetailProps) {
  const { user } = useCurrentUser();
  const ability = useAbility();
  const query = useChangeRequest(changeRequestId);
  const eventsQuery = useChangeRequestEvents(changeRequestId);
  const approveMutation = useApproveChangeRequest();
  const rejectMutation = useRejectChangeRequest();
  const [pending, setPending] = React.useState(false);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    return (
      <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />
    );
  }
  const request = query.data;
  if (!request) return null;

  const canApprove = ability.can("manage", "Allocation") && request.change_status === "PENDING";

  async function decide(kind: "approve" | "reject") {
    if (!request) return;
    setPending(true);
    try {
      if (kind === "approve") {
        await approveMutation.mutateAsync({ id: request.id, approverId: user?.id ?? "anonymous" });
        toast.success("Approved");
      } else {
        await rejectMutation.mutateAsync({ id: request.id, approverId: user?.id ?? "anonymous" });
        toast.success("Rejected");
      }
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setPending(false);
    }
  }

  return (
    <article className="space-y-6">
      <header className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h1 className="font-display text-[24px] font-bold leading-tight">
              Change request {request.id}
            </h1>
            <Link
              href={`/allocations/${request.compute_allocation_id}`}
              className="text-sm text-muted-foreground hover:underline"
            >
              For allocation {request.compute_allocation_id}
            </Link>
          </div>
          <StatusBadge
            variant={statusBadgeVariantFromChangeRequest(request.change_status)}
            label={request.change_status}
          />
        </div>
        {canApprove ? (
          <div className="flex justify-end gap-2">
            <Button variant="outline" disabled={pending} onClick={() => decide("reject")}>
              Reject
            </Button>
            <Button disabled={pending} onClick={() => decide("approve")}>
              Approve
            </Button>
          </div>
        ) : null}
      </header>

      <section>
        <dl className="grid gap-x-8 gap-y-3 sm:grid-cols-[max-content_1fr] text-sm">
          <dt className="text-muted-foreground">Requester</dt>
          <dd className="font-mono text-foreground">{request.requester_id}</dd>

          <dt className="text-muted-foreground">Submitted</dt>
          <dd className="text-foreground">{formatDate(request.timestamp)}</dd>

          <dt className="text-muted-foreground">Requested SUs</dt>
          <dd className="tabular-nums text-foreground">{formatSU(request.requested_su_amount)}</dd>

          <dt className="text-muted-foreground">Requested status</dt>
          <dd className="text-foreground">{request.requested_status}</dd>

          {request.approver_id ? (
            <>
              <dt className="text-muted-foreground">Reviewer</dt>
              <dd className="font-mono text-foreground">{request.approver_id}</dd>
            </>
          ) : null}

          <dt className="text-muted-foreground">Reason</dt>
          <dd className="whitespace-pre-wrap text-foreground">{request.reason}</dd>
        </dl>
      </section>

      <section className="space-y-2">
        <h2 className="text-sm font-semibold text-foreground">Timeline</h2>
        {eventsQuery.isLoading ? (
          <p className="text-sm text-muted-foreground">Loading events…</p>
        ) : eventsQuery.error ? (
          <ErrorState
            message={(eventsQuery.error as Error).message}
            onRetry={() => eventsQuery.refetch()}
          />
        ) : !eventsQuery.data || eventsQuery.data.length === 0 ? (
          <p className="text-sm text-muted-foreground">No events recorded.</p>
        ) : (
          <ol className="space-y-2">
            {eventsQuery.data.map((e) => (
              <li key={e.id} className="rounded-md border bg-card p-3 text-sm">
                <div className="flex items-center justify-between">
                  <span className="font-medium">{e.event_type}</span>
                  <span className="text-xs text-muted-foreground">{formatDate(e.timestamp)}</span>
                </div>
                {e.description ? (
                  <p className="mt-1 text-xs text-muted-foreground">{e.description}</p>
                ) : null}
              </li>
            ))}
          </ol>
        )}
      </section>
    </article>
  );
}
