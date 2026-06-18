import { ChangeRequestDetail } from "@/features/core/allocations/components/ChangeRequestDetail";

export default async function ChangeRequestDetailPage(props: {
  params: Promise<{ changeRequestId: string }>;
}) {
  const { changeRequestId } = await props.params;
  return <ChangeRequestDetail changeRequestId={changeRequestId} />;
}
