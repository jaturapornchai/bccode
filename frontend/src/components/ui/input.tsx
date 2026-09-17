import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => (
    <input
      type={type}
      className={cn(
        "flex h-8 w-full rounded-lg border border-input bg-background px-2 py-0 text-sm text-foreground shadow-[0_2px_6px_rgba(0,0,0,0.08),0_1px_2px_rgba(0,0,0,0.06)] dark:shadow-[0_2px_6px_rgba(0,0,0,0.45)] transition-[border-color,box-shadow] placeholder:text-muted-foreground hover:border-primary/60 hover:shadow-[0_3px_8px_rgba(0,0,0,0.12),0_1px_3px_rgba(0,0,0,0.08)] focus-visible:outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring focus-visible:shadow-[0_2px_8px_rgba(0,0,0,0.12)] disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none [&:-webkit-autofill]:[box-shadow:0_0_0_1000px_transparent_inset_!important] [&:-webkit-autofill]:[text-fill-color:inherit_!important] [&:-webkit-autofill]:transition-[background-color_5000s_ease-in-out_0s]",
        className,
      )}
      ref={ref}
      {...props}
    />
  ),
);
Input.displayName = "Input";

export { Input };
