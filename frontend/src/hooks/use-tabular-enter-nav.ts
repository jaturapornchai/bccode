"use client";

import { useCallback } from "react";

export interface TabularEnterNavOptions {
  /** Callback triggered when Enter is pressed on the last input of the table */
  onAddNewRow?: () => void;
  /** Custom selector for focusable elements inside the table container */
  selector?: string;
  /** Whether navigation is enabled */
  enabled?: boolean;
}

export interface MinimalKeyboardEvent {
  key: string;
  shiftKey?: boolean;
  ctrlKey?: boolean;
  altKey?: boolean;
  metaKey?: boolean;
  target: unknown;
  currentTarget?: unknown;
  preventDefault?: () => void;
}

/**
 * Pure helper to handle Enter key focus jumping across table input cells:
 * - Enter moves to the next input in the row
 * - At the last input of the last row, calls onAddNewRow to create a new row
 * - Simulates classic desktop accounting (Numpad enter-to-next-cell) flow
 */
export function handleTabularEnterKey(
  e: MinimalKeyboardEvent,
  options: TabularEnterNavOptions = {}
): boolean {
  if (options.enabled === false) return false;
  const isEnter = e.key === "Enter";
  const isTab = e.key === "Tab";
  if ((!isEnter && !isTab) || e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
    return false;
  }

  const target = e.target as HTMLElement | null;
  if (!target) return false;

  const tagName = target.tagName?.toLowerCase();
  if (tagName !== "input" && tagName !== "select") {
    return false;
  }

  const container =
    (e.currentTarget as HTMLElement | null) ||
    target.closest("table") ||
    target.closest("form");
  if (!container || typeof container.querySelectorAll !== "function") return false;

  const selector =
    options.selector ||
    'input:not([disabled]):not([readonly]):not([type="hidden"]), select:not([disabled])';

  const elements = Array.from(container.querySelectorAll<HTMLElement>(selector)).filter(
    (el) => !el.hasAttribute("data-enter-ignore")
  );

  const currentIndex = elements.indexOf(target);
  if (currentIndex === -1) return false;

  // On Tab: let browser handle normal tab movement until the last element
  if (isTab) {
    if (currentIndex === elements.length - 1 && options.onAddNewRow) {
      e.preventDefault?.();
      options.onAddNewRow();
      return true;
    }
    return false;
  }

  // On Enter: always advance focus or add new row
  e.preventDefault?.();

  if (currentIndex < elements.length - 1) {
    const nextEl = elements[currentIndex + 1];
    nextEl.focus?.();
    if (typeof (nextEl as HTMLInputElement).select === "function") {
      (nextEl as HTMLInputElement).select();
    }
    return true;
  }

  if (options.onAddNewRow) {
    options.onAddNewRow();
    return true;
  }

  return false;
}

/**
 * Hook for table-based accounting entry with Numpad / Enter key focus traversal
 */
export function useTabularEnterNav(options: TabularEnterNavOptions = {}) {
  const onKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      handleTabularEnterKey(e, options);
    },
    [options]
  );

  return { onKeyDown };
}
