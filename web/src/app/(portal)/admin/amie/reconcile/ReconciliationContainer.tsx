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
