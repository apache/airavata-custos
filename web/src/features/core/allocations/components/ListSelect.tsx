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

import { SelectField } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import type { UseQueryResult } from "@tanstack/react-query";

// A form picker over a list read: a failed read says so and retries instead of offering nothing.
export function ListSelect<T>({
  label,
  name,
  things,
  query,
  options,
}: {
  label: string;
  name: string;
  things: string;
  query: UseQueryResult<T>;
  options: (data: T | undefined) => Array<{ value: string; label: string }>;
}) {
  if (!query.error) return <SelectField label={label} name={name} options={options(query.data)} />;
  return (
    <div role="alert" className="space-y-1 text-sm">
      <p className="font-medium">{label}</p>
      <p className="text-destructive-text">
        Couldn't load {things}.{" "}
        <Button type="button" variant="link" size="xs" onClick={() => query.refetch()}>
          Try again
        </Button>
      </p>
    </div>
  );
}
