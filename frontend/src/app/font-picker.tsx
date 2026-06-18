"use client";

import { Check, ChevronDown, Type } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { type LanguageCode } from "@/lib/i18n";
import {
  appFonts,
  applyAppFont,
  defaultFontId,
  ensureFontLoaded,
  fontStorageKey,
  getAppFont,
  normalizeAppFont,
  writeFontCookie,
  type FontId,
} from "@/lib/font-data";

export function FontPicker({ language }: { language: LanguageCode }) {
  const [fontId, setFontId] = useState<FontId>(defaultFontId);
  const [open, setOpen] = useState(false);
  const [ready, setReady] = useState(false);
  const groupRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const saved = normalizeAppFont(localStorage.getItem(fontStorageKey));
    setFontId(saved);
    setReady(true);
  }, []);

  useEffect(() => {
    if (!ready) return;
    applyAppFont(fontId);
    localStorage.setItem(fontStorageKey, fontId);
    writeFontCookie(fontId);
  }, [fontId, ready]);

  // Preload every font stylesheet so the live preview renders instantly on open.
  useEffect(() => {
    appFonts.forEach((font) => ensureFontLoaded(font));
  }, []);

  useEffect(() => {
    if (!open) return;
    const handlePointerDown = (event: PointerEvent) => {
      if (!groupRef.current?.contains(event.target as Node)) setOpen(false);
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  const current = getAppFont(fontId);
  const label = language === "th" ? "เลือกฟอนต์" : "Choose font";

  return (
    <div className="font-control-group" ref={groupRef}>
      <button
        aria-expanded={open}
        aria-haspopup="dialog"
        aria-label={label}
        className="icon-button font-toggle"
        onClick={() => setOpen((value) => !value)}
        title={`${label}: ${current.short}`}
        type="button"
      >
        <Type aria-hidden="true" size={18} />
        <ChevronDown aria-hidden="true" size={11} className={open ? "font-toggle-chev open" : "font-toggle-chev"} />
      </button>

      {open ? (
        <div aria-label={label} className="font-popover" role="dialog">
          <p className="font-popover-title">{label}</p>
          <div className="font-popover-list">
            {appFonts.map((font) => {
              const selected = font.id === fontId;
              return (
                <button
                  aria-pressed={selected}
                  className={selected ? "font-choice selected" : "font-choice"}
                  key={font.id}
                  onClick={() => {
                    setFontId(font.id);
                    setOpen(false);
                  }}
                  style={{ fontFamily: font.family }}
                  type="button"
                >
                  <span className="font-choice-name">{font.name}</span>
                  <span className="font-choice-sample" aria-hidden="true">
                    กขค ABC 0123
                  </span>
                  {selected ? <Check aria-hidden="true" className="font-choice-check" size={16} /> : null}
                </button>
              );
            })}
          </div>
        </div>
      ) : null}
    </div>
  );
}
