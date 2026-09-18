"use client";

import { useState, useMemo } from "react";
import {
  CreditCard,
  CheckCircle2,
  AlertTriangle,
  ArrowDownLeft,
  ArrowUpRight,
  Clock,
  Ban,
  Plus,
  Search,
  Filter,
  RefreshCw,
  Building2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";

export interface ChequeRecord {
  id: string;
  chequeNo: string;
  type: "received" | "issued"; // received = เช็ครับ, issued = เช็คจ่าย
  date: string;
  dueDate: string;
  partyCode: string;
  partyName: string;
  bankName: string;
  bankBranch: string;
  amount: number;
  status: "on_hand" | "deposited" | "cleared" | "returned" | "cancelled";
  depositAccount?: string;
  returnReason?: string;
}

const INITIAL_CHEQUES: ChequeRecord[] = [
  {
    id: "chq-1",
    chequeNo: "9041284",
    type: "received",
    date: "2026-09-08",
    dueDate: "2026-09-15",
    partyCode: "AR-001",
    partyName: "บจก. สยามพาณิชย์การค้า (ลูกหนี้)",
    bankName: "กสิกรไทย",
    bankBranch: "สาขาสยามสแควร์",
    amount: 85000,
    status: "cleared",
    depositAccount: "111201 KBank ออมทรัพย์",
  },
  {
    id: "chq-2",
    chequeNo: "8834192",
    type: "received",
    date: "2026-09-12",
    dueDate: "2026-09-20",
    partyCode: "AR-004",
    partyName: "หจก. ทรัพย์อนันต์รุ่งเรือง",
    bankName: "ไทยพาณิชย์",
    bankBranch: "สาขาอโศก",
    amount: 42500,
    status: "deposited",
    depositAccount: "111201 KBank ออมทรัพย์",
  },
  {
    id: "chq-3",
    chequeNo: "5512093",
    type: "received",
    date: "2026-09-05",
    dueDate: "2026-09-10",
    partyCode: "AR-009",
    partyName: "บจก. เอเชีย โลจิสติกส์ มาร์เก็ตติ้ง",
    bankName: "กรุงเทพ",
    bankBranch: "สาขาสีลม",
    amount: 60000,
    status: "returned",
    returnReason: "เงินในบัญชีไม่พอจ่าย (เช็คเด้ง)",
  },
  {
    id: "chq-4",
    chequeNo: "1094821",
    type: "received",
    date: "2026-09-16",
    dueDate: "2026-09-25",
    partyCode: "AR-012",
    partyName: "ร้าน ป.พาณิชย์พัฒนา",
    bankName: "กรุงไทย",
    bankBranch: "สาขาราชเทวี",
    amount: 19500,
    status: "on_hand",
  },
  {
    id: "chq-5",
    chequeNo: "3309182",
    type: "issued",
    date: "2026-09-10",
    dueDate: "2026-09-18",
    partyCode: "AP-002",
    partyName: "บจก. ไทยเปเปอร์ ดิสทริบิวชั่น (เจ้าหนี้)",
    bankName: "กสิกรไทย",
    bankBranch: "บัญชีกระแสรายวันบริษัท",
    amount: 55000,
    status: "cleared",
  },
  {
    id: "chq-6",
    chequeNo: "3309183",
    type: "issued",
    date: "2026-09-14",
    dueDate: "2026-09-28",
    partyCode: "AP-005",
    partyName: "บจก. เมโทร ออฟฟิศ ซัพพลาย",
    bankName: "กสิกรไทย",
    bankBranch: "บัญชีกระแสรายวันบริษัท",
    amount: 28400,
    status: "on_hand",
  },
];

export function ChequeManagementView() {
  const [cheques, setCheques] = useState<ChequeRecord[]>(INITIAL_CHEQUES);
  const [activeType, setActiveType] = useState<"all" | "received" | "issued">("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const filteredCheques = useMemo(() => {
    return cheques.filter((c) => {
      if (activeType !== "all" && c.type !== activeType) return false;
      if (statusFilter !== "all" && c.status !== statusFilter) return false;
      if (!searchQuery.trim()) return true;
      const q = searchQuery.toLowerCase();
      return (
        c.chequeNo.toLowerCase().includes(q) ||
        c.partyName.toLowerCase().includes(q) ||
        c.bankName.toLowerCase().includes(q)
      );
    });
  }, [cheques, activeType, statusFilter, searchQuery]);

  // Status updates
  const handleUpdateStatus = (id: string, nextStatus: ChequeRecord["status"], reason?: string) => {
    setCheques((prev) =>
      prev.map((c) =>
        c.id === id ? { ...c, status: nextStatus, ...(reason ? { returnReason: reason } : {}) } : c
      )
    );
  };

  // KPI Calculations
  const stats = useMemo(() => {
    const rcv = cheques.filter((c) => c.type === "received");
    const iss = cheques.filter((c) => c.type === "issued");
    const onHandRcv = rcv.filter((c) => c.status === "on_hand" || c.status === "deposited");
    const returned = rcv.filter((c) => c.status === "returned");
    const pendingIss = iss.filter((c) => c.status === "on_hand");

    return {
      onHandTotal: onHandRcv.reduce((sum, c) => sum + c.amount, 0),
      onHandCount: onHandRcv.length,
      returnedTotal: returned.reduce((sum, c) => sum + c.amount, 0),
      returnedCount: returned.length,
      pendingIssTotal: pendingIss.reduce((sum, c) => sum + c.amount, 0),
      pendingIssCount: pendingIss.length,
    };
  }, [cheques]);

  return (
    <div className="flex flex-col gap-4">
      {/* 3 KPI Summary Cards */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <Card className="border-border/60 bg-card">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">เช็ครับรอขึ้นเงิน / ฝากแล้ว</span>
              <ArrowDownLeft className="size-4 text-blue-600" />
            </div>
            <div className="mt-2 text-xl font-bold font-mono text-blue-600">
              ฿{stats.onHandTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">{stats.onHandCount} ฉบับ อยู่ระหว่างรอเงินเข้าบัญชี</p>
          </CardContent>
        </Card>

        <Card className="border-destructive/30 bg-destructive/5">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-destructive">เช็ครับคืน (เช็คเด้งต้องตามทวง)</span>
              <AlertTriangle className="size-4 text-destructive" />
            </div>
            <div className="mt-2 text-xl font-bold font-mono text-destructive">
              ฿{stats.returnedTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-xs text-destructive/80">{stats.returnedCount} ฉบับ สรรพากร/ธนาคารคืนรายการ</p>
          </CardContent>
        </Card>

        <Card className="border-border/60 bg-card">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-muted-foreground">เช็คจ่ายรอเจ้าหนี้ขึ้นเงิน</span>
              <ArrowUpRight className="size-4 text-amber-600" />
            </div>
            <div className="mt-2 text-xl font-bold font-mono text-amber-600">
              ฿{stats.pendingIssTotal.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">{stats.pendingIssCount} ฉบับ ยังไม่ตัดบัญชีกระแสรายวัน</p>
          </CardContent>
        </Card>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-card p-3">
        <div className="flex flex-wrap items-center gap-2">
          {/* Cheque Type Toggle */}
          <div className="inline-flex rounded-lg border border-border bg-muted/40 p-0.5 text-xs">
            <button
              onClick={() => setActiveType("all")}
              className={`px-3 py-1.5 rounded-md font-medium transition-colors ${
                activeType === "all" ? "bg-background text-foreground shadow-sm font-semibold" : "text-muted-foreground"
              }`}
            >
              ทั้งหมด ({cheques.length})
            </button>
            <button
              onClick={() => setActiveType("received")}
              className={`px-3 py-1.5 rounded-md font-medium transition-colors ${
                activeType === "received" ? "bg-background text-primary shadow-sm font-semibold" : "text-muted-foreground"
              }`}
            >
              เช็ครับ (Cheque In)
            </button>
            <button
              onClick={() => setActiveType("issued")}
              className={`px-3 py-1.5 rounded-md font-medium transition-colors ${
                activeType === "issued" ? "bg-background text-amber-600 shadow-sm font-semibold" : "text-muted-foreground"
              }`}
            >
              เช็คจ่าย (Cheque Out)
            </button>
          </div>

          <div className="relative w-56">
            <Search className="absolute left-2.5 top-2 size-3.5 text-muted-foreground" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="ค้นหาเลขที่เช็ค / ชื่อคู่ค้า..."
              className="h-8 pl-8 text-xs"
            />
          </div>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground"
          >
            <option value="all">ทุกสถานะเช็ค</option>
            <option value="on_hand">ถือเช็คอยู่ / รอฝาก</option>
            <option value="deposited">นำฝากแล้ว</option>
            <option value="cleared">เช็คผ่านแล้ว</option>
            <option value="returned">เช็คคืน (เช็คเด้ง)</option>
          </select>
        </div>
      </div>

      {/* Cheques Table */}
      <div className="overflow-x-auto rounded-xl border border-border bg-card shadow-sm">
        <table className="w-full text-left text-xs border-collapse">
          <thead className="bg-muted/60 text-muted-foreground font-semibold border-b border-border">
            <tr>
              <th className="p-2.5 w-24">ประเภท</th>
              <th className="p-2.5 w-28">เลขที่เช็ค</th>
              <th className="p-2.5 w-24">วันที่บนเช็ค</th>
              <th className="p-2.5">คู่ค้า (ลูกค้า / เจ้าหนี้)</th>
              <th className="p-2.5 w-36">ธนาคาร / สาขา</th>
              <th className="p-2.5 text-right w-32">ยอดเงินบนเช็ค</th>
              <th className="p-2.5 text-center w-28">สถานะ</th>
              <th className="p-2.5 text-center w-40">การดำเนินการ</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 font-mono">
            {filteredCheques.map((c) => {
              const isRcv = c.type === "received";
              const isReturned = c.status === "returned";
              const isCleared = c.status === "cleared";

              return (
                <tr key={c.id} className={`hover:bg-muted/20 transition-colors ${isReturned ? "bg-destructive/5" : ""}`}>
                  <td className="p-2.5 font-sans">
                    <span
                      className={`inline-flex items-center gap-1 rounded px-2 py-0.5 text-[11px] font-semibold ${
                        isRcv ? "bg-blue-500/10 text-blue-600" : "bg-amber-500/10 text-amber-600"
                      }`}
                    >
                      {isRcv ? "เช็ครับ" : "เช็คจ่าย"}
                    </span>
                  </td>
                  <td className="p-2.5 font-bold text-foreground">{c.chequeNo}</td>
                  <td className="p-2.5 text-muted-foreground">{c.dueDate}</td>
                  <td className="p-2.5 font-sans font-medium text-foreground">
                    <div>{c.partyName}</div>
                    {c.returnReason && (
                      <div className="text-[11px] text-destructive font-normal mt-0.5 flex items-center gap-1">
                        <AlertTriangle className="size-3" />
                        {c.returnReason}
                      </div>
                    )}
                  </td>
                  <td className="p-2.5 font-sans text-muted-foreground">
                    {c.bankName} <span className="text-[10px] block text-slate-400">{c.bankBranch}</span>
                  </td>
                  <td className="p-2.5 text-right font-bold text-foreground">
                    ฿{c.amount.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </td>
                  <td className="p-2.5 text-center font-sans">
                    <Badge
                      variant="outline"
                      className={`text-xs ${
                        isCleared
                          ? "border-emerald-500/40 text-emerald-600 bg-emerald-500/10"
                          : isReturned
                            ? "border-destructive text-destructive bg-destructive/10 font-bold"
                            : c.status === "deposited"
                              ? "border-blue-500/40 text-blue-600 bg-blue-500/10"
                              : "border-slate-300 text-slate-600"
                      }`}
                    >
                      {c.status === "cleared"
                        ? "ผ่านแล้ว"
                        : c.status === "returned"
                          ? "เช็คเด้ง (คืน)"
                          : c.status === "deposited"
                            ? "นำฝากแล้ว"
                            : "ถืออยู่"}
                    </Badge>
                  </td>
                  <td className="p-2.5 text-center font-sans">
                    <div className="flex items-center justify-center gap-1.5">
                      {c.status !== "cleared" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleUpdateStatus(c.id, "cleared")}
                          className="h-7 px-2 text-xs text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
                          title="บันทึกเช็คผ่าน"
                        >
                          <CheckCircle2 className="size-3.5 mr-1" />
                          ผ่าน
                        </Button>
                      )}
                      {isRcv && c.status !== "returned" && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            const reason = window.prompt("ระบุสาเหตุเช็คคืน (เช็คเด้ง):", "เงินในบัญชีไม่พอจ่าย");
                            if (reason) handleUpdateStatus(c.id, "returned", reason);
                          }}
                          className="h-7 px-2 text-xs text-destructive hover:bg-destructive/10"
                          title="บันทึกเช็คเด้ง"
                        >
                          <AlertTriangle className="size-3.5 mr-1" />
                          เช็คเด้ง
                        </Button>
                      )}
                    </div>
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
