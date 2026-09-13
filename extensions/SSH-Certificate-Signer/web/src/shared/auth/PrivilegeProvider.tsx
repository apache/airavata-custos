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

import { createContext, type ReactNode, useContext } from "react";

const PrivilegeContext = createContext<ReadonlySet<string>>(new Set());

export function PrivilegeProvider({
  privileges,
  children,
}: {
  privileges: string[];
  children: ReactNode;
}) {
  return (
    <PrivilegeContext.Provider value={new Set(privileges)}>{children}</PrivilegeContext.Provider>
  );
}

export function useSignerPrivileges() {
  const privileges = useContext(PrivilegeContext);
  return {
    canRead: privileges.has("signer:certificates:read"),
    canWrite: privileges.has("signer:certificates:write"),
  };
}
