import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

const SEVERITIES = ["serious", "critical"] as const;

test.describe("admin AMIE packet detail", () => {
  test("clicking View opens the drawer with all four tabs", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets");
    await expect(page.getByRole("heading", { name: /AMIE packet inbox/i })).toBeVisible({
      timeout: 20_000,
    });

    const firstRow = page.locator("tbody tr").first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
    await firstRow.getByRole("button", { name: /^View$/ }).click();

    await expect(page.getByRole("tab", { name: /Overview/i })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole("tab", { name: /Raw JSON/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Timeline/i })).toBeVisible();
    await expect(page.getByRole("tab", { name: /Linked entity/i })).toBeVisible();
  });

  test("Timeline tab shows event rows; rows with trace_id expose a View trace link", async ({
    page,
  }) => {
    await signInAs(page, "admin");
    // pkt-100 carries an event with trace_id in the fixtures.
    await page.goto("/admin/amie/packets/pkt-100");

    await expect(page.getByRole("tab", { name: /Timeline/i })).toBeVisible({ timeout: 15_000 });
    await page.getByRole("tab", { name: /Timeline/i }).click();
    await expect(page.getByTestId("packet-events-table")).toBeVisible({ timeout: 10_000 });
    await expect(page.getByRole("link", { name: /View trace/i })).toBeVisible({
      timeout: 10_000,
    });
  });

  test("axe: no serious or critical violations on the open drawer", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/admin/amie/packets/pkt-100");
    await expect(page.getByRole("tab", { name: /Overview/i })).toBeVisible({ timeout: 20_000 });
    await page.waitForLoadState("networkidle");
    const results = await new AxeBuilder({ page })
      .include("main")
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();
    const critical = results.violations.filter((v) => SEVERITIES.includes(v.impact as "serious"));
    expect(critical).toEqual([]);
  });
});
