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

// "Aug 1, 2026" in the viewer's zone; blank when the timestamp is absent.
export const formatDate = (iso?: string | null) =>
  iso ? new Date(iso).toLocaleDateString(undefined, { dateStyle: "medium" }) : "";

// "Aug 1, 2026" for a date-only value stored at UTC midnight, so it never shows the previous day.
export const formatDay = (iso?: string | null) =>
  iso ? new Date(iso).toLocaleDateString(undefined, { dateStyle: "medium", timeZone: "UTC" }) : "";

// "Aug 1, 2026, 3:04 PM" in the viewer's zone; blank when the timestamp is absent.
export const formatDateTime = (iso?: string | null) =>
  iso ? new Date(iso).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" }) : "";

// Event timestamps across AMIE, traces and events: local time to the second, with formatUtc as the title.
export const formatTimestamp = (iso?: string | null) =>
  iso ? new Date(iso).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "medium" }) : "";

// "2026-08-01 15:04:05.123 UTC"; blank when the timestamp is absent.
export const formatUtc = (iso?: string | null) =>
  iso ? new Date(iso).toISOString().replace("T", " ").replace("Z", " UTC") : "";

// "1 credit", "2 credits": the noun for a count; pluralForm covers irregular ones ("reply", "replies").
export const pluralize = (noun: string, count: number, pluralForm?: string) =>
  count === 1 ? noun : (pluralForm ?? `${noun}s`);

// One sentence-case label for a backend enum value: "WAITING_APPROVAL" reads "Waiting approval".
export const enumLabel = (value: string) =>
  value ? value.charAt(0).toUpperCase() + value.slice(1).toLowerCase().replaceAll("_", " ") : "";

export const formatNumber = (n = 0) => n.toLocaleString(undefined, { maximumFractionDigits: 2 });

// Every SU amount (allocated, requested, used, remaining): master keeps SUs to 6 decimals.
export const formatSU = (n = 0) => n.toLocaleString(undefined, { maximumFractionDigits: 6 });

// A rate under one to 6 significant digits ("0.001"), one or more to 6 decimals with the whole part intact
// ("999,999.99"); only a rate under one millionth reads in exponent form.
export const formatRate = (n = 0) =>
  n.toLocaleString(
    undefined,
    Math.abs(n) >= 1
      ? { maximumFractionDigits: 6 }
      : {
          maximumSignificantDigits: 6,
          notation: n !== 0 && Math.abs(n) < 1e-6 ? "scientific" : "standard",
        },
  );
