import { describe, expect, it, vi } from "vitest";
import { handleTabularEnterKey, useTabularEnterNav } from "./use-tabular-enter-nav";

describe("handleTabularEnterKey", () => {
  it("focuses next input on Enter", () => {
    const input1 = {
      tagName: "INPUT",
      hasAttribute: () => false,
      focus: vi.fn(),
      select: vi.fn(),
    } as unknown as HTMLElement;

    const input2 = {
      tagName: "INPUT",
      hasAttribute: () => false,
      focus: vi.fn(),
      select: vi.fn(),
    } as unknown as HTMLElement;

    const container = {
      querySelectorAll: () => [input1, input2],
    } as unknown as HTMLElement;

    const preventDefault = vi.fn();

    const handled = handleTabularEnterKey(
      {
        key: "Enter",
        target: input1,
        currentTarget: container,
        preventDefault,
      },
      {}
    );

    expect(handled).toBe(true);
    expect(preventDefault).toHaveBeenCalled();
    expect((input2 as unknown as { focus: ReturnType<typeof vi.fn> }).focus).toHaveBeenCalled();
  });

  it("calls onAddNewRow when pressing Enter on the last input", () => {
    const input1 = {
      tagName: "INPUT",
      hasAttribute: () => false,
      focus: vi.fn(),
    } as unknown as HTMLElement;

    const container = {
      querySelectorAll: () => [input1],
    } as unknown as HTMLElement;

    const preventDefault = vi.fn();
    const onAddNewRow = vi.fn();

    const handled = handleTabularEnterKey(
      {
        key: "Enter",
        target: input1,
        currentTarget: container,
        preventDefault,
      },
      { onAddNewRow }
    );

    expect(handled).toBe(true);
    expect(preventDefault).toHaveBeenCalled();
    expect(onAddNewRow).toHaveBeenCalled();
  });

  it("ignores Shift+Enter or Ctrl+Enter", () => {
    const input1 = { tagName: "INPUT" } as HTMLElement;
    const handled = handleTabularEnterKey({
      key: "Enter",
      shiftKey: true,
      target: input1,
    });
    expect(handled).toBe(false);
  });

  it("exports useTabularEnterNav without error", () => {
    expect(typeof useTabularEnterNav).toBe("function");
  });
});
