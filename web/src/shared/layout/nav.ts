import {
  Activity,
  ClipboardList,
  FolderKanban,
  type LucideIcon,
  Server,
  UserCog,
} from "lucide-react";

export type AbilityCheck = { action: string; subject: string };

export type NavGroup = "allocations" | "admin";

export type NavItem = {
  href: string;
  label: string;
  icon: LucideIcon;
  group: NavGroup;
  ability?: AbilityCheck;
};

export const NAV_GROUP_LABELS: Record<NavGroup, string> = {
  allocations: "My allocations",
  admin: "Site administration",
};

export const portalNav: NavItem[] = [
  {
    href: "/allocations",
    label: "Allocations",
    icon: Server,
    group: "allocations",
    ability: { action: "read", subject: "Allocation" },
  },
  {
    href: "/projects",
    label: "Projects",
    icon: FolderKanban,
    group: "allocations",
    ability: { action: "read", subject: "Project" },
  },
  {
    href: "/admin/users",
    label: "Users & Permissions",
    icon: UserCog,
    group: "admin",
  },
  {
    href: "/admin/traces",
    label: "Tracing",
    icon: Activity,
    group: "admin",
    ability: { action: "read", subject: "Trace" },
  },
  {
    href: "/admin/amie",
    label: "AMIE",
    icon: ClipboardList,
    group: "admin",
    ability: { action: "read", subject: "AMIE" },
  },
];
