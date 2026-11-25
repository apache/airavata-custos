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

import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import * as React from "react";

const PRESET_DAYS = [7, 30, 90];

export function DaysRangePicker({
  value,
  onChange,
  label,
}: {
  value: number;
  onChange: (days: number) => void;
  label: string;
}) {
  const [customOpen, setCustomOpen] = React.useState(() => !PRESET_DAYS.includes(value));
  const [customDraft, setCustomDraft] = React.useState(String(value));
  const customInputId = React.useId();
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    setCustomDraft(String(value));
    setCustomOpen(!PRESET_DAYS.includes(value));
  }, [value]);

  const commitCustomDays = (raw: string) => {
    const days = Number(raw);
    if (Number.isInteger(days) && days >= 1 && days <= 365) {
      setError("");
      onChange(days);
    } else setError("Enter a whole number from 1 to 365.");
  };

  return (
    <fieldset className="flex flex-wrap items-center gap-2" aria-label={label}>
      {PRESET_DAYS.map((days) => (
        <Button
          key={days}
          type="button"
          size="sm"
          variant={!customOpen && value === days ? "brand" : "outline"}
          aria-pressed={!customOpen && value === days}
          onClick={() => {
            setError("");
            setCustomOpen(false);
            onChange(days);
          }}
        >
          {days} days
        </Button>
      ))}
      <Button
        type="button"
        size="sm"
        variant={customOpen ? "brand" : "outline"}
        aria-pressed={customOpen}
        onClick={() => {
          setCustomOpen(true);
          setCustomDraft(String(value));
        }}
      >
        Custom
      </Button>
      {customOpen ? (
        <label
          htmlFor={customInputId}
          className="flex shrink-0 items-center gap-1 text-xs text-muted-foreground"
        >
          <Input
            id={customInputId}
            type="number"
            min={1}
            max={365}
            aria-invalid={Boolean(error)}
            aria-describedby={error ? `${customInputId}-error` : undefined}
            value={customDraft}
            onChange={(event) => setCustomDraft(event.target.value)}
            onBlur={(event) => commitCustomDays(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") event.currentTarget.blur();
            }}
            aria-label={`${label} custom days`}
            className="h-8 w-20"
          />
          days
        </label>
      ) : null}
      {error ? (
        <span id={`${customInputId}-error`} role="alert" className="text-xs text-destructive">
          {error}
        </span>
      ) : null}
    </fieldset>
  );
}
