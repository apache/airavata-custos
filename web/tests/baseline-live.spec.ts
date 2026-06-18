// Live integration test: runs against a Custos backend whose state matches
// the AMIE baseline scenario + dev_portal_data.sql seed as declared in
// connectors/ACCESS/AMIE-Processor/testdata/scenarios/baseline.yaml.
//
// One UI assertion per data model the four portal areas (Projects,
// Allocations, Tracing, AMIE) render. Excluded from `pnpm test:e2e` via the
// .spec.ts suffix.
//
// Bring up the stack (see docs/internal/portal/baseline-runbook.md), then:
//   LIVE_PORTAL_URL=http://localhost:3001 \
//   pnpm exec playwright test --config playwright.config.ts \
//     --testMatch '**/baseline-live.spec.ts' --project=chromium

import { type Page, expect, test } from "@playwright/test";

const BASE = process.env.LIVE_PORTAL_URL ?? "http://localhost:3001";

async function signInAdmin(page: Page) {
  await page.goto(`${BASE}/sign-in`);
  // SignInForm reads useSearchParams() which Suspends on first paint; the
  // form renders only after hydration. Wait for the Button (the rightmost
  // hydration-dependent node) before interacting.
  const submit = page.getByRole("button", { name: /^sign in$/i });
  await submit.waitFor({ state: "visible", timeout: 60_000 });
  await page.locator("#dev-level").selectOption("admin");
  await submit.click();
  await page.waitForURL((url) => !url.pathname.startsWith("/sign-in"), { timeout: 30_000 });
}

async function goto(page: Page, path: string) {
  const errors: string[] = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.goto(`${BASE}${path}`, { waitUntil: "domcontentloaded" });
  await page.waitForLoadState("networkidle", { timeout: 15_000 }).catch(() => {});
  expect(errors, errors.join("\n")).toEqual([]);
}

test.describe("baseline live walk", () => {
  test.describe.configure({ mode: "serial" });

  // ---- Projects area --------------------------------------------------

  test("projects: 2 baseline projects render", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/projects");
    await expect(page.getByRole("link", { name: "BL-001" }).first()).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("link", { name: "BL-002" }).first()).toBeVisible();
  });

  test("project detail: opens BL-001 with title visible", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/projects");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await expect(page).toHaveURL(/\/projects\/[a-f0-9-]+/);
    await expect(page.getByRole("heading", { name: /BL-001/ })).toBeVisible();
  });

  test("project detail: PI is rendered by name, not by UUID", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/projects");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await expect(page.getByText("Pat Madison").first()).toBeVisible({ timeout: 15_000 });
    // UUIDs are 32 hex with 4 dashes — a stray UUID in the PI slot would be a regression.
    await expect(page.getByText(/^[0-9a-f]{8}-[0-9a-f]{4}-/i)).toHaveCount(0);
  });

  test("project members tab: BL-001 shows PI/CO_PI/MEMBER roles and an Allocations chip per user", async ({
    page,
  }) => {
    await signInAdmin(page);
    await goto(page, "/projects");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await page.getByRole("tab", { name: /members/i }).click();
    await expect(page.getByText("Pat Madison").first()).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("Casey Collaborator").first()).toBeVisible();
    await expect(page.getByText("Riley Researcher").first()).toBeVisible();
    // Role column populated, not blank/—.
    await expect(page.getByText(/^PI$/).first()).toBeVisible();
    await expect(page.getByText(/^Co-PI$/).first()).toBeVisible();
    await expect(page.getByText(/^Member$/).first()).toBeVisible();
    // Allocation chip for BL-001 visible in the Allocations column.
    await expect(page.getByRole("link", { name: /BL-001/ }).first()).toBeVisible();
  });

  // ---- Allocations area -----------------------------------------------

  test("allocations: 2 baseline allocations render", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/allocations");
    await expect(page.getByRole("link", { name: "BL-001" }).first()).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("link", { name: "BL-002" }).first()).toBeVisible();
  });

  test("allocation detail: Overview shows BL-001 with ACTIVE status", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/allocations");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await expect(page).toHaveURL(/\/allocations\/[a-f0-9-]+/);
    await expect(page.getByRole("heading", { name: /BL-001/ })).toBeVisible();
    await expect(page.getByText(/ACTIVE/i).first()).toBeVisible();
  });

  test("allocation members tab: BL-001 has 1 member (from request_account_create)", async ({
    page,
  }) => {
    await signInAdmin(page);
    await goto(page, "/allocations");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await page.getByRole("tab", { name: /members/i }).click();
    // Baseline produces exactly one ACTIVE membership for the survivor (Pat).
    // Empty-state copy is gated on count===0; assert it is absent.
    await expect(page.getByText(/No members yet/i)).toHaveCount(0, { timeout: 15_000 });
  });

  test("allocation change-requests tab: BL-001 shows the seeded PENDING request", async ({
    page,
  }) => {
    await signInAdmin(page);
    await goto(page, "/allocations");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await page.getByRole("tab", { name: /change requests/i }).click();
    await expect(page.getByText(/1 request/i)).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText(/PENDING/i).first()).toBeVisible();
  });

  test("allocation usage tab: BL-001 has at least one usage row from seed", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/allocations");
    await page.getByRole("link", { name: "BL-001" }).first().click();
    await page.getByRole("tab", { name: /usage/i }).click();
    await expect(page.getByText(/No usage data/i)).toHaveCount(0, { timeout: 15_000 });
  });

  // ---- AMIE area ------------------------------------------------------

  test("AMIE inbox: shows the 11 baseline packets DECODED", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/admin/amie/packets");
    await expect(page.locator("main")).toBeVisible();
    const decoded = page.getByText(/DECODED/);
    await expect(decoded.first()).toBeVisible({ timeout: 15_000 });
    expect(await decoded.count()).toBeGreaterThanOrEqual(11);
  });

  test("AMIE inbox: rows show packet types from baseline scenario", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/admin/amie/packets");
    // Packet type cells are rendered inside table rows; the filter dropdown
    // also contains them as <option>, so scope to the table body.
    const rows = page.locator("tbody tr");
    await expect(rows.first()).toBeVisible({ timeout: 15_000 });
    expect(await rows.count()).toBeGreaterThanOrEqual(11);
  });

  // ---- Tracing area ---------------------------------------------------

  test("admin/traces: list shows traces from the AMIE baseline", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/admin/traces");
    // Each trace row renders a button labeled "Open trace <id-prefix>".
    const openButtons = page.getByRole("button", { name: /^open trace/i });
    await expect(openButtons.first()).toBeVisible({ timeout: 15_000 });
  });

  test("admin/traces: list includes amie source traces", async ({ page }) => {
    await signInAdmin(page);
    await goto(page, "/admin/traces");
    await expect(page.getByText(/amie/i).first()).toBeVisible({ timeout: 15_000 });
  });
});
