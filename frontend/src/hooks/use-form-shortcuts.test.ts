import { describe, expect, it, vi } from "vitest";
import { evaluateFormShortcut, useFormShortcuts } from "./use-form-shortcuts";

describe("evaluateFormShortcut", () => {
  it("triggers onSave and prevents default on Ctrl+S", () => {
    const onSave = vi.fn();
    const preventDefault = vi.fn();
    const result = evaluateFormShortcut(
      { key: "s", ctrlKey: true, preventDefault },
      { onSave, canSave: true }
    );

    expect(result).toBe("save");
    expect(preventDefault).toHaveBeenCalled();
    expect(onSave).toHaveBeenCalled();
  });

  it("does not trigger onSave when canSave is false", () => {
    const onSave = vi.fn();
    const preventDefault = vi.fn();
    const result = evaluateFormShortcut(
      { key: "s", ctrlKey: true, preventDefault },
      { onSave, canSave: false }
    );

    expect(result).toBe("save");
    expect(preventDefault).toHaveBeenCalled();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("triggers onNew and prevents default on Alt+N", () => {
    const onNew = vi.fn();
    const preventDefault = vi.fn();
    const result = evaluateFormShortcut(
      { key: "n", altKey: true, preventDefault },
      { onNew }
    );

    expect(result).toBe("new");
    expect(preventDefault).toHaveBeenCalled();
    expect(onNew).toHaveBeenCalled();
  });

  it("triggers onCancel and prevents default on Escape", () => {
    const onCancel = vi.fn();
    const preventDefault = vi.fn();
    const result = evaluateFormShortcut(
      { key: "Escape", preventDefault },
      { onCancel }
    );

    expect(result).toBe("cancel");
    expect(preventDefault).toHaveBeenCalled();
    expect(onCancel).toHaveBeenCalled();
  });

  it("does not trigger callbacks when disabled is true", () => {
    const onSave = vi.fn();
    const preventDefault = vi.fn();
    const result = evaluateFormShortcut(
      { key: "s", ctrlKey: true, preventDefault },
      { onSave, disabled: true }
    );

    expect(result).toBeNull();
    expect(preventDefault).not.toHaveBeenCalled();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("does not trigger callbacks when defaultPrevented is true", () => {
    const onSave = vi.fn();
    const result = evaluateFormShortcut(
      { key: "s", ctrlKey: true, defaultPrevented: true },
      { onSave }
    );

    expect(result).toBeNull();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("exports useFormShortcuts hook function without crashing in node", () => {
    expect(typeof useFormShortcuts).toBe("function");
  });
});
