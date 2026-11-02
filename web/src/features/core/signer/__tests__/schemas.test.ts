// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0.

import { describe, expect, it } from "vitest";
import {
  certificateListSchema,
  certificateSchema,
  revokeResponseSchema,
} from "../schemas";

const certificate = {
  tenant_id: "tenant-1",
  client_id: "signer-client",
  serial_number: 42,
  key_id: "key-42",
  principal: "dev-admin",
  user_email: "admin@example.org",
  public_key_fingerprint: "SHA256:public",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 4_102_444_800,
  issued_at: 1_700_000_000,
  source_ip: "192.0.2.4",
  granted_extensions: ["permit-pty"],
  force_command: null,
  revoked: false,
};

describe("signer schemas", () => {
  it("accepts complete certificate detail metadata", () => {
    expect(certificateSchema.parse(certificate)).toEqual(certificate);
  });

  it("accepts a paginated certificate list", () => {
    expect(
      certificateListSchema.parse({ certificates: [certificate], total: 21, limit: 20, offset: 0 }),
    ).toMatchObject({ total: 21, limit: 20, offset: 0 });
  });

  it("accepts an idempotent revoke response", () => {
    expect(
      revokeResponseSchema.parse({
        success: true,
        message: "Certificate was already revoked",
        serial_number: 42,
        revoked: true,
        revoked_at: 1_700_500_000,
        reason: "original reason",
        already_revoked: true,
      }),
    ).toMatchObject({ already_revoked: true, reason: "original reason" });
  });

  it.each([
    ["certificate", { ...certificate, serial_number: "42" }, certificateSchema],
    [
      "list",
      { certificates: [certificate], total: -1, limit: 20, offset: 0 },
      certificateListSchema,
    ],
    [
      "revoke response",
      {
        success: true,
        message: "ok",
        serial_number: 42,
        revoked: true,
        revoked_at: "today",
        reason: "reason",
      },
      revokeResponseSchema,
    ],
  ])("rejects malformed %s data", (_label, value, schema) => {
    expect(() => schema.parse(value)).toThrow();
  });
});
