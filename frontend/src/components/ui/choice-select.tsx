"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { Combobox } from "./combobox";

export interface ChoiceOption<T = string | number> {
  value: T;
  label: React.ReactNode;
  disabled?: boolean;
  hint?: React.ReactNode;
}

export interface ChoiceSelectProps<T = string | number> {
  value?: T;
  onChange: (value: T) => void;
  options?: ChoiceOption<T>[];
  children?: React.ReactNode;
  radioThreshold?: number; // default 3: <= threshold renders as radio cards
  name?: string;
  disabled?: boolean;
  className?: string;
  radioClassName?: string;
  "aria-label"?: string;
  id?: string;
  layout?: "grid" | "flex";
}

/**
 * Extracts options from children <option> elements if options prop is not provided.
 */
function parseOptionsFromChildren<T = string | number>(children: React.ReactNode): ChoiceOption<T>[] {
  const result: ChoiceOption<T>[] = [];
  React.Children.forEach(children, (child) => {
    if (!React.isValidElement(child)) return;
    const props = child.props as Record<string, unknown> | undefined;
    if (child.type === "option" || (props && typeof props === "object" && "value" in props)) {
      const val = props?.value as T;
      const label = (props?.children as React.ReactNode) ?? String(val ?? "");
      const disabled = Boolean(props?.disabled);
      result.push({ value: val, label, disabled });
    }
  });
  return result;
}

/**
 * ChoiceSelect - Universal choice component.
 * Automatically displays as Radio Cards (min-h-[2.6em], big click target, Thai 40+ friendly)
 * when choices <= radioThreshold (default 3), or falls back to native <select> when > 3 choices.
 */
export function ChoiceSelect<T extends string | number = string>({
  value,
  onChange,
  options: optionsProp,
  children,
  radioThreshold = 3,
  name,
  disabled = false,
  className,
  radioClassName,
  "aria-label": ariaLabel,
  id,
  layout = "grid",
}: ChoiceSelectProps<T>) {
  const generatedName = React.useId();
  const radioName = name || `choice-${generatedName}`;

  const options = React.useMemo(() => {
    if (optionsProp && optionsProp.length > 0) {
      return optionsProp;
    }
    if (children) {
      return parseOptionsFromChildren<T>(children);
    }
    return [];
  }, [optionsProp, children]);

  const shouldRenderRadio = options.length > 0 && options.length <= radioThreshold;

  if (shouldRenderRadio) {
    const gridColsClass =
      options.length === 1
        ? "grid-cols-1"
        : options.length === 2
          ? "grid-cols-2"
          : "grid-cols-1 sm:grid-cols-3";

    return (
      <div
        role="radiogroup"
        aria-label={ariaLabel}
        className={cn(
          layout === "grid" ? cn("grid gap-2 w-full", gridColsClass) : "flex flex-wrap gap-2 w-full",
          className
        )}
      >
        {options.map((option) => {
          const isChecked = String(value ?? "") === String(option.value ?? "");
          const isOptionDisabled = disabled || Boolean(option.disabled);

          return (
            <label
              key={String(option.value)}
              className={cn(
                "group relative flex min-h-[2.6em] w-full cursor-pointer items-center justify-start gap-2.5 rounded-xl border px-3 py-1.5 text-[0.95rem] transition-all select-none",
                isChecked
                  ? "border-primary bg-primary/10 text-primary font-semibold ring-1 ring-primary/30 shadow-[0_3px_12px_rgba(0,0,0,0.16)]"
                  : "border-input bg-background text-foreground shadow-[0_3px_10px_rgba(0,0,0,0.14),0_1px_3px_rgba(0,0,0,0.1)] dark:shadow-[0_3px_10px_rgba(0,0,0,0.6)] hover:bg-muted/60 hover:border-primary/80 hover:shadow-[0_4px_16px_rgba(0,0,0,0.18)]",
                isOptionDisabled && "cursor-default text-foreground/90 pointer-events-none",
                radioClassName
              )}
            >
              <input
                type="radio"
                name={radioName}
                value={String(option.value)}
                checked={isChecked}
                disabled={isOptionDisabled}
                onChange={() => onChange(option.value)}
                className="sr-only"
                aria-label={typeof option.label === "string" ? option.label : undefined}
              />
              {/* Visual Radio Circle Indicator for WCAG / 40+ accessibility */}
              <span
                className={cn(
                  "flex size-4.5 shrink-0 items-center justify-center rounded-full border transition-colors",
                  isChecked
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-input bg-background group-hover:border-primary/50"
                )}
                aria-hidden="true"
              >
                {isChecked && <span className="size-2 rounded-full bg-background" />}
              </span>
              <span className="min-w-0 flex-1 truncate text-left leading-normal">{option.label}</span>
              {option.hint && <span className="text-xs text-muted-foreground ml-1">{option.hint}</span>}
            </label>
          );
        })}
      </div>
    );
  }

  // Fallback to Combobox dropdown when options > 3
  return (
    <Combobox
      id={id}
      name={name}
      aria-label={ariaLabel}
      disabled={disabled}
      value={value}
      onChange={(val) => onChange(val as T)}
      options={options}
      className={className}
    >
      {children}
    </Combobox>
  );
}
