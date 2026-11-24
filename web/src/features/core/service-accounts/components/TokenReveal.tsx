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

import * as React from "react";
import { Button } from "@/shared/ui/button";
import { CopyValue } from "@/shared/ui/CopyValue";
import { DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";

export type TokenRevealProps = {
  accountName: string;
  token: string;
  onDone: () => void;
};

export function TokenReveal({ accountName, token, onDone }: TokenRevealProps) {
  const [stored, setStored] = React.useState(false);
  return (
    <>
      <DialogHeader>
        <DialogTitle>Token for {accountName}</DialogTitle>
        <DialogDescription>
          Give this to the tool that uses this service account. It is shown only now. If it is lost,
          issue a new secret.
        </DialogDescription>
      </DialogHeader>
      <div className="min-w-0 space-y-4 text-sm">
        <CopyValue value={token} label="token" explicit>
          <span className="whitespace-normal break-all">{token}</span>
        </CopyValue>
        <label className="flex items-center gap-2">
          <input type="checkbox" checked={stored} onChange={(e) => setStored(e.target.checked)} />I
          have stored the token
        </label>
      </div>
      <DialogFooter>
        <Button onClick={onDone} disabled={!stored}>
          Done
        </Button>
      </DialogFooter>
    </>
  );
}
