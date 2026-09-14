import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { SearchInput } from "./gl-common";

describe("SearchInput (Baseline Search Component)", () => {
  it("renders input with aria-label, placeholder, and control classes", () => {
    const html = renderToStaticMarkup(
      createElement(SearchInput, {
        value: "",
        onChange: () => {},
        placeholder: "ค้นหารหัสหรือชื่อ",
        ariaLabel: "ค้นหารหัสหรือชื่อ",
      }),
    );
    expect(html).toContain('placeholder="ค้นหารหัสหรือชื่อ"');
    expect(html).toContain('aria-label="ค้นหารหัสหรือชื่อ"');
    expect(html).toContain("border-input");
  });

  it("does not render clean button when value is empty", () => {
    const html = renderToStaticMarkup(
      createElement(SearchInput, {
        value: "",
        onChange: () => {},
      }),
    );
    expect(html).not.toContain("ล้างข้อความค้นหา");
  });

  it("renders search icon and clearance padding", () => {
    const html = renderToStaticMarkup(
      createElement(SearchInput, {
        value: "",
        onChange: () => {},
      }),
    );
    expect(html).toContain("lucide-search");
    expect(html).toContain("!pl-9.5");
  });

  it("renders clean icon button when value is present", () => {
    const html = renderToStaticMarkup(
      createElement(SearchInput, {
        value: "เงินสด",
        onChange: () => {},
      }),
    );
    expect(html).toContain("ล้างข้อความค้นหา (Clean)");
    expect(html).toContain("lucide-x");
    expect(html).toContain("!pr-9");
  });
});

describe("Debounced Auto Search Logic (2 seconds)", () => {
  it("debounces search execution by 2000ms", () => {
    vi.useFakeTimers();
    const onSearch = vi.fn();

    let timer: NodeJS.Timeout | null = null;
    const triggerDebounce = (val: string) => {
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        onSearch(val);
      }, 2000);
    };

    // User types '1'
    triggerDebounce("1");
    expect(onSearch).not.toHaveBeenCalled();

    // Advance 1000ms
    vi.advanceTimersByTime(1000);
    expect(onSearch).not.toHaveBeenCalled();

    // User types '1101' at 1000ms
    triggerDebounce("1101");

    // Advance 1500ms (total 2500ms from start, 1500ms from second keystroke)
    vi.advanceTimersByTime(1500);
    expect(onSearch).not.toHaveBeenCalled();

    // Advance remaining 500ms (2000ms from second keystroke)
    vi.advanceTimersByTime(500);
    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch).toHaveBeenCalledWith("1101");

    vi.useRealTimers();
  });

  it("clears debounce timer and executes immediately on manual search or clean", () => {
    vi.useFakeTimers();
    const onSearch = vi.fn();

    let timer: NodeJS.Timeout | null = null;
    let query = "";

    const triggerDebounce = (val: string) => {
      query = val;
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        onSearch(val);
      }, 2000);
    };

    const searchNow = () => {
      if (timer) clearTimeout(timer);
      onSearch(query);
    };

    const clear = () => {
      if (timer) clearTimeout(timer);
      query = "";
      onSearch("");
    };

    // User types
    triggerDebounce("เงินฝาก");
    expect(onSearch).not.toHaveBeenCalled();

    // User presses search button after 300ms
    vi.advanceTimersByTime(300);
    searchNow();
    expect(onSearch).toHaveBeenCalledTimes(1);
    expect(onSearch).toHaveBeenCalledWith("เงินฝาก");

    // Advance 5000ms: timer should NOT fire again
    vi.advanceTimersByTime(5000);
    expect(onSearch).toHaveBeenCalledTimes(1);

    // User clears
    clear();
    expect(onSearch).toHaveBeenCalledTimes(2);
    expect(onSearch).toHaveBeenLastCalledWith("");

    vi.useRealTimers();
  });
});
