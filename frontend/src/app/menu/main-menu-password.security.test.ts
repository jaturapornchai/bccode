import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(fileURLToPath(new URL("./main-menu-screen.tsx", import.meta.url)), "utf8");

describe("main menu password security", () => {
  it("has no shared/default password flow", () => {
    expect(source).not.toMatch(/12345|isdefaultpassword|isDefaultPassword|default_password_warning/i);
  });

  it("validates 15-64 characters and clears the already-revoked session after a successful change", () => {
    expect(source).toContain("newPassword.length < 15 || newPassword.length > 64");
    expect(source).toMatch(/setPasswordDialogOpen\(false\);\s*clearAuthSession\(\);/s);
  });
});
