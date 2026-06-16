import { TableSkeleton } from "@/shared/ui/Loading";

export default function PacketsLoading() {
  return (
    <div className="space-y-4">
      <div className="h-40 rounded-md border bg-card" />
      <TableSkeleton />
    </div>
  );
}
