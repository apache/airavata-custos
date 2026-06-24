"use client";

import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as React from "react";
import { NAV_GROUP_LABELS, type NavGroup, type NavItem, portalNav } from "./nav";

const GROUP_ORDER: NavGroup[] = ["allocations", "admin"];

export function Sidebar() {
  const pathname = usePathname();
  const ability = useAbility();
  const navRef = React.useRef<HTMLElement>(null);
  const thumbRef = React.useRef<HTMLDivElement>(null);
  const timerRef = React.useRef<ReturnType<typeof setTimeout>>(undefined);
  const [scrolling, setScrolling] = React.useState(false);
  const [hovering, setHovering] = React.useState(false);
  const [isScrollable, setIsScrollable] = React.useState(false);

  const updateThumb = React.useCallback(() => {
    const nav = navRef.current;
    const thumb = thumbRef.current;
    if (!nav || !thumb) return;
    const ratio = nav.clientHeight / nav.scrollHeight;
    if (ratio >= 1) {
      setIsScrollable(false);
      return;
    }
    setIsScrollable(true);
    const thumbHeight = Math.max(ratio * nav.clientHeight, 32);
    const maxScroll = nav.scrollHeight - nav.clientHeight;
    const thumbTop = maxScroll > 0 ? (nav.scrollTop / maxScroll) * (nav.clientHeight - thumbHeight) : 0;
    thumb.style.height = `${thumbHeight}px`;
    thumb.style.transform = `translateY(${thumbTop}px)`;
  }, []);

  React.useEffect(() => {
    const el = navRef.current;
    if (!el) return;
    updateThumb();
    function onScroll() {
      updateThumb();
      setScrolling(true);
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setScrolling(false), 800);
    }
    el.addEventListener("scroll", onScroll, { passive: true });
    const ro = new ResizeObserver(updateThumb);
    ro.observe(el);
    return () => { el.removeEventListener("scroll", onScroll); clearTimeout(timerRef.current); ro.disconnect(); };
  }, [updateThumb]);

  const visible = portalNav.filter((item) => {
    if (!item.ability) return true;
    return ability.can(item.ability.action, item.ability.subject);
  });

  const groups = GROUP_ORDER.map((group) => ({
    group,
    items: visible.filter((item) => item.group === group),
  })).filter((g) => g.items.length > 0);

  const showThumb = isScrollable && (scrolling || hovering);

  return (
    <aside className="flex w-[240px] shrink-0 flex-col overflow-hidden border-r border-border bg-sidebar text-sidebar-foreground">
      <div className="px-6 pt-8 pb-6">
        <Link
          href="/"
          className="font-display text-2xl font-extrabold uppercase tracking-tight text-brand"
        >
          Custos
        </Link>
      </div>

      <div className="relative min-h-0 flex-1">
        <nav
          ref={navRef}
          className="sidebar-scroll-hidden flex h-full flex-col overflow-y-scroll"
          onMouseEnter={() => { setHovering(true); updateThumb(); }}
          onMouseLeave={() => setHovering(false)}
        >
          {groups.map(({ group, items }, idx) => (
            <div key={group} className={cn("flex flex-col", idx > 0 && "mt-4")}>
              <div className="px-6 pt-2 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                {NAV_GROUP_LABELS[group]}
              </div>
              {items.map((item) => (
                <SidebarLink key={item.href} item={item} active={isActive(pathname, item.href)} />
              ))}
            </div>
          ))}
        </nav>

        {/* Custom overlay scrollbar thumb — no layout space consumed */}
        <div
          ref={thumbRef}
          className="pointer-events-none absolute top-0 right-1 w-1 rounded-full transition-opacity duration-300"
          style={{
            backgroundColor: "color-mix(in srgb, var(--foreground) 40%, transparent)",
            opacity: showThumb ? 1 : 0,
          }}
        />
      </div>
    </aside>
  );
}

function isActive(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(`${href}/`);
}

function SidebarLink({ item, active }: { item: NavItem; active: boolean }) {
  const Icon = item.icon;
  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      className={cn(
        "relative flex h-11 items-center gap-3 px-6 text-sm font-medium transition",
        active
          ? "bg-[var(--sidebar-active)] font-semibold text-brand"
          : "text-muted-foreground hover:bg-[var(--sidebar-hover)] hover:text-foreground",
      )}
    >
      <Icon className="h-5 w-5 stroke-[1.75]" />
      <span className="truncate">{item.label}</span>
      {active && <span className="absolute top-2 right-0 bottom-2 w-1 rounded-l-full bg-brand" />}
    </Link>
  );
}
