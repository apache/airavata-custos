// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

"use client";

import { PencilIcon, PlusIcon, XIcon } from "lucide-react";
import * as React from "react";
import {
  type UserManagementRow,
  useDeleteIdentity,
  useSaveIdentity,
} from "@/features/core/users/queries";
import type { UserIdentity } from "@/generated/core/types.gen";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Field, FormDialog, text } from "@/shared/ui/FormDialog";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { DrawerSection } from "./DrawerSection";
import { IDENTITY_SOURCE_LABELS, identitySourceIcon, identitySourceLabel } from "./identities";

export function IdentitiesSection({
  user,
  canWrite,
}: {
  user: UserManagementRow;
  canWrite: boolean;
}) {
  const remove = useDeleteIdentity(user.id);
  // null: dialog closed; an empty identity adds, an existing one edits.
  const [editing, setEditing] = React.useState<UserIdentity | null>(null);

  return (
    <DrawerSection
      title="External Identities"
      action={
        canWrite ? (
          <Button variant="ghost" size="xs" onClick={() => setEditing({})}>
            <PlusIcon data-icon="inline-start" />
            Add
          </Button>
        ) : null
      }
      isLoading={user.identitiesLoading}
      isError={user.identitiesError}
      empty={user.identities.length === 0 && "No identities."}
      footer={
        editing ? (
          <IdentityFormDialog userId={user.id} identity={editing} onClose={() => setEditing(null)} />
        ) : null
      }
    >
      <ul className="space-y-1.5">
        {user.identities.map((identity) => {
          const Icon = identitySourceIcon(identity.source);
          const label = identitySourceLabel(identity.source);
          return (
            <li key={identity.id} className="flex items-center gap-2 text-sm">
              <Badge variant="outline">
                <Icon data-icon="inline-start" />
                {label}
              </Badge>
              <span className="min-w-0 flex-1 truncate font-mono text-xs text-muted-foreground">
                {identity.external_id}
              </span>
              {canWrite ? (
                <>
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    aria-label={`Edit ${label} identity`}
                    onClick={() => setEditing(identity)}
                  >
                    <PencilIcon />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    aria-label={`Remove ${label} identity`}
                    disabled={remove.isPending}
                    onClick={() =>
                      confirmToast(
                        `Remove the ${label} identity ${identity.external_id ?? ""}? The user can no longer be matched through it.`,
                        "Remove",
                        () =>
                          remove.mutate(identity.id ?? "", toastOnSuccess("Identity removed")),
                      )
                    }
                  >
                    <XIcon />
                  </Button>
                </>
              ) : null}
            </li>
          );
        })}
      </ul>
    </DrawerSection>
  );
}

function IdentityFormDialog({
  userId,
  identity,
  onClose,
}: {
  userId: string;
  identity: UserIdentity;
  onClose: () => void;
}) {
  const save = useSaveIdentity(userId);
  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={identity.id ? "Edit identity" : "Add identity"}
      submitLabel="Save"
      isPending={save.isPending}
      onSubmit={(form) =>
        save.mutate(
          {
            ...identity,
            source: text(form, "source"),
            external_id: text(form, "external_id"),
            email: text(form, "email"),
            oidc_sub: text(form, "oidc_sub"),
          },
          toastOnSuccess(identity.id ? "Identity updated" : "Identity added", onClose),
        )
      }
    >
      <Field
        label="Source"
        name="source"
        defaultValue={identity.source}
        list="identity-sources"
        required
      />
      <datalist id="identity-sources">
        {Object.keys(IDENTITY_SOURCE_LABELS).map((source) => (
          <option key={source} value={source} />
        ))}
      </datalist>
      <Field label="External ID" name="external_id" defaultValue={identity.external_id} required />
      <Field label="Email" name="email" type="email" defaultValue={identity.email} />
      <Field label="OIDC subject" name="oidc_sub" defaultValue={identity.oidc_sub} />
    </FormDialog>
  );
}
