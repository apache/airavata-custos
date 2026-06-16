"use client";

import { Check, Copy } from "lucide-react";
import * as React from "react";
import { Button } from "@/shared/ui/button";
import type { Span, Trace } from "../types";

export type TraceRawTabProps = {
  trace: Trace;
  spans: Span[];
};

type Highlighted = React.ReactNode[];

// Lex the stringified JSON once and tag tokens with their semantic colors so
// indent and punctuation render verbatim alongside coloured keys/values.
function highlightJson(obj: unknown): { text: string; nodes: Highlighted } {
  const text = JSON.stringify(obj, null, 2);
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
  const [copied, setCopied] = React.useState(false);
  const timerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const json = React.useMemo(
    () => ({
      trace_id: trace.trace_id,
      source: trace.source,
      status: trace.status,
      root_action: trace.root_name,
      span_count: trace.span_count,
      spans: spans.map((s) => ({
        span_id: s.span_id,
        parent_span_id: s.parent_span_id ?? null,
        action: s.name,
        source: trace.source,
        status: s.status,
        started_at: s.start_time,
        ended_at: s.end_time ?? null,
        ...(s.status_message ? { status_message: s.status_message } : {}),
        summary: "",
        attributes: (s.attributes ?? {}) as Record<string, unknown>,
      })),
    }),
    [trace, spans],
  );

  const { text, nodes } = React.useMemo(() => highlightJson(json), [json]);

  const onCopy = () => {
    void (async () => {
      try {
        await navigator.clipboard.writeText(text);
      } catch {
        // Best-effort copy; still surface the check.
      }
      setCopied(true);
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setCopied(false), 1100);
    })();
  };

  return (
    <div className="max-w-[920px]">
      <div className="mb-2.5 flex items-center justify-between">
        <div className="text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
          TRACE JSON{" "}
          <span className="font-medium normal-case tracking-normal">
            · {spans.length} spans
          </span>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={onCopy}
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
