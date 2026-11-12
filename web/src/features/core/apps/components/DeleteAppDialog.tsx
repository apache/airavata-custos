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

import { ConfirmDialog } from "@/shared/ui/ConfirmDialog";
import type { App } from "../api";
import { useAppAuthorizationCount, useDeleteApp } from "../queries";

export type DeleteAppDialogProps = {
  app: App;
  onClose: () => void;
};

export function DeleteAppDialog({ app, onClose }: DeleteAppDialogProps) {
  const count = useAppAuthorizationCount(app.id);
  const remove = useDeleteApp();
  const people = count.data ?? 0;

  return (
    <ConfirmDialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title={`Delete ${app.name}?`}
      confirmLabel="Delete app"
      pending={remove.isPending || count.isLoading}
      error={remove.error ?? count.error}
      onConfirm={() => remove.mutate(app.id, { onSuccess: onClose })}
    >
      {people > 0
        ? `${people} ${people === 1 ? "person has" : "people have"} allowed this app. They all lose it, and the app can't act for them again. `
        : null}
      This can't be undone. The app key {app.key} can't be used again.
    </ConfirmDialog>
  );
}
