"use client";

import { ImagePlus } from "lucide-react";
import {
  BusinessImageEditor,
  BusinessImageGallery,
} from "@/components/product-barcode/business-image-editor";
import { Input } from "@/components/ui/input";
import { getBarcodeText } from "@/lib/product-barcode/language";
import { type Product } from "@/lib/product-barcode/types";
import type { AuthSession } from "@/lib/workspace-models";
import { FieldRow, FieldGrid, Section, type ProductStateAction } from "./product-tab-shared";

export function TabProductMedia({
  value,
  onChange,
  auth,
  language = "th",
}: {
  value: Product;
  onChange: ProductStateAction;
  auth: AuthSession | null;
  language?: string;
}) {
  const textM = getBarcodeText(language);

  return (
    <div className="space-y-4">
      <Section title={textM.mediaUseSection}>
        <div className="flex gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={value.useimageorcolor ?? true}
              onChange={() => onChange((c) => c ? ({ ...c, useimageorcolor: true } as Product) : null)}
            />
            <ImagePlus className="h-4 w-4" />
            <span>แสดงรูปภาพหลัก</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={!(value.useimageorcolor ?? true)}
              onChange={() => onChange((c) => c ? ({ ...c, useimageorcolor: false } as Product) : null)}
            />
            <span className="inline-block size-4 rounded border" style={{ background: value.colorselecthex || "#888" }} />
            <span>แสดงสีป้ายสินค้า</span>
          </label>
        </div>
      </Section>

      <BusinessImageEditor
        auth={auth}
        language={language}
        value={value}
        onChange={(patch) =>
          onChange((current) =>
            current ? ({ ...current, ...patch } as Product) : null,
          )
        }
      />
      <BusinessImageGallery
        auth={auth}
        sources={(value.barcodes ?? []).map((barcode) => ({
          key: barcode.guidfixed || barcode.barcode,
          label: `${language === "th" ? "บาร์โค้ด" : "Barcode"} ${barcode.barcode}`,
          imageuri: barcode.imageuri,
          images: barcode.images,
          videos: barcode.videos,
        }))}
        title={language === "th" ? "สื่อจากบาร์โค้ด" : "Media from barcodes"}
      />

      {!(value.useimageorcolor ?? true) ? (
        <Section title={textM.mediaColorSection}>
          <FieldGrid>
            <FieldRow label={textM.mediaColorName}>
              <Input
                value={value.colorselect || ""}
                onChange={(event) => onChange((c) => c ? ({ ...c, colorselect: event.target.value } as Product) : null)}
              />
            </FieldRow>
            <FieldRow label={textM.mediaColorHex}>
              <div className="flex items-center gap-2">
                <Input
                  type="color"
                  value={value.colorselecthex || "#888888"}
                  onChange={(event) => onChange((c) => c ? ({ ...c, colorselecthex: event.target.value } as Product) : null)}
                  className="h-10 w-16 p-1 cursor-pointer"
                />
                <Input
                  value={value.colorselecthex || ""}
                  onChange={(event) => onChange((c) => c ? ({ ...c, colorselecthex: event.target.value } as Product) : null)}
                  placeholder="#RRGGBB"
                />
              </div>
            </FieldRow>
          </FieldGrid>
        </Section>
      ) : null}
    </div>
  );
}
