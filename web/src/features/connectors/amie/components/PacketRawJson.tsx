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

import { JsonView, defaultStyles } from "react-json-view-lite";
import "react-json-view-lite/dist/index.css";

export type PacketRawJsonProps = {
  rawJson: string;
};

export default function PacketRawJson({ rawJson }: PacketRawJsonProps) {
  let parsed: unknown;
  try {
    parsed = JSON.parse(rawJson);
  } catch {
    return (
      <pre className="overflow-x-auto rounded-md border bg-muted/20 p-3 text-xs">{rawJson}</pre>
    );
  }
  return (
    <div className="overflow-x-auto rounded-md border bg-background p-3 text-xs">
      <JsonView data={parsed as object} style={defaultStyles} />
    </div>
  );
}
