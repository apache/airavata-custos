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

import { notFound } from "next/navigation";
import { decodeSegment, queryUuid } from "./nav";

// A route page's record id from its raw segment. Middleware has already redirected other spellings
// (padded, uppercase) to the canonical one and served unparseable ids as not found.
export function routeRecordId(
  raw: string,
  parse: (id: string) => string | undefined = queryUuid,
): string {
  return parse(decodeSegment(raw)) ?? notFound();
}
