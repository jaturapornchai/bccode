# โดเมนสินค้า (Product / Barcode / Unit / BOM) — backend หลักถูกลบ 2026-09-23, frontend ยัง "รอพัฒนา"

> ตรวจล่าสุด: 2026-09-25 — เขียนใหม่หลัง ADR ถอด MongoDB/Kafka/Redis/ClickHouse; บทความเดิม (Mongo SoT + outbox + Kafka projection + listing v2) ใช้ไม่ได้แล้วเพราะโค้ดถูกลบจริง ไม่ใช่แค่ปิดการเชื่อมต่อ

## 1. สถานะปัจจุบัน

- backend สินค้าเดิมของ mainapi (`backend/internal/product/**` — product, productbarcode, productgroup, productcategory, bom, option, promotion ฯลฯ — รวม `productimport` และหน่วยนับ) **ไม่มีอยู่ในรีโปแล้ว**: `ls backend/internal` ไม่มีโฟลเดอร์เหล่านี้ และ `backend/main.go` ไม่ import แพ็กเกจสินค้าใด ๆ
- การลบเป็นส่วนหนึ่งของ ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md): จอที่ API เดิมอยู่บน MongoDB (รวมสินค้าทั้งโดเมน) ถูกลบ backend ทิ้ง และลุงจืดสั่งให้โฟกัส GL ก่อน
- **ห้ามเพิ่ม Mongo driver, outbox หรือ Kafka consumer กลับมาเพื่อ "ซ่อม" โดเมนนี้** — ถ้าจะทำสินค้าใหม่ต้องออกแบบ schema/endpoint บน PostgreSQL ต่อ holding ตั้งแต่ต้น

## 2. โค้ด goapi ที่ยังค้าง (อ่าน PostgreSQL แต่ไม่มีใครเติมข้อมูล)

- `backend/internal/goapi/bootstrap.go:373-382` ยังลงทะเบียน `POST /goapi/api/product/search`, `/api/product/barcode`, `/api/product/search/unified`, `/api/process/product-balance`, `GET /api/product/cache/stats`, `POST /api/product/cache/clear` และ `/api/search/aliases` — handler อยู่ที่ `backend/internal/goapi/handlers/product_search.go`, `product_cache.go`, `product_balance_update.go`, `search_aliases.go` อ่านตาราง `product` / `productbarcode` / `docdetail` / `search_aliases` ในฐานของ holding
- ไม่มีโค้ดนอกเทสต์สร้างตารางหรือ insert แถว `product` / `productbarcode` แล้ว (ทางเขียนเดิมคือ Kafka projection ที่ถูกลบ; เหลือแค่ UPDATE ยอดคงเหลือใน `backend/internal/goapi/process/process-stock/process-product-balance-update.go` และไฟล์ `backend/migrations/add_productbarcode_materialtype_projection.sql` ที่ไม่มีโค้ดใดรัน) และ BFF `/api/goapi/[...goPath]` ไม่ได้เปิด path เหล่านี้ (allowlist `frontend/src/app/api/goapi/[...goPath]/route.ts:14-32`) — จึงไม่มีจอเรียก
- ที่ยังใช้งานได้จริง: อัปโหลดรูป/วิดีโอ `POST /goapi/image/upload`, `/goapi/video/upload` (`bootstrap.go:396-397`) ลง S3/MinIO — BFF ของจอสินค้า `frontend/src/app/api/product-barcode/image/route.ts`, `video/route.ts` ชี้มาที่นี่

## 3. Frontend ยังอยู่ครบ แต่ขึ้น "รอพัฒนา"

- จอเดิมยังอยู่ใน `frontend/src/app/menu/` (`product-screen.tsx`, `product-barcode-screen.tsx`, `product-barcode-shelf-screen.tsx`, `product-price-history-screen.tsx`, `product-set-screen.tsx`, `tab-product-*.tsx`) และเลือกจอใน `frontend/src/app/menu/main-menu-screen.tsx:2861-2901`; route ประกาศใน `CUSTOM_MENU_SCREEN_ROUTES` (`frontend/src/lib/menu-screen-status.ts:12-16`)
- `/product`, `/productbarcode`, `/productbarcodeshelf`, `/pricehistory`, `/inventory/product-sets`, `/productset` อยู่ใน `RETIRED_BACKEND_ROUTES` (`frontend/src/lib/menu-screen-status.ts:40-43`) จึงถูก `isMenuBackendRetired()` (บรรทัด 51-56) ตัดสินว่า "ยังไม่พร้อม" เสมอ
- BFF proxy เดิมยังไม่ถูกลบ แต่ปลายทางไม่มีแล้ว: `frontend/src/app/api/product/[[...productPath]]/route.ts` → mainapi `/product*` (บรรทัด 33, 38, 51, 70, 98, 117), `frontend/src/app/api/product-barcode/[[...barcodePath]]/route.ts` → `/product/barcode*` (บรรทัด 22, 36, 56, 71, 79), `frontend/src/app/api/product-barcode/list/route.ts:39` → goapi `/api/product/barcode/list` (ไม่มี route นี้ใน goapi แล้ว) — จอไม่เรียกเพราะถูกกันด้วย "รอพัฒนา" ก่อน

## 4. สิ่งที่ยังไม่ตัดสินใจ

- ยังไม่มีกำหนดว่าจะสร้างโดเมนสินค้าใหม่บน PostgreSQL เมื่อไร/ขอบเขตแค่ไหน — ADR ระบุว่าสร้าง API บน PostgreSQL ทีละโมดูล **เมื่อลุงจืดสั่ง** AI ห้ามเริ่มเอง
- โค้ด goapi ใน §2 จะเก็บไว้เป็นฐานหรือลบทิ้งยังไม่มีคำตอบ — ถามลุงจืดก่อนแตะ
- ถ้าจะรื้อฟื้น ต้องเทียบเคียงกับ `D:\project-champ` ก่อนออกแบบฟิลด์/ผังเมนู ตามกฎ "ยึด Champ เป็นต้นแบบ"
