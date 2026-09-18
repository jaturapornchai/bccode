"use client";

import { useState, useMemo } from "react";
import {
  Boxes,
  Printer,
  Download,
  Search,
  ArrowDownLeft,
  ArrowUpRight,
  Package,
  Calendar,
  Warehouse,
  Filter,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  calculateStockCard,
  type RawStockMovement,
  type StockCardSummary,
} from "@/lib/inventory-stock-card";

// Initial sample movements for immediate viewing
const INITIAL_MOVEMENTS: RawStockMovement[] = [
  {
    id: "m1",
    date: "2026-09-01",
    docno: "PO2026-09001",
    taxno: "INV-9901",
    docType: "purchase",
    description: "บจก. ซัพพลายไทย อินดัสทรี (ซื้อสินค้าล็อต 1)",
    itemcode: "SKU-001",
    itemname: "กระดาษถ่ายเอกสาร A4 80g Double A",
    unitcode: "รีม",
    warehousecode: "WH-01",
    qty: 200,
    unitCost: 110,
    totalAmount: 22000,
  },
  {
    id: "m2",
    date: "2026-09-04",
    docno: "INV2026-09012",
    taxno: "TAX-202609012",
    docType: "sale",
    description: "บจก. ก้าวหน้าการค้า (ขายเชื่อ)",
    itemcode: "SKU-001",
    itemname: "กระดาษถ่ายเอกสาร A4 80g Double A",
    unitcode: "รีม",
    warehousecode: "WH-01",
    qty: 60,
  },
  {
    id: "m3",
    date: "2026-09-08",
    docno: "INV2026-09025",
    taxno: "TAX-202609025",
    docType: "sale",
    description: "ร้านค้าสวัสดิการ กทม. (ขายสด)",
    itemcode: "SKU-001",
    itemname: "กระดาษถ่ายเอกสาร A4 80g Double A",
    unitcode: "รีม",
    warehousecode: "WH-01",
    qty: 40,
  },
  {
    id: "m4",
    date: "2026-09-12",
    docno: "PO2026-09008",
    taxno: "INV-9985",
    docType: "purchase",
    description: "บจก. สยามเปเปอร์ แมนูแฟคเจอริ่ง (ซื้อล็อต 2)",
    itemcode: "SKU-001",
    itemname: "กระดาษถ่ายเอกสาร A4 80g Double A",
    unitcode: "รีม",
    warehousecode: "WH-01",
    qty: 150,
    unitCost: 115,
    totalAmount: 17250,
  },
  {
    id: "m5",
    date: "2026-09-15",
    docno: "ISSUE2026-004",
    docType: "issue",
    description: "เบิกใช้ภายในแผนกบัญชีและการตลาด",
    itemcode: "SKU-001",
    itemname: "กระดาษถ่ายเอกสาร A4 80g Double A",
    unitcode: "รีม",
    warehousecode: "WH-01",
    qty: 10,
  },
];

export function StockCardScreen() {
  const [selectedItem, setSelectedItem] = useState("SKU-001");
  const [selectedWarehouse, setSelectedWarehouse] = useState("WH-01");
  const [fromDate, setFromDate] = useState("2026-09-01");
  const [toDate, setToDate] = useState("2026-09-30");
  const [beginningQty, setBeginningQty] = useState(50);
  const [beginningCost, setBeginningCost] = useState(105);

  const stockSummary: StockCardSummary = useMemo(() => {
    return calculateStockCard(selectedItem, INITIAL_MOVEMENTS, {
      itemname: "กระดาษถ่ายเอกสาร A4 80g Double A",
      unitcode: "รีม",
      warehousecode: selectedWarehouse,
      fromDate,
      toDate,
      beginningQty,
      beginningCost,
    });
  }, [selectedItem, selectedWarehouse, fromDate, toDate, beginningQty, beginningCost]);

  const handleExportCsv = () => {
    const headers = [
      "วันที่",
      "เลขที่เอกสาร",
      "เลขที่ใบกำกับ",
      "ประเภท",
      "รายละเอียด",
      "รับ_จำนวน",
      "รับ_ต้นทุน",
      "รับ_มูลค่า",
      "จ่าย_จำนวน",
      "จ่าย_ต้นทุน",
      "จ่าย_มูลค่า",
      "คงเหลือ_จำนวน",
      "คงเหลือ_ต้นทุนเฉลี่ย",
      "คงเหลือ_มูลค่า",
    ];

    const rows = stockSummary.lines.map((l) => [
      l.date,
      l.docno,
      l.taxno,
      l.docTypeName,
      `"${(l.description || "").replace(/"/g, '""')}"`,
      l.inQty || "",
      l.inCost ? l.inCost.toFixed(2) : "",
      l.inTotal ? l.inTotal.toFixed(2) : "",
      l.outQty || "",
      l.outCost ? l.outCost.toFixed(2) : "",
      l.outTotal ? l.outTotal.toFixed(2) : "",
      l.balanceQty,
      l.balanceCost.toFixed(2),
      l.balanceTotal.toFixed(2),
    ]);

    const csvContent = "\uFEFF" + [headers.join(","), ...rows.map((r) => r.join(","))].join("\n");
    const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.setAttribute("download", `StockCard_${selectedItem}_${fromDate}_to_${toDate}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  return (
    <div className="mx-auto flex max-w-[1800px] flex-col gap-4 p-4 text-[0.95rem]">
      {/* Top Header */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border/60 pb-3">
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Boxes className="size-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">
                รายงานสินค้าและวัตถุดิบ (บัญชีคุมพิเศษสินค้า / Stock Card)
              </h1>
              <span className="rounded bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">
                มาตรา 87(3) สรรพากร
              </span>
            </div>
            <p className="text-xs text-muted-foreground">
              ถอดแบบระบบ BC Champ IC (ICRepSpecialAccountView) แสดงการเคลื่อนไหว เข้า-ออก-คงเหลือ-ต้นทุนถัวเฉลี่ย
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 print:hidden">
          <Button variant="outline" size="sm" onClick={handleExportCsv} className="gap-1.5">
            <Download className="size-4 text-emerald-600" />
            ส่งออก CSV
          </Button>
          <Button
            size="sm"
            onClick={() => window.print()}
            className="gap-1.5 bg-primary text-primary-foreground shadow-sm"
          >
            <Printer className="size-4" />
            พิมพ์รายงาน (Ctrl+P)
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <Card className="border-border/60 bg-card print:hidden">
        <CardContent className="grid grid-cols-1 gap-3 p-3 sm:grid-cols-2 lg:grid-cols-5">
          <div>
            <label className="mb-1 block text-xs font-semibold text-muted-foreground">
              รหัสสินค้า / สินค้า
            </label>
            <Input
              value={selectedItem}
              onChange={(e) => setSelectedItem(e.target.value)}
              className="h-8 text-xs font-mono"
              placeholder="ระบุรหัสสินค้า"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-semibold text-muted-foreground">
              คลังสินค้า
            </label>
            <Input
              value={selectedWarehouse}
              onChange={(e) => setSelectedWarehouse(e.target.value)}
              className="h-8 text-xs font-mono"
              placeholder="ระบุคลัง เช่น WH-01"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-semibold text-muted-foreground">
              ตั้งแต่วันที่
            </label>
            <Input
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
              className="h-8 text-xs"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-semibold text-muted-foreground">
              ถึงวันที่
            </label>
            <Input
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
              className="h-8 text-xs"
            />
          </div>

          <div className="flex items-end gap-2">
            <div className="flex-1">
              <label className="mb-1 block text-xs font-semibold text-muted-foreground">
                ยอดยกมาต้นงวด
              </label>
              <Input
                type="number"
                value={beginningQty}
                onChange={(e) => setBeginningQty(Number(e.target.value))}
                className="h-8 text-xs font-mono"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Summary KPI Cards */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 print:grid-cols-4">
        <Card className="border-border/60">
          <CardContent className="p-3">
            <span className="text-xs text-muted-foreground">ยอดยกมาต้นงวด</span>
            <div className="mt-1 text-lg font-bold font-mono">
              {stockSummary.beginningQty.toLocaleString()} {stockSummary.unitcode}
            </div>
            <p className="text-[11px] text-muted-foreground">
              มูลค่า: ฿{stockSummary.beginningTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border/60">
          <CardContent className="p-3">
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">รวมรับเข้า (IN)</span>
              <ArrowDownLeft className="size-3.5 text-blue-600" />
            </div>
            <div className="mt-1 text-lg font-bold font-mono text-blue-600">
              +{stockSummary.totalInQty.toLocaleString()} {stockSummary.unitcode}
            </div>
            <p className="text-[11px] text-muted-foreground">
              มูลค่า: ฿{stockSummary.totalInTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border/60">
          <CardContent className="p-3">
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">รวมจ่ายออก (OUT)</span>
              <ArrowUpRight className="size-3.5 text-amber-600" />
            </div>
            <div className="mt-1 text-lg font-bold font-mono text-amber-600">
              -{stockSummary.totalOutQty.toLocaleString()} {stockSummary.unitcode}
            </div>
            <p className="text-[11px] text-muted-foreground">
              มูลค่า: ฿{stockSummary.totalOutTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </p>
          </CardContent>
        </Card>

        <Card className="border-border/60 bg-emerald-500/5 border-emerald-500/30">
          <CardContent className="p-3">
            <span className="text-xs font-semibold text-emerald-700 dark:text-emerald-400">
              ยอดคงเหลือสิ้นงวด (Balance)
            </span>
            <div className="mt-1 text-lg font-bold font-mono text-emerald-600 dark:text-emerald-400">
              {stockSummary.endingQty.toLocaleString()} {stockSummary.unitcode}
            </div>
            <p className="text-[11px] text-emerald-700/80 dark:text-emerald-400/80">
              ทุนเฉลี่ย: ฿{stockSummary.endingCost.toFixed(2)} | รวม: ฿{stockSummary.endingTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* 14-Column Stock Card Table (Revenue Dept Format) */}
      <div className="overflow-x-auto rounded-xl border border-border/80 bg-card shadow-sm">
        <table className="w-full text-left text-xs border-collapse">
          {/* Main Table Headers */}
          <thead>
            <tr className="border-b border-border bg-muted/60 text-muted-foreground font-semibold">
              <th rowSpan={2} className="p-2 border-r border-border/60 text-center w-24">วันที่</th>
              <th rowSpan={2} className="p-2 border-r border-border/60 w-28">เลขที่เอกสาร</th>
              <th rowSpan={2} className="p-2 border-r border-border/60 w-28">เลขที่ใบกำกับ</th>
              <th rowSpan={2} className="p-2 border-r border-border/60 w-28">ประเภท</th>
              <th rowSpan={2} className="p-2 border-r border-border/60 min-w-[200px]">รายละเอียด / คู่ค้า</th>
              
              {/* Group IN */}
              <th colSpan={3} className="p-1.5 border-r border-border/60 text-center bg-blue-500/10 text-blue-700 dark:text-blue-300">
                รับเข้า (IN)
              </th>

              {/* Group OUT */}
              <th colSpan={3} className="p-1.5 border-r border-border/60 text-center bg-amber-500/10 text-amber-700 dark:text-amber-300">
                จ่ายออก (OUT)
              </th>

              {/* Group BALANCE */}
              <th colSpan={3} className="p-1.5 text-center bg-emerald-500/10 text-emerald-700 dark:text-emerald-300">
                คงเหลือสะสม (BALANCE)
              </th>
            </tr>
            <tr className="border-b border-border text-[11px] font-medium text-muted-foreground">
              {/* IN Subheaders */}
              <th className="p-1.5 text-right w-16 bg-blue-500/5">จำนวน</th>
              <th className="p-1.5 text-right w-20 bg-blue-500/5">ต้นทุน/หน.</th>
              <th className="p-1.5 text-right w-24 border-r border-border/60 bg-blue-500/5">มูลค่า</th>

              {/* OUT Subheaders */}
              <th className="p-1.5 text-right w-16 bg-amber-500/5">จำนวน</th>
              <th className="p-1.5 text-right w-20 bg-amber-500/5">ต้นทุน/หน.</th>
              <th className="p-1.5 text-right w-24 border-r border-border/60 bg-amber-500/5">มูลค่า</th>

              {/* BALANCE Subheaders */}
              <th className="p-1.5 text-right w-16 bg-emerald-500/5">จำนวน</th>
              <th className="p-1.5 text-right w-20 bg-emerald-500/5">ทุนเฉลี่ย</th>
              <th className="p-1.5 text-right w-24 bg-emerald-500/5">มูลค่ารวม</th>
            </tr>
          </thead>

          <tbody className="divide-y divide-border/40 font-mono">
            {stockSummary.lines.map((line, idx) => {
              const isBeginning = line.docType === "beginning";
              return (
                <tr
                  key={idx}
                  className={`hover:bg-muted/30 transition-colors ${
                    isBeginning ? "bg-muted/20 font-semibold" : ""
                  }`}
                >
                  <td className="p-2 border-r border-border/40 text-center text-muted-foreground">
                    {line.date}
                  </td>
                  <td className="p-2 border-r border-border/40 font-semibold text-primary">
                    {line.docno}
                  </td>
                  <td className="p-2 border-r border-border/40 text-muted-foreground">
                    {line.taxno}
                  </td>
                  <td className="p-2 border-r border-border/40 font-sans text-foreground">
                    {line.docTypeName}
                  </td>
                  <td className="p-2 border-r border-border/40 font-sans text-muted-foreground">
                    {line.description}
                  </td>

                  {/* IN Columns */}
                  <td className="p-2 text-right bg-blue-500/5 font-semibold text-blue-600">
                    {line.inQty ? line.inQty.toLocaleString() : "—"}
                  </td>
                  <td className="p-2 text-right bg-blue-500/5 text-muted-foreground">
                    {line.inCost ? line.inCost.toFixed(2) : "—"}
                  </td>
                  <td className="p-2 text-right border-r border-border/40 bg-blue-500/5 text-foreground">
                    {line.inTotal ? line.inTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 }) : "—"}
                  </td>

                  {/* OUT Columns */}
                  <td className="p-2 text-right bg-amber-500/5 font-semibold text-amber-600">
                    {line.outQty ? line.outQty.toLocaleString() : "—"}
                  </td>
                  <td className="p-2 text-right bg-amber-500/5 text-muted-foreground">
                    {line.outCost ? line.outCost.toFixed(2) : "—"}
                  </td>
                  <td className="p-2 text-right border-r border-border/40 bg-amber-500/5 text-foreground">
                    {line.outTotal ? line.outTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 }) : "—"}
                  </td>

                  {/* BALANCE Columns */}
                  <td className="p-2 text-right bg-emerald-500/5 font-bold text-foreground">
                    {line.balanceQty.toLocaleString()}
                  </td>
                  <td className="p-2 text-right bg-emerald-500/5 text-emerald-700 dark:text-emerald-400">
                    {line.balanceCost.toFixed(2)}
                  </td>
                  <td className="p-2 text-right bg-emerald-500/5 font-bold text-emerald-700 dark:text-emerald-400">
                    {line.balanceTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
