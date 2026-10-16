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

import {
  type UserManagementRow,
  useDeleteIdentity,
  useSaveIdentity,
  userIdentities,
} from "@/features/core/users/queries";
import type { UserIdentityInput } from "@/generated/core/types.gen";
import { Field, FormDialog, SelectField, edited, text } from "@/shared/ui/FormDialog";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { useQuery } from "@tanstack/react-query";
import { PencilIcon, PlusIcon, XIcon } from "lucide-react";
import * as React from "react";
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
  const [editing, setEditing] = React.useState<UserIdentityInput | null>(null);

  return (
    <DrawerSection
      title="External identities"
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
          <IdentityFormDialog
            userId={user.id}
            identity={editing}
            onClose={() => setEditing(null)}
          />
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
                <Button
                  variant="ghost"
                  size="icon-xs"
                  aria-label={`Edit ${label} identity ${identity.external_id}`}
                  onClick={() => setEditing(identity)}
                >
                  <PencilIcon />
                </Button>
              ) : null}
              {canWrite ? (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  aria-label={`Remove ${label} identity ${identity.external_id}`}
                  pending={remove.isPending && remove.variables === identity.id}
                  onClick={() =>
                    confirmToast(
                      `Remove the ${label} identity ${identity.external_id}? The user can no longer be matched through it.`,
                      "Remove",
                      () => toastOnSuccess(remove.mutateAsync(identity.id), "Identity removed"),
                    )
                  }
                >
                  <XIcon />
                </Button>
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
  identity: opened,
  onClose,
}: {
  userId: string;
  identity: UserIdentityInput;
  onClose: () => void;
}) {
  const save = useSaveIdentity(userId);
  // The drawer's identity list; the dialog's own opening refetches it.
  const record = useQuery(userIdentities(userId));
  const found = record.data?.find((i) => i.id === opened.id);
  const identity = found ?? opened;
  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={identity.id ? "Edit identity" : "Add identity"}
      submitLabel="Save"
      isPending={save.isPending}
      record={opened.id ? record : undefined}
      gone="This identity no longer exists"
      missing={Boolean(opened.id) && !found}
      onSubmit={(form) => {
        const fields = {
          source: text(form, "source"),
          external_id: text(form, "external_id"),
          email: text(form, "email"),
        };
        const body = identity.id ? edited(identity, fields) : fields;
        if (!body) return onClose();
        toastOnSuccess(
          save.mutateAsync(body),
          identity.id ? "Identity updated" : "Identity added",
          onClose,
        );
      }}
    >
      <SelectField
        label="Source"
        name="source"
        defaultValue={identity.source}
        options={[
          ...new Set([...Object.keys(IDENTITY_SOURCE_LABELS), identity.source ?? "oidc"]),
        ].map((value) => ({ value, label: identitySourceLabel(value) }))}
      />
      <Field
        label="External ID"
        name="external_id"
        defaultValue={identity.external_id}
        maxLength={255}
        required
      />
      {identity.email ? (
        <div className="space-y-1">
          <Field
            label="Email"
            name="email"
            type="email"
            defaultValue={identity.email}
            maxLength={255}
            required
            aria-describedby="identity-email-hint"
          />
          <p id="identity-email-hint" className="text-xs text-muted-foreground">
            An email can be changed, not removed.
          </p>
        </div>
      ) : (
        <Field label="Email" name="email" type="email" maxLength={255} />
      )}
    </FormDialog>
  );
}
