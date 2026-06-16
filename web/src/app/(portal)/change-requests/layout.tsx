import type { ReactNode } from "react";
import { ChangeRequestsPermissionGate } from "./PermissionGate";

export default function ChangeRequestsLayout({ children }: { children: ReactNode }) {
  return <ChangeRequestsPermissionGate>{children}</ChangeRequestsPermissionGate>;
}
