import type { Metadata } from "next";
import { TraceListContainer } from "@/features/core/audit/components/TraceListContainer";

export const metadata: Metadata = {
  title: "Tracing — Admin",
};

export default function AdminTracesPage() {
  return <TraceListContainer />;
}
