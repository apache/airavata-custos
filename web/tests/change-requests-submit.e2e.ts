import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("change request submit", () => {
  test("PI persona opens drawer, fills, submits", async ({ page }) => {
    await signInAs(page, "manager");
    await page.goto("/allocations/alloc-001");
    await expect(
      page.getByRole("heading", { name: /Genomic Sequencing - GPU Pool/i }),
    ).toBeVisible();

    await page.getByRole("tab", { name: /^change requests$/i }).click();
    await page.getByRole("button", { name: /submit change request/i }).click();

    await expect(
      page.getByRole("heading", { name: /submit change request/i }),
    ).toBeVisible({ timeout: 15_000 });

    await page.getByLabel(/additional sus requested/i).fill("10000");
    await page
      .getByLabel(/reason/i)
      .fill("Need additional SUs for the final modeling run before paper submission deadline.");

    await page.getByRole("button", { name: /submit request/i }).click();

    await expect(page.getByText(/change request submitted/i)).toBeVisible({ timeout: 15_000 });
  });
});
