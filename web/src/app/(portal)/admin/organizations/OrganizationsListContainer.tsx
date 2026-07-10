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

import * as React from "react";
import { toast } from "sonner";
import { CreateOrganizationDialog } from "@/features/core/organizations/components/CreateOrganizationDialog";
import { OrganizationsList } from "@/features/core/organizations/components/OrganizationsList";
import { useCreateOrganization, useOrganizations } from "@/features/core/organizations/queries";
import type { CreateOrganizationPayload } from "@/features/core/organizations/schemas";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";

const PAGE_SIZE = 50;

export function OrganizationsListContainer() {
  const ability = useAbility();
  const canCreate = ability.can("manage", "Organization");

  const [page, setPage] = React.useState(1);
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [createError, setCreateError] = React.useState<string | null>(null);

  const query = useOrganizations({ limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE });
  const createMutation = useCreateOrganization();

  async function handleCreate(payload: CreateOrganizationPayload) {
    setCreateError(null);
    try {
      await createMutation.mutateAsync(payload);
      toast.success("Organization created");
      setDialogOpen(false);
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : "Failed to create organization");
    }
  }

  return (
    <>
      <OrganizationsList
        rows={query.data?.items ?? []}
        isLoading={query.isLoading}
        error={(query.error as Error | null) ?? null}
        onRetry={() => query.refetch()}
        page={page}
        pageSize={PAGE_SIZE}
        total={query.data?.total ?? 0}
        onPageChange={setPage}
        headerCta={
          canCreate ? (
            <Button
              onClick={() => {
                setCreateError(null);
                setDialogOpen(true);
              }}
            >
              + Create organization
            </Button>
          ) : null
        }
      />
      {canCreate ? (
        <CreateOrganizationDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          onSubmit={handleCreate}
          isPending={createMutation.isPending}
          error={createError}
        />
      ) : null}
    </>
  );
}
