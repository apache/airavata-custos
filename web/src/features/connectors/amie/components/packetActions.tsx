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

import type { PacketResponse } from "@/generated/amie/types.gen";
import type { DataTableColumn } from "@/shared/ui/DataTable";
import { Button } from "@/shared/ui/button";
import { toggleId } from "@/shared/users-admin/permissions";

// Write actions; omitted when the caller lacks amie:packets:write.
export type PacketActions = {
  onRetryPackets?: (ids: string[]) => void;
  onResolvePackets?: (ids: string[]) => void;
};

// Retry and mark-processed buttons for one packet, or for the selection when bulk.
export function PacketActionButtons({
  ids,
  bulk = false,
  onRetryPackets,
  onResolvePackets,
}: PacketActions & { ids: string[]; bulk?: boolean }) {
  const props = { type: "button", variant: "outline", size: bulk ? "default" : "sm", disabled: ids.length === 0 } as const;
  return (
    <>
      {onRetryPackets ? (
        <Button {...props} onClick={() => onRetryPackets(ids)}>
          {bulk ? "Retry selected" : "Retry"}
        </Button>
      ) : null}
      {onResolvePackets ? (
        <Button {...props} onClick={() => onResolvePackets(ids)}>
          Mark processed
        </Button>
      ) : null}
    </>
  );
}

// Checkbox column for bulk actions; "select all" covers the rows on the current page.
export function selectColumn(
  rows: PacketResponse[],
  selected: Set<string>,
  onSelectChange: (selected: Set<string>) => void,
): DataTableColumn<PacketResponse> {
  const ids = rows.map((r) => r.id);
  const allSelected = ids.length > 0 && ids.every((id) => selected.has(id));
  return {
    key: "select",
    header: (
      <input
        type="checkbox"
        aria-label="Select all packets on this page"
        checked={allSelected}
        onChange={() => onSelectChange(new Set(allSelected ? [] : ids))}
      />
    ),
    width: "32px",
    interactive: true,
    cell: (row) => (
      <input
        type="checkbox"
        aria-label={`Select ${row.amie_id}`}
        checked={selected.has(row.id)}
        onChange={() => onSelectChange(toggleId(selected, row.id))}
      />
    ),
  };
}
