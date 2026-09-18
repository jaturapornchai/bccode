"use client";

import { useState, useMemo } from "react";
import {
  Wallet,
  Plus,
  ArrowDownLeft,
  Receipt,
  FileCheck,
  CheckCircle2,
  DollarSign,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";

export interface PettyExpenseItem {
  id: string;
  date: string;
  docno: string;
  description: string;
  category: string;
  amount: number;
  receiptNo: string;
  paidTo: string;
}

const INITIAL_EXPENSES: PettyExpenseItem[] = [
  {
    id: "pe-1",
    date: "2026-09-10",
    docno: "PC-001",
    description: "ค่าน้ำมันรถส่งของด่วน กทม.-นนทบุรี",
    category: "ค่ายานพาหนะ",
    amount: 850,
    receiptNo: "PTT-9921",
    paidTo: "ปั๊ม ปตท. แจ้งวัฒนะ",
  },
  {
    id: "pe-2",
    date: "2026-09-12",
    docno: "PC-002",
    description: "ค่าส่งพัสดุด่วน Kerry Express ส่งสัญญาให้ลูกค้า",
    category: "ค่าไปรษณีย์",
    amount: 140,
    receiptNo: "KEX-8812",
    paidTo: "เคอรี่ เอ็กซ์เพรส",
  },
  {
    id: "pe-3",
    date: "2026-09-14",
    docno: "PC-003",
    description: "ค่าอาหารว่างและเครื่องดื่มรับรองลูกค้าประชุมฝ่ายขาย",
    category: "ค่ารับรอง",
    amount: 620,
    receiptNo: "S&P-1049",
    paidTo: "ร้าน S&P",
  },
  {
    id: "pe-4",
    date: "2026-09-16",
    docno: "PC-004",
    description: "ซื้ออุปกรณ์สำนักงานเบ็ดเตล็ด เทปกาวและปากกาเคมี",
    category: "เครื่องเขียน",
    amount: 390,
    receiptNo: "OFM-5521",
    paidTo: "OfficeMate",
  },
];

export function PettyCashView() {
  const [imprestFund, setImprestFund] = useState<number>(10000); // วงเงินสดย่อยคงที่
  const [custodian, setCustodian] = useState<string>("สมศรี มีทรัพย์ (แผนกธุรการ)");
  const [expenses, setExpenses] = useState<PettyExpenseItem[]>(INITIAL_EXPENSES);
  const [replenishNotice, setReplenishNotice] = useState<string | null>(null);

  // Totals
  const totalSpent = useMemo(() => expenses.reduce((sum, e) => sum + e.amount, 0), [expenses]);
  const cashOnHand = imprestFund - totalSpent;
  const spentPercent = Math.min(100, Math.round((totalSpent / imprestFund) * 100));

  const handleReplenish = () => {
    setReplenishNotice(
      `สร้างใบขอเบิกชดเชยเงินสดย่อยเรียบร้อยแล้ว จำนวน ฿${totalSpent.toLocaleString("th-TH", {
        minimumFractionDigits: 2,
      })} (โอนเข้าบัญชีผู้ถือเงินสดย่อยเพื่อเติมเต็มวงเงิน ฿${imprestFund.toLocaleString()})`
    );
  };

  return (
    <div className="flex flex-col gap-4">
      {/* 3 KPI Cards */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <Card className="border-border/60 bg-card">
          <CardContent className="p-4">
            <span className="text-xs font-medium text-muted-foreground">วงเงินสดย่อยที่กำหนด (Imprest Fund)</span>
            <div className="mt-2 text-xl font-bold font-mono text-foreground">
              ฿{imprestFund.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">ผู้ดูแลวงเงิน: {custodian}</p>
          </CardContent>
        </Card>

        <Card className="border-border/60 bg-card">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">ยอดจ่ายไปแล้วรอบนี้ ({expenses.length} รายการ)</span>
              <Receipt className="size-4 text-amber-600" />
            </div>
            <div className="mt-2 text-xl font-bold font-mono text-amber-600">
              ฿{totalSpent.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <div className="mt-2 flex items-center gap-2">
              <div className="h-1.5 flex-1 rounded-full bg-muted overflow-hidden">
                <div className="h-full bg-amber-500 rounded-full" style={{ width: `${spentPercent}%` }} />
              </div>
              <span className="text-[11px] font-mono text-muted-foreground">{spentPercent}%</span>
            </div>
          </CardContent>
        </Card>

        <Card className="border-emerald-500/30 bg-emerald-500/5">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-emerald-700 dark:text-emerald-400">เงินสดคงเหลือในกล่อง</span>
              <Wallet className="size-4 text-emerald-600" />
            </div>
            <div className="mt-2 text-xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
              ฿{cashOnHand.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-xs text-emerald-700/80 dark:text-emerald-400/80">
              {cashOnHand < imprestFund * 0.3 ? "⚠️ เงินสดใกล้หมด ควรทำเบิกชดเชย" : "✓ พร้อมสำหรับการเบิกจ่ายประจำวัน"}
            </p>
          </CardContent>
        </Card>
      </div>

      {replenishNotice && (
        <div className="flex items-center gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-sm text-emerald-800">
          <CheckCircle2 className="size-4 text-emerald-600 shrink-0" />
          <span>{replenishNotice}</span>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-card p-3">
        <div className="flex items-center gap-2">
          <Receipt className="size-4 text-primary" />
          <h3 className="font-bold text-sm text-foreground">รายการใบเสร็จเงินสดย่อยที่รอเบิกชดเชย</h3>
        </div>

        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={handleReplenish}
            className="gap-1.5 bg-primary text-primary-foreground shadow-sm"
          >
            <FileCheck className="size-4" />
            ทำใบเบิกชดเชยเงินสดย่อย (฿{totalSpent.toLocaleString()})
          </Button>
        </div>
      </div>

      {/* Expense List Table */}
      <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
        <table className="w-full text-left text-xs border-collapse">
          <thead className="bg-muted/60 text-muted-foreground font-semibold border-b border-border">
            <tr>
              <th className="p-2.5 w-10 text-center">#</th>
              <th className="p-2.5 w-24">วันที่</th>
              <th className="p-2.5 w-28">เลขที่เอกสาร</th>
              <th className="p-2.5 w-32">หมวดค่าใช้จ่าย</th>
              <th className="p-2.5">รายละเอียดการจ่าย</th>
              <th className="p-2.5 w-36">จ่ายให้ (ผู้รับเงิน)</th>
              <th className="p-2.5 w-28">เลขที่ใบเสร็จ</th>
              <th className="p-2.5 text-right w-28">จำนวนเงิน</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 font-mono">
            {expenses.map((e, idx) => (
              <tr key={e.id} className="hover:bg-muted/20 transition-colors">
                <td className="p-2.5 text-center text-muted-foreground">{idx + 1}</td>
                <td className="p-2.5 text-muted-foreground">{e.date}</td>
                <td className="p-2.5 font-bold text-primary">{e.docno}</td>
                <td className="p-2.5 font-sans">
                  <span className="rounded bg-muted px-2 py-0.5 text-[11px] font-medium text-foreground">
                    {e.category}
                  </span>
                </td>
                <td className="p-2.5 font-sans font-medium text-foreground">{e.description}</td>
                <td className="p-2.5 font-sans text-muted-foreground">{e.paidTo}</td>
                <td className="p-2.5 text-muted-foreground">{e.receiptNo}</td>
                <td className="p-2.5 text-right font-bold text-foreground">
                  ฿{e.amount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
