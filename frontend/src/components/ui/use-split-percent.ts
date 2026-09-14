"use client";

import { useCallback, useEffect, useState } from "react";

export interface UseSplitPercentOptions {
  storageKey?: string;
  defaultLeft?: number;
  min?: number;
  max?: number;
  step?: number;
  mode?: "percent" | "pixel";
  containerRef?: React.RefObject<HTMLElement | null>;
}

export interface UseSplitPercentResult {
  splitPercent: number;
  setSplitPercent: React.Dispatch<React.SetStateAction<number>>;
  isResizing: boolean;
  setIsResizing: React.Dispatch<React.SetStateAction<boolean>>;
  startResize: (event: React.PointerEvent<HTMLDivElement>) => void;
  adjustWithKeyboard: (event: React.KeyboardEvent<HTMLDivElement>) => void;
  resetSplit: () => void;
  min: number;
  max: number;
}

export function useSplitPercent({
  storageKey,
  defaultLeft = 50,
  min = 20,
  max = 80,
  step,
  mode = "percent",
  containerRef,
}: UseSplitPercentOptions = {}): UseSplitPercentResult {
  const defaultStep = step ?? (mode === "pixel" ? 16 : 2);

  const [splitPercent, setSplitPercent] = useState<number>(() => {
    if (!storageKey || typeof window === "undefined") return defaultLeft;
    try {
      const saved = window.localStorage.getItem(storageKey);
      if (saved) {
        const parsed = Number(saved);
        if (Number.isFinite(parsed)) {
          return Math.min(max, Math.max(min, parsed));
        }
      }
    } catch {
      // Ignore localStorage errors
    }
    return defaultLeft;
  });

  const [isResizing, setIsResizing] = useState(false);

  useEffect(() => {
    if (!storageKey || typeof window === "undefined") return;
    try {
      const saved = window.localStorage.getItem(storageKey);
      if (saved) {
        const parsed = Number(saved);
        if (Number.isFinite(parsed)) {
          setSplitPercent(Math.min(max, Math.max(min, parsed)));
        }
      }
    } catch {
      // Ignore localStorage errors
    }
  }, [storageKey, min, max]);

  const saveValue = useCallback(
    (value: number) => {
      if (!storageKey || typeof window === "undefined") return;
      try {
        window.localStorage.setItem(storageKey, String(Math.round(value)));
      } catch {
        // Ignore localStorage errors
      }
    },
    [storageKey],
  );

  const startResize = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      const container = containerRef?.current ?? event.currentTarget.parentElement;
      if (!container) return;
      event.preventDefault();
      setIsResizing(true);
      const rect = container.getBoundingClientRect();

      const update = (clientX: number) => {
        let next: number;
        if (mode === "pixel") {
          next = Math.round(clientX - rect.left);
        } else {
          next = rect.width > 0 ? ((clientX - rect.left) / rect.width) * 100 : defaultLeft;
        }
        const clamped = Math.min(max, Math.max(min, next));
        setSplitPercent(clamped);
        saveValue(clamped);
      };

      update(event.clientX);

      const onMove = (moveEvent: PointerEvent) => update(moveEvent.clientX);
      const onUp = () => {
        window.removeEventListener("pointermove", onMove);
        window.removeEventListener("pointerup", onUp);
        setIsResizing(false);
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
      };

      document.body.style.cursor = "col-resize";
      document.body.style.userSelect = "none";
      window.addEventListener("pointermove", onMove);
      window.addEventListener("pointerup", onUp);
    },
    [containerRef, defaultLeft, max, min, mode, saveValue],
  );

  const adjustWithKeyboard = useCallback(
    (event: React.KeyboardEvent<HTMLDivElement>) => {
      if (event.key === "ArrowLeft") {
        event.preventDefault();
        setSplitPercent((prev) => {
          const next = Math.max(min, prev - defaultStep);
          saveValue(next);
          return next;
        });
      } else if (event.key === "ArrowRight") {
        event.preventDefault();
        setSplitPercent((prev) => {
          const next = Math.min(max, prev + defaultStep);
          saveValue(next);
          return next;
        });
      } else if (event.key === "Home") {
        event.preventDefault();
        setSplitPercent(min);
        saveValue(min);
      } else if (event.key === "End") {
        event.preventDefault();
        setSplitPercent(max);
        saveValue(max);
      }
    },
    [defaultStep, max, min, saveValue],
  );

  const resetSplit = useCallback(() => {
    setSplitPercent(defaultLeft);
    saveValue(defaultLeft);
  }, [defaultLeft, saveValue]);

  return {
    splitPercent,
    setSplitPercent,
    isResizing,
    setIsResizing,
    startResize,
    adjustWithKeyboard,
    resetSplit,
    min,
    max,
  };
}
