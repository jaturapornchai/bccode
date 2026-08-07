"use client";

import type { ImgHTMLAttributes, ReactNode, VideoHTMLAttributes } from "react";
import { useEffect, useMemo, useState } from "react";
import { imageNeedsAuthenticatedFetch } from "@/lib/image-upload-proxy";
import type { AuthSession } from "@/lib/workspace-models";

// ─── Authenticated image display (product/barcode/logo/gallery images) ────
//
// Protected GoAPI/S3 paths (see imageNeedsAuthenticatedFetch) require a
// bearer token, so a plain <img src=...> always renders broken for them. This
// hook fetches the image with the token and exposes a short-lived object URL
// instead. Object URLs are cached (LRU) per backend+user+token so repeated
// renders of the same image don't re-fetch. Moved here from
// system-settings-screen.tsx (where it was already used for logo/gallery
// fields) so product/barcode screens can reuse the same mechanism instead of
// building a parallel one. Lives in this "use client" component file (rather
// than the shared lib/image-upload-proxy.ts) because that lib file is also
// imported by server-only Route Handlers — mixing React hooks into it breaks
// the Next.js server/client module boundary.

type AuthenticatedImageCacheEntry = {
  objectUrl: string;
  lastAccessAt: number;
};

const AUTHENTICATED_IMAGE_CACHE_MAX_ENTRIES = 80;
const authenticatedImageObjectUrlCache = new Map<string, AuthenticatedImageCacheEntry>();
let authenticatedImageCacheOwner = "";

function authenticatedImageCacheKey(
  imageUrl: string,
  backendUrl: string,
  username: string,
  token: string,
): string {
  return [backendUrl.trim(), username.trim().toLowerCase(), token, imageUrl].join(
    String.fromCharCode(0),
  );
}

function syncAuthenticatedImageCacheOwner(backendUrl: string, username: string, token: string) {
  const owner = [backendUrl.trim(), username.trim().toLowerCase(), token].join(
    String.fromCharCode(0),
  );
  if (authenticatedImageCacheOwner && authenticatedImageCacheOwner !== owner) {
    clearAuthenticatedImageObjectUrlCache();
  }
  authenticatedImageCacheOwner = owner;
}

function getCachedAuthenticatedImageObjectUrl(cacheKey: string): string {
  const cached = authenticatedImageObjectUrlCache.get(cacheKey);
  if (!cached) return "";
  cached.lastAccessAt = Date.now();
  return cached.objectUrl;
}

function cacheAuthenticatedImageObjectUrl(cacheKey: string, objectUrl: string) {
  const existing = authenticatedImageObjectUrlCache.get(cacheKey);
  if (existing?.objectUrl && existing.objectUrl !== objectUrl) {
    URL.revokeObjectURL(existing.objectUrl);
  }
  authenticatedImageObjectUrlCache.set(cacheKey, { objectUrl, lastAccessAt: Date.now() });
  if (authenticatedImageObjectUrlCache.size <= AUTHENTICATED_IMAGE_CACHE_MAX_ENTRIES) return;
  const entries = [...authenticatedImageObjectUrlCache.entries()].sort(
    (first, second) => first[1].lastAccessAt - second[1].lastAccessAt,
  );
  for (const [key, entry] of entries.slice(
    0,
    authenticatedImageObjectUrlCache.size - AUTHENTICATED_IMAGE_CACHE_MAX_ENTRIES,
  )) {
    URL.revokeObjectURL(entry.objectUrl);
    authenticatedImageObjectUrlCache.delete(key);
  }
}

function clearAuthenticatedImageObjectUrlCache() {
  for (const entry of authenticatedImageObjectUrlCache.values()) {
    URL.revokeObjectURL(entry.objectUrl);
  }
  authenticatedImageObjectUrlCache.clear();
  authenticatedImageCacheOwner = "";
}

function mainApiDisplayBase(rawBackendUrl: unknown): string {
  const raw = stringValue(rawBackendUrl).trim();
  if (!raw) return "";
  try {
    const withProtocol = /^https?:\/\//i.test(raw) ? raw : `http://${raw}`;
    const parsed = new URL(withProtocol);
    const path = parsed.pathname.replace(/\/+$/, "");
    parsed.pathname = path.toLowerCase().endsWith("/goapi")
      ? path.slice(0, -"/goapi".length) || "/"
      : "/";
    parsed.search = "";
    parsed.hash = "";
    return parsed.toString().replace(/\/$/, "");
  } catch {
    return "";
  }
}

function imageDisplayUrl(value: unknown, backendUrl: unknown): string {
  const raw = stringValue(value).trim();
  if (!raw) return "";
  if (/^(blob:|data:|https?:\/\/)/i.test(raw)) return raw;
  if (raw.startsWith("//")) {
    return typeof window === "undefined" ? raw : `${window.location.protocol}${raw}`;
  }
  if (raw.startsWith("/api/")) return raw;

  const base = mainApiDisplayBase(backendUrl);
  if (!base) return raw;
  if (raw.startsWith("/")) return `${base}${raw}`;
  if (raw.toLowerCase().startsWith("images/")) return `${base}/${raw.replace(/^\/+/, "")}`;
  return `${base}/images/${raw.replace(/^\/+/, "")}`;
}

function stringValue(value: unknown): string {
  return typeof value === "string"
    ? value.trim()
    : value === null || value === undefined
      ? ""
      : String(value).trim();
}

/**
 * useAuthenticatedImageDisplaySource — resolves any product/barcode/logo/
 * gallery image value (relative GoAPI path or absolute URL) to a displayable
 * URL. Protected `/s3/file/` paths are fetched with the bearer token and
 * turned into a cached, short-lived object URL; public paths pass through
 * unchanged (no fetch).
 */
export function useAuthenticatedImageDisplaySource(
  value: unknown,
  auth: AuthSession | null,
): {
  displayUrl: string;
  failed: boolean;
  loading: boolean;
  requestedUrl: string;
} {
  const authBackendUrl = auth?.backendUrl ?? "";
  const authToken = auth?.token ?? "";
  const authUsername = auth?.username ?? "";
  const requestedUrl = useMemo(
    () => imageDisplayUrl(value, authBackendUrl),
    [authBackendUrl, value],
  );
  const [state, setState] = useState({
    displayUrl: "",
    failed: false,
    loading: false,
  });

  useEffect(() => {
    if (!requestedUrl) {
      setState({ displayUrl: "", failed: false, loading: false });
      return;
    }

    if (!imageNeedsAuthenticatedFetch(requestedUrl, authBackendUrl)) {
      setState({ displayUrl: requestedUrl, failed: false, loading: false });
      return;
    }

    if (!authToken) {
      clearAuthenticatedImageObjectUrlCache();
      setState({ displayUrl: "", failed: true, loading: false });
      return;
    }

    syncAuthenticatedImageCacheOwner(authBackendUrl, authUsername, authToken);
    const cacheKey = authenticatedImageCacheKey(
      requestedUrl,
      authBackendUrl,
      authUsername,
      authToken,
    );
    const cachedObjectUrl = getCachedAuthenticatedImageObjectUrl(cacheKey);
    if (cachedObjectUrl) {
      setState({ displayUrl: cachedObjectUrl, failed: false, loading: false });
      return;
    }

    const controller = new AbortController();
    let cancelled = false;

    setState({ displayUrl: "", failed: false, loading: true });
    void fetch(requestedUrl, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${authToken}` },
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.blob();
      })
      .then((blob) => {
        const nextObjectUrl = URL.createObjectURL(blob);
        if (cancelled) {
          URL.revokeObjectURL(nextObjectUrl);
          return;
        }
        cacheAuthenticatedImageObjectUrl(cacheKey, nextObjectUrl);
        setState({
          displayUrl: nextObjectUrl,
          failed: false,
          loading: false,
        });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        if (error instanceof DOMException && error.name === "AbortError")
          return;
        setState({ displayUrl: "", failed: true, loading: false });
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [authBackendUrl, authToken, authUsername, requestedUrl]);

  return { ...state, requestedUrl };
}

type AuthenticatedImgProps = Omit<ImgHTMLAttributes<HTMLImageElement>, "src"> & {
  src: string | null | undefined;
  auth: AuthSession | null;
  /** Shown when there is nothing displayable (no src, or the fetch failed). */
  fallback?: ReactNode;
};

/**
 * AuthenticatedImg — drop-in replacement for a plain <img> when `src` may be
 * a protected GoAPI/S3 path (`/goapi/s3/file/...`) that requires a bearer
 * token, which a browser cannot attach to a plain <img> request. Resolves the
 * image via useAuthenticatedImageDisplaySource above (auth fetch + cached
 * object URL) and renders the resulting blob URL; public URLs pass through
 * unchanged.
 */
export function AuthenticatedImg({
  src,
  auth,
  fallback = null,
  alt = "",
  className,
  ...imgProps
}: AuthenticatedImgProps) {
  const { displayUrl, loading } = useAuthenticatedImageDisplaySource(src ?? "", auth);
  if (displayUrl) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={displayUrl} alt={alt} className={className} {...imgProps} />;
  }
  if (loading) {
    return <div className={`${className ?? ""} animate-pulse bg-muted`.trim()} aria-hidden="true" />;
  }
  return <>{fallback}</>;
}

type AuthenticatedVideoProps = Omit<VideoHTMLAttributes<HTMLVideoElement>, "src"> & {
  src: string | null | undefined;
  posterSrc?: string | null | undefined;
  auth: AuthSession | null;
  loadLabel: string;
  failedLabel: string;
};

/** Loads a private video only when requested and releases its blob after unmount. */
export function AuthenticatedVideo({
  src,
  posterSrc,
  auth,
  loadLabel,
  failedLabel,
  className,
  ...videoProps
}: AuthenticatedVideoProps) {
  const [requested, setRequested] = useState(false);
  const [state, setState] = useState({ displayUrl: "", failed: false, loading: false });
  const requestedUrl = useMemo(
    () => imageDisplayUrl(src, auth?.backendUrl),
    [auth?.backendUrl, src],
  );
  const poster = useAuthenticatedImageDisplaySource(posterSrc ?? "", auth);

  useEffect(() => {
    setRequested(false);
    setState({ displayUrl: "", failed: false, loading: false });
  }, [requestedUrl]);

  useEffect(() => {
    if (!requested || !requestedUrl) return;
    if (!imageNeedsAuthenticatedFetch(requestedUrl, auth?.backendUrl ?? "")) {
      setState({ displayUrl: requestedUrl, failed: false, loading: false });
      return;
    }
    if (!auth?.token) {
      setState({ displayUrl: "", failed: true, loading: false });
      return;
    }

    const controller = new AbortController();
    let cancelled = false;
    let objectUrl = "";
    setState({ displayUrl: "", failed: false, loading: true });
    void fetch(requestedUrl, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${auth.token}` },
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.blob();
      })
      .then((blob) => {
        objectUrl = URL.createObjectURL(blob);
        if (cancelled) {
          URL.revokeObjectURL(objectUrl);
          objectUrl = "";
          return;
        }
        setState({ displayUrl: objectUrl, failed: false, loading: false });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        if (error instanceof DOMException && error.name === "AbortError") return;
        setState({ displayUrl: "", failed: true, loading: false });
      });

    return () => {
      cancelled = true;
      controller.abort();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [auth?.backendUrl, auth?.token, requested, requestedUrl]);

  if (!requested) {
    return (
      <button
        aria-label={loadLabel}
        className={`${className ?? ""} group relative flex items-center justify-center overflow-hidden bg-muted px-3 text-sm font-medium hover:bg-muted/80`.trim()}
        onClick={() => setRequested(true)}
        type="button"
      >
        {poster.displayUrl ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            alt=""
            className="absolute inset-0 h-full w-full object-cover"
            data-testid="video-poster"
            src={poster.displayUrl}
          />
        ) : poster.loading ? (
          <span className="absolute inset-0 animate-pulse bg-muted" aria-hidden="true" />
        ) : (
          <span className="relative z-10 px-3 text-center">{loadLabel}</span>
        )}
        {poster.displayUrl ? (
          <span
            aria-hidden="true"
            className="relative z-10 grid size-12 place-items-center rounded-full bg-black/65 text-xl text-white shadow-lg transition-transform group-hover:scale-105"
          >
            ▶
          </span>
        ) : null}
      </button>
    );
  }
  if (state.displayUrl) {
    return (
      <video
        className={className}
        controls
        playsInline
        poster={poster.displayUrl || undefined}
        preload="metadata"
        src={state.displayUrl}
        {...videoProps}
        onError={() => setState({ displayUrl: "", failed: true, loading: false })}
      />
    );
  }
  return (
    <button
      className={`${className ?? ""} flex items-center justify-center bg-muted p-3 text-center text-xs text-muted-foreground`.trim()}
      disabled={state.loading}
      onClick={() => setRequested(false)}
      type="button"
    >
      {state.loading ? loadLabel : failedLabel}
    </button>
  );
}
