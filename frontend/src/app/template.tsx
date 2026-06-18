"use client";

// Route-level page transition wrapper. Next.js renders template.tsx on every
// navigation, so wrapping children in a motion.div gives a subtle fade+lift on
// each route change. Uses motion's useReducedMotion hook (SSR-safe) so there is
// no hydration mismatch from reading window.matchMedia at module load.

import { motion, useReducedMotion } from "motion/react";

export default function Template({ children }: { children: React.ReactNode }) {
  // Returns undefined during SSR and the first client render, then updates after
  // mount — which keeps the server and client initial markup identical.
  const reduce = useReducedMotion();

  return (
    <motion.div
      initial={reduce ? { opacity: 0 } : { opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={
        reduce
          ? { duration: 0.15 }
          : { duration: 0.35, ease: [0.22, 1, 0.36, 1] }
      }
    >
      {children}
    </motion.div>
  );
}
