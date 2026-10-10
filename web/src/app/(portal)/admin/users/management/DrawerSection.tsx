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

import type * as React from "react";

// A titled section of the user drawer: a status line while loading, failed or empty, else its content; the footer always shows.
export function DrawerSection({
  title,
  action,
  isLoading = false,
  isError = false,
  empty,
  footer,
  children,
}: {
  title: string;
  action?: React.ReactNode;
  isLoading?: boolean;
  isError?: boolean;
  empty?: string | false;
  footer?: React.ReactNode;
  children: React.ReactNode;
}) {
  const note = isLoading ? "Loading…" : isError ? "Unavailable." : empty;
  return (
    <section className="border-t border-border pt-5 first:border-t-0 first:pt-0">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          {title}
        </h3>
        {action}
      </div>
      {note ? <p className="text-sm text-muted-foreground">{note}</p> : children}
      {footer}
    </section>
  );
}
