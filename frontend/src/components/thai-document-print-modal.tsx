"use client";

import { useEffect, useRef } from "react";
import { Printer, X, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { thaiBahtText } from "@/lib/thai-baht-text";
import type { ErpTransactionDoc } from "@/lib/erp-transaction";

export type ThaiDocPrintType = "tax_invoice" | "purchase_order" | "delivery_order" | "payment_voucher";

interface ThaiDocumentPrintModalProps {
  open: boolean;
  onClose: () => void;
  doc: ErpTransactionDoc | null;
  companyInfo?: {
    name?: string;
    taxId?: string;
    branch?: string;
    address?: string;
    tel?: string;
  };
  docType?: ThaiDocPrintType;
}

export function ThaiDocumentPrintModal({
  open,
  onClose,
  doc,
  companyInfo = {
    name: "บริษัท ตัวอย่าง จำกัด",
    taxId: "0105559001234",
    branch: "สำนักงานใหญ่",
    address: "99/9 หมู่ 5 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพฯ 10110",
    tel: "02-123-4567",
  },
  docType = "tax_invoice",
}: ThaiDocumentPrintModalProps) {
  const printRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
      if ((e.ctrlKey || e.metaKey) && e.key === "p") {
        e.preventDefault();
        window.print();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open || !doc) return null;

  const docTitleMap: Record<ThaiDocPrintType, { th: string; en: string }> = {
    tax_invoice: { th: "ใบเสร็จรับเงิน / ใบกำกับภาษี", en: "RECEIPT / TAX INVOICE" },
    purchase_order: { th: "ใบสั่งซื้อสินค้า", en: "PURCHASE ORDER" },
    delivery_order: { th: "ใบส่งของ / ใบแจ้งหนี้", en: "DELIVERY ORDER / INVOICE" },
    payment_voucher: { th: "ใบสำคัญจ่าย", en: "PAYMENT VOUCHER" },
  };

  const currentTitle = docTitleMap[docType] || docTitleMap.tax_invoice;

  const totalBeforeVat = doc.totalbeforevat ?? (doc.totalamount - (doc.totalvatvalue ?? 0));
  const vatValue = doc.totalvatvalue ?? 0;
  const netTotal = doc.totalamount;
  const bahtText = thaiBahtText(netTotal);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-2 sm:p-4 backdrop-blur-sm animate-in fade-in duration-150">
      {/* Container with print styles */}
      <div className="relative flex flex-col w-full max-w-4xl max-h-[96vh] rounded-xl border border-border bg-card text-card-foreground shadow-2xl overflow-hidden">
        {/* Modal Toolbar - Hidden during print */}
        <div className="flex items-center justify-between border-b border-border bg-muted/40 px-4 py-3 shrink-0 print:hidden">
          <div className="flex items-center gap-2">
            <Printer className="size-5 text-primary" />
            <h3 className="font-semibold text-foreground">
              พิมพ์เอกสาร: {currentTitle.th} ({doc.docno})
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              onClick={() => window.print()}
              className="gap-1.5 bg-primary text-primary-foreground shadow-sm"
            >
              <Printer className="size-4" />
              พิมพ์ (Ctrl+P)
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="size-8 p-0 text-muted-foreground hover:text-foreground"
            >
              <X className="size-4" />
            </Button>
          </div>
        </div>

        {/* Printable Document Body */}
        <div className="overflow-y-auto p-6 sm:p-8 bg-white text-slate-900 print:overflow-visible print:p-0">
          <div ref={printRef} className="mx-auto max-w-[210mm] text-[13px] leading-relaxed">
            {/* 1. Header: Company Info & Title */}
            <div className="flex justify-between items-start border-b border-slate-300 pb-4">
              <div className="max-w-[55%]">
                <h1 className="text-base font-bold text-slate-900">{companyInfo.name}</h1>
                <p className="text-slate-600 mt-0.5">{companyInfo.address}</p>
                <div className="flex gap-4 mt-1 text-slate-600 text-[12px]">
                  <span>เลขประจำตัวผู้เสียภาษี: <strong className="text-slate-900">{companyInfo.taxId}</strong></span>
                  <span>({companyInfo.branch})</span>
                </div>
                {companyInfo.tel && <p className="text-slate-600 text-[12px]">โทร: {companyInfo.tel}</p>}
              </div>

              <div className="text-right">
                <div className="inline-block border-2 border-slate-900 px-3 py-1.5 text-center rounded">
                  <h2 className="text-base font-bold text-slate-900 leading-tight">{currentTitle.th}</h2>
                  <p className="text-[10px] tracking-wider text-slate-600 uppercase font-semibold">{currentTitle.en}</p>
                </div>
                <div className="mt-3 text-left inline-block space-y-1">
                  <div className="flex justify-between gap-3">
                    <span className="text-slate-500 font-medium">เลขที่ / No:</span>
                    <span className="font-bold text-slate-900">{doc.docno}</span>
                  </div>
                  <div className="flex justify-between gap-3">
                    <span className="text-slate-500 font-medium">วันที่ / Date:</span>
                    <span className="text-slate-900">{doc.docdatetime ? new Date(doc.docdatetime).toLocaleDateString("th-TH") : "—"}</span>
                  </div>
                  {doc.branchcode && (
                    <div className="flex justify-between gap-3">
                      <span className="text-slate-500 font-medium">สาขา / Branch:</span>
                      <span className="text-slate-900">{doc.branchcode}</span>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* 2. Customer / Counterparty Box */}
            <div className="my-4 rounded border border-slate-300 p-3 text-[12px] bg-slate-50/50">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="flex gap-2">
                    <span className="text-slate-500 shrink-0 font-medium">ลูกค้า/คู่ค้า:</span>
                    <strong className="text-slate-900">{doc.custname || doc.custcode || "—"}</strong>
                  </div>
                  <div className="flex gap-2 mt-1">
                    <span className="text-slate-500 shrink-0 font-medium">รหัสคู่ค้า:</span>
                    <span className="text-slate-900">{doc.custcode || "—"}</span>
                  </div>
                </div>
                <div>
                  <div className="flex gap-2">
                    <span className="text-slate-500 shrink-0 font-medium">คำอธิบาย/อ้างอิง:</span>
                    <span className="text-slate-900">{doc.description || doc.remark || "—"}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* 3. Items Table */}
            <table className="w-full border-collapse text-[12px] my-3">
              <thead>
                <tr className="border-y-2 border-slate-900 bg-slate-100 text-slate-900">
                  <th className="py-1.5 px-2 text-center w-10">ลำดับ</th>
                  <th className="py-1.5 px-2 text-left w-28">รหัสสินค้า</th>
                  <th className="py-1.5 px-2 text-left">รายการสินค้า / บริการ</th>
                  <th className="py-1.5 px-2 text-right w-16">จำนวน</th>
                  <th className="py-1.5 px-2 text-center w-16">หน่วย</th>
                  <th className="py-1.5 px-2 text-right w-24">ราคา/หน่วย</th>
                  <th className="py-1.5 px-2 text-right w-20">ส่วนลด</th>
                  <th className="py-1.5 px-2 text-right w-24">จำนวนเงิน</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-200">
                {doc.details && doc.details.length > 0 ? (
                  doc.details.map((item, idx) => (
                    <tr key={idx} className="hover:bg-slate-50">
                      <td className="py-2 px-2 text-center text-slate-500">{idx + 1}</td>
                      <td className="py-2 px-2 font-mono text-slate-700">{item.itemcode}</td>
                      <td className="py-2 px-2 font-medium text-slate-900">
                        {item.itemname || item.itemcode}
                      </td>
                      <td className="py-2 px-2 text-right font-mono tabular-nums">{item.qty?.toLocaleString()}</td>
                      <td className="py-2 px-2 text-center text-slate-600">{item.unitname || item.unitcode || "ชิ้น"}</td>
                      <td className="py-2 px-2 text-right font-mono tabular-nums">{item.price?.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                      <td className="py-2 px-2 text-right font-mono tabular-nums text-slate-600">{item.discount || "—"}</td>
                      <td className="py-2 px-2 text-right font-mono font-medium tabular-nums">{item.sumamount?.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={8} className="py-6 text-center text-slate-400">
                      ไม่มีรายการสินค้าในเอกสารนี้
                    </td>
                  </tr>
                )}
              </tbody>
            </table>

            {/* 4. Summary & BahtText */}
            <div className="flex border-t-2 border-slate-900 pt-3 gap-4">
              <div className="flex-1 flex flex-col justify-between">
                <div className="rounded border border-slate-300 p-2.5 bg-slate-50">
                  <span className="text-slate-500 text-[11px] block">จำนวนเงินตัวอักษร:</span>
                  <strong className="text-slate-900 text-sm">{bahtText}</strong>
                </div>

                <div className="text-[11px] text-slate-500 mt-2">
                  * ใบกำกับภาษีนี้สมบูรณ์เมื่อมีลายมือชื่อของผู้รับมอบอำนาจและชำระเงินครบถ้วนแล้ว
                </div>
              </div>

              <div className="w-64 space-y-1 text-[12px]">
                <div className="flex justify-between py-0.5">
                  <span className="text-slate-600">รวมมูลค่าสินค้า:</span>
                  <span className="font-mono tabular-nums">{totalBeforeVat.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between py-0.5">
                  <span className="text-slate-600">ภาษีมูลค่าเพิ่ม 7%:</span>
                  <span className="font-mono tabular-nums">{vatValue.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between py-1 border-t border-slate-900 text-sm font-bold">
                  <span className="text-slate-900">ยอดเงินสุทธิ:</span>
                  <span className="text-slate-900 font-mono tabular-nums">{netTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
              </div>
            </div>

            {/* 5. Signatures Footer */}
            <div className="grid grid-cols-3 gap-6 pt-12 pb-4 text-center text-[12px] border-t border-slate-200 mt-8">
              <div className="space-y-8">
                <div className="border-b border-slate-400 mx-4"></div>
                <p className="text-slate-700">ผู้จัดทำ (Prepared by)<br/><span className="text-slate-400 text-[11px]">วันที่ ...../...../..........</span></p>
              </div>
              <div className="space-y-8">
                <div className="border-b border-slate-400 mx-4"></div>
                <p className="text-slate-700">ผู้รับเงิน / ผู้ส่งของ<br/><span className="text-slate-400 text-[11px]">วันที่ ...../...../..........</span></p>
              </div>
              <div className="space-y-8">
                <div className="border-b border-slate-400 mx-4"></div>
                <p className="text-slate-700">ผู้อนุมัติ (Authorized Signature)<br/><span className="text-slate-400 text-[11px]">วันที่ ...../...../..........</span></p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
