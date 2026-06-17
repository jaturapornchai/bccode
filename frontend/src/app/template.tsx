"use client";

// Route-level page transition wrapper. Next.js renders template.tsx on every
// navigation, so wrapping children in a motion.div gives a subtle fade+lift on
// each route change without AnimatePresence exit work (which template.tsx can't do).
// Respects prefers-reduced-motion: falls back to a tiny opacity fade only.

import { motion } from "motion/react";

const prefersReducedMotion =
  typeof window !== "undefined" &&
  window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;

export default function Template({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={prefersReducedMotion ? { opacity: 0 } : { opacity: 0, y: 8 }}
      animate={prefersReducedMotion ? { opacity: 1 } : { opacity: 1, y: 0 }}
      transition={prefersReducedMotion ? { duration: 0.15 } : { duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
    >
      {children}
    </motion.div>
  );
}
