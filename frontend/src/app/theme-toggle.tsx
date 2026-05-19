"use client";

import { Check, Moon, Palette, Sun } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { t, type LanguageCode } from "@/lib/i18n";

type ThemeMode = "light" | "dark";
type ColorThemeId =
  | "ban-chiang"
  | "sukhothai-jade"
  | "ayutthaya-gold"
  | "lanna-teak"
  | "andaman-blue"
  | "siam-rose"
  | "violet-bloom"
  | "coral-sunset"
  | "citrus-lime"
  | "berry-magenta";
type ThemeVars = Record<`--${string}`, string>;

type ColorTheme = {
  dark: ThemeVars;
  id: ColorThemeId;
  light: ThemeVars;
  name: string;
  swatches: string[];
};

type PaletteSeed = {
  accent: string;
  accentForeground: string;
  actionBg: string;
  actionText: string;
  background: string;
  bgAccentA: string;
  bgAccentB: string;
  border: string;
  card: string;
  foreground: string;
  iconSoftBg: string;
  inputFocusBg: string;
  muted: string;
  mutedBg: string;
  primary: string;
  primaryForeground: string;
  secondary: string;
  secondaryForeground: string;
  tertiary: string;
};

const themeStorageKey = "bc_theme";
const colorThemeStorageKey = "bc_color_theme";
const defaultColorTheme: ColorThemeId = "ban-chiang";

const colorThemes: ColorTheme[] = [
  createColorTheme({
    id: "ban-chiang",
    name: "Ban Chiang Terracotta",
    swatches: ["#812920", "#a04035", "#fbf9f8"],
    light: {
      accent: "#ffdad5",
      accentForeground: "#802820",
      actionBg: "#ffdad5",
      actionText: "#812920",
      background: "#fbf9f8",
      bgAccentA: "rgba(160, 64, 53, 0.08)",
      bgAccentB: "rgba(131, 40, 14, 0.08)",
      border: "#dcc0bc",
      card: "#ffffff",
      foreground: "#1b1c1c",
      iconSoftBg: "#ffdad5",
      inputFocusBg: "#ffffff",
      muted: "#615e53",
      mutedBg: "#f0eded",
      primary: "#812920",
      primaryForeground: "#ffffff",
      secondary: "#e4dfd1",
      secondaryForeground: "#49473c",
      tertiary: "#83280e",
    },
    dark: {
      accent: "#4b221b",
      accentForeground: "#ffdad5",
      actionBg: "rgba(255, 218, 213, 0.16)",
      actionText: "#ffb4a9",
      background: "#1b1c1c",
      bgAccentA: "rgba(255, 180, 169, 0.1)",
      bgAccentB: "rgba(255, 181, 161, 0.08)",
      border: "#6c524d",
      card: "#242120",
      foreground: "#f3f0f0",
      iconSoftBg: "rgba(255, 218, 213, 0.12)",
      inputFocusBg: "#2b2726",
      muted: "#cbc6b8",
      mutedBg: "#303030",
      primary: "#ffb4a9",
      primaryForeground: "#410001",
      secondary: "#302e2a",
      secondaryForeground: "#f3f0f0",
      tertiary: "#ffb5a1",
    },
  }),
  createColorTheme({
    id: "sukhothai-jade",
    name: "Jade Green",
    swatches: ["#0f766e", "#0d9488", "#f7fbf7"],
    light: {
      accent: "#d7f3ed",
      accentForeground: "#075e58",
      actionBg: "#d7f3ed",
      actionText: "#0f766e",
      background: "#f7fbf7",
      bgAccentA: "rgba(15, 118, 110, 0.09)",
      bgAccentB: "rgba(22, 101, 52, 0.08)",
      border: "#c7ddd7",
      card: "#ffffff",
      foreground: "#12211f",
      iconSoftBg: "#d7f3ed",
      inputFocusBg: "#ffffff",
      muted: "#51615d",
      mutedBg: "#eef5f2",
      primary: "#0f766e",
      primaryForeground: "#ffffff",
      secondary: "#dce8df",
      secondaryForeground: "#2f4a43",
      tertiary: "#166534",
    },
    dark: {
      accent: "#123b3a",
      accentForeground: "#99f6e4",
      actionBg: "rgba(153, 246, 228, 0.16)",
      actionText: "#5eead4",
      background: "#0b1f1d",
      bgAccentA: "rgba(94, 234, 212, 0.1)",
      bgAccentB: "rgba(134, 239, 172, 0.08)",
      border: "#355a54",
      card: "#102826",
      foreground: "#e7f7f3",
      iconSoftBg: "rgba(153, 246, 228, 0.12)",
      inputFocusBg: "#12312f",
      muted: "#a8c8c0",
      mutedBg: "#173633",
      primary: "#5eead4",
      primaryForeground: "#042f2e",
      secondary: "#173633",
      secondaryForeground: "#e7f7f3",
      tertiary: "#86efac",
    },
  }),
  createColorTheme({
    id: "ayutthaya-gold",
    name: "Amber Gold",
    swatches: ["#8a5a00", "#b7791f", "#fffaf0"],
    light: {
      accent: "#fff0c2",
      accentForeground: "#713f12",
      actionBg: "#fff0c2",
      actionText: "#8a5a00",
      background: "#fffaf0",
      bgAccentA: "rgba(183, 121, 31, 0.1)",
      bgAccentB: "rgba(124, 45, 18, 0.07)",
      border: "#e7d7ad",
      card: "#ffffff",
      foreground: "#211a10",
      iconSoftBg: "#fff0c2",
      inputFocusBg: "#ffffff",
      muted: "#6f6048",
      mutedBg: "#f6eddc",
      primary: "#8a5a00",
      primaryForeground: "#ffffff",
      secondary: "#eadfca",
      secondaryForeground: "#574529",
      tertiary: "#7c2d12",
    },
    dark: {
      accent: "#4a3211",
      accentForeground: "#f6c46b",
      actionBg: "rgba(246, 196, 107, 0.16)",
      actionText: "#f6c46b",
      background: "#211a10",
      bgAccentA: "rgba(246, 196, 107, 0.1)",
      bgAccentB: "rgba(251, 146, 60, 0.08)",
      border: "#6f5933",
      card: "#2d2416",
      foreground: "#fff4dc",
      iconSoftBg: "rgba(246, 196, 107, 0.12)",
      inputFocusBg: "#362a18",
      muted: "#d3bf98",
      mutedBg: "#3a2d1a",
      primary: "#f6c46b",
      primaryForeground: "#3b2200",
      secondary: "#3a2d1a",
      secondaryForeground: "#fff4dc",
      tertiary: "#fdba74",
    },
  }),
  createColorTheme({
    id: "lanna-teak",
    name: "Teak Brown",
    swatches: ["#7c3f16", "#a16207", "#fbf7f1"],
    light: {
      accent: "#fde8c8",
      accentForeground: "#7c2d12",
      actionBg: "#fde8c8",
      actionText: "#7c3f16",
      background: "#fbf7f1",
      bgAccentA: "rgba(124, 63, 22, 0.1)",
      bgAccentB: "rgba(161, 98, 7, 0.08)",
      border: "#dfc9ae",
      card: "#ffffff",
      foreground: "#21170f",
      iconSoftBg: "#fde8c8",
      inputFocusBg: "#ffffff",
      muted: "#6a5845",
      mutedBg: "#f2eadf",
      primary: "#7c3f16",
      primaryForeground: "#ffffff",
      secondary: "#e7d8c6",
      secondaryForeground: "#4d3b2a",
      tertiary: "#854d0e",
    },
    dark: {
      accent: "#4b2a13",
      accentForeground: "#fdba74",
      actionBg: "rgba(253, 186, 116, 0.16)",
      actionText: "#fdba74",
      background: "#21170f",
      bgAccentA: "rgba(253, 186, 116, 0.1)",
      bgAccentB: "rgba(245, 158, 11, 0.08)",
      border: "#6b4f34",
      card: "#2b1e13",
      foreground: "#fff6eb",
      iconSoftBg: "rgba(253, 186, 116, 0.12)",
      inputFocusBg: "#342415",
      muted: "#d6bea4",
      mutedBg: "#3a2a1a",
      primary: "#fdba74",
      primaryForeground: "#431407",
      secondary: "#3a2a1a",
      secondaryForeground: "#fff6eb",
      tertiary: "#f59e0b",
    },
  }),
  createColorTheme({
    id: "andaman-blue",
    name: "Ocean Blue",
    swatches: ["#075985", "#0284c7", "#f5fbff"],
    light: {
      accent: "#dff3ff",
      accentForeground: "#075985",
      actionBg: "#dff3ff",
      actionText: "#075985",
      background: "#f5fbff",
      bgAccentA: "rgba(2, 132, 199, 0.1)",
      bgAccentB: "rgba(15, 118, 110, 0.08)",
      border: "#bdd7e7",
      card: "#ffffff",
      foreground: "#0b1720",
      iconSoftBg: "#dff3ff",
      inputFocusBg: "#ffffff",
      muted: "#526b78",
      mutedBg: "#eef6fb",
      primary: "#075985",
      primaryForeground: "#ffffff",
      secondary: "#dbeaf2",
      secondaryForeground: "#294756",
      tertiary: "#0f766e",
    },
    dark: {
      accent: "#123449",
      accentForeground: "#bae6fd",
      actionBg: "rgba(186, 230, 253, 0.16)",
      actionText: "#7dd3fc",
      background: "#0b1720",
      bgAccentA: "rgba(125, 211, 252, 0.1)",
      bgAccentB: "rgba(94, 234, 212, 0.08)",
      border: "#33586f",
      card: "#102131",
      foreground: "#eff8ff",
      iconSoftBg: "rgba(186, 230, 253, 0.12)",
      inputFocusBg: "#132b3d",
      muted: "#b4ccda",
      mutedBg: "#173044",
      primary: "#7dd3fc",
      primaryForeground: "#082f49",
      secondary: "#173044",
      secondaryForeground: "#eff8ff",
      tertiary: "#5eead4",
    },
  }),
  createColorTheme({
    id: "siam-rose",
    name: "Rose Slate",
    swatches: ["#be123c", "#e11d48", "#fafafa"],
    light: {
      accent: "#ffe4e6",
      accentForeground: "#9f1239",
      actionBg: "#ffe4e6",
      actionText: "#be123c",
      background: "#fafafa",
      bgAccentA: "rgba(225, 29, 72, 0.09)",
      bgAccentB: "rgba(51, 65, 85, 0.07)",
      border: "#d7dce2",
      card: "#ffffff",
      foreground: "#111827",
      iconSoftBg: "#ffe4e6",
      inputFocusBg: "#ffffff",
      muted: "#64748b",
      mutedBg: "#f1f5f9",
      primary: "#be123c",
      primaryForeground: "#ffffff",
      secondary: "#e2e8f0",
      secondaryForeground: "#334155",
      tertiary: "#334155",
    },
    dark: {
      accent: "#4c1020",
      accentForeground: "#fecdd3",
      actionBg: "rgba(254, 205, 211, 0.16)",
      actionText: "#fda4af",
      background: "#111827",
      bgAccentA: "rgba(251, 113, 133, 0.1)",
      bgAccentB: "rgba(148, 163, 184, 0.08)",
      border: "#475569",
      card: "#1f2937",
      foreground: "#f8fafc",
      iconSoftBg: "rgba(254, 205, 211, 0.12)",
      inputFocusBg: "#273244",
      muted: "#cbd5e1",
      mutedBg: "#263244",
      primary: "#fda4af",
      primaryForeground: "#4c0519",
      secondary: "#263244",
      secondaryForeground: "#f8fafc",
      tertiary: "#cbd5e1",
    },
  }),
  createColorTheme({
    id: "violet-bloom",
    name: "Violet Bloom",
    swatches: ["#7c3aed", "#db2777", "#fbf8ff"],
    light: {
      accent: "#ede9fe",
      accentForeground: "#5b21b6",
      actionBg: "#ede9fe",
      actionText: "#7c3aed",
      background: "#fbf8ff",
      bgAccentA: "rgba(124, 58, 237, 0.1)",
      bgAccentB: "rgba(219, 39, 119, 0.08)",
      border: "#d8c8ff",
      card: "#ffffff",
      foreground: "#1c1229",
      iconSoftBg: "#ede9fe",
      inputFocusBg: "#ffffff",
      muted: "#6b5f7a",
      mutedBg: "#f3eefc",
      primary: "#7c3aed",
      primaryForeground: "#ffffff",
      secondary: "#e9ddff",
      secondaryForeground: "#46345f",
      tertiary: "#db2777",
    },
    dark: {
      accent: "#2d1d4d",
      accentForeground: "#ddd6fe",
      actionBg: "rgba(221, 214, 254, 0.16)",
      actionText: "#c4b5fd",
      background: "#181022",
      bgAccentA: "rgba(196, 181, 253, 0.1)",
      bgAccentB: "rgba(244, 114, 182, 0.08)",
      border: "#56416f",
      card: "#241733",
      foreground: "#fbf7ff",
      iconSoftBg: "rgba(221, 214, 254, 0.12)",
      inputFocusBg: "#2b1c3f",
      muted: "#d7c9e7",
      mutedBg: "#302044",
      primary: "#c4b5fd",
      primaryForeground: "#2e1065",
      secondary: "#302044",
      secondaryForeground: "#fbf7ff",
      tertiary: "#f0abfc",
    },
  }),
  createColorTheme({
    id: "coral-sunset",
    name: "Coral Sunset",
    swatches: ["#ea580c", "#e11d48", "#fff7ed"],
    light: {
      accent: "#ffedd5",
      accentForeground: "#9a3412",
      actionBg: "#ffedd5",
      actionText: "#ea580c",
      background: "#fff7ed",
      bgAccentA: "rgba(234, 88, 12, 0.1)",
      bgAccentB: "rgba(225, 29, 72, 0.08)",
      border: "#fed7aa",
      card: "#ffffff",
      foreground: "#24130a",
      iconSoftBg: "#ffedd5",
      inputFocusBg: "#ffffff",
      muted: "#73513c",
      mutedBg: "#fff1e3",
      primary: "#ea580c",
      primaryForeground: "#ffffff",
      secondary: "#fde1c4",
      secondaryForeground: "#5d3924",
      tertiary: "#e11d48",
    },
    dark: {
      accent: "#4a2111",
      accentForeground: "#fed7aa",
      actionBg: "rgba(254, 215, 170, 0.16)",
      actionText: "#fdba74",
      background: "#21140d",
      bgAccentA: "rgba(253, 186, 116, 0.1)",
      bgAccentB: "rgba(251, 113, 133, 0.08)",
      border: "#704832",
      card: "#2f1b11",
      foreground: "#fff6ef",
      iconSoftBg: "rgba(254, 215, 170, 0.12)",
      inputFocusBg: "#3a2114",
      muted: "#e0c2ac",
      mutedBg: "#3b2417",
      primary: "#fdba74",
      primaryForeground: "#431407",
      secondary: "#3b2417",
      secondaryForeground: "#fff6ef",
      tertiary: "#fb7185",
    },
  }),
  createColorTheme({
    id: "citrus-lime",
    name: "Citrus Lime",
    swatches: ["#65a30d", "#0f766e", "#fbfff4"],
    light: {
      accent: "#ecfccb",
      accentForeground: "#3f6212",
      actionBg: "#ecfccb",
      actionText: "#4d7c0f",
      background: "#fbfff4",
      bgAccentA: "rgba(101, 163, 13, 0.1)",
      bgAccentB: "rgba(15, 118, 110, 0.08)",
      border: "#d9efaa",
      card: "#ffffff",
      foreground: "#152010",
      iconSoftBg: "#ecfccb",
      inputFocusBg: "#ffffff",
      muted: "#566848",
      mutedBg: "#f2f9df",
      primary: "#65a30d",
      primaryForeground: "#ffffff",
      secondary: "#e2f2bf",
      secondaryForeground: "#3f4f2b",
      tertiary: "#0f766e",
    },
    dark: {
      accent: "#263611",
      accentForeground: "#d9f99d",
      actionBg: "rgba(217, 249, 157, 0.16)",
      actionText: "#bef264",
      background: "#111a0b",
      bgAccentA: "rgba(190, 242, 100, 0.1)",
      bgAccentB: "rgba(94, 234, 212, 0.08)",
      border: "#4e6732",
      card: "#182410",
      foreground: "#f7ffe8",
      iconSoftBg: "rgba(217, 249, 157, 0.12)",
      inputFocusBg: "#1e2d12",
      muted: "#c5d8aa",
      mutedBg: "#233215",
      primary: "#bef264",
      primaryForeground: "#1a2e05",
      secondary: "#233215",
      secondaryForeground: "#f7ffe8",
      tertiary: "#5eead4",
    },
  }),
  createColorTheme({
    id: "berry-magenta",
    name: "Berry Magenta",
    swatches: ["#c026d3", "#e11d48", "#fff7fc"],
    light: {
      accent: "#fae8ff",
      accentForeground: "#86198f",
      actionBg: "#fae8ff",
      actionText: "#a21caf",
      background: "#fff7fc",
      bgAccentA: "rgba(192, 38, 211, 0.1)",
      bgAccentB: "rgba(225, 29, 72, 0.08)",
      border: "#efc7f4",
      card: "#ffffff",
      foreground: "#251124",
      iconSoftBg: "#fae8ff",
      inputFocusBg: "#ffffff",
      muted: "#71556f",
      mutedBg: "#fbedf8",
      primary: "#c026d3",
      primaryForeground: "#ffffff",
      secondary: "#f3d9f4",
      secondaryForeground: "#583356",
      tertiary: "#e11d48",
    },
    dark: {
      accent: "#46124f",
      accentForeground: "#f5d0fe",
      actionBg: "rgba(245, 208, 254, 0.16)",
      actionText: "#f0abfc",
      background: "#1f1024",
      bgAccentA: "rgba(240, 171, 252, 0.1)",
      bgAccentB: "rgba(251, 113, 133, 0.08)",
      border: "#6a3f70",
      card: "#2b1730",
      foreground: "#fff4fd",
      iconSoftBg: "rgba(245, 208, 254, 0.12)",
      inputFocusBg: "#351b3b",
      muted: "#dfc2e4",
      mutedBg: "#3a2040",
      primary: "#f0abfc",
      primaryForeground: "#4a044e",
      secondary: "#3a2040",
      secondaryForeground: "#fff4fd",
      tertiary: "#fb7185",
    },
  }),
];

const colorThemeById = new Map<ColorThemeId, ColorTheme>(colorThemes.map((theme) => [theme.id, theme]));

export function ThemeToggle({ language }: { language: LanguageCode }) {
  const [themeMode, setThemeMode] = useState<ThemeMode>("light");
  const [colorTheme, setColorTheme] = useState<ColorThemeId>(defaultColorTheme);
  const [themeReady, setThemeReady] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const groupRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const savedTheme = localStorage.getItem(themeStorageKey);
    const savedColorTheme = normalizeColorTheme(localStorage.getItem(colorThemeStorageKey));
    const nextTheme =
      savedTheme === "light" || savedTheme === "dark"
        ? savedTheme
        : window.matchMedia("(prefers-color-scheme: dark)").matches
          ? "dark"
          : "light";

    applyVisualTheme(nextTheme, savedColorTheme);
    setThemeMode(nextTheme);
    setColorTheme(savedColorTheme);
    setThemeReady(true);
  }, []);

  useEffect(() => {
    if (!themeReady) return;
    applyVisualTheme(themeMode, colorTheme);
    localStorage.setItem(themeStorageKey, themeMode);
    localStorage.setItem(colorThemeStorageKey, colorTheme);
  }, [colorTheme, themeMode, themeReady]);

  useEffect(() => {
    if (!paletteOpen) return;

    const handlePointerDown = (event: PointerEvent) => {
      if (!groupRef.current?.contains(event.target as Node)) setPaletteOpen(false);
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setPaletteOpen(false);
    };

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [paletteOpen]);

  const selectedColorTheme = colorThemeById.get(colorTheme) ?? colorThemeById.get(defaultColorTheme)!;
  const nextThemeLabel = themeMode === "dark" ? t(language, "switchToLightTheme") : t(language, "switchToDarkTheme");
  const colorThemeLabel = t(language, "colorTheme");
  const selectColorThemeLabel = t(language, "selectColorTheme");

  return (
    <div className="theme-control-group" ref={groupRef}>
      <button
        aria-expanded={paletteOpen}
        aria-haspopup="dialog"
        aria-label={selectColorThemeLabel}
        className="icon-button theme-palette-toggle"
        onClick={() => setPaletteOpen((current) => !current)}
        title={`${selectColorThemeLabel}: ${selectedColorTheme.name}`}
        type="button"
      >
        <Palette aria-hidden="true" size={18} />
        <span aria-hidden="true" className="theme-palette-dot" style={{ background: selectedColorTheme.swatches[0] }} />
      </button>

      {paletteOpen ? (
        <div aria-label={selectColorThemeLabel} className="theme-palette-popover" role="dialog">
          <p className="theme-palette-title">{colorThemeLabel}</p>
          <div className="theme-palette-list">
            {colorThemes.map((theme) => {
              const selected = theme.id === colorTheme;
              return (
                <button
                  aria-pressed={selected}
                  className={selected ? "theme-palette-choice selected" : "theme-palette-choice"}
                  key={theme.id}
                  onClick={() => {
                    setColorTheme(theme.id);
                    setPaletteOpen(false);
                  }}
                  type="button"
                >
                  <span aria-hidden="true" className="theme-palette-swatches">
                    {theme.swatches.map((color) => (
                      <span key={color} style={{ background: color }} />
                    ))}
                  </span>
                  <span className="theme-palette-name">{theme.name}</span>
                  {selected ? <Check aria-hidden="true" className="theme-palette-check" size={16} /> : null}
                </button>
              );
            })}
          </div>
        </div>
      ) : null}

      <button
        aria-label={nextThemeLabel}
        aria-pressed={themeMode === "dark"}
        className="icon-button theme-toggle"
        onClick={() => setThemeMode((current) => (current === "dark" ? "light" : "dark"))}
        title={nextThemeLabel}
        type="button"
      >
        {themeMode === "dark" ? <Sun aria-hidden="true" size={18} /> : <Moon aria-hidden="true" size={18} />}
      </button>
    </div>
  );
}

function createColorTheme(theme: Omit<ColorTheme, "dark" | "light"> & { dark: PaletteSeed; light: PaletteSeed }): ColorTheme {
  return {
    ...theme,
    dark: paletteVars(theme.dark, "dark"),
    light: paletteVars(theme.light, "light"),
  };
}

function normalizeColorTheme(value: string | null): ColorThemeId {
  return colorThemeById.has(value as ColorThemeId) ? (value as ColorThemeId) : defaultColorTheme;
}

function applyVisualTheme(themeMode: ThemeMode, colorTheme: ColorThemeId) {
  const selectedTheme = colorThemeById.get(colorTheme) ?? colorThemeById.get(defaultColorTheme)!;
  const vars = selectedTheme[themeMode];

  document.documentElement.dataset.theme = themeMode;
  document.documentElement.dataset.colorTheme = selectedTheme.id;
  document.documentElement.style.colorScheme = themeMode;
  for (const [name, value] of Object.entries(vars)) {
    document.documentElement.style.setProperty(name, value);
  }
}

function paletteVars(seed: PaletteSeed, mode: ThemeMode): ThemeVars {
  const isDark = mode === "dark";
  return {
    "--accent": seed.accent,
    "--accent-foreground": seed.accentForeground,
    "--action-bg": seed.actionBg,
    "--action-text": seed.actionText,
    "--amber": isDark ? "#f6c46b" : "#b7791f",
    "--background": seed.background,
    "--bg": seed.background,
    "--bg-accent-a": seed.bgAccentA,
    "--bg-accent-b": seed.bgAccentB,
    "--border": seed.border,
    "--card": seed.card,
    "--card-foreground": seed.foreground,
    "--danger": isDark ? "#ffb4ab" : "#ba1a1a",
    "--destructive": isDark ? "#ffb4ab" : "#ba1a1a",
    "--focus-ring": colorMixAlpha(seed.primary, isDark ? 0.2 : 0.16),
    "--foreground": seed.foreground,
    "--icon-soft-bg": seed.iconSoftBg,
    "--indigo": seed.primary,
    "--input": seed.border,
    "--input-focus-bg": seed.inputFocusBg,
    "--line": seed.border,
    "--message-error-bg": isDark ? "rgba(180, 35, 24, 0.18)" : "#ffdad6",
    "--message-success-bg": isDark ? "rgba(19, 121, 91, 0.18)" : "#e8f5ef",
    "--muted": seed.muted,
    "--muted-bg": seed.mutedBg,
    "--muted-foreground": seed.muted,
    "--overlay-bg": isDark ? "rgba(12, 9, 8, 0.74)" : "rgba(48, 40, 37, 0.54)",
    "--panel": seed.card,
    "--panel-soft": seed.mutedBg,
    "--panel-strong": isDark ? colorWithAlpha(seed.card, 0.95) : colorWithAlpha(seed.card, 0.94),
    "--placeholder": seed.muted,
    "--popover": seed.card,
    "--popover-foreground": seed.foreground,
    "--primary": seed.primary,
    "--primary-foreground": seed.primaryForeground,
    "--primary-shadow": colorMixAlpha(seed.primary, isDark ? 0.14 : 0.18),
    "--primary-shadow-strong": colorMixAlpha(seed.primary, isDark ? 0.2 : 0.26),
    "--ring": seed.primary,
    "--secondary": seed.secondary,
    "--secondary-foreground": seed.secondaryForeground,
    "--shadow": isDark ? "0 24px 80px rgba(0, 0, 0, 0.42)" : `0 8px 24px ${colorMixAlpha(seed.primary, 0.12)}`,
    "--shadow-card": isDark ? "0 4px 12px rgba(0, 0, 0, 0.28)" : `0 4px 12px ${colorMixAlpha(seed.primary, 0.08)}`,
    "--shadow-popover": isDark ? "0 8px 24px rgba(0, 0, 0, 0.38)" : `0 8px 24px ${colorMixAlpha(seed.primary, 0.12)}`,
    "--success": "#13795b",
    "--teal": seed.tertiary,
    "--text": seed.foreground,
    "--text-soft": seed.muted,
  };
}

function colorMixAlpha(hex: string, alpha: number) {
  const value = hex.replace("#", "");
  const red = Number.parseInt(value.slice(0, 2), 16);
  const green = Number.parseInt(value.slice(2, 4), 16);
  const blue = Number.parseInt(value.slice(4, 6), 16);
  return `rgba(${red}, ${green}, ${blue}, ${alpha})`;
}

function colorWithAlpha(hex: string, alpha: number) {
  return colorMixAlpha(hex, alpha);
}
