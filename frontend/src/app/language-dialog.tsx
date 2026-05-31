"use client";

import { Check, Languages, X } from "lucide-react";
import Image from "next/image";
import { useEffect, useMemo, useState } from "react";
import { LANGUAGES, t, type LanguageCode } from "@/lib/i18n";

type LanguageDialogProps = {
  language: LanguageCode;
  onLanguageChange: (language: LanguageCode) => void;
  activeLanguageCodes?: string[];
};

export function LanguageDialog({ language, onLanguageChange, activeLanguageCodes }: LanguageDialogProps) {
  const [open, setOpen] = useState(false);
  const selectedLanguage = useMemo(
    () => LANGUAGES.find((item) => item.code === language) ?? LANGUAGES[0],
    [language],
  );

  const filteredLanguages = useMemo(() => {
    if (!activeLanguageCodes || activeLanguageCodes.length === 0) {
      return LANGUAGES;
    }
    return LANGUAGES.filter(
      (item) => activeLanguageCodes.includes(item.code) || item.code === language
    );
  }, [activeLanguageCodes, language]);

  useEffect(() => {
    if (!open) return;

    const previousOverflow = document.body.style.overflow;

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }

    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  function selectLanguage(nextLanguage: LanguageCode) {
    onLanguageChange(nextLanguage);
    setOpen(false);
  }

  return (
    <>
      <button
        aria-haspopup="dialog"
        aria-label={t(language, "selectLanguage")}
        className="language-dialog-trigger"
        onClick={() => setOpen(true)}
        title={t(language, "selectLanguage")}
        type="button"
      >
        <Languages aria-hidden="true" size={16} />
        <Image alt="" className="flag-icon" height={16} src={`/flags/${selectedLanguage.code}.png`} width={24} />
        <span>{selectedLanguage.name}</span>
      </button>

      {open ? (
        <div className="dialog-backdrop" onClick={() => setOpen(false)}>
          <section
            aria-label={t(language, "selectLanguage")}
            aria-modal="true"
            className="language-dialog"
            onClick={(event) => event.stopPropagation()}
            role="dialog"
          >
            <div className="dialog-header">
              <div>
                <p className="eyebrow">{t(language, "language")}</p>
                <h2>{t(language, "selectLanguage")}</h2>
              </div>
              <button aria-label={t(language, "close")} className="icon-button dialog-close" onClick={() => setOpen(false)} type="button">
                <X aria-hidden="true" size={18} />
              </button>
            </div>

            <div className="language-grid">
              {filteredLanguages.map((item) => {
                const selected = item.code === language;

                return (
                  <button
                    aria-pressed={selected}
                    className={selected ? "language-choice selected" : "language-choice"}
                    key={item.code}
                    onClick={() => selectLanguage(item.code)}
                    type="button"
                  >
                    <Image alt="" className="flag-icon" height={16} src={`/flags/${item.code}.png`} width={24} />
                    <span>{item.name}</span>
                    {selected ? <Check aria-hidden="true" size={17} /> : null}
                  </button>
                );
              })}
            </div>
          </section>
        </div>
      ) : null}
    </>
  );
}
