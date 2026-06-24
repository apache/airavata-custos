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

import { toast } from "sonner";
import { ReconciliationQueue } from "@/features/connectors/amie/components/ReconciliationQueue";
import {
  useLinkUnmapped,
  useResolvePacket,
  useUnmapped,
} from "@/features/connectors/amie/queries";
import type { Packet } from "@/features/connectors/amie/types";
import { listProjects } from "@/features/core/projects/api";
import type { ProjectAutocompletePick } from "@/shared/ui/ProjectAutocomplete";

async function searchProjectsAdapter(q: string): Promise<ProjectAutocompletePick[]> {
  if (!q.trim()) return [];
  const envelope = await listProjects({ q, limit: 10 });
  return envelope.items.map((p) => ({
    id: p.id,
    originated_id: p.originated_id,
    title: p.title,
  }));
}

export function ReconciliationContainer() {
  const unmappedQuery = useUnmapped({ limit: 100 });
  const linkMutation = useLinkUnmapped();
  const resolveMutation = useResolvePacket();
  const rows = unmappedQuery.data?.packets ?? [];

  async function handleLink(packet: Packet, entity_type: string, entity_id: string) {
    try {
      await linkMutation.mutateAsync({ id: packet.id, entity_type, entity_id });
      toast.success(`Linked to ${entity_type} ${entity_id} — audit entry created`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Link failed");
    }
  }

  async function handleSkip(packet: Packet, reason: string) {
    try {
      await resolveMutation.mutateAsync({ id: packet.id, reason: `Skip — ${reason}` });
      toast.success(`Skipped ${packet.amie_id}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Skip failed");
    }
  }

  return (
    <ReconciliationQueue
      rows={rows}
      total={unmappedQuery.data?.total ?? rows.length}
      isLoading={unmappedQuery.isLoading}
      error={unmappedQuery.error}
      searchProjects={searchProjectsAdapter}
      onLink={handleLink}
      onSkip={handleSkip}
      onRefresh={() => unmappedQuery.refetch()}
    />
  );
}
