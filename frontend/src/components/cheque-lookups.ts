"use client";

import { useEffect, useState } from "react";
import { listBankBooks, listBankCodes, type BankBook, type BankCode } from "@/lib/cheque-api";
import type { BankCodeKind } from "@/lib/bank-code-kinds";
import type { LanguageCode } from "@/lib/i18n";

const PICKER_LIMIT = 500;

export type PickerOption = { value: string; label: string };

/** ตัวเลือกต้องมีค่าที่เลือกอยู่เสมอ แม้รหัสนั้นถูกปิดการใช้งานภายหลัง (เอกสารเดิมต้องยังแสดงค่าเดิมได้) */
export const withCurrentOption = (options: PickerOption[], current: string): PickerOption[] =>
  current && !options.some((option) => option.value === current) ? [...options, { value: current, label: current }] : options;

/** สมุดเงินฝากที่เปิดใช้งาน สำหรับคอมโบเลือกสมุดของจอเช็ค */
export function useBankBooks(language: LanguageCode, reloadKey: number): { books: BankBook[]; error: unknown } {
  const [books, setBooks] = useState<BankBook[]>([]);
  const [error, setError] = useState<unknown>(null);
  useEffect(() => {
    let current = true;
    setError(null);
    listBankBooks({ limit: PICKER_LIMIT }, language)
      .then((result) => {
        if (current) setBooks((result.items ?? []).filter((book) => book.is_active));
      })
      .catch((cause: unknown) => {
        if (current) setError(cause);
      });
    return () => {
      current = false;
    };
  }, [language, reloadKey]);
  return { books, error };
}

/** รหัสชนิดหนึ่ง (ธนาคาร / สถานะเพิ่มเติม ฯลฯ) ที่เปิดใช้งาน; bankCode ใช้กรองสาขาของธนาคารนั้น */
export function useCodeList(kind: BankCodeKind, language: LanguageCode, reloadKey: number, bankCode?: string): { codes: BankCode[]; error: unknown } {
  const [codes, setCodes] = useState<BankCode[]>([]);
  const [error, setError] = useState<unknown>(null);
  useEffect(() => {
    if (kind === "bank_branch" && !bankCode) {
      setCodes([]);
      return;
    }
    let current = true;
    setError(null);
    listBankCodes(kind, { bankCode, activeOnly: true, limit: PICKER_LIMIT }, language)
      .then((result) => {
        if (current) setCodes(result.items ?? []);
      })
      .catch((cause: unknown) => {
        if (current) {
          setCodes([]);
          setError(cause);
        }
      });
    return () => {
      current = false;
    };
  }, [kind, bankCode, language, reloadKey]);
  return { codes, error };
}
