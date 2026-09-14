# บทเรียนจาก code review 2026-09-14 (GL v2 + คลังสินค้า + NumericInput)

รอบ review working tree ก่อน commit (8 มุม → verify 23 → แก้ 9 จุด) — เก็บเฉพาะกับดักที่จะเจอซ้ำ

- **`confirm()` จาก `useConfirmDialog` เป็น Promise** — `if (!confirm({...}))` ไม่เคยกันเพราะ Promise เป็น truthy เสมอ ต้อง `await` ใน handler ที่เป็น `async` (`frontend/src/app/system-settings/warehouse-tree-view.tsx` onClick เลือกคลัง; ตัวอย่างที่ถูกอยู่ที่ `handleDeleteWarehouse`)
- **backend `UpdateLocation` แทนที่ทั้ง embedded doc** (`backend/internal/warehouse/services/warehouse_location_http_service.go` `dataDoc.WarehouseLocation = doc`) — PUT จากตารางแก้ไขต้อง spread record เดิม (`companyguids`, `allow*/blocked*`, `sortcode`) ก่อน override field ที่แก้ ไม่งั้นค่าหายเงียบ ๆ
- **batch save ที่ยิงทีละแถว** — ต้องอัปเดต local state ทันทีหลังแต่ละ request สำเร็จ (`setDeletedLocationGuids` / `isNew:false` / `isModified:false`) ไม่ใช่รอจบทั้งชุด ไม่งั้นพังกลางทางแล้ว retry จะยิง DELETE/POST ซ้ำของที่สำเร็จไปแล้ว
- **GL Kafka projection worker ใช้ group คงที่ `gl-v2-projection`** (`backend/internal/generalledger/httpapi/http.go` `newRuntime`) — prod รัน consumer block ทั้งใน `mainapi` (ไม่มี `DEV_API_MODE`) และ `worker` (`CONSUMER_GROUP_NAME` ต่างกัน) ถ้า derive group จาก env ทุก event จะถูก apply 2 รอบ และ `recalculate` rebuild ซ้ำ
- **การลงทะเบียน consumer ห้าม `panic`** (`backend/main.go` บล็อก consumers) — config Kafka ผิดต้อง log แล้วให้ API ทำงานต่อ เพราะ mainapi เป็น process เดียวกัน
- **journal ปิดงบ/ยอดยกมามีได้ถึง `ProcessBalanceRowLimit` บรรทัด** (`backend/internal/generalledger/models.go` `Validate`) — เพดาน 500 ใช้กับรายการมือเท่านั้น เพราะ process สร้าง 1 บรรทัดต่อ (บัญชี, แผนก, โครงการ) ต่อสาขา
- **report GL หน้าละไม่เกิน 1000 แถว** (`postgres.go` clamp) — จอที่ต้องใช้ทุกแถว (statement designer) ต้องวนหน้าถัดไปด้วย `sequence` เดิม ไม่ใช่ขอ limit ใหญ่ครั้งเดียว
- **`NumericInput` default 2 ทศนิยมแต่ห้ามปัดค่าที่เก็บ** — `maximumFractionDigits = max(decimals, ทศนิยมจริงของค่า)` (0.125 ต้องโชว์ 0.125 ไม่ใช่ 0.13); test ใน `numeric-input.test.ts`
- **`DevDomInspector` บน production เป็นความตั้งใจ** — reviewer จะ flag ว่าไม่มี `NODE_ENV` gate แต่ลุงจืดสั่งเปิดตาม `decisions/2026-09-12-enable-dom-inspector-on-production.md` ห้ามใส่ gate กลับ
- **ยังไม่แก้ (ต้องตัดสินใจ):** เมนู `/line-oa` หายจากผังเมนูใหม่ (AGENTS.md ระบุว่าต้องอยู่ในเมนูหลัก) และหน้าจอที่ถูกย้ายไประดับ Holding (`/employee`, `/user`, `/permissiongroup`) ไม่อยู่ใน `MENU_SECTIONS` ทำให้ `useScreenActions` fallback เป็น ALL — ถามลุงจืดว่าจะวางเมนู LINE ไว้กลุ่มไหน และสิทธิ์ระดับ Holding ควร gate อย่างไร
