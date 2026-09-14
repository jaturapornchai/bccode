import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const inspectorSource = readFileSync(
  fileURLToPath(new URL("./dev-dom-inspector.tsx", import.meta.url)),
  "utf8",
);

describe("DevDomInspector", () => {
  it("does not restrict rendering to development mode so it runs on production", () => {
    // NODE_ENV !== "development" guard must not be present
    expect(inspectorSource).not.toContain('process.env.NODE_ENV !== "development"');
    expect(inspectorSource).not.toContain("process.env.NODE_ENV === 'development'");
  });

  it("contains the Copy DOM floating button and keyboard shortcut indicator", () => {
    expect(inspectorSource).toContain("data-dev-dom-inspector");
    expect(inspectorSource).toContain("Copy DOM");
    expect(inspectorSource).toContain("Alt+คลิก");
  });

  it("supports safe clipboard copying with fallback", () => {
    expect(inspectorSource).toContain("copyTextSafely");
    expect(inspectorSource).toContain("navigator.clipboard.writeText");
    expect(inspectorSource).toContain("execCommand");
  });

  it("handles mouse events and capture-phase click interception", () => {
    expect(inspectorSource).toContain('window.addEventListener("click", handleClick, true)');
    expect(inspectorSource).toContain("e.stopImmediatePropagation()");
  });
});
