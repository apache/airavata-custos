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

import { useUpdateUserName, useUser } from "@/features/core/users/queries";
import type { User } from "@/generated/core/types.gen";
import { Field, FormDialog, edited, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { displayName } from "@/shared/users-admin/displayName";
import * as React from "react";

const PARTS = [
  { name: "first_name", label: "First name" },
  { name: "middle_name", label: "Middle name" },
  { name: "last_name", label: "Last name" },
] as const;

// PUT /users/{id} replaces the whole name, so every part is sent.
export function EditUserNameDialog({ user: opened }: { user: User }) {
  const [open, setOpen] = React.useState(false);
  const record = useUser(open ? opened.id : undefined);
  const update = useUpdateUserName();
  const user = record.data ?? opened;
  const stored = {
    first_name: user.first_name,
    middle_name: user.middle_name ?? "",
    last_name: user.last_name,
  };
  return (
    <FormDialog
      open={open}
      onOpenChange={setOpen}
      trigger={
        <Button variant="outline" size="sm" aria-label={`Edit the name of ${displayName(user)}`}>
          Edit name
        </Button>
      }
      title="Edit name"
      submitLabel="Save"
      isPending={update.isPending}
      record={record}
      gone="This user no longer exists"
      onSubmit={(form, close) => {
        const body = edited(stored, {
          first_name: text(form, "first_name"),
          middle_name: text(form, "middle_name"),
          last_name: text(form, "last_name"),
        });
        if (!body) return close();
        toastOnSuccess(update.mutateAsync({ path: { id: user.id }, body }), "Name updated", close);
      }}
    >
      {PARTS.map(({ name, label }) => (
        <Field
          key={name}
          label={label}
          name={name}
          defaultValue={stored[name]}
          maxLength={255}
          // The API treats a blank part as unchanged, so a part already set cannot be cleared.
          required={stored[name] !== ""}
          title={`Enter a ${label.toLowerCase()}, not only spaces.`}
          placeholder={name === "middle_name" && stored[name] === "" ? "Optional" : undefined}
        />
      ))}
    </FormDialog>
  );
}
