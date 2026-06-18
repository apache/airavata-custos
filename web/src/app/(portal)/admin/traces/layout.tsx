import type { ReactNode } from "react";
import { TracePermissionGate } from "./PermissionGate";

export default function TracesLayout({ children }: { children: ReactNode }) {
  return <TracePermissionGate>{children}</TracePermissionGate>;
}
