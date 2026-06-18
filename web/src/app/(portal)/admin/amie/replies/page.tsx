import { AmieNav } from "../AmieNav";
import { ReplyTrackerContainer } from "./ReplyTrackerContainer";

export default function AmieRepliesPage() {
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">Reply tracker</h1>
        <p className="text-sm text-muted-foreground">
          Outgoing inform_* packets the connector sends back to ACCESS. Retry any that failed or are
          stuck PENDING.
        </p>
      </header>
      <AmieNav />
      <ReplyTrackerContainer />
    </div>
  );
}
