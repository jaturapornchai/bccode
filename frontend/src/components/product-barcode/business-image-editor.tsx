"use client";

import { Film, ImagePlus, Plus, Trash2, Upload, X } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import { AuthenticatedImg, AuthenticatedVideo } from "@/components/authenticated-image";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { uploadProductImage, uploadProductVideo } from "@/lib/product-barcode/api";
import { getBarcodeText } from "@/lib/product-barcode/language";
import {
  PRODUCT_VIDEO_MAX_BYTES,
  PRODUCT_VIDEO_MAX_MB,
  type ProductImage,
  type ProductVideo,
} from "@/lib/product-barcode/types";
import type { AuthSession } from "@/lib/workspace-models";

export type BusinessImageValue = {
  imageuri?: string;
  images?: ProductImage[];
  videos?: ProductVideo[];
};

export type BusinessImageSource = BusinessImageValue & {
  key: string;
  label: string;
};

export function BusinessImageEditor({
  auth,
  galleryTitle,
  language = "th",
  mainTitle,
  onChange,
  value,
}: {
  auth: AuthSession | null;
  galleryTitle?: string;
  language?: string;
  mainTitle?: string;
  onChange: (patch: BusinessImageValue) => void;
  value: BusinessImageValue;
}) {
  const text = getBarcodeText(language);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const [videoUploadError, setVideoUploadError] = useState("");

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
        onChange({ imageuri: result.data.url });
      } else {
        const images = value.images ?? [];
        onChange({
          images: [
            ...images,
            { xorder: images.length + 1, uri: result.data.url },
          ],
        });
      }
      setUploading(false);
    },
    [auth, language, onChange, value.images],
  );

  const handleVideoUpload = useCallback(
    async (file: File | null) => {
      if (!file) return;
      if (file.type !== "video/mp4" || !file.name.toLowerCase().endsWith(".mp4")) {
        setVideoUploadError(language === "th" ? "รองรับเฉพาะไฟล์วิดีโอ MP4" : "Only MP4 videos are supported.");
        return;
      }
      if (file.size > PRODUCT_VIDEO_MAX_BYTES) {
        setVideoUploadError(language === "th" ? `วิดีโอต้องมีขนาดไม่เกิน ${PRODUCT_VIDEO_MAX_MB} MB` : `Video size must not exceed ${PRODUCT_VIDEO_MAX_MB} MB.`);
        return;
      }

      setUploading(true);
      setVideoUploadError("");
      try {
        const posterFile = await createVideoPosterFile(file);
        const posterResult = await uploadProductImage(auth, posterFile);
        if (!posterResult.success || !posterResult.data?.url) {
          setVideoUploadError(posterResult.message ?? "Unable to upload video preview");
          return;
        }

        const result = await uploadProductVideo(auth, file);
        if (!result.success || !result.data?.url) {
          setVideoUploadError(result.message ?? "Upload failed");
          return;
        }
        const videos = value.videos ?? [];
        onChange({
          videos: [
            ...videos,
            {
              xorder: videos.length + 1,
              uri: result.data.url,
              posteruri: posterResult.data.url,
            },
          ],
        });
      } catch {
        setVideoUploadError(
          language === "th"
            ? "สร้างภาพตัวอย่างจากวิดีโอไม่สำเร็จ กรุณาใช้ไฟล์ MP4 (H.264)"
            : "Unable to create a video preview. Please use an MP4 (H.264) file.",
        );
      } finally {
        setUploading(false);
      }
    },
    [auth, language, onChange, value.videos],
  );

  return (
    <div className="space-y-3">
      <section className="rounded-xl border border-border p-3">
        <h3 className="mb-3 text-sm font-semibold">{mainTitle ?? text.mediaMainSection}</h3>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
          <ImagePreview
            auth={auth}
            emptyText={text.mediaNoImage}
            uri={value.imageuri ?? ""}
          />
          <div className="min-w-0 flex-1 space-y-2">
            <Input
              placeholder="https://… หรือ อัปโหลดไฟล์"
              value={value.imageuri ?? ""}
              onChange={(event) => onChange({ imageuri: event.target.value })}
            />
            <div className="flex flex-wrap items-center gap-2">
              <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input bg-background px-3 py-1.5 text-xs hover:bg-muted">
                <Upload className="h-3.5 w-3.5" />
                {uploading ? text.mediaUploading : text.mediaUploadBtn}
                <input
                  accept="image/png,image/jpeg"
                  className="hidden"
                  disabled={uploading}
                  type="file"
                  onChange={(event) => {
                    const file = event.currentTarget.files?.[0] ?? null;
                    event.currentTarget.value = "";
                    void handleUpload(file, "main");
                  }}
                />
              </label>
              {value.imageuri ? (
                <Button
                  onClick={() => onChange({ imageuri: "" })}
                  size="sm"
                  type="button"
                  variant="ghost"
                >
                  <Trash2 className="mr-1 h-3.5 w-3.5" />
                  {text.mediaDeleteBtn}
                </Button>
              ) : null}
            </div>
            {uploadError ? (
              <p className="text-xs text-destructive">{uploadError}</p>
            ) : null}
          </div>
        </div>
      </section>

      <section className="rounded-xl border border-border p-3">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <h3 className="text-sm font-semibold">{galleryTitle ?? text.mediaGallerySection}</h3>
          <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input bg-background px-3 py-1.5 text-xs hover:bg-muted">
            <Plus className="h-3.5 w-3.5" />
            {text.mediaAddGallery}
            <input
              accept="image/png,image/jpeg"
              className="hidden"
              disabled={uploading}
              type="file"
              onChange={(event) => {
                const file = event.currentTarget.files?.[0] ?? null;
                event.currentTarget.value = "";
                void handleUpload(file, "gallery");
              }}
            />
          </label>
        </div>
        {!value.images?.length ? (
          <p className="text-sm text-muted-foreground">{text.mediaNoGallery}</p>
        ) : (
          <ul className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6">
            {value.images.map((image, index) => (
              <li
                className="relative min-w-0 overflow-hidden rounded-md border border-border bg-background"
                key={`${image.uri}-${index}`}
              >
                <AuthenticatedImg
                  alt={`#${index + 1}`}
                  auth={auth}
                  className="aspect-square w-full object-cover"
                  src={image.uri}
                />
                <Button
                  aria-label={text.mediaDeleteBtn}
                  className="absolute right-1 top-1 size-6 bg-background/80"
                  onClick={() =>
                    onChange({
                      images: (value.images ?? [])
                        .filter((_, imageIndex) => imageIndex !== index)
                        .map((item, imageIndex) => ({
                          ...item,
                          xorder: imageIndex + 1,
                        })),
                    })
                  }
                  size="icon"
                  type="button"
                  variant="ghost"
                >
                  <X className="h-3 w-3" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="rounded-xl border border-border p-3">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <div>
            <h3 className="flex items-center gap-2 text-sm font-semibold">
              <Film className="h-4 w-4" />
              {language === "th" ? "วิดีโอ" : "Videos"}
            </h3>
            <p className="mt-1 text-xs text-muted-foreground">
              {language === "th" ? `MP4 สูงสุด ${PRODUCT_VIDEO_MAX_MB} MB แนะนำ H.264/AAC` : `MP4 up to ${PRODUCT_VIDEO_MAX_MB} MB; H.264/AAC recommended`}
            </p>
          </div>
          <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border border-input bg-background px-3 py-1.5 text-xs hover:bg-muted">
            <Plus className="h-3.5 w-3.5" />
            {language === "th" ? "เพิ่มวิดีโอ" : "Add video"}
            <input
              accept="video/mp4,.mp4"
              className="hidden"
              disabled={uploading}
              type="file"
              onChange={(event) => {
                const file = event.currentTarget.files?.[0] ?? null;
                event.currentTarget.value = "";
                void handleVideoUpload(file);
              }}
            />
          </label>
        </div>
        {!value.videos?.length ? (
          <p className="text-sm text-muted-foreground">
            {language === "th" ? "— ยังไม่มีวิดีโอ —" : "— No videos —"}
          </p>
        ) : (
          <ul className="grid gap-3 md:grid-cols-2">
            {value.videos.map((video, index) => (
              <li className="relative min-w-0 overflow-hidden rounded-md border border-border bg-background" key={`${video.uri}-${index}`}>
                <AuthenticatedVideo
                  auth={auth}
                  className="aspect-video w-full"
                  failedLabel={language === "th" ? "เปิดวิดีโอไม่สำเร็จ" : "Unable to open video"}
                  loadLabel={language === "th" ? "โหลดและเล่นวิดีโอ" : "Load and play video"}
                  posterSrc={video.posteruri}
                  src={video.uri}
                />
                <Button
                  aria-label={language === "th" ? "ลบวิดีโอ" : "Delete video"}
                  className="absolute right-1 top-1 size-7 bg-background/90"
                  onClick={() =>
                    onChange({
                      videos: (value.videos ?? [])
                        .filter((_, videoIndex) => videoIndex !== index)
                        .map((item, videoIndex) => ({ ...item, xorder: videoIndex + 1 })),
                    })
                  }
                  size="icon"
                  type="button"
                  variant="ghost"
                >
                  <X className="h-3 w-3" />
                </Button>
              </li>
            ))}
          </ul>
        )}
        {videoUploadError ? <p className="mt-2 text-xs text-destructive">{videoUploadError}</p> : null}
      </section>
    </div>
  );
}

export function BusinessImageGallery({
  auth,
  emptyText,
  sources,
  title,
}: {
  auth: AuthSession | null;
  emptyText?: string;
  sources: BusinessImageSource[];
  title: string;
}) {
  const visibleSources = useMemo(
    () =>
      sources
        .map((source) => ({ ...source, items: imageItems(source), videoItems: videoItems(source) }))
        .filter((source) => source.items.length > 0 || source.videoItems.length > 0),
    [sources],
  );

  if (visibleSources.length === 0 && !emptyText) return null;

  return (
    <section className="rounded-xl border border-border bg-background p-3">
      <h3 className="mb-3 flex items-center gap-2 text-sm font-semibold">
        <ImagePlus className="h-4 w-4" />
        {title}
      </h3>
      {visibleSources.length === 0 ? (
        <p className="text-sm text-muted-foreground">{emptyText}</p>
      ) : (
        <div className="space-y-3">
          {visibleSources.map((source) => (
            <div className="space-y-2" key={source.key}>
              <p className="break-words text-xs font-semibold text-muted-foreground">
                {source.label}
              </p>
              {source.items.length ? (
                <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6">
                  {source.items.map((item, index) => (
                  <figure
                    className="min-w-0 overflow-hidden rounded-lg border border-border bg-muted/20"
                    key={`${source.key}-${item.uri}-${index}`}
                  >
                    <AuthenticatedImg
                      alt={`${source.label} ${index + 1}`}
                      auth={auth}
                      className="aspect-square w-full object-cover"
                      src={item.uri}
                    />
                  </figure>
                  ))}
                </div>
              ) : null}
              {source.videoItems.length ? (
                <div className="grid gap-2 md:grid-cols-2">
                  {source.videoItems.map((item, index) => (
                    <figure className="min-w-0 overflow-hidden rounded-lg border border-border bg-muted/20" key={`${source.key}-video-${item.uri}-${index}`}>
                      <AuthenticatedVideo
                        auth={auth}
                        className="aspect-video w-full"
                        failedLabel="เปิดวิดีโอไม่สำเร็จ / Unable to open video"
                        loadLabel="โหลดและเล่นวิดีโอ / Load and play"
                        posterSrc={item.posteruri}
                        src={item.uri}
                      />
                    </figure>
                  ))}
                </div>
              ) : null}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function ImagePreview({
  auth,
  emptyText,
  uri,
}: {
  auth: AuthSession | null;
  emptyText: string;
  uri: string;
}) {
  return (
    <div className="size-32 shrink-0 overflow-hidden rounded-lg border border-border bg-muted">
      {uri ? (
        <AuthenticatedImg
          alt=""
          auth={auth}
          className="size-full object-cover"
          fallback={
            <div className="flex size-full items-center justify-center p-2 text-center text-xs text-muted-foreground">
              {emptyText}
            </div>
          }
          src={uri}
        />
      ) : (
        <div className="flex size-full items-center justify-center p-2 text-center text-xs text-muted-foreground">
          {emptyText}
        </div>
      )}
    </div>
  );
}

function imageItems(value: BusinessImageValue): ProductImage[] {
  const result: ProductImage[] = [];
  const seen = new Set<string>();
  const add = (uri: string, xorder: number) => {
    const normalized = uri.trim();
    if (!normalized || seen.has(normalized)) return;
    seen.add(normalized);
    result.push({ uri: normalized, xorder });
  };

  add(value.imageuri ?? "", 0);
  for (const image of value.images ?? []) add(image.uri ?? "", image.xorder);
  return result;
}

function videoItems(value: BusinessImageValue): ProductVideo[] {
  const seen = new Set<string>();
  return (value.videos ?? [])
    .map((video, index) => ({
      uri: String(video.uri ?? "").trim(),
      posteruri: String(video.posteruri ?? "").trim(),
      xorder: video.xorder || index + 1,
    }))
    .filter((video) => {
      if (!video.uri || seen.has(video.uri)) return false;
      seen.add(video.uri);
      return true;
    });
}

async function createVideoPosterFile(file: File): Promise<File> {
  const objectUrl = URL.createObjectURL(file);
  const video = document.createElement("video");
  video.muted = true;
  video.playsInline = true;
  video.preload = "auto";

  try {
    await new Promise<void>((resolve, reject) => {
      const cleanup = () => {
        video.removeEventListener("loadeddata", handleLoaded);
        video.removeEventListener("error", handleError);
      };
      const handleLoaded = () => {
        cleanup();
        resolve();
      };
      const handleError = () => {
        cleanup();
        reject(new Error("Video frame unavailable"));
      };
      video.addEventListener("loadeddata", handleLoaded);
      video.addEventListener("error", handleError);
      video.src = objectUrl;
      video.load();
    });

    if (!video.videoWidth || !video.videoHeight) throw new Error("Video dimensions unavailable");
    const canvas = document.createElement("canvas");
    const scale = Math.min(1, 960 / video.videoWidth);
    canvas.width = Math.max(1, Math.round(video.videoWidth * scale));
    canvas.height = Math.max(1, Math.round(video.videoHeight * scale));
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Canvas unavailable");
    context.drawImage(video, 0, 0, canvas.width, canvas.height);
    const blob = await new Promise<Blob | null>((resolve) =>
      canvas.toBlob(resolve, "image/jpeg", 0.82),
    );
    if (!blob) throw new Error("Video poster unavailable");
    const baseName = file.name.replace(/\.[^.]+$/, "") || "video";
    return new File([blob], `${baseName}-poster.jpg`, {
      type: "image/jpeg",
      lastModified: Date.now(),
    });
  } finally {
    video.removeAttribute("src");
    video.load();
    URL.revokeObjectURL(objectUrl);
  }
}
