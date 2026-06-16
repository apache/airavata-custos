import { expect, test } from "@playwright/test";
import { signInAs } from "./fixtures/auth";

test.describe("change request approver queue", () => {
  test("admin sees the queue and can approve a pending request", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/change-requests");
    await expect(page.getByRole("heading", { name: /^Change requests$/ })).toBeVisible();

    const approveBtn = page.getByRole("button", { name: /^Approve cr-001$/ });
    await expect(approveBtn).toBeVisible({ timeout: 15_000 });
    await approveBtn.click();
    await expect(page.getByText(/approved cr-001/i)).toBeVisible({ timeout: 15_000 });
  });

  test("admin can reject a pending request", async ({ page }) => {
    await signInAs(page, "admin");
    await page.goto("/change-requests");
    await expect(page.getByRole("heading", { name: /^Change requests$/ })).toBeVisible();

    const rejectBtn = page.getByRole("button", { name: /^Reject cr-002$/ });
    await expect(rejectBtn).toBeVisible({ timeout: 15_000 });
    await rejectBtn.click();
    await expect(page.getByText(/rejected cr-002/i)).toBeVisible({ timeout: 15_000 });
  });
});
