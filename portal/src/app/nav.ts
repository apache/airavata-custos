import {
  BadgeCheck,
  LayoutDashboard,
  Server,
  Users,
  Wrench,
} from "lucide-react";

// Only /signer/certificates is fully implemented; the others render
// PlaceholderPage stubs so the sidebar mirrors the reference layout.
// TODO: replace placeholders once the corresponding portal areas exist.
export const navItems = [
  {
    label: "Overview",
    href: "/",
    icon: LayoutDashboard,
  },
  {
    label: "Allocations",
    href: "/allocations",
    icon: Server,
  },
  {
    label: "Tools",
    href: "/tools",
    icon: Wrench,
  },
  {
    label: "SSH Certificates",
    href: "/signer/certificates",
    icon: BadgeCheck,
  },
  {
    label: "Clients",
    href: "/clients",
    icon: Users,
  },
];
