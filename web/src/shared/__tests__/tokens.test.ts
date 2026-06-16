import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

// Guards the design-token rename pass: any regression of the gray scale or
// the brand semantic tokens trips this test.
const tokensPath = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../../design-tokens/colors.css",
);
const tokens = readFileSync(tokensPath, "utf8");

describe("design-tokens/colors.css", () => {
  it("uses the reconciled gray scale values", () => {
    expect(tokens).toContain("--custos-gray-900: #111111");
    expect(tokens).toContain("--custos-gray-100: #f1f1f1");
  });

  it("defines the brand semantic in :root", () => {
    expect(tokens).toContain("--primary: oklch(0.205 0 0)");
    expect(tokens).toContain("--primary-foreground: oklch(0.985 0 0)");
    expect(tokens).toContain("--brand: var(--custos-blue-500)");
    expect(tokens).toContain("--brand-foreground: #ffffff");
    expect(tokens).toContain("--brand-tint: var(--custos-blue-50)");
  });

  it("mirrors the brand semantic under .dark", () => {
    const darkBlockMatch = tokens.match(/\.dark\s*{[^}]*}/);
    expect(darkBlockMatch, "expected a .dark { ... } block").not.toBeNull();
    const darkBlock = darkBlockMatch?.[0] ?? "";
    expect(darkBlock).toContain("--brand: var(--custos-blue-400)");
    expect(darkBlock).toContain("--brand-tint:");
  });
});
