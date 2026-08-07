import type { CSSProperties } from "react";
import type { Metadata, Viewport } from "next";
import { cookies } from "next/headers";
import { Inter, Outfit, Noto_Sans_Thai } from "next/font/google";
import "./globals.css";
import { colorThemeStorageKey, getThemesMap, normalizeColorTheme, themeStorageKey, type ThemeMode } from "@/lib/theme-data";
import { fontStorageKey, getAppFont, normalizeAppFont } from "@/lib/font-data";
import { ToastViewport } from "@/components/toast-viewport";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

const outfit = Outfit({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
  variable: "--font-display",
  display: "swap",
});

const notoSansThai = Noto_Sans_Thai({
  subsets: ["thai"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-thai",
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "BC Ai Account",
    template: "%s | BC Ai Account",
  },
  description: "ระบบบัญชีและจัดการร้านค้าสำหรับธุรกิจไทย",
  icons: {
    icon: "/favicon.ico",
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
};

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const cookieStore = await cookies();
  const savedTheme = cookieStore.get(themeStorageKey)?.value;
  const themeMode: ThemeMode = savedTheme === "dark" || savedTheme === "light" ? savedTheme : "light";
  const colorTheme = normalizeColorTheme(cookieStore.get(colorThemeStorageKey)?.value ?? null);
  const themesMap = getThemesMap();
  const themeVars = themesMap[colorTheme]?.[themeMode] ?? {};

  // Resolve the user-selected UI font from cookie so there is no flash on reload.
  const savedFontId = normalizeAppFont(cookieStore.get(fontStorageKey)?.value ?? null);
  const appFont = getAppFont(savedFontId);

  const htmlStyle = {
    colorScheme: themeMode,
    "--font-sans": appFont.family,
    ...themeVars,
  } as CSSProperties;

  return (
    <html
      lang="th"
      data-theme={themeMode}
      data-color-theme={colorTheme}
      data-app-font={savedFontId}
      style={htmlStyle}
      suppressHydrationWarning
      className={`${inter.variable} ${outfit.variable} ${notoSansThai.variable}`}
    >
      <head>
        {/* Preload all picker fonts so live previews render instantly. */}
        {[
          "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap",
          "https://fonts.googleapis.com/css2?family=Noto+Sans+Thai:wght@400;500;600;700&display=swap",
          "https://fonts.googleapis.com/css2?family=Prompt:wght@400;500;600;700;800&display=swap",
          "https://fonts.googleapis.com/css2?family=Sarabun:wght@400;500;600;700&display=swap",
          "https://fonts.googleapis.com/css2?family=Kanit:wght@400;500;600;700&display=swap",
          "https://fonts.googleapis.com/css2?family=IBM+Plex+Sans+Thai:wght@400;500;600;700&display=swap",
          "https://fonts.googleapis.com/css2?family=Mitr:wght@400;500;600;700&display=swap",
          "https://fonts.googleapis.com/css2?family=Bai+Jamjuree:wght@400;500;600;700&display=swap",
        ].map((href) => (
          <link key={href} rel="stylesheet" href={href} />
        ))}
      </head>
      <body>
        {children}
        <ToastViewport />
      </body>
    </html>
  );
}
