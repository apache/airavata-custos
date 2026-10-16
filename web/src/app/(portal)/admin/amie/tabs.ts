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

import { ROUTE_LABELS } from "@/shared/layout/nav";
import type { Metadata } from "next";

export const AMIE_TABS = [
  {
    href: "/admin/amie/packets",
    label: ROUTE_LABELS["/admin/amie/packets"],
    subject: "AmiePacket",
  },
  { href: "/admin/amie/failed", label: ROUTE_LABELS["/admin/amie/failed"], subject: "AmiePacket" },
] as const;

export function amieMetadata(href: (typeof AMIE_TABS)[number]["href"]): Metadata {
  return { title: `${ROUTE_LABELS[href]} · ${ROUTE_LABELS["/admin/amie"]} · Admin` };
}
