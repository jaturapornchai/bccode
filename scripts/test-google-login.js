// ทดสอบปุ่ม Google login บน production ด้วย Chromium จริง (popup ทำงาน)
// ใช้: node scripts/test-google-login.js [url]
const { chromium } = require('playwright');

(async () => {
  const url = process.argv[2] || 'https://account.bcaicloud.com/';
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();

  const logs = [];
  page.on('console', (msg) => {
    const text = msg.text();
    if (/gsi|google|origin|client|error|popup/i.test(text)) logs.push(`[${msg.type()}] ${text.slice(0, 300)}`);
  });
  page.on('pageerror', (err) => logs.push(`[pageerror] ${String(err).slice(0, 300)}`));

  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 });
  // รอ GIS โหลด + ปุ่ม render
  const frame = page.frameLocator('iframe[src*="accounts.google.com/gsi/button"]').first();
  await frame.locator('[role="button"], .nsm7Bb-HzV7m-LgbsSe').first().waitFor({ timeout: 20000 });
  await page.waitForTimeout(1500);

  const popupPromise = page.waitForEvent('popup', { timeout: 12000 }).catch(() => null);
  await frame.locator('[role="button"], .nsm7Bb-HzV7m-LgbsSe').first().click();
  const popup = await popupPromise;

  if (popup) {
    await popup.waitForLoadState('domcontentloaded').catch(() => {});
    await popup.waitForTimeout(2500);
    console.log('POPUP_OPENED:', popup.url().slice(0, 150));
    console.log('POPUP_TITLE:', (await popup.title()).slice(0, 80));
    const body = await popup.evaluate(() => document.body.innerText.replace(/\s+/g, ' ').slice(0, 250)).catch(() => '(no body)');
    console.log('POPUP_BODY:', body);
    await popup.screenshot({ path: 'test-results/google-popup.png' }).catch(() => {});
  } else {
    console.log('NO_POPUP');
    await page.screenshot({ path: 'test-results/google-nopopup.png' }).catch(() => {});
  }

  console.log('CONSOLE_LOGS:');
  logs.slice(0, 12).forEach((l) => console.log(' ', l));
  await browser.close();
})().catch((err) => {
  console.error('SCRIPT_ERROR:', String(err).slice(0, 300));
  process.exit(1);
});
