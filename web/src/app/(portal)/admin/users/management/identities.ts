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

import { BookUser, Fingerprint, Globe, KeyRound, type LucideIcon, Users } from "lucide-react";

// Display metadata for UserIdentity.source. Core sign-in writes "oidc", the AMIE
// connector "access" and COmanage "comanage".
export const IDENTITY_SOURCE_LABELS: Record<string, string> = {
  oidc: "OIDC",
  access: "ACCESS",
  comanage: "COmanage",
  cilogon: "CILogon",
  orcid: "ORCID",
  nairr: "NAIRR",
};

const IDENTITY_SOURCE_ICONS: Record<string, LucideIcon> = {
  oidc: KeyRound,
  access: Globe,
  comanage: Users,
  cilogon: KeyRound,
  orcid: BookUser,
  nairr: Fingerprint,
};

export function identitySourceLabel(source: string | undefined): string {
  if (!source) return "Unknown";
  return IDENTITY_SOURCE_LABELS[source] ?? source;
}

// A source as typed, by key or label in any case ("ACCESS", "access"), as the key master stores.
export const identitySourceKey = (typed: string) =>
  Object.entries(IDENTITY_SOURCE_LABELS).find(([key, label]) =>
    [key, label.toLowerCase()].includes(typed.toLowerCase()),
  )?.[0] ?? typed;

export function identitySourceIcon(source: string | undefined): LucideIcon {
  return (source ? IDENTITY_SOURCE_ICONS[source] : undefined) ?? Fingerprint;
}
