# บัญชีแยกประเภทใหม่สำหรับการอัปเกรด Champ

วันที่ 2026-09-11 · ขอบเขตที่ลุงจืดยืนยัน: ทั้ง 35 เมนู · runtime ล่าสุด mainapi/worker deploy `r20260911-gl-kafka-1` เวลา 18:11:35 น. ไทย และ frontend `r20260911-gl-demo-thai-1` เวลา 18:29:44 น. ไทย ผ่าน MongoDB → Kafka → PostgreSQL; mainapi/worker/frontend healthy ข้อมูลตัวอย่างวัสดุก่อสร้างผ่าน seed และการตรวจอิสระ Mongo/Kafka/PG รวม final UI หลัง frontend patch แล้ว: 5 reports ตรง expected และ smoke 35 routes (33 ใช้งาน / 2 รอข้อมูล) ดู [Kafka release](../decisions/2026-09-11-deploy-gl-kafka-demo.md)

## วัตถุประสงค์และหลักฐาน

ใช้ลำดับงาน Champ เป็นหลัก: ข้อมูลหลักและยอดยกมา → สมุดรายวัน → ตรวจ/ผ่านรายการ → รายงาน → ปิดงบและสิ้นปี โดยเพิ่มการทำงานบนเว็บ การตรวจสิทธิ์บริษัท/สาขา จำนวนเงินแม่นยำ และประวัติที่ตรวจสอบได้

- Champ `D:/project-champ/champ/champ/menuconfig.xml:552` เป็นที่มาของลำดับเมนู; `glfrmjournal.cpp:478` และ `:766` ตรวจเอกสารซ้ำ/เดบิตเครดิต; `glfrmcloseperiod.cpp:237` ใช้บัญชีกำไรขาดทุนและกำไรสะสมแยกกัน
- `GLReProcess.cpp:130` สร้างบัญชีจากเอกสารต้นทาง ส่วน `GLDlgReprocess.cpp:196` คำนวณยอดผ่านรายการใหม่ เป็นคนละกระบวนการ
- `YearEndProcess_Page1.cpp:559` มีการลบประวัติในระบบเดิม ระบบใหม่นี้เก็บประวัติและสร้างยอดยกมาในปีถัดไป
- รหัสเมนูและ route เดิมคงอยู่: `frontend/src/lib/menu-data.ts:574`; จุดเข้าหน้าจอ `frontend/src/app/gl/general-ledger-screen.tsx:19`

## วิธีใช้งาน

1. เปิด **ผังบัญชี** สร้างบัญชีคุม/บัญชีลงรายการ พร้อมหมวดและด้านบัญชี ระบุบัญชีเงินสดเพื่อใช้รายงานกระแสเงินสด
2. แท็บ **ปีบัญชีและบัญชีปิดปี** กำหนดช่วงวัน สกุลเงิน ทศนิยม และบัญชีทุนสองบัญชีสำหรับปิดกำไรขาดทุน/กำไรสะสม จากนั้นสร้างงวดใน **ล็อกงวดบัญชี**
3. บันทึกยอดยกมาในวันแรกของปี หรือสมุดรายวันขาย ซื้อ รับ จ่าย และทั่วไป ใส่จำนวนเงินเป็นเลขทศนิยมไม่ใส่จุลภาค เดบิตและเครดิตต้องเท่ากันก่อนบันทึก
4. ตรวจร่างและผ่านรายการ รายงานนับเฉพาะรายการที่ผ่านแล้ว แก้จำนวนเงินหรือลบรายการที่ผ่านแล้วไม่ได้ การกลับรายการสร้างเอกสารใหม่ที่สลับเดบิตเครดิตและเก็บต้นฉบับ
5. ปิดงบ: ตรวจยอดถึงวันที่เลือก สร้างฉบับร่างแยกสาขา ตรวจและผ่านรายการ แล้วล็อกงวดเอง ทุกแผนก/โครงการคงเดิม
6. สิ้นปี: ผ่านรายการร่างและปิดรายได้/ค่าใช้จ่ายให้หมด เลือกปีถัดไปที่ต่อเนื่อง สกุลเงินและทศนิยมตรงกัน ระบบปิดปีเดิมพร้อมสร้างร่างยอดยกมา ถ้าทุกยอดเป็นศูนย์จะปิดปีโดยไม่มีเอกสารเปล่า ยอดยกมาที่ระบบสร้างแก้/ลบไม่ได้ ให้ผ่านรายการและปรับปรุงด้วยรายการแยก
7. รายงานเลือกปีและช่วงวัน กรองบัญชี/สาขา/แผนก/โครงการ ส่งออก CSV ได้ครบทุกหน้า ส่วนสำรองข้อมูลส่ง JSON จำนวนเงินเป็นสตริง

## สัญญาจำนวนเงินและการบันทึก

- JSON ใช้ decimal string; Go ใช้ `shopspring/decimal`; MongoDB ใช้ `Decimal128`; PostgreSQL ใช้ `NUMERIC(38,8)` (`backend/internal/generalledger/amount.go:1`, `schema.sql:1`)
- ข้อมูลเข้าไม่เกิน 26 หลักหน้าจุดและ 8 ตำแหน่ง ปีบัญชีกำหนด scale 0–8; ไม่ปัดเศษให้อัตโนมัติ จำนวนเงินเกิน scale ถูกปฏิเสธ ฝั่งแสดงผลรองรับยอดรวมที่มากกว่าขนาดข้อมูลหนึ่งรายการ
- Decimal128 อาจคืน `1E-8`; BSON decoder แปลงเป็น decimal ปกติก่อนตรวจ precision โดย API ยังไม่รับเลขยกกำลัง (`amount.go:79`, `amount_precision_audit_test.go:9`)
- `_id` เป็น SHA-256 ย่อ 32 ตัวอักษร; วันบัญชีเป็นสตริง `YYYY-MM-DD`; เวลาประวัติ UTC เป็น BSON Date ความละเอียดมิลลิวินาที ไม่ใช้วันบัญชีเป็น timestamp
- Mongo transaction ใช้ snapshot/majority บันทึกข้อมูล + `gl_events` + ลำดับบริษัทพร้อมกัน การล็อก counter ต่อบริษัทป้องกัน race ระหว่างแก้ข้อมูลอ้างอิง ผ่านรายการ ล็อกงวด และปิดปี (`store.go:96`)
- `requestid` เดิมพร้อม actor/branch/body เดิมคืนผลเดิม; ถ้าข้อมูลต่างกันให้ปฏิเสธ เอกสารใช้ version เพื่อป้องกันเขียนทับพร้อมกัน
- ปีที่มีรายการหรืองบประมาณ/ประมาณการแล้วห้ามเปลี่ยนสกุลเงิน/ทศนิยม ปีปิดแล้วแผนเดิมห้ามแก้ ลบ หรือย้ายปี (`references.go:60`, `mutations.go`)

## การประมวลผลและสิทธิ์

- `/api/gl/*` เป็น Next.js BFF ไป `/gl/v2/*` ของ main API; ตรวจ membership/holding/company/branch และ permission ปัจจุบันจากต้นทาง สิทธิ์เข้าเมนูไม่เท่ากับสิทธิ์ create/update/delete (`httpapi/http.go:107`, `httpapi/permissions_test.go:5`)
- MongoDB: `chart_of_accounts`, `fiscal_year`, `gl_account_groups`, `gl_product_account_groups`, `gl_account_mappings`, `gl_budgets`, `gl_periods`, `gl_cash_forecast`, `gl_journals`, `gl_events`, `gl_controls`
- PostgreSQL ฐานตาม holding: สำเนาทุกทรัพยากรใน `gl_records` แยก company/kind/id, รายการผ่านแล้วใน `gl_lines`, ประวัติใน `gl_events`, ลำดับใน `gl_projection_state` ใช้ advisory lock และ transaction; trigger ห้ามแก้/ลบ/truncate ประวัติ
- **Transport Kafka deploy แล้วใน r20260911-gl-kafka-1:** MongoDB เป็นข้อมูลต้นฉบับที่เชื่อถือได้ คำสั่ง commit ข้อมูล + outbox ก่อน relay ส่งหัวคิวต่อบริษัทไป Kafka; เส้นทาง command ไม่เรียก Project PostgreSQL โดยตรง การส่งล้มเหลวคืนผลบันทึกสำเร็จพร้อม `projectionpending` เพราะ MongoDB commit แล้ว (`store.go:194`, `store.go:223`)
- Topic `bc-gl-projection-v1` ส่งเฉพาะ reference 6 ฟิลด์ `schemaversion/eventid/holdingcode/businesscode/sequence/eventhash` ไม่มี payload การเงิน ใช้ SHA-256 ของ holding/company เป็น partition key; consumer ตรวจ reference แล้วอ่าน immutable event ต้นฉบับจาก MongoDB เพื่อเทียบ hash ก่อนประมวลผล (`event_reference.go:23`, `kafkatransport/transport.go:127`, `store.go:252`)
- ลำดับยืนยันคือ **PostgreSQL Project สำเร็จ → Rebuild เมื่อเป็นคำสั่งคำนวณใหม่ → MongoDB delivered=true และลบ retryafter → commit Kafka offset**; broker ACK อย่างเดียวไม่เปลี่ยน delivered หาก handler หรือ commit offset ล้มเหลวไม่อ่าน offset ที่สูงกว่าและเปิด reader กลุ่มเดิมเพื่อ replay ข้อผิดพลาดบัญชีห้าม acknowledge-and-skip (`store.go:268`, `store.go:284`, `kafkatransport/transport.go:209`, `backend/internal/product/projection/consumer.go:23`)
- Relay ส่งครั้งละหัวคิวต่อบริษัท ไม่ให้เหตุการณ์ใหม่แซงเหตุการณ์เก่าที่ยังไม่ delivered; ใช้ฟิลด์ `retryafter` เดิม หลัง broker ACK พัก 2 วินาที หลังส่งผิดพลาดพัก 5 วินาที ไม่เพิ่ม persisted fields เพิ่มเพียง index `gl_outbox_company_pending` ตามลำดับ `holdingcode, businesscode, delivered, sequence` ทุกสมาชิก ascending และไม่ unique (`store.go:91`, `store.go:223`, `store.go:308`)
- รายงานเปิดเมื่อ MongoDB/PG sequence ตรงกันและไม่มี event ที่ยังไม่ delivered ถ้ายังไม่พร้อมตอบ HTTP 409 พร้อม `errorcode: GL_PROJECTION_PENDING`; frontend retry เฉพาะ GET รหัสนี้ 100/250/500/1000 ms ไม่เกินประมาณ 15 วินาทีพร้อม abort ไม่ส่ง POST ซ้ำและไม่ซ่อน version/snapshot conflict (`store.go:343`, `httpapi/http.go:293`, `frontend/src/lib/general-ledger-api.ts:17`)
- รายงานอ่าน PostgreSQL แบบ repeatable-read; export ตรวจ sequence ทุกหน้าและทุก resource ป้องกันไฟล์รวมข้อมูลต่างเวลาปะปน ไม่คำนวณรายงานจาก MongoDB
- ปิดงบและสิ้นปีอ่านยอดแยกบัญชี/สาขา/แผนก/โครงการจาก snapshot เดียว ตรวจ sequence อีกครั้งภายใน Mongo transaction; ไม่บังคับสมดุลต่อแผนก/โครงการเพราะบางบรรทัดต้นฉบับอาจไม่ได้กำหนดมิติ (`processes.go:22`, `reports_process.go`)
- คำนวณยอดผ่านรายการใหม่ replay จากประวัติ PG ที่แก้ไม่ได้ ไม่สร้างบัญชีจากเอกสารซื้อขายเดิม (`postgres.go:332`)

## ข้อผิดพลาด UI ที่พบจากภาพ UAT

กฎ CSS กลางนอก Tailwind layer กำหนด section กว้าง 100% ทำให้ pane รายการกินพื้นที่ทั้งหมดและบีบฟอร์มบัญชี แก้เฉพาะ GL โดยใช้ div เป็น pane พร้อม min-w-0; เพิ่มตรวจขนาด pane และดูภาพจริง การเช็คว่าไม่มี document overflow อย่างเดียวจับปัญหานี้ไม่ได้ (`frontend/src/app/globals.css:383`, `frontend/src/app/gl/gl-common.tsx:84`)

## การตั้งค่าและ dependency

ใช้ dependency ในโครงการ: Go, MongoDB replica set, PostgreSQL, Kafka และ segmentio/kafka-go, Redis/auth เดิม, shopspring/decimal, lib/pq, Next.js/React, BigInt และ Playwright ไม่มี plugin ใหม่

ต้องมีฐาน PostgreSQL ชื่อ holding ตัวพิมพ์เล็กก่อนเปิด GL; package สร้างตารางและ index เอง ใช้ MongoPersister/Persister config เดิม ไม่เพิ่มรหัสผ่านใน repo ส่วน `BCAI_LOCAL_BACKEND_URL` ใช้กำหนด backend ของ frontend ตอนเริ่ม/build

ซอร์ส Kafka ต้องมี `KAFKA_SERVER_URL` เป็น broker host:port และเตรียม topic `bc-gl-projection-v1` ไว้ก่อน; ไม่ auto-create topic และใช้ RequireAll สำหรับการส่ง Consumer group ใช้ `CONSUMER_GROUP_NAME` ต่อท้าย `-gl-v2` ค่าเริ่มต้น `03-gl-v2` รองรับ `KAFKA_SECURITY_PROTOCOL` ว่าง/PLAINTEXT หรือ SSL พร้อมชุด `KAFKA_SSL_CA_FILE/KAFKA_SSL_CERT_FILE/KAFKA_SSL_KEY_FILE`; ไม่รองรับ SASL ใน transport นี้ และ `ENABLE_KAFKA=false` ทำให้ GL runtime รอบนี้เริ่มไม่ได้ (`backend/internal/config/config_mq.go:21`, `httpapi/http.go:122`, `kafkatransport/transport.go:54`)

## ข้อจำกัดที่ยังต้องมีข้อมูลต้นทาง

- **ประมวลผลข้อมูลบัญชีจากเอกสารซื้อขายเดิม**: ยังไม่เปิดใช้งาน เพราะ writer เดิมเก็บยอดเป็น float64 และไม่มี company code ที่ยืนยันครบ (`backend/internal/transaction/models/transaction.go:42`, `backend/internal/goapi/models/mongo-trans-model.go:138`) ต้องมีต้นฉบับยอดที่กระทบแล้วและบริษัทที่ยืนยัน ห้าม cast float เป็น decimal แล้วถือว่าเป็นยอดถูกต้อง
- **ส่งออก XBRL**: เป็นหน้าเตรียมกระดาษทำการ ยังไม่สร้างไฟล์สำหรับยื่น ต้องมีแม่แบบและประเภทกิจการ/มาตรฐานที่ตรงกับบริษัท และตรวจด้วยตัวตรวจรับของ DBD ก่อน [คู่มือ DBD](https://efiling.dbd.go.th/efiling-documents/01_ManualFN.pdf) หน้า 6–10 ระบุการดาวน์โหลดแม่แบบแยกกิจการและขั้นตอนสร้างไฟล์
- งบ/กราฟและ cash flow เป็นรายงานจากบัญชีและประเภทเงินสดที่ผู้ใช้กำหนด ไม่ใช่แม่แบบงบที่ผ่านการรับรอง DBD; ประมาณการเงินสดเป็นแผนที่ผู้ใช้บันทึกเอง
- เอกสารไม่เกิน 500 บรรทัด; process ไม่เกิน 10,000 กลุ่มมิติ และต้องสร้างเอกสารที่สมดุลแต่ละสาขาภายในขนาดที่รองรับ; export ไม่เกิน 100,000 รายการ ไม่มีการส่งออกบางส่วนเงียบ ๆ
- JSON สำรองยังไม่มีหน้ากู้คืน และไม่ทดแทน backup ฐานข้อมูล/ประวัติทั้งหมด
- Deploy GL v2 วันที่ 2026-09-11 เวลา 17:25 น. ไทย และอัปเดต transport Kafka เวลา 18:11:35 น. ไทย โดยไม่แปลงรายการเดิม ดู [ผล Kafka deploy และ rollback](../decisions/2026-09-11-deploy-gl-kafka-demo.md)

## ผลตรวจและวิธีทดสอบ

**แยกรุ่นหลักฐาน:** ผล Go/CRUD/UAT และ deployment ที่ระบุด้านล่างเป็นรอบ `r20260911-gl-v2-1` ก่อนเปลี่ยน transport Kafka ห้ามใช้แทนผลรอบใหม่ รุ่น Kafka `r20260911-gl-kafka-1` deploy แล้ว: GL Linux regression 24 top-level tests + 63 subtests ไม่ fail/skip, real Kafka ตรวจ commit ordering/replay และยอดเดบิต=เครดิต `0.30000003`; release backend/outbox/projection และ frontend 393 tests ผ่าน มี 15 packages เดิมใน quarantine ที่ตรวจ compile อย่างเดียว ดู [metrics และข้อจำกัด](../../evidence/2026-09-11-gl-kafka/README.md)

ชุดทดสอบใช้ฐาน Mongo/PG แยกตาม UUID และตรวจ Mongo ทันทีทุกการสร้าง/แก้/ลบ/ผ่านรายการ จากนั้นตรวจ PG และล้างเฉพาะข้อมูลทดสอบ

```powershell
# จาก backend; URI/DSN ต้องเป็นฐานทดสอบแยกเท่านั้น
$env:BC_GL_TEST_MONGO_URI='mongodb://127.0.0.1:15444/?directConnection=true&replicaSet=gltest'
$env:BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:15443/postgres?sslmode=disable'
go test -tags=integration ./internal/generalledger
# httpapi ใช้ Linux เพราะ backend มี Kafka dependency ที่ไม่ build บน Windows
go test ./internal/generalledger/httpapi
```

Go GL รอบสุดท้ายผ่าน 16 top-level tests และ 38 subtests ไม่ข้ามทดสอบ ใช้เวลา 4.689 วินาที; HTTP permission test และ build ผ่าน Linux; frontend Vitest 5 ไฟล์ 29 tests, TypeScript และ lint ของไฟล์ที่เพิ่มผ่าน โดยครอบคลุมเงินแม่นยำ, CRUD และ soft-delete, scope holding/company/branch, การผ่าน/กลับรายการ, คำขอซ้ำและ double-click, ล็อกงวด, ปิดงบ/สิ้นปีสองสาขาหลายมิติ, snapshot ที่เก่า, ประวัติ PG แก้ไม่ได้ และรายงาน 15 แบบ (`store_integration_test.go`, `processes_integration_test.go`, `postgres*_integration_test.go`)

ตัวอย่างยอดยืนยัน: `0.1 + 0.2 = 0.3` และ `0.00000001 + 0.00000002 = 0.00000003`; กำไร `300.50 − 50.30 = 250.20` คงเดิมหลังปิดงบ พร้อม trial balance รายได้/ค่าใช้จ่ายปิดเป็นศูนย์และยอดทุกมิติยกไปปีใหม่ตรงกัน

UAT เบราว์เซอร์บน production build ผ่าน 1 test ใน 29.621 วินาที ไม่มี skipped/flaky/error (build ID `eikUmz6VbY1pwLz8ThIGu`) เปิดตรวจครบ 35 เส้นทาง และทดสอบ CRUD, ล็อกงวด, ผ่าน/กลับรายการ, คำขอซ้ำ, dirty form และโหลดบัญชีใหม่โดยคงฟอร์มเดิม ตรวจภาพ light/dark ด้วยปุ่มจริงครบ 1600/1280/1024/768 รวม 8 ภาพ ไม่มี console error ใช้ runtime ทดสอบแยก 3011/8891 ไม่ใช่เว็บจริง

- [ผล Playwright](../../evidence/2026-09-11-general-ledger/browser-results.json) และ [บันทึกการรัน](../../evidence/2026-09-11-general-ledger/browser-run.log)
- [ภาพหน้าจอ 8 แบบ](../../evidence/2026-09-11-general-ledger/browser/general-ledger-uat-GL-actu-8b3fd-outes-and-responsive-themes)
- [กระทบยอด MongoDB/PG](../../evidence/2026-09-11-general-ledger/production-reconciliation-r5.json): หลักฐาน Mongo ทันที 19 ขั้นตอน; เอกสารต้นฉบับและกลับรายการเดบิต=เครดิตใบละ `0.30000000` ตรง Decimal128 `0.3` และยอดสุทธิทุกบัญชีเป็นศูนย์
- [ผลล้างข้อมูลทดสอบ](../../evidence/2026-09-11-general-ledger/cleanup-result.json): ตรวจตัวตนก่อนล้างเฉพาะ Mongo `gl_browser_uat_20260911` และ PG `glbrowser20260911` ยืนยันฐาน PG อื่นคงเดิม หยุด frontend/API/MongoBrowser/Redis ของ UAT และลบ auth ชั่วคราวแล้ว

คำสั่งเบราว์เซอร์ใช้ `frontend/e2e/general-ledger-uat.spec.ts` โดยเปิด `GL_UAT_ENABLED=1`; ตัวทดสอบจำกัด URL/ฐาน/container ทดสอบและบันทึก seed `2026091101` ไม่เก็บ token ในหลักฐาน

## MongoModel ที่ตรวจจากระบบจริง

- ก่อน `projectRev=1423`; หลัง `projectRev=1480` ของ BC Ai Account
- แก้เฉพาะ field ของ `chart_of_accounts`/`fiscal_year` ใน diagram `efaf857d`; ตรวจฐาน appdb เดิมว่าทั้งสอง collection ไม่มีข้อมูลก่อนเปลี่ยนชนิด `_id`/วันบัญชีในแบบจำลอง
- เพิ่ม diagram `559002a1` บัญชีแยกประเภท 9 collections พร้อม relations และ workflow `gl-v2-exact-ledger` 8 steps สถานะ draft เพื่อบันทึก implementation ไม่อ้างว่าเป็นข้อกำหนดธุรกิจที่อนุมัติแล้ว
- `check_descriptions`, `lint_model` ของขอบเขตที่แก้ และ `lint_workflows` ผ่าน ไม่มี issue; ไม่มีการแก้ผลิตภัณฑ์ MCP หรือเขียนทับ code จาก generator
- รอบซอร์ส Kafka อ่านสดก่อนแก้ `projectRev=1480` → หลังแก้ `1482`: collection `gl_events` ID `7f90a74a` ใน diagram `559002a1` เพิ่ม index compound ที่ 4 โดยคงฟิลด์เดิมทั้งหมด และ workflow `gl-v2-exact-ledger` เปลี่ยนเป็น 12 steps/15 transitions สถานะ draft มี Mongo authoritative → Kafka reference → PG → Mongo delivered → offset พร้อมเส้นทางผิดพลาด
- ตรวจหลังแก้ด้วย `check_descriptions`, `lint_model`, `lint_workflows` ไม่มี issue และ reread diagram/workflow ยืนยัน revision 1482; MCP index schema ไม่มีช่องชื่อ index จึงระบุชื่อ runtime `gl_outbox_company_pending` ในคำอธิบาย collection และเก็บ field/order/unique จริงใน indexes ไม่เพิ่มฟิลด์ชดเชย

## การย้อนกลับ

Release ล่าสุด `r20260911-gl-kafka-1` สำรอง MongoDB/PG/config ก่อนปล่อยแล้ว วิธีคืน image และข้อจำกัด backup รอบนี้อยู่ใน [บันทึก Kafka deploy](../decisions/2026-09-11-deploy-gl-kafka-demo.md) การย้อน binary ต้องคง collection/tables/audit/topic/offset ไว้และตรวจ pending events เพราะรุ่นก่อนใช้ transport ต่างกัน ห้ามลบประวัติหรือแปลงยอดกลับเป็น float

## เมนูทั้งหมดและสถานะ

สถานะเชื่อมแล้วหมายถึงมีหน้าจอและ API ในโค้ดชุดนี้ ไม่ได้หมายถึงได้รับรองรูปแบบงบตามกฎหมาย

| เมนู | เส้นทาง | สถานะ |
|---|---|---|
| ผังบัญชี | `/gl/chartofaccounts` | เชื่อมหน้าจอและ API แล้ว |
| ยอดยกมาทางบัญชี | `/gl/openingbalance` | เชื่อมหน้าจอและ API แล้ว |
| กำหนดงบประมาณประจำปี | `/gl/budget` | เชื่อมหน้าจอและ API แล้ว |
| กลุ่มผังบัญชี | `/gl/account-groups` | เชื่อมหน้าจอและ API แล้ว |
| รูปแบบการเชื่อมโยงบัญชีอัตโนมัติ | `/gl/account-mapping` | เชื่อมหน้าจอและ API แล้ว |
| กลุ่มบัญชีสินค้า | `/gl/product-account-groups` | เชื่อมหน้าจอและ API แล้ว |
| ยอดสะสมประจำปี | `/gl/annual-balances` | เชื่อมหน้าจอและ API แล้ว |
| สมุดรายวันขาย | `/gl/journal/uv` | เชื่อมหน้าจอและ API แล้ว |
| สมุดรายวันซื้อ | `/gl/journal/sv` | เชื่อมหน้าจอและ API แล้ว |
| สมุดรายวันรับเงิน | `/gl/journal/rv` | เชื่อมหน้าจอและ API แล้ว |
| สมุดรายวันจ่ายเงิน | `/gl/journal/pv` | เชื่อมหน้าจอและ API แล้ว |
| สมุดรายวันทั่วไป | `/gl/journal/jv` | เชื่อมหน้าจอและ API แล้ว |
| กระดาษทำการ | `/gl/workingpaper` | เชื่อมหน้าจอและ API แล้ว |
| ล็อกงวดบัญชี | `/gl/periodlock` | เชื่อมหน้าจอและ API แล้ว |
| ปิดงบบัญชีสิ้นงวด | `/gl/financialclose` | เชื่อมหน้าจอและ API แล้ว |
| ตรวจสอบประจำวัน | `/checkdaily/dailyinfoscreen` | เชื่อมหน้าจอและ API แล้ว |
| ผ่านรายการบัญชี | `/gl/posting` | เชื่อมหน้าจอและ API แล้ว |
| ยกเลิกการผ่านรายการบัญชี | `/gl/unposting` | เชื่อมหน้าจอและ API แล้ว |
| ประมวลผลข้อมูลบัญชีใหม่ | `/gl/reprocess` | รอยอดและบริษัทของเอกสารต้นทาง |
| คำนวณยอดผ่านรายการใหม่ | `/gl/recalculate-posted` | เชื่อมหน้าจอและ API แล้ว |
| ประมวลผลสิ้นปี | `/gl/year-end` | เชื่อมหน้าจอและ API แล้ว |
| บัญชีแยกประเภท | `/report/ledger` | เชื่อมหน้าจอและ API แล้ว |
| งบทดลอง | `/report/trialbalance` | เชื่อมหน้าจอและ API แล้ว |
| งบกำไรขาดทุน | `/report/pnl` | เชื่อมหน้าจอและ API แล้ว |
| งบดุล | `/report/balancesheet` | เชื่อมหน้าจอและ API แล้ว |
| งบกระแสเงินสด | `/report/cashflow` | เชื่อมหน้าจอและ API แล้ว |
| ประมาณการกระแสเงินสด | `/report/cashflowforecast` | เชื่อมหน้าจอและ API แล้ว |
| กราฟประกอบงบการเงิน | `/report/financialgraphs` | เชื่อมหน้าจอและ API แล้ว |
| กำไรขาดทุนตามโครงการ | `/report/project-pnl` | เชื่อมหน้าจอและ API แล้ว |
| กำไรขาดทุนตามสาขาและแผนก | `/report/dimensionpnl` | เชื่อมหน้าจอและ API แล้ว |
| สรุปภาพรวมโครงการ | `/report/projectsummary` | เชื่อมหน้าจอและ API แล้ว |
| ภาพรวมธุรกิจ | `/report/dashboard` | เชื่อมหน้าจอและ API แล้ว |
| วิเคราะห์ธุรกิจสำหรับผู้บริหาร | `/report/executivesummary` | เชื่อมหน้าจอและ API แล้ว |
| ส่งออกงบการเงิน XBRL | `/report/xbrl` | เตรียมข้อมูล รอแม่แบบ DBD |
| สำรองและส่งออกข้อมูล | `/tools/databackup` | เชื่อมหน้าจอและ API แล้ว |
