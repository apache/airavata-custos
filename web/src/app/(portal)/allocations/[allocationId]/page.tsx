import { AllocationDetail } from "@/features/core/allocations/components/AllocationDetail";

export default async function AllocationDetailPage(props: {
  params: Promise<{ allocationId: string }>;
}) {
  const { allocationId } = await props.params;
  return <AllocationDetail allocationId={allocationId} />;
}
