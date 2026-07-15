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

// TODO(openapi): replace with generated once the SSH Certificate Signer ships
// an OpenAPI spec; the signer is a separate extension service with no spec today.
import { z } from "zod";

// Mirrors extensions/SSH-Certificate-Signer CertificateResponse. Unix-second
// timestamps are passed through as numbers and formatted in the UI layer.
export const certificateSchema = z.object({
  serial_number: z.number().int(),
  key_id: z.string(),
  principal: z.string(),
  public_key_fingerprint: z.string(),
  ca_fingerprint: z.string(),
  valid_after: z.number().int(),
  valid_before: z.number().int(),
  issued_at: z.number().int(),
  source_ip: z.string().optional(),
  granted_extensions: z.array(z.string()).optional(),
  force_command: z.string().nullable().optional(),
  revoked: z.boolean(),
  revoked_at: z.number().int().nullable().optional(),
  revocation_reason: z.string().optional(),
});
export type Certificate = z.infer<typeof certificateSchema>;

export const certificateListSchema = z.object({
  certificates: z.array(certificateSchema),
  total: z.number().int().nonnegative(),
  limit: z.number().int(),
  offset: z.number().int(),
});
export type CertificateList = z.infer<typeof certificateListSchema>;

// Response from POST /signer/certificates/{serial}/revoke. already_revoked is
// true when the certificate was revoked by a prior request (idempotent success).
export const revokeResponseSchema = z.object({
  success: z.boolean(),
  message: z.string(),
  serial_number: z.number().int(),
  revoked: z.boolean(),
  revoked_at: z.number().int(),
  reason: z.string(),
  already_revoked: z.boolean().optional(),
});
export type RevokeResponse = z.infer<typeof revokeResponseSchema>;
