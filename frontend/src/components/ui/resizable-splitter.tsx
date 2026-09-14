"use client";

import React, { forwardRef } from "react";
import { GripVertical } from "lucide-react";
import { cn } from "@/lib/utils";

export { useSplitPercent } from "./use-split-percent";
export type { UseSplitPercentOptions, UseSplitPercentResult } from "./use-split-percent";

export interface ResizableSplitterProps extends React.HTMLAttributes<HTMLDivElement> {
  isResizing?: boolean;
  min?: number;
  max?: number;
  value?: number;
  label?: string;
  pillSize?: "sm" | "md";
  breakpoint?: "md" | "lg" | "xl";
}

export const ResizableSplitter = forwardRef<HTMLDivElement, ResizableSplitterProps>(
  (
    {
      isResizing = false,
      min,
      max,
      value,
      label,
      title,
      tabIndex = 0,
      onPointerDown,
      onDoubleClick,
      onKeyDown,
      pillSize = "sm",
      breakpoint = "lg",
      className,
      children,
      ...props
    },
    ref,
  ) => {
    const visibilityClass =
      breakpoint === "xl"
        ? "hidden xl:flex"
        : breakpoint === "md"
          ? "hidden md:flex"
          : "hidden lg:flex";

    return (
      <div
        ref={ref}
        role="separator"
        aria-orientation="vertical"
        aria-valuenow={value}
        aria-valuemin={min}
        aria-valuemax={max}
        aria-label={label}
        title={title || label}
        tabIndex={tabIndex}
        onPointerDown={onPointerDown}
        onDoubleClick={onDoubleClick}
        onKeyDown={onKeyDown}
        className={cn(
          "group relative z-20 w-3 shrink-0 cursor-col-resize select-none touch-none items-center justify-center transition-colors -mx-1",
          visibilityClass,
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-1",
          isResizing && "bg-primary/10",
          className,
        )}
        {...props}
      >
        {/* Track Line */}
        <div
          className={cn(
            "h-full w-0.5 transition-colors duration-150 rounded-full",
            "group-hover:bg-primary/50",
            isResizing ? "bg-primary" : "bg-border/60",
          )}
        />

        {/* Floating Pill Handle with Grip Icon */}
        <div
          className={cn(
            "absolute z-10 flex items-center justify-center rounded-full border border-border/80 bg-background/95 shadow-2xs transition-all duration-150",
            pillSize === "sm" ? "h-8 w-3" : "h-10 w-3",
            pillSize === "sm"
              ? "group-hover:h-10 group-hover:border-primary/50 group-hover:bg-card group-hover:shadow"
              : "group-hover:h-12 group-hover:border-primary/50 group-hover:bg-card group-hover:shadow",
            isResizing && (pillSize === "sm" ? "h-12 border-primary bg-primary text-primary-foreground shadow-md" : "h-14 border-primary bg-primary text-primary-foreground shadow-md"),
          )}
        >
          <GripVertical
            className={cn(
              "h-3 w-3 text-muted-foreground/70 transition-colors",
              "group-hover:text-primary",
              isResizing && "text-primary-foreground",
            )}
          />
        </div>
        {children}
      </div>
    );
  },
);

ResizableSplitter.displayName = "ResizableSplitter";
