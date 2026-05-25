import { proxyImageUploadToGoApi } from "@/lib/image-upload-proxy";

/**
 * Image upload proxy.
 *
 * Frontend sends `multipart/form-data` with field `file`.
 * Backend GoAPI route: `POST /goapi/image/upload`.
 */
export async function POST(request: Request) {
  return proxyImageUploadToGoApi(request, {
    category: "products",
    timeoutMs: 60000,
  });
}
