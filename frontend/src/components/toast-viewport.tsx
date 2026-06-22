"use client";

import { AnimatePresence, motion } from "motion/react";
import { AlertCircle, AlertTriangle, CheckCircle2, Info, X } from "lucide-react";

import { dismissToast, useToasts, type ToastKind } from "@/lib/toast";

const ICONS: Record<ToastKind, typeof CheckCircle2> = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

// Reuse the app's existing themed notice colors (.message.success/.error/.info from globals.css);
// "warning" has no global modifier, so fall back to "info" colors.
function messageClass(kind: ToastKind): string {
  const modifier = kind === "warning" ? "info" : kind;
  return `message ${modifier}`;
}

export function ToastViewport() {
  const toasts = useToasts();

  return (
    <div
      className="pointer-events-none fixed bottom-4 right-4 z-[2000] flex w-[min(92vw,22rem)] flex-col items-stretch gap-2"
      role="region"
      aria-label="การแจ้งเตือน"
    >
      <AnimatePresence initial={false}>
        {toasts.map((toast) => {
          const Icon = ICONS[toast.kind] ?? Info;
          const isError = toast.kind === "error";
          return (
            <motion.div
              key={toast.id}
              layout
              initial={{ opacity: 0, y: 16, scale: 0.96 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, x: 24, scale: 0.95 }}
              transition={{ duration: 0.22, ease: "easeOut" }}
              role={isError ? "alert" : "status"}
              aria-live={isError ? "assertive" : "polite"}
              className={`${messageClass(toast.kind)} pointer-events-auto items-start gap-2.5 shadow-lg`}
            >
              <Icon size={18} className="mt-0.5 shrink-0" aria-hidden="true" />
              <span className="flex-1 text-sm leading-snug break-words">{toast.text}</span>
              <button
                type="button"
                onClick={() => dismissToast(toast.id)}
                aria-label="ปิดการแจ้งเตือน"
                className="-mr-1 -mt-0.5 shrink-0 rounded-md p-0.5 text-current opacity-60 transition hover:opacity-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2"
              >
                <X size={15} aria-hidden="true" />
              </button>
            </motion.div>
          );
        })}
      </AnimatePresence>
    </div>
  );
}
