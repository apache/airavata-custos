// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0.

import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { signInAs } from "./fixtures/auth";
import { expect, test } from "./fixtures/signer-api";

async function openRevokeDialog(page: Page, serial = 42) {
  await page.goto(`/admin/signer/certificates/${serial}`);
  await expect(page.getByRole("heading", { name: `Certificate ${serial}` })).toBeVisible();
  await page.getByRole("button", { name: /^Revoke$/ }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
}

async function submitReason(page: Page, reason: string) {
  await page.getByLabel(/reason/i).fill(reason);
  await page.getByRole("button", { name: /^Confirm revoke$/ }).click();
}

async function expectNoSeriousAxeViolations(page: Page, include: string) {
  const results = await new AxeBuilder({ page })
    .include(include)
    .withTags(["wcag2a", "wcag2aa"])
    .analyze();
  const blocking = results.violations.filter(
    (violation) => violation.impact === "serious" || violation.impact === "critical",
  );
  expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
}

test.describe("ssh certificates", () => {
  test("lists certificates and navigates to complete detail metadata", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/signer/certificates");
    await expect(page.getByRole("heading", { name: /^SSH Certificates$/ })).toBeVisible();
    await page.getByRole("link", { name: "42" }).click();
    await expect(page).toHaveURL(/\/admin\/signer\/certificates\/42/);
    await expect(page.getByRole("heading", { name: /^Certificate 42$/ })).toBeVisible();
    await expect(page.getByText("researcher@example.org")).toBeVisible();
    await expect(page.getByText("203.0.113.42")).toBeVisible();
    await expect(page.getByText(/permit-pty/)).toBeVisible();
  });

  test("search and status filters update URL state and filter the loaded page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/signer/certificates");
    await page.getByRole("searchbox", { name: /search certificates/i }).fill("44");
    await expect(page).toHaveURL(/[?&]q=44/);
    await expect(page.getByRole("link", { name: "44" })).toBeVisible();
    await expect(page.getByRole("link", { name: "42" })).toHaveCount(0);
    await page.getByLabel(/filter by status/i).selectOption("revoked");
    await expect(page).toHaveURL(/[?&]status=revoked/);
  });

  test("shows a true empty state", async ({ page, signerApi }) => {
    await signInAs(page, "admin");
    signerApi.setScenario("empty");
    await page.goto("/admin/signer/certificates");
    await expect(page.getByRole("heading", { name: /no certificates yet/i })).toBeVisible();
  });

  test("paginates against accurate server totals", async ({ page, signerApi }) => {
    await signInAs(page, "admin");
    signerApi.setScenario("paginated");
    await page.goto("/admin/signer/certificates");
    await expect(page.getByText(/Showing 1.*20 of 25/)).toBeVisible();
    await page.getByRole("button", { name: /^Next$/ }).click();
    await expect(page.getByText(/Showing 21.*25 of 25/)).toBeVisible();
    await expect(page.getByRole("link", { name: "120" })).toBeVisible();
  });

  test("distinguishes no matches on a loaded page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/signer/certificates");
    await page.getByRole("searchbox", { name: /search certificates/i }).fill("not-present");
    await expect(page.getByRole("heading", { name: /no matches on this page/i })).toBeVisible();
  });

  test("renders active, future, expired, and revoked details correctly", async ({ page }) => {
    await signInAs(page, "admin");
    for (const [serial, status] of [
      [42, "Active"],
      [43, "Expired"],
      [44, "Revoked"],
      [45, "Not yet valid"],
    ] as const) {
      await page.goto(`/admin/signer/certificates/${serial}`);
      await expect(page.getByRole("heading", { name: `Certificate ${serial}` })).toBeVisible();
      await expect(page.getByText(status, { exact: true }).first()).toBeVisible();
      if (serial === 44) await expect(page.getByText("Key compromised")).toBeVisible();
    }
  });

  test("renders not-found detail handling", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/signer/certificates/999");
    await expect(page.getByRole("heading", { name: /certificate not found/i })).toBeVisible();
  });

  test("revokes an active certificate and reflects backend state", async ({ page }) => {
    await signInAs(page, "admin");
    await openRevokeDialog(page);
    await submitReason(page, "Compromised in e2e");
    await expect(page.getByText("Compromised in e2e").first()).toBeVisible();
    await expect(page.getByRole("button", { name: /^Revoke$/ })).toHaveCount(0);
  });

  test("handles an already-revoked race using original backend state", async ({
    page,
    signerApi,
  }) => {
    await signInAs(page, "admin");
    signerApi.setScenario("already-revoked");
    await openRevokeDialog(page);
    await submitReason(page, "New request reason");
    await expect(page.getByText("Original backend reason")).toBeVisible();
    await expect(page.getByText("another-admin")).toBeVisible();
    await expect(page.getByRole("button", { name: /^Revoke$/ })).toHaveCount(0);
  });

  test("explains when a certificate becomes inactive before confirmation", async ({
    page,
    signerApi,
  }) => {
    await signInAs(page, "admin");
    signerApi.setScenario("inactive");
    await openRevokeDialog(page);
    await submitReason(page, "Race test");
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText(/no longer active/i)).toBeVisible();
  });

  test("keeps the dialog open after privilege loss", async ({ page, signerApi }) => {
    await signInAs(page, "admin");
    signerApi.setScenario("forbidden");
    await openRevokeDialog(page);
    await submitReason(page, "Privilege race");
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText(/no longer have permission/i)).toBeVisible();
  });

  test("keeps a retryable server failure in the dialog", async ({ page, signerApi }) => {
    await signInAs(page, "admin");
    signerApi.setScenario("server-error");
    await openRevokeDialog(page);
    await submitReason(page, "Retry test");
    await expect(page.getByRole("dialog")).toBeVisible();
    await expect(page.getByText("Signer temporarily unavailable")).toBeVisible();
    await expect(page.getByRole("button", { name: /^Confirm revoke$/ })).toBeEnabled();
  });

  test("read-only administrators can inspect but cannot revoke", async ({ page }) => {
    await signInAs(page, "signer-reader");
    await page.goto("/admin/signer/certificates/42");
    await expect(page.getByRole("heading", { name: /^Certificate 42$/ })).toBeVisible();
    await expect(page.getByRole("button", { name: /^Revoke$/ })).toHaveCount(0);
  });

  test("users without signer read privilege cannot access signer administration", async ({
    page,
  }) => {
    await signInAs(page, "viewer");
    await page.goto("/admin/signer/certificates");
    await expect(page.getByText(/not permitted/i)).toBeVisible();
    await expect(page.getByRole("heading", { name: /^SSH Certificates$/ })).toHaveCount(0);
  });

  test("future and expired certificates cannot be revoked", async ({ page }) => {
    await signInAs(page, "admin");
    for (const serial of [43, 45]) {
      await page.goto(`/admin/signer/certificates/${serial}`);
      await expect(page.getByRole("heading", { name: `Certificate ${serial}` })).toBeVisible();
      await expect(page.getByRole("button", { name: /^Revoke$/ })).toHaveCount(0);
    }
  });

  test("has no serious or critical accessibility violations", async ({ page }) => {
    test.setTimeout(45_000);
    await signInAs(page, "admin");
    await page.goto("/admin/signer/certificates");
    await expect(page.getByRole("heading", { name: /^SSH Certificates$/ })).toBeVisible();
    await expectNoSeriousAxeViolations(page, "main");

    await page.goto("/admin/signer/certificates/42");
    await expect(page.getByRole("heading", { name: /^Certificate 42$/ })).toBeVisible();
    await expectNoSeriousAxeViolations(page, "main");

    await page.getByRole("button", { name: /^Revoke$/ }).click();
    await expect(page.getByRole("dialog")).toBeVisible();
    await expectNoSeriousAxeViolations(page, "[role=dialog]");
  });
});
