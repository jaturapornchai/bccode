import { useSyncExternalStore } from "react";

// Global toast/notification store — framework-agnostic singleton so any screen (or non-React
// code) can raise a toast that renders once, bottom-right, via <ToastViewport/> in the root
// layout. Toasts auto-dismiss after DEFAULT_DURATION. This replaces the per-screen inline
// `notice` banners; legacy `setNotice({type,text})` call sites forward through `pushNotice`.

export type ToastKind = "success" | "error" | "info" | "warning";

export type ToastItem = {
  id: number;
  kind: ToastKind;
  text: string;
};

export const TOAST_DURATION = 5000;
const MAX_TOASTS = 5;

const EMPTY: ToastItem[] = [];
let toasts: ToastItem[] = EMPTY;
let nextId = 1;
const listeners = new Set<() => void>();
const timers = new Map<number, ReturnType<typeof setTimeout>>();

function emit() {
  for (const listener of listeners) listener();
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

function getSnapshot() {
  return toasts;
}

function getServerSnapshot() {
  return EMPTY;
}

function clearTimer(id: number) {
  const timer = timers.get(id);
  if (timer) {
    clearTimeout(timer);
    timers.delete(id);
  }
}

export function dismissToast(id: number) {
  clearTimer(id);
  if (toasts.some((toast) => toast.id === id)) {
    toasts = toasts.filter((toast) => toast.id !== id);
    emit();
  }
}

function normalizeKind(kind: ToastKind | string | undefined): ToastKind {
  return kind === "success" || kind === "error" || kind === "warning" ? kind : "info";
}

export function notify(kind: ToastKind, text: string, duration: number = TOAST_DURATION): number {
  const message = (text ?? "").trim();
  if (!message) return -1;

  const id = nextId++;
  toasts = [...toasts, { id, kind, text: message }];

  // Cap the visible stack — drop the oldest beyond the limit (and cancel its timer).
  if (toasts.length > MAX_TOASTS) {
    const overflow = toasts.slice(0, toasts.length - MAX_TOASTS);
    overflow.forEach((toast) => clearTimer(toast.id));
    toasts = toasts.slice(toasts.length - MAX_TOASTS);
  }

  if (duration > 0 && typeof window !== "undefined") {
    timers.set(id, setTimeout(() => dismissToast(id), duration));
  }

  emit();
  return id;
}

export const toast = {
  success: (text: string) => notify("success", text),
  error: (text: string) => notify("error", text),
  info: (text: string) => notify("info", text),
  warning: (text: string) => notify("warning", text),
};

// Compatibility shim for the legacy per-screen `setNotice({type,text})` API: screens swap
// `const [notice, setNotice] = useState<Notice>(null)` for `const setNotice = pushNotice` and
// delete their inline notice render block. A `null` (the old "clear" call) is a no-op because
// toasts self-expire.
export type LegacyNotice = { type: ToastKind | string; text?: string } | null | undefined;

export function pushNotice(notice: LegacyNotice): void {
  if (!notice || !notice.text) return;
  notify(normalizeKind(notice.type), notice.text);
}

export function useToasts(): ToastItem[] {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
