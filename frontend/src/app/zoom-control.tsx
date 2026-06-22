"use client";

import { Check, ChevronDown, Minus, Plus, ZoomIn } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { backendText, type BackendLanguageDictionary } from "@/lib/backend-language";
import { t, type LanguageCode } from "@/lib/i18n";

const zoomStorageKey = "bc_ui_zoom";
const defaultZoom = 100;
const minZoom = 50;
const maxZoom = 150;
const zoomStep = 5;
const zoomOptions = [150, 100, 75, 50] as const;

type ZoomControlProps = {
  dictionary?: BackendLanguageDictionary;
  language: LanguageCode;
};

export function ZoomControl({ dictionary, language }: ZoomControlProps) {
  const [zoom, setZoom] = useState(defaultZoom);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    // On the first entry per browser session, open at 100% (default) instead of restoring the
    // last saved zoom — the menu should start clean at 100%. Within the same session the user's
    // zoom choice is still remembered.
    const sessionKey = "bc_ui_zoom_session";
    const firstThisSession = sessionStorage.getItem(sessionKey) !== "1";
    let initialZoom = normalizeZoom(localStorage.getItem(zoomStorageKey));
    if (firstThisSession) {
      sessionStorage.setItem(sessionKey, "1");
      initialZoom = defaultZoom;
      localStorage.setItem(zoomStorageKey, String(defaultZoom));
    }
    applyUiZoom(initialZoom);
    setZoom(initialZoom);
    setReady(true);
  }, []);

  useEffect(() => {
    if (!ready) return;
    applyUiZoom(zoom);
    localStorage.setItem(zoomStorageKey, String(zoom));
  }, [ready, zoom]);

  const labels = useMemo(
    () => ({
      selectZoom: languageText(dictionary, "select_zoom", t(language, "selectZoom")),
      zoom: languageText(dictionary, "zoom", t(language, "zoom")),
      zoomIn: languageText(dictionary, "zoom_in", t(language, "zoomIn")),
      zoomOut: languageText(dictionary, "zoom_out", t(language, "zoomOut")),
    }),
    [dictionary, language],
  );

  return (
    <div aria-label={labels.zoom} className="zoom-control" role="group">
      <Button
        aria-label={labels.zoomIn}
        className="zoom-control-button"
        disabled={zoom >= maxZoom}
        onClick={() => setZoom((current) => stepZoom(current, 1))}
        title={labels.zoomIn}
        type="button"
        variant="ghost"
      >
        <Plus aria-hidden="true" className="h-4 w-4" />
      </Button>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button aria-label={labels.selectZoom} className="zoom-control-trigger" title={`${labels.zoom}: ${zoom}%`} type="button" variant="ghost">
            <ZoomIn aria-hidden="true" className="h-4 w-4" />
            <span className="tabular-nums">{zoom}%</span>
            <ChevronDown aria-hidden="true" className="h-3.5 w-3.5 opacity-70" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="zoom-control-menu">
          {zoomOptions.map((option) => (
            <DropdownMenuItem key={option} className="zoom-control-option" onClick={() => setZoom(option)}>
              <span className="tabular-nums">{option}%</span>
              {option === zoom ? <Check aria-hidden="true" className="ml-auto h-4 w-4" /> : null}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>

      <Button
        aria-label={labels.zoomOut}
        className="zoom-control-button"
        disabled={zoom <= minZoom}
        onClick={() => setZoom((current) => stepZoom(current, -1))}
        title={labels.zoomOut}
        type="button"
        variant="ghost"
      >
        <Minus aria-hidden="true" className="h-4 w-4" />
      </Button>
    </div>
  );
}

function languageText(dictionary: BackendLanguageDictionary | undefined, key: string, fallback: string): string {
  return dictionary ? backendText(dictionary, key, fallback) || fallback : fallback;
}

function normalizeZoom(value: string | null): number {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return defaultZoom;
  const rounded = Math.round(parsed / zoomStep) * zoomStep;
  return Math.min(Math.max(rounded, minZoom), maxZoom);
}

function stepZoom(current: number, direction: 1 | -1): number {
  return Math.min(Math.max(current + direction * zoomStep, minZoom), maxZoom);
}

function applyUiZoom(zoom: number) {
  document.documentElement.dataset.uiZoom = String(zoom);
  document.documentElement.style.fontSize = `${zoom}%`;
  document.body.style.removeProperty("zoom");
}
