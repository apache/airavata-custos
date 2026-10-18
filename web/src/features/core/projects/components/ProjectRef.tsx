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

import type { ProjectResponse } from "@/generated/core/zod.gen";
import { Badge } from "@/shared/ui/badge";

export function ProjectRef({ project }: { project: Pick<ProjectResponse, "ref"> }) {
  return (
    <Badge variant="outline" className="h-4 shrink-0 px-1.5 py-0 font-mono text-[10px]">
      {project.ref}
    </Badge>
  );
}
