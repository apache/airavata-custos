// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0.

import { expect, test as base, type BrowserContext, type Route } from "@playwright/test";
import certificatesFixture from "../../src/features/core/signer/__fixtures__/certificates.json";
import type { Certificate } from "../../src/features/core/signer/schemas";

export type SignerScenario =
  | "default"
  | "empty"
  | "paginated"
  | "already-revoked"
  | "inactive"
  | "forbidden"
  | "server-error";

type SignerApiController = {
  setScenario(scenario: SignerScenario): void;
};

type SignerFixtures = {
  signerApi: SignerApiController;
};

function cloneCertificates(): Certificate[] {
  return (certificatesFixture as Certificate[]).map((certificate) => ({
    ...certificate,
    granted_extensions: certificate.granted_extensions
      ? [...certificate.granted_extensions]
      : undefined,
  }));
}

function cookiePrivileges(route: Route): string[] {
  const cookie = route.request().headers().cookie ?? "";
  const match = cookie
    .split("; ")
    .find((entry) => entry.startsWith("custos.test-privileges="));
  if (!match) return [];
  return decodeURIComponent(match.slice("custos.test-privileges=".length))
    .split(",")
    .filter(Boolean);
}

function paginatedCertificates(certificates: Certificate[]): Certificate[] {
  const baseCertificate = certificates[0];
  if (!baseCertificate) return [];
  return Array.from({ length: 25 }, (_, index) => ({
    ...baseCertificate,
    granted_extensions: baseCertificate.granted_extensions
      ? [...baseCertificate.granted_extensions]
      : undefined,
    serial_number: 100 + index,
    key_id: `key-${100 + index}`,
    principal: `user-${index + 1}`,
  }));
}

async function json(route: Route, body: unknown, status = 200) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

async function installSignerRoutes(context: BrowserContext) {
  const certificates = cloneCertificates();
  const unexpectedRequests: string[] = [];
  const interceptedRequests: string[] = [];
  let scenario: SignerScenario = "default";

  await context.route("**/api/v1/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    const requestLabel = `${request.method()} ${url.pathname}${url.search}`;
    interceptedRequests.push(requestLabel);

    if (request.method() === "GET" && url.pathname === "/api/v1/user/privileges") {
      await json(route, { privileges: cookiePrivileges(route) });
      return;
    }

    if (
      request.method() === "GET" &&
      url.pathname === "/api/v1/signer/admin/certificates"
    ) {
      const source =
        scenario === "empty"
          ? []
          : scenario === "paginated"
            ? paginatedCertificates(certificates)
            : certificates;
      const limit = Number(url.searchParams.get("limit") ?? source.length);
      const offset = Number(url.searchParams.get("offset") ?? 0);
      await json(route, {
        certificates: source.slice(offset, offset + limit),
        total: source.length,
        limit,
        offset,
      });
      return;
    }

    const detailMatch = url.pathname.match(
      /^\/api\/v1\/signer\/admin\/certificates\/(\d+)$/,
    );
    if (request.method() === "GET" && detailMatch) {
      const certificate = certificates.find(
        (candidate) => candidate.serial_number === Number(detailMatch[1]),
      );
      await json(
        route,
        certificate ?? { error: "certificate not found" },
        certificate ? 200 : 404,
      );
      return;
    }

    const revokeMatch = url.pathname.match(
      /^\/api\/v1\/signer\/admin\/certificates\/(\d+)\/revoke$/,
    );
    if (request.method() === "POST" && revokeMatch) {
      if (scenario === "forbidden") {
        await json(
          route,
          { error: "insufficient_privilege", message: "Caller lacks required privilege" },
          403,
        );
        return;
      }
      if (scenario === "inactive") {
        await json(
          route,
          { error: "certificate_not_active", message: "Certificate is not active" },
          409,
        );
        return;
      }
      if (scenario === "server-error") {
        await json(
          route,
          { error: "authorization_unavailable", message: "Signer temporarily unavailable" },
          503,
        );
        return;
      }
      if (!cookiePrivileges(route).includes("signer:certificates:write")) {
        await json(
          route,
          { error: "insufficient_privilege", message: "Caller lacks required privilege" },
          403,
        );
        return;
      }

      const serial = Number(revokeMatch[1]);
      const certificate = certificates.find((candidate) => candidate.serial_number === serial);
      if (!certificate) {
        await json(route, { error: "certificate not found" }, 404);
        return;
      }

      let body: { reason?: string } = {};
      try {
        body = request.postDataJSON() as { reason?: string };
      } catch {
        // Leave the body empty so the mock returns the same validation response
        // that the signer API would return for malformed JSON.
      }
      const reason = (body.reason ?? "").trim();
      if (!reason || reason.length > 255) {
        await json(
          route,
          { error: "invalid_reason", message: "Reason must be between 1 and 255 characters" },
          400,
        );
        return;
      }

      const now = Math.floor(Date.now() / 1000);
      if (!certificate.revoked && (certificate.valid_after > now || certificate.valid_before <= now)) {
        await json(
          route,
          { error: "certificate_not_active", message: "Certificate is not active" },
          409,
        );
        return;
      }

      if (scenario === "already-revoked") {
        certificate.revoked = true;
        certificate.revoked_at = 1_700_500_000;
        certificate.revocation_reason = "Original backend reason";
        certificate.revoked_by = "another-admin";
      }
      const alreadyRevoked = certificate.revoked;
      if (!alreadyRevoked) {
        certificate.revoked = true;
        certificate.revoked_at = now;
        certificate.revocation_reason = reason;
        certificate.revoked_by = "test-admin";
      }
      await json(route, {
        success: true,
        message: alreadyRevoked
          ? "Certificate was already revoked"
          : "Certificate revoked successfully",
        serial_number: serial,
        revoked: true,
        revoked_at: certificate.revoked_at ?? now,
        reason: certificate.revocation_reason ?? reason,
        already_revoked: alreadyRevoked,
      });
      return;
    }

    unexpectedRequests.push(requestLabel);
    await route.abort("failed");
  });

  return {
    controller: {
      setScenario(nextScenario: SignerScenario) {
        scenario = nextScenario;
      },
    },
    interceptedRequests,
    unexpectedRequests,
  };
}

export const test = base.extend<SignerFixtures>({
  signerApi: [
    async ({ context }, use, testInfo) => {
      const routes = await installSignerRoutes(context);
      await use(routes.controller);

      await testInfo.attach("signer-api-requests", {
        body:
          routes.interceptedRequests.join("\n") ||
          "No API request was made; the client-side permission gate handled this scenario.",
        contentType: "text/plain",
      });
      expect(routes.unexpectedRequests, "unexpected API requests escaped signer mocks").toEqual(
        [],
      );
    },
    { auto: true },
  ],
});

export { expect } from "@playwright/test";
