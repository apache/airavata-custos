import { AmieNav } from "../AmieNav";
import { ReconciliationContainer } from "./ReconciliationContainer";

export default function AmieReconcilePage() {
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">Reconciliation queue</h1>
        <p className="text-sm text-muted-foreground">
          Decoded packets that couldn't be mapped to a domain entity. Link to an existing project,
          account, or person — or skip with a reason.
        </p>
      </header>
      <AmieNav />
      <ReconciliationContainer />
    </div>
  );
}
