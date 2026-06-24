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

import { AmieNav } from "../AmieNav";
import { FailedQueueContainer } from "./FailedQueueContainer";

export default function AmieFailedPage() {
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">Failed packet queue</h1>
        <p className="text-sm text-muted-foreground">
          Packets the handler couldn't process. Retry one-shot or in bulk; flag stuck packets with a
          manual resolution reason.
        </p>
      </header>
      <AmieNav />
      <FailedQueueContainer />
    </div>
  );
}
