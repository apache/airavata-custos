import type { ReactNode } from "react";
import { Breadcrumbs } from "@/shared/ui/Breadcrumbs";

export default async function AllocationDetailLayout(props: {
  children: ReactNode;
  params: Promise<{ allocationId: string }>;
}) {
  const { allocationId } = await props.params;
  return (
    <div className="space-y-6">
      <Breadcrumbs
        items={[{ label: "Allocations", href: "/allocations" }, { label: allocationId }]}
      />
      {props.children}
    </div>
  );
}
