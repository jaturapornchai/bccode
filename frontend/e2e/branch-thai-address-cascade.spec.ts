import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the Thai province -> district -> subdistrict -> auto-zipcode cascade picker
 * added directly into the branch ("สาขา") tree-view screen (`/branch`,
 * `frontend/src/app/system-settings/company-branch-tree-view.tsx`, `BranchGeoAddressPicker`).
 *
 * Unlike creditor/debtor (generic `main-crud` config screens using the `thai-address` field type),
 * the branch screen is a bespoke tree-view component: company/branch nodes in a left tree, an
 * inline add/edit form on the right, native `<select>` elements for จังหวัด/อำเภอ/ตำบล (not the
 * shared `ThailandAddressFieldEditor`), and a 4-digit random-code confirm dialog before every
 * save/delete (`ยืนยันการบันทึกข้อมูล` / `ยืนยันการลบข้อมูล`). See `trade-partners-crud.spec.ts` for
 * the base DOM-interaction conventions this file reuses (labels wrap inputs via
 * `input.closest("label")`, native selects need the `HTMLSelectElement` value setter + `change`
 * event to update React state).
 *
 * Backend fields added to `BranchOrgDoc` (`backend/internal/organization/branch/models/branch_org.go`):
 * `countrycode`/`provincecode`/`districtcode`/`subdistrictcode`/`zipcode`, top-level and
 * language-independent (separate from the existing per-language `addresses[]` free-text list).
 */

// The branch tree-view form (unlike the generic `main-crud` FieldEditor screens) renders each
// field as a sibling `<label>` + `<Input>` pair inside a shared wrapper div, not a `<label>` that
// wraps the input — so `input.closest("label")` never matches here. Walk up to the nearest
// common ancestor that contains both a label with matching text and the input, closest wins.
async function fillFieldNearLabel(page: Page, labelPattern: RegExp, value: string) {
  await page.evaluate(
    ({ pattern, value }) => {
      const re = new RegExp(pattern);
      const setValue = (el: HTMLInputElement | HTMLTextAreaElement, v: string) => {
        const proto = el instanceof HTMLTextAreaElement ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype;
        const setter = Object.getOwnPropertyDescriptor(proto, "value")!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const labels = [...document.querySelectorAll<HTMLElement>("label")].filter((l) => re.test(l.innerText ?? ""));
      for (const label of labels) {
        let wrapper: HTMLElement | null = label.parentElement;
        for (let i = 0; i < 4 && wrapper; i++) {
          const input = wrapper.querySelector<HTMLInputElement | HTMLTextAreaElement>("input,textarea");
          if (input) {
            setValue(input, value);
            return;
          }
          wrapper = wrapper.parentElement;
        }
      }
      throw new Error(`No input found near label matching ${pattern}`);
    },
    { pattern: labelPattern.source, value },
  );
}

// The free-text address field (`AddressesEditor`) labels its textarea "ภาษาแรก TH" (language flag +
// code), not the address wording — the address text only appears as the `placeholder`. Match on
// placeholder instead of label text for this one field.
async function fillFieldByPlaceholder(page: Page, placeholder: string, value: string) {
  await page.evaluate(
    ({ placeholder, value }) => {
      const el = document.querySelector<HTMLInputElement | HTMLTextAreaElement>(
        `textarea[placeholder="${placeholder}"],input[placeholder="${placeholder}"]`,
      );
      if (!el) throw new Error(`No input/textarea found with placeholder "${placeholder}"`);
      const proto = el instanceof HTMLTextAreaElement ? window.HTMLTextAreaElement.prototype : window.HTMLInputElement.prototype;
      const setter = Object.getOwnPropertyDescriptor(proto, "value")!.set!;
      setter.call(el, value);
      el.dispatchEvent(new Event("input", { bubbles: true }));
      el.dispatchEvent(new Event("change", { bubbles: true }));
    },
    { placeholder, value },
  );
}

async function clickButtonByText(page: Page, textPattern: RegExp) {
  const deadline = Date.now() + 8000;
  for (;;) {
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
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No clickable element found matching ${textPattern}`);
    await page.waitForTimeout(300);
  }
}

// The "เพิ่มสาขา" (add branch) / "แก้ไขสาขา" (edit branch) / "ลบสาขา" (delete branch) row actions
// are icon-only buttons identified by their `title` attribute, not visible text content. Rows in
// the tree share ancestors a few levels up (the whole branch list), so walking up too far to find
// `withinText` false-matches a DIFFERENT row's action button whose distant ancestor happens to also
// contain the target row's text somewhere else in the list — take the CLOSEST candidate whose own
// row (nearest ancestor containing the row's code badge) matches, not the first one found overall.
async function clickButtonByTitle(page: Page, title: string, withinText?: string) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const clicked = await page.evaluate(
      ({ title, withinText }) => {
        const candidates = [...document.querySelectorAll<HTMLElement>(`[title="${title}"]`)].filter(
          (e) => e.offsetParent !== null,
        );
        if (!withinText) {
          if (candidates[0]) {
            candidates[0].click();
            return true;
          }
          return false;
        }
        // For each candidate, find the smallest ancestor (walking up at most 3 levels — the row
        // wrapper, not the whole list) that contains withinText; pick whichever candidate's row
        // match is narrowest (fewest characters), which is the actual owning row, not a distant
        // list-level ancestor that happens to contain multiple rows' text.
        let best: { btn: HTMLElement; textLen: number } | null = null;
        for (const btn of candidates) {
          let el: HTMLElement | null = btn;
          for (let i = 0; i < 3 && el; i++) {
            el = el.parentElement;
            const text = el?.innerText ?? "";
            if (el && text.includes(withinText)) {
              if (!best || text.length < best.textLen) best = { btn, textLen: text.length };
              break;
            }
          }
        }
        if (best) {
          best.btn.click();
          return true;
        }
        return false;
      },
      { title, withinText },
    );
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No [title="${title}"] button found${withinText ? ` near "${withinText}"` : ""}`);
    await page.waitForTimeout(300);
  }
}

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

async function clickCompanyCard(page: Page) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const clicked = await page.evaluate(() => {
      const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
        (b) => (b.textContent ?? "").includes("บริษัท") && (b.textContent ?? "").trim().length > 15,
      );
      if (btn) {
        btn.click();
        return true;
      }
      return false;
    });
    if (clicked) return;
    if (Date.now() > deadline) throw new Error("No company card button found");
    await page.waitForTimeout(300);
  }
}

async function loginOnly(page: Page) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /test(?!\d)/i);
  await page.waitForTimeout(2500);
  await clickCompanyCard(page);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /สำนักงานใหญ่|headquarters/i);
  await page.waitForTimeout(2500);
  const hasProductUnitModal = await bodyHasText(page, "ยังไม่มีหน่วยนับสินค้า");
  if (hasProductUnitModal) {
    await clickButtonByText(page, /^เข้าเมนูก่อน$/);
    await page.waitForTimeout(1500);
  }
}

// Reads the 4-digit code shown in the save/delete confirm dialog and submits it, matching the
// bespoke `ยืนยันการบันทึกข้อมูล` / `ยืนยันการลบข้อมูล` dialog on the branch tree-view screen (not the
// same confirm-modal shape as the generic `main-crud` screens' "ต้องการลบ" text modal).
async function confirmCodeDialog(page: Page, confirmButtonText: RegExp) {
  await page.waitForTimeout(800);
  const code = await page.evaluate(() => {
    const codeEl = [...document.querySelectorAll<HTMLElement>("span")].find(
      (s) => /^\d{4}$/.test((s.textContent ?? "").trim()) && s.className.includes("tracking-widest"),
    );
    return codeEl?.textContent?.trim() ?? "";
  });
  if (!/^\d{4}$/.test(code)) throw new Error(`Confirm dialog code not found or not 4 digits (got "${code}")`);
  await page.evaluate((c) => {
    const input = document.querySelector<HTMLInputElement>('input[placeholder="กรอกรหัส 4 หลักที่แสดงด้านบน"]');
    if (!input) throw new Error("confirm code input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, c);
    input.dispatchEvent(new Event("input", { bubbles: true }));
    input.dispatchEvent(new Event("change", { bubbles: true }));
  }, code);
  await clickButtonByText(page, confirmButtonText);
  await page.waitForTimeout(1500);
}

// `ThailandAddressSelect`/postal-code `<label>` renders `<label><span>{label}</span><select|input>
// ...`. A `<label>`'s `.innerText` includes its descendant `<select>`'s own option text (browsers
// serialize `<select>` text content into ancestor innerText), so an exact-match label regex like
// `/^จังหวัด$/` never matches — use the label's leading `<span>` text instead, which holds just the
// field label with no option noise.
function setSelect(page: Page, labelPattern: RegExp, valueOrOptionTextPattern: string | RegExp) {
  return page.evaluate(
    ({ pattern, value, isText }) => {
      const re = new RegExp(pattern);
      const spanText = (label: Element | null): string => label?.querySelector("span")?.textContent ?? label?.textContent ?? "";
      const sel = [...document.querySelectorAll<HTMLSelectElement>("select")].find((s) =>
        re.test(spanText(s.closest("label"))),
      );
      if (!sel) throw new Error(`No select found near label matching ${pattern}`);
      let optionValue = value;
      if (isText) {
        const opt = [...sel.options].find((o) => new RegExp(value).test(o.text));
        if (!opt) throw new Error(`No option matching ${value} in select near ${pattern}`);
        optionValue = opt.value;
      }
      const setter = Object.getOwnPropertyDescriptor(window.HTMLSelectElement.prototype, "value")!.set!;
      setter.call(sel, optionValue);
      sel.dispatchEvent(new Event("change", { bubbles: true }));
    },
    {
      pattern: labelPattern.source,
      value: valueOrOptionTextPattern instanceof RegExp ? valueOrOptionTextPattern.source : valueOrOptionTextPattern,
      isText: valueOrOptionTextPattern instanceof RegExp,
    },
  );
}

function readSelectValue(page: Page, labelPattern: RegExp) {
  return page.evaluate((pattern) => {
    const re = new RegExp(pattern);
    const spanText = (label: Element | null): string => label?.querySelector("span")?.textContent ?? label?.textContent ?? "";
    const sel = [...document.querySelectorAll<HTMLSelectElement>("select")].find((s) => re.test(spanText(s.closest("label"))));
    return sel?.value ?? null;
  }, labelPattern.source);
}

// The geo picker hydrates the province/district/subdistrict selects from the branch's saved
// geo fields async (fetch Thailand address dataset, then resolve saved codes against it) — on a
// freshly-opened edit form the selects can briefly show disabled/unselected before that resolves.
// Poll instead of a fixed sleep so the test isn't flaky under CI/headless load.
async function waitForSelectEnabled(page: Page, labelPattern: RegExp, timeoutMs = 8000) {
  const deadline = Date.now() + timeoutMs;
  for (;;) {
    const enabled = await page.evaluate((pattern) => {
      const re = new RegExp(pattern);
      const spanText = (label: Element | null): string => label?.querySelector("span")?.textContent ?? label?.textContent ?? "";
      const sel = [...document.querySelectorAll<HTMLSelectElement>("select")].find((s) => re.test(spanText(s.closest("label"))));
      return Boolean(sel && !sel.disabled && sel.value);
    }, labelPattern.source);
    if (enabled) return;
    if (Date.now() > deadline) throw new Error(`Select near ${labelPattern} did not become enabled+populated in time`);
    await page.waitForTimeout(300);
  }
}

function readInputValue(page: Page, labelPattern: RegExp) {
  return page.evaluate((pattern) => {
    const re = new RegExp(pattern);
    const spanText = (label: Element | null): string => label?.querySelector("span")?.textContent ?? label?.textContent ?? "";
    const input = [...document.querySelectorAll<HTMLInputElement>("input")].find((i) => re.test(spanText(i.closest("label"))));
    return input?.value ?? null;
  }, labelPattern.source);
}

test("branch — Thai geo cascade forward+reverse, addresses untouched on geo-only edit, delete", async ({
  page,
  request,
}) => {
  const uid = Date.now().toString().slice(-4);
  // Branch code must be exactly 5 numeric digits — leading "9" avoids clashing with seeded 00000-00003.
  const numericCode = `9${uid}`;
  const name = `สาขาทดสอบGeoE2E${uid}`;
  const address = `123/45 ถนนทดสอบ E2E ไม่ควรหาย ${uid}`;
  const MAINAPI = "http://localhost:8888";

  await loginOnly(page);
  // Navigate directly rather than via menu search — the branch tree-view is a bespoke screen
  // (not a generic `main-crud` menu item with a guaranteed-unique single-word label to search for).
  await page.goto("/branch", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await expect(page.locator("body")).toContainText(/โครงสร้างองค์กร/);

  // Add a new branch under the first company node via the icon-only "เพิ่มสาขา" button.
  await clickButtonByTitle(page, "เพิ่มสาขา");
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, "เพิ่มสาขาใหม่")).toBe(true);
  await fillFieldNearLabel(page, /รหัสสาขา/, numericCode);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await fillFieldByPlaceholder(page, "ที่อยู่สำหรับออกเอกสาร", address);

  // Forward cascade: province -> district -> subdistrict -> auto zipcode (เชียงใหม่ / เมืองเชียงใหม่ / ศรีภูมิ -> 50200).
  await page.waitForTimeout(1500); // let the Thailand address dataset fetch finish before touching selects
  await setSelect(page, /^จังหวัด$/, /เชียงใหม่ \(50\)/);
  await page.waitForTimeout(500);
  await setSelect(page, /อำเภอ\/เขต/, /เมืองเชียงใหม่/);
  await page.waitForTimeout(500);
  await setSelect(page, /ตำบล\/แขวง/, /^ศรีภูมิ/);
  await page.waitForTimeout(500);
  expect(await readInputValue(page, /รหัสไปรษณีย์/)).toBe("50200");

  await clickButtonByText(page, /^บันทึก$/);
  await confirmCodeDialog(page, /^ยืนยันบันทึก$/);
  await page.waitForTimeout(1500); // left tree refetches after save; the form panel itself stays open (not a redirect)
  expect(await bodyHasText(page, name)).toBe(true);

  // Verify via the real backend API: the 5 new geo fields + the free-text address, both present.
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const listRes = await (await request.get(`${MAINAPI}/organization/branch?limit=200`, { headers: H })).json();
  type BranchDoc = {
    guidfixed: string;
    code: string;
    countrycode?: string;
    provincecode?: string;
    districtcode?: string;
    subdistrictcode?: string;
    zipcode?: string;
    addresses?: { code: string; address: string }[];
  };
  const created = (listRes.data as BranchDoc[]).find((b) => b.code === numericCode);
  expect(created).toBeTruthy();
  const guid = created!.guidfixed;
  expect(created!.provincecode).toBe("50");
  expect(created!.districtcode).toBe("5001");
  expect(created!.subdistrictcode).toBe("500101");
  expect(created!.zipcode).toBe("50200");
  expect(created!.addresses?.[0]?.address).toBe(address);

  // Geo-only edit: change subdistrict (still same province/district), leave the free-text address
  // untouched, save, and confirm via API that addresses[] survived unchanged (no regression to the
  // pre-existing per-language Addresses feature).
  await page.goto("/branch", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickButtonByTitle(page, "แก้ไขสาขา", numericCode);
  await page.waitForTimeout(1500);
  // Confirm the correct row's edit form actually opened before touching the geo picker.
  expect(
    await page.evaluate(
      () => document.querySelector<HTMLInputElement>('input[placeholder="ระบุรหัสสาขา 5 หลัก เช่น 00001"]')?.value ?? null,
    ),
  ).toBe(numericCode);
  await waitForSelectEnabled(page, /อำเภอ\/เขต/);
  await setSelect(page, /ตำบล\/แขวง/, /^พระสิงห์/);
  await page.waitForTimeout(500);
  await clickButtonByText(page, /^บันทึก$/);
  await confirmCodeDialog(page, /^ยืนยันบันทึก$/);
  await page.waitForTimeout(1000);

  const afterGeoEdit = await (await request.get(`${MAINAPI}/organization/branch/${guid}`, { headers: H })).json();
  expect(afterGeoEdit.data.subdistrictcode).toBe("500102");
  expect(afterGeoEdit.data.provincecode).toBe("50");
  expect(afterGeoEdit.data.districtcode).toBe("5001");
  expect(afterGeoEdit.data.addresses?.[0]?.address).toBe(address); // untouched by the geo-only edit

  // Reverse lookup — unique zipcode (10550 -> exactly one subdistrict: คลองด่าน, สมุทรปราการ/บางบ่อ)
  // auto-fills province+district+subdistrict with no ambiguity hint.
  await page.goto("/branch", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickButtonByTitle(page, "แก้ไขสาขา", numericCode);
  await page.waitForTimeout(1500);
  await waitForSelectEnabled(page, /อำเภอ\/เขต/);
  await fillFieldNearLabel(page, /รหัสไปรษณีย์/, "10550");
  await page.waitForTimeout(600);
  expect(await readSelectValue(page, /^จังหวัด$/)).toBe("11");
  expect(await readSelectValue(page, /อำเภอ\/เขต/)).toBe("1102");
  expect(await readSelectValue(page, /ตำบล\/แขวง/)).toBe("110205");

  // Reverse lookup — ambiguous zipcode (83000 spans 2 provinces / 6 subdistricts): the picker must
  // NOT silently guess; province/district are cleared until the user narrows it down, and an
  // ambiguous-match hint is shown instead of crashing or picking one arbitrarily.
  await fillFieldNearLabel(page, /รหัสไปรษณีย์/, "83000");
  await page.waitForTimeout(600);
  const ambiguousHint = await bodyHasText(page, "พบ");
  expect(ambiguousHint).toBe(true);
  expect(await readSelectValue(page, /^จังหวัด$/)).toBe("");

  // Resolve the ambiguous zip by picking a province, then subdistrict — confirms the picker
  // recovers cleanly from the ambiguous state.
  await setSelect(page, /^จังหวัด$/, /ภูเก็ต \(83\)/);
  await page.waitForTimeout(500);
  await setSelect(page, /ตำบล\/แขวง/, /^ตลาดใหญ่/);
  await page.waitForTimeout(500);
  expect(await readInputValue(page, /รหัสไปรษณีย์/)).toBe("83000");

  await clickButtonByText(page, /^บันทึก$/);
  await confirmCodeDialog(page, /^ยืนยันบันทึก$/);
  await page.waitForTimeout(1000);

  const afterReverseLookup = await (
    await request.get(`${MAINAPI}/organization/branch/${guid}`, { headers: H })
  ).json();
  expect(afterReverseLookup.data.provincecode).toBe("83");
  expect(afterReverseLookup.data.districtcode).toBe("8301");
  expect(afterReverseLookup.data.subdistrictcode).toBe("830101");
  expect(afterReverseLookup.data.zipcode).toBe("83000");
  expect(afterReverseLookup.data.addresses?.[0]?.address).toBe(address); // still untouched

  // Cleanup — delete the test branch via its row action + confirm-code dialog.
  await clickButtonByTitle(page, "ลบสาขา", numericCode);
  await page.waitForTimeout(800);
  expect(await bodyHasText(page, "ยืนยันการลบข้อมูล")).toBe(true);
  await confirmCodeDialog(page, /^ยืนยันลบ$/);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, name)).toBe(false);

  const afterDelete = await request.get(`${MAINAPI}/organization/branch/${guid}`, { headers: H });
  expect([400, 404]).toContain(afterDelete.status());
});
