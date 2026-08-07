"use client";

import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type ImportRow = {
  row: number;
  username: string;
  email: string;
  name: string;
  role: number;
  position: string;
  department: string;
  valid: boolean;
  message: string;
  imported?: boolean;
};

type ImportResult = {
  total: number;
  valid: number;
  imported: number;
  failed: number;
  rows: ImportRow[];
};

const ROLE_LABEL: Record<number, string> = { 0: "ผู้ใช้", 1: "แอดมิน", 2: "เจ้าของ" };
const MAX_FILE_BYTES = 5 * 1024 * 1024; // 5MB — mirrors the backend cap

async function fileToBase64(file: File): Promise<string> {
  const bytes = new Uint8Array(await file.arrayBuffer());
  let binary = "";
  const chunk = 0x8000;
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk));
  }
  return btoa(binary);
}

// Self-contained modal: pick a .csv/.xlsx → preview (validate, no write) → confirm (write).
export function BulkUserImport({
  open,
  onClose,
  holdingcode,
  authToken,
  onImported,
}: {
  open: boolean;
  onClose: () => void;
  holdingcode: string;
  authToken: string;
  onImported: () => void;
}) {
  const [filename, setFilename] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<ImportResult | null>(null);
  const [committed, setCommitted] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  if (!open) return null;

  const reset = () => {
    setFilename("");
    setError("");
    setResult(null);
    setCommitted(false);
    setSelectedFile(null);
    if (fileRef.current) fileRef.current.value = "";
  };

  const callImport = async (file: File, commit: boolean): Promise<ImportResult> => {
    const contentbase64 = await fileToBase64(file);
    const res = await fetch("/api/holding-users-import", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${authToken}`,
      },
      body: JSON.stringify({ holdingcode, filename: file.name, contentbase64, commit }),
    });
    const json = (await res.json().catch(() => ({ success: false }))) as { success?: boolean; message?: string; data?: ImportResult };
    if (!res.ok || json?.success === false) {
      throw new Error(json?.message || "นำเข้าไม่สำเร็จ");
    }
    return json.data as ImportResult;
  };

  const onPickFile = async (file: File | undefined) => {
    if (!file) return;
    setFilename(file.name);
    if (file.size > MAX_FILE_BYTES) {
      setSelectedFile(null);
      setResult(null);
      setCommitted(false);
      setError("ไฟล์ใหญ่เกินไป (สูงสุด 5MB)");
      return;
    }
    setBusy(true);
    setError("");
    setResult(null);
    setCommitted(false);
    setSelectedFile(file);
    try {
      setResult(await callImport(file, false));
    } catch (e) {
      setError(e instanceof Error ? e.message : "อ่านไฟล์ไม่สำเร็จ");
    } finally {
      setBusy(false);
    }
  };

  const onCommit = async () => {
    if (!selectedFile) return;
    setBusy(true);
    setError("");
    try {
      setResult(await callImport(selectedFile, true));
      setCommitted(true);
      onImported();
    } catch (e) {
      setError(e instanceof Error ? e.message : "นำเข้าไม่สำเร็จ");
    } finally {
      setBusy(false);
    }
  };

  const invalidCount = result ? result.total - result.valid : 0;

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={onClose}>
      <div
        className="flex max-h-[85vh] w-full max-w-2xl flex-col gap-3 overflow-hidden rounded-xl border border-border bg-card p-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold">นำเข้าผู้ใช้จาก Excel/CSV</h2>
          <Button type="button" size="sm" variant="ghost" onClick={onClose}>
            ปิด
          </Button>
        </div>

        <p className="text-xs text-muted-foreground">
          หัวตาราง: <b>usercode</b> (จำเป็น), username (ชื่อแสดง), email (ไม่บังคับ), role (ผู้ใช้/แอดมิน/เจ้าของ หรือ 0/1/2), position, department
        </p>

        <div className="flex flex-wrap items-center gap-2">
          <input
            ref={fileRef}
            type="file"
            accept=".csv,.xlsx,.xls,.txt"
            disabled={busy}
            onChange={(e) => void onPickFile(e.target.files?.[0])}
            className="text-sm"
          />
          {filename ? (
            <Button type="button" size="sm" variant="outline" onClick={reset} disabled={busy}>
              ล้าง
            </Button>
          ) : null}
        </div>

        {error ? (
          <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive">
            {error}
          </div>
        ) : null}

        {result ? (
          <>
            <div className="text-sm">
              พบ {result.total} แถว · ถูกต้อง <b className="text-emerald-600">{result.valid}</b>
              {invalidCount > 0 ? (
                <>
                  {" "}· มีปัญหา <b className="text-destructive">{invalidCount}</b>
                </>
              ) : null}
              {committed ? (
                <>
                  {" "}· นำเข้าแล้ว <b className="text-emerald-600">{result.imported}</b>
                  {result.failed > 0 ? (
                    <>
                      {" "}· ล้มเหลว <b className="text-destructive">{result.failed}</b>
                    </>
                  ) : null}
                </>
              ) : null}
            </div>
            <div className="min-h-0 flex-1 overflow-auto rounded-lg border border-border">
              <table className="w-full text-left text-xs">
                <thead className="sticky top-0 bg-muted">
                  <tr>
                    <th className="px-2 py-1">#</th>
					<th className="px-2 py-1">รหัสผู้ใช้</th>
					<th className="px-2 py-1">ชื่อผู้ใช้</th>
					<th className="px-2 py-1">อีเมล</th>
                    <th className="px-2 py-1">สิทธิ์</th>
                    <th className="px-2 py-1">สถานะ</th>
                  </tr>
                </thead>
                <tbody>
                  {result.rows.map((r) => (
                    <tr key={r.row} className={cn("border-t border-border", !r.valid && "bg-destructive/5")}>
                      <td className="px-2 py-1 text-muted-foreground">{r.row}</td>
                      <td className="px-2 py-1 break-all">{r.username || <span className="text-muted-foreground">—</span>}</td>
                      <td className="px-2 py-1">{r.name}</td>
					  <td className="px-2 py-1 break-all">{r.email || <span className="text-muted-foreground">—</span>}</td>
                      <td className="px-2 py-1">{ROLE_LABEL[r.role] ?? r.role}</td>
                      <td className="px-2 py-1">
                        {committed ? (
                          r.imported ? (
                            <span className="text-emerald-600">✓ นำเข้าแล้ว</span>
                          ) : (
                            <span className="text-destructive">{r.message || (r.valid ? "ล้มเหลว" : "ข้าม")}</span>
                          )
                        ) : r.valid ? (
                          <span className="text-emerald-600">✓ พร้อม</span>
                        ) : (
                          <span className="text-destructive">{r.message}</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        ) : null}

        <div className="flex items-center justify-end gap-2">
          {result && !committed && result.valid > 0 ? (
            <Button type="button" onClick={() => void onCommit()} disabled={busy}>
              {busy ? "กำลังนำเข้า..." : `ยืนยันนำเข้า ${result.valid} รายการ`}
            </Button>
          ) : null}
          {committed ? (
            <Button type="button" onClick={onClose}>
              เสร็จสิ้น
            </Button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
