import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => (
    <input
      type={type}
      className={cn(
        "flex h-8 w-full rounded-lg border border-input bg-transparent px-1.5 py-0.5 text-sm text-foreground shadow-[var(--shadow-card)] transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 [&:-webkit-autofill]:[box-shadow:0_0_0_1000px_transparent_inset_!important] [&:-webkit-autofill]:[text-fill-color:inherit_!important] [&:-webkit-autofill]:transition-[background-color_5000s_ease-in-out_0s]",
        className,
      )}
      ref={ref}
      {...props}
    />
  ),
);
Input.displayName = "Input";

export { Input };
