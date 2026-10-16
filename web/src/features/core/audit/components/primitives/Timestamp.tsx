/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { formatTimestamp, formatUtc } from "@/shared/format";
import type * as React from "react";

export type TimestampProps = {
  iso?: string | null;
  // Keeps milliseconds, for trace steps that land within one second.
  precise?: boolean;
  className?: string;
  // Replaces the local time as the text, such as a relative time; the title stays UTC.
  children?: React.ReactNode;
};

const PRECISE: Intl.DateTimeFormatOptions = {
  year: "numeric",
  month: "short",
  day: "numeric",
  hour: "numeric",
  minute: "2-digit",
  second: "2-digit",
  fractionalSecondDigits: 3,
};

// Every timestamp renders in local time with the exact UTC value as its title; blank when absent.
export function Timestamp({ iso, precise, className, children }: TimestampProps) {
  if (!iso) return null;
  return (
    <time dateTime={iso} title={formatUtc(iso)} className={className}>
      {children ??
        (precise ? new Date(iso).toLocaleString(undefined, PRECISE) : formatTimestamp(iso))}
    </time>
  );
}
