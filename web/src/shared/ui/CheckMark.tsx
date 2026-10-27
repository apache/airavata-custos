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
import { Check } from "lucide-react";

// A read-only checkbox; the label names what it marks.
export function CheckMark({ checked, label }: { checked: boolean; label: string }) {
  return (
    <span
      role="img"
      aria-label={label}
      title={label}
      className={cn(
        "mx-auto flex size-4 items-center justify-center rounded-[4px]",
        checked ? "bg-foreground text-background" : "border border-muted-foreground/40",
      )}
    >
      {checked ? <Check className="size-3" strokeWidth={3} aria-hidden /> : null}
    </span>
  );
}
