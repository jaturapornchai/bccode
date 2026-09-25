"use client";

import { GL_MENU_ITEMS, type GLResource } from "@/lib/general-ledger";
import { isMenuScreenPending } from "@/lib/menu-screen-status";
import { type LanguageCode } from "@/lib/i18n";
import { GLMasters } from "./gl-masters";
import { GLJournals } from "./gl-journals";
import { GLReports } from "./gl-reports";
import { GLProcesses } from "./gl-processes";
import { GLStatementDesigner } from "./gl-statement-designer";
import { GLLanguageProvider, panel, useGLText } from "./gl-common";
import { SmartBreadcrumb } from "@/components/smart-breadcrumb";

const masterRoutes: Record<string, GLResource> = {
  "/gl/chartofaccounts": "accounts",
  "/gl/fiscal-years": "fiscal-years",
  "/gl/account-groups": "account-groups",
  "/gl/account-mapping": "mappings",
  "/gl/product-account-groups": "product-account-groups",
  "/gl/periodlock": "periods",
  "/gl/journal-books": "journal-books",
};
const processRoutes = { "/gl/financialclose": "close", "/gl/year-end": "year-end", "/gl/recalculate-posted": "recalculate", "/gl/reprocess": "reprocess" } as const;
export function GeneralLedgerScreen({ route, embedded = false, language = "th" }: { route: string; embedded?: boolean; language?: LanguageCode }) {
  return <GLLanguageProvider language={language}><GeneralLedgerWorkbench route={route} embedded={embedded} /></GLLanguageProvider>;
}
function GeneralLedgerWorkbench({ route, embedded }: { route: string; embedded: boolean }) {
  const tr = useGLText();
  const cleanRoute = route.split("?")[0], item = GL_MENU_ITEMS.find((entry) => entry.route === cleanRoute);
  let content;
  if (isMenuScreenPending(cleanRoute)) content = <GLPendingPanel />;
  else if (cleanRoute === "/gl/statement-designer") content = <GLStatementDesigner route={cleanRoute} />;
  else if (masterRoutes[cleanRoute]) content = <GLMasters key={cleanRoute} resource={masterRoutes[cleanRoute] as Exclude<GLResource, "journals">} route={cleanRoute} />;
  else if (cleanRoute === "/gl/openingbalance") content = <GLJournals route={cleanRoute} kind="opening" />;
  else if (cleanRoute === "/gl/journals" || cleanRoute === "/gl/journal") content = <GLJournals key="all" route={cleanRoute} book="" />;
  else if (cleanRoute.startsWith("/gl/journal/")) content = <GLJournals key={cleanRoute} route={cleanRoute} book={cleanRoute.split("/").at(-1)!.toUpperCase()} />;
  else if (cleanRoute === "/gl/posting" || cleanRoute === "/gl/unposting") content = <GLJournals key={cleanRoute} route={cleanRoute} mode={cleanRoute === "/gl/posting" ? "post" : "reverse"} />;
  else if (cleanRoute in processRoutes) content = <GLProcesses key={cleanRoute} route={cleanRoute} action={processRoutes[cleanRoute as keyof typeof processRoutes]} />;
  else content = <GLReports key={cleanRoute} name={cleanRoute.split("/").at(-1) ?? "trialbalance"} />;
  if (!item) return <div className={panel}>{tr("gl_account_page_not_found", "ไม่พบหน้าบัญชีที่ต้องการ")}</div>;
  return <main className={`gl-workbench flex min-w-0 flex-1 flex-col gap-2 text-[0.95rem] leading-relaxed ${embedded ? "p-2 h-full min-h-0 overflow-hidden" : "mx-auto max-w-[1800px] p-3 min-h-[calc(100dvh-2rem)]"}`} data-gl-route={cleanRoute}>
    {!embedded && <SmartBreadcrumb currentTitle={item ? (item.label.key ? tr(item.label.key, item.label.th) : item.label.th) : undefined} />}
    <div className="flex-1 min-h-0 flex flex-col">{content}</div>
  </main>;
}
/** Champ menu item whose GL view is not built yet — same wording as the main-menu planned-workflow card. */
function GLPendingPanel() {
  const tr = useGLText();
  return <section className={`${panel} grid gap-2`} data-gl-pending="true">
    <h2 className="text-lg font-semibold">{tr("menu_planned_workflow", "เมนูในแผนพัฒนา")}</h2>
    <p className="leading-relaxed text-muted-foreground">{tr("menu_planned_description", "หน้าจอนี้ยังอยู่ระหว่างเตรียมพัฒนา จึงยังบันทึกหรือประมวลผลข้อมูลไม่ได้ เลือกใช้งานเมนูอื่นจากแถบเมนูได้ตามปกติ")}</p>
  </section>;
}
