"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { GL_MENU_ITEMS, type GLResource } from "@/lib/general-ledger";
import { type LanguageCode } from "@/lib/i18n";
import { GLMasters } from "./gl-masters";
import { GLJournals } from "./gl-journals";
import { GLReports } from "./gl-reports";
import { GLProcesses } from "./gl-processes";
import { GLExport, GLXbrlPreparation } from "./gl-export";
import { GLStatementDesigner } from "./gl-statement-designer";
import { GLAllocations } from "./gl-allocations";
import { GLLanguageProvider, actionClass, panel, useGLText } from "./gl-common";

const masterRoutes: Record<string, GLResource> = {
  "/gl/chartofaccounts": "accounts",
  "/gl/fiscal-years": "fiscal-years",
  "/gl/budget": "budgets",
  "/gl/account-groups": "account-groups",
  "/gl/account-mapping": "mappings",
  "/gl/product-account-groups": "product-account-groups",
  "/gl/periodlock": "periods",
};
const reportRoutes: Record<string, string> = { "/gl/annual-balances": "annual-balances", "/gl/workingpaper": "workingpaper", "/checkdaily/dailyinfoscreen": "daily-check" };
const processRoutes = { "/gl/financialclose": "close", "/gl/year-end": "year-end", "/gl/recalculate-posted": "recalculate", "/gl/reprocess": "reprocess" } as const;
export function GeneralLedgerScreen({ route, embedded = false, language = "th" }: { route: string; embedded?: boolean; language?: LanguageCode }) {
  return <GLLanguageProvider language={language}><GeneralLedgerWorkbench route={route} embedded={embedded} /></GLLanguageProvider>;
}
function GeneralLedgerWorkbench({ route, embedded }: { route: string; embedded: boolean }) {
  const tr = useGLText();
  const cleanRoute = route.split("?")[0], item = GL_MENU_ITEMS.find((entry) => entry.route === cleanRoute);
  const [tab, setTab] = useState("main"), [dirty, setDirty] = useState(false);
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  useEffect(() => {
    const listener = (event: Event) => { const detail = (event as CustomEvent<{ route: string; dirty: boolean }>).detail; if (detail?.route === cleanRoute) setDirty(detail.dirty); };
    window.addEventListener("bc-gl-dirty", listener);
    return () => window.removeEventListener("bc-gl-dirty", listener);
  }, [cleanRoute]);
  async function changeTab(next: string) {
    if (next === tab) return;
    if (dirty && !await confirm({ title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"), description: tr("gl_save_before_switch_or_discard", "กรุณาบันทึกก่อนสลับ หรือยืนยันละทิ้งการแก้ไข"), confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข") })) return;
    setTab(next);
  }
  let content;
  if (cleanRoute === "/gl/statement-designer") content = <GLStatementDesigner route={cleanRoute} />;
  else if (cleanRoute === "/gl/allocations") content = <GLAllocations route={cleanRoute} />;
  else if (masterRoutes[cleanRoute]) content = <GLMasters key={cleanRoute} resource={masterRoutes[cleanRoute] as Exclude<GLResource, "journals">} route={cleanRoute} />;
  else if (cleanRoute === "/gl/openingbalance") content = <GLJournals route={cleanRoute} kind="opening" />;
  else if (cleanRoute.startsWith("/gl/journal/")) content = <GLJournals key={cleanRoute} route={cleanRoute} book={cleanRoute.split("/").at(-1)!.toUpperCase()} />;
  else if (cleanRoute === "/gl/posting" || cleanRoute === "/gl/unposting") content = <GLJournals key={cleanRoute} route={cleanRoute} mode={cleanRoute === "/gl/posting" ? "post" : "reverse"} />;
  else if (cleanRoute in processRoutes) content = <GLProcesses key={cleanRoute} route={cleanRoute} action={processRoutes[cleanRoute as keyof typeof processRoutes]} />;
  else if (cleanRoute === "/tools/databackup") content = <GLExport />;
  else if (cleanRoute === "/report/xbrl") content = <GLXbrlPreparation />;
  else if (cleanRoute === "/report/cashflowforecast" && tab === "forecast") content = <GLMasters resource="forecast" route={cleanRoute} />;
  else content = <GLReports key={cleanRoute} name={reportRoutes[cleanRoute] ?? cleanRoute.split("/").at(-1) ?? "trialbalance"} />;
  if (!item) return <div className={panel}>{tr("gl_account_page_not_found", "ไม่พบหน้าบัญชีที่ต้องการ")}</div>;
  return <main className={`gl-workbench flex min-w-0 flex-1 flex-col gap-2 text-[0.95rem] leading-relaxed ${embedded ? "p-2 h-full min-h-0 overflow-hidden" : "mx-auto max-w-[1800px] p-3 min-h-[calc(100dvh-2rem)]"}`} data-gl-route={cleanRoute}>
    {cleanRoute === "/report/cashflowforecast" && <nav aria-label={tr("gl_cash_flow_projection", "ประมาณการกระแสเงินสด")} className="shrink-0 flex flex-wrap gap-2"><Button className={actionClass} variant={tab === "main" ? "default" : "outline"} aria-pressed={tab === "main"} onClick={() => void changeTab("main")}>{tr("gl_projection_report", "รายงานประมาณการ")}</Button><Button className={actionClass} variant={tab === "forecast" ? "default" : "outline"} aria-pressed={tab === "forecast"} onClick={() => void changeTab("forecast")}>{tr("gl_record_cash_flow_est", "บันทึกประมาณการเงินเข้าออก")}</Button></nav>}
    <div className="flex-1 min-h-0 flex flex-col">{content}</div>{confirmationDialog}
  </main>;
}
