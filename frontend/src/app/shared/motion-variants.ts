// Shared motion variants for premium card stagger across screens.
// Reuse in holding/workspace/menu/login to keep animation feel consistent.
// Reduced-motion is handled globally via <MotionConfig reducedMotion="user">.

import type { Variants } from "motion/react";

export const cardStaggerParent: Variants = {
  animate: { transition: { staggerChildren: 0.06, delayChildren: 0.08 } },
};

export const cardStaggerChild: Variants = {
  initial: { opacity: 0, y: 12 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease: [0.22, 1, 0.36, 1] },
  },
};

export const panelFadeIn: Variants = {
  initial: { opacity: 0, y: 16 },
  animate: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.45, ease: [0.22, 1, 0.36, 1] },
  },
};

export const noticeSlide: Variants = {
  initial: { opacity: 0, y: -6, height: 0 },
  animate: {
    opacity: 1,
    y: 0,
    height: "auto",
    transition: { duration: 0.24, ease: [0.22, 1, 0.36, 1] },
  },
  exit: {
    opacity: 0,
    y: -6,
    height: 0,
    transition: { duration: 0.2, ease: [0.22, 1, 0.36, 1] },
  },
};
