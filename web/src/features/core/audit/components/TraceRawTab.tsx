/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

"use client";

import { Button } from "@/shared/ui/button";
import { Check, Copy } from "lucide-react";
import * as React from "react";
import type { TraceDetail } from "../types";
import { useCopy } from "./primitives/CopyValue";
import { JsonBlock } from "./primitives/JsonBlock";

export type TraceRawTabProps = {
  detail: TraceDetail;
  stepCount: number;
};

export function TraceRawTab({ detail, stepCount }: TraceRawTabProps) {
  const [copied, copy] = useCopy();
  const text = React.useMemo(() => JSON.stringify(detail, null, 2), [detail]);

  return (
    <div className="max-w-[920px]">
      <div className="mb-2.5 flex items-center justify-between">
        <div className="text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
          TRACE JSON{" "}
          <span className="font-medium normal-case tracking-normal">
            · {stepCount} {stepCount === 1 ? "row" : "rows"}, {detail.deliveries.length}{" "}
            {detail.deliveries.length === 1 ? "delivery" : "deliveries"}
          </span>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => copy(text)}
          aria-label={copied ? "Copied JSON" : "Copy trace JSON"}
        >
          {copied ? (
            <Check className="h-3.5 w-3.5 text-[color:var(--tone-ok-fg)]" />
          ) : (
            <Copy className="h-3.5 w-3.5" />
          )}
          <span>{copied ? "Copied" : "Copy JSON"}</span>
        </Button>
      </div>
      <JsonBlock text={text} label="Trace JSON" data-testid="trace-raw-json" />
    </div>
  );
}
