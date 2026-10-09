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

import { Check, Copy } from "lucide-react";
import * as React from "react";
import { cn } from "@/lib/utils";

export type CopyValueProps = {
  value: string;
  label?: string;
  explicit?: boolean;
  className?: string;
  children?: React.ReactNode;
};

const COPIED_RESET_MS = 1100;

// Copies best-effort and flashes `copied` either way so the click always gets feedback.
export function useCopy(): [boolean, (text: string) => void] {
  const [copied, setCopied] = React.useState(false);
  const timer = React.useRef<ReturnType<typeof setTimeout>>(undefined);
  React.useEffect(() => () => clearTimeout(timer.current), []);
  const copy = React.useCallback((text: string) => {
    void navigator.clipboard
      .writeText(text)
      .catch(() => {})
      .finally(() => {
        setCopied(true);
        clearTimeout(timer.current);
        timer.current = setTimeout(() => setCopied(false), COPIED_RESET_MS);
      });
  }, []);
  return [copied, copy];
}

export function CopyValue({ value, label, explicit = false, className, children }: CopyValueProps) {
  const [copied, copy] = useCopy();

  const handleCopy = (event: React.MouseEvent) => {
    // Copying a trace id inside a list row should not also open the drawer.
    event.stopPropagation();
    event.preventDefault();
    copy(value);
  };

  const Icon = copied ? Check : Copy;
  const ariaLabel = `Copy ${label ?? value}`;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 font-mono text-xs text-foreground",
        className,
      )}
    >
      <span className="truncate">{children ?? value}</span>
      <button
        type="button"
        onClick={handleCopy}
        aria-label={ariaLabel}
        className={cn(
          "inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-md text-muted-foreground",
          "hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          copied ? "text-[color:var(--custos-green-700)]" : null,
          explicit ? "opacity-100" : "opacity-60",
        )}
      >
        <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
    </span>
  );
}
