"use client";

import { JsonView, defaultStyles } from "react-json-view-lite";
import "react-json-view-lite/dist/index.css";

export type PacketRawJsonProps = {
  rawJson: string;
};

export default function PacketRawJson({ rawJson }: PacketRawJsonProps) {
  let parsed: unknown;
  try {
    parsed = JSON.parse(rawJson);
  } catch {
    return (
      <pre className="overflow-x-auto rounded-md border bg-muted/20 p-3 text-xs">{rawJson}</pre>
    );
  }
  return (
    <div className="overflow-x-auto rounded-md border bg-background p-3 text-xs">
      <JsonView data={parsed as object} style={defaultStyles} />
    </div>
  );
}
