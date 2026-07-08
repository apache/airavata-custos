"use client";

import type { UserIdentity } from "@/generated/core/types.gen";
import { Badge } from "@/shared/ui/badge";
import { identitySourceIcon, identitySourceLabel } from "./identities";

const VISIBLE_IDENTITY_COUNT = 2;

export function IdentitiesCell({
  identities,
  expanded,
  onToggleExpand,
}: {
  identities: UserIdentity[];
  expanded: boolean;
  onToggleExpand: () => void;
}) {
  const visible = expanded ? identities : identities.slice(0, VISIBLE_IDENTITY_COUNT);
  const hiddenCount = identities.length - visible.length;

  return (
    <div className="flex w-[220px] flex-wrap items-center gap-1.5">
      {identities.length === 0 ? (
        <span className="text-sm text-muted-foreground">No identities</span>
      ) : (
        visible.map((identity) => {
          const Icon = identitySourceIcon(identity.source);
          return (
            <Badge key={identity.id} variant="outline">
              <Icon data-icon="inline-start" />
              {identitySourceLabel(identity.source)}
            </Badge>
          );
        })
      )}
      {hiddenCount > 0 ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          +{hiddenCount}
        </button>
      ) : expanded && identities.length > VISIBLE_IDENTITY_COUNT ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          Show less
        </button>
      ) : null}
    </div>
  );
}
