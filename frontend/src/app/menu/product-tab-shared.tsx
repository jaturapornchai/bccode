"use client";

/**
 * Shared field/layout helpers for Product screen detail tabs.
 * Extracted from product-screen.tsx (mechanical move, no behavior change) so each
 * tab can live in its own file per the existing tab-product-units.tsx precedent.
 */

import { type ChangeEvent, type Dispatch, type ReactNode, type SetStateAction } from "react";
import { Input } from "@/components/ui/input";
import { NumericInput } from "@/components/ui/numeric-input";
import { cn } from "@/lib/utils";
import { Checkbox } from "@/components/ui/checkbox";
import type { Product } from "@/lib/product-barcode/types";

export type ProductStateAction = Dispatch<SetStateAction<Product | null>>;

export function FieldRow({ label, hint, children, required }: { label: string; hint?: string; children: ReactNode; required?: boolean }) {
  return (
    <label className="flex flex-col gap-1 text-sm">
      <span className="font-medium text-foreground">
        {label}
        {required ? <span className="ml-1 text-destructive">*</span> : null}
      </span>
      {children}
      {hint ? <span className="text-xs text-muted-foreground">{hint}</span> : null}
    </label>
  );
}

export function SectionCard({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <section className="rounded-lg border border-border bg-card text-card-foreground">
      <header className="border-b border-border/60 p-3">
        <h3 className="text-sm font-semibold">{title}</h3>
        {description ? <p className="text-xs text-muted-foreground">{description}</p> : null}
      </header>
      <div className="p-3">{children}</div>
    </section>
  );
}

export function FieldGrid({ children }: { children: ReactNode }) {
  return <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{children}</div>;
}

export function Section({ title, action, children }: { title: string; action?: ReactNode; children: ReactNode }) {
  return (
    <section className="mb-4 overflow-hidden rounded-lg border border-border bg-card">
      <header className="flex items-center justify-between border-b border-border px-3 py-2 bg-muted/20">
        <h3 className="text-sm font-semibold">{title}</h3>
        {action}
      </header>
      <div className="p-3">{children}</div>
    </section>
  );
}

export function Toggle({
  checked,
  onCheckedChange,
  label,
  disabled,
}: {
  checked: boolean;
  onCheckedChange: (next: boolean) => void;
  label: string;
  disabled?: boolean;
}) {
  return (
    <label
      className={cn(
        "inline-flex min-h-10 cursor-pointer items-center gap-2.5 rounded-lg border border-input/80 bg-background/90 px-3 py-1.5 text-sm font-medium text-foreground",
        "shadow-[0_1px_4px_rgba(0,0,0,0.06)] hover:border-primary/50 hover:bg-accent/40 transition-all select-none",
        checked && "border-primary/40 bg-primary/5",
        disabled && "cursor-not-allowed opacity-60 pointer-events-none"
      )}
    >
      <Checkbox
        checked={checked}
        disabled={disabled}
        onCheckedChange={onCheckedChange}
      />
      <span>{label}</span>
    </label>
  );
}

export type RadioOptionValue = string | number | boolean;

export function RadioOptionGroup<T extends RadioOptionValue>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: Array<{ value: T; label: string; disabled?: boolean }>;
  onChange: (next: T) => void;
}) {
  return (
    <fieldset className="rounded-md border border-border bg-background px-2.5 py-2">
      <legend className="px-1 text-xs font-semibold text-muted-foreground">{label}</legend>
      <div className="grid gap-1.5 sm:grid-cols-2 lg:grid-cols-3">
        {options.map((option) => (
          <label
            key={String(option.value)}
            className={cn(
              "flex min-h-7 w-full cursor-pointer items-center gap-2 rounded-md px-2 py-1 text-xs font-medium hover:bg-muted/60",
              option.disabled && "cursor-not-allowed opacity-60",
            )}
          >
            <input
              type="radio"
              checked={Object.is(value, option.value)}
              disabled={option.disabled}
              onChange={() => onChange(option.value)}
              className="size-4"
            />
            <span className="whitespace-normal break-words">{option.label}</span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}

export function NumberField({
  value,
  onChange,
  min,
  step = "any",
  className,
}: {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  step?: string | number;
  className?: string;
}) {
  return (
    <NumericInput value={value} onChange={onChange} min={min} step={step} className={className} />
  );
}
