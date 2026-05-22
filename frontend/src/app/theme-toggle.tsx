"use client";

import { Check, Moon, Palette, Sun } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { t, type LanguageCode } from "@/lib/i18n";
import {
  ThemeMode,
  ColorThemeId,
  themeStorageKey,
  colorThemeStorageKey,
  defaultColorTheme,
  colorThemes,
  normalizeColorTheme,
  applyVisualTheme,
} from "@/lib/theme-data";

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

  const selectedColorTheme = colorThemes.find((t) => t.id === colorTheme) || colorThemes[0];
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
