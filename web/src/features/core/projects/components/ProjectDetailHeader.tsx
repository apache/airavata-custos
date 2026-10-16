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

import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import type { ProjectResponse } from "@/generated/core/types.gen";
import { formatDate } from "@/shared/format";
import { statusLabel } from "@/shared/ui/FormDialog";
import { MetaItem, MetaRow } from "@/shared/ui/MetaRow";
import { Building2, Calendar, UserSquare } from "lucide-react";
import { originationLabel } from "../format";

export type ProjectDetailHeaderProps = {
  project: ProjectResponse;
};

export function ProjectDetailHeader({ project }: ProjectDetailHeaderProps) {
  const pi = project.project_pi_display_name ?? project.project_pi_id;
  const tone: "success" | "warning" | "danger" =
    project.status === "ACTIVE" ? "success" : project.status === "INACTIVE" ? "warning" : "danger";
  return (
    <header className="min-w-0 space-y-4">
      <h1 className="font-display text-[28px] font-bold leading-tight text-foreground [overflow-wrap:anywhere]">
        {project.title}
      </h1>
      <div className="flex flex-wrap items-center gap-3">
        <MetaItem
          variant="status"
          tone={tone}
          value={statusLabel(project.status)}
          className="py-1.5"
        />
        <MetaRow className="max-lg:flex max-lg:gap-x-6 max-lg:rounded-none max-lg:bg-transparent max-lg:p-0 max-lg:[&>*:not(:first-child)]:ml-0 max-lg:[&>*:not(:first-child)]:border-l-0 max-lg:[&>*:not(:first-child)]:pl-0 max-sm:flex-col max-sm:items-start">
          <MetaItem
            icon={UserSquare}
            label="PI"
            value={
              <span className="block max-w-[32ch] truncate" title={pi}>
                {pi}
              </span>
            }
          />
          <MetaItem
            icon={Building2}
            label="Origination"
            value={originationLabel(project.origination)}
          />
          <MetaItem
            icon={Calendar}
            label="Created"
            value={
              <Timestamp iso={project.created_time}>{formatDate(project.created_time)}</Timestamp>
            }
          />
        </MetaRow>
      </div>
    </header>
  );
}
