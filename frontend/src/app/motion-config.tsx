"use client";

// Wraps the app so motion respects the user's prefers-reduced-motion setting
// automatically, without each motion.div needing to read window.matchMedia.
// This is SSR-safe: motion renders the final state on the server and only
// applies transforms on the client after mount.

import { MotionConfig } from "motion/react";

export function MotionConfigProvider({ children }: { children: React.ReactNode }) {
  return <MotionConfig reducedMotion="user">{children}</MotionConfig>;
}
