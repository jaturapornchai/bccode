import { expect, test } from "@playwright/test";

const token = { id: "12345678123412341234123456789012", name: "Audit assistant", kind: "mcp", mode: "readonly", holdingCode: "TEST", companyCodes: ["C01", "C02"], createdBy: "admin", createdAt: "2026-09-20T00:00:00Z", expiresAt: "2099-01-01T00:00:00Z", revokedAt: null, lastUsedAt: null };
const companies = [{ code: "C01", name: "บริษัทหนึ่ง" }, { code: "C02", name: "บริษัทสอง" }];
test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => {
    localStorage.setItem("bc_auth", JSON.stringify({ username: "token-admin-test", backendUrl: location.origin, holdingcode: "TEST" }));
    localStorage.setItem("bc_workspace", JSON.stringify({ shop: { holdingcode: "TEST" } }));
  });
  await page.route("**/api/auth/refresh", (route) => route.fulfill({ json: { success: true, token: "test-session-only" } }));
});

test("administrator creates a scoped token, sees its secret once and confirms deletion", async ({ page }, info) => {
  let created = false, revoked = false;
  const commands: Record<string, unknown>[] = [];
  await page.route("**/api/mcp-tokens**", async (route) => {
    if (route.request().url().endsWith("/companies")) { await route.fulfill({ json: { success: true, data: companies } }); return; }
    if (route.request().method() === "POST") {
      if (route.request().url().endsWith("/revoke")) { revoked = true; await route.fulfill({ json: { success: true } }); return; }
      commands.push(route.request().postDataJSON()); created = true;
      await route.fulfill({ json: { success: true, data: { ...token, token: "test-secret-never-persist-this" } } }); return;
    }
    await route.fulfill({ json: { success: true, data: created ? [{ ...token, revokedAt: revoked ? "2026-09-20T01:00:00Z" : null }] : [] } });
  });
  await page.goto("/mcp-tokens");
  await expect(page.getByText("ยังไม่มี token สำหรับ Holding นี้")).toBeVisible();
  await page.getByRole("button", { name: "เพิ่ม MCP token", exact: true }).click();
  await expect(page.getByLabel("ประเภทการเชื่อมต่อ")).toHaveValue("mcp");
  await expect(page.getByLabel("สิทธิ์การเข้าถึง")).toHaveValue("readonly");
  await page.getByLabel("ชื่อการเชื่อมต่อ").fill("Audit assistant");
  await page.getByLabel("สิทธิ์การเข้าถึง").selectOption("readwrite");
  await page.getByRole("button", { name: "สร้าง token", exact: true }).last().click();
  await expect(page.getByText("กรุณาเลือกบริษัทที่อนุญาตอย่างน้อยหนึ่งบริษัท")).toBeVisible();
  expect(commands).toHaveLength(0);
  await page.getByText("C01 — บริษัทหนึ่ง", { exact: true }).click();
  await page.getByText("C02 — บริษัทสอง", { exact: true }).click();
  await page.getByRole("button", { name: "สร้าง token", exact: true }).last().click();
  await expect(page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })).toHaveValue("test-secret-never-persist-this");
  expect(commands).toHaveLength(1); expect(commands[0]).toMatchObject({ name: "Audit assistant", kind: "mcp", mode: "readwrite", companyCodes: ["C01", "C02"] });
  expect(commands[0]).not.toHaveProperty("companyCode");
  expect(await page.evaluate(() => JSON.stringify({ ...localStorage, ...sessionStorage }))).not.toContain("test-secret-never-persist-this");
  await page.setViewportSize({ width: 1500, height: 1000 });
  const detailBox = await page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true }).boundingBox();
  expect(detailBox!.width).toBeGreaterThan(400);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: info.outputPath("tokens-light.png"), fullPage: true, mask: [page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })] });
  await page.getByRole("button", { name: "เปลี่ยนเป็นธีมมืด", exact: true }).click();
  await expect(page.getByRole("button", { name: "เปลี่ยนเป็นธีมสว่าง", exact: true })).toBeVisible();
  await page.screenshot({ path: info.outputPath("tokens-dark.png"), fullPage: true, animations: "disabled", mask: [page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })] });
  await page.getByRole("button", { name: "ลบ token", exact: true }).click();
  await expect(page.getByRole("dialog")).toBeVisible(); expect(revoked).toBe(false);
  await page.getByRole("dialog").getByRole("button", { name: "ลบ token", exact: true }).click();
  await expect(page.getByText("ลบ token แล้ว", { exact: true })).toBeVisible();
  await expect(page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("button", { name: /Audit assistant/ })).toHaveCount(0);
  await page.getByRole("button", { name: "โหลดใหม่", exact: true }).click();
  await expect(page.getByRole("button", { name: /Audit assistant/ })).toHaveCount(0);
});

test("API tokens use their own create action and deletion is cancellable and persistent", async ({ page }) => {
  let submitted: Record<string, unknown> | null = null;
  let created = false, revoked = false, revokeRequests = 0;
  const apiToken = { ...token, name: "API accounting", kind: "api", companyCodes: ["C02"] };
  await page.route("**/api/mcp-tokens**", async (route) => {
    if (route.request().url().endsWith("/companies")) { await route.fulfill({ json: { success: true, data: companies } }); return; }
    if (route.request().method() === "POST") {
      if (route.request().url().endsWith("/revoke")) {
        revokeRequests++; revoked = true;
        await route.fulfill({ json: { success: true } }); return;
      }
      submitted = route.request().postDataJSON(); created = true;
      await route.fulfill({ json: { success: true, data: { ...apiToken, token: "test-api-secret-only" } } }); return;
    }
    // The server retains revocation history; the management list must hide deleted tokens.
    await route.fulfill({ json: { success: true, data: created ? [{ ...apiToken, revokedAt: revoked ? "2026-09-20T01:00:00Z" : null }] : [] } });
  });
  await page.goto("/mcp-tokens");
  await page.getByRole("button", { name: "เพิ่ม API token", exact: true }).click();
  await expect(page.getByLabel("ประเภทการเชื่อมต่อ")).toHaveValue("api");
  await page.getByLabel("ชื่อการเชื่อมต่อ").fill("API accounting");
  await page.getByText("C02 — บริษัทสอง", { exact: true }).click();
  await expect(page.getByText(/API URL:.*\/api\/integration\/gl/)).toBeVisible();
  await page.getByRole("button", { name: "สร้าง token", exact: true }).click();
  await expect(page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })).toHaveValue("test-api-secret-only");
  expect(submitted).toMatchObject({ name: "API accounting", kind: "api", mode: "readonly", companyCodes: ["C02"] });
  await expect(page.getByText("API token และ MCP token แยกกัน ใช้ข้ามช่องทางไม่ได้")).toBeVisible();
  await expect(page.getByText(/API URL:.*\/api\/integration\/gl/)).toBeVisible();
  await page.getByRole("button", { name: "ลบ token", exact: true }).click();
  await page.getByRole("dialog").getByRole("button", { name: "ยกเลิก", exact: true }).filter({ hasText: "ยกเลิก" }).click();
  expect(revokeRequests).toBe(0); expect(revoked).toBe(false);
  await expect(page.getByRole("button", { name: /API accounting/ })).toHaveCount(1);
  await page.getByRole("button", { name: "ลบ token", exact: true }).click();
  await page.getByRole("dialog").getByRole("button", { name: "ลบ token", exact: true }).click();
  await expect(page.getByText("ลบ token แล้ว", { exact: true })).toBeVisible();
  expect(revokeRequests).toBe(1); expect(revoked).toBe(true);
  await expect(page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("button", { name: /API accounting/ })).toHaveCount(0);
  await page.getByRole("button", { name: "โหลดใหม่", exact: true }).click();
  await expect(page.getByRole("button", { name: /API accounting/ })).toHaveCount(0);
  await expect(page.getByText("ยังไม่มี token สำหรับ Holding นี้")).toBeVisible();
});

test("premium empty state and create form fit desktop and mobile", async ({ page }, info) => {
  await page.route("**/api/mcp-tokens**", (route) => route.fulfill({ json: { success: true, data: route.request().url().endsWith("/companies") ? companies : [] } }));
  await page.setViewportSize({ width: 1500, height: 950 });
  await page.goto("/mcp-tokens");
  await expect(page.getByRole("heading", { name: "เริ่มต้นเชื่อมต่อห้องบัญชี" })).toBeVisible();
  await page.screenshot({ path: info.outputPath("empty-desktop.png"), fullPage: true });
  await page.getByRole("button", { name: "เพิ่ม MCP token", exact: true }).click();
  await page.screenshot({ path: info.outputPath("create-desktop.png"), fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(page.getByLabel("ชื่อการเชื่อมต่อ")).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await page.screenshot({ path: info.outputPath("create-mobile.png"), fullPage: true });
});

test("non-administrator cannot access token management", async ({ page }) => {
  await page.route("**/api/mcp-tokens**", (route) => route.fulfill({ status: 403, json: { success: false, message: "ปฏิเสธสิทธิ์" } }));
  await page.goto("/mcp-tokens");
  await expect(page.getByText("ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")).toBeVisible();
  await expect(page.getByRole("button", { name: /^เพิ่ม (API|MCP) token$/ })).toHaveCount(0);
});

test("workspace changes clear an already displayed secret", async ({ page }) => {
  await page.route("**/api/mcp-tokens**", (route) => route.fulfill({ json: { success: true, data: route.request().url().endsWith("/companies") ? companies : route.request().method() === "POST" ? { ...token, token: "test-secret-clear-on-workspace" } : [] } }));
  await page.goto("/mcp-tokens");
  await page.getByRole("button", { name: "เพิ่ม MCP token", exact: true }).click();
  await page.getByLabel("ชื่อการเชื่อมต่อ").fill("Audit assistant");
  await page.getByText("C01 — บริษัทหนึ่ง", { exact: true }).click();
  await page.getByRole("button", { name: "สร้าง token", exact: true }).last().click();
  await expect(page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })).toHaveValue("test-secret-clear-on-workspace");
  await page.evaluate(() => { localStorage.setItem("bc_workspace", JSON.stringify({ shop: { holdingcode: "NEXT" } })); window.dispatchEvent(new Event("bc-workspace-changed")); });
  await expect(page.getByLabel("Token สำหรับเชื่อมต่อ", { exact: true })).toHaveCount(0);
  await expect(page.getByText("Holding: NEXT", { exact: true })).toBeVisible();
});

test("denied screen can reload after server-side access is corrected", async ({ page }) => {
  let permitted = false;
  await page.route("**/api/mcp-tokens**", (route) => route.fulfill(permitted
    ? { json: { success: true, data: route.request().url().endsWith("/companies") ? companies : [] } }
    : { status: 403, json: { success: false, message: "ปฏิเสธสิทธิ์" } }));
  await page.goto("/mcp-tokens");
  await expect(page.getByText("ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")).toBeVisible();
  await expect(page.getByText("Holding: TEST", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: /^เพิ่ม (API|MCP) token$/ })).toHaveCount(0);
  permitted = true;
  await page.getByRole("button", { name: "โหลดใหม่", exact: true }).click();
  await expect(page.getByRole("button", { name: "เพิ่ม MCP token", exact: true })).toBeEnabled();
  await expect(page.getByRole("button", { name: "เพิ่ม API token", exact: true })).toBeEnabled();
  await expect(page.getByText("ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")).toHaveCount(0);
});

for (const permitted of [true, false]) {
test(`Holding settings sidebar keeps token management visible (${permitted ? "admin" : "denied"})`, async ({ page }, info) => {
  const holding = { holdingcode: "TEST", name: "Holding test", role: 2, iscreator: true, language: "th", activelanguages: ["th", "en"], languageconfigs: [{ code: "th", name: "ไทย", isuse: true, isdefault: true }], companies: [{ code: "C01", guidfixed: "company-1", names: [{ code: "th", name: "บริษัทหนึ่ง" }] }], branches: [] };
  await page.route("**/api/language/**", (route) => route.fulfill({ json: {} }));
  await page.route("**/api/workspace/**", (route) => {
    const path = new URL(route.request().url()).pathname;
    const data = path.endsWith("/holdings") ? [holding] : path.endsWith("/holding-info") ? holding : [];
    return route.fulfill({ json: { success: true, data } });
  });
  await page.route("**/api/system-settings/**", (route) => route.fulfill({ json: { success: true, data: [] } }));
  let tokenRequests = 0;
  await page.route("**/api/mcp-tokens**", (route) => {
    tokenRequests++;
    return route.fulfill(permitted
      ? { json: { success: true, data: route.request().url().endsWith("/companies") ? companies : [] } }
      : { status: 403, json: { success: false, message: "ปฏิเสธสิทธิ์" } });
  });
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto("/workspace");
  await page.getByRole("button", { name: /ตั้งค่าระบบ/ }).first().click();
  const link = page.getByRole("link", { name: "จัดการ API / MCP token", exact: true });
  await expect(link).toBeVisible();
  await page.locator("aside").evaluate((element) => { element.scrollTop = element.scrollHeight; });
  const linkBox = (await link.boundingBox())!;
  expect(linkBox.y + linkBox.height).toBeLessThan(720 - 48);
  const audit = page.getByRole("button", { name: /ตรวจสอบสิทธิ์/ });
  await expect(audit).toBeVisible();
  expect((await link.boundingBox())!.y).toBeGreaterThan((await audit.boundingBox())!.y);
  await page.screenshot({ path: info.outputPath("holding-token-navigation.png"), fullPage: true });
  expect(tokenRequests).toBe(0);
  await link.click();
  await expect(page).toHaveURL(/\/mcp-tokens$/);
  if (!permitted) {
    await expect(page.getByText("ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")).toBeVisible();
    await expect(page.getByRole("button", { name: /^เพิ่ม (API|MCP) token$/ })).toHaveCount(0);
    return;
  }
  await expect(page.getByText("Holding: TEST", { exact: true })).toBeVisible();
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem("bc_workspace") ?? "{}").company)).toBeUndefined();
  await page.getByRole("button", { name: "เพิ่ม MCP token", exact: true }).click();
  await expect(page.getByRole("checkbox", { name: "C01 — บริษัทหนึ่ง" })).toBeVisible();
  await expect(page.getByRole("checkbox", { name: "C02 — บริษัทสอง" })).toBeVisible();
});
}
