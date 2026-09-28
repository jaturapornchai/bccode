import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const inspectorSource = readFileSync(
  fileURLToPath(new URL("./dev-dom-inspector.tsx", import.meta.url)),
  "utf8",
);

const DEV_GATE = '{process.env.NODE_ENV === "development" && (';

describe("DevDomInspector", () => {
  it("keeps the Alt+click feature on production (ADR 2026-09-12): only the floating button is dev-gated", () => {
    // No component-wide guard: the listeners must mount on production builds too.
    expect(inspectorSource).not.toContain('process.env.NODE_ENV !== "development"');
    expect(inspectorSource).not.toContain("process.env.NODE_ENV === 'development'");
    expect(inspectorSource.split("process.env.NODE_ENV").length - 1).toBe(1);
    expect(inspectorSource.indexOf(DEV_GATE)).toBeGreaterThan(
      inspectorSource.indexOf('window.addEventListener("click", handleClick, true)'),
    );
  });

  it("renders the floating Copy DOM button only in development so it never covers production controls", () => {
    const gateIndex = inspectorSource.indexOf(DEV_GATE);
    const buttonWrapperIndex = inspectorSource.indexOf("fixed bottom-4 left-4 z-[999999]");
    expect(gateIndex).toBeGreaterThan(-1);
    expect(buttonWrapperIndex).toBeGreaterThan(gateIndex);
    // The toggle into "active" mode lives inside the gated button only.
    const toggleIndex = inspectorSource.indexOf("setActive(next)");
    expect(toggleIndex).toBeGreaterThan(gateIndex);
    expect(inspectorSource.split("setActive(true)").length - 1).toBe(0);
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

  it("trusts the event's own altKey so a stuck Alt (Alt+Tab) cannot swallow ordinary clicks", () => {
    expect(inspectorSource).toContain("if (!activeRef.current && !e.altKey) return;");
    expect(inspectorSource).not.toContain("if (!activeRef.current && !altHeldRef.current) return;");
    expect(inspectorSource).toContain("if (!activeRef.current && !e.altKey) {");
  });

  it("resets Alt/highlight state when the window loses focus or the tab is hidden", () => {
    expect(inspectorSource).toContain('window.addEventListener("blur", resetAlt)');
    expect(inspectorSource).toContain('window.removeEventListener("blur", resetAlt)');
    expect(inspectorSource).toContain('document.addEventListener("visibilitychange", handleVisibilityChange)');
    expect(inspectorSource).toContain('document.removeEventListener("visibilitychange", handleVisibilityChange)');
    expect(inspectorSource).toContain('document.visibilityState === "hidden"');
  });

  it("still exits active mode with Escape", () => {
    expect(inspectorSource).toContain('e.key === "Escape" && activeRef.current');
  });
});
