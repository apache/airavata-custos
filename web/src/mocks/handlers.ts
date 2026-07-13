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

import type { RequestHandler } from "msw";
import { allocationsHandlers } from "./handlers/allocations";
import { amieHandlers } from "./handlers/amie";
import { clustersHandlers } from "./handlers/clusters";
import { healthzHandlers } from "./handlers/healthz";
import { identityHandlers } from "./handlers/identity";
import { organizationsHandlers } from "./handlers/organizations";
import { privilegesHandlers } from "./handlers/privileges";
import { projectsHandlers } from "./handlers/projects";
import { resourcesHandlers } from "./handlers/resources";
import { tracesHandlers } from "./handlers/traces";

export const handlers: RequestHandler[] = [
  ...healthzHandlers,
  ...privilegesHandlers,
  ...identityHandlers,
  ...projectsHandlers,
  ...organizationsHandlers,
  ...allocationsHandlers,
  ...tracesHandlers,
  ...amieHandlers,
  ...clustersHandlers,
  ...resourcesHandlers,
];
