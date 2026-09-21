import { expect, test } from "@playwright/test";
import { resolve } from "node:path";
import { emptyJournal } from "../src/lib/general-ledger";

let harness = "";
test.beforeAll(async () => {
  const { build } = await import("vite");
  const output = await build({ configFile: false, define: { __dirname: JSON.stringify("/") }, root: process.cwd(), resolve: { alias: { "@": resolve("src") } },
    plugins: [{ name: "report-test-entry", resolveId(id) { if (id === "report-test-entry") return id; }, load(id) { if (id === "report-test-entry") return 'import { createElement } from "react"; import { createRoot } from "react-dom/client"; import { GLReports } from "' + resolve("src/app/gl/gl-reports.tsx").replaceAll("\\", "/") + '"; createRoot(document.getElementById("root")).render(createElement(GLReports, {name:"trialbalance"}));'; } }],
    build: { write: false, minify: false, rollupOptions: { input: "report-test-entry", output: { inlineDynamicImports: true } } } });
  const result = Array.isArray(output) ? output[0] : output;
  if (!("output" in result)) throw new Error("Missing test bundle");
  harness = result.output.filter(item => item.type === "chunk").map(item => item.code).join("\n");
});

test("report drill uses applied filters and exact source ID", async ({ page }) => {
  const pageErrors: string[] = [];
  page.on("pageerror", error => pageErrors.push(error.message));
  const reads: URL[] = [];
  await page.addInitScript(() => localStorage.setItem("bc_auth", JSON.stringify({username:"test",backendUrl:location.origin})));
  await page.route("**/api/auth/refresh", route => route.fulfill({json:{success:true,token:"test"}}));
  await page.route("**/api/gl/**", async route => {
    const url = new URL(route.request().url()); reads.push(url);
    const path = url.pathname;
    let data: unknown;
    if(path.includes("/reports/")) data={sequence:1,columns:[{key:"accountcode",label:"บัญชี",amount:false},{key:"docno",label:"เลขที่",amount:false}],rows:[{accountcode:"1000",docno:"JV1",journalid:"exact-source"}],totals:{},totalrows:1,warnings:[],asof:"2026-01-31"};
    else if(path.endsWith("/journals/exact-source")) data={...emptyJournal(),id:"exact-source",docno:"JV1",description:"หลักฐานที่ถูกต้อง"};
    else data={items:path.endsWith("/fiscal-years")?[{code:"2026",startdate:"2026-01-01",enddate:"2026-12-31",scale:2}]:[],total:path.endsWith("/fiscal-years")?1:0,page:1,limit:1000,sequence:1};
    await route.fulfill({json:{success:true,data}});
  });
  await page.route("**/report-test", route => route.fulfill({contentType:"text/html",body:'<div id="root"></div><script type="module" src="/report-test.js"></script>'}));
  await page.route("**/report-test.js", route => route.fulfill({contentType:"text/javascript",body:harness}));
  await page.goto("http://localhost:3017/report-test");
  await page.getByRole("combobox",{name:"ปีบัญชี",exact:true}).click();
  await page.getByRole("option",{name:"2026",exact:true}).click();
  const dates=page.locator('input[type="date"]');
  await dates.nth(1).fill("2026-01-31");
  await page.getByRole("button",{name:"แสดงรายงาน",exact:true}).click();
  await expect(page.getByRole("button",{name:"JV1",exact:true})).toBeVisible();
  await dates.nth(0).fill("2026-02-01");
  await dates.nth(1).fill("2026-02-28");
  await page.getByRole("button",{name:"ดูรายงานแยกประเภท",exact:true}).click();
  await expect.poll(()=>reads.some(url=>url.pathname.endsWith("/reports/ledger"))).toBe(true);
  const ledger=reads.find(url=>url.pathname.endsWith("/reports/ledger"))!;
  expect(ledger.searchParams.get("from")).toBe("2026-01-01");
  expect(ledger.searchParams.get("to")).toBe("2026-01-31");
  await page.getByRole("button",{name:"JV1",exact:true}).click();
  await expect(page.getByText("หลักฐานที่ถูกต้อง",{exact:true})).toBeVisible();
  expect(reads.some(url=>url.pathname.endsWith("/journals") && url.searchParams.has("q"))).toBe(false);
  await page.keyboard.press("Escape");
  await page.getByRole("button",{name:"กลับไปรายงานก่อนหน้า",exact:true}).click();
  await expect(dates.nth(0)).toHaveValue("2026-01-01");
  await expect(dates.nth(1)).toHaveValue("2026-01-31");
  expect(pageErrors).toEqual([]);
});
