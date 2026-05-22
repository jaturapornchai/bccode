"use client";

import { AlertTriangle, Info, ShieldAlert, X } from "lucide-react";
import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

export type ConfirmDialogTone = "danger" | "warning" | "info";

export type ConfirmDialogOptions = {
  title: string;
  description?: ReactNode;
  details?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  tone?: ConfirmDialogTone;
};

type PendingConfirm = Required<Pick<ConfirmDialogOptions, "title" | "confirmLabel" | "cancelLabel" | "tone">>
  & Omit<ConfirmDialogOptions, "title" | "confirmLabel" | "cancelLabel" | "tone">;

const toneClass: Record<ConfirmDialogTone, { icon: string; panel: string; button: "default" | "destructive" }> = {
  danger: {
    icon: "bg-destructive/10 text-destructive",
    panel: "border-destructive/30",
    button: "destructive",
  },
  warning: {
    icon: "bg-amber-500/10 text-amber-600 dark:text-amber-300",
    panel: "border-amber-500/30",
    button: "default",
  },
  info: {
    icon: "bg-primary/10 text-primary",
    panel: "border-border",
    button: "default",
  },
};

export function useConfirmDialog() {
  const resolverRef = useRef<((value: boolean) => void) | null>(null);
  const [pending, setPending] = useState<PendingConfirm | null>(null);

  const close = useCallback((confirmed: boolean) => {
    resolverRef.current?.(confirmed);
    resolverRef.current = null;
    setPending(null);
  }, []);

  const confirm = useCallback((options: ConfirmDialogOptions) => {
    resolverRef.current?.(false);
    return new Promise<boolean>((resolve) => {
      resolverRef.current = resolve;
      setPending({
        title: options.title,
        description: options.description,
        details: options.details,
        confirmLabel: options.confirmLabel ?? "ยืนยัน",
        cancelLabel: options.cancelLabel ?? "ยกเลิก",
        tone: options.tone ?? "warning",
      });
    });
  }, []);

  useEffect(() => {
    if (!pending) return;
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") close(false);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [close, pending]);

  useEffect(() => () => resolverRef.current?.(false), []);

  const confirmationDialog = pending ? (
    <div
      aria-modal="true"
      className="fixed inset-0 z-[100] grid place-items-center bg-black/45 p-2 text-foreground backdrop-blur-[2px]"
      onPointerDown={(event) => {
        if (event.target === event.currentTarget) close(false);
      }}
      role="dialog"
    >
      <section
        className={cn(
          "grid w-full max-w-[min(560px,calc(100vw-16px))] gap-3 rounded-2xl border bg-card p-3 text-card-foreground shadow-xl",
          toneClass[pending.tone].panel,
        )}
      >
        <header className="flex min-w-0 items-start justify-between gap-3">
          <div className="flex min-w-0 items-start gap-2">
            <span className={cn("grid size-10 shrink-0 place-items-center rounded-2xl", toneClass[pending.tone].icon)}>
              {pending.tone === "danger" ? <ShieldAlert size={21} /> : pending.tone === "warning" ? <AlertTriangle size={21} /> : <Info size={21} />}
            </span>
            <div className="min-w-0">
              <h2 className="text-base font-semibold leading-6">{pending.title}</h2>
              {pending.description ? <div className="mt-1 text-sm leading-6 text-muted-foreground">{pending.description}</div> : null}
            </div>
          </div>
          <Button type="button" variant="outline" size="icon" onClick={() => close(false)} aria-label={pending.cancelLabel}>
            <X />
          </Button>
        </header>

        {pending.details ? (
          <div className="max-h-[38dvh] overflow-y-auto rounded-xl border border-border bg-background p-2 text-sm leading-6 text-muted-foreground">
            {pending.details}
          </div>
        ) : null}

        <footer className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={() => close(false)}>
            {pending.cancelLabel}
          </Button>
          <Button type="button" variant={toneClass[pending.tone].button} onClick={() => close(true)}>
            {pending.confirmLabel}
          </Button>
        </footer>
      </section>
    </div>
  ) : null;

  return { confirm, confirmationDialog };
}
