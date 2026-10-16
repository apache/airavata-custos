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

import { useOrganizations } from "@/features/core/organizations/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Field, FormDialog, SelectField, named, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useCreateTempAccount } from "../queries";

export function CreateTempAccountDialog({ onCreated }: { onCreated: (userId: string) => void }) {
  const create = useCreateTempAccount();
  // The account needs an existing organization; without Organization read, or when the list fails
  // to load, its id is typed.
  const canReadOrganizations = useAbility().can("read", "Organization");
  const organizations = useOrganizations({ limit: 200 }, canReadOrganizations);
  const canPickOrganization = canReadOrganizations && !organizations.isError;
  return (
    <FormDialog
      trigger={<Button>+ Create temporary account</Button>}
      title="Create temporary account"
      description="Creates a temporary user that can be granted an allocation."
      submitLabel="Create account"
      isPending={create.isPending}
      onSubmit={(form, close) =>
        toastOnSuccess(
          create.mutateAsync({
            first_name: text(form, "first_name"),
            last_name: text(form, "last_name"),
            email: text(form, "email"),
            organization_id: text(form, "organization"),
          }),
          (user) => `Created ${user.email}`,
          (user) => {
            close();
            onCreated(user.id);
          },
        )
      }
    >
      <div className="grid grid-cols-2 gap-4">
        <Field
          label="First name"
          name="first_name"
          maxLength={255}
          required
          title="Enter a name, not only spaces."
        />
        <Field label="Last name" name="last_name" maxLength={255} />
      </div>
      <Field label="Email" name="email" type="email" maxLength={255} required />
      {canPickOrganization ? (
        <SelectField
          label="Organization"
          name="organization"
          options={named(organizations.data?.items)}
        />
      ) : (
        <Field label="Organization ID" name="organization" maxLength={255} required />
      )}
    </FormDialog>
  );
}
