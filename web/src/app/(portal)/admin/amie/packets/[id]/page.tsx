import { AmieNav } from "../../AmieNav";
import { PacketInboxContainer } from "../PacketInboxContainer";

export default async function AmiePacketDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">AMIE packet inbox</h1>
        <p className="text-sm text-muted-foreground">
          Deep link to packet <span className="font-mono">{id}</span>.
        </p>
      </header>
      <AmieNav />
      <PacketInboxContainer initialPacketId={id} />
    </div>
  );
}
