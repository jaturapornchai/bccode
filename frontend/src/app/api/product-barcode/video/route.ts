import { proxyImageUploadToGoApi } from "@/lib/image-upload-proxy";
import {
  PRODUCT_VIDEO_REQUEST_MAX_BYTES,
  PRODUCT_VIDEO_UPLOAD_TIMEOUT_MS,
} from "@/lib/product-barcode/types";

export async function POST(request: Request) {
  return proxyImageUploadToGoApi(request, {
    backendPath: "/goapi/video/upload",
    category: "products/videos",
    forwardRequestBody: true,
    kind: "video",
    maxRequestBytes: PRODUCT_VIDEO_REQUEST_MAX_BYTES,
    timeoutMs: PRODUCT_VIDEO_UPLOAD_TIMEOUT_MS,
  });
}
