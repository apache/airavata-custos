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

import { ReconciliationQueue } from "@/features/connectors/amie/components/ReconciliationQueue";
import { useAmieAction, useUnmapped } from "@/features/connectors/amie/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";

export function ReconciliationContainer() {
  const unmappedQuery = useUnmapped();
  const rows = unmappedQuery.data?.packets ?? [];
  const canWrite = useAbility().can("write", "AmieUnmapped");
  const link = useAmieAction("linkUnmapped");

  return (
    <ReconciliationQueue
      rows={rows}
      total={unmappedQuery.data?.total ?? rows.length}
      isLoading={unmappedQuery.isLoading}
      error={unmappedQuery.error}
      onRefresh={() => unmappedQuery.refetch()}
      onLink={canWrite ? (id) => link.mutate([id]) : undefined}
    />
  );
}
