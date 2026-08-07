import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "สูตรผลิต" (Product BOM / Recipe) screen — /productbom.
 * Backing store: MongoDB `appdb.productbarcodeboms` (recipe root doc + embedded bom[] items +
 * boms[] versions; soft delete via `deletedat`). Backend: `internal/product/bom/` (MongoDB-first,
 * recipe components validated against `productbarcodes` with materialtype 1/2/4 only).
 *
 * Covers improvements made 2026-07-02 via a uat-crud-mongo run (screen was the last of the
 * 16-item menu sweep never fully tested):
 *  1. Recipe DELETE button — backend DELETE /product/bom/:id existed but the UI had NO delete
 *     entry point at all (recipes could be created but never removed).
 *  2. Editable cost-per-unit on material rows — `averagecost` existed in the backend BOM item
 *     model but the UI rendered it read-only, permanently ฿0.00 for picker-added items, making
 *     the whole cost-rollup/margin feature useless for restaurants/factories.
 *  3. Output-per-batch quantity (`outputqty` → recipe root `qty`) — new backend field; enables
 *     "1 สูตรผลิตได้ N หน่วย" and cost-per-produced-unit (rollup ÷ N), the number a restaurant
 *     or factory actually prices from.
 *
 * Prerequisite data (seeded once via API, stays in dev DB): material products MAT44941 /
 * MAT244941 with barcodes BMAT44941 / BMAT244941 (materialtype=1). The test creates its own
 * unique recipe and deletes it through the UI, so it is re-runnable without cleanup.
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

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

/** Fill the input that lives in the same container as a <label> containing labelText. */
async function fillInputNearLabel(page: Page, labelText: string, value: string) {
  const ok = await page.evaluate(
    ({ labelText, value }) => {
      const lbl = [...document.querySelectorAll("label")].find((l) => (l.textContent ?? "").includes(labelText));
      const input = lbl?.parentElement?.querySelector("input");
      if (!input) return false;
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(input, value);
      input.dispatchEvent(new Event("input", { bubbles: true }));
      return true;
    },
    { labelText, value },
  );
  if (!ok) throw new Error(`No input found near label "${labelText}"`);
}

async function loginAndOpenBomScreen(page: Page) {
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
  await page.evaluate(() => {
    const input = document.querySelector<HTMLInputElement>('input[placeholder*="ค้นหาเมนู"]');
    if (!input) throw new Error("Menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "สูตรผลิต");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1000);
  await clickButtonByText(page, /^สูตรผลิต$/);
  await page.waitForTimeout(2500);
  await expect(page.locator("body")).toContainText(/รายการสูตรผลิต|สูตรผลิต/);
}

test("product BOM — create recipe with ingredient+cost+output qty, save, reopen, delete", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `RCPE2E${uid}`;
  const name = `สูตรE2E${uid}`;

  await loginAndOpenBomScreen(page);

  // NEW recipe (round + button in the list header)
  await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => b.offsetParent !== null && b.querySelector("svg.lucide-plus") && b.className.includes("rounded-full"),
    );
    if (!btn) throw new Error("Add-recipe button not found");
    btn.click();
  });
  await page.waitForTimeout(1500);

  await fillInputNearLabel(page, "รหัสสูตรผลิต", code);
  await fillInputNearLabel(page, "ผลิตได้ต่อสูตร", "5");
  // Thai name field — the shared <NamesEditor> renders each language input with the language
  // code as its placeholder (see names-editor.tsx), so `input[placeholder="th"]` is the stable
  // selector (an ancestor-text walk does NOT work here: the label wraps a flag icon, not "TH").
  await page.evaluate((value) => {
    const nameInput = [...document.querySelectorAll<HTMLInputElement>('input[placeholder="th"]')].find(
      (i) => i.offsetParent !== null,
    );
    if (!nameInput) throw new Error("Thai name input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(nameInput, value);
    nameInput.dispatchEvent(new Event("input", { bubbles: true }));
  }, name);

  // Add ingredient via the material picker (seeded material แป้งสาลีตราทดสอบ, single unit → no unit dialog)
  await clickButtonByText(page, /เพิ่มวัตถุดิบ/);
  await page.waitForTimeout(1500);
  await page.evaluate(() => {
    const input = [...document.querySelectorAll<HTMLInputElement>("input")].find(
      (i) => i.offsetParent !== null && /ค้นหา/.test(i.placeholder ?? ""),
    );
    if (!input) throw new Error("Picker search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "แป้งสาลี");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1500);
  await clickButtonByText(page, /แป้งสาลีตราทดสอบ/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, "แป้งสาลีตราทดสอบ")).toBe(true);

  // Editable cost (the fix this test guards): set cost/unit = 40 on the material row
  await page.evaluate(() => {
    const row = [...document.querySelectorAll("tbody tr")][0];
    if (!row) throw new Error("No ingredient row");
    const costInput = [...row.querySelectorAll<HTMLInputElement>('input[type="number"]')].find(
      (i) => i.placeholder === "0.00",
    );
    if (!costInput) throw new Error("Cost input not found (editable-cost fix regressed?)");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(costInput, "40");
    costInput.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(300);

  // SAVE
  await clickButtonByText(page, /บันทึกสูตร/);
  await expect.poll(() => bodyHasText(page, "บันทึกสำเร็จ"), { timeout: 15000 }).toBe(true);
  await clickButtonByText(page, /^ตกลง$/);
  await page.waitForTimeout(1500);

  // API + stored-state check: reopen from the list — recipe carries cost 40 and output qty 5
  await page.evaluate((n) => {
    const el = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => (b.textContent ?? "").includes(n) && b.offsetParent !== null,
    );
    if (!el) throw new Error("Saved recipe not in list");
    el.click();
  }, name);
  await page.waitForTimeout(2500);
  const reopened = await page.evaluate(() => {
    const outLbl = [...document.querySelectorAll("label")].find((l) => (l.textContent ?? "").includes("ผลิตได้ต่อสูตร"));
    const outVal = outLbl?.parentElement?.querySelector("input")?.value;
    const row = [...document.querySelectorAll("tbody tr")][0];
    const costVal = row
      ? [...row.querySelectorAll<HTMLInputElement>('input[type="number"]')].find((i) => i.placeholder === "0.00")?.value
      : undefined;
    return { outVal, costVal };
  });
  expect(reopened.outVal).toBe("5");
  expect(reopened.costVal).toBe("40");

  // DELETE (the other fix this test guards) — button did not exist before 2026-07-02
  await clickButtonByText(page, /^ลบสูตร$/);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, "ยืนยันการลบสูตรผลิต")).toBe(true);
  // Confirm dialog's own "ลบสูตร" button is the LAST matching button on screen
  await page.evaluate(() => {
    const btns = [...document.querySelectorAll<HTMLElement>("button")].filter(
      (b) => (b.textContent ?? "").trim() === "ลบสูตร" && b.offsetParent !== null,
    );
    if (!btns.length) throw new Error("Confirm delete button not found");
    btns[btns.length - 1].click();
  });
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, name)).toBe(false);

  // Cleanup safety net: remove any leftover E2E recipes via API if the UI delete ever regresses.
  await page.evaluate(async () => {
    const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
    if (!auth) return;
    const headers = {
      Authorization: `Bearer ${auth.token}`,
      "x-bc-backend-url": auth.backendUrl,
      "Content-Type": "application/json",
    };
    const list = await (
      await fetch("/api/system-settings/productbom/list?holdingcode=test&limit=100&offset=0", { headers })
    ).json();
    for (const row of list.data ?? []) {
      if (typeof row.barcode === "string" && row.barcode.startsWith("RCPE2E")) {
        await fetch(`/api/system-settings/productbom/${row.guidfixed}?holdingcode=test`, {
          method: "DELETE",
          headers,
          body: JSON.stringify({ guidfixed: row.guidfixed, holdingcode: "test", backendUrl: auth.backendUrl }),
        });
      }
    }
  });
});

/**
 * Live sub-recipe reference regression (2026-07-02): reftype=="recipe" components used to be a
 * save-time snapshot of the sub-recipe's ingredients; editing the sub-recipe never propagated to
 * parents unless they were resaved. The backend now resolves reftype=="recipe" items LIVE on every
 * GET (see internal/product/bom/services/bom_http_service.go resolveRecipeTree), and rejects cycles
 * at save time (checkNoCycle). This test drives the real UI: create a shared sub-recipe + two
 * parents referencing it, edit ONLY the sub-recipe, reopen both parents WITHOUT resaving them and
 * confirm both show the updated content — plus a cycle-attempt that must reject fast, not hang.
 *
 * Requires the same seeded materials as the test above (BMAT44941 / BMAT244941).
 */
test("product BOM — live sub-recipe cascade: edit shared sub-recipe, both parents update without resave; cycle rejected fast", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const subCode = `RCPE2ESUB${uid}`;
  const subName = `สูตรย่อยE2E${uid}`;
  const parentACode = `RCPE2EA${uid}`;
  const parentAName = `สูตรหลักAE2E${uid}`;
  const parentBCode = `RCPE2EB${uid}`;
  const parentBName = `สูตรหลักBE2E${uid}`;

  await loginAndOpenBomScreen(page);

  async function createNewRecipe(code: string, name: string, outputQty: string) {
    // The list's own "ค้นหาสูตร..." filter must be empty before creating: a leftover filter that
    // doesn't match the new blank-barcode virtual record hides it from visibleRecords, which makes
    // selectedRecord resolve to null even though selectedRecordId is set (see system-settings-screen.tsx
    // visibleRecords/selectedRecord memos) — the editor then silently shows its empty-state instead
    // of the new recipe form.
    await page.evaluate(() => {
      const input = document.querySelector<HTMLInputElement>('input[placeholder="ค้นหาสูตร..."]');
      if (!input) return;
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(input, "");
      input.dispatchEvent(new Event("input", { bubbles: true }));
    });
    await page.waitForTimeout(300);
    await page.evaluate(() => {
      const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
        (b) => b.offsetParent !== null && b.querySelector("svg.lucide-plus") && b.className.includes("rounded-full"),
      );
      if (!btn) throw new Error("Add-recipe button not found");
      btn.click();
    });
    await page.waitForTimeout(1500);
    // Verify we actually landed on a blank/new recipe form before filling it in.
    const gotNewForm = await page.evaluate(() => {
      const codeInput = document.querySelector<HTMLInputElement>('input[placeholder="RECIPE-001"]');
      return Boolean(codeInput && codeInput.offsetParent !== null && codeInput.value === "");
    });
    if (!gotNewForm) throw new Error("New-recipe form did not open (recipe code field not blank/visible)");
    await fillInputNearLabel(page, "รหัสสูตรผลิต", code);
    await fillInputNearLabel(page, "ผลิตได้ต่อสูตร", outputQty);
    await page.evaluate((value) => {
      const nameInput = [...document.querySelectorAll<HTMLInputElement>('input[placeholder="th"]')].find(
        (i) => i.offsetParent !== null,
      );
      if (!nameInput) throw new Error("Thai name input not found");
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(nameInput, value);
      nameInput.dispatchEvent(new Event("input", { bubbles: true }));
    }, name);
    // Hard guard: confirm the code field actually holds THIS recipe's code, not a stale/different
    // record still mid-transition from a just-completed save (previously caused edits meant for a
    // new recipe to land on the still-open previous record instead).
    const codeStuck = await page.evaluate(() => {
      const input = document.querySelector<HTMLInputElement>('input[placeholder="RECIPE-001"]');
      return input?.value ?? null;
    });
    if (codeStuck !== code) {
      throw new Error(`Recipe code field shows "${codeStuck}", expected "${code}" — new-recipe form did not take the input`);
    }
  }

  async function addSubRecipeIngredient(subRecipeName: string, qty: string) {
    const rowCountBefore = await page.evaluate(() => document.querySelectorAll("tbody tr").length);
    await clickButtonByText(page, /เพิ่มสูตรย่อย/);
    await page.waitForTimeout(800);
    // ROOT CAUSE of a real, previously-mysterious bug (found via manual repro): the "เลือกสูตรย่อย"
    // (sub-recipe picker) modal has NO role="dialog" and its item buttons re-use the SAME visible
    // text as the corresponding row in the recipe LIST on the left, which stays mounted underneath
    // the modal overlay. An unscoped "find the first button whose text matches" click picks the
    // LIST's row (earlier in DOM order) instead of the modal's button — silently SELECTING that
    // recipe in the list (calling the list's own onClick) rather than adding it as a sub-recipe.
    // That corrupts `selectedRecord`/`parentItemCode` right under the test's feet. Scope strictly to
    // the modal's own item buttons (distinguishing class from product-bom-editor.tsx's picker item:
    // "rounded-lg border border-border bg-card").
    await page.evaluate((name) => {
      const btn = [...document.querySelectorAll<HTMLElement>("button.rounded-lg.border.border-border.bg-card")].find(
        (b) => b.offsetParent !== null && (b.textContent ?? "").includes(name),
      );
      if (!btn) throw new Error(`Sub-recipe picker item "${name}" not found (modal-scoped)`);
      btn.click();
    }, subRecipeName);
    // Wait for the picker modal to close AND the new row to actually land in the table — a fixed
    // sleep here previously raced the click's async state has not applied to the table by the time
    // the next selector ran, so the "Save" button acted on stale (empty) bomItems.
    await expect
      .poll(() => page.evaluate(() => document.querySelectorAll("tbody tr").length), { timeout: 8000 })
      .toBeGreaterThan(rowCountBefore);
    await page.evaluate((qtyVal) => {
      const rows = [...document.querySelectorAll("tbody tr")];
      const row = rows[rows.length - 1];
      const qtyInput = row?.querySelector<HTMLInputElement>('td:nth-child(3) input[type="number"]');
      if (!qtyInput) throw new Error("Sub-recipe qty input not found");
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(qtyInput, qtyVal);
      qtyInput.dispatchEvent(new Event("input", { bubbles: true }));
    }, qty);
    await page.waitForTimeout(300);
  }

  async function addMaterialIngredient(searchTerm: RegExp, cost: string) {
    await clickButtonByText(page, /เพิ่มวัตถุดิบ/);
    await page.waitForTimeout(1200);
    // Scope to the picker's own dialog — the header's global menu search also matches a broad
    // "contains ค้นหา" selector and must not be touched here.
    await page.evaluate((term) => {
      const dialog = document.querySelector('[role="dialog"]');
      const input = dialog?.querySelector<HTMLInputElement>("input");
      if (!input) throw new Error("Picker search input not found (dialog scope)");
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(input, term);
      input.dispatchEvent(new Event("input", { bubbles: true }));
    }, searchTerm.source.replace(/\\/g, ""));
    await page.waitForTimeout(1200);
    await clickButtonByText(page, searchTerm);
    await page.waitForTimeout(1200);
    // Set cost on the LAST row (the one just added)
    await page.evaluate((costVal) => {
      const rows = [...document.querySelectorAll("tbody tr")];
      const row = rows[rows.length - 1];
      if (!row) throw new Error("No ingredient row after add");
      const costInput = [...row.querySelectorAll<HTMLInputElement>('input[type="number"]')].find(
        (i) => i.placeholder === "0.00",
      );
      if (!costInput) throw new Error("Cost input not found");
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(costInput, costVal);
      costInput.dispatchEvent(new Event("input", { bubbles: true }));
    }, cost);
    await page.waitForTimeout(300);
  }

  async function checkSavedViaApi(code: string): Promise<boolean> {
    return page.evaluate(async (c) => {
      const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
      if (!auth) return false;
      const headers = { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl };
      const list = await (
        await fetch("/api/system-settings/productbom/list?holdingcode=test&limit=200&offset=0", { headers })
      ).json();
      return (list.data ?? []).some((r: { barcode?: string }) => r.barcode === c);
    }, code);
  }

  // Returns the recipe-code input's current value, or null if not present/visible.
  async function currentCodeFieldValue(): Promise<string | null> {
    return page.evaluate(() => {
      const input = document.querySelector<HTMLInputElement>('input[placeholder="RECIPE-001"]');
      return input && input.offsetParent !== null ? input.value : null;
    });
  }

  async function saveRecipe(expectedCode: string) {
    // Cheap pre-flight guard: confirm the code field still shows what we typed right before
    // clicking Save. Fails loudly instead of silently saving over the wrong recipe if selection
    // ever drifts (see addSubRecipeIngredient's DOM-scoping fix above for the actual bug this once
    // exposed — kept here as a regression tripwire, it's a 3-line check).
    const codeNow = await currentCodeFieldValue();
    if (codeNow !== expectedCode) {
      throw new Error(`Recipe code field shows "${codeNow}" right before save, expected "${expectedCode}" — selection drifted`);
    }
    await clickButtonByText(page, /บันทึกสูตร/);
    await expect.poll(() => bodyHasText(page, "บันทึกสำเร็จ"), { timeout: 15000 }).toBe(true);
    await clickButtonByText(page, /^ตกลง$/);
    await page.waitForTimeout(1500);
    const saved = await checkSavedViaApi(expectedCode);
    if (!saved) throw new Error(`Recipe ${expectedCode} was not found via API after save`);
  }

  async function openRecipeByName(name: string) {
    await page.evaluate((n) => {
      const el = [...document.querySelectorAll<HTMLElement>("button")].find(
        (b) => (b.textContent ?? "").includes(n) && b.offsetParent !== null,
      );
      if (!el) throw new Error(`Saved recipe "${n}" not in list`);
      el.click();
    }, name);
    await page.waitForTimeout(2000);
  }

  // 1. Create the shared sub-recipe: 1 ingredient (แป้งสาลีตราทดสอบ), outputqty=1
  await createNewRecipe(subCode, subName, "1");
  await addMaterialIngredient(/แป้งสาลีตราทดสอบ/, "40");
  await saveRecipe(subCode);

  // 2. Create parent A referencing the sub-recipe (qty=0.05)
  await createNewRecipe(parentACode, parentAName, "4");
  await addSubRecipeIngredient(subName, "0.05");
  await saveRecipe(parentACode);

  // 3. Create parent B also referencing the sub-recipe (qty=0.08, independent of A)
  await createNewRecipe(parentBCode, parentBName, "2");
  await addSubRecipeIngredient(subName, "0.08");
  await saveRecipe(parentBCode);

  // 4. Edit ONLY the shared sub-recipe: bump its material cost 40 -> 55 (no touch of A or B)
  await openRecipeByName(subName);
  await page.evaluate(() => {
    const row = [...document.querySelectorAll("tbody tr")][0];
    if (!row) throw new Error("No ingredient row on sub-recipe");
    const costInput = [...row.querySelectorAll<HTMLInputElement>('input[type="number"]')].find(
      (i) => i.placeholder === "0.00",
    );
    if (!costInput) throw new Error("Cost input not found on sub-recipe");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(costInput, "55");
    costInput.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await saveRecipe(subCode);

  // 5. Reopen A and B WITHOUT resaving — both must show the updated ฿55 sub-recipe cost baked into
  // their resolved sub-recipe cost/unit cell (proves live resolution, not a stale save-time copy).
  await openRecipeByName(parentAName);
  const aSauceRow = await page.evaluate((subBarcode) => {
    const rows = [...document.querySelectorAll("tbody tr")];
    const row = rows.find((r) => (r.textContent ?? "").includes(subBarcode));
    return row ? row.textContent ?? "" : null;
  }, subCode);
  expect(aSauceRow).not.toBeNull();
  expect(aSauceRow).toContain("55.00");

  await openRecipeByName(parentBName);
  const bSauceRow = await page.evaluate((subBarcode) => {
    const rows = [...document.querySelectorAll("tbody tr")];
    const row = rows.find((r) => (r.textContent ?? "").includes(subBarcode));
    return row ? row.textContent ?? "" : null;
  }, subCode);
  expect(bSauceRow).not.toBeNull();
  expect(bSauceRow).toContain("55.00");

  // 6. Attempt a circular reference via the real API (PUT sub-recipe to reference parent A back) —
  // must reject fast with a clear error, never hang.
  const cycleResult = await page.evaluate(async ({ subBarcode, parentABarcode, subName2 }) => {
    const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
    if (!auth) return { error: "no auth" };
    const headers = {
      Authorization: `Bearer ${auth.token}`,
      "x-bc-backend-url": auth.backendUrl,
      "Content-Type": "application/json",
    };
    const listRes = await (
      await fetch("/api/system-settings/productbom/list?holdingcode=test&limit=200&offset=0", { headers })
    ).json();
    const subRow = (listRes.data ?? []).find((r: { barcode?: string }) => r.barcode === subBarcode);
    if (!subRow) return { error: "sub-recipe not found in list" };

    const start = Date.now();
    const res = await fetch(`/api/system-settings/productbom/${subRow.guidfixed}?holdingcode=test`, {
      method: "PUT",
      headers,
      body: JSON.stringify({
        guidfixed: subRow.guidfixed,
        holdingcode: "test",
        backendUrl: auth.backendUrl,
        barcode: subBarcode,
        names: [{ code: "th", name: subName2 }],
        itemunitcode: "RECIPE",
        price: 0,
        outputqty: 1,
        bom: [{ barcode: parentABarcode, reftype: "recipe", qty: 0.01, yieldpercent: 100, averagecost: 0 }],
      }),
    });
    const elapsedMs = Date.now() - start;
    const json = await res.json();
    return { status: res.status, elapsedMs, json };
  }, { subBarcode: subCode, parentABarcode: parentACode, subName2: subName });

  expect(cycleResult.error).toBeUndefined();
  expect(cycleResult.status).toBe(400);
  expect(cycleResult.elapsedMs).toBeLessThan(5000);
  expect(String(cycleResult.json?.message ?? "")).toMatch(/circular/i);

  // Cleanup: soft-delete the 3 E2E recipes this test created via API (mirrors the pattern above).
  // Does NOT touch the persistent Thai demo dataset (RCPSAUCE197609 / RCPPADTHAI197609 / etc).
  await page.evaluate(async (codes) => {
    const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
    if (!auth) return;
    const headers = {
      Authorization: `Bearer ${auth.token}`,
      "x-bc-backend-url": auth.backendUrl,
      "Content-Type": "application/json",
    };
    const list = await (
      await fetch("/api/system-settings/productbom/list?holdingcode=test&limit=200&offset=0", { headers })
    ).json();
    for (const row of list.data ?? []) {
      if (typeof row.barcode === "string" && codes.includes(row.barcode)) {
        await fetch(`/api/system-settings/productbom/${row.guidfixed}?holdingcode=test`, {
          method: "DELETE",
          headers,
          body: JSON.stringify({ guidfixed: row.guidfixed, holdingcode: "test", backendUrl: auth.backendUrl }),
        });
      }
    }
  }, [subCode, parentACode, parentBCode]);
});

/**
 * Costing fields regression (added for the labor/overhead/scrap/costmode/standardcost + sale-qty
 * calculator feature): covers persistence of the 5 new root-level costing fields and the CostMode-
 * aware total cost (current-mode rollup formula, then standard-mode frozen override), plus the
 * sale-quantity calculator inside the Exploded BOM modal. Creates its own recipe (does not touch the
 * persistent Thai demo dataset) and deletes it via the UI at the end.
 */
test("product BOM — labor/overhead/scrap/costmode persist and drive the CostMode-aware total; sale-quantity calculator scales correctly", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const code = `RCPCOST${uid}`;
  const name = `สูตรต้นทุนE2E${uid}`;

  await loginAndOpenBomScreen(page);

  // NEW recipe
  await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => b.offsetParent !== null && b.querySelector("svg.lucide-plus") && b.className.includes("rounded-full"),
    );
    if (!btn) throw new Error("Add-recipe button not found");
    btn.click();
  });
  await page.waitForTimeout(1500);

  await fillInputNearLabel(page, "รหัสสูตรผลิต", code);
  await fillInputNearLabel(page, "ผลิตได้ต่อสูตร", "4"); // outputqty=4
  await page.evaluate((value) => {
    const nameInput = [...document.querySelectorAll<HTMLInputElement>('input[placeholder="th"]')].find(
      (i) => i.offsetParent !== null,
    );
    if (!nameInput) throw new Error("Thai name input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(nameInput, value);
    nameInput.dispatchEvent(new Event("input", { bubbles: true }));
  }, name);

  // One material ingredient: qty=2, cost=25 -> rollup = 50
  await clickButtonByText(page, /เพิ่มวัตถุดิบ/);
  await page.waitForTimeout(1500);
  await page.evaluate(() => {
    const input = [...document.querySelectorAll<HTMLInputElement>("input")].find(
      (i) => i.offsetParent !== null && /ค้นหา/.test(i.placeholder ?? ""),
    );
    if (!input) throw new Error("Picker search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "แป้งสาลี");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1500);
  await clickButtonByText(page, /แป้งสาลีตราทดสอบ/);
  await page.waitForTimeout(2000);

  await page.evaluate(() => {
    const row = [...document.querySelectorAll("tbody tr")][0];
    if (!row) throw new Error("No ingredient row");
    const qtyInput = row.querySelector<HTMLInputElement>('td:nth-child(3) input[type="number"]');
    const costInput = [...row.querySelectorAll<HTMLInputElement>('input[type="number"]')].find(
      (i) => i.placeholder === "0.00",
    );
    if (!qtyInput || !costInput) throw new Error("Qty/cost input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(qtyInput, "2");
    qtyInput.dispatchEvent(new Event("input", { bubbles: true }));
    setter.call(costInput, "25");
    costInput.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(300);

  // Labor=10, Overhead=6, Scrap=20% -> current-mode cost/unit = (50+10+6)/4/(1-0.2) = 20.625
  await fillInputNearLabel(page, "ค่าแรงงาน", "10");
  await fillInputNearLabel(page, "ค่าโสหุ้ย", "6");
  await fillInputNearLabel(page, "% ของเสีย", "20");
  await page.waitForTimeout(300);

  await clickButtonByText(page, /บันทึกสูตร/);
  await expect.poll(() => bodyHasText(page, "บันทึกสำเร็จ"), { timeout: 15000 }).toBe(true);
  await clickButtonByText(page, /^ตกลง$/);
  await page.waitForTimeout(1500);

  // Reopen from the list — three-way check via UI + API.
  await page.evaluate((n) => {
    const el = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => (b.textContent ?? "").includes(n) && b.offsetParent !== null,
    );
    if (!el) throw new Error("Saved recipe not in list");
    el.click();
  }, name);
  await page.waitForTimeout(2500);

  const reloaded = await page.evaluate(() => {
    const getVal = (labelText: string) => {
      const lbl = [...document.querySelectorAll("label")].find((l) => (l.textContent ?? "").includes(labelText));
      return lbl?.parentElement?.querySelector("input")?.value;
    };
    return { labor: getVal("ค่าแรงงาน"), overhead: getVal("ค่าโสหุ้ย"), scrap: getVal("% ของเสีย") };
  });
  expect(reloaded).toEqual({ labor: "10", overhead: "6", scrap: "20" });

  const apiData = await page.evaluate(async (c) => {
    const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
    const headers = { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl };
    const list = await (
      await fetch("/api/system-settings/productbom/list?holdingcode=test&limit=200&offset=0", { headers })
    ).json();
    const row = (list.data ?? []).find((r: { barcode?: string }) => r.barcode === c);
    if (!row) return null;
    const info = await (
      await fetch(`/api/system-settings/productbom/${row.guidfixed}?holdingcode=test`, { headers })
    ).json();
    return info.data;
  }, code);
  expect(apiData?.laborcost).toBe(10);
  expect(apiData?.overheadcost).toBe(6);
  expect(apiData?.scrappercent).toBe(20);
  expect(apiData?.costmode).toBe("current");

  // Open Exploded BOM modal: current-mode cost/unit must be (50+10+6)/4/0.8 = 20.625 -> ฿20.63.
  await clickButtonByText(page, /โครงสร้างสูตรละเอียด/);
  await page.waitForTimeout(1000);
  const costPerUnitText = await page.evaluate(
    () => document.body.innerText.match(/ต้นทุน\/หน่วยผลิต[\s\S]{0,60}/)?.[0] ?? "",
  );
  expect(costPerUnitText).toContain("20.63");

  // Sale-quantity calculator: enter qty=8 (= 2x the good-unit gross batch of outputqty/(1-scrap%)=5)
  // -> total should be 8 * 20.625 = 165.00.
  await fillInputNearLabel(page, "จำนวนที่ต้องการขาย", "8");
  await page.waitForTimeout(500);
  const totalForQtyText = await page.evaluate(
    () => document.body.innerText.match(/ต้นทุนรวมสำหรับจำนวนนี้[\s\S]{0,30}/)?.[0] ?? "",
  );
  expect(totalForQtyText).toContain("165.00");

  // Close modal, switch to Standard Cost mode with value 42, save, reopen, confirm the headline
  // now shows the frozen 42.00 instead of the live-computed 20.63.
  await page.evaluate(() => {
    const closeButtons = [...document.querySelectorAll<HTMLElement>("button")].filter(
      (b) => b.offsetParent !== null && b.querySelector("svg.lucide-x"),
    );
    closeButtons[closeButtons.length - 1]?.click();
  });
  await page.waitForTimeout(500);
  await page.evaluate(() => {
    const radios = [...document.querySelectorAll<HTMLInputElement>('input[name="costmode"]')];
    radios[1]?.click();
  });
  await page.waitForTimeout(300);
  await page.evaluate(() => {
    const radios = [...document.querySelectorAll<HTMLInputElement>('input[name="costmode"]')];
    const container = radios[1]?.closest(".flex.flex-wrap");
    const input = container?.querySelector<HTMLInputElement>('input[type="number"]');
    if (!input) throw new Error("Standard cost input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "42");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(300);
  await clickButtonByText(page, /บันทึกสูตร/);
  await expect.poll(() => bodyHasText(page, "บันทึกสำเร็จ"), { timeout: 15000 }).toBe(true);
  await clickButtonByText(page, /^ตกลง$/);
  await page.waitForTimeout(1500);

  await page.evaluate((n) => {
    const el = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => (b.textContent ?? "").includes(n) && b.offsetParent !== null,
    );
    if (!el) throw new Error("Saved recipe not in list");
    el.click();
  }, name);
  await page.waitForTimeout(2000);
  await clickButtonByText(page, /โครงสร้างสูตรละเอียด/);
  await page.waitForTimeout(1000);
  const standardModeText = await page.evaluate(
    () => document.body.innerText.match(/ต้นทุนมาตรฐาน\/หน่วย[\s\S]{0,40}/)?.[0] ?? "",
  );
  expect(standardModeText).toContain("42.00");
  expect(await bodyHasText(page, "ราคาปัจจุบัน")).toBe(true); // ingredient list still shown as reference

  // Cleanup: delete this test recipe via API (does not touch the persistent Thai demo dataset).
  await page.evaluate(async (c) => {
    const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
    if (!auth) return;
    const headers = {
      Authorization: `Bearer ${auth.token}`,
      "x-bc-backend-url": auth.backendUrl,
      "Content-Type": "application/json",
    };
    const list = await (
      await fetch("/api/system-settings/productbom/list?holdingcode=test&limit=200&offset=0", { headers })
    ).json();
    for (const row of list.data ?? []) {
      if (row.barcode === c) {
        await fetch(`/api/system-settings/productbom/${row.guidfixed}?holdingcode=test`, {
          method: "DELETE",
          headers,
          body: JSON.stringify({ guidfixed: row.guidfixed, holdingcode: "test", backendUrl: auth.backendUrl }),
        });
      }
    }
  }, code);
});
