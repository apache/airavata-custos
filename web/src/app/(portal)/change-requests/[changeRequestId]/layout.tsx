import type { ReactNode } from "react";
import { Breadcrumbs } from "@/shared/ui/Breadcrumbs";

export default async function ChangeRequestDetailLayout(props: {
  children: ReactNode;
  params: Promise<{ changeRequestId: string }>;
}) {
  const { changeRequestId } = await props.params;
  return (
    <div className="space-y-6">
      <Breadcrumbs
        items={[
          { label: "Change requests", href: "/change-requests" },
          { label: changeRequestId },
        ]}
      />
      {props.children}
    </div>
  );
}
