import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { CommandPalette } from "./command-palette";

// Mock next/navigation useRouter
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
}));

describe("CommandPalette Component", () => {
  it("renders without crashing in static markup", () => {
    const html = renderToStaticMarkup(createElement(CommandPalette));
    // When not mounted on client, returns null (clean SSR)
    expect(html).toBe("");
  });
});
