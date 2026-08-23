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
 * failure on this exact file (2026-07-08 regression) showed a small
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
 * The UI's "เพิ่มกลุ่มย่อย" path now has direct CRUD regression coverage below. The drag-specific
 * tests still seed exact parent/xorder values through the same record PUT used by the drag handler,
 * keeping those tests focused on reorder/reparent persistence rather than duplicating UI creation.
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

type ProductGroupRecord = {
  guidfixed: string;
  code?: string;
  parentguid?: string;
  parentguidall?: string;
  names?: { code: string; name: string }[];
  xsorts?: { code: string; xorder: number }[];
};

const createdGroupGuids = new WeakMap<Page, string[]>();

function rememberCreatedGroup(page: Page, guid: string) {
  const guids = createdGroupGuids.get(page) ?? [];
  guids.push(guid);
  createdGroupGuids.set(page, guids);
}

function getAuthHeaders(page: Page) {
  return page.evaluate(() => {
    const auth = JSON.parse(localStorage.getItem("bc_auth")!) as { token: string; backendUrl: string };
    return { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl };
  });
}

test.afterEach(async ({ page }) => {
  const guids = createdGroupGuids.get(page) ?? [];
  if (guids.length === 0) return;

  const headers = await getAuthHeaders(page);
  for (const guid of [...guids].reverse()) {
    const response = await page.request.delete(
      `/api/system-settings/productgroup/${guid}?holdingcode=test`,
      { headers },
    );
    if (!response.ok() && response.status() !== 404) {
      throw new Error(`Failed to clean up product group ${guid}: HTTP ${response.status()}`);
    }
  }
  createdGroupGuids.delete(page);
});

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

test("productgroup — full UI CRUD nests subgroup and deletes code-less groups", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `UATกลุ่มหลัก${uid}`;
  const updatedRootName = `UATกลุ่มหลักแก้ไข${uid}`;
  const childName = `UATกลุ่มย่อย${uid}`;

  await loginAndSearchMenu(page, "กลุ่มสินค้า");
  await clickButtonByText(page, /^กลุ่มสินค้า$/);
  await page.waitForTimeout(2000);

  const headers = await getAuthHeaders(page);
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;
  const activeList = async () => {
    const response = await page.request.get(`${MAINAPI}/product/group/list?limit=100000`, { headers });
    expect(response.ok(), "fresh product-group list read").toBe(true);
    return ((await response.json()) as { data: ProductGroupRecord[] }).data;
  };
  const primaryName = (record: ProductGroupRecord) =>
    record.names?.find((name) => name.code === "th")?.name ?? record.names?.[0]?.name ?? "";

  await clickButtonByText(page, /^เพิ่มกลุ่มหลัก$/);
  await fillPrimaryNameInput(page, rootName);
  const [createRootResponse] = await Promise.all([
    page.waitForResponse(
      (response) =>
        response.url().includes("/api/system-settings/productgroup") &&
        response.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(createRootResponse.ok(), "create root response").toBe(true);
  const rootGuid = ((await createRootResponse.json()) as { id: string }).id;
  expect(rootGuid, "created root guid").toBeTruthy();
  rememberCreatedGroup(page, rootGuid);

  const rootReadResponse = await page.request.get(recordUrl(rootGuid), { headers });
  expect(rootReadResponse.ok(), "fresh root detail read").toBe(true);
  const rootRecord = ((await rootReadResponse.json()) as { data: ProductGroupRecord }).data;
  expect(primaryName(rootRecord)).toBe(rootName);
  expect(rootRecord.parentguid ?? "").toBe("");

  const search = page.getByPlaceholder("ค้นหา...");
  await search.fill(rootName);
  await expect(page.getByText(rootName, { exact: true })).toBeVisible();
  await page.getByText(rootName, { exact: true }).click();
  const addSubgroupButtons = page.getByRole("button", { name: "เพิ่มกลุ่มย่อย", exact: true });
  await expect(addSubgroupButtons).toHaveCount(2);
  await addSubgroupButtons.first().click();
  await fillPrimaryNameInput(page, childName);
  const [createChildResponse] = await Promise.all([
    page.waitForResponse(
      (response) =>
        response.url().includes("/api/system-settings/productgroup") &&
        response.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(createChildResponse.ok(), "create subgroup response").toBe(true);
  const childGuid = ((await createChildResponse.json()) as { id: string }).id;
  expect(childGuid, "created subgroup guid").toBeTruthy();
  rememberCreatedGroup(page, childGuid);

  const childReadResponse = await page.request.get(recordUrl(childGuid), { headers });
  expect(childReadResponse.ok(), "fresh subgroup detail read").toBe(true);
  const childRecord = ((await childReadResponse.json()) as { data: ProductGroupRecord }).data;
  expect(primaryName(childRecord)).toBe(childName);
  expect(childRecord.parentguid).toBe(rootGuid);
  expect(childRecord.parentguidall).toBe(rootGuid);
  expect(childRecord.xsorts?.[0]?.xorder).toBe(1);

  await page.goto("/productgroup", { waitUntil: "domcontentloaded" });
  await expect(search).toBeVisible({ timeout: 10_000 });
  await search.fill(childName);
  await expect(page.getByText(childName, { exact: true })).toBeVisible();

  await search.fill(rootName);
  await expect(page.getByRole("button", { name: "แก้ไข", exact: true })).toHaveCount(1);
  await page.getByRole("button", { name: "แก้ไข", exact: true }).click();
  const editRootForm = page.getByRole("form", { name: "แก้ไข" });
  await expect(editRootForm).toBeVisible();
  const inlineActions = editRootForm.getByTestId("inline-form-actions");
  const editRootHeading = editRootForm.getByRole("heading", { level: 2 });
  await expect(inlineActions).toBeInViewport();
  const [inlineActionsBox, editRootHeadingBox] = await Promise.all([
    inlineActions.boundingBox(),
    editRootHeading.boundingBox(),
  ]);
  expect(inlineActionsBox).not.toBeNull();
  expect(editRootHeadingBox).not.toBeNull();
  expect(inlineActionsBox!.y).toBeLessThan(editRootHeadingBox!.y + 64);
  await fillPrimaryNameInput(page, updatedRootName);
  const [updateRootResponse] = await Promise.all([
    page.waitForResponse(
      (response) =>
        response.url().includes(`/api/system-settings/productgroup/${rootGuid}`) &&
        response.request().method() === "PUT",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(updateRootResponse.ok(), "update root response").toBe(true);

  const updatedRootReadResponse = await page.request.get(recordUrl(rootGuid), { headers });
  expect(updatedRootReadResponse.ok(), "fresh updated-root detail read").toBe(true);
  const updatedRoot = ((await updatedRootReadResponse.json()) as { data: ProductGroupRecord }).data;
  expect(primaryName(updatedRoot)).toBe(updatedRootName);
  expect(updatedRoot.parentguid ?? "").toBe("");
  await search.fill(updatedRootName);
  await expect(page.getByText(updatedRootName, { exact: true })).toBeVisible();

  const deleteFromUI = async (name: string, guid: string) => {
    await search.fill(name);
    await expect(page.getByText(name, { exact: true })).toBeVisible();
    await page.getByText(name, { exact: true }).click();
    await expect(page.getByRole("button", { name: "ลบ", exact: true })).toHaveCount(1);
    await page.getByRole("button", { name: "ลบ", exact: true }).click();
    const dialog = page.getByRole("dialog");
    await expect(dialog).toBeVisible();
    const [deleteResponse] = await Promise.all([
      page.waitForResponse(
        (response) =>
          response.url().includes(`/api/system-settings/productgroup/${guid}`) &&
          response.request().method() === "DELETE",
      ),
      dialog.getByRole("button", { name: "ลบ", exact: true }).click(),
    ]);
    expect(deleteResponse.ok(), `delete ${name} response`).toBe(true);
    await expect(page.getByText(name, { exact: true })).toHaveCount(0);
    expect((await activeList()).some((record) => record.guidfixed === guid)).toBe(false);
  };

  await deleteFromUI(childName, childGuid);
  await deleteFromUI(updatedRootName, rootGuid);
});

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
  rememberCreatedGroup(page, rootGuid);
  await page.waitForTimeout(1000);

  const H = await getAuthHeaders(page);
  const listUrl = `${MAINAPI}/product/group/list?limit=100000`;
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;

  // Seed exact parent/xorder values through the same record-PUT contract used by drag persistence;
  // the full CRUD test above covers the real "เพิ่มกลุ่มย่อย" UI path.
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
    rememberCreatedGroup(page, childGuid);
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
  rememberCreatedGroup(page, rootGuid);
  await page.waitForTimeout(1000);

  const H = await getAuthHeaders(page);
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;
  const listUrl = `${MAINAPI}/product/group/list?limit=100000`;

  // Seed the exact "child1, child2 both under root" baseline through the same record-PUT contract
  // used by drag persistence; the full CRUD test above covers the real subgroup-creation UI.
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
    rememberCreatedGroup(page, childGuid);
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

/**
 * Regression coverage for the 2026-07-09 addition to `product-group-tree-view.tsx`: per-row
 * up/down move buttons (call the same `reorderGroup()` the drag handler uses) plus a session-local
 * Undo/Redo stack that replays moves through the real `saveGroupRecord`/`saveXSorts` calls (not a
 * local-only revert). Independently re-verified live this session via claude-in-chrome real button
 * clicks + a real pointer drag reparent, each checked against a fresh API read -- see
 * the 2026-07-09 regression run. Drives the same real button elements and the real `xsort` PUT the
 * UI uses, asserting against fresh API reads (not the optimistic in-memory state) at every step.
 */
test("productgroup — up/down move buttons + undo/redo reorder and revert via the real xsort API", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eกลุ่มปุ่ม${uid}`;
  const childNames = [`E2Eปุ่มลูก1_${uid}`, `E2Eปุ่มลูก2_${uid}`];

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
  rememberCreatedGroup(page, rootGuid);
  await page.waitForTimeout(1000);

  const H = await getAuthHeaders(page);
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;
  const listUrl = `${MAINAPI}/product/group/list?limit=100000`;

  // Seed exact parent/xorder values through the same record-PUT contract used by drag persistence;
  // the full CRUD test above covers the real subgroup-creation UI.
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
    rememberCreatedGroup(page, childGuid);
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  const [child1Guid, child2Guid] = childGuids;
  for (let i = 0; i < childGuids.length; i++) {
    const before = ((await (await page.request.get(recordUrl(childGuids[i]), { headers: H })).json()) as {
      data: Record<string, unknown>;
    }).data;
    const nestRes = await page.request.put(recordUrl(childGuids[i]), {
      headers: { ...H, "Content-Type": "application/json" },
      data: { ...before, parentguid: rootGuid, parentguidall: rootGuid, xsorts: [{ code: "X", xorder: i + 1 }] },
    });
    expect(nestRes.ok(), `nest child${i + 1} under root`).toBe(true);
  }

  // Reload so the tree view picks up the freshly-nested children (fresh JS memory, no stale
  // in-memory overrides), then expand the root to reveal the up/down buttons on its children.
  await page.goto("/productgroup", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await page.evaluate((guid) => {
    const row = document.querySelector<HTMLElement>(`[data-group-row-guid="${guid}"]`);
    const toggle = row?.querySelector<HTMLButtonElement>("button");
    if (!toggle) throw new Error("No expand toggle found on root row");
    toggle.click();
  }, rootGuid);
  await page.waitForTimeout(600);

  const orderOf = async () => {
    const data = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
      guidfixed: string;
      xsorts?: { xorder: number }[];
    }[]);
    return {
      child1: data.find((r) => r.guidfixed === child1Guid)?.xsorts?.[0]?.xorder,
      child2: data.find((r) => r.guidfixed === child2Guid)?.xsorts?.[0]?.xorder,
    };
  };

  // Boundary check: the first sibling's up button and the last sibling's down button must be
  // disabled (can't move a first child further up, or a last child further down).
  const boundaryDisabled = await page.evaluate(({ c1, c2 }) => {
    const row1 = document.querySelector<HTMLElement>(`[data-group-row-guid="${c1}"]`);
    const row2 = document.querySelector<HTMLElement>(`[data-group-row-guid="${c2}"]`);
    const up1 = row1?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายขึ้น"]');
    const down2 = row2?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายลง"]');
    return { firstUpDisabled: up1?.disabled, lastDownDisabled: down2?.disabled };
  }, { c1: child1Guid, c2: child2Guid });
  expect(boundaryDisabled.firstUpDisabled, "first sibling's up button is disabled").toBe(true);
  expect(boundaryDisabled.lastDownDisabled, "last sibling's down button is disabled").toBe(true);

  expect(await orderOf(), "children start in creation order").toEqual({ child1: 1, child2: 2 });

  // Click the "ย้ายลง" (move down) button on child1's row -- calls the exact same `reorderGroup()`
  // the drag handler uses, just with a fixed "before"/"after" target instead of a pointer position.
  const [downRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    page.evaluate((guid) => {
      const row = document.querySelector<HTMLElement>(`[data-group-row-guid="${guid}"]`);
      const btn = row?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายลง"]');
      if (!btn) throw new Error("No move-down button found on child1's row");
      btn.click();
    }, child1Guid),
  ]);
  expect(downRes.ok(), "move-down xsort PUT").toBe(true);
  expect(await orderOf(), "down button swapped child1/child2 via the real API").toEqual({ child1: 2, child2: 1 });

  // Undo -- must revert through the real record/xsort API (`applyMoveSnapshot`), not just local state.
  const [undoRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^เลิกทำ$/),
  ]);
  expect(undoRes.ok(), "undo xsort PUT").toBe(true);
  expect(await orderOf(), "undo restored creation order via the real API").toEqual({ child1: 1, child2: 2 });

  // Redo -- must reapply the exact same move via the real API.
  const [redoRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^ทำซ้ำ$/),
  ]);
  expect(redoRes.ok(), "redo xsort PUT").toBe(true);
  expect(await orderOf(), "redo re-applied the swap via the real API").toEqual({ child1: 2, child2: 1 });
});

/**
 * Regression coverage for the 2026-07-09 (cont. 2) drag-jank redesign's independent re-verification:
 * the existing undo/redo test above only exercises a single linear undo-then-redo sequence, which
 * does not catch a redo tail that fails to clear after a genuinely NEW move is made post-undo (i.e.
 * a stale/leftover "future" branch of history incorrectly staying replayable after being superseded).
 * Standard undo/redo semantics require a new action taken after an undo to discard that redo tail --
 * confirmed still correct here via live browser testing in the 2026-07-09 regression run,
 * not previously covered by an automated test. Uses 3 siblings so "move A" and
 * "move B" are unambiguously different operations (not just the same swap re-applied).
 */
test("productgroup — a new move after Undo clears the Redo tail (does not replay the undone move)", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eกลุ่มล้าง${uid}`;
  const childNames = [`E2Eล้าง1_${uid}`, `E2Eล้าง2_${uid}`, `E2Eล้าง3_${uid}`];

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
  rememberCreatedGroup(page, rootGuid);
  await page.waitForTimeout(1000);

  const H = await getAuthHeaders(page);
  const recordUrl = (guid: string) => `/api/system-settings/productgroup/${guid}?holdingcode=test`;
  const listUrl = `${MAINAPI}/product/group/list?limit=100000`;

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
    rememberCreatedGroup(page, childGuid);
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  const [child1Guid, child2Guid, child3Guid] = childGuids;
  for (let i = 0; i < childGuids.length; i++) {
    const before = ((await (await page.request.get(recordUrl(childGuids[i]), { headers: H })).json()) as {
      data: Record<string, unknown>;
    }).data;
    const nestRes = await page.request.put(recordUrl(childGuids[i]), {
      headers: { ...H, "Content-Type": "application/json" },
      data: { ...before, parentguid: rootGuid, parentguidall: rootGuid, xsorts: [{ code: "X", xorder: i + 1 }] },
    });
    expect(nestRes.ok(), `nest child${i + 1} under root`).toBe(true);
  }

  await page.goto("/productgroup", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await page.evaluate((guid) => {
    const row = document.querySelector<HTMLElement>(`[data-group-row-guid="${guid}"]`);
    const toggle = row?.querySelector<HTMLButtonElement>("button");
    if (!toggle) throw new Error("No expand toggle found on root row");
    toggle.click();
  }, rootGuid);
  await page.waitForTimeout(600);

  const orderOf = async () => {
    const data = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
      guidfixed: string;
      xsorts?: { xorder: number }[];
    }[]);
    return {
      child1: data.find((r) => r.guidfixed === child1Guid)?.xsorts?.[0]?.xorder,
      child2: data.find((r) => r.guidfixed === child2Guid)?.xsorts?.[0]?.xorder,
      child3: data.find((r) => r.guidfixed === child3Guid)?.xsorts?.[0]?.xorder,
    };
  };
  const clickMoveDown = async (guid: string) => {
    const [res] = await Promise.all([
      page.waitForResponse((r) => r.url().includes("/xsort") && r.request().method() === "PUT"),
      page.evaluate((g) => {
        const row = document.querySelector<HTMLElement>(`[data-group-row-guid="${g}"]`);
        const btn = row?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายลง"]');
        if (!btn) throw new Error("No move-down button found");
        btn.click();
      }, guid),
    ]);
    expect(res.ok(), "move-down xsort PUT").toBe(true);
  };

  expect(await orderOf(), "children start in creation order").toEqual({ child1: 1, child2: 2, child3: 3 });

  // Move A: swap child1/child2 (down-button on child1) -> child2, child1, child3.
  await clickMoveDown(child1Guid);
  expect(await orderOf(), "move A applied").toEqual({ child1: 2, child2: 1, child3: 3 });

  // Undo move A -> back to creation order.
  const [undoRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^เลิกทำ$/),
  ]);
  expect(undoRes.ok(), "undo xsort PUT").toBe(true);
  expect(await orderOf(), "undo reverted move A").toEqual({ child1: 1, child2: 2, child3: 3 });

  // Redo button must be enabled here (move A is replayable) before we supersede it.
  const redoEnabledBeforeMoveB = await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => (b.textContent ?? "").trim() === "ทำซ้ำ",
    );
    return btn ? !btn.disabled : null;
  });
  expect(redoEnabledBeforeMoveB, "redo is enabled right after undo, before move B").toBe(true);

  // Move B: a DIFFERENT move (swap child2/child3 via down-button on child2) -> child1, child3, child2.
  await clickMoveDown(child2Guid);
  expect(await orderOf(), "move B applied").toEqual({ child1: 1, child2: 3, child3: 2 });

  // The redo tail (move A) must now be cleared -- the button must be disabled, not just unclicked.
  const redoDisabledAfterMoveB = await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => (b.textContent ?? "").trim() === "ทำซ้ำ",
    );
    return btn ? btn.disabled : null;
  });
  expect(redoDisabledAfterMoveB, "redo is disabled after a new move supersedes the undone one").toBe(true);

  // Order must still reflect move B (not move A) -- fresh API read, not optimistic state.
  expect(await orderOf(), "order reflects move B, unaffected by the cleared redo tail").toEqual({
    child1: 1,
    child2: 3,
    child3: 2,
  });

  // Undo move B, back to creation order, for a clean final state.
  const [undoBRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^เลิกทำ$/),
  ]);
  expect(undoBRes.ok(), "undo move B xsort PUT").toBe(true);
  expect(await orderOf(), "undo move B restored creation order").toEqual({ child1: 1, child2: 2, child3: 3 });
});
