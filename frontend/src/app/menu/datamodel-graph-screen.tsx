"use client";

import { Brain, ExternalLink, Loader2, RefreshCcw } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { type LanguageCode } from "@/lib/i18n";

// 3D vault graph served by obsidian-jarvis-ui (D:/bccode/jarvisui — run start.bat)
const JARVIS_URL = "http://localhost:5173";

type DataModelGraphScreenProps = {
  embedded?: boolean;
  language?: LanguageCode;
};

const TEXT = {
  th: {
    title: "โครงสร้างข้อมูล (สมอง)",
    subtitle: "แผนผัง 3D สำหรับดูโมเดลหลักที่เก็บใน MongoDB — คลิกเพื่อดู collection, field และความสัมพันธ์",
    openTab: "เปิดเต็มจอ",
    retry: "ลองใหม่",
    checking: "กำลังเชื่อมต่อ Jarvis UI...",
    down: "Jarvis UI ยังไม่ทำงาน",
    downHint: "ดับเบิลคลิก D:\\bccode\\jarvisui\\start.bat แล้วกดลองใหม่",
  },
  en: {
    title: "Data Model Graph (Brain)",
    subtitle: "3D browser for MongoDB root models — click to inspect collections, fields, and relationships",
    openTab: "Open full screen",
    retry: "Retry",
    checking: "Connecting to Jarvis UI...",
    down: "Jarvis UI is not running",
    downHint: "Double-click D:\\bccode\\jarvisui\\start.bat then retry",
  },
};

export function DataModelGraphScreen({ language = "th" }: DataModelGraphScreenProps) {
  const text = TEXT[language === "th" ? "th" : "en"];
  const jarvisUrl = `${JARVIS_URL}?lang=${language === "th" ? "th" : "en"}`;
  const [status, setStatus] = useState<"checking" | "up" | "down">("checking");

  const check = useCallback(() => {
    setStatus("checking");
    fetch(`${JARVIS_URL}/api/graph/status`, { cache: "no-store" })
      .then((response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        setStatus("up");
      })
      .catch(() => setStatus("down"));
  }, []);
  useEffect(check, [check]);
  useEffect(() => {
    if (status !== "down") return;
    const retryTimer = window.setTimeout(check, 1_000);
    return () => window.clearTimeout(retryTimer);
  }, [check, status]);

  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      <div className="flex flex-wrap items-center gap-3">
        <span className="grid h-10 w-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
          <Brain className="h-5 w-5" />
        </span>
        <div className="min-w-0 flex-1">
          <h2 className="text-base font-semibold leading-tight">{text.title}</h2>
          <p className="truncate text-xs text-muted-foreground">{text.subtitle}</p>
        </div>
        <Button variant="outline" size="sm" className="h-9 gap-1" onClick={() => window.open(jarvisUrl, "_blank")}>
          <ExternalLink className="h-3.5 w-3.5" />
          {text.openTab}
        </Button>
      </div>

      <Card className="min-h-[560px] flex-1 overflow-hidden lg:min-h-0">
        <CardContent className="relative h-full min-h-[560px] p-0 lg:min-h-0">
          {status === "up" ? (
            <iframe src={jarvisUrl} title="Jarvis UI — Data Model Graph" className="h-full min-h-[560px] w-full border-0 bg-black lg:min-h-0" />
          ) : (
            <div className="grid h-full min-h-[560px] place-items-center text-sm text-muted-foreground lg:min-h-0">
              {status === "checking" ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  {text.checking}
                </span>
              ) : (
                <div className="flex flex-col items-center gap-3 text-center">
                  <p className="font-medium text-foreground">{text.down}</p>
                  <p className="text-xs">{text.downHint}</p>
                  <Button variant="secondary" size="sm" className="gap-1" onClick={check}>
                    <RefreshCcw className="h-3.5 w-3.5" />
                    {text.retry}
                  </Button>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
