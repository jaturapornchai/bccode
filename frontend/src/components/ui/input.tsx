import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => (
    <input
      type={type}
      className={cn(
        "flex h-8 w-full rounded-lg border border-input bg-transparent px-1 py-0 text-sm text-foreground shadow-xs transition-[border-color,box-shadow] placeholder:text-muted-foreground hover:border-primary/40 focus-visible:outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring focus-visible:shadow-sm disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none [&:-webkit-autofill]:[box-shadow:0_0_0_1000px_transparent_inset_!important] [&:-webkit-autofill]:[text-fill-color:inherit_!important] [&:-webkit-autofill]:transition-[background-color_5000s_ease-in-out_0s]",
        className,
      )}
      ref={ref}
      {...props}
    />
  ),
);
Input.displayName = "Input";

export { Input };
