import type { ReactNode } from "react";
import { ProjectsPermissionGate } from "./PermissionGate";

export default function ProjectsLayout({ children }: { children: ReactNode }) {
  return <ProjectsPermissionGate>{children}</ProjectsPermissionGate>;
}
