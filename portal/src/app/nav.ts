// Sidebar navigation manifest. Each entry maps a label and Lucide icon to
// a route under the App Router. Only /signer/certificates is fully
// implemented today; the rest point at PlaceholderPage stubs so the
// sidebar matches the reference layout.
// TODO: replace placeholders once the corresponding portal areas exist.
import {
  BadgeCheck,
  LayoutDashboard,
  Server,
  Users,
  Wrench,
} from "lucide-react";

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
