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
