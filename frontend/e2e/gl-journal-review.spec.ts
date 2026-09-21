import { expect, test } from "@playwright/test";
import { emptyJournal } from "../src/lib/general-ledger";
import { colorThemes, defaultColorTheme } from "../src/lib/theme-data";

// Actual journal screen, isolated HTTP fixtures: no accounting records are written.
test("review notes, history, concurrency reload and unsaved navigation", async ({ page }, info) => {
  const journals = [1, 2].map((n) => ({ ...emptyJournal(), id: `review-${n}`, docno: `JV-REVIEW-${n}`, version: 4, fiscalyear: "2026", description: `เอกสารตรวจสอบ ${n}` }));
  let eventno = 7, status = 1, stale = false;
  const events: Record<string, unknown>[] = [];
  const commands: Record<string, unknown>[] = [];
  await page.addInitScript(() => localStorage.setItem("bc_auth", JSON.stringify({ username: "review-test", backendUrl: location.origin, holdingcode: "test" })));
  await page.route("**/api/auth/refresh", (route) => route.fulfill({ json: { success: true, token: "test-only" } }));
  await page.route("**/api/gl/**", async (route) => {
    const path = new URL(route.request().url()).pathname.replace("/api/gl/", "");
    let data: unknown;
    if (path === "command") {
      const body = route.request().postDataJSON(); commands.push(body);
      status = body.review.status; eventno++;
      events.unshift({ eventno, version: 4, status, note: body.review.note, reviewedby: "ผู้ตรวจทดสอบ", reviewedat: "2026-09-20T10:00:00Z" });
      data = { id: journals[0].id, version: 4 };
    } else if (path.startsWith("journal-reviews/")) {
      data = { journalid: path.split("/")[1], version: stale ? 5 : 4, status, eventno, events };
    } else if (path.startsWith("journals/")) {
      data = { ...journals.find((journal) => journal.id === path.split("/")[1]), version: stale ? 5 : 4 };
    } else {
      data = { items: path === "journals" ? journals : [], total: path === "journals" ? 2 : 0, page: 1, limit: 30, sequence: 0 };
    }
    await route.fulfill({ json: { success: true, data } });
  });
  await page.goto("/gl/journals");
  const first = page.getByRole("row").filter({ hasText: "JV-REVIEW-1" });
  const second = page.getByRole("row").filter({ hasText: "JV-REVIEW-2" });
  await first.click();
  const panel = page.getByRole("region", { name: "ผลตรวจและข้อแตกต่าง" });
  await expect(panel.getByText("รอตรวจ", { exact: true })).toBeVisible();
  await panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true }).click();
  await panel.getByLabel("สถานะผลตรวจ").selectOption("2");
  await expect(panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true })).toBeDisabled();
  await expect(panel.getByText("กรุณาระบุข้อแตกต่างที่พบ")).toBeVisible();
  await panel.getByLabel("หมายเหตุการตรวจ / ข้อแตกต่าง").fill("  ยอดธนาคารไม่ตรง  ");
  await second.click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.getByRole("dialog").getByRole("button", { name: "ยกเลิก", exact: true }).last().click();
  await expect(panel.getByLabel("หมายเหตุการตรวจ / ข้อแตกต่าง")).toHaveValue("  ยอดธนาคารไม่ตรง  ");
  await panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true }).click();
  await expect(panel.getByText("บันทึกผลตรวจแล้ว", { exact: true })).toBeVisible();
  expect(commands).toHaveLength(1);
  expect(commands[0]).toMatchObject({ resource: "journals", action: "review", id: "review-1", version: 4, review: { status: 2, note: "ยอดธนาคารไม่ตรง", expectedEventNo: 7 } });
  await panel.getByText("ประวัติผลตรวจ", { exact: true }).click();
  await expect(panel.getByText("ยอดธนาคารไม่ตรง", { exact: true })).toBeVisible();
  await expect(panel.getByText(/ผู้ตรวจทดสอบ/)).toBeVisible();
  await page.setViewportSize({ width: 1600, height: 1000 });
  await page.screenshot({ path: info.outputPath("review-light.png"), fullPage: true });
  await page.evaluate((vars) => { document.documentElement.dataset.theme = "dark"; for (const [key, value] of Object.entries(vars)) document.documentElement.style.setProperty(key, value); }, colorThemes.find((theme) => theme.id === defaultColorTheme)!.dark);
  await page.screenshot({ path: info.outputPath("review-dark.png"), fullPage: true });
  await page.evaluate((vars) => { document.documentElement.dataset.theme = "light"; for (const [key, value] of Object.entries(vars)) document.documentElement.style.setProperty(key, value); }, colorThemes.find((theme) => theme.id === defaultColorTheme)!.light);
  await panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true }).click();
  await panel.getByLabel("สถานะผลตรวจ").selectOption("3");
  await second.click();
  await page.getByRole("dialog").getByRole("button", { name: "ละทิ้งการแก้ไข", exact: true }).click();
  await expect(panel.getByLabel("สถานะผลตรวจ")).toHaveCount(0);
  expect(commands).toHaveLength(1);
  // Simulate review loading a newer journal revision than the already loaded document.
  await page.route("**/api/gl/journal-reviews/review-1", async (route) => {
    stale = true;
    await route.fulfill({ json: { success: true, data: { journalid: "review-1", version: 5, status: 1, eventno, events } } });
  });
  await first.click();
  await expect(panel.getByText("เอกสารเปลี่ยนแล้ว กรุณาโหลดใหม่", { exact: true })).toBeVisible();
  await expect(panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true })).toHaveCount(0);
  await panel.getByRole("button", { name: "โหลดเอกสารล่าสุด" }).click();
  await expect(panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true })).toBeVisible();
  expect(commands).toHaveLength(1);
  await page.route("**/api/gl/command", (route) => route.fulfill({ json: { success: false, code: "version_conflict", message: "ข้อมูลถูกแก้ไขโดยผู้ใช้อื่น กรุณาโหลดใหม่" } }));
  await panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true }).click();
  await panel.getByLabel("สถานะผลตรวจ").selectOption("2");
  await panel.getByLabel("หมายเหตุการตรวจ / ข้อแตกต่าง").fill("รอตรวจเอกสารใหม่");
  await panel.getByRole("button", { name: "บันทึกผลตรวจ", exact: true }).click();
  await expect(panel.getByText("ข้อมูลถูกแก้ไขโดยผู้ใช้อื่น กรุณาโหลดใหม่", { exact: true })).toBeVisible();
  await expect(panel.getByLabel("หมายเหตุการตรวจ / ข้อแตกต่าง")).toBeVisible();
  await expect(panel.getByLabel("หมายเหตุการตรวจ / ข้อแตกต่าง")).toHaveValue("รอตรวจเอกสารใหม่");
  await second.click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.getByRole("dialog").getByRole("button", { name: "ยกเลิก", exact: true }).last().click();
  expect(commands).toHaveLength(1);
});
