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

import { cn } from "@/lib/utils";
import type { HTMLAttributes, ReactNode } from "react";

export type StatCardRowCols = 3 | 4 | 5;

export type StatCardRowProps = Omit<HTMLAttributes<HTMLDivElement>, "children"> & {
  children: ReactNode;
  cols?: StatCardRowCols;
};

const COL_CLASS: Record<StatCardRowCols, string> = {
  3: "md:grid-cols-3",
  4: "sm:grid-cols-2 lg:grid-cols-4",
  5: "sm:grid-cols-2 lg:grid-cols-5",
};

export function StatCardRow({ children, cols = 3, className, ...rest }: StatCardRowProps) {
  return (
    <div className={cn("grid grid-cols-1 gap-4", COL_CLASS[cols], className)} {...rest}>
      {children}
    </div>
  );
}
