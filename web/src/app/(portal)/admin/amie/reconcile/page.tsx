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
import { ReconciliationContainer } from "./ReconciliationContainer";

export default function AmieReconcilePage() {
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">Reconciliation queue</h1>
        <p className="text-sm text-muted-foreground">
          Decoded packets that couldn't be mapped to a domain entity. Link to an existing project,
          account, or person — or skip with a reason.
        </p>
      </header>
      <AmieNav />
      <ReconciliationContainer />
    </div>
  );
}
