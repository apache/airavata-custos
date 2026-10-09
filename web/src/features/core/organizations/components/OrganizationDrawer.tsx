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

import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { useOrganization } from "../queries";

export function OrganizationDrawer({
  organizationId,
  onClose,
}: {
  organizationId: string | undefined;
  onClose: () => void;
}) {
  const query = useOrganization(organizationId);
  return (
    <RecordDrawer
      title={query.data?.name ?? "Organization"}
      id={organizationId}
      onClose={onClose}
      query={query}
      fields={(org) => ({
        Name: org.name,
        "Originated ID": <Mono>{org.originated_id || "None"}</Mono>,
        ID: <Mono>{org.id}</Mono>,
      })}
    />
  );
}
