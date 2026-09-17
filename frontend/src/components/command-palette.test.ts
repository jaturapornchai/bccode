import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi, beforeEach } from "vitest";
import {
  CommandPalette,
  getStoredRecentRoutes,
  saveStoredRecentRoutes,
  MAX_RECENT_ITEMS,
} from "./command-palette";

// Mock next/navigation useRouter and usePathname
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
  usePathname: () => "/gl/gl-journals",
}));

describe("CommandPalette Component", () => {
  it("renders without crashing in static markup (clean SSR)", () => {
    const html = renderToStaticMarkup(createElement(CommandPalette));
    expect(html).toBe("");
  });
});

describe("Recent Screens Storage Helpers", () => {
  beforeEach(() => {
    // In node test environment, mock window / localStorage safely
    const storage: Record<string, string> = {};
    (globalThis as unknown as { localStorage: unknown }).localStorage = {
      getItem: (key: string) => storage[key] ?? null,
      setItem: (key: string, val: string) => {
        storage[key] = val;
      },
      removeItem: (key: string) => {
        delete storage[key];
      },
    };
  });

  it("retrieves empty array when nothing is stored", () => {
    expect(getStoredRecentRoutes()).toEqual([]);
  });

  it("saves and retrieves recent routes up to MAX_RECENT_ITEMS", () => {
    const testRoutes = [
      "/gl/chartofaccounts",
      "/gl/gl-journals",
      "/report/financial-position",
      "/report/income-statement",
    ];
    saveStoredRecentRoutes(testRoutes);
    expect(getStoredRecentRoutes()).toEqual(testRoutes);

    // Should cap at MAX_RECENT_ITEMS
    const excessRoutes = Array.from({ length: 15 }, (_, i) => `/route-${i}`);
    saveStoredRecentRoutes(excessRoutes);
    const retrieved = getStoredRecentRoutes();
    expect(retrieved.length).toBe(MAX_RECENT_ITEMS);
    expect(retrieved[0]).toBe("/route-0");
  });
});
