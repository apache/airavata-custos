import type { Metadata } from "next";
import { TraceListContainer } from "@/features/core/audit/components/TraceListContainer";

export const metadata: Metadata = {
  title: "Trace · Admin",
};

export default async function AdminTraceDetailPage({
  params,
}: {
  params: Promise<{ traceId: string }>;
}) {
  const { traceId } = await params;
  return <TraceListContainer initialTraceId={traceId} />;
}
