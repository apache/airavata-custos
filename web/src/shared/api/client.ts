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

import type { CreateClientConfig } from "@/generated/core/client.gen";
import { z } from "zod";

const errorBodySchema = z.object({
  error: z.string().optional(),
  message: z.string().optional(),
  code: z.string().optional(),
});
type ApiErrorBody = z.infer<typeof errorBodySchema> | string | null;

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly path: string,
    public readonly body: ApiErrorBody,
    message?: string,
  ) {
    super(message ?? `API ${status} on ${path}`);
    this.name = "ApiError";
  }
}

function parseErrorBody(text: string): ApiErrorBody {
  try {
    return errorBodySchema.parse(JSON.parse(text));
  } catch {
    return text || null;
  }
}

// Every generated SDK client fetches through here; non-2xx responses throw ApiError.
const portalFetch: typeof fetch = async (input, init) => {
  const response = await fetch(input, init);
  if (response.ok) return response;
  const url = input instanceof Request ? input.url : String(input);
  const body = parseErrorBody(await response.text());
  // The backend reports {"error": "..."}, some with a readable "message" beside a code.
  const message = typeof body === "object" ? body?.message || body?.error : undefined;
  throw new ApiError(response.status, url, body, message || undefined);
};

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  fetch: portalFetch,
});
