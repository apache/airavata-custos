"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useRouter } from "next/navigation";
import { Breadcrumbs } from "./Breadcrumbs";
import { UserPill } from "./UserPill";
import { useNavHistory } from "./useNavHistory";

function NavArrows() {
  const router = useRouter();
  const { canGoBack, canGoForward } = useNavHistory();
  const baseClass =
    "h-8 w-8 rounded-full text-muted-foreground hover:bg-muted/40 hover:text-foreground";
  const disabledClass = "opacity-50 cursor-not-allowed";
  return (
    <div className="flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon"
        aria-label="Go back"
        disabled={!canGoBack}
        className={cn(baseClass, !canGoBack && disabledClass)}
        onClick={() => router.back()}
      >
        <ChevronLeft className="h-4 w-4" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        aria-label="Go forward"
        disabled={!canGoForward}
        className={cn(baseClass, !canGoForward && disabledClass)}
        onClick={() => router.forward()}
      >
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  );
}

export function Topbar() {
  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-background px-8">
      <div className="flex items-center gap-4">
        <NavArrows />
        <Breadcrumbs />
      </div>

      <div className="flex items-center gap-4">
        <UserPill />
      </div>
    </header>
  );
}
