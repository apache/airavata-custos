import type { ReactNode } from "react";
import { AmiePermissionGate } from "./PermissionGate";

export default function AmieLayout({ children }: { children: ReactNode }) {
  return <AmiePermissionGate>{children}</AmiePermissionGate>;
}
