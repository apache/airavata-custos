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

import type { TraceEvent, TraceSummary } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import { Check, Copy } from "lucide-react";
import * as React from "react";
import { useCopy } from "./primitives/CopyValue";

export type TraceRawTabProps = {
  trace: TraceSummary;
  spans: TraceEvent[];
};

type Highlighted = React.ReactNode[];

// Lex the stringified JSON once and tag tokens with their semantic colors so
// indent and punctuation render verbatim alongside coloured keys/values.
function highlightJson(text: string): { text: string; nodes: Highlighted } {
  const parts: Highlighted = [];
  const re =
    /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(?:true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g;
  let last = 0;
  let m: RegExpExecArray | null;
  let key = 0;
  // biome-ignore lint/suspicious/noAssignInExpressions: idiomatic regex loop
  while ((m = re.exec(text))) {
    if (m.index > last) parts.push(text.slice(last, m.index));
    const tok = m[0];
    const isKey = tok.startsWith('"') && tok.trimEnd().endsWith(":");
    const isStr = tok.startsWith('"') && !isKey;
    const isBool = /^(true|false|null)$/.test(tok);
    const color = isKey
      ? "var(--syntax-key)"
      : isStr
        ? "var(--syntax-str)"
        : isBool
          ? "var(--syntax-bool)"
          : "var(--syntax-num)";
    const tokenType = isKey ? "key" : isStr ? "str" : isBool ? "bool" : "num";
    parts.push(
      <span key={`tk-${key++}`} data-token={tokenType} style={{ color }}>
        {tok}
      </span>,
    );
    last = m.index + tok.length;
  }
  if (last < text.length) parts.push(text.slice(last));
  return { text, nodes: parts };
}

export function TraceRawTab({ trace, spans }: TraceRawTabProps) {
  const [copied, copy] = useCopy();

  const { text, nodes } = React.useMemo(
    () => highlightJson(JSON.stringify({ ...trace, spans }, null, 2)),
    [trace, spans],
  );

  return (
    <div className="max-w-[920px]">
      <div className="mb-2.5 flex items-center justify-between">
        <div className="text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
          TRACE JSON{" "}
          <span className="font-medium normal-case tracking-normal">· {spans.length} spans</span>
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
      <pre
        data-testid="trace-raw-json"
        className="m-0 overflow-auto rounded-[10px] border border-[color:var(--border)] bg-[color:var(--muted-2)] p-4 font-mono text-xs leading-[1.6] text-foreground"
      >
        {nodes}
      </pre>
    </div>
  );
}
