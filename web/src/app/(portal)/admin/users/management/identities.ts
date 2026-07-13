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

import { BookUser, Fingerprint, Globe, KeyRound, type LucideIcon } from "lucide-react";

// Display metadata for the `source` field on UserIdentity (GET
// /users/{id}/user-identities) — "the source's native identifier" per the
// API doc, e.g. "access", "cilogon", "orcid", "nairr".
export const IDENTITY_SOURCE_LABELS: Record<string, string> = {
  access: "ACCESS",
  cilogon: "CILogon",
  orcid: "ORCID",
  nairr: "NAIRR",
};

const IDENTITY_SOURCE_ICONS: Record<string, LucideIcon> = {
  access: Globe,
  cilogon: KeyRound,
  orcid: BookUser,
  nairr: Fingerprint,
};

export function identitySourceLabel(source: string | undefined): string {
  if (!source) return "Unknown";
  return IDENTITY_SOURCE_LABELS[source] ?? source;
}

export function identitySourceIcon(source: string | undefined): LucideIcon {
  return (source ? IDENTITY_SOURCE_ICONS[source] : undefined) ?? Fingerprint;
}
