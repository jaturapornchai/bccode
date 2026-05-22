import type { Metadata, Viewport } from "next";
import "./globals.css";
import { getThemesMap } from "@/lib/theme-data";

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

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const themesMap = getThemesMap();

  return (
    <html lang="th" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `
              (function() {
                try {
                  var theme = localStorage.getItem('bc_theme');
                  var colorTheme = localStorage.getItem('bc_color_theme') || 'ban-chiang';
                  if (!theme) {
                    theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
                  }
                  document.documentElement.setAttribute('data-theme', theme);
                  document.documentElement.setAttribute('data-color-theme', colorTheme);
                  document.documentElement.style.colorScheme = theme;

                  var themes = ${JSON.stringify(themesMap)};
                  var vars = themes[colorTheme] ? themes[colorTheme][theme] : null;
                  if (vars) {
                    for (var name in vars) {
                      document.documentElement.style.setProperty(name, vars[name]);
                    }
                  }
                } catch (e) {}
              })();
            `,
          }}
        />
      </head>
      <body>{children}</body>
    </html>
  );
}
