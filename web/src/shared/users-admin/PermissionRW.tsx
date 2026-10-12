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
import { actionChipClass } from "./permissions";

const CHIP = "inline-flex h-6 items-center justify-center rounded px-2 text-xs font-medium";

// Only granted actions show, so "not granted" never rests on styling alone.
export function PermissionRW({ read, write }: { read: boolean; write: boolean }) {
  return (
    <div className="flex items-center gap-1">
      {read ? <span className={cn(CHIP, actionChipClass("read"))}>read</span> : null}
      {write ? <span className={cn(CHIP, actionChipClass("write"))}>write</span> : null}
    </div>
  );
}
