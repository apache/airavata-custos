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

import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import { activityStatus } from "../lib";
import type { UserActivityRow } from "../schemas";

function ActivityPill({
  variant,
  label,
}: {
  variant: StatusBadgeVariant;
  label?: string;
}) {
  return (
    <StatusBadge variant={variant} label={label} dot className="[&>span:first-child]:size-1.5" />
  );
}

export function ActivityStatus({
  user,
  windowDays,
}: { user: UserActivityRow; windowDays: number }) {
  const status = activityStatus(user, windowDays);
  if (status === "active") return <ActivityPill variant="active" />;
  if (status === "dormant") return <ActivityPill variant="warning" label="Dormant" />;
  if (status === "unknown") return <ActivityPill variant="inactive" label="Unknown" />;
  return <ActivityPill variant="deleted" label="Never signed in" />;
}
