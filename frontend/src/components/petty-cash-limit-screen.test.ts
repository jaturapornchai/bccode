import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const read = (file: string) => readFileSync(resolve(process.cwd(), file), "utf8");
const pettyCashScreen = read("src/components/petty-cash-limit-screen.tsx");

describe("petty-cash-limit-screen", () => {
  it("follows the language prop through the backend dictionary, not localStorage or a local t(th,en)", () => {
    expect(pettyCashScreen).toContain("useBackendLanguage(language");
    expect(pettyCashScreen).not.toContain("localStorage");
    expect(pettyCashScreen).not.toMatch(/\bt\(\s*"[^"]*"\s*,\s*"[^"]*"\s*\)/);
  });

  it("warns before the tab closes while there are unsaved edits", () => {
    expect(pettyCashScreen).toContain('window.addEventListener("beforeunload"');
    expect(pettyCashScreen).toContain("dirty");
  });

  it("never lets an icon-only button go unlabeled", () => {
    const buttons = pettyCashScreen.split("<Button").slice(1).map((chunk) => chunk.split("</Button>")[0] ?? "");
    for (const button of buttons.filter((chunk) => chunk.includes('size="icon"'))) {
      expect(button).toContain("aria-label=");
    }
  });

  it("loads and saves petty cash using the cheque-api methods", () => {
    expect(pettyCashScreen).toContain("listPettyCash(language)");
    expect(pettyCashScreen).toContain("savePettyCash(payload, language)");
  });

  it("limits code and name input lengths according to database schema", () => {
    expect(pettyCashScreen).toContain("maxLength={20}");
    expect(pettyCashScreen).toContain("maxLength={100}");
  });

  it("supports recording and viewing petty cash replenishment and expenses (Champ 5178/5179)", () => {
    expect(pettyCashScreen).toContain("createPettyCashTrans");
    expect(pettyCashScreen).toContain("listPettyCashTrans");
    expect(pettyCashScreen).toContain("deletePettyCashTrans");
    expect(pettyCashScreen).toContain("handleOpenTransModal");
  });
});
