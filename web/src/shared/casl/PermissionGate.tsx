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

import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { normalizeRouteId } from "@/shared/layout/nav";
import { NotPermitted } from "@/shared/ui/ErrorState";
import { useParams } from "next/navigation";
import type * as React from "react";
import { useAbility } from "./AbilityProvider";

// Renders children when the caller can read any of the subjects, else the titled NotPermitted.
// A gated detail route's crumb reads `crumb` in place of the record id it cannot show.
export function PermissionGate({
  read,
  title,
  crumb,
  children,
}: { read: string[]; title: string; crumb?: string; children: React.ReactNode }) {
  const ability = useAbility();
  const allowed = read.some((subject) => ability.can("read", subject));
  const id = Object.values(useParams() ?? {}).find((v) => typeof v === "string");
  useBreadcrumbLabel(allowed ? undefined : normalizeRouteId(id), crumb);
  return allowed ? children : <NotPermitted title={title} />;
}
