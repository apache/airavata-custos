import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("allocations list", () => {
  test("renders the list and navigates to a detail page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/allocations");
    await expect(page.getByRole("heading", { name: /^Allocations$/ })).toBeVisible();

    const firstRow = page.locator('a[href^="/allocations/alloc-"]').first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
    await firstRow.click();
    await expect(page).toHaveURL(/\/allocations\/alloc-/);
  });

  test("status filter updates URL state", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/allocations");
    await page.getByLabel(/filter by status/i).selectOption("INACTIVE");
    await expect(page).toHaveURL(/[?&]status=INACTIVE/);
  });

  test("axe sweep on allocations list", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/allocations");
    await expect(page.getByRole("heading", { name: /^Allocations$/ })).toBeVisible();
    const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa"]).analyze();
    const blocking = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
