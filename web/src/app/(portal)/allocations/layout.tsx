import type { ReactNode } from "react";
import { AllocationsPermissionGate } from "./PermissionGate";

export default function AllocationsLayout({ children }: { children: ReactNode }) {
  return <AllocationsPermissionGate>{children}</AllocationsPermissionGate>;
}
