// Smooth visible mouse for Playwright MCP demos (headed browser).
//
// CDP-driven page.mouse.* dispatches synthetic events: the real OS cursor
// does not move. This helper injects a cursor overlay div that glides to the
// target while page.mouse.move follows the same path, so a human watching the
// headed window sees the pointer travel and click.
//
// browser_run_code_unsafe runs each call in a fresh vm, so state lives in the
// page (window.__smoothJob / __smoothPos). Two-step usage:
//
//   1) browser_evaluate:      () => { window.__smoothJob = { do: "click", selector: "text=ข้อมูลหลัก" } }
//   2) browser_run_code_unsafe: filename: "tools/smooth-mouse.js"
//
// Job shapes:
//   { do: "click", selector: "<playwright selector>" }
//   { do: "type",  selector: "<playwright selector>", text: "...", delayMs?: 40 }   // append
//   { do: "set",   selector: "<playwright selector>", text: "...", delayMs?: 40 }   // clear first
//   { do: "move",  x: 400, y: 300, steps?: 30 }

async (page) => {
  const CURSOR_ID = "__smooth_cursor_overlay";
  const STEP_MS = 12; // 30 steps ≈ 360ms glide

  const sleep = (ms) => page.waitForTimeout(ms);

  async function ensureCursor() {
    await page.evaluate((id) => {
      if (document.getElementById(id)) return;
      const el = document.createElement("div");
      el.id = id;
      el.innerHTML =
        '<svg width="30" height="30" viewBox="0 0 24 24">' +
        '<path d="M4 2l16 9-7 1.5L9 20z" fill="#e11d48" stroke="#fff" stroke-width="1.6"/>' +
        "</svg>";
      Object.assign(el.style, {
        position: "fixed",
        top: "0",
        left: "0",
        zIndex: "2147483647",
        pointerEvents: "none",
        transform: "translate(-100px,-100px)",
        filter: "drop-shadow(0 2px 4px rgba(0,0,0,.45))",
      });
      document.body.appendChild(el);
    }, CURSOR_ID);
  }

  async function paint(x, y, scale) {
    await page.evaluate(
      ([id, x, y, s]) => {
        const el = document.getElementById(id);
        if (el) el.style.transform = `translate(${x - 4}px, ${y - 2}px) scale(${s})`;
      },
      [CURSOR_ID, x, y, scale ?? 1]
    );
  }

  async function currentPos() {
    const saved = await page.evaluate(() => window.__smoothPos || null);
    if (saved) return saved;
    const vp = page.viewportSize() || { width: 1280, height: 720 };
    return { x: Math.round(vp.width / 2), y: Math.round(vp.height / 2) };
  }

  async function move(x, y, steps = 30) {
    await ensureCursor();
    const from = await currentPos();
    const n = Math.max(2, steps);
    for (let i = 1; i <= n; i++) {
      const xi = from.x + ((x - from.x) * i) / n;
      const yi = from.y + ((y - from.y) * i) / n;
      await page.mouse.move(xi, yi);
      await paint(xi, yi);
      await sleep(STEP_MS);
    }
    await page.evaluate((p) => (window.__smoothPos = p), { x, y });
    return { movedTo: { x, y } };
  }

  async function pointOf(selector) {
    const loc = page.locator(selector).first();
    await loc.scrollIntoViewIfNeeded();
    const box = await loc.boundingBox();
    if (!box) throw new Error("element has no bounding box: " + selector);
    return { x: box.x + box.width / 2, y: box.y + box.height / 2 };
  }

  async function click(selector) {
    const p = await pointOf(selector);
    await move(p.x, p.y);
    await paint(p.x, p.y, 0.7); // press-in pulse
    await sleep(120);
    await page.mouse.down();
    await sleep(60);
    await page.mouse.up();
    await paint(p.x, p.y, 1);
    return { clicked: selector, at: p };
  }

  async function type(selector, text, delayMs = 40) {
    await click(selector);
    await page.keyboard.type(text, { delay: delayMs });
    return { typed: text.length, into: selector };
  }

  async function set(selector, text, delayMs = 40) {
    await click(selector);
    await page.keyboard.press("ControlOrMeta+a");
    await sleep(50);
    await page.keyboard.press("Delete");
    await sleep(50);
    await page.keyboard.type(text, { delay: delayMs });
    return { setTo: text, into: selector };
  }

  const job = await page.evaluate(() => window.__smoothJob || null);
  if (!job) {
    return 'no job — set window.__smoothJob first, e.g. { do: "click", selector: "text=..." }';
  }
  await page.evaluate(() => delete window.__smoothJob);

  if (job.do === "click") return await click(job.selector);
  if (job.do === "type") return await type(job.selector, job.text, job.delayMs);
  if (job.do === "set") return await set(job.selector, job.text, job.delayMs);
  if (job.do === "move") return await move(job.x, job.y, job.steps);
  throw new Error("unknown job: " + job.do);
}