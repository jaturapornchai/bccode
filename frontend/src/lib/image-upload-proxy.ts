import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  readJsonOrText,
  requireBearerToken,
} from "@/lib/workspace-api";

type ImageUploadProxyOptions = {
  category: string;
  requireClientCategory?: boolean;
  timeoutMs?: number;
};

type ImageUploadPayload = Record<string, unknown>;

const forwardedTextFields = ["description", "tags", "uploadedby"] as const;

export async function proxyImageUploadToGoApi(
  request: Request,
  options: ImageUploadProxyOptions,
): Promise<NextResponse> {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      {
        success: false,
        message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง",
      },
      { status: 400 },
    );
  }

  const form = await request.formData().catch(() => null);
  const file = form?.get("file");
  if (!form || !(file instanceof File)) {
    return NextResponse.json(
      { success: false, message: "ไม่พบไฟล์รูป" },
      { status: 400 },
    );
  }

  const category = resolveUploadCategory(form.get("category"), options);
  if (!category) {
    return NextResponse.json(
      { success: false, message: "ไม่พบหมวดหมู่รูปภาพสำหรับอัปโหลด" },
      { status: 400 },
    );
  }

  const uploadForm = new FormData();
  uploadForm.append("file", file, file.name || "image.jpg");
  uploadForm.append("category", category);

  for (const field of forwardedTextFields) {
    const value = form.get(field);
    if (typeof value === "string" && value.trim()) {
      uploadForm.append(field, value.trim().slice(0, 500));
    }
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), options.timeoutMs ?? 60000);

  try {
    const response = await fetch(`${mainApiUrl}/goapi/image/upload`, {
      method: "POST",
      headers: {
        "Accept-Language": request.headers.get("accept-language") ?? "th",
        Authorization: authorization,
      },
      body: uploadForm,
      signal: controller.signal,
      cache: "no-store",
    });
    const payload = await readJsonOrText(response);
    return NextResponse.json(normalizeImageUploadPayload(payload, response.ok), {
      status: response.status,
    });
  } catch (error) {
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : "อัปโหลดรูปไม่สำเร็จ";
    return NextResponse.json({ success: false, message }, { status: 504 });
  } finally {
    clearTimeout(timeout);
  }
}

export function normalizeImageUploadPayload(
  payload: unknown,
  responseOk = true,
): ImageUploadPayload {
  if (!isRecord(payload)) {
    return {
      success: responseOk,
      message: typeof payload === "string" ? payload : "",
    };
  }

  const data = isRecord(payload.data) ? { ...payload.data } : {};
  const status = stringValue(payload.status).toLowerCase();
  const success =
    payload.success === true ||
    status === "success" ||
    (responseOk && payload.success !== false && status !== "error");

  if (!success) {
    return { ...payload, success: false };
  }

  const existingUrl = normalizeImageProxyUrl(
    stringValue(
      payload.url ??
        data.url ??
        payload.uri ??
        data.uri ??
        payload.file_url ??
        data.file_url,
    ),
  );
  const objectKey = imageObjectKey(data);
  const url = existingUrl || (objectKey ? `/goapi/s3/file/${objectKey}` : "");

  return {
    ...payload,
    success: true,
    url,
    data: {
      ...data,
      key: objectKey,
      url,
    },
  };
}

function imageObjectKey(data: Record<string, unknown>): string {
  const holdingCode = stringValue(data.holdingcode ?? data.holdingcode ?? data.holdingcode);
  const category = sanitizeUploadCategory(data.category);
  const fileName = stringValue(
    data.file_name ?? data.filename ?? data.fileName ?? data.name,
  );
  if (!holdingCode || !fileName) return "";
  return [holdingCode, category, fileName]
    .filter(Boolean)
    .map((part) => encodeURIComponent(part).replace(/%2F/gi, "/"))
    .join("/");
}

function normalizeImageProxyUrl(value: string): string {
  if (!value) return "";
  if (/^https?:\/\//i.test(value) || value.startsWith("data:")) return value;
  if (value.startsWith("/goapi/s3/file/")) return value;
  if (value.startsWith("/s3/file/")) return `/goapi${value}`;
  if (value.startsWith("s3/file/")) return `/goapi/${value}`;
  return value;
}

export function imageNeedsAuthenticatedFetch(imageUrl: string): boolean {
  if (!imageUrl || /^(blob:|data:)/i.test(imageUrl)) return false;
  const fallbackOrigin =
    typeof window === "undefined"
      ? "http://localhost"
      : window.location.origin;
  try {
    const parsed = new URL(imageUrl, fallbackOrigin);
    return parsed.pathname.includes("/s3/file/");
  } catch {
    return imageUrl.includes("/s3/file/");
  }
}

function resolveUploadCategory(
  value: unknown,
  options: ImageUploadProxyOptions,
): string {
  if (options.requireClientCategory) {
    return sanitizeUploadCategory(value);
  }
  return sanitizeUploadCategory(options.category);
}

function sanitizeUploadCategory(value: unknown): string {
  const source = stringValue(value);
  return source
    .replaceAll("\\", "/")
    .split("/")
    .map((part) => part.replace(/[^a-z0-9_-]/gi, "").slice(0, 64))
    .filter(Boolean)
    .join("/");
}

function stringValue(value: unknown): string {
  return typeof value === "string"
    ? value.trim()
    : value === null || value === undefined
      ? ""
      : String(value).trim();
}
