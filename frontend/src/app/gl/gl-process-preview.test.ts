import { readdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it, vi } from "vitest";
import { glRequest } from "@/lib/general-ledger-api";
import { emptyReportFilters, fetchReport } from "./gl-reports";
vi.mock("@/lib/general-ledger-api",()=>({glRequest:vi.fn().mockResolvedValue({})}));
describe("process preview company scope",()=>{
  it.each(["workingpaper","trialbalance"])("requests all branches explicitly for %s without altering regular report filters",async(name)=>{
    const applied={...emptyReportFilters,fiscalyear:"2026",from:"2026-01-01",to:"2026-12-31"};
    await fetchReport(name,applied,2,50,7,true);
    let query=new URLSearchParams(vi.mocked(glRequest).mock.calls.at(-1)![0].split("?")[1]);
    expect(query.get("companywide")).toBe("true");expect(query.get("snapshot")).toBe("7");expect(query.get("page")).toBe("2");
    await fetchReport(name,applied);
    query=new URLSearchParams(vi.mocked(glRequest).mock.calls.at(-1)![0].split("?")[1]);
    expect(query.has("companywide")).toBe(false);expect(applied).not.toHaveProperty("companywide");
  });
});
// adversarial review 2026-09-24: the book column shows the book name only when ReportGrid gets the company books
describe("every ReportGrid gets the journal books",()=>{
  it("passes books= so the bookcode column shows names, not raw codes",()=>{
    const dir=resolve(process.cwd(),"src","app","gl");
    const missing=readdirSync(dir).filter((name)=>name.endsWith(".tsx")).flatMap((name)=>[...readFileSync(resolve(dir,name),"utf8").matchAll(/<ReportGrid\b[\s\S]*?\/>/g)].filter((match)=>!/\bbooks=/.test(match[0])).map(()=>name));
    expect(missing).toEqual([]);
  });
});
// Champ checks the print / export right before it prints or exports (BaseReportView.cpp:736-745,2406-2415): every
// page of a GL report pulled for print / export carries purpose so the backend checks report:print / report:export
describe("GL report print / export rights",()=>{
  it("marks the whole-report pulls with their purpose and leaves screen reads unmarked",async()=>{
    const applied={...emptyReportFilters,fiscalyear:"2026"};
    await fetchReport("pnl",applied,2,500,7,false,"export");
    expect(new URLSearchParams(vi.mocked(glRequest).mock.calls.at(-1)![0].split("?")[1]).get("purpose")).toBe("export");
    await fetchReport("pnl",applied);
    expect(new URLSearchParams(vi.mocked(glRequest).mock.calls.at(-1)![0].split("?")[1]).has("purpose")).toBe(false);
    const source=readFileSync(resolve(process.cwd(),"src/app/gl/gl-reports.tsx"),"utf8");
    const complete=source.slice(source.indexOf("async function completeReport("),source.indexOf("async function printReport("));
    expect(complete.match(/fetchReport\([^)]*, purpose\)/g)).toHaveLength(2);
    expect(source.slice(source.indexOf("async function printReport("),source.indexOf("async function exportCsv("))).toMatch(/completeReport\([\s\S]*, "print"\);/);
    expect(source.slice(source.indexOf("async function exportCsv("),source.indexOf("const resultReady"))).toMatch(/completeReport\([\s\S]*, "export"\);/);
  });
  // print / CSV outside the report screen also go through a purpose-checked read; Champ checks the right on click
  it("re-reads the financial statement with its purpose before printing or exporting it",()=>{
    const source=readFileSync(resolve(process.cwd(),"src/app/gl/gl-statement-designer.tsx"),"utf8");
    expect(source).toContain('fetchReport("statement", query, 1, 50, undefined, false, purpose)');
    expect(source).toContain('statementFor("print").then((report) => report && statementPrint.print(statementView(report)');
    expect(source).toContain('const report = await statementFor("export");');
    expect(source).not.toContain("statementPrint.print(statementView(calculated)");
    expect(source).not.toContain("reportCsv(calculated)");
  });
  it("checks the right before printing or exporting the chart of accounts already on screen",()=>{
    const masters=readFileSync(resolve(process.cwd(),"src/app/gl/gl-masters.tsx"),"utf8");
    expect(masters.slice(masters.indexOf("const handleExportCsv"),masters.indexOf("chartOfAccountsCsv(treeAccounts"))).toContain('await ensureReportRight("accounts", "export");');
    const conditions=readFileSync(resolve(process.cwd(),"src/app/gl/gl-print-conditions.tsx"),"utf8");
    expect(conditions.slice(conditions.indexOf("export function ChartPrintDialog("),conditions.indexOf("<GLChartOfAccountsPrint"))).toContain('await ensureReportRight("accounts", "print");');
  });
});
