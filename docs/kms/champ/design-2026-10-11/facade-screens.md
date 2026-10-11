# ออกแบบ: จอ UI-only (facade) → "รอพัฒนา" จนกว่าจะสร้างจริง

- วันที่: 2026-10-11 · ขอบเขต audit: `ORD-03`, `CRIT-U2`, cluster F (`_notes_3/4/7/8`), `D1b` (transaction-lock), `A4-1`, `A4-2`
- คำตัดสินลุงจืด 2026-10-11: จอที่เป็น UI อย่างเดียว (ไม่มี backend จริง) ให้เป็น **"รอพัฒนา"** จนกว่าจะสร้างจริง; ERP cancel ระดับหัวเอกสาร; รายงานทุกชื่อตาม Champ แต่เปิด **จอรายงานกลางจอเดียว** ผ่าน report registry ฝั่ง backend
- เอกสารนี้เป็นแบบเท่านั้น ยังไม่แก้โค้ด (workflow อื่นกำลังแก้ `menu-screen-status.ts`, `backend/internal/transaction/**`, `erp-*`) → ลงมือหลัง workflow นั้น commit แล้ว

## 1. กลไกเดิมที่ใช้ (ไม่สร้างกลไกใหม่)

| ส่วน | ไฟล์:บรรทัด | พฤติกรรม |
|---|---|---|
| รายการ route ที่มีจอ custom | `frontend/src/lib/menu-screen-status.ts:12-46` `CUSTOM_MENU_SCREEN_ROUTES` | route ในชุดนี้ = "มีจอแล้ว" |
| ตัดสินว่ารอพัฒนา | `menu-screen-status.ts:100-114` `isMenuScreenPending(route)` | true เมื่อไม่อยู่ใน CUSTOM และไม่ match registry ใด (GL, ทรัพย์สิน, ERP transaction, thai tax, tax form, report, tools, operations, system setting) |
| จอ fallback | `frontend/src/app/menu/main-menu-screen.tsx:2242-2244` `plannedCard` (= `MenuPlannedCard`), `return plannedCard;` ท้าย `WorkTabPanel` | route ที่ไม่มี branch → การ์ด "เมนูในแผนพัฒนา" |
| ป้าย | `frontend/src/app/menu/menu-pending-badge.tsx` | **คืน `null` ตั้งแต่ commit `6c4447f8` (2026-10-08 Champ parity)** — เดิมแสดง `menu_pending_development` "รอพัฒนา" / `menu_data_pending` "รอเชื่อมข้อมูล" |
| การ์ด | `frontend/src/app/menu/menu-planned-card.tsx:33` | เรียก `<MenuPendingBadge/>` → ปัจจุบันไม่แสดงอะไร |
| test คู่ขนาน | `frontend/src/lib/menu-screen-status.test.ts:9-13` | regex `activeTab.route === "..."` ใน `main-menu-screen.tsx` ต้องเท่ากับ `CUSTOM_MENU_SCREEN_ROUTES` → **แก้ 2 ไฟล์พร้อมกันเสมอ** |

ตรวจแล้ว: route facade ทั้งหมดด้านล่าง **ไม่ match registry อื่น** (grep ใน `frontend/src/lib/*.ts` เจอแค่ `menu-icons.ts` ซึ่งเป็นแผนที่ไอคอน) และ `getSystemSettingConfig` (`frontend/src/lib/system-setting-screens.ts:1362`) match ด้วย route/slug ตรงตัว ไม่มี `/settings/currency` ฯลฯ → ถอดจาก CUSTOM แล้ว `isMenuScreenPending` = true ทันที, `/transaction/receivedeposit` match `isErpTransactionRoute` (config `frontend/src/lib/erp-transaction.ts:522-524`, allowlist BFF `erp-transaction-modules.ts:28`) → ไม่ pending

## 2. รายการ facade (ตรงตัว)

เกณฑ์ facade = จอไม่มีการเรียก API จริง (ข้อมูล hard-code / state ใน memory / localStorage) แต่แสดงตัวเหมือนใช้งานได้ · ทุกไฟล์ถูก import ที่ `main-menu-screen.tsx` ที่เดียว (+ test ของตัวเอง)

| # | menu id (`menu-data.ts`) | route | component (`frontend/src/components/`) | บรรทัด | ทำไมเป็น facade | backend จริง? | Champ (หลักฐาน) | การดำเนินการ |
|---|---|---|---|---|---|---|---|---|
| 1 | `receive-deposit` :268 | `/transaction/receivedeposit` | `deposit-receipt-screen.tsx` | 875 | state ใน memory, ไทย hard-code 104 บรรทัด, ไม่ fetch | **มี**: doctype `receivedeposit` flag 32 `RDEP` (`backend/internal/doctype/doctype.go:78`) + CRUD generic `/transaction/:module` + config ERP | 5327 บันทึกใบรับเงินมัดจำ · `bcspec/หน้าจอเมนูจัดขาย/บันทึกใบรับเงินมัดจำ.jpg` | **ถอดเงา → ErpCrudWorkbench** (ไม่ pending) |
| 2 | `billing-note-auto` :363 | `/ar/billing-note-auto` | `billing-note-auto-screen.tsx` | 686 | สร้างใบวางบิลจากข้อมูลจำลอง | บางส่วน: doctype `billingnote` 26 (`doctype.go:76`), `/transaction/billingnote/auto-print` (พิมพ์) แต่**ไม่มี auto-generate** | `bcspec/หน้าจอเมนูลูกหนี้/บันทึกใบวางบิลอัตโนมัติ.jpg` | pending |
| 3 | `transaction-lock` :577 | `/tools/transaction-lock` | `transaction-lock-screen.tsx` | 794 | ล็อกใน state ไม่บันทึก, ไม่มีผลกับการบันทึกเอกสาร | ไม่มี | `bcspec/หน้าจอเมนูสินค้าคงคลัง/Transaction Lock.jpg`; `ICDlgLockCostDlg.cpp:132-270` (islockcost) | pending (D1b ทาง B) → สร้างจริงตาม D1b ทาง A |
| 4 | `ap-payment-prepare` :319 | `/ap/payment-prepare` | `ap-payment-prepare-screen.tsx` | 808 | ข้อมูลจำลอง | ไม่มี | `bcspec/หน้าจอเมนูเจ้าหนี้/บันทึกเตรียมจ่ายจ่ายชำระหนี้.jpg` | pending → C1 |
| 5 | `cheque-movement` :463 | `/banking/cheque-movement` | `cheque-movement-screen.tsx` | 695 | ข้อมูลจำลอง | ไม่มี (ยังไม่มีทะเบียนเช็ค B0) | `bcspec/หน้าจอเมนูเช็ค บัตร/ข้อมูลการเคลื่อนไหวเช็ค.jpg` | pending → B0, B2 |
| 6 | `bank-service` :443 | `/banking/bank-service` | `bank-service-screen.tsx` | 598 | ข้อมูลจำลอง | ไม่มี | `bcspec/หน้าจอเมนูธนาคาร เงินสด/บันทึกรายละเอียดธนาคาร.jpg` (tab บริการธนาคาร) | pending → B4 |
| 7 | `cheque-issued-prepare` :415 | `/banking/cheque-issued-prepare` | `cheque-issued-prepare-screen.tsx` | 983 | ข้อมูลจำลอง | ไม่มี | `bcspec/หน้าจอเมนูเช็ค บัตร/บันทึกเช็คเตรียมจ่าย.jpg` | pending → B3 |
| 8 | `bank-contact` :444 | `/banking/contacts` | `banking-contacts-screen.tsx` | 1357 | localStorage, ไทย hard-code 157 บรรทัด | ไม่มี | 5680 บันทึกรายละเอียดผู้ติดต่อธนาคาร (`menuconfig.xml:365-391`) | pending |
| 9 | `petty-cash-limit` :436 | `/banking/petty-cash-limit` | `petty-cash-limit-screen.tsx` | 864 | ข้อมูลจำลอง | ไม่มี | 5177 กำหนดวงเงินสดย่อย | pending |
| 10 | `sale-reservation` :255 | `/sale/reservation` | `sale-reservation-screen.tsx` | 656 | ข้อมูลจำลอง | บางส่วน: `saleorder` SO (`doctype.go:70`) แต่ไม่มี `sostatus` | 5326 บันทึกใบสั่งขาย/สั่งจองสินค้า, 5001 Flow · `ใบสั่งจองสินค้า.jpg` | pending → A4-1 |
| 11 | `sale-dispatch` :259 | `/sale/dispatch` | `sale-dispatch-screen.tsx` | 627 | ข้อมูลจำลอง | ไม่มี (DO 434 ยังไม่มี) | `bcspec/หน้าจอเมนูจัดขาย/ใบสั่งจ่ายสินค้า.jpg` | pending → A4-2 |
| 12 | `stock-count-handheld` :531 | `/inventory/stock-count-handheld` | `stock-count-handheld-screen.tsx` | 562 | ข้อมูลจำลอง | ไม่มี (มีแค่ doctype stock-count 433 `doctype.go:66`) | `Stock to count by Hand held.jpg` | pending |
| 13 | `product-supplier-update` :505 | `/inventory/product-supplier-update` | `product-supplier-update-screen.tsx` | 487 | ข้อมูลจำลอง | ไม่มี (ลุงจืด: product master ทำทีหลัง) | `ปรับปรุง ข้อมูล Supplier ตามสินค้า.jpg` | pending |
| 14-23 | `system-journal-format` :714, `system-work-date` :715, `system-product-by-warehouse` :716, `system-default-warehouse` :717, `system-currency` :718, `system-main-map` :719, `system-credit-debit-reasons` :720, `system-payment-terms` :721, `system-customer-inquiry` :722, `system-product-inquiry` :723 | `/settings/journal-format`, `/settings/work-date`, `/settings/product-by-warehouse`, `/settings/default-warehouse`, `/settings/currency`, `/settings/main-map`, `/settings/credit-debit-reasons`, `/settings/payment-terms`, `/settings/customer-inquiry`, `/settings/product-inquiry` | `system-*-screen.tsx` (10 ไฟล์) | 299–784 | ฟอร์มไม่บันทึกที่ใด | ไม่มี | `bcspec/หน้าจอเมนูจัดการระบบ/*.jpg` | pending |
| 24 | `help-about` :756 | `/settings/about` | `system-about-screen.tsx` | 222 | release/เวอร์ชัน hard-code | มีแค่ `VersionNumber = "1.1.1121"` (`backend/internal/goapi/bootstrap.go:34`), `/api/health` (:323) | เมนูช่วยเหลือ Champ | pending |
| 25 | `help-standard-software` :757 | `/settings/standard-software` | `system-standard-software-screen.tsx` | 204 | **อ้างการรับรองซอฟต์แวร์มาตรฐานสรรพากรโดยไม่มีเลขทะเบียน** | ไม่มี | — | pending (เสี่ยงข้อกฎหมาย ดู §7) |
| 26 | `etax-invoice` :283 | `/transaction/etax` | `etax-invoice-screen.tsx` | 637 | คิวเอกสาร `MOCK_DOCUMENTS` (:44) พร้อมสถานะส่งสรรพากรปลอม | บางส่วน: `POST /api/report/tax/etax/generate-xml` (`bootstrap.go:371`, test `etax_handler_test.go`) จริง แต่ไม่มีคิวจากใบกำกับจริง | ไม่มีใน Champ (ข้อยกเว้น ADR `2026-09-19-champ-parity-menu-cut.md:65`) | pending (คง backend XML) — ดูคำตัดสิน D-2 |
| 27-31 | `window-font-report`, `-title`, `-column`, `-detail`, `-company` :738-742 | `/settings/font-report`, `/settings/font-report-title`, `/settings/font-report-column`, `/settings/font-report-detail`, `/settings/font-report-company` | `screen-font-settings.tsx` (ใช้ร่วมกับ `/settings/font`) | — | เก็บ `bc_app_font_report*` ใน localStorage **แต่ไม่มีโค้ดใดอ่านไปใช้**; ปุ่ม "ค่าเริ่มต้นบริษัท" (`:195-203`, ปุ่ม `:276-283`) เขียน localStorage อย่างเดียว | ไม่มี | `bcspec/หน้าเมนูหน้าต่าง/เลือกตัวอักษรในการแสดงผล.jpg` | pending; **คง `/settings/font`** (ฟอนต์จอทำงานจริงผ่าน cookie `:155-165`, ธีม `:180`) |

ไม่ใช่ facade (คงไว้): `/shortcuts`, `/product`, `/productbarcode`, `/productbarcodeshelf`, `/pricehistory`, `/datamodelgraph`, `/inventory/product-sets`, `/productset`, `/report/vatsummary`, `/customer/label-print`, `/ar/billing-note-auto-print`, `/settings/font`

## 3. การแก้โค้ด (ตรงตัว)

### 3.1 `frontend/src/lib/menu-screen-status.ts:12-46`
`CUSTOM_MENU_SCREEN_ROUTES` เหลือ 12 รายการ:
```ts
"/shortcuts", "/product", "/productbarcode", "/productbarcodeshelf", "/pricehistory",
"/datamodelgraph", "/inventory/product-sets", "/productset", "/report/vatsummary",
"/customer/label-print", "/ar/billing-note-auto-print", "/settings/font",
```
ลบ 31 route: ทั้งหมดในตาราง §2 (รวม `/transaction/receivedeposit` เพราะ ERP registry รับต่อ)

### 3.2 `frontend/src/app/menu/main-menu-screen.tsx`
- ลบ import บรรทัด 137-162 (คง :136 `ScreenFontSettings`)
- ลบ branch ใน `WorkTabPanel`: receivedeposit :2321-2323 (**ต้องลบเพราะอยู่ก่อน `isErpTransactionRoute` :2325** — จอเงากำลังบัง ERP จริง), :2357-2455 ทุก branch facade (contacts, petty, cheque-issued-prepare, billing-note-auto, transaction-lock, ap/payment-prepare, cheque-movement, bank-service, sale/reservation, sale/dispatch, stock-count-handheld, product-supplier-update, settings ×10, about, standard-software, etax)
- branch ฟอนต์ :2346-2355 เหลือเฉพาะ `activeTab.route === "/settings/font"` (regex test ต้องเจอ literal นี้)
- ไม่เพิ่ม branch ใหม่: route ที่ถูกถอดจะตกไป `return plannedCard;` เอง
- ไฟล์ ≥950 บรรทัด → รัน `pwsh -NoProfile -File tools/gen-code-map.ps1` อัปเดต `docs/reference/CODE-MAP.md` (:408 main-menu-screen, :926 banking-contacts-screen ต้องหาย)

### 3.3 `frontend/src/app/menu/menu-planned-card.tsx:33`
แทน `<MenuPendingBadge .../>` ด้วยชิปในการ์ดโดยตรง (ข้อความจาก key เดิม `menu_pending_development`, `languages.tsv:4656`) ใช้ `menuText` แบบเดียวกับหัวการ์ด และ class เดิมของป้ายก่อน `6c4447f8`: `inline-flex shrink-0 whitespace-nowrap rounded-md border border-border bg-muted px-1.5 text-[0.9rem] leading-normal text-foreground` + `data-menu-pending={route}` (ให้ test/UAT จับได้). ผลที่ผู้ใช้เห็น: คลิกเมนู → แท็บเปิดการ์ด "เมนูในแผนพัฒนา" + ชื่อเมนูตาม Champ + ชิป "รอพัฒนา" + `menu_planned_description` — ไม่มีฟอร์มให้กรอก
- `MenuPendingBadge` ในเมนู (8 จุด: `champ-menu-bar.tsx:161`, `dashboard-home.tsx:150`, `manage-shortcuts-screen.tsx:306/450/556`, `main-menu-screen.tsx:1270/1424/1947`) คงคืน `null` ตาม Champ parity — ดูคำตัดสิน D-1

### 3.4 `frontend/src/components/screen-font-settings.tsx`
ตัด `FONT_TARGET_CONFIGS` (:38-97) เหลือ target `screen`, ลบ tab เลือก target (:229), ลบ `handleSaveCompanyDefault` (:195-203) + ปุ่ม (:276-283); คง `handleSelectFont` (cookie) และ `handleSetTheme`. key ภาษาที่กำพร้าหลังตัด → ลบแถวใน `languages.tsv` ใน commit เดียวกัน (ตรวจด้วย grep key ทั้ง `frontend/src` + `backend`)

### 3.5 ลบไฟล์ (กฎ dev: ลบของเก่าใน change เดียวกัน, ไม่มี flag)
26 component + 26 test (`frontend/src/components/`):
`ap-payment-prepare-screen`, `bank-service-screen`, `banking-contacts-screen`, `billing-note-auto-screen`, `cheque-issued-prepare-screen`, `cheque-movement-screen`, `deposit-receipt-screen`, `etax-invoice-screen`, `petty-cash-limit-screen`, `product-supplier-update-screen`, `sale-dispatch-screen`, `sale-reservation-screen`, `stock-count-handheld-screen`, `system-journal-format-screen`, `system-work-date-screen`, `system-product-by-warehouse-screen`, `system-default-warehouse-screen`, `system-currency-screen`, `system-payment-terms-screen`, `system-credit-debit-reasons-screen`, `system-main-map-screen`, `system-customer-inquiry-screen`, `system-product-inquiry-screen`, `system-about-screen`, `system-standard-software-screen`, `transaction-lock-screen` (`.tsx` + `.test.ts`) — รวม ≈16,000 บรรทัด

Import ที่ตรวจแล้ว: ทุกไฟล์ถูก import จาก `main-menu-screen.tsx` เท่านั้น. helper ที่ใช้ร่วมและ**ต้องคงไว้**: `@/lib/thai-banks` (ใช้ใน `system-settings-screen`), `thai-baht-text` (ใช้ใน `thai-document-print-modal`), `@/lib/client-auth-session` `apiFetch`. Backend `ETaxGenerateXMLHandler` คงไว้ (มี test, ADR ข้อยกเว้น)
- key ภาษา: facade ส่วนใหญ่ hard-code ไทย (ผิดกฎภาษาอยู่แล้ว); สแกนแล้วมี key จาก tsv 11 ตัว, ผู้สมัครกำพร้า ≤4 (กลุ่ม `pending_approval*`, `petty_cash*`, `credit_note*`, `out_of*`) → หลังลบให้รันสแกนซ้ำ (grep ทั้ง frontend+backend) แล้วลบเฉพาะที่กำพร้าจริง; **key เมนู (`bank_contact`, `petty_cash_limit`, `sale_reservation`, `sale_dispatch_order`, `ap_payment_prepare`, `billing_note_auto`, `cheque_movement`, `bank_service`, `product_supplier_adjust`, `stock_count_handheld`, `transaction_lock`, `system_*`, `help_about`, `help_standard_software`, `etax_invoice`) ห้ามลบ** (เมนู id/route/key คงเดิม)
- `frontend/src/lib/menu-data.ts:704,728,747` คอมเมนต์ "/pending/*" ล้าสมัย → แก้ให้ตรง (route จริงไม่เปลี่ยน)

### 3.6 Backend / DDL / REST
งานนี้ **ไม่มี DDL และไม่มี endpoint ใหม่** (เป็นการถอด UI ปลอม). receivedeposit ใช้ของเดิม: `GET/POST/PUT/DELETE /transaction/receivedeposit` (generic `/transaction/:module`), approve/cancel ระดับหัวเอกสาร (คำตัดสิน 2026-10-11). ข้อจำกัดที่ต้องรายงาน: flag 32 **ไม่อยู่ใน `postingFlags`** (`doctype.go:234-236`) → บันทึกมัดจำแล้วไม่ลง GL และยังไม่มีการตัดมัดจำในใบขาย — งานต่อเนื่องอยู่ในแบบ ERP posting (ไม่ใช่ไฟล์นี้)

## 4. Test ที่ต้องแก้

| ไฟล์ | การแก้ |
|---|---|
| `frontend/src/lib/audit-pending.test.ts:51` | เปลี่ยน `expect(pendingList.length).toBe(0)` เป็น `expect(pendingList.sort()).toEqual(FACADE_PENDING_ROUTES)` (30 route เรียงตัวอักษร = §2 ลบ receivedeposit); คอมเมนต์ประวัติ :38-50 สรุปเหลือบรรทัดเดียวอ้างเอกสารนี้ |
| `frontend/src/lib/menu-screen-status.test.ts` | `:16` `CHAMP_PENDING_ROUTES` ใส่ 30 route; `:36-38` expected list ใช้ชุดนี้; `:40` เปลี่ยนเป็น `isMenuScreenPending("/transaction/etax")` → `true`; เพิ่ม `isMenuScreenPending("/transaction/receivedeposit")` → `false` และ `getErpTransactionConfig("/transaction/receivedeposit")` มีค่า; `:9-13` regex parity ผ่านเองเมื่อ §3.1+§3.2 แก้คู่กัน; `:33` นับเมนู 238 ไม่เปลี่ยน |
| `frontend/src/lib/menu-bars.test.ts:169-181` | `window-font-report*` ×5 และ `help-about`, `help-standard-software` → expect pending; `window-font-screen` (`/settings/font`) → ไม่ pending; `:23` `CHAMP_ONLY_PENDING_IDS` คงว่าง (convention `/pending/<id>` ใช้กับเมนูที่ไม่มี route จริงเท่านั้น); `:271-273` ไม่เปลี่ยน |
| `frontend/src/components/screen-font-settings.test.ts` | เหลือ assert `/settings/font` ไม่ pending + ไม่มี tab target/ปุ่มค่าเริ่มต้นบริษัท |
| `frontend/src/app/menu/menu-planned-card.test.tsx` (ใหม่, เล็ก) | render ด้วย `react-dom/server` → มี `data-menu-pending` และข้อความ "รอพัฒนา" |
| `frontend/src/lib/erp-transaction.test.ts:91-102` | ไม่เปลี่ยน (etax/petty/contacts ไม่มี ERP config อยู่แล้ว) |
| 26 test ของ facade | ลบพร้อม component |
| `tests/uat-phase5.spec.ts` | เขียนใหม่: (ก) วนเมนู facade ทั้ง 30 → เห็น `[data-menu-pending]` + "รอพัฒนา" ไม่มี input; (ข) receivedeposit ผ่าน ErpCrudWorkbench แบบ CRUD ทีละ step + ตรวจ PostgreSQL |

UAT CRUD receivedeposit (Demo → `rungrueng` → บริษัท `01` → สาขา `00000`, ตรวจ psql ใน container `postgres` ฐานของ holding นั้น ทันทีหลังแต่ละขั้น):
1. Create เลขที่ `RDEP-UAT-<seed>` → `SELECT id,docno,transflag,custcode,iscancel FROM doc WHERE docno=$1;` ได้ 1 แถว `transflag=32`; ยอดใน `docdetail` เป็น numeric ตรงที่กรอก
2. Update หมายเหตุ/ยอด → แถว `id` เดิมเปลี่ยน ไม่มีแถวใหม่
3. Cancel (หัวเอกสาร) → `iscancel=true` + มีแถวใน `erp_doc_events`
4. Delete (ร่าง) → แถวหาย, record อื่นของ holding นับเท่าเดิม
5. Edge: เลขซ้ำ (error ไทย), ช่องว่าง, double-click บันทึก (ได้ 1 แถว)
เคลียร์ด้วย `docno` เป้าหมายเท่านั้น

Verify ก่อนบอกเสร็จ: `npm test -- src/lib/audit-pending.test.ts src/lib/menu-screen-status.test.ts src/lib/menu-bars.test.ts src/components/screen-font-settings.test.ts src/app/menu/menu-planned-card.test.tsx` + `npx tsc --noEmit` + `npm run verify` + Browser pane: Demo เปิด 3 เมนูตัวอย่าง (`/settings/currency`, `/banking/contacts`, `/transaction/etax`) light/dark ไม่มี console error

## 5. CRIT-U2: รายงานเช็ค/บัตร 3 รายการ — ตำแหน่งตาม Champ

หลักฐาน: `champ/champ/menuconfig.xml` (UTF-16LE, decode ด้วย Node) :309-363 + `bcspec/หน้าจอเมนูเช็ค บัตร/เมนู-ประกอบ.jpg` (ลำดับ: ทะเบียนหลักต่างๆ ›, ข้อมูลการเคลื่อนไหวเช็ค, รายการปฏิบัติการ, **รายงานข้อมูลหลัก ›**, **รายงานรายวัน ›**, คำนวณยอดคงเหลือของเช็คใหม่)

- **ตำแหน่งชั้นบนของ BC ถูกแล้ว**: `frontend/src/lib/menu-bars.ts:192-195` submenu `cheque-reports-master` (`menu_sub_reports_master_data` "รายงานข้อมูลหลัก") มี `cheque-received-report`, `cheque-issued-report`, `credit-card-report` (`menu-data.ts:472-474`) → ข้อกล่าวหา "วางผิดที่" ใน audit ล้าสมัย
- **ที่ยังขาด (ย้ายเข้างาน report registry + B0 ไม่ใช่งาน facade)**:
  - เมนูย่อยชั้นที่ 3 ของรายงานข้อมูลหลัก (`menuconfig.xml:328`): รายงานเช็ครับ 5380-5384 (ตามวันที่รับ, ตามวันครบกำหนด, ตามลูกค้า, ตามเลขที่, ตามแผนก); รายงานเช็คจ่าย 5386-5391 (ตามวันที่จ่าย, ตามวันครบกำหนด, ตามเจ้าหนี้, ตามเลขที่, ตามแผนก, ตามสมุดบัญชี); รายงานบัตรเครดิต 5393-5397 (ตามวันที่จ่าย, ตามวันครบกำหนด, ตามลูกค้า, ตามเลขที่, ตามแผนก) = 16 ชื่อ → ตามคำตัดสิน "ทุกชื่อในเมนู แต่เปิดจอรายงานกลาง" ให้แต่ละชื่อเป็น report id ใน registry (เรียงลำดับ = พารามิเตอร์ sort/group) ไม่ใช่จอใหม่
  - **รายงานรายวัน ›** (`menuconfig.xml:352-363`, ทั้งหมด "...พร้อมรายการย่อย"): 5409 ใบนำฝากเช็ครับ, 5410 รายวันเช็ครับผ่าน, 5411 ยกเลิกเช็ครับ, 5416 ขายลดเช็ครับ, │ 5412 รายวันเช็คจ่ายผ่าน, 5413 ยกเลิกเช็คจ่าย, │ 5414 รายวันขึ้นเงิน, 5415 ยกเลิกบัตร = 8 ชื่อ, ยังไม่มีใน BC; key หัว submenu `menu_sub_reports_daily` มีแล้ว (`REPORTS_DAILY` `menu-bars.ts:82`)
  - bcspec `Report_ระบบเช็ค บัตร/` มี PDF 28 ไฟล์ รวมแบบที่ menuconfig ไม่มี (ตามพนักงาน, ตามสถานะเช็ค, ประวัติเช็ครับ) → ยึด menuconfig เป็นรายการเมนู, PDF เป็นแบบฟอร์ม/คอลัมน์
- ข้อมูล: config `frontend/src/lib/erp-reports.ts:583/603/623` ไม่อยู่ใน `API_READY_REPORTS` (:833) → viewer แสดง `report_not_available` (`erp-report-viewer.tsx:198`) ซึ่งซื่อตรงแล้ว ไม่ใช่ facade; ข้อมูลจริงรอ B0 (ทะเบียนเช็ค) → **ไม่แก้ในงานนี้**

## 6. Language keys
ไม่ต้องเพิ่ม key ใหม่: ใช้ `menu_pending_development` (:4656), `menu_planned_workflow` (:4740), `menu_planned_description` (:4741). รายงานรายวัน/เมนูย่อย CRIT-U2 จะเพิ่ม key ในงาน report registry (แถวละ 13 คอลัมน์, th = ถ้อยคำ Champ ตรงตัว). ลบเฉพาะ key กำพร้าตาม §3.5

## 7. ความเสี่ยง
1. **ชนกับ workflow ที่กำลังแก้** `menu-screen-status.ts`/`menu-screen-status.test.ts` (:18, :89-97 ถูกแก้อยู่) → ทำหลัง commit นั้น แล้ว `git log`/`git diff` ก่อนแก้
2. receivedeposit: จอ ERP จริงหน้าตาต่างจากจอเงา; flag 32 ไม่ลง GL และไม่มีตัดมัดจำ → ผู้ใช้บันทึกได้แต่ยังไม่ครบตาม Champ; ต้องรายงานชัด
3. ผู้ใช้เสียจอที่ "ดูเหมือนใช้ได้" 30 จอ (แต่ข้อมูลไม่เคยถูกบันทึก — เสี่ยงน้อยกว่าคงไว้)
4. `help-standard-software` อ้างสถานะซอฟต์แวร์มาตรฐาน: ตามคำสั่งกรมสรรพากร ท.378/2543 แก้ไขโดย ท.643/2556 ข้อ 6 ซอฟต์แวร์มาตรฐานต้องแสดงเลขทะเบียน software house/เลขซอฟต์แวร์บนหน้าจอแรก (https://www.rd.go.th/27996.html, ทะเบียน https://www.rd.go.th/5991.html) — **ไม่ทราบว่า BC มีเลขทะเบียนหรือไม่** → แสดงข้ออ้างโดยไม่มีเลข = เสี่ยง ต้อง pending
5. etax: ADR `2026-09-19-champ-parity-menu-cut.md:65` ระบุ `etax-invoice` เป็นข้อยกเว้นที่ "คงไว้" → การทำ pending ไม่ได้ลบเมนู แต่ต้องบันทึกใน ADR/KMS ว่าจอคิวเป็น mock; backend XML คงไว้
6. Playwright `tests/uat-phase5.spec.ts` พังทันทีถ้าไม่เขียนใหม่ใน commit เดียวกัน
7. Pre-commit code-map check (ไฟล์ ≥950 บรรทัด) → ต้อง regenerate
8. Rule 20 float64 ใน `transaction/httpapi/models.go` เป็นของเดิม (`_notes_*`) ไม่อยู่ในงานนี้ — flag เท่านั้น

## 8. คำตัดสินที่ต้องให้ลุงจืดเลือก
- **D-1 ป้าย "รอพัฒนา" ในเมนู**: แนะนำ ชิปบนการ์ดในจออย่างเดียว (เมนูคงเหมือน Champ); ทางเลือก คืนป้ายในเมนูเฉพาะ `isMenuScreenPending` (ส่วนเสริมเล็ก) หรือลบ `MenuPendingBadge` ที่คืน null ทิ้งทั้ง 8 จุด
- **D-2 etax**: แนะนำ pending ตอนนี้ (คง XML backend) แล้วสร้างคิวจากใบกำกับขายจริงเมื่อ ERP sale posting เสร็จ; ทางเลือก สร้างคิวจริงทันที
- **D-3 transaction-lock (D1b)**: แนะนำ ทาง B (pending) ตอนนี้ แล้วทำทาง A (islockcost ตาม `ICDlgLockCostDlg.cpp:132-270`) ร่วมกับ D1 period lock; ทางเลือก ทำทาง A ทันที
- **D-4 help-about**: แนะนำ pending แล้วทำจริงโดยอ่าน release tag/VersionNumber จาก `/api/health`; ทางเลือก ใช้ `release.env`
- **D-5 help-standard-software**: แนะนำ pending จนลุงจืดให้เลขทะเบียน software house/ซอฟต์แวร์จากกรมสรรพากร; ทางเลือก ลบเมนู (Champ มีเมนูนี้หรือไม่ต้องตรวจเพิ่ม)
- **D-6 ฟอนต์รายงาน 5 เมนู**: แนะนำ pending; ทางเลือก ต่อ CSS var เข้ากับ report viewer (ต้องเก็บค่าใน backend ต่อบริษัท ไม่ใช่ localStorage)

## 9. ลำดับงาน
F1 ถอดเงา receivedeposit → F2 pending facade + ลบไฟล์ + test → F3 ตัดจอฟอนต์ → F4 ชิปบนการ์ด → F5 เขียน UAT ใหม่ (F1–F5 ทำเป็น commit เดียวได้ เพราะ test regex บังคับให้แก้คู่กัน) · F6 CRIT-U2 อยู่ในงาน report registry + B0 · จอที่ pending จะกลับมาเมื่อสร้างจริง: A4-1, A4-2, D1b-A, B0, B2, B3, B4, C1
