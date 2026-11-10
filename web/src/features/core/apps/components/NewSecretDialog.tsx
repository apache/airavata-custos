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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import type { App } from "../api";
import { useIssueAppSecret } from "../queries";
import { TokenReveal } from "./TokenReveal";

export type NewSecretDialogProps = {
  app: App;
  onClose: () => void;
};

export function NewSecretDialog({ app, onClose }: NewSecretDialogProps) {
  const issue = useIssueAppSecret();
  const result = issue.data;

  return (
    <Dialog
      open
      onOpenChange={(next) => {
        // The token is shown once, so only Done closes this state.
        if (!next && !issue.isPending && !result) onClose();
      }}
    >
      <DialogContent>
        {result ? (
          <TokenReveal appName={app.name} token={result.token} onDone={onClose} />
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>New secret for {app.name}?</DialogTitle>
              <DialogDescription>
                The current token stops working at once. The tool must switch to the new token.
              </DialogDescription>
            </DialogHeader>
            {issue.error ? (
              <p role="alert" className="text-sm text-destructive">
                {issue.error.message}
              </p>
            ) : null}
            <DialogFooter>
              <Button variant="outline" onClick={onClose} disabled={issue.isPending}>
                Cancel
              </Button>
              <Button onClick={() => issue.mutate(app.id)} disabled={issue.isPending}>
                Issue new secret
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
