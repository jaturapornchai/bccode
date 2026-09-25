# โมดูล legacy/อื่น ๆ ใน backend/internal — เกือบทั้งหมดถูกลบทิ้ง 2026-09-23

> ตรวจล่าสุด: 2026-09-25 — เขียนใหม่ทั้งหมด บทความเดิม (แผนที่โมดูล legacy 40+ ตัว) ใช้ไม่ได้แล้วเพราะ **เกือบทุกโมดูลที่บทความเดิมสำรวจถูกลบออกจากรีโปจริง** ไม่ใช่แค่ปิดการเชื่อมต่อ

## 1. สิ่งที่ถูกลบทั้งหมด (ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md))

- โมดูลต่อไปนี้ที่บทความเดิมสำรวจไว้ **ไม่มีไฟล์ใดใน git ใต้ `backend/internal/` แล้ว** (ตรวจด้วย `git ls-files backend/internal/<module>` 2026-09-25): `vfgl/*` (chartofaccount, accountgroup, journal, journalreport ฯลฯ — แทนที่ด้วย `backend/internal/generalledger/`), `debtaccount/*`, `creditorprocess`, `debtorprocess`, `coupon`, `shopcoupon`, `task`, `mastersync`, `member`, `notify`, `images`, `slipimage`, `ocr`, `currency`, `dimension`, `channel`, `form`, `productsection`, `smlaiproduct`, `restaurant`, `shopdesign`, `pos`, `order`, `pickandpack`, `payment`, `paymentmaster`, `masterexpense`, `masterincome`, `purchasetype`, `filestatus`, `smsreceive`, `apikeyservice`, `requestapi`, `report` (`reportqueryc`/`reportquerym`), `reportquery`, `repositories` (generic Mongo repo), `uat`, `storefront`, `logistics`, `datatransfer`, `stockprocess`, `transaction`, `warehouse`, `product` — ดูโดเมนที่ยังมีจอค้างใน [`04-product-domain.md`](04-product-domain.md), [`05-transaction-sales-purchase.md`](05-transaction-sales-purchase.md), [`06-transaction-stock.md`](06-transaction-stock.md)
- บนดิสก์อาจยังเห็นโฟลเดอร์ `backend/internal/{vfgl,member,datatransfer,transaction,stockprocess}/` แต่ข้างในมีแค่ `logs/*.log` ที่ถูก `.gitignore:54` (`backend/internal/**/logs/`) ข้าม — **ไม่มีไฟล์ใดใน git** ลบทิ้งในเครื่องได้โดยไม่กระทบ build
- `backend/cmd/*` เดิม (`cmd/app`, `cmd/datatransfer`, `cmd/removeshopdata` ฯลฯ) ถูกลบ — `ls backend/cmd` เหลือแค่ `glseed/` (คำสั่งเติมข้อมูลหลักฐานลูกหนี้/เจ้าหนี้/ธนาคาร/งบประมาณ/ใบสำคัญให้จอ GL ผ่าน service layer — `backend/cmd/glseed/main.go:1-11`)
- `backend/pkg/stockcalculator` และ engine ต้นทุนรุ่นเก่า (`process-stock-calc-cost.go`, `processstockcalculationstate*`, `distributedlocks`) ไม่มีแล้ว

## 2. โมดูลจากบทความเดิมที่ยังมีชีวิต

| module | หน้าที่ | สถานะ | อ้างอิง |
|---|---|---|---|
| `media` | อัปโหลดวิดีโอ/รูปผ่าน file persister (S3/MinIO, env `S3_*`) | LIVE แต่ยังไม่มีจอเรียก (grep `media/upload` ใน `frontend/src` = 0) | `POST /media/upload/video`, `POST /media/upload/image` (`backend/internal/media/media_http.go:37-38`); ลงทะเบียนที่ `backend/main.go:166` |
| `demo` | เปิด/ปิดปุ่ม Demo login ด้วย env `BCAI_DEMO_LOGIN_ENABLED`, `BCAI_DEMO_USERNAME` | LIVE (lib) | `backend/internal/demo/demo.go:15-16,20-21,30`; ผู้ใช้ `authentication/demo_login.go`, `authentication/authentication_http.go`, `organization/creator_access.go` |
| `encrypt` | `GenerateSHA256Hash` | LIVE (lib) | `backend/internal/encrypt/encrypt.go:13`; ผู้ใช้ `backend/pkg/microservice/auth.go` |
| `models` | struct กลาง (identity, activity, pagination, api_response, postgresql ฯลฯ) | LIVE (lib) | `backend/internal/models/` |
| `utils` | helper กลาง (auth, business_code, holding_code, password, random, request_param) | LIVE (lib) | `backend/internal/utils/` |

- `backend/internal/database/` ไม่มีโค้ด Go แต่เก็บสคริปต์ SQL 3 ไฟล์ (`fresh_provision.sql`, `fresh_provision_all.sql`, `schema_full.sql`) = สำเนา DDL ตารางฐานกลาง + GL ที่ยังอัปเดตตามงาน (ดู ADR [`decisions/2026-09-25-company-tax-address.md`](decisions/2026-09-25-company-tax-address.md)) — ไม่มีโค้ดใดรันตอน runtime; ถูกอ้างในคอมเมนต์ `backend/internal/generalledger/uat_fixture_test.go:18`
- โมดูลอื่นใน §1 **ไม่มีอยู่แล้ว** — จะสร้างใหม่บน PostgreSQL เมื่อลุงจืดสั่งเท่านั้น (ADR) ห้ามกู้โค้ดเดิมจาก git กลับมาเพราะพึ่ง Mongo/Kafka

## 3. `backend/main.go` คือแหล่งความจริงเดียวตอนนี้

- `backend/main.go` (192 บรรทัด, binary เดียว) ลงทะเบียน `authentication`, `shop` (+`ShopMemberHttp`), `shop/employee`, `fixedasset` (`fahttp`), `generalledger` (`glhttp`), `mcptoken`, `organization/{company,branch,businesstype,rolepermission}` (`backend/main.go:153-165`), `media` (บรรทัด 166) และ `goapi` ใต้ `/goapi/*` + `GET /api/language/:lang` (บรรทัด 168-179)
- ไม่มีการสลับโหมดด้วย `DEV_API_MODE` แล้ว (ไม่มีโหมด legacy consumer/migration) — ค่า `DEV_API_MODE=2` ยังค้างใน `backend/Dockerfile:42`, `backend/Dockerfile.local:34`, `backend/docker-compose.yml:39`, `deploy/account/provision-server.sh:85` และมีแค่ mapping ใน `backend/internal/setupconfig/loader.go:34`, `backend/internal/goapi/setupconfig/loader.go:56` แต่ไม่มีโค้ดใดอ่านค่านี้ไปแยกทาง
- โมดูลอื่นที่ยังมีไฟล์ `.go` ใน `backend/internal/` และมีบทความของตัวเอง/อยู่นอกขอบเขต "legacy": `centraldb`, `config`, `goapi`, `logger`, `mcpgateway`, `middlewares`, `pdftext`, `rdfile`, `rdform`, `setupconfig`, `taxaddress`, `whtcert` — ตรวจจากโค้ดโดยตรงหรือบทความ `01`–`03`, `07`, `09`–`11`, `13`–`14`
