// Proof-of-concept: launch Chrome exactly like Playwright MCP will after the
// session restart — persistent profile + chromiumSandbox:true (no --no-sandbox).
// Run: node tools/test-mcp-chrome-launch.js   (Ctrl+C or close window to exit)
const { chromium } = require("playwright");

(async () => {
  const ctx = await chromium.launchPersistentContext(
    "C:/Users/jatur/AppData/Local/bcai-chrome-profile",
    {
      channel: "chrome",
      headless: false,
      chromiumSandbox: true,
      viewport: null,
    }
  );
  const page = ctx.pages()[0] || (await ctx.newPage());
  await page.goto("http://127.0.0.1:3000/");
  console.log("LAUNCHED OK — window is open, banner should be absent");
  await new Promise(() => {}); // keep alive until killed
})();
