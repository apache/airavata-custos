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
import { ResourcesTab } from "@/features/core/resources/components/ResourcesTab";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { TabsRouter, type TabsRouterTab } from "@/shared/ui/TabsRouter";

export function ResourcesClustersTabs() {
  const ability = useAbility();

  const tabs: TabsRouterTab[] = [];
  if (ability.can("read", "Cluster")) {
    tabs.push({ value: "clusters", label: "Clusters", content: <ClustersTab /> });
  }
  if (ability.can("read", "Allocation")) {
    tabs.push({ value: "resources", label: "Resources", content: <ResourcesTab /> });
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-[28px] font-bold leading-tight">Resources &amp; Clusters</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Compute clusters and their allocation resources and rates.
        </p>
      </div>
      {tabs[0] ? <TabsRouter defaultValue={tabs[0].value} tabs={tabs} /> : null}
    </div>
  );
}
