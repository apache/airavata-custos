import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("allocation members tab", () => {
  test("admin adds, edits, and removes a member", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/allocations/alloc-002");
    await expect(
      page.getByRole("heading", { name: /Genomic Sequencing - CPU Pool/i }),
    ).toBeVisible();

    await page.getByRole("tab", { name: /^members$/i }).click();

    // Add
    await page.getByRole("button", { name: /\+ add member/i }).click();
    await expect(page.getByRole("heading", { name: /^Add member$/i })).toBeVisible();
    await page.getByLabel(/user id/i).fill("user-new");
    await page.getByRole("button", { name: /^add member$/i }).click();
    await expect(page.getByText(/user-new/i).first()).toBeVisible({ timeout: 15_000 });

    // Edit
    const editButtons = page.getByRole("button", { name: /^edit /i });
    await editButtons.first().click();
    await expect(page.getByRole("heading", { name: /^Edit member$/i })).toBeVisible();
    await page.getByLabel(/^role$/i).selectOption("co_pi");
    await page.getByRole("button", { name: /^save$/i }).click();
    await expect(page.getByRole("heading", { name: /^Edit member$/i })).toBeHidden({
      timeout: 15_000,
    });

    // Remove
    const removeButtons = page.getByRole("button", { name: /^remove /i });
    const initialCount = await removeButtons.count();
    expect(initialCount).toBeGreaterThan(0);
    await removeButtons.first().click();
    await expect
      .poll(async () => removeButtons.count(), { timeout: 15_000 })
      .toBeLessThan(initialCount);
  });
});
