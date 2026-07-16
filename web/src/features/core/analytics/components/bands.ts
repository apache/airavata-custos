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

import type { UrgencyBand } from "../lib";

// Amber maps to the warn tone, red to the error tone. "ok" gets no pill or rule.
export const BAND_PILL_LABEL: Record<Exclude<UrgencyBand, "ok">, string> = {
  amber: "Low",
  red: "Critical",
};

export const BAND_PILL_CLASS: Record<Exclude<UrgencyBand, "ok">, string> = {
  amber: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
  red: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
};

export const BAND_RULE_CLASS: Record<Exclude<UrgencyBand, "ok">, string> = {
  amber: "border-l-2 border-l-[color:var(--tone-warn-fg)]",
  red: "border-l-2 border-l-[color:var(--tone-error-fg)]",
};

export const BAND_METER_CLASS: Record<UrgencyBand, string> = {
  ok: "bg-[color:var(--custos-green-500)]",
  amber: "bg-[color:var(--custos-amber-500)]",
  red: "bg-[color:var(--custos-red-500)]",
};
