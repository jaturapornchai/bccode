// Font picker data — user-selectable font families for the app UI.
// Each font supports Thai + Latin and loads from Google Fonts CDN so the picker
// preview can render with the real font without bundling every weight.
//
// Apply pattern: writeFontCookie sets a 1-year cookie so layout.tsx can read it
// server-side and set --font-sans on <html>; the FontPicker also sets the CSS
// variable at runtime for instant switching.

export type FontId =
  | "inter"
  | "noto-sans-thai"
  | "prompt"
  | "sarabun"
  | "kanit"
  | "ibm-plex-sans-thai"
  | "mitr"
  | "bai-jamjuree";

export type AppFont = {
  id: FontId;
  /** Display name shown in the picker UI (rendered in the font itself for live preview). */
  name: string;
  /** Short label for the trigger button. */
  short: string;
  /** CSS font-family value applied to --font-sans. */
  family: string;
  /** Google Fonts CSS2 URL with the weights we need. */
  href: string;
};

export const fontStorageKey = "bc_app_font";

export const defaultFontId: FontId = "inter";

export const appFonts: AppFont[] = [
  {
    id: "inter",
    name: "Inter — อ่านง่าย โมเดิร์น",
    short: "Inter",
    family: '"Inter", "Noto Sans Thai", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap",
  },
  {
    id: "noto-sans-thai",
    name: "Noto Sans Thai — ฟอนต์ไทยมาตรฐาน",
    short: "Noto Thai",
    family: '"Noto Sans Thai", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Noto+Sans+Thai:wght@400;500;600;700&family=Inter:wght@400;500;600;700;800&display=swap",
  },
  {
    id: "prompt",
    name: "Prompt — โค้งมน เป็นมิตร",
    short: "Prompt",
    family: '"Prompt", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Prompt:wght@400;500;600;700;800&display=swap",
  },
  {
    id: "sarabun",
    name: "Sarabun — อ่านสบาย ทางการ",
    short: "Sarabun",
    family: '"Sarabun", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Sarabun:wght@400;500;600;700&display=swap",
  },
  {
    id: "kanit",
    name: "Kanit — โค้งเด่น มีบุคลิก",
    short: "Kanit",
    family: '"Kanit", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Kanit:wght@400;500;600;700&display=swap",
  },
  {
    id: "ibm-plex-sans-thai",
    name: "IBM Plex Thai — เทคโนโลยี คมชัด",
    short: "IBM Plex",
    family: '"IBM Plex Sans Thai", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=IBM+Plex+Sans+Thai:wght@400;500;600;700&display=swap",
  },
  {
    id: "mitr",
    name: "Mitr — กลม ทันสมัย",
    short: "Mitr",
    family: '"Mitr", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Mitr:wght@400;500;600;700&display=swap",
  },
  {
    id: "bai-jamjuree",
    name: "Bai Jamjuree — เรขาคณิต สะอาด",
    short: "Bai Jamjuree",
    family: '"Bai Jamjuree", "Inter", system-ui, sans-serif',
    href: "https://fonts.googleapis.com/css2?family=Bai+Jamjuree:wght@400;500;600;700&display=swap",
  },
];

const fontIdSet = new Set<FontId>(appFonts.map((font) => font.id));

export function normalizeAppFont(value: string | null | undefined): FontId {
  if (value && fontIdSet.has(value as FontId)) return value as FontId;
  return defaultFontId;
}

export function getAppFont(id: FontId): AppFont {
  return appFonts.find((font) => font.id === id) ?? appFonts[0];
}

/** Inject a Google Fonts <link> once per font; returns the href for keying. */
export function ensureFontLoaded(font: AppFont): void {
  if (typeof document === "undefined") return;
  const linkId = `app-font-${font.id}`;
  if (document.getElementById(linkId)) return;
  const link = document.createElement("link");
  link.id = linkId;
  link.rel = "stylesheet";
  link.href = font.href;
  link.crossOrigin = "anonymous";
  document.head.appendChild(link);
}

/** Apply a font at runtime: set --font-sans on <html> so Tailwind var() picks it up. */
export function applyAppFont(id: FontId): void {
  const font = getAppFont(id);
  ensureFontLoaded(font);
  if (typeof document !== "undefined") {
    document.documentElement.style.setProperty("--font-sans", font.family);
    document.documentElement.dataset.appFont = id;
  }
}

/** Persist the font choice as a cookie so layout.tsx can read it server-side (no flash). */
export function writeFontCookie(id: FontId): void {
  if (typeof document === "undefined") return;
  document.cookie = `${fontStorageKey}=${encodeURIComponent(id)}; path=/; max-age=31536000; SameSite=Lax`;
}
