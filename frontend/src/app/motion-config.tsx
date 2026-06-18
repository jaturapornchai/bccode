"use client";

// Use this inside client-only mounted subtrees so Motion can respect the user's
// prefers-reduced-motion setting without changing server-rendered styles during
// hydration. Do not wrap the root layout with this for entry animations.

import { MotionConfig } from "motion/react";

export function MotionConfigProvider({ children }: { children: React.ReactNode }) {
  return <MotionConfig reducedMotion="user">{children}</MotionConfig>;
}
