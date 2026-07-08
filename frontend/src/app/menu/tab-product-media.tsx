"use client";

import { useCallback, useState } from "react";
import { ImagePlus, Plus, Trash2, Upload, X } from "lucide-react";
import { AuthenticatedImg } from "@/components/authenticated-image";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { uploadProductImage } from "@/lib/product-barcode/api";
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
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string>("");

  const handleUpload = useCallback(
    async (file: File | null, target: "main" | "gallery") => {
      if (!file) return;
      if (file.type !== "image/png" && file.type !== "image/jpeg") {
        setUploadError(
          language === "th"
            ? "รองรับเฉพาะไฟล์ PNG และ JPG"
            : "Only PNG and JPG files are supported.",
        );
        return;
      }
      setUploading(true);
      setUploadError("");
      const result = await uploadProductImage(auth, file);
      if (!result.success || !result.data?.url) {
        setUploadError(result.message ?? "Upload failed");
      } else if (target === "main") {
        onChange((c) => c ? ({ ...c, imageuri: result.data!.url }) : null);
      } else {
        onChange((c) => c ? ({
          ...c,
          images: [...(c.images || []), { xorder: (c.images || []).length + 1, uri: result.data!.url }],
        }) : null);
      }
      setUploading(false);
    },
    [auth, onChange, language],
  );

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

      {value.useimageorcolor ?? true ? (
        <>
          <Section title={textM.mediaMainSection}>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <div className="size-32 overflow-hidden rounded-lg border border-border bg-muted">
                {value.imageuri ? (
                  <AuthenticatedImg
                    src={value.imageuri}
                    auth={auth}
                    alt="Main Product"
                    className="size-full object-cover"
                    fallback={
                      <div className="flex size-full items-center justify-center text-xs text-muted-foreground">
                        {textM.mediaNoImage}
                      </div>
                    }
                  />
                ) : (
                  <div className="flex size-full items-center justify-center text-xs text-muted-foreground">
                    {textM.mediaNoImage}
                  </div>
                )}
              </div>
              <div className="space-y-2 flex-1 max-w-md">
                <Input
                  placeholder="https://… หรือ อัปโหลดไฟล์"
                  value={value.imageuri || ""}
                  onChange={(event) => onChange((c) => c ? ({ ...c, imageuri: event.target.value } as Product) : null)}
                />
                <div className="flex flex-wrap items-center gap-2">
                  <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input px-3 py-1.5 text-xs hover:bg-muted bg-background">
                    <Upload className="h-3.5 w-3.5" />
                    {uploading ? textM.mediaUploading : textM.mediaUploadBtn}
                    <input
                      type="file"
                      accept="image/png,image/jpeg"
                      className="hidden"
                      onChange={(event) => handleUpload(event.target.files?.[0] ?? null, "main")}
                      disabled={uploading}
                    />
                  </label>
                  {value.imageuri ? (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => onChange((c) => c ? ({ ...c, imageuri: "" } as Product) : null)}
                    >
                      <Trash2 className="mr-1 h-3.5 w-3.5" />
                      {textM.mediaDeleteBtn}
                    </Button>
                  ) : null}
                </div>
                {uploadError ? <p className="text-xs text-destructive">{uploadError}</p> : null}
              </div>
            </div>
          </Section>
          <Section
            title={textM.mediaGallerySection}
            action={
              <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input px-3 py-1.5 text-xs hover:bg-muted bg-background">
                <Plus className="h-3.5 w-3.5" />
                {textM.mediaAddGallery}
                <input
                  type="file"
                  accept="image/png,image/jpeg"
                  className="hidden"
                  onChange={(event) => handleUpload(event.target.files?.[0] ?? null, "gallery")}
                  disabled={uploading}
                />
              </label>
            }
          >
            {(!value.images || value.images.length === 0) ? (
              <p className="text-sm text-muted-foreground">{textM.mediaNoGallery}</p>
            ) : (
              <ul className="grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6">
                {value.images.map((img, idx) => (
                  <li key={img.xorder} className="relative overflow-hidden rounded-md border border-border">
                    <AuthenticatedImg
                      src={img.uri}
                      auth={auth}
                      alt={`#${img.xorder}`}
                      className="aspect-square w-full object-cover"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="absolute right-1 top-1 size-6 bg-background/80"
                      onClick={() =>
                        onChange((c) => c ? ({ ...c, images: (c.images || []).filter((_, imgIdx) => imgIdx !== idx) } as Product) : null)
                      }
                      aria-label="Delete Image"
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </li>
                ))}
              </ul>
            )}
          </Section>
        </>
      ) : (
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
      )}
    </div>
  );
}
