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

import { useAmieAction } from "@/features/connectors/amie/queries";
import { pluralize } from "@/features/connectors/amie/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { confirmToast } from "@/shared/ui/sonner";

// Retry and mark-processed handlers shared by the inbox, failed queue and packet drawer;
// empty without amie:packets:write so callers hide the controls.
export function usePacketActions(onDone: () => void) {
  const canWrite = useAbility().can("write", "AmiePacket");
  const retry = useAmieAction("retryPacket");
  const resolve = useAmieAction("resolvePacket");
  if (!canWrite) return {};
  return {
    onRetryPackets: (ids: string[]) => retry.mutate(ids, { onSuccess: onDone }),
    onResolvePackets: (ids: string[]) =>
      confirmToast(
        `Mark ${ids.length} ${pluralize("packet", ids.length)} processed without the handler?`,
        "Mark processed",
        () => resolve.mutate(ids, { onSuccess: onDone }),
      ),
  };
}
