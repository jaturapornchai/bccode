"use client";

import { useEffect, useRef } from "react";

export interface FormShortcutsOptions {
  /** Callback triggered on Ctrl+S / Cmd+S */
  onSave?: () => void | Promise<void>;
  /** Callback triggered on Alt+N */
  onNew?: () => void | Promise<void>;
  /** Callback triggered on Escape */
  onCancel?: () => void | Promise<void>;
  /** Disables all shortcuts (e.g. during busy/submitting state) */
  disabled?: boolean;
  /** Guard condition whether saving is allowed (e.g. form is dirty, valid) */
  canSave?: boolean;
  /** Whether the hook is currently active in the page */
  enabled?: boolean;
}

export interface ShortcutKeyEvent {
  key: string;
  ctrlKey?: boolean;
  metaKey?: boolean;
  altKey?: boolean;
  defaultPrevented?: boolean;
  preventDefault?: () => void;
}

/**
 * Pure evaluation function for shortcut keyboard events:
 * - Ctrl+S / Cmd+S: Save document / draft (calls preventDefault to suppress browser dialog)
 * - Alt+N: New document / create entry
 * - Escape: Cancel edit / close workbench form
 */
export function evaluateFormShortcut(
  e: ShortcutKeyEvent,
  options: {
    onSave?: () => void | Promise<void>;
    onNew?: () => void | Promise<void>;
    onCancel?: () => void | Promise<void>;
    disabled?: boolean;
    canSave?: boolean;
  }
): "save" | "new" | "cancel" | null {
  if (e.defaultPrevented || options.disabled) return null;

  const key = e.key.toLowerCase();
  const isCtrlOrMeta = Boolean(e.ctrlKey || e.metaKey);

  // 1. Ctrl+S or Cmd+S -> Save
  if (isCtrlOrMeta && key === "s") {
    e.preventDefault?.();
    if (options.canSave !== false && options.onSave) {
      void options.onSave();
    }
    return "save";
  }

  // 2. Alt+N -> New Document
  if (Boolean(e.altKey) && !isCtrlOrMeta && key === "n") {
    e.preventDefault?.();
    if (options.onNew) {
      void options.onNew();
    }
    return "new";
  }

  // 3. Escape -> Cancel Edit
  if (e.key === "Escape" && !e.altKey && !isCtrlOrMeta) {
    if (options.onCancel) {
      e.preventDefault?.();
      void options.onCancel();
    }
    return "cancel";
  }

  return null;
}

/**
 * Global Keyboard Shortcuts for Transaction Forms & Workbenches:
 * - Ctrl+S / Cmd+S: Save document / draft
 * - Alt+N: New document / create entry
 * - Escape: Cancel current edit / close workbench form
 */
export function useFormShortcuts({
  onSave,
  onNew,
  onCancel,
  disabled = false,
  canSave = true,
  enabled = true,
}: FormShortcutsOptions) {
  const onSaveRef = useRef(onSave);
  const onNewRef = useRef(onNew);
  const onCancelRef = useRef(onCancel);
  const disabledRef = useRef(disabled);
  const canSaveRef = useRef(canSave);

  useEffect(() => {
    onSaveRef.current = onSave;
    onNewRef.current = onNew;
    onCancelRef.current = onCancel;
    disabledRef.current = disabled;
    canSaveRef.current = canSave;
  });

  useEffect(() => {
    if (!enabled || typeof window === "undefined") return;

    const handleKeyDown = (e: KeyboardEvent) => {
      evaluateFormShortcut(e, {
        onSave: onSaveRef.current,
        onNew: onNewRef.current,
        onCancel: onCancelRef.current,
        disabled: disabledRef.current,
        canSave: canSaveRef.current,
      });
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [enabled]);
}
