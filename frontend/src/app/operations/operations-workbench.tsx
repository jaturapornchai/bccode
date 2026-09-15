"use client";

import { useState } from "react";
import { getOperationsConfig } from "@/lib/erp-operations";
import type { LanguageCode } from "@/lib/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import {
  CheckCircle,
  XCircle,
  Clock,
  FileSpreadsheet,
  Upload,
  Search,
  Package,
  Boxes,
  Tag,
  ShieldCheck,
  Calendar,
  Barcode,
} from "lucide-react";

interface OperationsWorkbenchProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
}

export function OperationsWorkbench({
  route,
  embedded: _embedded = false,
  language = "th",
}: OperationsWorkbenchProps) {
  const config = getOperationsConfig(route) || {
    route,
    code: "operations_workflow",
    category: "approval" as const,
    title: { th: "ระบบงานปฏิบัติการและเวิร์กโฟลว์", en: "Operations & Workflow" },
    description: { th: "จัดการกระบวนการทำงานและเอกสาร", en: "Process & workflow management" },
    primaryActionLabel: { th: "ดำเนินการ", en: "Execute" },
  };

  const [searchTerm, setSearchTerm] = useState<string>("");
  const [feedbackNotice, setFeedbackNotice] = useState<string | null>(null);

  // Approval Sample Data
  const [approvalList, setApprovalList] = useState([
    { id: "app-1", docno: `${config.documentType || "DOC"}-202609-001`, date: "2026-09-12", requestor: "สมเกียรติ ยิ่งเจริญ", counterparty: "บริษัท สยามพาณิชย์ จำกัด", amount: 128400, status: "pending" },
    { id: "app-2", docno: `${config.documentType || "DOC"}-202609-002`, date: "2026-09-14", requestor: "นภาพร สดใส", counterparty: "ห้างหุ้นส่วนจำกัด บุญส่งการค้า", amount: 45000, status: "pending" },
  ]);

  // Serial Registry Sample Data
  const [serialList] = useState([
    { id: "sn-1", serialno: "SN-2026-990142", itemcode: "MACH-01", itemname: "เครื่องพิมพ์เลเซอร์ความเร็วสูง", status: "in_stock", warranty: "2027-09-15", customer: "-" },
    { id: "sn-2", serialno: "SN-2026-990143", itemcode: "MACH-01", itemname: "เครื่องพิมพ์เลเซอร์ความเร็วสูง", status: "sold", warranty: "2027-09-10", customer: "บจก. สยามการค้า" },
    { id: "sn-3", serialno: "SN-2026-881200", itemcode: "NOTE-PRO", itemname: "คอมพิวเตอร์พกพาสำหรับองค์กร", status: "in_stock", warranty: "2028-01-01", customer: "-" },
  ]);

  // BOM Recipe Sample Data
  const [bomRecipe] = useState([
    { itemcode: "COMP-01", itemname: "ชิ้นส่วนโครงสร้างหลัก A", reqqty: 1, unit: "ชิ้น", available: 150 },
    { itemcode: "COMP-02", itemname: "ชุดแผงควบคุมอิเล็กทรอนิกส์", reqqty: 1, unit: "ชุด", available: 85 },
    { itemcode: "COMP-03", itemname: "น็อตและหมุดยึดสแตนเลส", reqqty: 8, unit: "ตัว", available: 1200 },
  ]);

  function handleAction(docId: string, action: "approve" | "reject") {
    setApprovalList((prev) => prev.filter((item) => item.id !== docId));
    setFeedbackNotice(
      action === "approve"
        ? (language === "th" ? `อนุมัติเอกสาร ${docId} เรียบร้อยแล้ว` : `Approved ${docId}`)
        : (language === "th" ? `ปฏิเสธเอกสาร ${docId} แล้ว` : `Rejected ${docId}`),
    );
    setTimeout(() => setFeedbackNotice(null), 3500);
  }

  return (
    <div className="flex flex-col gap-5 p-4 lg:p-6 max-w-6xl mx-auto">
      {/* Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
        <div className="flex items-center gap-3">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            {config.category === "approval" ? (
              <ShieldCheck className="h-6 w-6" />
            ) : config.category === "bom" ? (
              <Boxes className="h-6 w-6" />
            ) : config.category === "import" ? (
              <FileSpreadsheet className="h-6 w-6" />
            ) : (
              <Package className="h-6 w-6" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                {language === "th" ? config.title.th : config.title.en}
              </h1>
              <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
                {config.category}
              </span>
            </div>
            <p className="text-sm text-muted-foreground">
              {language === "th" ? config.description.th : config.description.en}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {config.category === "import" && (
            <Button variant="outline" className="gap-2">
              <FileSpreadsheet className="h-4 w-4" />
              {language === "th" ? "ดาวน์โหลดไฟล์ตัวอย่าง" : "Download Template"}
            </Button>
          )}
          <Button variant="default" className="gap-2 min-h-[44px]">
            {config.category === "import" ? <Upload className="h-4 w-4" /> : <CheckCircle className="h-4 w-4" />}
            {language === "th" ? config.primaryActionLabel.th : config.primaryActionLabel.en}
          </Button>
        </div>
      </div>

      {/* Feedback Banner */}
      {feedbackNotice && (
        <div className="flex items-center gap-2.5 rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-3.5 text-emerald-700 dark:text-emerald-400 font-semibold text-sm animate-in fade-in">
          <CheckCircle className="h-5 w-5" />
          <span>{feedbackNotice}</span>
        </div>
      )}

      {/* Category Specific Content */}
      {config.category === "approval" && (
        <Card className="overflow-hidden">
          <div className="flex items-center justify-between border-b p-4 bg-muted/40">
            <div className="flex items-center gap-2 font-semibold text-sm">
              <Clock className="h-4 w-4 text-primary" />
              <span>{language === "th" ? `รายการรอการอนุมัติ (${approvalList.length} รายการ)` : `Pending Approvals (${approvalList.length})`}</span>
            </div>
            <div className="relative w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={language === "th" ? "ค้นหาเอกสาร..." : "Search..."}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-8 h-8 text-xs"
              />
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="border-b bg-muted/60 text-xs font-semibold text-muted-foreground uppercase">
                <tr>
                  <th className="px-4 py-3">เลขที่เอกสาร</th>
                  <th className="px-4 py-3">วันที่</th>
                  <th className="px-4 py-3">ผู้ขออนุมัติ</th>
                  <th className="px-4 py-3">คู่ค้า / ลูกค้า</th>
                  <th className="px-4 py-3 text-right">ยอดเงินรวม</th>
                  <th className="px-4 py-3 text-center">จัดการ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {approvalList.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-12 text-center text-muted-foreground">
                      {language === "th" ? "ไม่มีเอกสารค้างรอการอนุมัติ" : "No pending documents"}
                    </td>
                  </tr>
                ) : (
                  approvalList.map((item) => (
                    <tr key={item.id} className="hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-semibold text-primary font-mono">{item.docno}</td>
                      <td className="px-4 py-3 font-mono text-xs">{item.date}</td>
                      <td className="px-4 py-3">{item.requestor}</td>
                      <td className="px-4 py-3">{item.counterparty}</td>
                      <td className="px-4 py-3 text-right font-mono font-bold">
                        {item.amount.toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท
                      </td>
                      <td className="px-4 py-3 text-center">
                        <div className="flex items-center justify-center gap-2">
                          <Button
                            size="sm"
                            variant="default"
                            onClick={() => handleAction(item.id, "approve")}
                            className="h-8 gap-1 text-xs bg-emerald-600 hover:bg-emerald-700 text-white"
                          >
                            <CheckCircle className="h-3.5 w-3.5" />
                            {language === "th" ? "อนุมัติ" : "Approve"}
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => handleAction(item.id, "reject")}
                            className="h-8 gap-1 text-xs text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/20"
                          >
                            <XCircle className="h-3.5 w-3.5" />
                            {language === "th" ? "ไม่อนุมัติ" : "Reject"}
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {config.category === "bom" && route === "/productserialregistry" ? (
        <Card className="overflow-hidden">
          <div className="flex items-center justify-between border-b p-4 bg-muted/40">
            <div className="flex items-center gap-2 font-semibold text-sm">
              <Barcode className="h-4 w-4 text-primary" />
              <span>{language === "th" ? "ทะเบียนเลขเครื่องและระยะเวลารับประกัน (Serial Numbers)" : "Serial Registry"}</span>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="border-b bg-muted/60 text-xs font-semibold text-muted-foreground uppercase">
                <tr>
                  <th className="px-4 py-3">Serial Number</th>
                  <th className="px-4 py-3">รหัสสินค้า</th>
                  <th className="px-4 py-3">ชื่อสินค้า</th>
                  <th className="px-4 py-3 text-center">สถานะ</th>
                  <th className="px-4 py-3">วันหมดประกัน</th>
                  <th className="px-4 py-3">ลูกค้าผู้ซื้อ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {serialList.map((sn) => (
                  <tr key={sn.id} className="hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono font-bold text-primary">{sn.serialno}</td>
                    <td className="px-4 py-3 font-mono text-xs">{sn.itemcode}</td>
                    <td className="px-4 py-3 font-medium">{sn.itemname}</td>
                    <td className="px-4 py-3 text-center">
                      <span className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ${
                        sn.status === "in_stock" ? "bg-emerald-500/10 text-emerald-600" : "bg-blue-500/10 text-blue-600"
                      }`}>
                        {sn.status === "in_stock" ? (language === "th" ? "อยู่ในสต็อก" : "In Stock") : (language === "th" ? "จำหน่ายแล้ว" : "Sold")}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-mono text-xs">{sn.warranty}</td>
                    <td className="px-4 py-3 text-muted-foreground">{sn.customer}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      ) : config.category === "bom" ? (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
          <Card className="lg:col-span-1 p-5 space-y-4">
            <h2 className="font-bold text-base flex items-center gap-2">
              <Boxes className="h-4 w-4 text-primary" />
              {language === "th" ? "สินค้าชุดแม่ (Finished Kit)" : "Parent Kit"}
            </h2>
            <div className="space-y-3 text-sm">
              <div>
                <label className="text-xs text-muted-foreground block mb-1">รหัสชุดสินค้า:</label>
                <Input value="SET-PC-OFFICE" readOnly className="font-mono font-semibold" />
              </div>
              <div>
                <label className="text-xs text-muted-foreground block mb-1">ชื่อชุดสินค้า:</label>
                <Input value="ชุดคอมพิวเตอร์สำนักงานพร้อมใช้งาน" readOnly className="font-medium" />
              </div>
              <div>
                <label className="text-xs text-muted-foreground block mb-1">จำนวนที่ต้องการประกอบ/แยก:</label>
                <Input type="number" defaultValue={10} className="font-mono font-bold text-primary" />
              </div>
              <Button className="w-full mt-2 font-semibold">
                {language === "th" ? "ตรวจสอบยอดสต็อกชิ้นส่วน" : "Verify Stock Availability"}
              </Button>
            </div>
          </Card>

          <Card className="lg:col-span-2 overflow-hidden">
            <div className="border-b p-4 bg-muted/40 font-semibold text-sm">
              {language === "th" ? "ชิ้นส่วนประกอบตามสูตรการผลิต (BOM Components)" : "Components"}
            </div>
            <table className="w-full text-left text-sm">
              <thead className="border-b bg-muted/60 text-xs font-semibold text-muted-foreground uppercase">
                <tr>
                  <th className="px-4 py-3">รหัสชิ้นส่วน</th>
                  <th className="px-4 py-3">ชื่อชิ้นส่วน</th>
                  <th className="px-4 py-3 text-right">จำนวนต่อชุด</th>
                  <th className="px-4 py-3 text-right">สต็อกคงเหลือ</th>
                  <th className="px-4 py-3 text-center">ความพร้อม</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {bomRecipe.map((c, i) => (
                  <tr key={i} className="hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-xs text-primary">{c.itemcode}</td>
                    <td className="px-4 py-3 font-medium">{c.itemname}</td>
                    <td className="px-4 py-3 text-right font-mono">{c.reqqty} {c.unit}</td>
                    <td className="px-4 py-3 text-right font-mono font-semibold">{c.available} {c.unit}</td>
                    <td className="px-4 py-3 text-center">
                      <span className="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-600 font-semibold">
                        {language === "th" ? "เพียงพอ" : "Ready"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        </div>
      ) : null}

      {config.category === "import" && (
        <Card className="border-2 border-dashed border-border p-8 text-center bg-muted/10 hover:bg-muted/20 transition-colors cursor-pointer rounded-2xl">
          <div className="flex flex-col items-center justify-center gap-3">
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 text-primary">
              <Upload className="h-8 w-8" />
            </div>
            <div>
              <h3 className="text-base font-bold text-foreground">
                {language === "th" ? "ลากและวางไฟล์ Excel (.xlsx) หรือ CSV ที่นี่" : "Drag and drop your file here"}
              </h3>
              <p className="text-sm text-muted-foreground mt-1">
                {language === "th"
                  ? "รองรับไฟล์ UTF-8 ขนาดไม่เกิน 50 MB ระบบจะตรวจสอบความถูกต้องของหัวคอลัมน์ให้อัตโนมัติ"
                  : "Supports Excel and CSV UTF-8 up to 50 MB"}
              </p>
            </div>
            <Button variant="outline" className="mt-2 font-semibold">
              {language === "th" ? "คลิกเพื่อเลือกไฟล์จากคอมพิวเตอร์" : "Browse Computer Files"}
            </Button>
          </div>
        </Card>
      )}

      {config.category === "pricing" && (
        <Card className="p-5 space-y-4">
          <div className="flex items-center justify-between border-b pb-3">
            <div className="font-semibold text-base flex items-center gap-2">
              <Tag className="h-4 w-4 text-primary" />
              <span>{language === "th" ? "ปรับราคาขายเป็นกลุ่ม (Mass Price Adjustment)" : "Mass Price Update"}</span>
            </div>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            <div>
              <label className="text-xs text-muted-foreground block mb-1">กลุ่มสินค้าที่ต้องการปรับ:</label>
              <select className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
                <option>ทุกกลุ่มสินค้า</option>
                <option>สินค้าอุปโภคบริโภค</option>
                <option>เครื่องเขียนและสำนักงาน</option>
              </select>
            </div>
            <div>
              <label className="text-xs text-muted-foreground block mb-1">รูปแบบการปรับราคา:</label>
              <select className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm">
                <option>ปรับเพิ่มตามเปอร์เซ็นต์ (%)</option>
                <option>ปรับลดตามเปอร์เซ็นต์ (%)</option>
                <option>ปรับตามจำนวนเงินบาท (+/- บาท)</option>
              </select>
            </div>
            <div>
              <label className="text-xs text-muted-foreground block mb-1">ค่าที่ต้องการปรับ:</label>
              <Input type="number" defaultValue={5} className="font-mono font-bold text-primary" />
            </div>
          </div>
        </Card>
      )}

      {config.category === "reservation" && (
        <Card className="p-5 space-y-3">
          <div className="font-semibold text-base flex items-center gap-2">
            <Calendar className="h-4 w-4 text-primary" />
            <span>{language === "th" ? "ติดตามการสั่งจองและแผนจัดส่ง (Delivery Schedule & Reservations)" : "Reservations"}</span>
          </div>
          <p className="text-sm text-muted-foreground leading-relaxed">
            {language === "th"
              ? "ระบบตรวจสอบความพร้อมของสินค้าในคลัง เชื่อมโยงกับใบสั่งขาย (SO) และกำหนดคิวจัดส่งให้แก่ลูกค้า เพื่อป้องกันสินค้าขาดสต็อกและการส่งมอบล่าช้า"
              : "Synchronizes stock allocations with confirmed sales orders and logistics dispatch queues."}
          </p>
        </Card>
      )}
    </div>
  );
}
