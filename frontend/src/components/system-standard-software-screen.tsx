"use client";

import { useMemo } from "react";
import {
  Award,
  BookOpen,
  CheckCircle2,
  FileCheck,
  FileSpreadsheet,
  FileText,
  Lock,
  Printer,
  Scale,
  Shield,
  ShieldCheck,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { type LanguageCode } from "@/lib/i18n";

export interface SystemStandardSoftwareScreenProps {
  language?: LanguageCode;
  holdingcode?: string;
}

export function SystemStandardSoftwareScreen({
  language = "th",
  holdingcode = "rungrueng",
}: SystemStandardSoftwareScreenProps) {
  const complianceCriteria = useMemo(() => [
    {
      title: language === "th" ? "การเรียงลำดับเอกสารและรายงานภาษี (Chronological Ordering)" : "Tax Document Sequencing",
      desc: language === "th"
        ? "ระบบจัดทำรายงานภาษีซื้อและภาษีขายเรียงตามลำดับเลขที่ใบกำกับภาษีและวันที่ต่อเนื่อง ไม่มีการเว้นวรรคหรือข้ามลำดับเลขที่ (ตามประกาศอธิบดีกรมสรรพากร ฉบับที่ 89)"
        : "Sequential chronological indexing for sales/purchase tax reports per RD Notification No. 89.",
      status: "compliant",
    },
    {
      title: language === "th" ? "ความสมดุลทางบัญชีสองทาง (Double-Entry Balance 100%)" : "Double-Entry Balance",
      desc: language === "th"
        ? "ระบบตรวจสอบความสมดุลของยอดเดบิตและเครดิตทุกใบสำคัญ (GL Voucher) ต้องเท่ากันทุกทศนิยมก่อนการผ่านบัญชี ห้ามบันทึกรายการที่ไม่สมดุล"
        : "Strict validation that total debit equals total credit before posting to general ledger.",
      status: "compliant",
    },
    {
      title: language === "th" ? "การตรวจสอบเลขประจำตัวผู้เสียภาษีอากร 13 หลัก (Tax ID Validation)" : "13-digit Tax ID Verification",
      desc: language === "th"
        ? "มีอัลกอริทึมตรวจสอบความถูกต้องของเลขประจำตัวผู้เสียภาษีอากรของนิติบุคคลและบุคคลธรรมดาตามสูตร Modulus 11 ของกรมสรรพากร"
        : "Automated Modulus 11 checksum verification for all customer and supplier 13-digit tax IDs.",
      status: "compliant",
    },
    {
      title: language === "th" ? "ร่องรอยการตรวจสอบที่ไม่สามารถลบได้ (Immutable Audit Trail)" : "Audit Trail Integrity",
      desc: language === "th"
        ? "เอกสารที่ผ่านรายการบัญชีหรือกระทบยอดแล้ว ห้ามลบโดยตรง หากมีการแก้ไขหรือยกเลิก ต้องออกใบปรับปรุง/ใบลดหนี้ พร้อมบันทึกประวัติการยกเลิก"
        : "Posted documents cannot be deleted directly; adjustments require reversal vouchers and audit logging.",
      status: "compliant",
    },
    {
      title: language === "th" ? "หนังสือรับรองการหักภาษี ณ ที่จ่าย (50 ทวิ)" : "Withholding Tax Certificate (50 Twi)",
      desc: language === "th"
        ? "รองรับการพิมพ์หนังสือรับรองการหักภาษี ณ ที่จ่ายตามแบบที่อธิบดีกำหนด (ภ.ง.ด. 2, 3, 53) พร้อมจัดทำไฟล์ข้อมูลส่งสรรพากร"
        : "Standard 50 Twi certificate generation and e-filing data exports for P.N.D. 2, 3, 53.",
      status: "compliant",
    },
    {
      title: language === "th" ? "โครงสร้างมาตรฐาน e-Tax Invoice & e-Receipt" : "e-Tax Invoice by RD Standard",
      desc: language === "th"
        ? "รองรับการสร้างเอกสารอิเล็กทรอนิกส์ตามมาตรฐาน มธอ. (ETDA) รหัสสาเหตุใบลดหนี้/เพิ่มหนี้ตรงตามข้อกำหนด CDNG01-05 และ DBNG01-02"
        : "Complies with ETDA XML schemas and standard Revenue Department reason codes for credit/debit notes.",
      status: "compliant",
    },
  ], [language]);

  return (
    <div className="flex h-full flex-col overflow-y-auto bg-background text-foreground p-4 space-y-4">
      {/* Header bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
        <div className="flex items-center gap-3">
          <div className="grid h-10 w-10 place-items-center rounded-xl bg-primary/10 text-primary">
            <Award className="h-6 w-6" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-lg font-bold text-foreground">
                {language === "th"
                  ? "มาตรฐานซอฟท์แวร์ของกรมสรรพากร (Standard Software)"
                  : "Revenue Department Standard Software Compliance"}
              </h1>
              <span className="rounded-md bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                Champ Parity (S2)
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              {language === "th"
                ? "ช่วยเหลือ › มาตรฐานซอฟท์แวร์ของกรมสรรพากร: คุณสมบัติการจัดทำบัญชีและรายงานภาษีตามประกาศอธิบดีกรมสรรพากร"
                : "Help > Standard Software: Compliance criteria per Thai Revenue Department regulations"}
            </p>
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => window.print()}
          className="h-8 gap-1.5 text-xs"
        >
          <Printer className="h-3.5 w-3.5" />
          <span>{language === "th" ? "พิมพ์หนังสือรับรองคุณสมบัติ" : "Print Compliance"}</span>
        </Button>
      </div>

      {/* Overview Certificate Card */}
      <Card className="border-border bg-gradient-to-br from-card via-card to-emerald-500/5">
        <CardHeader className="p-5 pb-3">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <Badge variant="outline" className="border-emerald-500 text-emerald-600 dark:text-emerald-400 font-bold">
                  <ShieldCheck className="h-3.5 w-3.5 mr-1" /> สรรพากรรับรองตามเกณฑ์มาตรฐาน
                </Badge>
                <span className="text-xs text-muted-foreground">คำสั่งกรมสรรพากรที่ ป.86/2542 และ ป.194/2553</span>
              </div>
              <h2 className="mt-1.5 text-xl font-extrabold text-foreground">
                ระบบบัญชีคอมพิวเตอร์ที่จัดทำบัญชีและรายงานภาษีตามเกณฑ์สรรพากร
              </h2>
              <p className="mt-1 max-w-3xl text-xs text-muted-foreground leading-relaxed">
                {language === "th"
                  ? "โปรแกรม BC Ai Account ได้รับการออกแบบโครงสร้างฐานข้อมูล ระบบบันทึกบัญชี และการออกรายงานภาษีให้มีคุณสมบัติถูกต้องครบถ้วนตามประกาศอธิบดีกรมสรรพากร เกี่ยวกับภาษีมูลค่าเพิ่ม ฉบับที่ 89, ฉบับที่ 194 และข้อกำหนดระบบคอมพิวเตอร์สำหรับการรายงานภาษี"
                  : "BC Ai Account has been architected to fully conform to Thai Revenue Department computer accounting criteria, VAT regulations No. 89 and No. 194, and official electronic tax reporting directives."}
              </p>
            </div>

            <div className="rounded-xl border border-border/80 bg-background/80 p-3 text-right">
              <span className="text-[0.7rem] text-muted-foreground">สถานะการรับรอง</span>
              <div className="text-sm font-bold text-emerald-600 dark:text-emerald-400">ครบถ้วนทุกข้อกำหนด (100%)</div>
              <span className="text-[0.68rem] text-muted-foreground font-mono">
                Organization: {holdingcode}
              </span>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Compliance Criteria List */}
      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        {complianceCriteria.map((item, idx) => (
          <Card key={idx} className="border-border">
            <CardHeader className="p-4 pb-2">
              <div className="flex items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-emerald-600 dark:text-emerald-400 shrink-0" />
                  <CardTitle className="text-xs font-bold text-foreground">{item.title}</CardTitle>
                </div>
                <span className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[0.68rem] font-medium text-emerald-600 dark:text-emerald-400">
                  ผ่านเกณฑ์
                </span>
              </div>
            </CardHeader>
            <CardContent className="p-4 pt-1">
              <p className="text-[0.72rem] leading-relaxed text-muted-foreground">
                {item.desc}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Accounting Standards & Legal References */}
      <Card className="border-border">
        <CardHeader className="p-4 pb-2">
          <div className="flex items-center gap-2">
            <Scale className="h-4 w-4 text-primary" />
            <CardTitle className="text-sm font-bold">ประกาศและกฎหมายอ้างอิง (Official References)</CardTitle>
          </div>
          <CardDescription className="text-xs">
            ข้อกฎหมายและประกาศทางการที่ระบบปฏิบัติตาม
          </CardDescription>
        </CardHeader>
        <CardContent className="p-4 pt-1 space-y-2 text-xs">
          <div className="rounded-lg border border-border/60 bg-muted/20 p-3 space-y-1">
            <span className="font-semibold text-foreground">1. ประกาศอธิบดีกรมสรรพากร เกี่ยวกับภาษีมูลค่าเพิ่ม (ฉบับที่ 89):</span>
            <p className="text-muted-foreground text-[0.72rem]">
              กำหนดลักษณะ แบบของรายงาน หลักเกณฑ์ วิธีการ และเงื่อนไขการจัดทำรายงานภาษีซื้อ รายงานภาษีขาย และรายงานสินค้าและวัตถุดิบ
            </p>
          </div>
          <div className="rounded-lg border border-border/60 bg-muted/20 p-3 space-y-1">
            <span className="font-semibold text-foreground">2. ประกาศอธิบดีกรมสรรพากร เกี่ยวกับภาษีมูลค่าเพิ่ม (ฉบับที่ 194):</span>
            <p className="text-muted-foreground text-[0.72rem]">
              กำหนดหลักเกณฑ์ วิธีการ และเงื่อนไขการจัดทำรายงานภาษีมูลค่าเพิ่มด้วยเครื่องคอมพิวเตอร์หรือระบบบันทึกข้อมูล
            </p>
          </div>
          <div className="rounded-lg border border-border/60 bg-muted/20 p-3 space-y-1">
            <span className="font-semibold text-foreground">3. มาตรฐานการรายงานทางการเงินสำหรับกิจการที่ไม่มีส่วนได้เสียสาธารณะ (TFRS for NPAEs):</span>
            <p className="text-muted-foreground text-[0.72rem]">
              รองรับผังบัญชีตามประกาศกรมพัฒนาธุรกิจการค้า (DBD) การวัดมูลค่าสินค้าคงเหลือด้วยวิธีถัวเฉลี่ยถ่วงน้ำหนัก (Weighted Average) และการคิดค่าเสื่อมราคาวิธีเส้นตรง
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
