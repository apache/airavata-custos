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

import { useRouter } from "next/navigation";
import * as React from "react";
import { useUserEmailSearch } from "@/features/core/users/queries";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { SuggestInput } from "@/shared/ui/SuggestInput";
import { useCreateProject } from "../queries";

export type CreateProjectDialogProps = {
  onOpenChange: (open: boolean) => void;
};

export function CreateProjectDialog({ onOpenChange }: CreateProjectDialogProps) {
  const router = useRouter();
  const create = useCreateProject();
  const [title, setTitle] = React.useState("");
  const [piEmail, setPiEmail] = React.useState("");
  const pi = useUserEmailSearch(piEmail, true);

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New project</DialogTitle>
          <DialogDescription>
            Create a project; add allocations and members after.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!title.trim() || !pi.match) return;
            create.mutate(
              {
                title: title.trim(),
                origination: "internal",
                project_pi_id: pi.match.id,
              },
              { onSuccess: (project) => router.push(`/projects/${project.id}`) },
            );
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="new-project-title">Title</Label>
            <Input
              id="new-project-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              maxLength={200}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-project-pi">PI</Label>
            <SuggestInput
              id="new-project-pi"
              type="email"
              value={piEmail}
              onValueChange={setPiEmail}
              suggestions={pi.suggestions}
              emptyMessage={pi.emptyMessage}
              placeholder="pi@example.edu"
              required
            />
          </div>
          {create.error ? (
            <p className="text-destructive text-sm">{create.error.message}</p>
          ) : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={create.isPending || !title.trim() || !pi.match}>
              {create.isPending ? "Creating…" : "Create project"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
