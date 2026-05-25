import { proxyImageUploadToGoApi } from "@/lib/image-upload-proxy";

export async function POST(request: Request) {
  return proxyImageUploadToGoApi(request, {
    category: "system-settings",
    requireClientCategory: true,
    timeoutMs: 60000,
  });
}
