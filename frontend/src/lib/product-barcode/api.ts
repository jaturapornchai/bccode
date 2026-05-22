/**
 * Client-side API for Product Barcode feature.
 *
 * Conventions:
 * - Every call goes through a Next.js API route (`/api/...`) — never hit backend directly from the browser.
 * - Auth context (Bearer token + backend URL) is read from session storage helpers in workspace-models.
 * - Returns plain objects with `success` boolean.
 */

import type { AuthSession } from "@/lib/workspace-models";
import type {
  NameX,
  ProductBarcode,
  ProductBarcodeListRequest,
  ProductBarcodeListResponse,
  ProductBarcodeListRow,
} from "./types";

/** Standard JSON response envelope. */
export type ApiEnvelope<T = unknown> = {
  success: boolean;
  message?: string;
  data?: T;
  total?: number;
};

function authHeaders(auth: AuthSession | null): Record<string, string> {
  return {
    "Content-Type": "application/json",
    ...(auth ? { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl } : {}),
  };
}

async function jsonRequest<T = unknown>(input: RequestInfo | URL, init: RequestInit): Promise<ApiEnvelope<T>> {
  try {
    const response = await fetch(input, init);
    const payload = (await response.json().catch(() => ({}))) as ApiEnvelope<T>;
    if (!response.ok && payload.success !== false) {
      return { ...payload, success: false, message: payload.message ?? `HTTP ${response.status}` };
    }
    return payload;
  } catch (error) {
    return { success: false, message: error instanceof Error ? error.message : "Network error" };
  }
}

/** List barcodes via PG endpoint (filters + sort + offset). */
export function listBarcodes(
  auth: AuthSession | null,
  body: ProductBarcodeListRequest,
): Promise<ProductBarcodeListResponse> {
  return jsonRequest<ProductBarcodeListRow[]>("/api/product-barcode/list", {
    method: "POST",
    headers: authHeaders(auth),
    body: JSON.stringify({ ...body, backendUrl: auth?.backendUrl }),
  }).then((envelope) => ({
    success: envelope.success,
    message: envelope.message,
    data: envelope.data,
    total: envelope.total,
  }));
}

/** Get a single barcode by GUID. */
export function getBarcode(auth: AuthSession | null, guid: string): Promise<ApiEnvelope<unknown>> {
  return jsonRequest("/api/product-barcode/" + encodeURIComponent(guid), {
    method: "GET",
    headers: authHeaders(auth),
  });
}

/** Create a new barcode. */
export function createBarcode(auth: AuthSession | null, data: ProductBarcode): Promise<ApiEnvelope<unknown>> {
  return jsonRequest("/api/product-barcode", {
    method: "POST",
    headers: authHeaders(auth),
    body: JSON.stringify({ backendUrl: auth?.backendUrl, data }),
  });
}

/** Update an existing barcode by GUID. */
export function updateBarcode(
  auth: AuthSession | null,
  guid: string,
  data: ProductBarcode,
): Promise<ApiEnvelope<unknown>> {
  return jsonRequest("/api/product-barcode/" + encodeURIComponent(guid), {
    method: "PUT",
    headers: authHeaders(auth),
    body: JSON.stringify({ backendUrl: auth?.backendUrl, data }),
  });
}

/** Delete by GUID list (bulk) or single. */
export function deleteBarcodes(auth: AuthSession | null, guids: string[]): Promise<ApiEnvelope<unknown>> {
  if (guids.length === 1) {
    return jsonRequest("/api/product-barcode/" + encodeURIComponent(guids[0]), {
      method: "DELETE",
      headers: authHeaders(auth),
    });
  }
  return jsonRequest("/api/product-barcode", {
    method: "DELETE",
    headers: authHeaders(auth),
    body: JSON.stringify({ backendUrl: auth?.backendUrl, guids }),
  });
}

/** Get BOM tree by barcode (read-only view). */
export function getBarcodeBom(auth: AuthSession | null, barcode: string): Promise<ApiEnvelope<unknown>> {
  return jsonRequest("/api/product-barcode/bom/" + encodeURIComponent(barcode), {
    method: "GET",
    headers: authHeaders(auth),
  });
}

/** Get price history for a barcode. */
export function getPriceHistory(
  auth: AuthSession | null,
  barcode: string,
  page = 1,
  limit = 20,
): Promise<ApiEnvelope<unknown>> {
  const qs = new URLSearchParams({ page: String(page), limit: String(limit) }).toString();
  return jsonRequest("/api/product-barcode/price-history/" + encodeURIComponent(barcode) + "?" + qs, {
    method: "GET",
    headers: authHeaders(auth),
  });
}

/** Master data entry returned by picker endpoints. */
export type MasterEntry = {
  guidfixed: string;
  code: string;
  names: NameX[];
};

/**
 * Master picker — fetch a paginated list of master records.
 * Backend path is selected by `master` name. All hit the same Next.js proxy
 * (`/api/product-barcode/master/[master]`) which translates to the mainapi.
 */
export type MasterName =
  | "group"
  | "groupsubone"
  | "groupsubtwo"
  | "brand"
  | "category"
  | "class"
  | "design"
  | "grade"
  | "model"
  | "pattern"
  | "unit"
  | "producttype"
  | "ordertype"
  | "businesstype"
  | "branch";

export interface MasterListRequest {
  q?: string;
  page?: number;
  limit?: number;
  lang?: string;
}

export function listMaster(
  auth: AuthSession | null,
  master: MasterName,
  request: MasterListRequest = {},
): Promise<ApiEnvelope<MasterEntry[]>> {
  const qs = new URLSearchParams();
  if (request.q) qs.set("q", request.q);
  if (request.page) qs.set("page", String(request.page));
  if (request.limit) qs.set("limit", String(request.limit));
  if (request.lang) qs.set("lang", request.lang);
  const queryString = qs.toString();
  const url = `/api/product-barcode/master/${master}${queryString ? `?${queryString}` : ""}`;
  return jsonRequest<MasterEntry[]>(url, {
    method: "GET",
    headers: authHeaders(auth),
  });
}

/** Upload main product image. Returns absolute URL (or relative as backend prefers). */
export function uploadProductImage(
  auth: AuthSession | null,
  file: File,
): Promise<ApiEnvelope<{ url: string; key?: string }>> {
  const form = new FormData();
  form.append("file", file);
  form.append("category", "products");
  return fetch("/api/product-barcode/image", {
    method: "POST",
    headers: {
      ...(auth ? { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl } : {}),
    },
    body: form,
  })
    .then(async (res) => {
      const payload = (await res.json().catch(() => ({}))) as ApiEnvelope<{ url: string; key?: string }>;
      if (!res.ok && payload.success !== false) {
        return { ...payload, success: false, message: payload.message ?? `HTTP ${res.status}` };
      }
      return payload;
    })
    .catch((error) => ({ success: false, message: error instanceof Error ? error.message : "Upload failed" }));
}
