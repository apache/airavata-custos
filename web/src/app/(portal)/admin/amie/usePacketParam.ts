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

import { pageParam } from "@/shared/hooks/usePageClamp";
import { queryUuid } from "@/shared/layout/nav";
import { useRouter, useSearchParams } from "next/navigation";

// Writes one key of path's query, or removes it when value is undefined.
function useReplaceParam(path: string) {
  const router = useRouter();
  const searchParams = useSearchParams();
  return (key: string, value: string | undefined) => {
    const params = new URLSearchParams(searchParams.toString());
    if (value) params.set(key, value);
    else params.delete(key);
    const next = params.toString();
    router.replace(next ? `${path}?${next}` : path, { scroll: false });
  };
}

// The open packet lives in ?packet=, so a reload or a shared link reopens its drawer.
// Replacing always targets path, so a /packets/[id] deep link doesn't keep its stale id.
export function usePacketParam(path: string) {
  const searchParams = useSearchParams();
  const replace = useReplaceParam(path);
  return [
    queryUuid(searchParams.get("packet")),
    (id: string | undefined) => replace("packet", id),
  ] as const;
}

// The list page lives in ?page=, a whole number from 1; a page past the end shows First page.
export function usePageParam(path: string) {
  const searchParams = useSearchParams();
  const replace = useReplaceParam(path);
  return [
    pageParam(searchParams),
    (next: number) => replace("page", next > 1 ? String(next) : undefined),
  ] as const;
}
