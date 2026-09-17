"use client";

import * as React from "react";
import { Check, ChevronDown, Search, X } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ComboboxOption<T = string | number> {
  value: T;
  label: React.ReactNode;
  disabled?: boolean;
  hint?: React.ReactNode;
}

export interface ComboboxProps<T = string | number> {
  value?: T;
  defaultValue?: T;
  onChange?: (value: T) => void;
  options?: ComboboxOption<T>[];
  children?: React.ReactNode;
  placeholder?: string;
  searchable?: boolean;
  searchPlaceholder?: string;
  disabled?: boolean;
  readOnly?: boolean;
  className?: string;
  buttonClassName?: string;
  wrapperClassName?: string;
  dropdownClassName?: string;
  id?: string;
  name?: string;
  "aria-label"?: string;
  "aria-labelledby"?: string;
  "data-field"?: string;
  tabIndex?: number;
}

/**
 * Extracts options from children <option> elements if options prop is not provided.
 */
function parseOptionsFromChildren<T = string | number>(children: React.ReactNode): ComboboxOption<T>[] {
  const result: ComboboxOption<T>[] = [];
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
 * Global Combobox component designed for Thai 40+ UX/UI and premium BC Ai design.
 * Features:
 * - 44px+ touch-friendly target (`min-h-[2.6em]`)
 * - Rich depth elevation shadow with sharp borders
 * - Accessible WAI-ARIA combobox pattern with keyboard navigation
 * - Built-in search filtering for large option lists
 * - Drop-in replacement supporting both `options` prop and children `<option>` tags
 */
function ComboboxInner<T extends string | number = string | number>(
  {
    value: controlledValue,
    defaultValue,
    onChange,
      options: optionsProp,
      children,
      placeholder = "เลือก...",
      searchable,
      searchPlaceholder = "ค้นหา...",
      disabled = false,
      readOnly = false,
      className,
      buttonClassName,
      wrapperClassName,
      dropdownClassName,
      id,
      name,
      "aria-label": ariaLabel,
      "aria-labelledby": ariaLabelledBy,
      "data-field": dataField,
      tabIndex,
    }: ComboboxProps<T>,
    ref: React.ForwardedRef<HTMLButtonElement>
  ) {
    const wrapperLayoutClasses = React.useMemo(() => {
      if (!className) return "";
      return className
        .split(/\s+/)
        .filter((c) => /^(w-|max-w-|min-w-|col-|row-|shrink|grow|basis-|self-)/.test(c))
        .join(" ");
    }, [className]);
    const isControlled = controlledValue !== undefined;
    const [internalValue, setInternalValue] = React.useState<T | undefined>(defaultValue);
    const currentValue = isControlled ? controlledValue : internalValue;

    const [isOpen, setIsOpen] = React.useState(false);
    const [searchQuery, setSearchQuery] = React.useState("");
    const [highlightedIndex, setHighlightedIndex] = React.useState<number>(-1);

    const containerRef = React.useRef<HTMLDivElement>(null);
    const internalTriggerRef = React.useRef<HTMLButtonElement>(null);
    const searchInputRef = React.useRef<HTMLInputElement>(null);
    const listboxRef = React.useRef<HTMLUListElement>(null);

    // Merge forwarded ref and internal ref
    React.useImperativeHandle(ref, () => internalTriggerRef.current as HTMLButtonElement);

    const generatedId = React.useId();
    const componentId = id || `combobox-${generatedId}`;
    const listboxId = `${componentId}-listbox`;

    // Parse options from prop or children
    const rawOptions = React.useMemo(() => {
      if (optionsProp && optionsProp.length > 0) {
        return optionsProp;
      }
      if (children) {
        return parseOptionsFromChildren<T>(children);
      }
      return [];
    }, [optionsProp, children]);

    // Determine if search should be enabled (explicit true, or auto if > 6 options and not explicitly false)
    const isSearchable = searchable ?? rawOptions.length > 6;

    // Filter options by search query
    const filteredOptions = React.useMemo(() => {
      if (!searchQuery.trim()) return rawOptions;
      const query = searchQuery.trim().toLowerCase();
      return rawOptions.filter((opt) => {
        const labelText = typeof opt.label === "string" ? opt.label : String(opt.label ?? opt.value ?? "");
        return (
          labelText.toLowerCase().includes(query) ||
          String(opt.value ?? "").toLowerCase().includes(query)
        );
      });
    }, [rawOptions, searchQuery]);

    // Find current selected option
    const selectedOption = React.useMemo(() => {
      return rawOptions.find((opt) => String(opt.value) === String(currentValue));
    }, [rawOptions, currentValue]);

    const handleSelect = React.useCallback(
      (val: T) => {
        if (disabled || readOnly) return;
        if (!isControlled) {
          setInternalValue(val);
        }
        onChange?.(val);
        setIsOpen(false);
        setSearchQuery("");
        internalTriggerRef.current?.focus();
      },
      [disabled, readOnly, isControlled, onChange]
    );

    // Close on outside click
    React.useEffect(() => {
      if (!isOpen) return;
      const handleClickOutside = (event: MouseEvent | TouchEvent) => {
        if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
          setIsOpen(false);
          setSearchQuery("");
        }
      };
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("touchstart", handleClickOutside);
      return () => {
        document.removeEventListener("mousedown", handleClickOutside);
        document.removeEventListener("touchstart", handleClickOutside);
      };
    }, [isOpen]);

    // Focus search input when menu opens
    React.useEffect(() => {
      if (isOpen && isSearchable) {
        const timer = setTimeout(() => {
          searchInputRef.current?.focus();
        }, 30);
        return () => clearTimeout(timer);
      }
    }, [isOpen, isSearchable]);

    // Reset highlighted index when menu opens or search changes
    React.useEffect(() => {
      if (!isOpen) {
        setHighlightedIndex(-1);
        return;
      }
      const selectedIndex = filteredOptions.findIndex(
        (opt) => String(opt.value) === String(currentValue)
      );
      setHighlightedIndex(selectedIndex >= 0 ? selectedIndex : 0);
    }, [isOpen, filteredOptions, currentValue]);

    // Scroll active item into view
    React.useEffect(() => {
      if (!isOpen || highlightedIndex < 0 || !listboxRef.current) return;
      const listElement = listboxRef.current;
      const activeElement = listElement.children[highlightedIndex] as HTMLElement | undefined;
      if (activeElement) {
        const itemTop = activeElement.offsetTop;
        const itemBottom = itemTop + activeElement.offsetHeight;
        const listTop = listElement.scrollTop;
        const listBottom = listTop + listElement.offsetHeight;

        if (itemTop < listTop) {
          listElement.scrollTop = itemTop;
        } else if (itemBottom > listBottom) {
          listElement.scrollTop = itemBottom - listElement.offsetHeight;
        }
      }
    }, [highlightedIndex, isOpen]);

    // Keyboard navigation handler
    const handleKeyDown = (event: React.KeyboardEvent) => {
      if (disabled || readOnly) return;

      if (!isOpen) {
        if (event.key === "ArrowDown" || event.key === "ArrowUp" || event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          setIsOpen(true);
        }
        return;
      }

      switch (event.key) {
        case "ArrowDown": {
          event.preventDefault();
          if (filteredOptions.length === 0) return;
          let nextIndex = highlightedIndex + 1;
          while (nextIndex < filteredOptions.length && filteredOptions[nextIndex].disabled) {
            nextIndex++;
          }
          if (nextIndex < filteredOptions.length) {
            setHighlightedIndex(nextIndex);
          }
          break;
        }
        case "ArrowUp": {
          event.preventDefault();
          if (filteredOptions.length === 0) return;
          let prevIndex = highlightedIndex - 1;
          while (prevIndex >= 0 && filteredOptions[prevIndex].disabled) {
            prevIndex--;
          }
          if (prevIndex >= 0) {
            setHighlightedIndex(prevIndex);
          }
          break;
        }
        case "Enter": {
          event.preventDefault();
          if (highlightedIndex >= 0 && highlightedIndex < filteredOptions.length) {
            const opt = filteredOptions[highlightedIndex];
            if (!opt.disabled) {
              handleSelect(opt.value);
            }
          }
          break;
        }
        case "Escape": {
          event.preventDefault();
          setIsOpen(false);
          setSearchQuery("");
          internalTriggerRef.current?.focus();
          break;
        }
        case "Tab": {
          setIsOpen(false);
          setSearchQuery("");
          break;
        }
        case "Home": {
          event.preventDefault();
          const firstValid = filteredOptions.findIndex((opt) => !opt.disabled);
          if (firstValid >= 0) setHighlightedIndex(firstValid);
          break;
        }
        case "End": {
          event.preventDefault();
          for (let i = filteredOptions.length - 1; i >= 0; i--) {
            if (!filteredOptions[i].disabled) {
              setHighlightedIndex(i);
              break;
            }
          }
          break;
        }
      }
    };

    return (
      <div
        ref={containerRef}
        className={cn("relative w-full min-w-0 inline-block text-left", wrapperLayoutClasses, wrapperClassName)}
        onKeyDown={handleKeyDown}
      >
        {/* Hidden input for HTML form integration if name is passed */}
        {name && (
          <input
            type="hidden"
            name={name}
            value={currentValue !== undefined ? String(currentValue) : ""}
            disabled={disabled}
          />
        )}

        {/* Trigger Button */}
        <button
          ref={internalTriggerRef}
          type="button"
          id={componentId}
          data-field={dataField}
          role="combobox"
          aria-expanded={isOpen}
          aria-haspopup="listbox"
          aria-controls={listboxId}
          aria-label={ariaLabel}
          aria-labelledby={ariaLabelledBy}
          disabled={disabled}
          tabIndex={tabIndex ?? 0}
          onClick={() => {
            if (!disabled && !readOnly) {
              setIsOpen((prev) => !prev);
            }
          }}
          className={cn(
            "group flex min-h-[2.6em] w-full items-center justify-between gap-2 rounded-xl border border-input bg-background px-3 py-1.5 text-[0.95rem] leading-normal text-foreground select-none text-left",
            // Depth Shadow elevation
            "shadow-[0_3px_10px_rgba(0,0,0,0.14),0_1px_3px_rgba(0,0,0,0.1)] dark:shadow-[0_3px_10px_rgba(0,0,0,0.6)] transition-[border-color,box-shadow,background-color] duration-150",
            // Hover state
            "hover:border-primary/80 hover:shadow-[0_4px_16px_rgba(0,0,0,0.18),0_1px_4px_rgba(0,0,0,0.12)]",
            // Focus state
            "focus-visible:outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring focus-visible:shadow-[0_4px_16px_rgba(0,0,0,0.2)]",
            // Open state
            isOpen && "border-primary ring-2 ring-ring/80 shadow-[0_4px_16px_rgba(0,0,0,0.2)]",
            // Disabled state - preserve depth shadow for Thai 40+ read-only clarity
            disabled &&
              "cursor-default bg-muted/20 border-border text-foreground shadow-[0_2px_8px_rgba(0,0,0,0.1),0_1px_2px_rgba(0,0,0,0.07)] hover:border-border hover:shadow-[0_2px_8px_rgba(0,0,0,0.1),0_1px_2px_rgba(0,0,0,0.07)] pointer-events-none opacity-90",
            className,
            buttonClassName
          )}
        >
          <span className="min-w-0 flex-1 truncate">
            {selectedOption ? (
              <span className="font-medium text-foreground">{selectedOption.label}</span>
            ) : (
              <span className="text-muted-foreground">{placeholder}</span>
            )}
          </span>

          <ChevronDown
            className={cn(
              "size-4.5 shrink-0 text-muted-foreground transition-transform duration-200",
              isOpen && "rotate-180 text-primary",
              disabled && "opacity-40"
            )}
            aria-hidden="true"
          />
        </button>

        {/* Dropdown Floating Panel */}
        {isOpen && !disabled && (
          <div
            className={cn(
              "absolute left-0 top-full z-50 mt-1.5 w-full min-w-[14rem] overflow-hidden rounded-xl border border-border/80 bg-popover text-popover-foreground shadow-[0_12px_36px_rgba(0,0,0,0.2),0_4px_12px_rgba(0,0,0,0.1)] dark:shadow-[0_12px_36px_rgba(0,0,0,0.8)] backdrop-blur-md animate-in fade-in-50 zoom-in-95 duration-100",
              dropdownClassName
            )}
          >
            {/* Search Input Filter */}
            {isSearchable && (
              <div className="flex items-center gap-2 border-b border-border/70 bg-muted/30 px-3 py-2">
                <Search className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                <input
                  ref={searchInputRef}
                  type="text"
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setHighlightedIndex(0);
                  }}
                  placeholder={searchPlaceholder}
                  className="min-h-[2em] w-full bg-transparent text-[0.95rem] placeholder:text-muted-foreground focus:outline-none"
                  aria-label={searchPlaceholder}
                />
                {searchQuery && (
                  <button
                    type="button"
                    onClick={() => {
                      setSearchQuery("");
                      searchInputRef.current?.focus();
                    }}
                    className="inline-flex size-5 shrink-0 items-center justify-center rounded-full text-muted-foreground hover:bg-muted hover:text-foreground"
                    aria-label="ล้างคำค้นหา"
                  >
                    <X className="size-3.5" />
                  </button>
                )}
              </div>
            )}

            {/* Options Listbox */}
            <ul
              ref={listboxRef}
              id={listboxId}
              role="listbox"
              aria-label={ariaLabel || "ตัวเลือก"}
              className="max-h-60 overflow-y-auto p-1.5 focus:outline-none"
              tabIndex={-1}
            >
              {filteredOptions.length === 0 ? (
                <li className="px-3 py-3 text-center text-[0.9rem] text-muted-foreground">
                  ไม่พบข้อมูลที่ค้นหา
                </li>
              ) : (
                filteredOptions.map((option, index) => {
                  const isSelected = String(option.value) === String(currentValue);
                  const isHighlighted = index === highlightedIndex;

                  return (
                    <li
                      key={String(option.value)}
                      id={`${componentId}-opt-${index}`}
                      role="option"
                      aria-selected={isSelected}
                      aria-disabled={option.disabled}
                      onClick={() => {
                        if (!option.disabled) {
                          handleSelect(option.value);
                        }
                      }}
                      onMouseEnter={() => {
                        if (!option.disabled) {
                          setHighlightedIndex(index);
                        }
                      }}
                      className={cn(
                        "group relative flex min-h-[2.4em] cursor-pointer items-center justify-between gap-2 rounded-lg px-3 py-2 text-[0.95rem] transition-colors select-none",
                        isHighlighted && !isSelected && "bg-accent text-accent-foreground",
                        isSelected && "bg-primary/10 text-primary font-semibold hover:bg-primary/15",
                        option.disabled && "cursor-not-allowed opacity-40 pointer-events-none"
                      )}
                    >
                      <div className="flex min-w-0 flex-1 items-center gap-2">
                        <span className="truncate leading-normal">{option.label}</span>
                        {option.hint && (
                          <span className="text-xs text-muted-foreground shrink-0">{option.hint}</span>
                        )}
                      </div>

                      {isSelected && (
                        <Check className="size-4 shrink-0 text-primary" aria-hidden="true" />
                      )}
                    </li>
                  );
                })
              )}
            </ul>
          </div>
        )}
      </div>
    );
  }

export const Combobox = React.forwardRef(ComboboxInner) as <
  T extends string | number = string | number
>(
  props: ComboboxProps<T> & React.RefAttributes<HTMLButtonElement>
) => React.ReactElement | null;

(Combobox as unknown as { displayName: string }).displayName = "Combobox";
