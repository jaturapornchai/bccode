import { execFileSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { expect, test, type Page, type TestInfo } from "@playwright/test";
import { menuText } from "../src/lib/menu-data";
import { GL_MENU_ITEMS, emptyAccount, emptyFiscalYear, emptyJournal, emptyMaster, emptyLine, type GLCommand } from "../src/lib/general-ledger";

// Deliberately bound to the isolated browser UAT stack. Never target production.
test.use({ baseURL: "http://127.0.0.1:3011", trace: "off", actionTimeout: 20000, navigationTimeout: 60000 });
test.afterEach(async ({ page }, info) => { if (info.status !== info.expectedStatus) { await page.screenshot({ path: info.outputPath("failure.png"), fullPage: true }).catch(() => {}); await info.attach("failure-location", { body: page.url(), contentType: "text/plain" }); } });
test.skip(process.env.GL_UAT_ENABLED !== "1", "Run explicitly against the isolated GL browser UAT containers");
const mongoContainer = "bc-gl-mongo-browser-20260911";
const database = "gl_browser_uat_20260911";
const holding = "glbrowser20260911";
const company = "GLUAT";
const seed = process.env.GL_UAT_SEED ?? "20260911";
if (!/^\d{6,14}$/.test(seed)) throw new Error("GL_UAT_SEED must be 6–14 decimal digits");
const setupPrefix = `GU${seed}`;
const attempt = process.env.GL_UAT_ATTEMPT ?? "1";
if (!/^\d{1,3}$/.test(attempt)) throw new Error("GL_UAT_ATTEMPT must be 1–3 digits");
const prefix = `${setupPrefix}R${attempt}`;
const fiscalyear = `${setupPrefix}FY`;
const codes = { cash: `${setupPrefix}CA`, equity: `${setupPrefix}EQ`, profit: `${setupPrefix}PL`, retained: `${setupPrefix}RE` };
const name = (text: string) => [{ code: "th", name: text }];
type Json = Record<string, unknown>;
const collections: Record<string, string> = { accounts: "chart_of_accounts", "fiscal-years": "fiscal_year", periods: "gl_periods", journals: "gl_journals", budgets: "gl_budgets" };
function mongoOne(collection: string, filter: Json): Json | null {
  if (!/^[a-z_]+$/.test(collection)) throw new Error("Unexpected test collection");
  const script = `const doc=db.getSiblingDB(${JSON.stringify(database)}).getCollection(${JSON.stringify(collection)}).findOne(${JSON.stringify(filter)});print(EJSON.stringify(doc,{relaxed:false}));`;
  return JSON.parse(execFileSync("docker", ["exec", mongoContainer, "mongosh", "--quiet", "--eval", script], { encoding: "utf8" }).trim());
}
function decimal(value: unknown): string { return (value as { $numberDecimal: string }).$numberDecimal; }
async function proof(info: TestInfo, stage: string, collection: string, filter: Json) {
  const doc = mongoOne(collection, filter);
  await info.attach(`mongo-${stage}`, { body: JSON.stringify({ seed, collection, filter, doc }, null, 2), contentType: "application/json" });
  return doc;
}
async function openMenu(page: Page, route: string) {
  const search = page.getByPlaceholder("ค้นหาเมนู เอกสาร หรือหน้าจอ");
  await search.fill(route);
  const item = GL_MENU_ITEMS.find((candidate) => candidate.route === route)!;
  await page.getByRole("search").getByRole("button").filter({ has: page.getByText(menuText(item.label, "th"), { exact: true }) }).click();
  await expect(page.locator(`.gl-workbench[data-gl-route="${route}"]:visible`)).toBeVisible();
  await expect(page.locator(".gl-workbench:visible").getByRole("heading", { name: menuText(item.label, "th"), exact: true })).toBeVisible();
}
async function uiCommand(page: Page, button: ReturnType<Page["getByRole"]>, confirmLabel?: string) {
  const response = page.waitForResponse((result) => result.url().endsWith("/api/gl/command") && result.request().method() === "POST");
  await button.click();
  if (confirmLabel) await page.getByRole("dialog").getByRole("button", { name: confirmLabel, exact: true }).click();
  const result = await response, payload = await result.json();
  expect(result.ok(), payload.message ?? "GL command failed").toBe(true);
  expect(payload.success).toBe(true);
  return payload.data as { id: string; version: number };
}

test("GL actual UI, CRUD, posting, exact balances, 35 routes and responsive themes", async ({ page }, info) => {
  test.setTimeout(720000);
  await info.attach("seed", { body: JSON.stringify({ seed, attempt }), contentType: "application/json" });
  await page.goto("/");
  const loginResponse = page.waitForResponse((response) => response.url().endsWith("/api/auth/demo-login"));
  await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
  const login = await (await loginResponse).json();
  expect(login.success).toBe(true);
  await expect(page).toHaveURL(/\/holding$/);
  await expect(page.getByRole("heading", { name: "เลือกกลุ่มกิจการ", exact: true }).first()).toBeVisible();
  const headers = { Authorization: `Bearer ${login.token}`, "x-bc-backend-url": "http://127.0.0.1:3011/backend/goapi" };
  const post = async (path: string, data: unknown) => {
    let response = await page.request.post(path, { headers, data });
    if (response.status() === 401) { const refreshed = await (await page.request.post("/api/auth/refresh")).json(); expect(refreshed.success).toBe(true); headers.Authorization = `Bearer ${refreshed.token}`; response = await page.request.post(path, { headers, data }); }
    const payload = await response.json();
    expect(response.ok(), payload.message ?? `Request failed ${path}`).toBe(true);
    expect(payload.success).not.toBe(false);
    return payload;
  };
  const cmd = async (command: Omit<GLCommand, "requestid">) => (await post("/api/gl/command", { ...command, requestid: crypto.randomUUID() })).data as { id: string; version: number };
  const holdingDoc = mongoOne("shops", { holdingcode: holding });
  if (!holdingDoc) {
    await post("/api/workspace/create-holding", { holdingcode: holding, names: name("กลุ่มทดสอบบัญชีแยกประเภท"), name1: "กลุ่มทดสอบบัญชีแยกประเภท" });
    expect(await proof(info, "holding-created", "shops", { holdingcode: holding })).not.toBeNull();
  }
  await post("/api/workspace/select-holding", { holdingcode: holding });
  let companyDoc = mongoOne("organizationcompanies", { holdingcode: holding, code: company });
  if (!companyDoc) {
    await post("/backend/organization/company", { code: company, names: name("บริษัททดสอบบัญชีแยกประเภท"), isactive: true });
    companyDoc = await proof(info, "company-created", "organizationcompanies", { holdingcode: holding, code: company });
  }
  expect(companyDoc).not.toBeNull();
  let branch = mongoOne("organizationbranches", { holdingcode: holding, companyguid: companyDoc!.guidfixed, code: "00000" });
  if (!branch) {
    await post("/api/workspace/branch", { branch: { code: "00000", companyguid: companyDoc!.guidfixed, names: name("สำนักงานใหญ่ทดสอบบัญชี"), languages: ["th"], departments: [], timezone: "Asia/Bangkok", basecurrency: "THB", language: "th", isactive: true } });
    branch = await proof(info, "branch-created", "organizationbranches", { holdingcode: holding, companyguid: companyDoc!.guidfixed, code: "00000" });
  }
  const membership = mongoOne("shopusers", { holdingcode: holding, username: "gluat20260911" });
  if (!(membership?.accessscopes as Json[] | undefined)?.some((scope) => scope.companyuid === companyDoc!.companyuid)) {
    await post("/backend/holding/permission", { username: "gluat20260911", editusername: "gluat20260911", role: 2, isaccessdisabled: false, accessscopes: [{ scopetype: "company", companyuid: companyDoc!.companyuid, businesscode: company, allbranches: true }], permissionsets: [] });
    expect((await proof(info, "membership-scope", "shopusers", { holdingcode: holding, username: "gluat20260911" }))?.accessscopes).toHaveLength(1);
  }
  await page.reload();
  await page.getByRole("button", { name: /กลุ่มทดสอบบัญชีแยกประเภท/ }).first().click();
  await page.getByRole("button", { name: /บริษัททดสอบบัญชีแยกประเภท/ }).first().click();
  await page.getByRole("button", { name: /สำนักงานใหญ่ทดสอบบัญชี/ }).first().click();
  const skipUnits = page.getByRole("button", { name: "เข้าเมนูก่อน", exact: true }).first();
  await expect.poll(async () => page.url().endsWith("/menu") || await skipUnits.isVisible()).toBe(true);
  if (await skipUnits.isVisible()) await skipUnits.click();
  await expect(page).toHaveURL(/\/menu$/, { timeout: 45000 });
  const storageDir = join(tmpdir(), "bc-gl-browser-auth-20260911"); mkdirSync(storageDir, { recursive: true });
  await page.context().storageState({ path: join(storageDir, "storage-state.json") });
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });

  // Setup accounting through real commands and verify Mongo immediately, before the next write.
  for (const [key, code] of Object.entries(codes)) {
    if (mongoOne(collections.accounts, { holdingcode: holding, accountcode: code, isdeleted: false })) continue;
    const result = await cmd({ resource: "accounts", action: "create", account: { ...emptyAccount(), accountcode: code, names: name(`บัญชีทดสอบ ${key}`), accounttype: key === "cash" ? "asset" : "equity", normalbalance: key === "cash" ? "debit" : "credit", iscash: key === "cash" } });
    expect((await proof(info, `account-${key}-created`, collections.accounts, { _id: result.id }))?.accountcode).toBe(code);
  }
  const year = { ...emptyFiscalYear(), code: fiscalyear, startdate: "2026-01-01", enddate: "2026-12-31", currency: "THB", retainedearningsaccount: codes.retained, profitlossaccount: codes.profit };
  if (!mongoOne(collections["fiscal-years"], { holdingcode: holding, code: fiscalyear, isdeleted: false })) {
    const yearResult = await cmd({ resource: "fiscal-years", action: "create", fiscalyear: year });
    expect((await proof(info, "year-created", collections["fiscal-years"], { _id: yearResult.id }))?.code).toBe(fiscalyear);
  }
  const persistedYear = mongoOne(collections["fiscal-years"], { holdingcode: holding, code: fiscalyear, isdeleted: false });
  expect(persistedYear).toMatchObject({ holdingcode: holding, businesscode: company, startdate: "2026-01-01", enddate: "2026-12-31", currency: "THB", scale: { $numberInt: "2" }, closed: false });
  await page.reload();
  await expect(page).toHaveURL(/\/menu$/, { timeout: 45000 });
  for (const item of GL_MENU_ITEMS) { await openMenu(page, item.route); }
  await info.attach("routes-opened", { body: JSON.stringify(GL_MENU_ITEMS.map((item) => item.route)), contentType: "application/json" });

  while (await page.getByRole("button", { name: /^ปิดแท็บ / }).count()) await page.getByRole("button", { name: /^ปิดแท็บ / }).first().click();
  await openMenu(page, "/gl/chartofaccounts");
  let screen = page.locator(".gl-workbench:visible");
  await screen.getByRole("button", { name: "เพิ่มรายการ", exact: true }).click();
  await screen.getByLabel("รหัสบัญชี", { exact: true }).fill(`${prefix}CRUD`);
  await screen.getByLabel("ชื่อบัญชีภาษาไทย", { exact: true }).fill("ทดสอบสร้างบัญชี อักขระ & ' ไทย");
  expect(await screen.getByLabel("ระดับบัญชี", { exact: true }).inputValue()).toBe("1");
  const account = await uiCommand(page, screen.getByRole("button", { name: "บันทึก", exact: true }));
  const createdDoc = await proof(info, "account-ui-created", collections.accounts, { _id: account.id });
  expect(createdDoc?.accountcode).toBe(`${prefix}CRUD`);
  expect(Number(createdDoc?.level ?? 1)).toBe(1);

  // Create child account with auto-suggested level 2
  await screen.getByRole("button", { name: "เพิ่มรายการ", exact: true }).click();
  await screen.getByLabel("รหัสบัญชี", { exact: true }).fill(`${prefix}SUB`);
  await screen.getByLabel("ชื่อบัญชีภาษาไทย", { exact: true }).fill("ทดสอบบัญชีย่อย ระดับ 2");
  await screen.getByLabel("บัญชีแม่", { exact: true }).selectOption(`${prefix}CRUD`);
  expect(await screen.getByLabel("ระดับบัญชี", { exact: true }).inputValue()).toBe("2");
  const childAccount = await uiCommand(page, screen.getByRole("button", { name: "บันทึก", exact: true }));
  const childDoc = await proof(info, "child-account-created", collections.accounts, { _id: childAccount.id });
  expect(childDoc?.accountcode).toBe(`${prefix}SUB`);
  expect(Number(childDoc?.level ?? 2)).toBe(2);

  // Update child account
  await screen.getByLabel("ชื่อบัญชีภาษาไทย", { exact: true }).fill("แก้ไขบัญชีย่อยแล้ว");
  await uiCommand(page, screen.getByRole("button", { name: "บันทึก", exact: true }), "บันทึกการแก้ไข");
  expect(JSON.stringify((await proof(info, "child-account-updated", collections.accounts, { _id: childAccount.id }))?.names)).toContain("แก้ไขบัญชีย่อยแล้ว");

  // Delete child account first
  await screen.getByLabel("เหตุผลการเปลี่ยนแปลง").fill("ล้างบัญชีย่อย");
  await uiCommand(page, screen.getByRole("button", { name: "ลบรายการ", exact: true }), "ลบรายการ");
  expect((await proof(info, "child-account-deleted", collections.accounts, { _id: childAccount.id }))?.isdeleted).toBe(true);

  // Then delete parent account
  await screen.getByRole("button", { name: `${prefix}CRUD`, exact: true }).click();
  await screen.getByLabel("เหตุผลการเปลี่ยนแปลง").fill("ล้างข้อมูลทดสอบ CRUD");
  await uiCommand(page, screen.getByRole("button", { name: "ลบรายการ", exact: true }), "ลบรายการ");
  expect((await proof(info, "account-ui-deleted", collections.accounts, { _id: account.id }))?.isdeleted).toBe(true);

  await openMenu(page, "/gl/periodlock"); screen = page.locator(".gl-workbench:visible");
  await screen.getByRole("button", { name: "เพิ่มรายการ", exact: true }).click();
  await screen.getByLabel("รหัส", { exact: true }).fill(`${prefix}PER`);
  await screen.getByLabel("ชื่อ", { exact: true }).fill("งวดทดสอบกันยายน");
  await screen.getByLabel("ปีบัญชี", { exact: true }).selectOption(fiscalyear);
  await screen.getByLabel("วันเริ่มต้น", { exact: true }).fill("2026-12-01");
  await screen.getByLabel("วันสิ้นสุด", { exact: true }).fill("2026-12-31");
  const period = await uiCommand(page, screen.getByRole("button", { name: "บันทึก", exact: true }));
  expect((await proof(info, "period-created", collections.periods, { _id: period.id }))?.locked).toBe(false);
  await screen.getByLabel("ชื่อ", { exact: true }).fill("งวดทดสอบแก้ไขแล้ว");
  await uiCommand(page, screen.getByRole("button", { name: "บันทึก", exact: true }), "บันทึกการแก้ไข");
  expect((await proof(info, "period-updated", collections.periods, { _id: period.id }))?.name).toBe("งวดทดสอบแก้ไขแล้ว");
  await screen.getByLabel("เหตุผลการเปลี่ยนแปลง").fill("ทดสอบการล็อกงวด");
  await uiCommand(page, screen.getByRole("button", { name: "ล็อกงวด", exact: true }), "ล็อกงวดบัญชี");
  expect((await proof(info, "period-locked", collections.periods, { _id: period.id }))?.locked).toBe(true);
  await screen.getByRole("button", { name: `${prefix}PER`, exact: true }).click();
  await screen.getByLabel("เหตุผลการเปลี่ยนแปลง").fill("ทดสอบปลดล็อกงวด");
  await uiCommand(page, screen.getByRole("button", { name: "ปลดล็อกงวด", exact: true }), "ปลดล็อกงวดบัญชี");
  expect((await proof(info, "period-unlocked", collections.periods, { _id: period.id }))?.locked).toBe(false);
  await screen.getByRole("button", { name: `${prefix}PER`, exact: true }).click();
  await screen.getByLabel("เหตุผลการเปลี่ยนแปลง").fill("ล้างงวดทดสอบ");
  await uiCommand(page, screen.getByRole("button", { name: "ลบรายการ", exact: true }), "ลบรายการ");
  expect((await proof(info, "period-deleted", collections.periods, { _id: period.id }))?.isdeleted).toBe(true);

  const existingWorkPeriod = mongoOne(collections.periods, { holdingcode: holding, fiscalyear, startdate: "2026-09-01", enddate: "2026-09-30", isdeleted: false });
  if (!existingWorkPeriod) {
    const workPeriod = await cmd({ resource: "periods", action: "create", master: { ...emptyMaster(), code: `${prefix}WORK`, name: "งวดใช้งานสำหรับรายการทดสอบ", isactive: true, fiscalyear, startdate: "2026-09-01", enddate: "2026-09-30", locked: false } });
    expect((await proof(info, "work-period-created", collections.periods, { _id: workPeriod.id }))?.fiscalyear).toBe(fiscalyear);
  } else expect(existingWorkPeriod).toMatchObject({ businesscode: company, locked: false });
  await openMenu(page, "/gl/journal/jv"); screen = page.locator(".gl-workbench:visible");
  await screen.getByRole("button", { name: "เพิ่มรายการ", exact: true }).click();
  await screen.getByLabel("เลขที่เอกสาร", { exact: true }).fill(`${prefix}JV`);
  await screen.getByLabel("ปีบัญชี", { exact: true }).selectOption(fiscalyear);
  await screen.getByLabel("วันที่เอกสาร", { exact: true }).fill("2026-09-11");
  await screen.getByLabel("คำอธิบายรายการ", { exact: true }).fill("ทดสอบยอด 0.1 + 0.2 = 0.3");
  await screen.getByLabel("บัญชีบรรทัด 1", { exact: true }).selectOption(codes.cash);
  await screen.getByLabel("เดบิตบรรทัด 1", { exact: true }).fill("0.1");
  await screen.getByLabel("บัญชีบรรทัด 2", { exact: true }).selectOption(codes.cash);
  await screen.getByLabel("เดบิตบรรทัด 2", { exact: true }).fill("0.2");
  await screen.getByRole("button", { name: "เพิ่มบรรทัด", exact: true }).click();
  await screen.getByLabel("บัญชีบรรทัด 3", { exact: true }).selectOption(codes.equity);
  await screen.getByLabel("เครดิตบรรทัด 3", { exact: true }).fill("0.2");
  await screen.getByRole("button", { name: "บันทึกฉบับร่าง", exact: true }).click();
  await expect(screen.getByRole("alert")).toContainText("ยอดเดบิตและเครดิตต้องเท่ากัน");
  expect(mongoOne(collections.journals, { holdingcode: holding, docno: `${prefix}JV` })).toBeNull();
  // Searching and switching tabs must preserve the unfinished journal.
  const search = page.getByPlaceholder("ค้นหาเมนู เอกสาร หรือหน้าจอ");
  await search.fill("งบทดลอง"); await search.fill("");
  await expect(screen.getByLabel("เลขที่เอกสาร", { exact: true })).toHaveValue(`${prefix}JV`);
  await page.getByRole("tab").first().click();
  await page.getByRole("tab").filter({ hasText: "สมุดรายวันทั่วไป" }).click();
  await expect(screen.getByLabel("เครดิตบรรทัด 3", { exact: true })).toHaveValue("0.2");
  const journalTab = page.getByRole("tab").filter({ hasText: "สมุดรายวันทั่วไป" });
  await journalTab.locator("..").getByRole("button", { name: /ปิด/ }).click();
  await page.getByRole("dialog").getByRole("button", { name: "กลับไปบันทึก", exact: true }).last().click();
  await expect(screen.getByLabel("เลขที่เอกสาร", { exact: true })).toHaveValue(`${prefix}JV`);
  const lookupCode = `${prefix}LOOK`;
  const lookup = await cmd({ resource: "accounts", action: "create", account: { ...emptyAccount(), accountcode: lookupCode, names: name("บัญชีทดสอบโหลด lookup ใหม่") } });
  expect((await proof(info, "lookup-created", collections.accounts, { _id: lookup.id }))?.accountcode).toBe(lookupCode);
  await screen.getByRole("button", { name: "โหลดใหม่", exact: true }).click();
  await expect(screen.getByLabel("บัญชีบรรทัด 1", { exact: true }).getByRole("option").filter({ hasText: lookupCode })).toHaveCount(1);
  await expect(screen.getByLabel("เครดิตบรรทัด 3", { exact: true })).toHaveValue("0.2");
  await cmd({ resource: "accounts", action: "delete", id: lookup.id, version: lookup.version, reason: "ล้างบัญชี lookup ทดสอบ" });
  expect((await proof(info, "lookup-deleted", collections.accounts, { _id: lookup.id }))?.isdeleted).toBe(true);
  // This developer-only overlay is absent from production and must not intercept the product controls.
  await page.addStyleTag({ content: "[data-dev-dom-inspector] { display: none !important; }" });
  for (const width of [1600, 1280, 1024, 768]) {
    await page.setViewportSize({ width, height: width === 768 ? 1024 : 900 });
    for (const theme of ["light", "dark"]) {
      if (await page.locator("html").getAttribute("data-theme") !== theme) await page.locator("button.theme-toggle").click();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await page.waitForTimeout(350);
      await screen.getByRole("button", { name: "บันทึกฉบับร่าง", exact: true }).hover();
      await screen.getByRole("button", { name: "บันทึกฉบับร่าง", exact: true }).focus();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
      const paneWidths = await screen.locator("[data-gl-pane]").evaluateAll((panes) => panes.map((pane) => pane.getBoundingClientRect().width));
      if (width >= 1280) { expect(paneWidths[0]).toBeLessThan(width * 0.65); expect(paneWidths[1]).toBeGreaterThan(width * 0.4); }
      await screen.getByRole("heading", { name: "สมุดรายวันทั่วไป", exact: true }).scrollIntoViewIfNeeded();
      await page.screenshot({ path: info.outputPath(`gl-journal-${width}-${theme}.png`), fullPage: true, animations: "disabled" });
    }
  }
  await page.setViewportSize({ width: 1600, height: 900 });
  await screen.getByLabel("เครดิตบรรทัด 3", { exact: true }).fill("0.3");
  const journal = await uiCommand(page, screen.getByRole("button", { name: "บันทึกฉบับร่าง", exact: true }));
  const created = await proof(info, "journal-created", collections.journals, { _id: journal.id });
  expect(decimal((created!.lines as Json[])[0].debit)).toBe("0.1");
  expect(decimal((created!.lines as Json[])[2].credit)).toBe("0.3");
  await screen.getByLabel("คำอธิบายรายการ", { exact: true }).fill("แก้ไขคำอธิบายทดสอบก่อนผ่านบัญชี");
  await uiCommand(page, screen.getByRole("button", { name: "บันทึกฉบับร่าง", exact: true }), "บันทึกฉบับร่าง");
  expect((await proof(info, "journal-updated", collections.journals, { _id: journal.id }))?.description).toBe("แก้ไขคำอธิบายทดสอบก่อนผ่านบัญชี");
  await uiCommand(page, screen.getByRole("button", { name: "ผ่านรายการบัญชี", exact: true }), "ผ่านรายการบัญชี");
  expect((await proof(info, "journal-posted", collections.journals, { _id: journal.id }))?.status).toBe("posted");
  await screen.getByRole("button", { name: `${prefix}JV`, exact: true }).click();
  await expect(screen.getByLabel("คำอธิบายรายการ", { exact: true })).toBeDisabled();
  await screen.getByLabel("เหตุผล", { exact: true }).fill("ทดสอบกลับรายการบัญชี");
  await screen.getByLabel("เลขที่ใบกลับรายการ").fill(`${prefix}REV`);
  await screen.getByLabel("วันที่กลับรายการ").fill("2026-09-12");
  await uiCommand(page, screen.getByRole("button", { name: "สร้างรายการกลับบัญชี", exact: true }), "สร้างรายการกลับบัญชี");
  const reversed = await proof(info, "journal-reversed", collections.journals, { holdingcode: holding, reversalof: journal.id });
  expect(reversed).not.toBeNull();
  expect(decimal((reversed!.lines as Json[])[0].credit)).toBe("0.1");
  const draft = { ...emptyJournal(), docno: `${prefix}DEL`, date: "2026-09-11", fiscalyear, currency: "THB", description: "ฉบับร่างทดสอบลบ", lines: [{ ...emptyLine(), accountcode: codes.cash, debit: "1" }, { ...emptyLine(), accountcode: codes.equity, credit: "1" }] };
  const draftResult = await cmd({ resource: "journals", action: "create", journal: draft });
  expect((await proof(info, "draft-created", collections.journals, { _id: draftResult.id }))?.status).toBe("draft");
  await cmd({ resource: "journals", action: "delete", id: draftResult.id, version: draftResult.version, reason: "ล้างฉบับร่างทดสอบ" });
  expect((await proof(info, "draft-deleted", collections.journals, { _id: draftResult.id }))?.isdeleted).toBe(true);
  const rejected = await page.request.post("/api/gl/command", { headers, data: { resource: "journals", action: "create", requestid: crypto.randomUUID(), journal: { ...draft, docno: `${prefix}BAD`, lines: [{ ...draft.lines[0], debit: 0.1 }, draft.lines[1]] } } });
  expect(rejected.status()).toBe(400);
  expect(mongoOne(collections.journals, { holdingcode: holding, docno: `${prefix}BAD` })).toBeNull();
  // An unused fiscal year exercises update/delete without removing posted accounting truth.
  const temporaryYear = { ...year, code: `${prefix}TEMP`, startdate: "2030-01-01", enddate: "2030-12-31" };
  const temp = await cmd({ resource: "fiscal-years", action: "create", fiscalyear: temporaryYear });
  expect((await proof(info, "temporary-year-created", collections["fiscal-years"], { _id: temp.id }))?.code).toBe(temporaryYear.code);
  const updated = await cmd({ resource: "fiscal-years", action: "update", id: temp.id, version: temp.version, fiscalyear: { ...temporaryYear, enddate: "2030-11-30" } });
  expect((await proof(info, "temporary-year-updated", collections["fiscal-years"], { _id: temp.id }))?.enddate).toBe("2030-11-30");
  await cmd({ resource: "fiscal-years", action: "delete", id: temp.id, version: updated.version, reason: "ล้างปีบัญชีทดสอบ" });
  expect((await proof(info, "temporary-year-deleted", collections["fiscal-years"], { _id: temp.id }))?.isdeleted).toBe(true);
  await openMenu(page, "/report/trialbalance"); screen = page.locator(".gl-workbench:visible");
  await screen.getByLabel("ปีบัญชี", { exact: true }).selectOption(fiscalyear);
  await screen.getByRole("button", { name: "แสดงรายงาน", exact: true }).click();
  await expect(screen.getByText("ข้อมูล ณ", { exact: false })).toBeVisible();
  const report = await (await page.request.get(`/api/gl/reports/trialbalance?fiscalyear=${fiscalyear}`, { headers })).json();
  expect(report.success).toBe(true);
  await info.attach("trialbalance", { body: JSON.stringify(report.data, null, 2), contentType: "application/json" });
  expect(errors).toEqual([]);
});
