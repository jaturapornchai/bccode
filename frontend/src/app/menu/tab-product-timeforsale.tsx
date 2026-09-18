"use client";

import { useCallback } from "react";
import { useBarcodeText } from "@/components/product-barcode/use-barcode-text";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import { type Product, type ProductTimeForSale } from "@/lib/product-barcode/types";
import { FieldRow, Section, type ProductStateAction } from "./product-tab-shared";

export function TabProductTimeForSale({
  value,
  onChange,
  language = "th",
}: {
  value: Product;
  onChange: ProductStateAction;
  language?: string;
}) {
  const textT = useBarcodeText(language);
  const setRows = useCallback(
    (mutator: (rows: ProductTimeForSale[]) => ProductTimeForSale[]) =>
      onChange((c) => c ? ({ ...c, timeforsales: mutator(c.timeforsales || []) } as Product) : null),
    [onChange],
  );

  const DAYS = [textT.sun, textT.mon, textT.tue, textT.wed, textT.thu, textT.fri, textT.sat];

  return (
    <Section
      title={textT.timeSection}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [
              ...rows,
              { daysofweek: [1, 2, 3, 4, 5], fromdate: "", todate: "", fromtime: "00:00", totime: "23:59" },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {textT.timeAddBtn}
        </Button>
      }
    >
      {(!value.timeforsales || value.timeforsales.length === 0) ? (
        <p className="text-sm text-muted-foreground">{textT.timeNoLimit}</p>
      ) : (
        <div className="space-y-4">
          {value.timeforsales.map((entry, idx) => (
            <div key={idx} className="p-3 border border-border rounded bg-muted/10 space-y-3">
              <div className="flex items-center justify-between border-b border-border/60 pb-1">
                <span className="text-xs font-semibold text-muted-foreground">ข้อกำหนดเวลาขายที่ #{idx + 1}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 text-destructive hover:bg-destructive/10"
                  onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>

              {/* Days of week checklist */}
              <div className="space-y-1">
                <label className="text-xs text-muted-foreground font-semibold">{textT.timeDaysLabel}</label>
                <div className="flex flex-wrap gap-2">
                  {DAYS.map((dayLabel, dayIdx) => {
                    const checked = (entry.daysofweek || []).includes(dayIdx as any);
                    return (
                      <label
                        key={dayIdx}
                        className={cn(
                          "flex items-center gap-2 text-xs font-medium cursor-pointer px-2.5 py-1.5 rounded-lg border transition-all select-none shadow-[0_1px_3px_rgba(0,0,0,0.04)]",
                          checked
                            ? "bg-primary/10 border-primary/40 text-foreground"
                            : "bg-background border-border hover:bg-muted/40 text-muted-foreground hover:text-foreground"
                        )}
                      >
                        <Checkbox
                          checked={checked}
                          onCheckedChange={(isNextChecked) => {
                            const activeDays = entry.daysofweek || [];
                            const nextDays = isNextChecked
                              ? [...activeDays, dayIdx as any]
                              : activeDays.filter((d) => d !== dayIdx);
                            setRows((rows) =>
                              rows.map((row, rowIdx) =>
                                rowIdx === idx ? { ...row, daysofweek: nextDays } : row,
                              ),
                            );
                          }}
                        />
                        <span>{dayLabel}</span>
                      </label>
                    );
                  })}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-4">
                <FieldRow label={textT.timeFromDate}>
                  <Input
                    type="date"
                    value={entry.fromdate || ""}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, fromdate: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textT.timeToDate}>
                  <Input
                    type="date"
                    value={entry.todate || ""}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, todate: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textT.timeFromTime}>
                  <Input
                    type="time"
                    value={entry.fromtime || "00:00"}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, fromtime: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
                <FieldRow label={textT.timeToTime}>
                  <Input
                    type="time"
                    value={entry.totime || "23:59"}
                    onChange={(e) =>
                      setRows((rows) =>
                        rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, totime: e.target.value } : row)),
                      )
                    }
                  />
                </FieldRow>
              </div>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}
