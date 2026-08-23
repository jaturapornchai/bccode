"use client";

import { authFetch } from "@/lib/client-auth-session";
import {
  Edit3,
  ImageIcon,
  Loader2,
  Trash2,
  UploadCloud,
  X,
} from "lucide-react";
import Image from "next/image";
import { useEffect, useMemo, useRef, useState } from "react";
import { useAuthenticatedImageDisplaySource } from "@/components/authenticated-image";
import { Button } from "@/components/ui/button";
import type { LanguageCode } from "@/lib/i18n";
import type { SystemSettingField } from "@/lib/system-setting-screens";
import type { AuthSession } from "@/lib/workspace-models";
import { type FormState, isRecord, stringValue } from "../types";

// --- i18n text dictionary ---

type UploadTextKey =
  | "applyCrop"
  | "cancel"
  | "chooseImage"
  | "cropImage"
  | "cropTitle"
  | "imageDisplayFailed"
  | "imageTooLarge"
  | "imageUploadFailed"
  | "imageUploadHint"
  | "removeImage"
  | "uploading";

const uploadText: Record<
  UploadTextKey,
  Partial<Record<LanguageCode, string>> & { en: string; th: string }
> = {
  applyCrop: { th: "ใช้รูปนี้", en: "Apply crop" },
  cancel: { th: "ยกเลิก", en: "Cancel" },
  chooseImage: {
    th: "เลือกรูป",
    en: "Choose image",
    cn: "选择图片",
    ja: "画像を選択",
    ko: "이미지 선택",
    lo: "ເລືອກຮູບ",
    my: "ပုံရွေးပါ",
    km: "ជ្រើសរូបភាព",
    vi: "Chọn hình ảnh",
    ms: "Pilih imej",
    id: "Pilih gambar",
    fil: "Pumili ng larawan",
  },
  cropImage: { th: "แก้ไขรูป", en: "Edit image" },
  cropTitle: { th: "ครอปรูป", en: "Crop image" },
  imageTooLarge: {
    th: "ไฟล์รูปต้องไม่เกิน 12 MB",
    en: "Image must not exceed 12 MB",
    cn: "图片不能超过 12 MB",
    ja: "画像は 12 MB 以下にしてください",
    ko: "이미지는 12 MB를 초과할 수 없습니다",
    lo: "ຮູບຕ້ອງບໍ່ເກີນ 12 MB",
    my: "ပုံဖိုင်သည် 12 MB ထက်မကြီးရပါ",
    km: "រូបភាពមិនត្រូវលើស 12 MB",
    vi: "Ảnh không được vượt quá 12 MB",
    ms: "Imej tidak boleh melebihi 12 MB",
    id: "Gambar tidak boleh lebih dari 12 MB",
    fil: "Hindi dapat lumampas sa 12 MB ang larawan",
  },
  imageUploadFailed: {
    th: "อัปโหลดรูปไม่สำเร็จ",
    en: "Image upload failed",
    cn: "图片上传失败",
    ja: "画像のアップロードに失敗しました",
    ko: "이미지 업로드 실패",
    lo: "ອັບໂຫຼດຮູບບໍ່ສຳເລັດ",
    my: "ပုံတင်ခြင်း မအောင်မြင်ပါ",
    km: "អាប់ឡូតរូបភាពមិនបានសម្រេច",
    vi: "Tải ảnh lên thất bại",
    ms: "Muat naik imej gagal",
    id: "Unggah gambar gagal",
    fil: "Nabigo ang pag-upload ng larawan",
  },
  imageUploadHint: {
    th: "รองรับ JPG, PNG และย่อรูปก่อนอัปโหลด",
    en: "JPG, PNG. The image is resized before upload.",
    cn: "支持 JPG、PNG，上传前会缩小图片",
    ja: "JPG、PNG 対応。アップロード前にリサイズします。",
    ko: "JPG, PNG 지원. 업로드 전에 크기를 줄입니다.",
    lo: "ຮອງຮັບ JPG, PNG ແລະຫຍໍ້ຮູບກ່ອນອັບໂຫຼດ",
    my: "JPG, PNG. မတင်မီ အရွယ်အစားလျှော့ပါမည်",
    km: "គាំទ្រ JPG, PNG ហើយបង្រួមរូបភាពមុនអាប់ឡូត",
    vi: "Hỗ trợ JPG, PNG và tự thu nhỏ trước khi tải lên",
    ms: "Sokong JPG, PNG dan saiz imej dikecilkan sebelum muat naik",
    id: "Mendukung JPG, PNG dan gambar diperkecil sebelum diunggah",
    fil: "Suportado ang JPG, PNG at nire-resize bago i-upload",
  },
  imageDisplayFailed: {
    th: "ไม่สามารถแสดงรูปได้",
    en: "Cannot display image",
    cn: "无法显示图片",
    ja: "画像を表示できません",
    ko: "이미지를 표시할 수 없습니다",
    lo: "ບໍ່ສາມາດສະແດງຮູບໄດ້",
    my: "ပုံကို ပြသ၍မရပါ",
    km: "មិនអាចបង្ហាញរូបភាពបាន",
    vi: "Không thể hiển thị ảnh",
    ms: "Tidak dapat memaparkan imej",
    id: "Tidak dapat menampilkan gambar",
    fil: "Hindi maipakita ang larawan",
  },
  removeImage: {
    th: "ลบรูป",
    en: "Remove image",
    cn: "移除图片",
    ja: "画像を削除",
    ko: "이미지 제거",
    lo: "ລຶບຮູບ",
    my: "ပုံဖယ်ရှားပါ",
    km: "លុបរូបភាព",
    vi: "Xóa hình ảnh",
    ms: "Buang imej",
    id: "Hapus gambar",
    fil: "Alisin ang larawan",
  },
  uploading: {
    th: "กำลังอัปโหลด",
    en: "Uploading",
    cn: "正在上传",
    ja: "アップロード中",
    ko: "업로드 중",
    lo: "ກຳລັງອັບໂຫຼດ",
    my: "တင်နေသည်",
    km: "កំពុងអាប់ឡូត",
    vi: "Đang tải lên",
    ms: "Sedang memuat naik",
    id: "Mengunggah",
    fil: "Ina-upload",
  },
};

export function uploadUiText(language: LanguageCode, key: UploadTextKey): string {
  return uploadText[key][language] ?? uploadText[key].en;
}

// --- Helpers ---

function isFailed(payload: unknown): boolean {
  if (!isRecord(payload)) return false;
  return payload.success === false || payload.status === "error";
}

function extractMessage(payload: unknown): string | undefined {
  if (!isRecord(payload)) return undefined;
  const msg = payload.message ?? payload.error ?? payload.err;
  return typeof msg === "string" && msg ? msg : undefined;
}

export function extractUploadUri(payload: unknown): string {
  if (!isRecord(payload)) return "";
  const direct = stringValue(
    payload.uri ??
      payload.url ??
      payload.fileurl ??
      payload.fileUrl ??
      payload.imageurl ??
      payload.imageuri,
  );
  if (direct) return direct;
  const data = payload.data;
  if (!isRecord(data)) return "";
  return stringValue(
    data.uri ??
      data.url ??
      data.fileurl ??
      data.fileUrl ??
      data.imageurl ??
      data.imageuri,
  );
}

function clampNumber(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function loadImageElement(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new window.Image();
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error("image load failed"));
    image.src = src;
  });
}

function canvasToBlob(
  canvas: HTMLCanvasElement,
  type: string,
  quality?: number,
): Promise<Blob | null> {
  return new Promise((resolve) => canvas.toBlob(resolve, type, quality));
}

export async function resizeLogoFile(file: File): Promise<File> {
  const isPngByName = /\.png$/i.test(file.name);
  const isPngByType = file.type === "image/png";
  if (!isPngByName && !isPngByType) {
    throw new Error("โลโก้ต้องเป็นไฟล์ PNG เท่านั้น");
  }
  const objectUrl = URL.createObjectURL(file);
  try {
    const image = await loadImageElement(objectUrl);
    const maxSide = Math.max(image.naturalWidth, image.naturalHeight);
    if (!maxSide) return file;
    const scale = Math.min(1, 512 / maxSide);
    if (scale === 1 && file.size <= 300 * 1024) return file;
    const canvas = document.createElement("canvas");
    canvas.width = Math.max(1, Math.round(image.naturalWidth * scale));
    canvas.height = Math.max(1, Math.round(image.naturalHeight * scale));
    const context = canvas.getContext("2d");
    if (!context) return file;
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = "high";
    context.drawImage(image, 0, 0, canvas.width, canvas.height);
    const blob = await canvasToBlob(canvas, "image/png");
    if (!blob) return file;
    const baseName = file.name.replace(/\.[^.]+$/, "") || "company-logo";
    return new File([blob], `${baseName}.png`, {
      type: "image/png",
      lastModified: Date.now(),
    });
  } finally {
    URL.revokeObjectURL(objectUrl);
  }
}

export async function resizeImageFile(file: File, maxSide = 1280): Promise<File> {
  const objectUrl = URL.createObjectURL(file);
  try {
    const image = await loadImageElement(objectUrl);
    const largest = Math.max(image.naturalWidth, image.naturalHeight);
    if (!largest) return file;
    const scale = Math.min(1, maxSide / largest);
    if (scale === 1 && file.size <= 400 * 1024) return file;
    const canvas = document.createElement("canvas");
    canvas.width = Math.max(1, Math.round(image.naturalWidth * scale));
    canvas.height = Math.max(1, Math.round(image.naturalHeight * scale));
    const context = canvas.getContext("2d");
    if (!context) return file;
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = "high";
    context.drawImage(image, 0, 0, canvas.width, canvas.height);
    const outputType = file.type === "image/png" ? "image/png" : "image/jpeg";
    const blob = await canvasToBlob(canvas, outputType, 0.85);
    if (!blob) return file;
    const baseName = file.name.replace(/\.[^.]+$/, "") || "image";
    const ext = outputType === "image/png" ? "png" : "jpg";
    return new File([blob], `${baseName}.${ext}`, {
      type: outputType,
      lastModified: Date.now(),
    });
  } finally {
    URL.revokeObjectURL(objectUrl);
  }
}

export async function makeThumbnailFile(file: File, maxSide = 256): Promise<File> {
  const objectUrl = URL.createObjectURL(file);
  try {
    const image = await loadImageElement(objectUrl);
    const largest = Math.max(image.naturalWidth, image.naturalHeight);
    const scale = largest > 0 ? Math.min(1, maxSide / largest) : 1;
    const canvas = document.createElement("canvas");
    canvas.width = Math.max(1, Math.round(image.naturalWidth * scale));
    canvas.height = Math.max(1, Math.round(image.naturalHeight * scale));
    const context = canvas.getContext("2d");
    if (!context) return file;
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = "high";
    context.drawImage(image, 0, 0, canvas.width, canvas.height);
    const outputType = file.type === "image/png" ? "image/png" : "image/jpeg";
    const blob = await canvasToBlob(canvas, outputType, 0.82);
    if (!blob) return file;
    const baseName = file.name.replace(/\.[^.]+$/, "") || "image";
    const ext = outputType === "image/png" ? "png" : "jpg";
    return new File([blob], `${baseName}_thumb.${ext}`, {
      type: outputType,
      lastModified: Date.now(),
    });
  } finally {
    URL.revokeObjectURL(objectUrl);
  }
}

// --- ImageCropDialog ---

function drawCroppedImage(
  canvas: HTMLCanvasElement,
  image: HTMLImageElement,
  crop: { offsetX: number; offsetY: number; zoom: number },
) {
  const size = Math.min(image.naturalWidth, image.naturalHeight);
  const cropSize = Math.max(1, size / Math.max(1, crop.zoom));
  const maxX = Math.max(0, image.naturalWidth - cropSize);
  const maxY = Math.max(0, image.naturalHeight - cropSize);
  const sourceX = clampNumber(
    (image.naturalWidth - cropSize) / 2 + (crop.offsetX / 100) * (maxX / 2),
    0,
    maxX,
  );
  const sourceY = clampNumber(
    (image.naturalHeight - cropSize) / 2 + (crop.offsetY / 100) * (maxY / 2),
    0,
    maxY,
  );
  const context = canvas.getContext("2d");
  if (!context) return;
  context.clearRect(0, 0, canvas.width, canvas.height);
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.drawImage(
    image,
    sourceX,
    sourceY,
    cropSize,
    cropSize,
    0,
    0,
    canvas.width,
    canvas.height,
  );
}

export function ImageCropDialog({
  imageUrl,
  language,
  onApply,
  onCancel,
  outputType = "image/jpeg",
}: {
  imageUrl: string;
  language: LanguageCode;
  onApply: (file: File) => void;
  onCancel: () => void;
  outputType?: string;
}) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [image, setImage] = useState<HTMLImageElement | null>(null);
  const [offsetX, setOffsetX] = useState(0);
  const [offsetY, setOffsetY] = useState(0);
  const [zoom, setZoom] = useState(1);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    const nextImage = new window.Image();
    nextImage.crossOrigin = "anonymous";
    nextImage.onload = () => {
      if (!active) return;
      setImage(nextImage);
      setOffsetX(0);
      setOffsetY(0);
      setZoom(1);
      setError("");
    };
    nextImage.onerror = () => {
      if (active) setError(uploadUiText(language, "imageUploadFailed"));
    };
    nextImage.src = imageUrl;
    return () => {
      active = false;
    };
  }, [imageUrl, language]);

  useEffect(() => {
    if (!image || !canvasRef.current) return;
    drawCroppedImage(canvasRef.current, image, { offsetX, offsetY, zoom });
  }, [image, offsetX, offsetY, zoom]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onCancel]);

  async function apply() {
    if (!canvasRef.current) return;
    try {
      const blob = await canvasToBlob(canvasRef.current, outputType, 0.86);
      if (!blob) throw new Error(uploadUiText(language, "imageUploadFailed"));
      const ext = outputType === "image/png" ? "png" : "jpg";
      onApply(
        new File([blob], `cropped-image.${ext}`, {
          type: outputType,
          lastModified: Date.now(),
        }),
      );
    } catch {
      setError(uploadUiText(language, "imageUploadFailed"));
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-2"
      role="dialog"
      aria-modal="true"
    >
      <section className="grid w-[min(420px,calc(100vw-16px))] gap-2 rounded-2xl border border-border bg-card p-3 text-card-foreground shadow-xl">
        <div className="flex items-center justify-between gap-2">
          <div className="text-sm font-semibold">
            {uploadUiText(language, "cropTitle")}
          </div>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onCancel}
            aria-label={uploadUiText(language, "cancel")}
          >
            <X />
          </Button>
        </div>
        <div className="grid place-items-center rounded-2xl border border-input bg-muted/40 p-2">
          <canvas
            ref={canvasRef}
            width={512}
            height={512}
            className="aspect-square w-full max-w-80 rounded-xl bg-background object-contain"
          />
        </div>
        <label className="grid gap-1 text-xs font-semibold">
          <span>Zoom</span>
          <input
            type="range"
            min="1"
            max="3"
            step="0.01"
            value={zoom}
            onChange={(event) => setZoom(Number(event.target.value))}
          />
        </label>
        <label className="grid gap-1 text-xs font-semibold">
          <span>X</span>
          <input
            type="range"
            min="-100"
            max="100"
            step="1"
            value={offsetX}
            onChange={(event) => setOffsetX(Number(event.target.value))}
          />
        </label>
        <label className="grid gap-1 text-xs font-semibold">
          <span>Y</span>
          <input
            type="range"
            min="-100"
            max="100"
            step="1"
            value={offsetY}
            onChange={(event) => setOffsetY(Number(event.target.value))}
          />
        </label>
        {error ? (
          <p className="text-xs font-semibold text-destructive">{error}</p>
        ) : null}
        <div className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            {uploadUiText(language, "cancel")}
          </Button>
          <Button type="button" onClick={() => void apply()} disabled={!image}>
            {uploadUiText(language, "applyCrop")}
          </Button>
        </div>
      </section>
    </div>
  );
}

// --- toUriArray helper ---

export function toUriArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map((item) => stringValue(item)).filter(Boolean);
  }
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  if (trimmed.startsWith("[")) {
    try {
      const parsed = JSON.parse(trimmed) as unknown;
      if (Array.isArray(parsed)) {
        return parsed.map((item) => stringValue(item)).filter(Boolean);
      }
    } catch {
      // fall through to single value
    }
  }
  return [trimmed];
}

// --- ImageUploadFieldEditor ---

export function ImageUploadFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState("");
  const [cropSource, setCropSource] = useState("");
  const [localPreview, setLocalPreview] = useState("");
  const [uploading, setUploading] = useState(false);
  const value = stringValue(form[field.key]);
  const {
    displayUrl: previewValue,
    failed: previewLoadFailed,
    loading: previewLoading,
  } = useAuthenticatedImageDisplaySource(localPreview || value, auth);
  const hasImageValue = Boolean(localPreview || value);
  const previewStyle = previewValue
    ? { backgroundImage: `url(${JSON.stringify(previewValue)})` }
    : undefined;

  const formGuid = String(form.guidfixed || form.guidfixed || form.guid || "");
  useEffect(() => {
    setLocalPreview("");
    setError("");
  }, [formGuid]);

  useEffect(() => {
    if (!localPreview) return;
    return () => URL.revokeObjectURL(localPreview);
  }, [localPreview]);

  function clearImage() {
    setLocalPreview("");
    setCropSource("");
    const next: FormState = { ...form, [field.key]: "" };
    if (field.thumbnailKey) next[field.thumbnailKey] = "";
    setForm(next);
  }

  async function uploadImageFile(file: File) {
    if (!auth) {
      setError(uploadUiText(language, "imageUploadFailed"));
      return;
    }
    const session = auth;
    const uploadOne = async (toUpload: File): Promise<string> => {
      const uploadForm = new FormData();
      uploadForm.append("file", toUpload, toUpload.name);
      uploadForm.append("category", `system-settings/${field.key}`);
      const response = await authFetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": session.backendUrl,
          Authorization: `Bearer ${session.token}`,
        },
        body: uploadForm,
      });
      const payload = (await response.json()) as unknown;
      if (!response.ok || isFailed(payload))
        throw new Error(
          extractMessage(payload) ?? uploadUiText(language, "imageUploadFailed"),
        );
      const uri = extractUploadUri(payload);
      if (!uri) throw new Error(uploadUiText(language, "imageUploadFailed"));
      return uri;
    };

    setUploading(true);
    setError("");
    try {
      if (field.thumbnailKey) {
        const originalUri = await uploadOne(file);
        const thumbUri = await uploadOne(await makeThumbnailFile(file));
        setForm({
          ...form,
          [field.key]: originalUri,
          [field.thumbnailKey]: thumbUri,
        });
      } else {
        const uri = await uploadOne(await resizeLogoFile(file));
        setForm({ ...form, [field.key]: uri });
      }
    } catch (uploadError) {
      setError(
        uploadError instanceof Error && uploadError.message
          ? uploadError.message
          : uploadUiText(language, "imageUploadFailed"),
      );
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  }

  async function applyCroppedFile(file: File) {
    const objectUrl = URL.createObjectURL(file);
    setLocalPreview(objectUrl);
    setCropSource("");
    await uploadImageFile(file);
  }

  async function handleFile(file: File | undefined) {
    if (!file || uploading) return;
    const allowedTypes = (field.acceptTypes ?? "image/png,image/jpeg")
      .split(",")
      .map((type) => type.trim())
      .filter(Boolean);
    if (!allowedTypes.includes(file.type)) {
      const onlyPng = allowedTypes.length === 1 && allowedTypes[0] === "image/png";
      setError(
        language === "th"
          ? onlyPng
            ? "รองรับเฉพาะไฟล์ PNG"
            : "รองรับเฉพาะไฟล์ PNG และ JPG"
          : onlyPng
            ? "Only PNG files are supported."
            : "Only PNG and JPG files are supported.",
      );
      return;
    }
    if (file.size > 12 * 1024 * 1024) {
      setError(uploadUiText(language, "imageTooLarge"));
      return;
    }
    setLocalPreview(URL.createObjectURL(file));
    await uploadImageFile(file);
  }

  return (
    <section className="grid gap-1 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <div className="flex w-full flex-wrap items-center gap-3">
        <button
          aria-label={label}
          className="grid size-20 place-items-center overflow-hidden rounded-2xl border border-input bg-card bg-contain bg-center bg-no-repeat text-muted-foreground shadow-sm"
          onClick={() => inputRef.current?.click()}
          style={previewStyle}
          type="button"
        >
          {previewValue ? (
            <span className="sr-only">{label}</span>
          ) : previewLoading ? (
            <Loader2 className="size-8 animate-spin" />
          ) : (
            <ImageIcon className="size-8" />
          )}
        </button>
        <div className="grid min-w-48 flex-1 gap-2">
          <div className="flex w-full flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => inputRef.current?.click()}
              disabled={uploading || !auth}
            >
              {uploading ? <Loader2 className="animate-spin" /> : <UploadCloud />}
              {uploading
                ? uploadUiText(language, "uploading")
                : uploadUiText(language, "chooseImage")}
            </Button>
            {previewValue && !field.thumbnailKey ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => setCropSource(previewValue)}
                disabled={uploading || previewLoading}
              >
                <Edit3 />
                {uploadUiText(language, "cropImage")}
              </Button>
            ) : null}
            {hasImageValue ? (
              <Button
                type="button"
                variant="outline"
                onClick={clearImage}
                disabled={uploading}
              >
                <Trash2 />
                {uploadUiText(language, "removeImage")}
              </Button>
            ) : null}
          </div>
          <p className="text-xs font-normal text-muted-foreground">
            {field.acceptTypes === "image/png"
              ? language === "th"
                ? "รองรับเฉพาะไฟล์ PNG พื้นหลังโปร่งใสได้ ใช้สำหรับออกแบบฟอร์มและพิมพ์เอกสาร"
                : "PNG only. Transparent background supported. Used for form design and document printing."
              : uploadUiText(language, "imageUploadHint")}
          </p>
          {error ? (
            <p className="text-xs font-semibold text-destructive">{error}</p>
          ) : null}
          {previewLoadFailed && value ? (
            <p className="break-all text-xs font-semibold text-destructive">
              {uploadUiText(language, "imageDisplayFailed")}
              <br />
              {value}
            </p>
          ) : null}
        </div>
      </div>
      <input
        ref={inputRef}
        className="sr-only"
        type="file"
        accept={field.acceptTypes ?? "image/png,image/jpeg"}
        onChange={(event) => void handleFile(event.target.files?.[0])}
      />
      {cropSource ? (
        <ImageCropDialog
          imageUrl={cropSource}
          language={language}
          outputType={field.acceptTypes === "image/png" ? "image/png" : "image/jpeg"}
          onCancel={() => setCropSource("")}
          onApply={(file) => void applyCroppedFile(file)}
        />
      ) : null}
    </section>
  );
}

// --- ImageUploadReadOnlyDetail ---

export function ImageUploadReadOnlyDetail({
  auth,
  label,
  language,
  value,
}: {
  auth: AuthSession | null;
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const rawValue = stringValue(value);
  const {
    displayUrl,
    failed: fetchFailed,
    requestedUrl,
  } = useAuthenticatedImageDisplaySource(rawValue, auth);
  const [elementFailed, setElementFailed] = useState(false);
  const canShowImage = Boolean(displayUrl) && !fetchFailed && !elementFailed;

  useEffect(() => {
    setElementFailed(false);
  }, [displayUrl]);

  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] transition-all duration-200 border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">{label}</span>
      {canShowImage ? (
        <div className="flex w-full flex-wrap items-start gap-3">
          <div className="relative h-28 w-28 shrink-0 overflow-hidden rounded-xl border border-input bg-muted">
            <Image
              alt={label}
              className="object-cover"
              fill
              sizes="112px"
              src={displayUrl}
              unoptimized
              onError={() => setElementFailed(true)}
            />
          </div>
          <b className="min-w-0 flex-1 break-all text-sm font-medium text-foreground">
            {rawValue || requestedUrl}
          </b>
        </div>
      ) : (
        <div className="grid gap-1">
          {(fetchFailed || elementFailed) && rawValue ? (
            <span className="text-xs font-semibold text-destructive">
              {uploadUiText(language, "imageDisplayFailed")}
            </span>
          ) : null}
          <b className="min-w-0 break-words text-foreground font-medium">
            {rawValue || "-"}
          </b>
        </div>
      )}
    </div>
  );
}

// --- ImageGalleryFieldEditor ---

export function ImageGalleryFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState("");
  const [uploading, setUploading] = useState(false);
  const [cropTarget, setCropTarget] = useState<{ url: string; index: number } | null>(null);
  const values = useMemo(() => toUriArray(form[field.key]), [field.key, form]);

  const addLabel = language === "th" ? "เพิ่มรูป" : "Add image";
  const editLabel = language === "th" ? "แก้ไข" : "Edit";
  const removeLabel = language === "th" ? "ลบ" : "Remove";

  function updateValues(next: string[]) {
    setForm({ ...form, [field.key]: next });
  }

  function removeAt(index: number) {
    updateValues(values.filter((_, idx) => idx !== index));
  }

  function replaceAt(index: number, uri: string) {
    const next = values.slice();
    next[index] = uri;
    updateValues(next);
  }

  async function uploadFile(file: File, replaceIndex?: number) {
    if (!auth) {
      setError(uploadUiText(language, "imageUploadFailed"));
      return;
    }
    setUploading(true);
    setError("");
    try {
      const resizedFile = await resizeImageFile(file);
      const uploadForm = new FormData();
      uploadForm.append("file", resizedFile, resizedFile.name);
      uploadForm.append("category", `system-settings/${field.key}`);
      const response = await authFetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: uploadForm,
      });
      const payload = (await response.json()) as unknown;
      if (!response.ok || isFailed(payload))
        throw new Error(
          extractMessage(payload) ?? uploadUiText(language, "imageUploadFailed"),
        );
      const uri = extractUploadUri(payload);
      if (!uri) throw new Error(uploadUiText(language, "imageUploadFailed"));
      if (typeof replaceIndex === "number") {
        replaceAt(replaceIndex, uri);
      } else {
        updateValues([...values, uri]);
      }
    } catch (uploadError) {
      setError(
        uploadError instanceof Error && uploadError.message
          ? uploadError.message
          : uploadUiText(language, "imageUploadFailed"),
      );
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  }

  async function handleFile(file: File | undefined) {
    if (!file || uploading) return;
    const allowedTypes = (field.acceptTypes ?? "image/png,image/jpeg")
      .split(",")
      .map((type) => type.trim())
      .filter(Boolean);
    if (!allowedTypes.includes(file.type)) {
      const onlyPng = allowedTypes.length === 1 && allowedTypes[0] === "image/png";
      setError(
        language === "th"
          ? onlyPng
            ? "รองรับเฉพาะไฟล์ PNG"
            : "รองรับเฉพาะไฟล์ PNG และ JPG"
          : onlyPng
            ? "Only PNG files are supported."
            : "Only PNG and JPG files are supported.",
      );
      return;
    }
    if (file.size > 12 * 1024 * 1024) {
      setError(uploadUiText(language, "imageTooLarge"));
      return;
    }
    await uploadFile(file);
  }

  async function applyCroppedFile(file: File) {
    const target = cropTarget;
    setCropTarget(null);
    if (!target) return;
    await uploadFile(file, target.index);
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => inputRef.current?.click()}
          disabled={uploading || !auth}
        >
          {uploading ? <Loader2 className="animate-spin" /> : <UploadCloud />}
          {uploading ? uploadUiText(language, "uploading") : addLabel}
        </Button>
      </div>
      {values.length === 0 ? (
        <div className="grid h-20 place-items-center rounded-xl border border-dashed border-input text-xs font-normal text-muted-foreground">
          {language === "th"
            ? "ยังไม่มีรูป — กด \"เพิ่มรูป\" เพื่ออัปโหลด"
            : 'No images yet — click "Add image" to upload'}
        </div>
      ) : (
        <ul className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
          {values.map((uri, index) => (
            <GalleryItemCard
              key={`${uri}-${index}`}
              auth={auth}
              editLabel={editLabel}
              language={language}
              onCrop={(displayUrl) => setCropTarget({ url: displayUrl, index })}
              onRemove={() => removeAt(index)}
              removeLabel={removeLabel}
              uri={uri}
            />
          ))}
        </ul>
      )}
      <p className="text-xs font-normal text-muted-foreground">
        {uploadUiText(language, "imageUploadHint")}
      </p>
      {error ? (
        <p className="text-xs font-semibold text-destructive">{error}</p>
      ) : null}
      <input
        ref={inputRef}
        className="sr-only"
        type="file"
        accept={field.acceptTypes ?? "image/png,image/jpeg"}
        onChange={(event) => void handleFile(event.target.files?.[0])}
      />
      {cropTarget ? (
        <ImageCropDialog
          imageUrl={cropTarget.url}
          language={language}
          onCancel={() => setCropTarget(null)}
          onApply={(file) => void applyCroppedFile(file)}
        />
      ) : null}
    </section>
  );
}

// --- GalleryItemCard ---

function GalleryItemCard({
  auth,
  editLabel,
  language,
  onCrop,
  onRemove,
  removeLabel,
  uri,
}: {
  auth: AuthSession | null;
  editLabel: string;
  language: LanguageCode;
  onCrop: (displayUrl: string) => void;
  onRemove: () => void;
  removeLabel: string;
  uri: string;
}) {
  const { displayUrl, failed, loading } = useAuthenticatedImageDisplaySource(uri, auth);
  return (
    <li className="grid gap-1 rounded-xl border border-input bg-card p-1 shadow-sm">
      <div
        className="relative h-24 w-full overflow-hidden rounded-lg border border-border bg-muted bg-contain bg-center bg-no-repeat"
        style={
          displayUrl
            ? { backgroundImage: `url(${JSON.stringify(displayUrl)})` }
            : undefined
        }
      >
        {!displayUrl ? (
          <div className="grid h-full place-items-center text-muted-foreground">
            {loading ? (
              <Loader2 className="size-6 animate-spin" />
            ) : failed ? (
              <span className="px-2 text-center text-[10px] font-semibold text-destructive">
                {uploadUiText(language, "imageDisplayFailed")}
              </span>
            ) : (
              <ImageIcon className="size-6" />
            )}
          </div>
        ) : null}
      </div>
      <div className="flex flex-wrap items-center justify-between gap-1 px-1 text-[11px] text-muted-foreground">
        <span className="line-clamp-1 min-w-0 break-all" title={uri}>
          {uri}
        </span>
        <div className="flex shrink-0 gap-1">
          {displayUrl ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => onCrop(displayUrl)}
              aria-label={editLabel}
            >
              <Edit3 className="size-3" />
            </Button>
          ) : null}
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onRemove}
            aria-label={removeLabel}
          >
            <Trash2 className="size-3" />
          </Button>
        </div>
      </div>
    </li>
  );
}

// --- ImageGalleryReadOnlyDetail ---

export function ImageGalleryReadOnlyDetail({
  auth,
  label,
  language,
  value,
}: {
  auth: AuthSession | null;
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const uris = useMemo(() => toUriArray(value), [value]);
  if (uris.length === 0) {
    return (
      <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
        <span className="text-xs font-semibold text-muted-foreground">{label}</span>
        <b className="text-foreground font-medium">-</b>
      </div>
    );
  }
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">{label}</span>
      <ul className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
        {uris.map((uri, index) => (
          <GalleryReadOnlyItem
            key={`${uri}-${index}`}
            auth={auth}
            label={label}
            language={language}
            uri={uri}
          />
        ))}
      </ul>
    </div>
  );
}

function GalleryReadOnlyItem({
  auth,
  label,
  language,
  uri,
}: {
  auth: AuthSession | null;
  label: string;
  language: LanguageCode;
  uri: string;
}) {
  const { displayUrl, failed } = useAuthenticatedImageDisplaySource(uri, auth);
  const [elementFailed, setElementFailed] = useState(false);
  const canShow = Boolean(displayUrl) && !failed && !elementFailed;

  useEffect(() => {
    setElementFailed(false);
  }, [displayUrl]);

  return (
    <li className="grid gap-1 rounded-lg border border-input bg-background p-1">
      <div className="relative h-24 w-full overflow-hidden rounded-md border border-border bg-muted">
        {canShow ? (
          <Image
            alt={label}
            className="object-cover"
            fill
            sizes="160px"
            src={displayUrl}
            unoptimized
            onError={() => setElementFailed(true)}
          />
        ) : (
          <div className="grid h-full place-items-center text-[10px] font-semibold text-muted-foreground">
            {failed || elementFailed
              ? uploadUiText(language, "imageDisplayFailed")
              : "…"}
          </div>
        )}
      </div>
      <span className="line-clamp-2 break-all px-1 text-[10px] text-muted-foreground" title={uri}>
        {uri}
      </span>
    </li>
  );
}
