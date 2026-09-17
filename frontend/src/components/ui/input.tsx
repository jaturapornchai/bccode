import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => (
    <input
      type={type}
      className={cn(
        "flex h-8 w-full rounded-lg border border-input bg-background px-2 py-0 text-sm text-foreground shadow-[0_3px_10px_rgba(0,0,0,0.14),0_1px_3px_rgba(0,0,0,0.1)] dark:shadow-[0_3px_10px_rgba(0,0,0,0.6)] transition-[border-color,box-shadow] placeholder:text-muted-foreground hover:border-primary/80 hover:shadow-[0_4px_16px_rgba(0,0,0,0.18),0_1px_4px_rgba(0,0,0,0.12)] focus-visible:outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring focus-visible:shadow-[0_4px_16px_rgba(0,0,0,0.2)] disabled:cursor-not-allowed disabled:bg-muted/20 disabled:border-border disabled:shadow-[0_2px_8px_rgba(0,0,0,0.1),0_1px_2px_rgba(0,0,0,0.07)] [&:-webkit-autofill]:[box-shadow:0_0_0_1000px_transparent_inset_!important] [&:-webkit-autofill]:[text-fill-color:inherit_!important] [&:-webkit-autofill]:transition-[background-color_5000s_ease-in-out_0s]",
        className,
      )}
      ref={ref}
      {...props}
    />
  ),
);
Input.displayName = "Input";

export { Input };
