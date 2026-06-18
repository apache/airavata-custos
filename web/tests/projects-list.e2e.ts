import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("projects list", () => {
  test("renders the list and navigates to a detail page", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects");
    await expect(page.getByRole("heading", { name: /^Projects$/ })).toBeVisible();

    const firstRow = page.locator('a[href^="/projects/project-"]').first();
    await expect(firstRow).toBeVisible({ timeout: 15_000 });
    await firstRow.click();
    await expect(page).toHaveURL(/\/projects\/project-/);
  });

  test("status filter updates URL state", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects");
    await page.getByLabel(/filter by status/i).selectOption("INACTIVE");
    await expect(page).toHaveURL(/[?&]status=INACTIVE/);
  });

  test("axe sweep on projects list", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects");
    await expect(page.getByRole("heading", { name: /^Projects$/ })).toBeVisible();
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();
    const blocking = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
