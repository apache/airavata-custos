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

import { z } from "zod";

export const serverSchema = z
  .object({
    NODE_ENV: z.enum(["development", "test", "production"]).default("development"),
    PORTAL_AUTH_MODE: z.enum(["dev", "oidc"]).default("dev"),
    NEXTAUTH_SECRET: z.string().min(8).default("dev-secret-do-not-use-in-prod"),
    NEXTAUTH_URL: z.string().url().optional(),

    CUSTOS_CORE_API_BASE_URL: z.string().url().default("http://localhost:8080"),
    CUSTOS_ADMIN_CLIENT_ID: z.string().optional(),
    CUSTOS_ADMIN_CLIENT_SECRET: z.string().optional(),

    OIDC_ISSUER_URL: z.string().url().optional(),
    OIDC_CLIENT_ID: z.string().optional(),
    OIDC_CLIENT_SECRET: z.string().optional(),
  })
  .superRefine((env, ctx) => {
    // OIDC mode demands these — otherwise NextAuth boots a misconfigured
    // provider that silently fails at sign-in.
    if (env.PORTAL_AUTH_MODE === "oidc") {
      const required: Array<["OIDC_ISSUER_URL" | "OIDC_CLIENT_ID" | "OIDC_CLIENT_SECRET", string | undefined]> = [
        ["OIDC_ISSUER_URL", env.OIDC_ISSUER_URL],
        ["OIDC_CLIENT_ID", env.OIDC_CLIENT_ID],
        ["OIDC_CLIENT_SECRET", env.OIDC_CLIENT_SECRET],
      ];
      for (const [name, value] of required) {
        if (!value || value.trim().length === 0) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            path: [name],
            message: `${name} is required when PORTAL_AUTH_MODE='oidc'`,
          });
        }
      }
    }
  });

const clientSchema = z.object({
  NEXT_PUBLIC_PORTAL_USE_MSW: z.enum(["true", "false"]).default("false"),
  NEXT_PUBLIC_PORTAL_BUILD_SHA: z.string().min(1).default("dev"),
});

function parseServer() {
  const parsed = serverSchema.safeParse(process.env);
  if (!parsed.success) {
    console.error("Invalid server env:", parsed.error.flatten().fieldErrors);
    throw new Error("Invalid server env");
  }
  return parsed.data;
}

function parseClient() {
  const parsed = clientSchema.safeParse({
    NEXT_PUBLIC_PORTAL_USE_MSW: process.env.NEXT_PUBLIC_PORTAL_USE_MSW,
    NEXT_PUBLIC_PORTAL_BUILD_SHA: process.env.NEXT_PUBLIC_PORTAL_BUILD_SHA,
  });
  if (!parsed.success) {
    console.error("Invalid client env:", parsed.error.flatten().fieldErrors);
    throw new Error("Invalid client env");
  }
  return parsed.data;
}

export const serverEnv = parseServer();
export const clientEnv = parseClient();
