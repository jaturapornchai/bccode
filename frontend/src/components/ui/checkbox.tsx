"use client";

import React, { forwardRef } from "react";
import { Check as CheckIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export interface CheckboxProps extends React.InputHTMLAttributes<HTMLInputElement> {
  onCheckedChange?: (checked: boolean) => void;
  indicatorClassName?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, indicatorClassName, checked, disabled, onChange, onCheckedChange, id, ...props }, ref) => {
    return (
      <span className="relative inline-flex shrink-0 items-center justify-center">
        <input
          type="checkbox"
          ref={ref}
          id={id}
          checked={checked}
          disabled={disabled}
          onChange={(e) => {
            onChange?.(e);
            onCheckedChange?.(e.target.checked);
          }}
          className="peer sr-only"
          {...props}
        />
        <span
          aria-hidden="true"
          className={cn(
            "flex size-5 shrink-0 items-center justify-center rounded-md border-2 transition-all duration-150 cursor-pointer select-none",
            "peer-focus-visible:ring-2 peer-focus-visible:ring-ring peer-focus-visible:ring-offset-1 peer-focus-visible:outline-none",
            checked
              ? "border-primary bg-primary text-primary-foreground shadow-[0_1px_3px_rgba(0,0,0,0.15)]"
              : "border-muted-foreground/40 bg-background hover:border-primary/60 shadow-inner",
            disabled && "cursor-default opacity-50 bg-muted/40 border-border pointer-events-none shadow-none",
            className
          )}
        >
          {checked && <CheckIcon className={cn("size-3.5 stroke-[3.2]", indicatorClassName)} />}
        </span>
      </span>
    );
  }
);
Checkbox.displayName = "Checkbox";

export interface CheckboxCardProps extends CheckboxProps {
  label: React.ReactNode;
  hint?: React.ReactNode;
  cardClassName?: string;
}

export function CheckboxCard({
  label,
  hint,
  checked,
  disabled,
  cardClassName,
  className,
  onChange,
  onCheckedChange,
  id,
  ...props
}: CheckboxCardProps) {
  return (
    <label
      htmlFor={id}
      className={cn(
        "group inline-flex min-h-11 items-center gap-3 rounded-xl border px-3.5 py-2 text-[0.95rem] font-medium leading-normal cursor-pointer select-none transition-all duration-150",
        "shadow-[0_2px_8px_rgba(0,0,0,0.06),0_1px_2px_rgba(0,0,0,0.04)] hover:shadow-[0_3px_12px_rgba(0,0,0,0.1)]",
        "focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-1 focus-within:outline-none",
        checked
          ? "border-primary/50 bg-primary/10 text-foreground hover:border-primary/70 hover:bg-primary/[0.14]"
          : "border-input/90 bg-background text-foreground/90 hover:border-border hover:bg-muted/40",
        disabled && "cursor-default opacity-60 pointer-events-none shadow-none bg-muted/20 border-border text-muted-foreground",
        cardClassName
      )}
    >
      <Checkbox
        id={id}
        checked={checked}
        disabled={disabled}
        onChange={onChange}
        onCheckedChange={onCheckedChange}
        className={className}
        {...props}
      />
      <span className="flex flex-col min-w-0 flex-1">
        <span className="tracking-tight text-foreground">{label}</span>
        {hint && <span className="text-xs text-muted-foreground font-normal mt-0.5">{hint}</span>}
      </span>
    </label>
  );
}
