import { AmieNav } from "../AmieNav";
import { PacketInboxContainer } from "./PacketInboxContainer";

export default function AmiePacketsPage() {
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">AMIE packet inbox</h1>
        <p className="text-sm text-muted-foreground">
          All packets flowing through the AMIE connector. Filter, drill into a packet, retry, or
          mark as processed.
        </p>
      </header>
      <AmieNav />
      <PacketInboxContainer />
    </div>
  );
}
