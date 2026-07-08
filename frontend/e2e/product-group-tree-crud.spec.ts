import { test, expect, type Page } from "@playwright/test";

/**
 * Regression coverage for the 2026-07-08 fix to `product-group-tree-view.tsx` (`/productgroup`,
 * "กลุ่มสินค้า"), the sibling of the `product-category-tree-view.tsx` drop-zone precision fix
 * covered by `product-category-list-crud.spec.ts`'s two companion tests below "sibling reorder via
 * the real xsort API" / "reparent via the real record+xsort API". Both trees shared the same defect
 * class: `getPointerDropCandidate` resolved the drop target via live `document.elementFromPoint()` +
 * DOM `closest()`, which is contaminated by the drag's own rendered side effects -- the dragged row
 * is `pointer-events-none` while dragging, and this file's drag-only "วางเป็นลูก" chip could
 * out-compete a neighboring row's own zone -- instead of stable row geometry. A live-reproduced
 * failure on this exact file (see `.agents/worklog.md` 2026-07-08 "cont." entries) showed a small
 * in-place jiggle-drag on a leaf row silently reparenting a DIFFERENT, unrelated sibling row into a
 * neighboring group (worse than the category screen's own repro, which at least moved the dragged
 * node itself) -- this file's `previewSiblingReorder` optimistically FLIP-animates rows during the
 * drag, so the DOM being hit-tested was a moving target on top of the ghosted-row defect.
 *
 * FIXED 2026-07-08 (later): `getPointerDropCandidate` (`product-group-tree-view.tsx`) rewritten to
 * resolve purely from live `getBoundingClientRect()` containment against `[data-group-row-guid]`
 * elements, with an explicit guard: if the row whose rect contains the pointer is the dragged row
 * itself, return no-candidate rather than falling through to a nearest-row scan (the exact fallthrough
 * that caused the unrelated-row mutation above). The dead `data-group-child-drop-guid` hit-test
 * attribute was removed from the "วางเป็นลูก" chip (kept as a pure visual affordance).
 *
 * Independently re-verified live via real claude-in-chrome pointer press-move-release drags in a
 * fresh browser session (cleared localStorage/sessionStorage, redone dev-test-login): a 13px in-place
 * jiggle on a nested leaf row is now a correct no-op (zero PUT calls, fresh API read unchanged); an
 * adjacent same-parent sibling swap correctly fires a single `xsort` PUT and persists the new order;
 * dragging one row onto another's middle/"inside" band correctly fires the reparent (record PUT +
 * xsort PUT) and persists new parentguid/parentguidall, confirmed against a fresh API read each time.
 *
 * Same scope note as the category screen's tests: a real Chromium pointer drag is NOT simulated here
 * -- this component's live FLIP-animation preview reflow during a drag makes a scripted multi-tick
 * drag path flaky to author reliably, and the pixel-to-zone geometry decision inside
 * `getPointerDropCandidate`/`getRowDropPosition` runs entirely client-side before any API call, so no
 * API-level assertion can exercise it directly -- only a real/simulated pointer drag can (verified
 * manually this session via claude-in-chrome, not via CI). These tests instead drive the SAME real
 * `xsort` PUT / record PUT `saveXSorts`/`saveGroupRecord` send (same payload shapes), covering the
 * backend/persistence/display half of the flow: that a correctly-resolved target really persists and
 * really displays, for both the "before/after" (reorder) and "inside" (reparent) branches of the fix.
 *
 * Child records below are created as top-level roots via the UI's "เพิ่มกลุ่มหลัก" (Add Root) button,
 * then nested via the same record-PUT contract the fix relies on, INSTEAD of via the UI's own
 * "เพิ่มกลุ่มย่อย" (Add Subgroup) button. This is deliberate, not a shortcut: independently confirmed
 * live this session (see `.agents/worklog.md`) that "เพิ่มกลุ่มย่อย" has a real, separate,
 * already-flagged bug -- `handleOpenGroupCreate`'s `parentguid` is only ever set on FORM state, but
 * `buildPayload` (`system-settings-screen.tsx`) skips every field marked `readOnly` in
 * `system-setting-screens.ts`, and `productgroup`'s `parentguid` field IS `readOnly` (unlike the
 * category screen, which has a dedicated slug-specific block that re-adds `parentguid`/`parentguidall`
 * to the payload after `buildPayload` strips it -- that block does not cover `productgroup`). The
 * create form visibly shows the correct parent guid in the (disabled) "กลุ่มแม่" field, but the save
 * silently sends no `parentguid` at all, so the new row is created as an orphaned ROOT instead of a
 * child, with no error shown to the user. Confirmed via a live create + fresh API read this session
 * (new group's `parentguid`/`parentguidall` both `""` despite a parent being selected). This exact
 * defect was ALSO already flagged by an earlier session -- the live dataset already contains a row
 * literally named "[VERIFY-BUG-เพิ่มกลุ่มย่อย ห้ามใช้ orphan]" for this reason. Out of scope for the
 * drop-zone precision fix this file covers, so NOT fixed here -- flagged separately instead (see
 * worklog + task report). These tests route around it by nesting via the record PUT directly.
 */

async function clickButtonByText(page: Page, textPattern: RegExp) {
  const clicked = await page.evaluate((pattern) => {
    const re = new RegExp(pattern);
    const el = [...document.querySelectorAll<HTMLElement>("button,a,[role=button]")].find(
      (e) => re.test((e.textContent ?? "").trim()) && e.offsetParent !== null,
    );
    if (el) {
      el.click();
      return true;
    }
    return false;
  }, textPattern.source);
  if (!clicked) throw new Error(`No clickable element found matching ${textPattern}`);
}

async function loginAndSearchMenu(page: Page, menuQuery: string) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /Test/);
  await page.waitForTimeout(2500);
  await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => (b.textContent ?? "").includes("บริษัท") && (b.textContent ?? "").trim().length > 15,
    );
    if (!btn) throw new Error("No company card button found");
    btn.click();
  });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /สำนักงานใหญ่|headquarters/i);
  await page.waitForTimeout(2500);
  await page.evaluate((q) => {
    const input = document.querySelector<HTMLInputElement>('input[placeholder*="ค้นหาเมนู"]');
    if (!input) throw new Error("Menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, q);
    input.dispatchEvent(new Event("input", { bubbles: true }));
  }, menuQuery);
  await page.waitForTimeout(1000);
}

const MAINAPI = "http://localhost:8888";

function getAuthHeaders(page: Page) {
  return page.evaluate(() => {
    const auth = JSON.parse(localStorage.getItem("bc_auth")!) as { token: string; backendUrl: string };
    return { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl };
  });
}

/** Fills the primary-language input of the "names" multilingual editor (`NamesEditor` component) --
 *  same anchor strategy as the category screen's `fillPrimaryNameInput`: the wrapping `<label>`
 *  containing "ภาษาแรก" ("Primary language") is the reliable anchor regardless of the resolved
 *  placeholder/code. */
async function fillPrimaryNameInput(page: Page, value: string) {
  await page.evaluate((v) => {
    const label = [...document.querySelectorAll<HTMLElement>("label")].find(
      (l) => (l.textContent ?? "").includes("ภาษาแรก") && l.offsetParent !== null,
    );
    if (!label) throw new Error('No "ภาษาแรก" (primary language) label found');
    const input = label.querySelector<HTMLInputElement>("input");
    if (!input) throw new Error('No input found under the "ภาษาแรก" label');
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, v);
    input.dispatchEvent(new Event("input", { bubbles: true }));
    input.dispatchEvent(new Event("change", { bubbles: true }));
  }, value);
}

test("productgroup — sibling reorder via the real xsort API persists and displays in the new order", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eกลุ่มเรียง${uid}`;
  const childNames = [`E2Eลูกที่1_${uid}`, `E2Eลูกที่2_${uid}`, `E2Eลูกที่3_${uid}`];

  await loginAndSearchMenu(page, "กลุ่มสินค้า");
  await clickButtonByText(page, /^กลุ่มสินค้า$/);
  await page.waitForTimeout(2000);

  // Create the root, then 3 children under it (creation order == initial xorder 1,2,3).
  await clickButtonByText(page, /^เพิ่มกลุ่มหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productgroup") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root group").toBe(true);
  const rootGuid = ((await rootRes.json()) as { id: string }).id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  const H = await getAuthHeaders(page);
  const listUrl = `${MAINAPI}/product/group/list?limit=100000`;
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;

  // Create the 3 children as top-level roots via the UI (the working "เพิ่มกลุ่มหลัก" path -- see the
  // file-level comment above for why "เพิ่มกลุ่มย่อย" is deliberately NOT used), then nest each one
  // under `rootGuid` via the real record-PUT contract, seeding xorder 1,2,3 in creation order.
  const childGuids: string[] = [];
  for (const childName of childNames) {
    await clickButtonByText(page, /^เพิ่มกลุ่มหลัก$/);
    await page.waitForTimeout(500);
    await fillPrimaryNameInput(page, childName);
    const [childRes] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/system-settings/productgroup") && res.request().method() === "POST",
      ),
      clickButtonByText(page, /^บันทึก$/),
    ]);
    expect(childRes.ok(), `create subgroup ${childName}`).toBe(true);
    const childGuid = ((await childRes.json()) as { id: string }).id;
    expect(childGuid).toBeTruthy();
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  for (let i = 0; i < childGuids.length; i++) {
    const childBefore = ((await (await page.request.get(recordUrl(childGuids[i]), { headers: H })).json()) as {
      data: Record<string, unknown>;
    }).data;
    const nestRes = await page.request.put(recordUrl(childGuids[i]), {
      headers: { ...H, "Content-Type": "application/json" },
      data: { ...childBefore, parentguid: rootGuid, parentguidall: rootGuid, xsorts: [{ code: "X", xorder: i + 1 }] },
    });
    expect(nestRes.ok(), `nest child${i + 1} under root`).toBe(true);
  }

  // Confirm the initial order really is 1,2,3 (just seeded) before touching it.
  const before = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
    guidfixed: string;
    xsorts?: { xorder: number }[];
  }[]);
  const orderOf = (recs: typeof before, guid: string) => recs.find((r) => r.guidfixed === guid)?.xsorts?.[0]?.xorder;
  expect(orderOf(before, childGuids[0]), "child1 starts at xorder 1").toBe(1);
  expect(orderOf(before, childGuids[1]), "child2 starts at xorder 2").toBe(2);
  expect(orderOf(before, childGuids[2]), "child3 starts at xorder 3").toBe(3);

  // Drive the SAME real xsort endpoint the drag handler's `saveXSorts` calls, with the same payload
  // shape, to reorder child3 to before child2 (child1, child3, child2).
  const xsortRes = await page.request.put(`/api/system-settings/productgroup/xsort?holdingcode=test`, {
    headers: { ...H, "Content-Type": "application/json" },
    data: [
      { guidfixed: childGuids[0], code: "X", xorder: 1 },
      { guidfixed: childGuids[2], code: "X", xorder: 2 },
      { guidfixed: childGuids[1], code: "X", xorder: 3 },
    ],
  });
  expect(xsortRes.ok(), "xsort reorder request").toBe(true);

  // Verify via a FRESH API read (not the optimistic response) that the reorder really persisted.
  const after = ((await (await page.request.get(listUrl, { headers: H })).json()).data as typeof before);
  expect(orderOf(after, childGuids[0]), "child1 stays at xorder 1").toBe(1);
  expect(orderOf(after, childGuids[2]), "child3 moved to xorder 2").toBe(2);
  expect(orderOf(after, childGuids[1]), "child2 moved to xorder 3").toBe(3);

  // Verify via a hard reload (fresh JS memory, no in-memory/optimistic state) that the real UI
  // displays the new order too, not just that the API accepted the write. Expand the root's caret
  // toggle (the row's own first <button>, no dedicated aria-label on this screen) to reveal children
  // -- selecting the row alone only opens its detail panel, it does not auto-expand the tree.
  await page.goto("/productgroup", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await page.evaluate((guid) => {
    const row = document.querySelector<HTMLElement>(`[data-group-row-guid="${guid}"]`);
    const toggle = row?.querySelector<HTMLButtonElement>("button");
    if (!toggle) throw new Error("No expand toggle found on root row");
    toggle.click();
  }, rootGuid);
  await page.waitForTimeout(600);
  const bodyText = await page.evaluate(() => document.body.innerText);
  const idx1 = bodyText.indexOf(childNames[0]);
  const idx3 = bodyText.indexOf(childNames[2]);
  const idx2 = bodyText.indexOf(childNames[1]);
  expect(idx1, "child1 rendered").toBeGreaterThan(-1);
  expect(idx3, "child3 rendered").toBeGreaterThan(-1);
  expect(idx2, "child2 rendered").toBeGreaterThan(-1);
  expect(idx1 < idx3 && idx3 < idx2, "UI displays child1, child3, child2 in that order after reload").toBe(true);
});

/**
 * Companion to the sibling-reorder test above -- covers the OTHER branch of the same fixed decision
 * (`getPointerDropCandidate` choosing "inside" instead of "before"/"after"): a real reparent. Drives
 * the exact record-PUT + xsort-PUT payload shape `moveGroupAsChild` sends
 * (`product-group-tree-view.tsx`), not a raw arbitrary PUT, so this exercises the real persistence
 * contract the fixed drop-resolution code relies on. Does NOT simulate the pointer drag or the
 * pixel-to-zone geometry decision itself -- see the file-level comment above for why that part can
 * only be covered by a real/simulated drag, not an API-level test.
 */
test("productgroup — reparent via the real record+xsort API (moveGroupAsChild path) persists correct parentguid/parentguidall", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eกลุ่มย้าย${uid}`;
  const childNames = [`E2Eลูกที่1_${uid}`, `E2Eลูกที่2_${uid}`];

  await loginAndSearchMenu(page, "กลุ่มสินค้า");
  await clickButtonByText(page, /^กลุ่มสินค้า$/);
  await page.waitForTimeout(2000);

  await clickButtonByText(page, /^เพิ่มกลุ่มหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productgroup") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root group").toBe(true);
  const rootGuid = ((await rootRes.json()) as { id: string }).id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  const H = await getAuthHeaders(page);
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;
  const listUrl = `${MAINAPI}/product/group/list?limit=100000`;

  // Create both children as top-level roots via the UI (see the file-level comment above for why
  // "เพิ่มกลุ่มย่อย" is deliberately NOT used), then nest both under `rootGuid` via the real
  // record-PUT contract -- establishing the "child1, child2 both under root" baseline the actual
  // reparent-under-test (child2 -> under child1) starts from.
  const childGuids: string[] = [];
  for (const childName of childNames) {
    await clickButtonByText(page, /^เพิ่มกลุ่มหลัก$/);
    await page.waitForTimeout(500);
    await fillPrimaryNameInput(page, childName);
    const [childRes] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/system-settings/productgroup") && res.request().method() === "POST",
      ),
      clickButtonByText(page, /^บันทึก$/),
    ]);
    expect(childRes.ok(), `create subgroup ${childName}`).toBe(true);
    const childGuid = ((await childRes.json()) as { id: string }).id;
    expect(childGuid).toBeTruthy();
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  const [child1Guid, child2Guid] = childGuids;
  for (let i = 0; i < childGuids.length; i++) {
    const childBefore = ((await (await page.request.get(recordUrl(childGuids[i]), { headers: H })).json()) as {
      data: Record<string, unknown>;
    }).data;
    const nestRes = await page.request.put(recordUrl(childGuids[i]), {
      headers: { ...H, "Content-Type": "application/json" },
      data: { ...childBefore, parentguid: rootGuid, parentguidall: rootGuid, xsorts: [{ code: "X", xorder: i + 1 }] },
    });
    expect(nestRes.ok(), `nest child${i + 1} under root`).toBe(true);
  }

  // Reparent child2 to become a child of child1 -- the exact payload shape `moveGroupAsChild` builds:
  // spread the fetched record, override parentguid/parentguidall/xsorts.
  const child2Before = ((await (await page.request.get(recordUrl(child2Guid), { headers: H })).json()) as {
    data: Record<string, unknown>;
  }).data;
  const newParentGuidAll = `${rootGuid},${child1Guid}`;
  const reparentRes = await page.request.put(recordUrl(child2Guid), {
    headers: { ...H, "Content-Type": "application/json" },
    data: { ...child2Before, parentguid: child1Guid, parentguidall: newParentGuidAll, xsorts: [{ code: "X", xorder: 1 }] },
  });
  expect(reparentRes.ok(), "reparent record PUT").toBe(true);
  const xsortRes = await page.request.put(`/api/system-settings/productgroup/xsort?holdingcode=test`, {
    headers: { ...H, "Content-Type": "application/json" },
    data: [
      { guidfixed: child1Guid, code: "X", xorder: 1 }, // renumbered root-level sibling left behind
      { guidfixed: child2Guid, code: "X", xorder: 1 }, // first (only) child of its new parent
    ],
  });
  expect(xsortRes.ok(), "post-reparent xsort PUT").toBe(true);

  // Verify via a FRESH API read (not the optimistic response) that the reparent really persisted,
  // and that the old parent's remaining child (child1) was left structurally untouched.
  const after = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
    guidfixed: string;
    parentguid?: string;
    parentguidall?: string;
    xsorts?: { xorder: number }[];
  }[]);
  const child1After = after.find((r) => r.guidfixed === child1Guid);
  const child2After = after.find((r) => r.guidfixed === child2Guid);
  expect(child1After, "child1 present via API").toBeTruthy();
  expect(child2After, "child2 present via API").toBeTruthy();
  expect(child2After!.parentguid, "child2 now nests under child1, not root").toBe(child1Guid);
  expect(child2After!.parentguidall, "child2's ancestor chain includes root then child1").toBe(newParentGuidAll);
  expect(child1After!.parentguid, "child1 stays under root (untouched by the reparent)").toBe(rootGuid);
  expect(child1After!.xsorts?.[0]?.xorder, "child1 keeps xorder 1 under root").toBe(1);
});
