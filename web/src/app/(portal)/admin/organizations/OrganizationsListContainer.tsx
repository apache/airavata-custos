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
import { CreateOrganizationDialog } from "@/features/core/organizations/components/CreateOrganizationDialog";
import { OrganizationDrawer } from "@/features/core/organizations/components/OrganizationDrawer";
import { OrganizationsList } from "@/features/core/organizations/components/OrganizationsList";
import { useOrganizations } from "@/features/core/organizations/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";

const PAGE_SIZE = 50;

export function OrganizationsListContainer() {
  const ability = useAbility();
  const canCreate = ability.can("write", "Organization");

  const [page, setPage] = React.useState(1);
  const [selectedId, setSelectedId] = React.useState<string>();

  const query = useOrganizations({ limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE });
  return (
    <>
      <OrganizationsList
        rows={query.data?.items ?? []}
        isLoading={query.isLoading}
        error={query.error}
        onRetry={() => query.refetch()}
        page={page}
        pageSize={PAGE_SIZE}
        total={query.data?.total ?? 0}
        onPageChange={setPage}
        onRowClick={(row) => setSelectedId(row.id)}
        headerCta={canCreate ? <CreateOrganizationDialog /> : null}
      />
      <OrganizationDrawer organizationId={selectedId} onClose={() => setSelectedId(undefined)} />
    </>
  );
}
