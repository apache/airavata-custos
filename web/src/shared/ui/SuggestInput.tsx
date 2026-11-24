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

import { Autocomplete } from "@base-ui/react/autocomplete";
import * as React from "react";
import { Input } from "./input";

type Suggestion = { value: string; detail?: string };

type SuggestInputProps = Omit<React.ComponentProps<typeof Input>, "value" | "onChange"> & {
  value: string;
  onValueChange: (value: string) => void;
  suggestions: Suggestion[];
  emptyMessage?: string;
};

// Suggestions arrive pre-filtered by the caller (typically a server search), so Base UI does not filter.
export function SuggestInput({
  value,
  onValueChange,
  suggestions,
  emptyMessage,
  ...inputProps
}: SuggestInputProps) {
  const [open, setOpen] = React.useState(false);
  return (
    <Autocomplete.Root
      items={suggestions}
      itemToStringValue={(s) => s.value}
      value={value}
      onValueChange={(next) => {
        onValueChange(next);
        if (!next) setOpen(true);
      }}
      mode="none"
      filter={null}
      autoHighlight="always"
      open={open}
      onOpenChange={setOpen}
      openOnInputClick
    >
      <Autocomplete.Input render={<Input {...inputProps} />} onFocus={() => setOpen(true)} />
      <Autocomplete.Portal>
        <Autocomplete.Positioner sideOffset={4} className="isolate z-50">
          <Autocomplete.Popup className="max-h-(--available-height) w-(--anchor-width) origin-(--transform-origin) overflow-y-auto rounded-lg bg-popover p-1 text-popover-foreground shadow-md ring-1 ring-foreground/10 duration-100 data-[ending-style]:opacity-0 data-[starting-style]:opacity-0 empty:hidden">
            <Autocomplete.List>
              {(s: Suggestion) => (
                <Autocomplete.Item
                  key={s.value}
                  value={s}
                  className="flex cursor-default select-none flex-col rounded-md px-2 py-1.5 text-sm outline-hidden data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
                >
                  <span>{s.value}</span>
                  {s.detail ? (
                    <span className="text-muted-foreground text-xs">{s.detail}</span>
                  ) : null}
                </Autocomplete.Item>
              )}
            </Autocomplete.List>
            {emptyMessage ? (
              <Autocomplete.Empty className="px-2 py-1.5 text-muted-foreground text-sm">
                {emptyMessage}
              </Autocomplete.Empty>
            ) : null}
          </Autocomplete.Popup>
        </Autocomplete.Positioner>
      </Autocomplete.Portal>
    </Autocomplete.Root>
  );
}
