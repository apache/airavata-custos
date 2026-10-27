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

import { cn } from "@/lib/utils";
import type { LucideIcon } from "lucide-react";

// A record's kind icon, green only when active; the hover title names the actual state.
export function StatusIcon({
  icon: Icon,
  active,
  label,
}: {
  icon: LucideIcon;
  active: boolean;
  label: string;
}) {
  return (
    <Icon
      role="img"
      aria-label={label}
      className={cn("size-4 shrink-0", active ? "text-emerald-600" : "text-muted-foreground/50")}
    >
      <title>{label}</title>
    </Icon>
  );
}
