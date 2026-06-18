import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("project detail", () => {
  test("renders header, overview, and members tab", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/project-001");
    await expect(
      page.getByRole("heading", { name: /Genomic Sequencing Pipeline/i }),
    ).toBeVisible();

    await page.getByRole("tab", { name: /members/i }).click();
    await expect(page.getByText(/Ada Lovelace/i).first()).toBeVisible({ timeout: 15_000 });
  });

  test("axe sweep on projects detail", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/projects/project-001");
    await expect(
      page.getByRole("heading", { name: /Genomic Sequencing Pipeline/i }),
    ).toBeVisible();
    const results = await new AxeBuilder({ page })
      .withTags(["wcag2a", "wcag2aa"])
      .analyze();
    const blocking = results.violations.filter(
      (v) => v.impact === "serious" || v.impact === "critical",
    );
    expect(blocking, JSON.stringify(blocking, null, 2)).toEqual([]);
  });
});
