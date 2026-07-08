"use client";

import { useCallback, useMemo, useState } from "react";
import { Plus, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { pickName } from "@/lib/product-barcode/utils";
import {
  type Product,
  type ProductOption,
  type ProductOrderType,
  type ProductRestaurant,
} from "@/lib/product-barcode/types";
import type { AuthSession } from "@/lib/workspace-models";
import {
  FieldRow,
  NumberField,
  RadioOptionGroup,
  Section,
  Toggle,
  type ProductStateAction,
} from "./product-tab-shared";

function cryptoRandomId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return (crypto as Crypto).randomUUID();
  }
  return Math.random().toString(36).slice(2);
}

export function TabProductRestaurant({
  value,
  onChange,
  auth,
  language,
  shopLanguages,
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language: string;
  shopLanguages: string[];
}) {
  const text = getBarcodeText(language);
  const foodTypes = useMemo(() => [
    { value: 0, label: text.foodTypeFood },
    { value: 1, label: text.foodTypeDrink },
    { value: 2, label: text.foodTypeAlcohol },
    { value: 3, label: text.foodTypeOther },
  ], [text]);

  const updR = useCallback(
    (key: keyof ProductRestaurant, val: boolean) =>
      onChange((c) => {
        if (!c) return null;
        return {
          ...c,
          restaurant: {
            ...(c.restaurant || {
              isforrestaurant: false,
              isfortakeaway: false,
              isfordelivery: false,
              isforcustomer: false,
              isforcustomerpreorder: false,
            }),
            [key]: val,
          },
        } as Product;
      }),
    [onChange],
  );

  return (
    <div className="space-y-4">
      <Section title={text.tabRestaurant}>
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          <Toggle checked={value.restaurant?.isforrestaurant ?? false} onCheckedChange={(n) => updR("isforrestaurant", n)} label={text.isForRestaurant} />
          <Toggle checked={value.restaurant?.isfortakeaway ?? false} onCheckedChange={(n) => updR("isfortakeaway", n)} label={text.isForTakeaway} />
          <Toggle checked={value.restaurant?.isfordelivery ?? false} onCheckedChange={(n) => updR("isfordelivery", n)} label={text.isForDelivery} />
          <Toggle checked={value.restaurant?.isforcustomer ?? false} onCheckedChange={(n) => updR("isforcustomer", n)} label={text.isForCustomer} />
          <Toggle checked={value.restaurant?.isforcustomerpreorder ?? false} onCheckedChange={(n) => updR("isforcustomerpreorder", n)} label={text.isForCustomerPreOrder} />
          <Toggle checked={value.isalacarte ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isalacarte: n } as Product) : null)} label={text.isALaCarte} />
          <Toggle checked={value.isstockforrestaurant ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isstockforrestaurant: n } as Product) : null)} label={text.isStockForRestaurant} />
          <Toggle checked={value.issplitunitprint ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, issplitunitprint: n } as Product) : null)} label={text.isSplitUnitPrint} />
          <Toggle checked={value.isonlystaff ?? false} onCheckedChange={(n) => onChange((c) => c ? ({ ...c, isonlystaff: n } as Product) : null)} label={text.isOnlyStaff} />
        </div>
        <div className="mt-3">
          <RadioOptionGroup
            label={text.foodType}
            value={value.foodtype ?? 0}
            onChange={(val) => onChange((c) => c ? ({ ...c, foodtype: val } as Product) : null)}
            options={foodTypes}
          />
        </div>
      </Section>

      <ProductOrderTypesEditor value={value} onChange={onChange} language={language} auth={auth} />

      <ProductOptionsEditor value={value} onChange={onChange} shopLanguages={shopLanguages} language={language} />
    </div>
  );
}

function ProductOrderTypesEditor({
  value,
  onChange,
  language,
  auth,
}: {
  value: Product;
  onChange: ProductStateAction;
  language: string;
  auth: AuthSession | null;
}) {
  const [picker, setPicker] = useState({ open: false, idx: undefined as number | undefined });
  const setRows = useCallback(
    (mutator: (rows: ProductOrderType[]) => ProductOrderType[]) =>
      onChange((c) => c ? ({ ...c, ordertypes: mutator(c.ordertypes || []) } as Product) : null),
    [onChange],
  );

  const textOT = getBarcodeText(language);
  return (
    <Section
      title={textOT.orderTypes}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setRows((rows) => [...rows, { guidfixed: "", code: "", names: [], chargeprice: 0, isdisabled: false } as any])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {textOT.orderTypeAdd}
        </Button>
      }
    >
      {(!value.ordertypes || value.ordertypes.length === 0) ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-2">
          {value.ordertypes.map((entry, idx) => (
            <div key={idx} className="grid grid-cols-1 items-center gap-2 md:grid-cols-[2fr_1fr_40px]">
              <button
                type="button"
                onClick={() => setPicker({ open: true, idx })}
                className="flex h-10 w-full items-center justify-between rounded-lg border border-input bg-background px-3 text-left text-sm hover:bg-muted/40"
              >
                <span className="truncate">{pickName(entry.names, language) || entry.code || "— เลือกบริการสั่งอาหาร —"}</span>
                <span className="text-xs text-muted-foreground">{entry.code}</span>
              </button>
              <NumberField
                value={entry.chargeprice ?? 0}
                onChange={(n) =>
                  setRows((rows) => rows.map((row, rowIdx) => (rowIdx === idx ? { ...row, chargeprice: n } : row)))
                }
                step="any"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setRows((rows) => rows.filter((_, rowIdx) => rowIdx !== idx))}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
      <MasterPicker
        open={picker.open}
        onClose={() => setPicker({ open: false, idx: undefined })}
        auth={auth}
        language={language}
        master="ordertype"
        title={textOT.orderTypes}
        onSelect={(entry) =>
          setRows((rows) =>
            rows.map((row, rowIdx) =>
              rowIdx === picker.idx
                ? ({ ...row, guidfixed: entry.guidfixed, code: entry.code, names: entry.names } as any)
                : row,
            ),
          )
        }
      />
    </Section>
  );
}

function ProductOptionsEditor({
  value,
  onChange,
  shopLanguages,
  language,
}: {
  value: Product;
  onChange: ProductStateAction;
  shopLanguages: string[];
  language: string;
}) {
  const textOpt = getBarcodeText(language);
  const setOptions = useCallback(
    (mutator: (rows: ProductOption[]) => ProductOption[]) =>
      onChange((c) => c ? ({ ...c, options: mutator(c.options || []) } as Product) : null),
    [onChange],
  );

  return (
    <Section
      title={textOpt.options}
      action={
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            setOptions((rows) => [
              ...rows,
              { guid: cryptoRandomId(), names: [], choicetype: 0, choices: [] },
            ])
          }
        >
          <Plus className="mr-1 h-4 w-4" />
          {textOpt.optionAdd}
        </Button>
      }
    >
      {(!value.options || value.options.length === 0) ? (
        <p className="text-sm text-muted-foreground">—</p>
      ) : (
        <div className="space-y-4">
          {value.options.map((opt, optIdx) => (
            <div key={opt.guid} className="space-y-2 rounded-md border border-border p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground font-semibold">กลุ่มตัวเลือกที่ #{optIdx + 1}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="text-destructive hover:bg-destructive/10"
                  onClick={() => setOptions((rows) => rows.filter((_, idx) => idx !== optIdx))}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>

              <NamesEditor
                names={opt.names || []}
                onChange={(nextNames) =>
                  setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, names: nextNames } : row)))
                }
                languages={shopLanguages}
                label={textOpt.optionGroupName}
                language={language}
              />

              <div className="grid gap-3 sm:grid-cols-3">
                <RadioOptionGroup
                  label={textOpt.optionChoiceType}
                  value={opt.choicetype}
                  onChange={(n) =>
                    setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, choicetype: n } : row)))
                  }
                  options={[
                    { value: 0, label: textOpt.optionChoiceTypeMulti },
                    { value: 1, label: textOpt.optionChoiceTypeSingle },
                  ]}
                />
                <FieldRow label={textOpt.optionMinSelectLabel}>
                  <NumberField
                    value={opt.minselect ?? 0}
                    onChange={(n) =>
                      setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, minselect: n } : row)))
                    }
                    min={0}
                    step={1}
                  />
                </FieldRow>
                <FieldRow label={textOpt.optionMaxSelectLabel}>
                  <NumberField
                    value={opt.maxselect ?? 0}
                    onChange={(n) =>
                      setOptions((rows) => rows.map((row, idx) => (idx === optIdx ? { ...row, maxselect: n } : row)))
                    }
                    min={0}
                    step={1}
                  />
                </FieldRow>
              </div>

              {/* Choices inside Option */}
              <div className="mt-3 border-t border-border pt-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-semibold text-muted-foreground">{textOpt.optionChoiceList}</span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setOptions((rows) =>
                        rows.map((row, idx) =>
                          idx === optIdx
                            ? {
                                ...row,
                                choices: [
                                  ...(row.choices || []),
                                  {
                                    guid: cryptoRandomId(),
                                    names: [],
                                    imageuri: "",
                                    refbarcode: "",
                                    refbarcodenames: [],
                                    refproductcode: "",
                                    refunitcode: "",
                                    isstock: false,
                                    isdefault: false,
                                    qty: 1,
                                    price: "",
                                    vatcal: 0,
                                  },
                                ],
                              }
                            : row,
                        ),
                      )
                    }
                  >
                    <Plus className="mr-1 h-3.5 w-3.5" />
                    {textOpt.optionAddChoice}
                  </Button>
                </div>

                {(!opt.choices || opt.choices.length === 0) ? (
                  <p className="text-xs text-muted-foreground text-center py-2">—</p>
                ) : (
                  <div className="space-y-3">
                    {opt.choices.map((choice, choiceIdx) => (
                      <div key={choice.guid} className="p-3 border border-border/60 rounded bg-muted/20 space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-xs text-muted-foreground font-semibold">ตัวเลือกย่อย #{choiceIdx + 1}</span>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="size-6 text-destructive hover:bg-destructive/10"
                            onClick={() =>
                              setOptions((rows) =>
                                rows.map((row, idx) =>
                                  idx === optIdx
                                    ? { ...row, choices: (row.choices || []).filter((_, cIdx) => cIdx !== choiceIdx) }
                                    : row,
                                ),
                              )
                            }
                          >
                            <X className="h-3.5 w-3.5" />
                          </Button>
                        </div>

                        <NamesEditor
                          names={choice.names || []}
                          onChange={(nextNames) =>
                            setOptions((rows) =>
                              rows.map((row, idx) =>
                                idx === optIdx
                                  ? {
                                      ...row,
                                      choices: (row.choices || []).map((c, cIdx) =>
                                        cIdx === choiceIdx ? { ...c, names: nextNames } : c,
                                      ),
                                    }
                                  : row,
                              ),
                            )
                          }
                          languages={shopLanguages}
                          label={textOpt.optionChoiceName}
                          language={language}
                        />

                        <div className="grid gap-3 sm:grid-cols-4">
                          <FieldRow label={textOpt.optionChoicePriceLabel}>
                            <Input
                              value={choice.price || ""}
                              onChange={(e) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, price: e.target.value } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                              placeholder="0.00"
                            />
                          </FieldRow>
                          <FieldRow label={textOpt.optionChoiceQtyLabel}>
                            <NumberField
                              value={choice.qty ?? 0}
                              onChange={(n) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, qty: n } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                            />
                          </FieldRow>
                          <div className="flex items-center pt-5">
                            <Toggle
                              checked={choice.isdefault}
                              onCheckedChange={(n) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, isdefault: n } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                              label={textOpt.optionChoiceDefaultLabel}
                            />
                          </div>
                          <div className="flex items-center pt-5">
                            <Toggle
                              checked={choice.isstock}
                              onCheckedChange={(n) =>
                                setOptions((rows) =>
                                  rows.map((row, idx) =>
                                    idx === optIdx
                                      ? {
                                          ...row,
                                          choices: (row.choices || []).map((c, cIdx) =>
                                            cIdx === choiceIdx ? { ...c, isstock: n } : c,
                                          ),
                                        }
                                      : row,
                                  ),
                                )
                              }
                              label={textOpt.optionChoiceStockLabel}
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </Section>
  );
}
