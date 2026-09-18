"use client";

import { useMemo } from "react";
import { useBackendText } from "@/components/backend-text-provider";
import { Card, CardContent } from "@/components/ui/card";
import {
  TrendingUp,
  Activity,
  ShieldCheck,
  AlertTriangle,
  AlertCircle,
  Clock,
  ArrowUpRight,
  ArrowDownRight,
  DollarSign,
  PieChart,
} from "lucide-react";
import {
  calculateCFOFinancialHealth,
  calculateCashFlowStatement,
  type CFOBalanceMetrics,
} from "@/lib/cfo-financial-health";

interface CFODashboardViewProps {
  metrics?: Partial<CFOBalanceMetrics>;
}

export function CFODashboardView({ metrics }: CFODashboardViewProps) {
  const tr = useBackendText();

  const standardMetrics: CFOBalanceMetrics = useMemo(
    () => ({
      cashAndBank: metrics?.cashAndBank ?? 1450000,
      tradeReceivables: metrics?.tradeReceivables ?? 890000,
      inventory: metrics?.inventory ?? 620000,
      otherCurrentAssets: metrics?.otherCurrentAssets ?? 110000,
      nonCurrentAssets: metrics?.nonCurrentAssets ?? 3200000,
      totalAssets: metrics?.totalAssets ?? 6270000,

      tradePayables: metrics?.tradePayables ?? 540000,
      shortTermLoans: metrics?.shortTermLoans ?? 250000,
      otherCurrentLiabilities: metrics?.otherCurrentLiabilities ?? 180000,
      longTermLiabilities: metrics?.longTermLiabilities ?? 600000,
      totalLiabilities: metrics?.totalLiabilities ?? 1570000,

      totalEquity: metrics?.totalEquity ?? 4700000,

      revenue: metrics?.revenue ?? 5400000,
      costOfGoodsSold: metrics?.costOfGoodsSold ?? 3240000,
      grossProfit: metrics?.grossProfit ?? 2160000,
      operatingExpenses: metrics?.operatingExpenses ?? 1350000,
      netProfit: metrics?.netProfit ?? 810000,

      monthlyOperatingBurnRate: metrics?.monthlyOperatingBurnRate ?? 112500,
    }),
    [metrics],
  );

  const healthData = useMemo(
    () => calculateCFOFinancialHealth(standardMetrics),
    [standardMetrics],
  );

  const cashFlow = useMemo(
    () =>
      calculateCashFlowStatement({
        beginningCash: 1100000,
        netProfit: standardMetrics.netProfit,
        depreciation: 180000,
        beginningAR: 750000,
        endingAR: standardMetrics.tradeReceivables,
        beginningInventory: 580000,
        endingInventory: standardMetrics.inventory,
        beginningAP: 480000,
        endingAP: standardMetrics.tradePayables,
        capexPurchases: 150000,
        assetDisposalProceeds: 30000,
        loanProceeds: 0,
        loanRepayments: 80000,
        dividendsPaid: 150000,
      }),
    [standardMetrics],
  );

  return (
    <div className="flex flex-col gap-5">
      {/* Executive Health Score Banner */}
      <Card className="border-border/60 shadow-sm bg-gradient-to-r from-background to-muted/30">
        <CardContent className="flex flex-col gap-4 p-5 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center gap-4">
            <div
              className={`flex h-16 w-16 items-center justify-center rounded-2xl font-bold text-2xl shadow-inner ${
                healthData.healthScore >= 75
                  ? "bg-emerald-500/15 text-emerald-600 border border-emerald-500/30"
                  : healthData.healthScore >= 50
                    ? "bg-amber-500/15 text-amber-600 border border-amber-500/30"
                    : "bg-rose-500/15 text-rose-600 border border-rose-500/30"
              }`}
            >
              {healthData.healthScore}
            </div>

              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-xl font-bold tracking-tight text-foreground">
                    {tr("gl_cfo_dashboard_title", "แดชบอร์ดสุขภาพการเงินผู้บริหาร (CFO Financial Health)")}
                  </h2>
                  <span
                    className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${
                      healthData.overallStatus === "HEALTHY"
                        ? "bg-emerald-500/15 text-emerald-700"
                        : healthData.overallStatus === "WARNING"
                          ? "bg-amber-500/15 text-amber-700"
                          : "bg-rose-500/15 text-rose-700"
                    }`}
                  >
                    {healthData.overallStatus === "HEALTHY"
                      ? tr("gl_cfo_status_healthy", "สถานะแข็งแกร่ง (Healthy)")
                      : healthData.overallStatus === "WARNING"
                        ? tr("gl_cfo_status_warning", "ควรระวัง (Warning)")
                        : tr("gl_cfo_status_critical", "มีความเสี่ยง (Critical)")}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {tr("gl_cfo_dashboard_subtitle", "วิเคราะห์อัตราส่วนทางการเงิน TFRS, กระแสเงินสดสุทธิ, และระยะเวลาความอยู่รอดของเงินสดสำหรับเจ้าของกิจการ")}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-4 border-t pt-3 lg:border-t-0 lg:pt-0">
              <div className="rounded-xl border bg-card p-3 text-right">
                <span className="text-[11px] font-semibold text-muted-foreground block">
                  {tr("gl_cfo_cash_runway", "เงินสดสำรองพอใช้อีก (Runway)")}
                </span>
                <span className="text-lg font-bold text-primary flex items-center justify-end gap-1">
                  <Clock className="h-4 w-4 text-primary" />
                  {healthData.cashRunwayMonths.toFixed(1)} {tr("gl_cfo_months", "เดือน")}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 6 Key Financial Ratio Cards */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {healthData.ratios.map((r) => (
            <Card key={r.key} className="border-border/60 shadow-xs">
              <CardContent className="p-4 flex flex-col justify-between h-full">
                <div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-semibold text-muted-foreground">{r.nameTh}</span>
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold ${
                        r.status === "HEALTHY"
                          ? "bg-emerald-500/15 text-emerald-700"
                          : r.status === "WARNING"
                            ? "bg-amber-500/15 text-amber-700"
                            : "bg-rose-500/15 text-rose-700"
                      }`}
                    >
                      {r.status === "HEALTHY" ? (
                        <ShieldCheck className="h-3 w-3" />
                      ) : r.status === "WARNING" ? (
                        <AlertTriangle className="h-3 w-3" />
                      ) : (
                        <AlertCircle className="h-3 w-3" />
                      )}
                      {tr("gl_cfo_benchmark", "เกณฑ์")} {r.benchmark}
                    </span>
                  </div>
                  <div className="mt-2 text-2xl font-bold text-foreground">{r.formatted}</div>
                </div>

                <div className="mt-3 border-t pt-2 text-xs text-muted-foreground">
                  <p className="leading-relaxed">{r.advice}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        {/* Cash Flow Statement (Indirect Method) Card */}
        <Card className="border-border/60 shadow-sm">
          <CardContent className="p-5">
            <div className="flex items-center justify-between border-b pb-3">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-primary" />
                <h3 className="font-bold text-foreground">
                  {tr("gl_cfo_cash_flow_statement_title", "งบกระแสเงินสด (Cash Flow Statement - Indirect Method)")}
                </h3>
              </div>
              <span className="text-xs text-muted-foreground">{tr("gl_cfo_standard_tfrs", "มาตรฐาน TFRS for NPAEs")}</span>
            </div>

            <div className="mt-4 space-y-4 text-xs sm:text-sm">
              {/* 1. Operating Activities */}
              <div className="rounded-lg border bg-muted/20 p-4 space-y-2">
                <span className="font-bold text-foreground block">
                  {tr("gl_cfo_cf_operating", "1. กระแสเงินสดจากกิจกรรมดำเนินงาน (Cash Flows from Operating Activities)")}
                </span>

                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_net_profit", "กำไรสุทธิประจำงวด (Net Profit)")}</span>
                  <span>{cashFlow.operating.netProfit.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_depreciation", "บวก: ค่าเสื่อมราคาและค่าตัดจำหน่าย (Depreciation Addback)")}</span>
                  <span>+{cashFlow.operating.depreciation.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_ar_change", "การเปลี่ยนแปลงในลูกหนี้การค้า (Increase in AR)")}</span>
                  <span>{cashFlow.operating.arChange >= 0 ? "+" : ""}{cashFlow.operating.arChange.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_inv_change", "การเปลี่ยนแปลงในสินค้าคงเหลือ (Change in Inventory)")}</span>
                  <span>{cashFlow.operating.inventoryChange >= 0 ? "+" : ""}{cashFlow.operating.inventoryChange.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_ap_change", "การเปลี่ยนแปลงในเจ้าหนี้การค้า (Increase in AP)")}</span>
                  <span>{cashFlow.operating.apChange >= 0 ? "+" : ""}{cashFlow.operating.apChange.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>

                <div className="border-t pt-1 flex justify-between font-bold text-foreground">
                  <span>{tr("gl_cfo_cf_net_operating", "เงินสดสุทธิได้มาจากกิจกรรมดำเนินงาน")}</span>
                  <span className="text-emerald-600">
                    {cashFlow.operating.netOperatingCashFlow.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>

              {/* 2. Investing Activities */}
              <div className="rounded-lg border bg-muted/20 p-4 space-y-2">
                <span className="font-bold text-foreground block">
                  {tr("gl_cfo_cf_investing", "2. กระแสเงินสดจากกิจกรรมลงทุน (Cash Flows from Investing Activities)")}
                </span>

                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_capex", "เงินสดจ่ายซื้อสินทรัพย์ถาวร (Capex Purchases)")}</span>
                  <span>{cashFlow.investing.capexPurchases.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_disposal", "เงินสดรับจากการขายสินทรัพย์ถาวร (Disposal Proceeds)")}</span>
                  <span>+{cashFlow.investing.assetDisposalProceeds.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>

                <div className="border-t pt-1 flex justify-between font-bold text-foreground">
                  <span>{tr("gl_cfo_cf_net_investing", "เงินสดสุทธิใช้ไปในกิจกรรมลงทุน")}</span>
                  <span className="text-rose-600">
                    {cashFlow.investing.netInvestingCashFlow.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>

              {/* 3. Financing Activities */}
              <div className="rounded-lg border bg-muted/20 p-4 space-y-2">
                <span className="font-bold text-foreground block">
                  {tr("gl_cfo_cf_financing", "3. กระแสเงินสดจากกิจกรรมจัดหาเงิน (Cash Flows from Financing Activities)")}
                </span>

                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_loan_repay", "เงินสดจ่ายชำระคืนเงินกู้ยืม (Loan Repayments)")}</span>
                  <span>{cashFlow.financing.loanRepayments.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>
                <div className="flex justify-between pl-3 text-muted-foreground">
                  <span>{tr("gl_cfo_cf_dividends", "เงินสดจ่ายเงินปันผล (Dividends Paid)")}</span>
                  <span>{cashFlow.financing.dividendsPaid.toLocaleString("th-TH", { minimumFractionDigits: 2 })}</span>
                </div>

                <div className="border-t pt-1 flex justify-between font-bold text-foreground">
                  <span>{tr("gl_cfo_cf_net_financing", "เงินสดสุทธิใช้ไปในกิจกรรมจัดหาเงิน")}</span>
                  <span className="text-rose-600">
                    {cashFlow.financing.netFinancingCashFlow.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>

              {/* Summary Net Change in Cash */}
              <div className="rounded-lg border-2 border-primary/30 bg-primary/5 p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 font-bold text-sm">
                <div className="space-y-1">
                  <span className="text-muted-foreground block text-xs">{tr("gl_cfo_cf_summary_title", "สรุปการเปลี่ยนแปลงเงินสดสุทธิ")}</span>
                  <span className="text-foreground">
                    {tr("gl_cfo_cf_beg_cash", "เงินสดต้นงวด")}: {cashFlow.summary.beginningCash.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                    {" → "}
                    {tr("gl_cfo_cf_end_cash", "เงินสดปลายงวด")}: {cashFlow.summary.endingCash.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </span>
                </div>

                <div className="text-right">
                  <span className="text-xs text-muted-foreground block">{tr("gl_cfo_cf_net_change", "เงินสดเพิ่มขึ้นสุทธิ (Net Change)")}</span>
                  <span className="text-lg text-emerald-600">
                    +{cashFlow.summary.netCashChange.toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                  </span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
    </div>
  );
}
