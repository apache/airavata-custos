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

import { type QueryKey, useMutation, useQueryClient } from "@tanstack/react-query";

// A mutation that refreshes the given query keys once it succeeds.
export function useInvalidating<O, R>(mutationFn: (options: O) => Promise<R>, ...keys: QueryKey[]) {
  const client = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: () => Promise.all(keys.map((queryKey) => client.invalidateQueries({ queryKey }))),
  });
}
