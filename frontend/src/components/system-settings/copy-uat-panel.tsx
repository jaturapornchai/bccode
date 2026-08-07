"use client";

// Copy UAT Panel - extracted from system-settings-screen.tsx
// Allows copying data from UAT/PRO environment to current DEV shop

import { Copy, Loader2, Search } from "lucide-react";
import type { LanguageCode } from "@/lib/i18n";
import type { SystemSettingConfig } from "@/lib/system-setting-screens";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import {
  type SettingRecord,
} from "@/components/system-settings/types";
import { recordTitle, type UiTextKey } from "@/components/system-settings/utils";

export function CopyUatPanel({
  config,
  copyPreview,
  language,
  loading,
  records,
  runCopy,
  saving,
  selected,
  setSelected,
  sourceEnvironment,
  setSourceEnvironment,
  targetHoldingCode,
  text,
}: {
  config: SystemSettingConfig;
  copyPreview: unknown;
  language: LanguageCode;
  loading: boolean;
  records: SettingRecord[];
  runCopy: (action: "copy" | "preview") => void;
  saving: boolean;
  selected: string;
  setSelected: (value: string) => void;
  sourceEnvironment: "uat" | "pro";
  setSourceEnvironment: (value: "uat" | "pro") => void;
  targetHoldingCode: string;
  text: (key: UiTextKey) => string;
}) {
  return (
    <Card>
      <CardContent className="grid gap-3 p-3">
        <div className="grid gap-2 lg:grid-cols-[minmax(180px,220px)_minmax(0,1fr)_auto_auto]">
          <label className="grid gap-1 text-sm font-semibold">
            <span>{text("sourceEnvironment")}</span>
            <select
              className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              disabled={loading || saving}
              onChange={(event) => {
                setSelected("");
                setSourceEnvironment(
                  event.target.value === "pro" ? "pro" : "uat",
                );
              }}
              value={sourceEnvironment}
            >
              <option value="uat">{text("sourceUat")}</option>
              <option value="pro">{text("sourcePro")}</option>
            </select>
          </label>
          <label className="grid gap-1 text-sm font-semibold">
            <span>{text("sourceShop")}</span>
            <select
              className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              disabled={loading}
              onChange={(event) => setSelected(event.target.value)}
              value={selected}
            >
              <option value="">{text("selectSourceShop")}</option>
              {records.map((record, index) => {
                const id = String(
                  record.guidfixed ?? record.holdingcode ?? record._id ?? "",
                );
                return (
                  <option key={`${id}-${index}`} value={id}>
                    {recordTitle(record, config, language)} ({id})
                  </option>
                );
              })}
            </select>
          </label>
          <Button
            type="button"
            variant="outline"
            onClick={() => runCopy("preview")}
            disabled={!selected || saving}
          >
            {saving ? <Loader2 className="animate-spin" /> : <Search />}
            {text("preview")}
          </Button>
          <Button
            type="button"
            onClick={() => runCopy("copy")}
            disabled={!selected || saving}
          >
            {saving ? <Loader2 className="animate-spin" /> : <Copy />}
            {text("copyNow")}
          </Button>
        </div>
        <div className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm">
          <b>
            {text("sourceEnvironment")}: {sourceEnvironment.toUpperCase()} → DEV
          </b>
          <b>
            {text("targetShop")}: {targetHoldingCode || "-"}
          </b>
          <p className="text-muted-foreground">
            Preview ก่อน copy ทุกครั้ง เพราะ action นี้มีผลกับข้อมูล MongoDB ของ
            shop ปัจจุบัน
          </p>
        </div>
        {copyPreview ? (
          <pre className="max-h-80 overflow-auto rounded-2xl border border-border bg-muted p-3 text-xs text-muted-foreground">
            {JSON.stringify(copyPreview, null, 2)}
          </pre>
        ) : null}
      </CardContent>
    </Card>
  );
}
