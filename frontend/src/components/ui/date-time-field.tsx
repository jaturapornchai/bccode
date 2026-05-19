import { CalendarDays, Clock3 } from "lucide-react";
import * as React from "react";
import { formatLocalDate, type CalendarYearType } from "@/lib/date-time";
import type { LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";

type SharedFieldProps = {
  className?: string;
  helper?: string;
  label?: string;
  timezoneLabel?: string;
};

type DateFieldProps = Omit<React.InputHTMLAttributes<HTMLInputElement>, "type"> & SharedFieldProps & {
  language: LanguageCode;
  yearType: CalendarYearType;
};

type TimeFieldProps = Omit<React.InputHTMLAttributes<HTMLInputElement>, "type"> & SharedFieldProps & {
  utcPreview?: string;
};

const dateTimeFieldBase =
  "group grid min-w-0 gap-1 rounded-lg border border-input bg-background px-2 py-1 shadow-[var(--shadow-card)] transition-colors focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/25";
const inputBase = "min-w-0 border-0 bg-transparent p-0 text-sm font-semibold text-foreground outline-none [color-scheme:inherit]";

export const DateField = React.forwardRef<HTMLInputElement, DateFieldProps>(
  ({ className, helper, label, language, timezoneLabel, value, yearType, ...props }, ref) => {
    const displayValue = typeof value === "string" ? formatLocalDate(value, language, yearType) : "";
    return (
      <label className={cn(dateTimeFieldBase, className)}>
        <span className="flex min-w-0 items-center gap-2 text-[0.7rem] font-bold text-muted-foreground">
          <CalendarDays className="size-3.5 shrink-0 text-primary" />
          <span className="truncate">{label}</span>
          {timezoneLabel ? <span className="ml-auto truncate">{timezoneLabel}</span> : null}
        </span>
        <input ref={ref} type="date" className={inputBase} value={value} {...props} />
        {displayValue || helper ? <small className="truncate text-[0.68rem] font-semibold text-muted-foreground">{displayValue || helper}</small> : null}
      </label>
    );
  },
);
DateField.displayName = "DateField";

export const TimeField = React.forwardRef<HTMLInputElement, TimeFieldProps>(
  ({ className, helper, label, timezoneLabel, utcPreview, value, ...props }, ref) => (
    <label className={cn(dateTimeFieldBase, className)}>
      <span className="flex min-w-0 items-center gap-2 text-[0.7rem] font-bold text-muted-foreground">
        <Clock3 className="size-3.5 shrink-0 text-primary" />
        <span className="truncate">{label}</span>
        {timezoneLabel ? <span className="ml-auto truncate">{timezoneLabel}</span> : null}
      </span>
      <input ref={ref} type="time" className={inputBase} value={value} {...props} />
      {utcPreview || helper ? <small className="truncate text-[0.68rem] font-semibold text-muted-foreground">{utcPreview || helper}</small> : null}
    </label>
  ),
);
TimeField.displayName = "TimeField";
