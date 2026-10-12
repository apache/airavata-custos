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

import { ClustersTab } from "@/features/core/clusters/components/ClustersTab";
import { CreateClusterDialog } from "@/features/core/clusters/components/CreateClusterDialog";
import { useClusters } from "@/features/core/clusters/queries";
import { CreateResourceDialog } from "@/features/core/resources/components/CreateResourceDialog";
import { ResourcesTab } from "@/features/core/resources/components/ResourcesTab";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useShallowSearchParams } from "@/shared/hooks/useShallowSearchParams";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { WriteOnlyNotice } from "@/shared/ui/EmptyState";
import { NotPermitted } from "@/shared/ui/ErrorState";
import { TabsRouter, type TabsRouterTab } from "@/shared/ui/TabsRouter";
import { DeclaredParams } from "@/shared/ui/ViewParamsDropped";

export function ResourcesClustersTabs() {
  const ability = useAbility();
  const tab = useShallowSearchParams().get("tab");

  // Each tab's drawer params; the other tab's are dropped so they never reopen a drawer on return.
  const tabs: Array<TabsRouterTab & { params: string[] }> = [];
  if (ability.can("read", "Cluster")) {
    tabs.push({
      value: "clusters",
      label: "Clusters",
      content: <ClustersTab />,
      params: ["cluster", "account"],
    });
  }
  if (ability.can("read", "Allocation")) {
    tabs.push({
      value: "resources",
      label: "Resources",
      content: <ResourcesTab />,
      params: ["resource", "rates"],
    });
  }
  const active = tabs.find((t) => t.value === tab) ?? tabs[0];

  // A create the caller can make without reading its list sits in the header.
  const registerCluster = ability.can("write", "Cluster") && !ability.can("read", "Cluster");
  const newResource = ability.can("write", "Allocation") && !ability.can("read", "Allocation");
  const clusters = useClusters(newResource && ability.can("read", "Cluster")).data;
  if (!active && !registerCluster && !newResource)
    return <NotPermitted title={ROUTE_LABELS["/admin/resources"]} />;

  return (
    <div className="space-y-6">
      <DeclaredParams keys={active ? ["tab", ...active.params] : []} />
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ROUTE_LABELS["/admin/resources"]}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Compute clusters and their allocation resources and rates.
          </p>
        </div>
        {registerCluster || newResource ? (
          <div className="flex flex-wrap gap-2">
            {registerCluster ? <CreateClusterDialog /> : null}
            {newResource ? <CreateResourceDialog clusters={clusters} /> : null}
          </div>
        ) : null}
      </div>
      {tabs[0] ? (
        <TabsRouter defaultValue={tabs[0].value} tabs={tabs} />
      ) : (
        <WriteOnlyNotice
          things={
            registerCluster && newResource
              ? "clusters or resources"
              : registerCluster
                ? "clusters"
                : "resources"
          }
        />
      )}
    </div>
  );
}
