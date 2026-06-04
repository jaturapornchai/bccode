import type { CSSProperties } from "react";
import type { Metadata, Viewport } from "next";
import { cookies } from "next/headers";
import "./globals.css";
import { colorThemeStorageKey, getThemesMap, normalizeColorTheme, themeStorageKey, type ThemeMode } from "@/lib/theme-data";

export const metadata: Metadata = {
  title: "BC Ai Account Login",
  description: "BC Ai Account login",
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
  const htmlStyle = {
    colorScheme: themeMode,
    ...themeVars,
  } as CSSProperties;

  return (
    <html lang="th" data-theme={themeMode} data-color-theme={colorTheme} style={htmlStyle} suppressHydrationWarning>
      <body>{children}</body>
    </html>
  );
}
