/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

"use client";

import { useSubscriptions } from "@/features/core/events/queries";
import { leftThisRun, relativeTime } from "@/features/core/events/utils";
import { cn } from "@/lib/utils";
import { enumLabel } from "@/shared/format";
import { buttonVariants } from "@/shared/ui/button";
import { useScrollStop } from "@/shared/ui/focus";
import { ExternalLinkIcon } from "lucide-react";
import Link from "next/link";
import type * as React from "react";
import { tokenLabel } from "../eventLabels";
import type { Step, TraceDelivery } from "../types";
import type { FlowModel, FlowRow } from "../utils";
import {
  deliveryLabel,
  deliveryListStatus,
  errorText,
  formatOffset,
  isCodeShaped,
  parseDetails,
  plainName,
  retriedLabel,
  tryTitle,
} from "../utils";
import { RelatedTracesLink } from "./RelatedTracesLink";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";
import { Timestamp } from "./primitives/Timestamp";

export type StepDetailPanelProps = {
  row: FlowRow | null;
  model: FlowModel;
  traceId: string;
  className?: string;
};

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mt-4">
      <h3 className="mb-1.5 text-[11px] font-bold uppercase tracking-[0.05em] text-muted-foreground">
        {title}
      </h3>
      {children}
    </section>
  );
}

function Fact({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-[30px] items-start gap-3 text-[13px]">
      <span className="w-[88px] shrink-0 pt-[1px] text-muted-foreground">{label}</span>
      <div className="min-w-0 flex-1 text-foreground [overflow-wrap:anywhere]">{children}</div>
    </div>
  );
}

function ErrorBox({ text }: { text: string }) {
  const scrollStop = useScrollStop<HTMLPreElement>();
  return (
    <pre
      data-testid="step-error"
      {...scrollStop}
      aria-label="Error"
      className="mt-3 max-h-40 overflow-auto whitespace-pre-wrap [overflow-wrap:anywhere] rounded-md bg-[color:var(--tone-error-bg)] p-3 font-mono text-xs text-[color:var(--tone-error-fg)]"
    >
      {text}
    </pre>
  );
}

function OpenDeliveryButton({ delivery }: { delivery: TraceDelivery }) {
  return (
    <Link
      href={`/admin/events?delivery=${encodeURIComponent(delivery.id)}`}
      className={cn(buttonVariants({ variant: "outline", size: "sm" }), "mt-4")}
    >
      <ExternalLinkIcon className="size-3.5" aria-hidden="true" />
      Open delivery in Events
    </Link>
  );
}

function Details({ description }: { description: string | undefined }) {
  const parsed = parseDetails(description);
  if (!parsed) return null;
  if ("text" in parsed) {
    return (
      <Section title="Details">
        <pre className="whitespace-pre-wrap [overflow-wrap:anywhere] rounded-md border border-border bg-[color:var(--muted-2)] p-3 font-mono text-xs text-foreground">
          {tokenLabel(parsed.text)}
        </pre>
      </Section>
    );
  }
  return (
    <Section title="Details">
      <dl className="overflow-hidden rounded-md border border-border text-[12.5px]">
        {parsed.entries.map(([k, v], i) => (
          <div
            key={k}
            className={cn(
              "flex items-start gap-3 px-2.5 py-1.5",
              i % 2 === 1 && "bg-[color:var(--muted-2)]",
            )}
          >
            <dt className="w-[120px] shrink-0 font-mono text-[11.5px] text-muted-foreground">
              {k}
            </dt>
            <dd className="min-w-0 flex-1 font-mono [overflow-wrap:anywhere]">
              <CopyValue value={v} label={k} explicit>
                <span className="whitespace-pre-wrap">{v}</span>
              </CopyValue>
            </dd>
          </div>
        ))}
      </dl>
    </Section>
  );
}

function Ids({ step, traceId }: { step?: Step; traceId: string }) {
  return (
    <Section title="Ids">
      {step && (
        <Fact label="Step">
          <CopyValue value={step.id} label="step ID" />
        </Fact>
      )}
      {step && (
        <Fact label="Span">
          <CopyValue value={step.span_id} label="span ID" />
        </Fact>
      )}
      {step?.parent_span_id && (
        <Fact label="Parent span">
          <CopyValue value={step.parent_span_id} label="parent span ID" />
        </Fact>
      )}
      <Fact label="Trace">
        <CopyValue value={traceId} label="trace ID" />
      </Fact>
    </Section>
  );
}

function StepBody({ step, model, traceId }: { step: Step; model: FlowModel; traceId: string }) {
  const name = plainName(step.event_type);
  const title =
    step.event_type === "EVENT_DELIVERY_RETRIED"
      ? retriedLabel(step)
      : (tryTitle(model, step) ?? name ?? step.event_type);
  const err = errorText(step);
  const deliveryId = model.tryOf.get(step.id) ?? model.ownerOfSpan.get(step.span_id);
  const delivery = deliveryId
    ? [
        ...model.hopsBySpan.values(),
        ...model.groups.flatMap((g) => (g.kind === "hop" ? [[g.hop]] : [])),
      ]
        .flat()
        .flatMap((h) => h.deliveries)
        .find((d) => d.delivery.id === deliveryId)?.delivery
    : undefined;
  const offset = Date.parse(step.created_at) - Date.parse(model.startedAt);

  return (
    <>
      <div
        className={cn("text-[15px] font-bold text-foreground", !name && "font-mono text-[13.5px]")}
      >
        {title}
      </div>
      {name && (
        <div className="mt-0.5 font-mono text-[12px] text-muted-foreground">{step.event_type}</div>
      )}
      <div className="mt-2 flex flex-wrap items-center gap-1.5">
        <StatusPill status={step.status} size="sm" />
        <SourcePill source={step.source} size="sm" />
      </div>
      {err && <ErrorBox text={err} />}
      <Section title="Facts">
        <Fact label="Time">
          <Timestamp iso={step.created_at} precise className="tabular-nums" />
          <span className="text-muted-foreground"> ({relativeTime(step.created_at)})</span>
        </Fact>
        <Fact label="Offset">
          <span className="font-mono tabular-nums">{formatOffset(offset)}</span>
        </Fact>
        {step.entity_type && step.entity_id && (
          <Fact label="Entity">
            <span className="text-muted-foreground">{enumLabel(step.entity_type)} </span>
            <CopyValue value={step.entity_id} label={step.entity_type} />
            <div className="mt-1">
              <RelatedTracesLink
                entityId={step.entity_id}
                entityType={step.entity_type}
                className="text-[12.5px] text-foreground underline-offset-4 hover:underline"
              />
            </div>
          </Fact>
        )}
        {delivery && (
          <Fact label="Delivery">
            <span className="font-mono text-[12.5px]">{delivery.subscriber}</span>
            <span className="text-muted-foreground">
              , {deliveryLabel(delivery, model.tryCounts.get(delivery.id) ?? 0).toLowerCase()}
            </span>
          </Fact>
        )}
      </Section>
      <Details description={step.description} />
      {delivery && <OpenDeliveryButton delivery={delivery} />}
      <Ids step={step} traceId={traceId} />
    </>
  );
}

function DeliveryBody({
  delivery,
  tries,
  traceId,
}: {
  delivery: TraceDelivery;
  tries: number;
  traceId: string;
}) {
  const status = deliveryListStatus(delivery);
  const loaded = useSubscriptions().data?.items?.find(
    (s) => s.subscriber === delivery.subscriber && s.event_type === delivery.event_type,
  )?.loaded;
  return (
    <>
      <div className="text-[13.5px] font-bold text-foreground">
        <span className="font-medium text-muted-foreground">Delivery to </span>
        <span className="font-mono">{delivery.subscriber}</span>
      </div>
      <div className="mt-2 flex flex-wrap items-center gap-1.5">
        <StatusPill status={status} label={deliveryLabel(delivery, tries)} size="sm" />
      </div>
      {delivery.last_error && delivery.status !== "SUCCEEDED" && (
        <ErrorBox text={delivery.last_error} />
      )}
      <Section title="Facts">
        <Fact label="Event">
          <span className="font-mono text-[12.5px]">{delivery.event_type}</span>
        </Fact>
        <Fact label="Tries">
          <span className="tabular-nums">{Math.max(tries, delivery.attempts)}</span>
          {delivery.status === "PENDING" && (
            <span className="text-muted-foreground"> ({leftThisRun(delivery.attempts)})</span>
          )}
        </Fact>
        {delivery.status === "PENDING" && (
          <Fact label="Next try">
            <span className="tabular-nums">{relativeTime(delivery.next_run_at)}</span>
            <span className="text-muted-foreground">
              {" "}
              (<Timestamp iso={delivery.next_run_at} className="tabular-nums" />)
            </span>
          </Fact>
        )}
        {delivery.status === "PENDING" && delivery.attempts === 0 && (
          <p className="mt-1 text-[12.5px] text-muted-foreground">
            {loaded === undefined
              ? "No try yet."
              : loaded
                ? "No try yet. The connector is running and tries it when it is due."
                : "No try yet. The connector is not running on this server."}
          </p>
        )}
      </Section>
      <OpenDeliveryButton delivery={delivery} />
      <Section title="Ids">
        <Fact label="Delivery">
          <CopyValue value={delivery.id} label="delivery ID" />
        </Fact>
        <Fact label="Trace">
          <CopyValue value={traceId} label="trace ID" />
        </Fact>
      </Section>
    </>
  );
}

export function StepDetailPanel({ row, model, traceId, className }: StepDetailPanelProps) {
  let body: React.ReactNode;
  if (!row || row.kind === "fold" || row.kind === "ghost" || row.kind === "truncated") {
    body = <p className="text-sm text-muted-foreground">Select a step to see its details.</p>;
  } else if (row.kind === "step") {
    body = <StepBody step={row.step} model={model} traceId={traceId} />;
  } else if (row.kind === "group" && row.delivery) {
    body = <DeliveryBody delivery={row.delivery} tries={row.tries ?? 0} traceId={traceId} />;
  } else if (row.kind === "group") {
    body = (
      <>
        <div
          className={cn(
            "text-[15px] font-bold text-foreground",
            isCodeShaped(row.title) && "font-mono text-[13.5px]",
          )}
        >
          {row.title}
        </div>
        <div className="mt-2 flex flex-wrap items-center gap-1.5">
          {row.source && <SourcePill source={row.source} size="sm" />}
          {row.status !== "muted" && <StatusPill status={row.status} size="sm" />}
        </div>
        {row.status === "muted" && (
          <p className="mt-3 text-[13px] text-muted-foreground">
            These steps name a parent span that wrote no row in this trace.
          </p>
        )}
        <Ids traceId={traceId} />
      </>
    );
  } else {
    body = (
      <>
        <div className="font-mono text-[13.5px] font-bold text-foreground">{row.eventType}</div>
        <p className="mt-1 text-[13px] text-muted-foreground">
          Sent to {row.deliveries.length} {row.deliveries.length === 1 ? "connector" : "connectors"}
          .
        </p>
        <Section title="Deliveries">
          <ul className="space-y-1.5">
            {row.deliveries.map((d) => (
              <li key={d.id} className="flex items-center gap-2 text-[12.5px]">
                <StatusPill
                  status={deliveryListStatus(d)}
                  label={deliveryLabel(d, model.tryCounts.get(d.id) ?? 0)}
                  size="sm"
                />
                <span className="truncate font-mono">{d.subscriber}</span>
              </li>
            ))}
          </ul>
        </Section>
        <Ids traceId={traceId} />
      </>
    );
  }

  return (
    <aside
      aria-label="Step details"
      data-testid="step-panel"
      className={cn("rounded-[10px] border border-border bg-card p-4", className)}
    >
      {body}
    </aside>
  );
}
