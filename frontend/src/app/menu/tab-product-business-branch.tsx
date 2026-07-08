"use client";

import { useCallback, useState } from "react";
import { Minus, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import {
  type Product,
  type ProductBarcodeBranch,
  type ProductBarcodeBusinessType,
} from "@/lib/product-barcode/types";
import type { AuthSession } from "@/lib/workspace-models";
import { Section, type ProductStateAction } from "./product-tab-shared";

export function TabProductBusinessBranch({
  value,
  onChange,
  auth,
  language,
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language: string;
}) {
  const textBB = getBarcodeText(language);
  const [pickerType, setPickerType] = useState<"branch" | "businesstype" | null>(null);

  const setBusinessRows = useCallback(
    (mutator: (rows: ProductBarcodeBusinessType[]) => ProductBarcodeBusinessType[]) =>
      onChange((c) => c ? ({ ...c, businesstypes: mutator(c.businesstypes || []) } as Product) : null),
    [onChange],
  );

  const setBranchRows = useCallback(
    (mutator: (rows: ProductBarcodeBranch[]) => ProductBarcodeBranch[]) =>
      onChange((c) => c ? ({ ...c, ignorebranches: mutator(c.ignorebranches || []) } as Product) : null),
    [onChange],
  );

  return (
    <div className="space-y-4">
      {/* Business Types (Ignore) */}
      <Section
        title={textBB.businessTypeSection}
        action={
          <Button type="button" variant="outline" size="sm" onClick={() => setPickerType("businesstype")}>
            <Plus className="mr-1 h-4 w-4" />
            {textBB.businessTypeAddBtn}
          </Button>
        }
      >
        {(!value.businesstypes || value.businesstypes.length === 0) ? (
          <p className="text-sm text-muted-foreground">{textBB.businessTypeNoData}</p>
        ) : (
          <ul className="space-y-1">
            {value.businesstypes.map((entry, idx) => (
              <li key={entry.guidfixed || idx} className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1.5 bg-muted/10 text-sm">
                <span>{pickName(entry.names, language) || entry.code}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 text-destructive hover:bg-destructive/10"
                  onClick={() => setBusinessRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Minus className="h-4 w-4" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </Section>

      {/* Ignore Branches */}
      <Section
        title={textBB.branchSection}
        action={
          <Button type="button" variant="outline" size="sm" onClick={() => setPickerType("branch")}>
            <Plus className="mr-1 h-4 w-4" />
            {textBB.branchAddBtn}
          </Button>
        }
      >
        <p className="mb-2 text-xs text-muted-foreground">{textBB.branchHintDetail}</p>
        {(!value.ignorebranches || value.ignorebranches.length === 0) ? (
          <p className="text-sm text-muted-foreground">{textBB.branchNoData}</p>
        ) : (
          <ul className="space-y-1">
            {value.ignorebranches.map((entry, idx) => (
              <li key={entry.guidfixed || idx} className="flex items-center justify-between gap-2 rounded border border-border px-2 py-1.5 bg-muted/10 text-sm">
                <span>{pickName(entry.names, language) || entry.code}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 text-destructive hover:bg-destructive/10"
                  onClick={() => setBranchRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
                >
                  <Minus className="h-4 w-4" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </Section>

      <MasterPicker
        open={pickerType !== null}
        onClose={() => setPickerType(null)}
        auth={auth}
        language={language}
        master={pickerType === "branch" ? "branch" : "businesstype"}
        title={pickerType === "branch" ? textBB.pickerBranch : textBB.pickerBusinessType}
        onSelect={(entry) => {
          if (pickerType === "branch") {
            setBranchRows((rows) => [
              ...rows,
              { guidfixed: entry.guidfixed, code: entry.code, names: entry.names, isignore: true },
            ]);
          } else {
            setBusinessRows((rows) => [
              ...rows,
              { guidfixed: entry.guidfixed, code: entry.code, names: entry.names, isignore: false },
            ]);
          }
        }}
      />
    </div>
  );
}
