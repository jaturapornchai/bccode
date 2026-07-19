// Logo thumbnail URL helper.
// Logos are stored as original PNGs on Cloudflare R2/S3 (via GoAPI). Thumbnails are
// produced on demand by Cloudflare Image Resizing using the /cdn-cgi/image path prefix,
// so we only persist the original URI and derive thumbnail URLs on the client.

const CF_IMAGE_PREFIX = "/cdn-cgi/image";

/**
 * Default thumbnail width (in pixels) used across BC Ai Account forms, list rows,
 * and previews. Logos are square-ish brand marks, so we keep a single size and
 * let CSS handle the visible box.
 */
export const LOGO_THUMB_DEFAULT_WIDTH = 160;

/**
 * Build a Cloudflare Image Resizing URL for a logo original.
 *
 * The original URI is what we persist on the record (e.g. `logos/company/BCS001.png`).
 * Cloudflare serves a resized, cached thumbnail at runtime; we never generate or
 * store a separate thumbnail file.
 *
 * Example: `logos/company/BCS001.png` -> `/cdn-cgi/image/width=160,format=auto,quality=85/logos/company/BCS001.png`
 *
 * Public-logo URIs are returned as-is. Protected (GoAPI/R2 authenticated) URIs must go
 * through the existing authenticated preview path and are not resized by CF; callers
 * should fall back to the original URI in that case.
 *
 * @param originalUri  The persisted logo URI/path (must be a public R2/S3 object path).
 * @param width        Target width in pixels. Defaults to LOGO_THUMB_DEFAULT_WIDTH.
 * @returns A Cloudflare-resized thumbnail URL, or the original URI when it cannot be resized.
 */
export function logoThumbUri(
  originalUri: string | null | undefined,
  width: number = LOGO_THUMB_DEFAULT_WIDTH,
): string {
  const uri = (originalUri ?? "").trim();
  if (!uri) return "";
  // Skip remote absolute URLs that already carry their own host/CDN handling.
  if (/^https?:\/\//i.test(uri)) return uri;
  // Skip protected GoAPI/S3 object-stream paths; CF cannot resize authenticated streams.
  if (uri.startsWith("/goapi/") || uri.startsWith("/s3/") || uri.startsWith("/api/")) {
    return uri;
  }
  // Normalize: strip a leading slash so the CF prefix stays clean.
  const path = uri.startsWith("/") ? uri.slice(1) : uri;
  return `${CF_IMAGE_PREFIX}/width=${Math.max(1, Math.round(width))},format=auto,quality=85/${path}`;
}
