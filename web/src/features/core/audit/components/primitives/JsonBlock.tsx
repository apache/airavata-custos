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

import { cn } from "@/lib/utils";
import { useScrollStop } from "@/shared/ui/focus";
import * as React from "react";

// Themed token colours, so every JSON view clears 4.5:1 on --muted in both themes.
const TOKEN_COLORS = {
  key: "var(--foreground)",
  str: "var(--link)",
  num: "var(--muted-foreground)",
  bool: "var(--muted-foreground)",
} as const;

const TOKEN =
  /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(?:true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g;

// Tags the tokens of JSON text so indent and punctuation render verbatim beside coloured keys and values.
function highlight(text: string): React.ReactNode[] {
  const parts: React.ReactNode[] = [];
  let last = 0;
  for (const m of text.matchAll(TOKEN)) {
    const [tok] = m;
    const at = m.index ?? 0;
    if (at > last) parts.push(text.slice(last, at));
    const type = tok.startsWith('"')
      ? m[2] !== undefined
        ? "key"
        : "str"
      : /^(true|false|null)$/.test(tok)
        ? "bool"
        : "num";
    parts.push(
      <span
        key={at}
        data-token={type}
        className={type === "key" ? "font-semibold" : undefined}
        style={{ color: TOKEN_COLORS[type] }}
      >
        {tok}
      </span>,
    );
    last = at + tok.length;
  }
  if (last < text.length) parts.push(text.slice(last));
  return parts;
}

export type JsonBlockProps = {
  // JSON text as it should read, already indented.
  text: string;
  label: string;
  className?: string;
  "data-testid"?: string;
};

export function JsonBlock({ text, label, className, ...rest }: JsonBlockProps) {
  const nodes = React.useMemo(() => highlight(text), [text]);
  const scrollStop = useScrollStop<HTMLElement>();
  return (
    <section
      data-testid={rest["data-testid"]}
      {...scrollStop}
      aria-label={label}
      className={cn(
        "overflow-auto rounded-[10px] border border-[color:var(--border)] bg-[color:var(--muted)] p-4",
        className,
      )}
    >
      <pre className="m-0 font-mono text-xs leading-[1.6] text-foreground">{nodes}</pre>
    </section>
  );
}
