"use client";

import * as React from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

export interface SelectOption {
  value: string | number;
  label: string;
}

export interface CustomSelectProps {
  value: string | number;
  onChange: (value: any) => void;
  options: SelectOption[];
  className?: string;
  disabled?: boolean;
}

export const CustomSelect = React.forwardRef<HTMLButtonElement, CustomSelectProps>(
  ({ value, onChange, options, className, disabled }, ref) => {
    const selectedOption = options.find((opt) => opt.value === value);

    return (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            ref={ref}
            variant="outline"
            role="combobox"
            disabled={disabled}
            className={cn(
              "w-full justify-between h-9 px-3 text-left font-normal border border-input bg-transparent hover:bg-accent/50 text-sm shadow-sm transition-colors focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
              className
            )}
          >
            <span className="truncate">{selectedOption ? selectedOption.label : "Select option..."}</span>
            <ChevronDown className="h-4 w-4 shrink-0 opacity-50 ml-2" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-[var(--radix-dropdown-menu-trigger-width)] min-w-[12rem] bg-popover text-popover-foreground border border-border rounded-lg shadow-md p-1">
          {options.map((option) => (
            <DropdownMenuItem
              key={option.value}
              onClick={() => onChange(option.value)}
              className={cn(
                "cursor-pointer hover:bg-accent hover:text-accent-foreground px-2 py-1.5 text-sm rounded-md",
                value === option.value && "bg-accent text-accent-foreground font-semibold"
              )}
            >
              {option.label}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }
);

CustomSelect.displayName = "CustomSelect";
