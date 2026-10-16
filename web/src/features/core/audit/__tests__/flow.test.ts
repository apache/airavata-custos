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

import approveDone from "@/features/core/audit/__fixtures__/trace.approve.done.json";
import approveFailed from "@/features/core/audit/__fixtures__/trace.approve.failed.json";
import createdDone from "@/features/core/audit/__fixtures__/trace.created.done.json";
import packetFailed from "@/features/core/audit/__fixtures__/trace.packet.failed.json";
import { traceDetailSchema } from "@/features/core/audit/schemas";
import type { TraceDetail } from "@/features/core/audit/types";
import {
  buildFlow,
  containersOf,
  defaultExpanded,
  deliveryLabel,
  expandableKeys,
  flattenFlow,
  pathRails,
} from "@/features/core/audit/utils";
import { describe, expect, it } from "vitest";

const failed = traceDetailSchema.parse(approveFailed);
const done = traceDetailSchema.parse(approveDone);
const created = traceDetailSchema.parse(createdDone);
const packet = traceDetailSchema.parse(packetFailed);

function allRows(detail: TraceDetail) {
  const model = buildFlow(detail);
  return { model, rows: flattenFlow(model, new Set(expandableKeys(model))) };
}

describe("buildFlow", () => {
  // Make sure the trace reads as origin, then the published event, then one
  // group per delivery, so the hop across the event bus is drawn, not implied.
  it("hangs one delivery group per connector under the published event", () => {
    const { rows } = allRows(failed);
    const kinds = rows.map((r) => r.kind);
    expect(kinds.slice(0, 3)).toEqual(["group", "step", "hop"]);
    const groups = rows.filter((r) => r.kind === "group" && r.delivery);
    expect(groups.map((g) => g.kind === "group" && g.title)).toEqual([
      "amie-processor",
      "comanage-identity-provisioner",
    ]);
  });

  // Make sure the failing step is the connector's own error row on the last
  // try, not the worker's try row that repeats it, and that the path to it
  // runs from the origin through the hop into that delivery and no other.
  it("picks the connector's last error as the failing step and marks the path to it", () => {
    const { model, rows } = allRows(failed);
    const failing = model.steps.get(model.failingStepId ?? "")?.step;
    expect(failing?.event_type).toBe("ComanageProvisioningFailed");
    const onPath = rows.filter((r) => "onPath" in r && r.onPath);
    expect(model.failingStepId).toBe(model.errorStepIds.at(-1));
    expect(onPath.slice(0, 4).map((r) => r.kind)).toEqual(["group", "step", "hop", "group"]);
    expect(onPath.at(-1)?.key).toBe(model.failingStepId);
    const amie = rows.find((r) => r.kind === "group" && r.title === "amie-processor");
    expect(amie?.kind === "group" && amie.onPath).toBe(false);
  });

  // Make sure a failed try inside a delivery that later succeeded is history
  // and does not make the trace look broken.
  it("ignores errors inside a delivery that succeeded", () => {
    const { model } = allRows(done);
    expect(model.failingStepId).toBeNull();
    expect(model.errorStepIds).toEqual([]);
  });

  // Make sure a done trace has no failing step, so an AMIE attempt a later one superseded reads as a past try.
  it("finds no failing step in a done trace", () => {
    const { model, rows } = allRows({ ...packet, status: "ok" });
    expect(model.failingStepId).toBeNull();
    expect(rows.find((r) => r.kind === "group")).toMatchObject({ status: "past" });
    expect(allRows(packet).model.failingStepId).not.toBeNull();
  });

  // Make sure only the last AMIE attempt of a gave-up trace reads as failed; earlier ones read Retried.
  it("marks an attempt a later attempt superseded as a past try", () => {
    const retry = packet.tree.map((n) => ({
      ...n,
      id: `${n.id}-2`,
      span_id: "9999e099218d2af1",
      created_at: n.created_at.replace("20:29", "20:39"),
    }));
    const { rows } = allRows({ ...packet, tree: [...packet.tree, ...retry] });
    const groups = rows.filter((r) => r.kind === "group");
    expect(groups.map((g) => g.kind === "group" && g.status)).toEqual(["past", "error"]);
  });

  // Make sure a connector that published its own event shows that hop under
  // its steps, with the next connector's delivery hanging off it.
  it("nests a hop published by a connector under that connector's steps", () => {
    const { rows } = allRows(done);
    const hops = rows.filter((r) => r.kind === "hop");
    expect(hops.map((h) => h.kind === "hop" && h.eventType)).toEqual([
      "compute_cluster_user::approve",
      "compute_cluster_user::provision",
    ]);
    expect(hops[1]?.depth ?? 0).toBeGreaterThan(hops[0]?.depth ?? 0);
  });

  // Make sure rows a connector writes in the delivery's own span, with no span
  // of their own, still belong to that delivery and not to the origin.
  it("keeps rows written in the delivery span inside the delivery", () => {
    const { rows } = allRows(created);
    expect(rows[0]?.kind).toBe("hop");
    const steps = rows.filter((r) => r.kind === "step");
    expect(steps.map((s) => s.kind === "step" && s.step.event_type)).toEqual([
      "NOTIFICATION_SENT",
      "EVENT_DELIVERY_SUCCEEDED",
    ]);
  });

  // Make sure a packet whose rows share one span under the request span reads
  // as one origin group, not as steps without a parent.
  it("groups roots of one span as a single origin", () => {
    const { model, rows } = allRows(packet);
    const first = rows[0];
    expect(first?.kind).toBe("group");
    expect(rows.slice(1).map((r) => r.kind)).toEqual(["step", "step", "step"]);
    expect(model.groups.some((g) => g.kind === "orphans")).toBe(false);
  });

  // Make sure a long run of failed tries folds to the first and last try so a
  // ten try delivery is three rows, and opens in place.
  it("folds the middle tries", () => {
    const model = buildFlow(failed);
    const rows = flattenFlow(model, defaultExpanded(model));
    const fold = rows.find((r) => r.kind === "fold");
    expect(fold?.kind === "fold" && fold.count).toBe(8);
    const tries = rows.filter((r) => r.kind === "step" && r.try?.outcome === "failed");
    expect(tries.map((t) => t.kind === "step" && t.try?.n)).toEqual([1, 10]);
  });

  // Make sure tries keep counting past an admin retry, which resets the
  // server's attempt, and the fold and the delivery's label count tries only, not the retry row.
  it("numbers tries continuously across an admin retry", () => {
    const deliveryId = "8aaf8ebb-91e2-4df0-9026-b09f76ed05d1";
    const row = (id: string, event_type: string, created_at: string, description: string) => ({
      ...failed.tree[0]?.children[0],
      id,
      event_type,
      created_at,
      description,
      span_id: id.slice(0, 16),
      entity_type: "event_delivery",
      entity_id: deliveryId,
      status: "ok",
      children: [],
    });
    const root = failed.tree[0];
    const detail = traceDetailSchema.parse({
      ...failed,
      deliveries: failed.deliveries.map((d) =>
        d.id === deliveryId ? { ...d, status: "SUCCEEDED", attempts: 1 } : d,
      ),
      tree: [
        {
          ...root,
          children: [
            ...(root?.children ?? []),
            row("aaaaaaaaaaaaaaaa1", "EVENT_DELIVERY_RETRIED", "2026-10-09T21:00:00Z", "{}"),
            row(
              "bbbbbbbbbbbbbbbb1",
              "EVENT_DELIVERY_SUCCEEDED",
              "2026-10-09T21:00:05Z",
              '{"attempt":1}',
            ),
          ],
        },
      ],
    });
    const model = buildFlow(detail);
    const rows = flattenFlow(model, defaultExpanded(model).add(`d:${deliveryId}`));
    const fold = rows.find((r) => r.kind === "fold");
    expect(fold?.kind === "fold" && fold.count).toBe(9);
    const last = rows.filter((r) => r.kind === "step" && r.try).at(-1);
    expect(last?.kind === "step" && last.try).toEqual({ n: 11, outcome: "succeeded" });
    const delivery = rows.find((r) => r.kind === "group" && r.delivery?.id === deliveryId);
    expect(
      delivery?.kind === "group" &&
        delivery.delivery &&
        deliveryLabel(delivery.delivery, delivery.tries ?? 0),
    ).toBe("Done after 11 tries");
  });
});

describe("pathRails", () => {
  // Make sure the red rail runs down the spine past a sibling delivery that
  // is fine and stops after the failing step, so it leads to one place.
  it("colors only the rails that lead to the failing step", () => {
    const { rows } = allRows(failed);
    const rails = pathRails(rows);
    const amieStep = rows.findIndex((r) => r.kind === "step" && r.step.event_type === "REPLY_SENT");
    expect(rails[amieStep]).toEqual([true, true, false]);
    const failing = rows.findIndex((r) => r.kind === "step" && r.failing);
    expect(rails[failing]?.every(Boolean)).toBe(true);
    expect(rails[failing + 1]?.some(Boolean)).toBe(false);
  });
});

describe("containersOf", () => {
  // Make sure a step linked from the URL can be shown even when its delivery
  // group and fold are collapsed.
  it("lists the groups that must open to show a folded try", () => {
    const { model, rows } = allRows(failed);
    const ninthTry = rows.filter((r) => r.kind === "step" && r.try).at(-2);
    const keys = containersOf(model, ninthTry?.key ?? "");
    expect(keys.some((k) => k.startsWith("o:"))).toBe(true);
    expect(keys.some((k) => k.startsWith("d:"))).toBe(true);
    expect(keys.some((k) => k.startsWith("f:"))).toBe(true);
  });
});
