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

import "server-only";
import { z } from "zod";

const schema = z.object({
  NODE_ENV: z.enum(["development", "test", "production"]).default("development"),
  NEXTAUTH_SECRET: z.string().min(8),
  CUSTOS_SIGNER_API_BASE_URL: z.string().url().default("http://localhost:8084"),
  CUSTOS_PORTAL_BASE_URL: z.string().url().default("http://localhost:3000"),
});

const parsed = schema.safeParse(process.env);
if (!parsed.success) {
  console.error("Invalid signer web environment:", parsed.error.flatten().fieldErrors);
  throw new Error("Invalid signer web environment");
}

export const serverEnv = parsed.data;
