import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  isRecord,
  readJsonOrText,
  requireBearerToken,
} from "@/lib/workspace-api";

type ImageUploadProxyOptions = {
  backendPath?: string;
  category: string;
  forwardRequestBody?: boolean;
  kind?: "image" | "video";
  maxRequestBytes?: number;
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

  let uploadBody: BodyInit;
  let uploadContentType = "";
  let uploadExceededLimit = false;
  if (options.forwardRequestBody) {
    const contentLength = Number(request.headers.get("content-length") ?? 0);
    if (options.maxRequestBytes && contentLength > options.maxRequestBytes) {
      return NextResponse.json(
        { success: false, message: "ไฟล์มีขนาดใหญ่เกินกำหนด" },
        { status: 413 },
      );
    }
    uploadContentType = request.headers.get("content-type") ?? "";
    if (!request.body || !uploadContentType.toLowerCase().startsWith("multipart/form-data;")) {
      return NextResponse.json(
        { success: false, message: options.kind === "video" ? "ไม่พบไฟล์วิดีโอ" : "ไม่พบไฟล์รูป" },
        { status: 400 },
      );
    }
    if (options.maxRequestBytes) {
      let receivedBytes = 0;
      uploadBody = request.body.pipeThrough(
        new TransformStream<Uint8Array, Uint8Array>({
          transform(chunk, controller) {
            receivedBytes += chunk.byteLength;
            if (receivedBytes > options.maxRequestBytes!) {
              uploadExceededLimit = true;
              controller.error(new Error("upload body exceeds limit"));
              return;
            }
            controller.enqueue(chunk);
          },
        }),
      );
    } else {
      uploadBody = request.body;
    }
  } else {
    const form = await request.formData().catch(() => null);
    const file = form?.get("file");
    if (!form || !(file instanceof File)) {
      return NextResponse.json(
        { success: false, message: options.kind === "video" ? "ไม่พบไฟล์วิดีโอ" : "ไม่พบไฟล์รูป" },
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
    uploadForm.append("file", file, file.name || (options.kind === "video" ? "video.mp4" : "image.jpg"));
    uploadForm.append("category", category);

    for (const field of forwardedTextFields) {
      const value = form.get(field);
      if (typeof value === "string" && value.trim()) {
        uploadForm.append(field, value.trim().slice(0, 500));
      }
    }
    uploadBody = uploadForm;
  }

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), options.timeoutMs ?? 60000);

  try {
    const headers: Record<string, string> = {
      "Accept-Language": request.headers.get("accept-language") ?? "th",
      Authorization: authorization,
    };
    if (uploadContentType) headers["Content-Type"] = uploadContentType;
    const uploadRequest: RequestInit & { duplex?: "half" } = {
      method: "POST",
      headers,
      body: uploadBody,
      signal: controller.signal,
      cache: "no-store",
    };
    if (options.forwardRequestBody) uploadRequest.duplex = "half";
    const response = await fetch(
      `${mainApiUrl}${options.backendPath ?? "/goapi/image/upload"}`,
      uploadRequest,
    );
    const payload = await readJsonOrText(response);
    return NextResponse.json(normalizeImageUploadPayload(payload, response.ok), {
      status: response.status,
    });
  } catch (error) {
    if (uploadExceededLimit) {
      return NextResponse.json(
        { success: false, message: "ไฟล์มีขนาดใหญ่เกินกำหนด" },
        { status: 413 },
      );
    }
    const message =
      error instanceof Error && error.name === "AbortError"
        ? "Server ไม่ตอบกลับทันเวลา"
        : options.kind === "video" ? "อัปโหลดวิดีโอไม่สำเร็จ" : "อัปโหลดรูปไม่สำเร็จ";
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
        data.file_url ??
        payload.fileurl ??
        data.fileurl,
    ),
  );
  const objectKey = stringValue(payload.objectkey ?? data.objectkey) || imageObjectKey(data);
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

export function imageNeedsAuthenticatedFetch(imageUrl: string, trustedBackendUrl = ""): boolean {
  if (!imageUrl || /^(blob:|data:)/i.test(imageUrl)) return false;
  const fallbackOrigin =
    typeof window === "undefined"
      ? "http://localhost"
      : window.location.origin;
  try {
    const parsed = new URL(imageUrl, fallbackOrigin);
    if (!parsed.pathname.includes("/s3/file/")) return false;
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return false;
    const frontendOrigin = new URL(fallbackOrigin).origin;
    const trustedOrigin = uploadOrigin(trustedBackendUrl, fallbackOrigin);
    return parsed.origin === frontendOrigin || parsed.origin === trustedOrigin;
  } catch {
    return false;
  }
}

function uploadOrigin(value: string, fallbackOrigin: string): string {
  const normalized = value.trim();
  if (!normalized) return fallbackOrigin;
  try {
    return new URL(/^https?:\/\//i.test(normalized) ? normalized : `http://${normalized}`).origin;
  } catch {
    return fallbackOrigin;
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
    .map((part) => part.replace(/[^a-z0-9_~-]/gi, "").slice(0, 64))
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
