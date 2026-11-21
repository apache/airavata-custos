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

import type { ComputeAllocation } from "@/generated/core/zod.gen";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import * as React from "react";
import { toast } from "sonner";
import { useSetAllocationStatus } from "../queries";

// AllocationStatusButton deactivates an active allocation after a confirm, and reactivates an inactive one at once.
export function AllocationStatusButton({ allocation }: { allocation: ComputeAllocation }) {
  const can = usePermissions();
  const status = useSetAllocationStatus(allocation.id);
  const [confirming, setConfirming] = React.useState(false);
  const active = allocation.status === "ACTIVE";
  return (
    <>
      <GuardedButton
        variant={active ? "destructive" : "outline"}
        allowed={can.allocationsWrite}
        reason={requires("allocationsWrite")}
        onClick={() => {
          if (!active)
            return status.mutate("ACTIVE", { onError: (err) => toast.error(err.message) });
          status.reset();
          setConfirming(true);
        }}
        disabled={status.isPending}
      >
        {active ? "Deactivate" : "Reactivate"}
      </GuardedButton>
      <ConfirmDialog
        open={confirming}
        onOpenChange={setConfirming}
        title={`Deactivate ${allocation.name}?`}
        confirmLabel="Deactivate allocation"
        pending={status.isPending}
        error={status.error}
        onConfirm={() => status.mutate("INACTIVE", { onSuccess: () => setConfirming(false) })}
      >
        Its members lose access until it is reactivated.
      </ConfirmDialog>
    </>
  );
}
