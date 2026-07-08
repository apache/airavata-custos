"use client";

import { Pencil, ShieldCheck } from "lucide-react";
import { Badge } from "@/shared/ui/badge";
import { Card, CardContent, CardHeader } from "@/shared/ui/card";
import { PermissionRW } from "../PermissionRW";
import { rwStateFor } from "../permissions";
import type { RoleRow } from "../types";
import { RoleFormDialog } from "./RoleFormDialog";

export function RoleCard({ role, memberCount }: { role: RoleRow; memberCount: number }) {
  const rwPermissions = rwStateFor(role.permissions);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div>
            <div className="flex items-center gap-1.5 font-heading text-base font-semibold text-foreground">
              {role.name}
              {role.is_system ? (
                <ShieldCheck
                  className="size-3.5 text-muted-foreground"
                  aria-label="System role"
                />
              ) : null}
            </div>
            <p className="mt-1 text-sm text-muted-foreground">{role.description}</p>
          </div>
          <Badge variant="secondary" className="shrink-0">
            {memberCount} {memberCount === 1 ? "member" : "members"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="border-t border-border" />

        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Permissions
          </h4>
          <ul className="space-y-2">
            {rwPermissions.map((p) => (
              <li key={p.section} className="flex items-center justify-between text-sm">
                <span className="text-foreground">{p.label}</span>
                <PermissionRW read={p.read} write={p.write} />
              </li>
            ))}
          </ul>
        </div>

        <div className="border-t border-border" />

        <RoleFormDialog
          role={role}
          triggerRender={
            <button
              type="button"
              className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground"
            />
          }
          triggerContent={
            <>
              <Pencil className="size-3.5" />
              Edit
            </>
          }
        />
      </CardContent>
    </Card>
  );
}
