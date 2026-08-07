"use client";

import { useEffect, useMemo, useState } from "react";
import { Building2, type LucideIcon } from "lucide-react";
import { logoThumbUri } from "@/lib/logo-thumb";
import { imageNeedsAuthenticatedFetch } from "@/lib/image-upload-proxy";

/**
 * LogoAvatar — shared brand-logo avatar used across BC Ai Account screens.
 *
 * Logos are stored as protected R2 objects (served via `/goapi/s3/file/...`)
 * so we fetch them with the current bearer token and render a short-lived
 * browser object URL. Public-logo paths fall through to the Cloudflare Image
 * Resizing thumbnail helper for cheap, cached resizing.
 *
 * When there is no logo (or it fails to load) we fall back to a Building2 icon
 * so every card keeps the same compact avatar footprint.
 */

type AuthLike = {
  backendUrl?: string;
  token?: string;
  username?: string;
} | null;

type CacheEntry = { objectUrl: string; lastAccessAt: number };

const AUTHENTICATED_LOGO_CACHE_MAX = 48;
const objectUrlCache = new Map<string, CacheEntry>();
let cacheOwner = "";

function cacheKeyFor(url: string, backendUrl: string, username: string, token: string): string {
  return [backendUrl.trim(), username.trim().toLowerCase(), token, url].join("\u0000");
}

function syncCacheOwner(backendUrl: string, username: string, token: string) {
  const owner = [backendUrl.trim(), username.trim().toLowerCase(), token].join("\u0000");
  if (cacheOwner && cacheOwner !== owner) {
    for (const entry of objectUrlCache.values()) URL.revokeObjectURL(entry.objectUrl);
    objectUrlCache.clear();
  }
  cacheOwner = owner;
}

function getCached(cacheKey: string): string {
  const entry = objectUrlCache.get(cacheKey);
  if (!entry) return "";
  entry.lastAccessAt = Date.now();
  return entry.objectUrl;
}

function setCached(cacheKey: string, objectUrl: string) {
  const existing = objectUrlCache.get(cacheKey);
  if (existing?.objectUrl && existing.objectUrl !== objectUrl) {
    URL.revokeObjectURL(existing.objectUrl);
  }
  objectUrlCache.set(cacheKey, { objectUrl, lastAccessAt: Date.now() });
  if (objectUrlCache.size <= AUTHENTICATED_LOGO_CACHE_MAX) return;
  const sorted = [...objectUrlCache.entries()].sort(
    (a, b) => a[1].lastAccessAt - b[1].lastAccessAt,
  );
  for (const [key, entry] of sorted.slice(0, objectUrlCache.size - AUTHENTICATED_LOGO_CACHE_MAX)) {
    URL.revokeObjectURL(entry.objectUrl);
    objectUrlCache.delete(key);
  }
}

function mainApiBase(backendUrl: string): string {
  const raw = (backendUrl ?? "").trim();
  if (!raw) return "";
  try {
    const withProtocol = /^https?:\/\//i.test(raw) ? raw : `http://${raw}`;
    const parsed = new URL(withProtocol);
    const path = parsed.pathname.replace(/\/+$/, "");
    parsed.pathname = path.toLowerCase().endsWith("/goapi")
      ? path.slice(0, -"/goapi".length) || "/"
      : "/";
    parsed.search = "";
    return parsed.toString().replace(/\/+$/, "");
  } catch {
    return raw;
  }
}

function resolveDisplayUrl(value: string, backendUrl: string): string {
  const raw = (value ?? "").trim();
  if (!raw) return "";
  if (/^(blob:|data:|https?:\/\/)/i.test(raw)) return raw;
  if (raw.startsWith("//")) {
    return typeof window === "undefined" ? raw : `${window.location.protocol}${raw}`;
  }
  // Authenticated GoAPI/S3 proxy paths must be resolved against the backend host,
  // not the frontend origin. Otherwise fetch() hits localhost:3000 and 404s.
  if (raw.startsWith("/api/") || raw.startsWith("/goapi/")) {
    const base = mainApiBase(backendUrl);
    return base ? `${base}${raw}` : raw;
  }
  const base = mainApiBase(backendUrl);
  if (!base) return raw;
  if (raw.startsWith("/")) return `${base}${raw}`;
  if (raw.toLowerCase().startsWith("images/")) return `${base}/${raw.replace(/^\/+/, "")}`;
  return `${base}/images/${raw.replace(/^\/+/, "")}`;
}

/**
 * useLogoImage — internal hook that resolves a logo URI to a displayable URL.
 * Protected GoAPI/R2 paths are fetched with the bearer token and turned into a
 * short-lived browser object URL; public paths go through logoThumbUri() so
 * Cloudflare can serve a cached thumbnail.
 */
function useLogoImage(uri: string, auth: AuthLike, width: number) {
  const backendUrl = auth?.backendUrl ?? "";
  const token = auth?.token ?? "";
  const username = auth?.username ?? "";
  const isProtected = useMemo(
    () => imageNeedsAuthenticatedFetch(resolveDisplayUrl(uri, backendUrl), backendUrl),
    [uri, backendUrl],
  );
  // For public paths, ask Cloudflare for a resized thumbnail.
  const publicThumb = useMemo(
    () => (isProtected ? "" : logoThumbUri(uri, width)),
    [isProtected, uri, width],
  );
  const resolvedUrl = useMemo(
    () => (isProtected ? resolveDisplayUrl(uri, backendUrl) : publicThumb),
    [isProtected, uri, backendUrl, publicThumb],
  );
  const [displayUrl, setDisplayUrl] = useState("");
  const [failed, setFailed] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!resolvedUrl) {
      setDisplayUrl("");
      setFailed(false);
      setLoading(false);
      return;
    }
    if (!isProtected) {
      setDisplayUrl(resolvedUrl);
      setFailed(false);
      setLoading(false);
      return;
    }
    if (!token) {
      setDisplayUrl("");
      setFailed(true);
      setLoading(false);
      return;
    }
    syncCacheOwner(backendUrl, username, token);
    const cacheKey = cacheKeyFor(resolvedUrl, backendUrl, username, token);
    const cached = getCached(cacheKey);
    if (cached) {
      setDisplayUrl(cached);
      setFailed(false);
      setLoading(false);
      return;
    }
    const controller = new AbortController();
    let cancelled = false;
    setLoading(true);
    setFailed(false);
    setDisplayUrl("");
    void fetch(resolvedUrl, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.blob();
      })
      .then((blob) => {
        const objectUrl = URL.createObjectURL(blob);
        if (cancelled) {
          URL.revokeObjectURL(objectUrl);
          return;
        }
        setCached(cacheKey, objectUrl);
        setDisplayUrl(objectUrl);
        setFailed(false);
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setDisplayUrl("");
        setFailed(true);
        setLoading(false);
      });
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [resolvedUrl, isProtected, backendUrl, username, token]);

  return { displayUrl, failed, loading };
}

export type LogoAvatarProps = {
  uri: string | null | undefined;
  auth: AuthLike;
  alt: string;
  /** Tailwind size classes for the avatar box, e.g. "size-10". */
  sizeClass?: string;
  /** Lucide icon size for the fallback icon. */
  iconSize?: number;
  /** Thumbnail width in pixels (used when the logo is on a public path). */
  width?: number;
  /** Extra Tailwind classes for the avatar box. */
  className?: string;
  /** Lucide icon shown when there is no image. Defaults to Building2 (use UserRound for people). */
  fallbackIcon?: LucideIcon;
};

export function LogoAvatar({
  uri,
  auth,
  alt,
  sizeClass = "size-10",
  iconSize = 20,
  width = 160,
  className = "",
  fallbackIcon: FallbackIcon = Building2,
}: LogoAvatarProps) {
  const uriValue = (uri ?? "").trim();
  const { displayUrl, failed, loading } = useLogoImage(uriValue, auth, width);
  const hasLogo = Boolean(uriValue) && Boolean(displayUrl) && !failed;
  return (
    <span
      className={`grid ${sizeClass} shrink-0 place-items-center overflow-hidden rounded-xl bg-primary/10 text-primary ${className}`}
      aria-label={alt}
      role="img"
    >
      {hasLogo ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={displayUrl}
          alt={alt}
          className="h-full w-full object-contain"
          loading="lazy"
          onError={() => {
            /* surfaced via failed state from hook; keep img silent */
          }}
        />
      ) : loading ? (
        <span className="size-1/2 animate-pulse rounded-full bg-primary/30" />
      ) : (
        <FallbackIcon size={iconSize} aria-hidden="true" />
      )}
    </span>
  );
}

/**
 * useProfileAvatar — fetches the logged-in user's avatar thumbnail from
 * GET /api/auth/profile (best-effort). Returns "" when unavailable so the
 * caller can fall back to an icon. Never throws; failures are swallowed.
 */
export function useProfileAvatar(auth: AuthLike): string {
  const [avatar, setAvatar] = useState("");
  const token = auth?.token;
  const backendUrl = auth?.backendUrl;
  useEffect(() => {
    if (!token || !backendUrl) {
      setAvatar("");
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const response = await fetch(
          `/api/auth/profile?backendUrl=${encodeURIComponent(backendUrl)}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "x-bc-backend-url": backendUrl,
            },
            cache: "no-store",
          },
        );
        if (!response.ok) return;
        const payload = (await response.json()) as {
          data?: { avatar?: string; avatarthumb?: string };
        };
        if (cancelled) return;
        setAvatar(payload.data?.avatarthumb || payload.data?.avatar || "");
      } catch {
        if (!cancelled) setAvatar("");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token, backendUrl]);
  return avatar;
}
