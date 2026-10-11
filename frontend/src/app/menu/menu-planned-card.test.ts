import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import { MenuPlannedCard } from "./menu-planned-card";

describe("MenuPlannedCard", () => {
  it("shows the รอพัฒนา chip with data-menu-pending and no form controls (2026-10-11)", () => {
    const html = renderToStaticMarkup(
      createElement(MenuPlannedCard, {
        route: "/settings/currency",
        title: "กำหนดสกุลเงิน",
        language: "th",
        backendLanguage: {} as BackendLanguageDictionary,
      }),
    );
    expect(html).toContain('data-menu-pending="/settings/currency"');
    expect(html).toContain("รอพัฒนา");
    expect(html).toContain("เมนูในแผนพัฒนา");
    expect(html).toContain("กำหนดสกุลเงิน");
    expect(html).not.toMatch(/<(input|textarea|select|form)[ >]/);
  });

  it("renders the chip in English from the same language key", () => {
    const html = renderToStaticMarkup(
      createElement(MenuPlannedCard, {
        route: "/transaction/etax",
        title: "e-Tax Invoice",
        language: "en",
        backendLanguage: {} as BackendLanguageDictionary,
      }),
    );
    expect(html).toContain('data-menu-pending="/transaction/etax"');
    expect(html).toContain("In development");
  });
});
