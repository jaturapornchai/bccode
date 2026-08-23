import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const setupScreens = [
  "settings",
  "menu",
  "currency",
  "workspace",
  "activelanguages",
  "company",
  "permissiondefinition",
  "permissiongroup",
  "user",
  "useraccessaudit",
] as const;

describe("setup basic manuals", () => {
  it.each(setupScreens)("%s has concept, workflow, guidance, and limitations", (screen) => {
    const manual = JSON.parse(
      readFileSync(resolve(process.cwd(), "manual", `${screen}.json`), "utf8"),
    ) as {
      screen: string;
      translations: Record<
        "en" | "th",
        {
          commonMistakes: string[];
          config: string[];
          limitations: string[];
          nextSteps: string[];
          objective: string;
          sources: Array<{ href: string; name: string }>;
          workflow: string[];
        }
      >;
    };

    expect(manual.screen).toBe(screen);
    expect(manual.translations.th.objective).toContain("แนวคิด");
    expect(manual.translations.en.objective).toContain("Concept");
    for (const language of ["th", "en"] as const) {
      expect(manual.translations[language].sources.length).toBeGreaterThanOrEqual(1);
      expect(manual.translations[language].sources.every((source) => source.href.startsWith("https://"))).toBe(true);
      expect(manual.translations[language].workflow.length).toBeGreaterThanOrEqual(4);
      expect(manual.translations[language].config.length).toBeGreaterThanOrEqual(3);
      expect(manual.translations[language].commonMistakes.length).toBeGreaterThanOrEqual(2);
      expect(manual.translations[language].limitations.length).toBeGreaterThanOrEqual(2);
      expect(manual.translations[language].nextSteps.length).toBeGreaterThanOrEqual(1);
    }
  });

  it("shows the active screen manual in the embedded setup header", () => {
    const source = readFileSync(
      resolve(process.cwd(), "src", "app", "workspace", "workspace-screen.tsx"),
      "utf8",
    );

    expect(source).toContain("getSystemSettingConfig(activeAccessRoute)?.manual");
    expect(source).toContain('label={language === "th" ? "คู่มือเบื้องต้น" : t(language, "manual")}');
  });

  it("routes screens without reviewed content to the available-guide index", () => {
    const source = readFileSync(
      resolve(process.cwd(), "src", "app", "manual-link.tsx"),
      "utf8",
    );

    for (const screen of setupScreens) expect(source).toContain(`"${screen}"`);
    expect(source).toContain("/manual?lang=");
    expect(source).toContain("encodeURIComponent(screen)");
  });

  it("marks guide content with its actual reviewed language", () => {
    const indexSource = readFileSync(
      resolve(process.cwd(), "src", "app", "manual", "page.tsx"),
      "utf8",
    );
    const detailSource = readFileSync(
      resolve(process.cwd(), "src", "app", "manual", "[screen]", "page.tsx"),
      "utf8",
    );

    for (const source of [indexSource, detailSource]) {
      expect(source).toContain('requestedLanguage === "th" ? "th" : "en"');
      expect(source).toContain("lang={language}");
      expect(source).toContain('item.code === "th" || item.code === "en"');
    }
  });
});
