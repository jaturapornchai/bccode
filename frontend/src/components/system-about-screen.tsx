"use client";

import { useMemo, useState } from "react";
import {
  Award,
  CheckCircle2,
  Clock,
  Cpu,
  Database,
  ExternalLink,
  Globe,
  HardDrive,
  Info,
  Layers,
  RefreshCw,
  Server,
  Shield,
  ShieldCheck,
  Terminal,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { type LanguageCode } from "@/lib/i18n";

export interface SystemAboutScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemAboutScreen({
  language = "th",
  holdingcode = "rungrueng",
}: SystemAboutScreenProps) {
  const [refreshedAt, setRefreshedAt] = useState<string>(() => new Date().toLocaleTimeString("th-TH"));

  const systemMetrics = useMemo(() => [
    { label: language === "th" ? "เวอร์ชันระบบ (Version)" : "Release Version", value: "r20261010-phase8-parity-6", status: "active" },
    { label: language === "th" ? "แกนสถาปัตยกรรม (Core)" : "Core Architecture", value: "Go Pure + Next.js Turbopack", status: "active" },
    { label: language === "th" ? "ฐานข้อมูลหลัก (Database)" : "Primary Database", value: "PostgreSQL 18 (ACID Sync)", status: "active" },
    { label: language === "th" ? "ระบบไฟล์และเอกสาร (Storage)" : "Document Storage", value: "S3 / MinIO (Encrypted)", status: "active" },
    { label: language === "th" ? "โซนเวลา (Timezone)" : "Timezone", value: "Asia/Bangkok (UTC+07:00)", status: "active" },
    { label: language === "th" ? "โฮลดิ้งปัจจุบัน (Active Holding)" : "Current Holding", value: holdingcode || "rungrueng", status: "active" },
  ], [language, holdingcode]);

  return (
    <div className="flex h-full flex-col overflow-y-auto bg-background text-foreground p-4 space-y-4">
      {/* Header bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <div className="flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-xl bg-primary/10 text-primary">
            <Info className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-lg font-bold text-foreground">
                {language === "th" ? "เกี่ยวกับโปรแกรม (About BC Ai Account)" : "About BC Ai Account"}
              </h1>
              <span className="rounded-md bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              {language === "th"
                ? "ช่วยเหลือ › เกี่ยวกับโปรแกรม: ข้อมูลเวอร์ชัน สถาปัตยกรรมระบบ การรับรอง และการวินิจฉัยความพร้อม"
                : "Help > About: System architecture, release version, certification, and diagnostics"}
            </p>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => setRefreshedAt(new Date().toLocaleTimeString("th-TH"))}
          className="h-8 gap-1.5 text-xs"
        >
          <RefreshCw className="h-3.5 w-3.5" />
          <span>{language === "th" ? `ตรวจสถานะ (ณ ${refreshedAt})` : `Refresh (${refreshedAt})`}</span>
        </Button>
      </div>

      {/* Main Hero Card */}
      <Card className="border-border bg-gradient-to-br from-card via-card to-primary/5">
        <CardHeader className="p-5 pb-3">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="border-primary text-primary font-bold">
                  Enterprise Cloud ERP
                </Badge>
                <span className="text-xs text-muted-foreground">High Precision Thai Accounting Standard</span>
              </div>
              <h2 className="mt-1.5 text-xl font-extrabold text-foreground">
                BC Ai Account Cloud ERP
              </h2>
              <p className="mt-1 max-w-2xl text-xs text-muted-foreground leading-relaxed">
                {language === "th"
                  ? "ระบบบัญชีและบริหารทรัพยากรองค์กรบนคลาวด์มาตรฐานสูง รองรับการทำงานเทียบเท่าโปรแกรม Desktop ERP (Champ Parity) ครอบคลุมการจัดซื้อ การขาย สินค้าคงคลัง ธนาคาร การเงิน สินทรัพย์ บัญชีแยกประเภท และแบบภาษีสรรพากรไทย"
                  : "High precision cloud ERP and accounting system with Champ Parity desktop ergonomics, covering procurement, sales, inventory, banking, assets, GL, and Thai Revenue Department tax forms."}
              </p>
            </div>

            <div className="rounded-xl border border-border/80 bg-background/80 p-3 text-right">
              <span className="text-[0.7rem] text-muted-foreground">Build Release</span>
              <div className="font-mono text-sm font-bold text-foreground">r20261010-phase8-parity-6</div>
              <span className="inline-flex items-center gap-1 text-[0.68rem] text-emerald-600 dark:text-emerald-400 font-medium">
                <CheckCircle2 className="h-3 w-3" /> Production Live
              </span>
            </div>
          </div>
        </CardHeader>

        {/* System Attributes Grid */}
        <CardContent className="p-5 pt-2">
          <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
            {systemMetrics.map((item, idx) => (
              <div key={idx} className="rounded-lg border border-border/60 bg-card p-3 space-y-1">
                <span className="text-[0.7rem] text-muted-foreground">{item.label}</span>
                <p className="text-xs font-semibold text-foreground font-mono truncate">{item.value}</p>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Architecture & Engineering Standards */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Card className="border-border">
          <CardHeader className="p-4 pb-2">
            <div className="flex items-center gap-2">
              <Server className="h-4 w-4 text-primary" />
              <CardTitle className="text-sm font-bold">สถาปัตยกรรมแบ็กเอนด์ (Backend Architecture)</CardTitle>
            </div>
            <CardDescription className="text-xs">
              ระบบประมวลผลธุรกรรมบัญชี synchronous ACID 100%
            </CardDescription>
          </CardHeader>
          <CardContent className="p-4 pt-1 space-y-2 text-xs">
            <div className="flex items-start gap-2">
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600 mt-0.5 shrink-0" />
              <div>
                <span className="font-semibold text-foreground">PostgreSQL 18 Synchronization:</span>
                <p className="text-muted-foreground text-[0.72rem]">
                  ธุรกรรมทุกรายการบันทึกผ่าน sql.Tx synchronous ACID ไร้ eventual consistency
                </p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600 mt-0.5 shrink-0" />
              <div>
                <span className="font-semibold text-foreground">Pure Go Service (CGO_ENABLED=0):</span>
                <p className="text-muted-foreground text-[0.72rem]">
                  คอมไพล์เป็นไบนารีเดียว น้ำหนักเบา ปลอดภัยสูง ไร้การพึ่งพา external C libraries
                </p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600 mt-0.5 shrink-0" />
              <div>
                <span className="font-semibold text-foreground">REST API & MCP Server Built-in:</span>
                <p className="text-muted-foreground text-[0.72rem]">
                  ให้บริการทั้ง Web Client และ AI Agent ผ่าน MCP Tools สำหรับการดึงข้อมูลและบันทึกบัญชี
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-border">
          <CardHeader className="p-4 pb-2">
            <div className="flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
              <CardTitle className="text-sm font-bold">ความปลอดภัยและมาตรฐานตัวเลข (Precision & Security)</CardTitle>
            </div>
            <CardDescription className="text-xs">
              นโยบายห้ามใช้ Floating Point (Rule 20) และระบบป้องกันข้อมูล
            </CardDescription>
          </CardHeader>
          <CardContent className="p-4 pt-1 space-y-2 text-xs">
            <div className="flex items-start gap-2">
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600 mt-0.5 shrink-0" />
              <div>
                <span className="font-semibold text-foreground">No Float Accounting (Rule 20):</span>
                <p className="text-muted-foreground text-[0.72rem]">
                  ตัวเลขบัญชีทุกจุดใช้ Minor Unit Integer (สตางค์) หรือ Numeric/Decimal เท่านั้น ปราศจากความคลาดเคลื่อน
                </p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600 mt-0.5 shrink-0" />
              <div>
                <span className="font-semibold text-foreground">Multi-tenant Isolation:</span>
                <p className="text-muted-foreground text-[0.72rem]">
                  แยกฐานข้อมูลตาม holding_code ป้องกันการเข้าถึงข้ามองค์กรอย่างเด็ดขาด
                </p>
              </div>
            </div>
            <div className="flex items-start gap-2">
              <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600 mt-0.5 shrink-0" />
              <div>
                <span className="font-semibold text-foreground">Audit Trail 100%:</span>
                <p className="text-muted-foreground text-[0.72rem]">
                  บันทึกประวัติการสร้าง แก้ไข ยกเลิก พร้อม timestamp และระบุผู้ใช้งานจริงทุกครั้ง
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Copyright Footer Card */}
      <div className="rounded-xl border border-border bg-card p-4 text-center text-xs text-muted-foreground space-y-1">
        <p className="font-medium text-foreground">
          © 2026 BC Ai Account. All Rights Reserved.
        </p>
        <p className="text-[0.72rem]">
          สงวนลิขสิทธิ์ตามพระราชบัญญัติลิขสิทธิ์ พ.ศ. 2537 และฉบับแก้ไขเพิ่มเติม
        </p>
      </div>
    </div>
  );
}
