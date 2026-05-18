import { expect, test } from "@playwright/test";
import { spawn, type ChildProcess } from "node:child_process";

const baseURL = "http://127.0.0.1:3105";
let serverProcess: ChildProcess | null = null;

test.setTimeout(60_000);

test.beforeAll(async () => {
  serverProcess = spawn(
    process.execPath,
    [
      "./node_modules/next/dist/bin/next",
      "dev",
      "--hostname",
      "127.0.0.1",
      "--port",
      "3105",
    ],
    {
      cwd: process.cwd(),
      env: { ...process.env, NEXT_TELEMETRY_DISABLED: "1" },
      stdio: "ignore",
    }
  );

  await waitForServer();
});

test.afterAll(() => {
  serverProcess?.kill();
  serverProcess = null;
});

const activeCertificate = {
  serial_number: 42,
  key_id: "researcher@login.cluster.example.org",
  principal: "researcher",
  public_key_fingerprint: "SHA256:sample",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 1_800_000_000,
  issued_at: 1_700_000_000,
  source_ip: "127.0.0.1",
  granted_extensions: ["permit-pty"],
  force_command: null,
  revoked: false,
};

test("list to detail to revoke", async ({ page }) => {
  let revoked = false;
  let revokePayload: unknown = null;

  await page.route("**/api/v1/userinfo", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        subject: "sub",
        issuer: "issuer",
        email: "researcher@example.org",
        principal: "researcher",
      },
    });
  });

  await page.route("**/api/v1/certificates?**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        certificates: [certificateState(revoked)],
        total: 1,
        limit: 20,
        offset: 0,
      },
    });
  });

  await page.route("**/api/v1/certificates/42", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: certificateState(revoked),
    });
  });

  await page.route("**/api/v1/revoke", async (route) => {
    revokePayload = route.request().postDataJSON();
    revoked = true;

    await route.fulfill({
      contentType: "application/json",
      json: {
        success: true,
        message: "Certificate revoked.",
        revoked_count: 1,
      },
    });
  });

  await page.goto("/signer/certificates");

  await expect(page.getByText("researcher@example.org")).toBeVisible();
  await expect(page.getByRole("cell", { name: "42" })).toBeVisible();

  await page.getByRole("link", { name: "More" }).click();

  await expect(page).toHaveURL(/\/signer\/certificates\/42$/);
  await expect(
    page.getByRole("heading", { name: "Certificate 42" })
  ).toBeVisible();
  await expect(page.getByText("SHA256:sample")).toBeVisible();

  await page.getByRole("button", { name: "Revoke" }).click();
  await expect(page.getByText("Revoke certificate?")).toBeVisible();
  await page.getByRole("button", { name: "Confirm revoke" }).click();

  await expect
    .poll(() => revokePayload)
    .toEqual({ serial_number: 42, reason: "User requested revocation" });
  await expect(page.locator("dd").filter({ hasText: /^Revoked$/ })).toBeVisible();
  await expect(page.getByRole("button", { name: "Revoke" })).toHaveCount(0);
});

async function waitForServer() {
  const deadline = Date.now() + 30_000;

  while (Date.now() < deadline) {
    try {
      const response = await fetch(baseURL);

      if (response.status < 500) return;
    } catch {
      await new Promise((resolve) => setTimeout(resolve, 250));
    }
  }

  throw new Error("Next.js test server did not start.");
}

function certificateState(revoked: boolean) {
  return {
    ...activeCertificate,
    revoked,
    revoked_at: revoked ? 1_700_000_500 : undefined,
    revocation_reason: revoked ? "User requested revocation" : undefined,
  };
}
