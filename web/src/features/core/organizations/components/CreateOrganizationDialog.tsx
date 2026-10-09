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

import { toastOnSuccess } from "@/shared/ui/sonner";
import { Button } from "@/shared/ui/button";
import { Field, FormDialog, text } from "@/shared/ui/FormDialog";
import { useCreateOrganization } from "../queries";

export function CreateOrganizationDialog() {
  const create = useCreateOrganization();
  return (
    <FormDialog
      trigger={<Button>+ Create organization</Button>}
      title="Create organization"
      description="Register an institution or resource provider."
      submitLabel="Create organization"
      isPending={create.isPending}
      onSubmit={(form, close) =>
        create.mutate(
          {
            body: {
              name: text(form, "name"),
              originated_id: text(form, "originated_id") || undefined,
            },
          },
          toastOnSuccess("Organization created", close),
        )
      }
    >
      <Field label="Name" name="name" placeholder="Georgia Institute of Technology" required />
      <Field
        label="Originated ID"
        name="originated_id"
        placeholder="Optional, e.g. ACCESS org code"
      />
    </FormDialog>
  );
}
