import { type Page, expect } from "@playwright/test";

export type DevLevel = "viewer" | "manager" | "admin";

export async function signInAs(page: Page, level: DevLevel = "admin") {
  await page.goto("/sign-in");
  await page.getByLabel(/dev privilege level/i).selectOption(level);
  await page.getByRole("button", { name: /^sign in$/i }).click();
  // Land on the portal home — the layout redirects unauthed users to /sign-in.
  await page.waitForURL((url) => !url.pathname.startsWith("/sign-in"), { timeout: 20_000 });
}
