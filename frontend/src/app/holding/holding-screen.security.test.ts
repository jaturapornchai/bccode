import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(fileURLToPath(new URL("./holding-screen.tsx", import.meta.url)), "utf8");

describe("holding screen authorization", () => {
  it("uses the backend capability instead of inferring create permission from email", () => {
    expect(source).toContain("payload.cancreateholding === true");
    expect(source).not.toMatch(/canCreateHolding\s*=\s*Boolean\(auth\?\.profile\?\.email/);
  });
});
