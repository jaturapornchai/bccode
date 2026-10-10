import { chromium } from "playwright";
import fs from "fs";
import path from "path";

const BASE_URL = process.env.UAT_TARGET_URL || "https://account.bcaicloud.com";
const OUTPUT_DIR = path.resolve("scratch/uat-round7");

if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

async function run() {
  console.log("=== Starting Live Production UAT Round 7 on:", BASE_URL, "===");
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  const consoleErrors = [];
  page.on("pageerror", (err) => {
    console.error("PAGE ERROR:", err.message);
    consoleErrors.push(err.message);
  });
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      consoleErrors.push(msg.text());
    }
  });

  try {
    // 1. Visit Login & Demo Login
    console.log("[1/5] Demo login...");
    await page.goto(BASE_URL, { waitUntil: "networkidle" });
    await page.waitForTimeout(2000);
    const demoButton = page.locator('button:has-text("ทดลองใช้ระบบ (Demo)"), button:has-text("ทดลองใช้ระบบ")');
    if (await demoButton.isVisible()) {
      await demoButton.click();
    } else {
      await page.click('text="ทดลองใช้ระบบ"');
    }
    await page.waitForURL("**/holding", { timeout: 15000 }).catch(() => {});
    await page.waitForTimeout(2500);

    // 2. Select Workspace
    console.log("[2/5] Selecting workspace RUNGRUENG -> 01 -> 00000...");
    const holdingCard = page.locator('button:has-text("RUNGRUENG")').first();
    await holdingCard.click();

    const compBtn = page.locator('button:has-text("บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด"), button:has-text("บริษัท รุ่งเรือง ค้าไม้ จำกัด")').first();
    await compBtn.waitFor({ state: "visible", timeout: 15000 });
    await compBtn.click();

    const branchBtn = page.locator('button:has-text("สำนักงานใหญ่")').first();
    await branchBtn.waitFor({ state: "visible", timeout: 15000 });
    await branchBtn.click();

    await page.waitForURL("**/menu", { timeout: 15000 });
    await page.waitForTimeout(2000);

    // Helper: open GL menu item
    async function openGlMenu(itemText, submenuText = null) {
      console.log(`Navigating menu: บัญชี -> ${submenuText ? submenuText + " -> " : ""}${itemText}`);
      await page.keyboard.press("Escape");
      await page.waitForTimeout(300);

      const glBar = page.getByRole("button", { name: "บัญชี" }).first();
      await glBar.click();
      await page.waitForTimeout(500);

      if (submenuText) {
        const sub = page.locator('[role="menuitem"], [data-radix-collection-item]').filter({ hasText: submenuText }).first();
        if (await sub.isVisible()) {
          await sub.hover();
          await page.waitForTimeout(500);
        }
      }

      const item = page.locator('[role="menuitem"], [data-radix-collection-item]').filter({ hasText: itemText }).last();
      await item.waitFor({ state: "visible", timeout: 8000 });
      await item.click();
      await page.waitForTimeout(3000);
    }

    // --- 1. Test Trial Balance Collapsible Filter ---
    console.log("[3/5] Testing Trial Balance Collapsible Filter...");
    await openGlMenu("รายงานงบทดลอง", "รายงานงบการเงิน");

    const yearTrigger = page.locator('button[aria-label="ปีบัญชี"], button:has-text("เลือกปีบัญชี")').first();
    if (await yearTrigger.isVisible()) {
      await yearTrigger.click();
      await page.waitForTimeout(400);
      const yearItem = page.locator('[role="option"], li, div').filter({ hasText: /2026|2569/ }).first();
      if (await yearItem.isVisible()) {
        await yearItem.click();
        await page.waitForTimeout(600);
      }
    }

    const showBtn = page.locator('button:has-text("แสดงรายงาน")').first();
    await showBtn.waitFor({ state: "visible", timeout: 5000 });
    await showBtn.click();
    await page.waitForTimeout(3000);

    // ตรวจสอบและกดปุ่มย่อเงื่อนไข
    const toggleFilterBtn = page.locator('button:has-text("ย่อเงื่อนไข"), button:has-text("ขยายเงื่อนไข")').first();
    if (await toggleFilterBtn.isVisible()) {
      console.log("  Clicking toggle filter button to collapse...");
      await toggleFilterBtn.click();
      await page.waitForTimeout(1000);
    }

    await page.screenshot({ path: path.join(OUTPUT_DIR, "01-trialbalance-collapsed-filter.png"), fullPage: false });
    console.log("  Saved: 01-trialbalance-collapsed-filter.png");

    // --- 2. Test SME Presets in Journal Entry ---
    console.log("[4/5] Testing SME Presets in Journal Entry...");
    await openGlMenu("บันทึกข้อมูลรายวัน", "สมุดรายวัน");

    // กดปุ่มเพิ่มรายการ
    const addRowBtn = page.locator('button:has-text("เพิ่มรายการ")').first();
    if (await addRowBtn.isVisible()) {
      console.log("  Clicking add row button...");
      await addRowBtn.click();
      await page.waitForTimeout(1500);
    }

    // กดปุ่มเลือกรูปแบบ
    const templateBtn = page.locator('button:has-text("เลือกรูปแบบ"), button[title*="เลือกรูปแบบ"]').first();
    if (await templateBtn.isVisible()) {
      console.log("  Clicking 'เลือกรูปแบบ' button...");
      await templateBtn.click();
      await page.waitForTimeout(1500);
      await page.screenshot({ path: path.join(OUTPUT_DIR, "02-sme-presets-modal.png"), fullPage: false });
      console.log("  Saved: 02-sme-presets-modal.png");

      // ปิด modal
      const closeBtn = page.locator('dialog button:has-text("ปิด")').first();
      if (await closeBtn.isVisible()) {
        await closeBtn.click();
        await page.waitForTimeout(500);
      }
    } else {
      console.log("  Warning: templateBtn not visible");
    }

    // --- 3. Test Bank Unmatched Report ---
    console.log("[5/5] Testing Bank Unmatched Report...");
    await openGlMenu("รายงาน Statement ธนาคารที่ยังไม่จับคู่", "รายงานงบการเงิน");

    const showBankBtn = page.locator('button:has-text("แสดงรายงาน")').first();
    if (await showBankBtn.isVisible()) {
      await showBankBtn.click();
      await page.waitForTimeout(2500);
    }
    await page.screenshot({ path: path.join(OUTPUT_DIR, "03-bank-unmatched-report.png"), fullPage: false });
    console.log("  Saved: 03-bank-unmatched-report.png");

    console.log("=== All Live Production UAT Steps Completed Successfully! ===");
    console.log("Total Console Errors:", consoleErrors.length);
  } catch (err) {
    console.error("UAT FAILED:", err);
    await page.screenshot({ path: path.join(OUTPUT_DIR, "error-state.png") }).catch(() => {});
  } finally {
    await browser.close();
  }
}

run();
