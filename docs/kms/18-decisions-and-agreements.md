# ข้อตกลงและการตัดสินใจของโปรเจ็กต์ (ที่ไม่ได้อยู่ในโค้ด)
> ตรวจล่าสุด: 2026-09-09 (HEAD `098107e1`) — กฎที่ "บังคับใช้" อยู่ใน `AGENTS.md` (ไฟล์นั้นชนะเสมอ); ไฟล์นี้เก็บ**ที่มา/วันที่/เหตุผล**และการตัดสินใจอื่นที่ AI ทุกตัวต้องรู้ตรงกัน; ADR ฉบับเต็มอยู่ใน `docs/kms/decisions/`

## ลำดับการอ่านสำหรับ AI ทุกตัว (On-Demand ประหยัด Context)

1. `AGENTS.md` (กฎบังคับหลัก)
2. **โหลดเฉพาะเรื่องที่ต้องใช้จริง (On-Demand Loading เพื่อไม่เปลือง Context Window)**:
   - งานทั่วไป / เล็กน้อย / แก้บั๊ก: ตรวจและแก้ที่โค้ดจริงโดยตรง (`code = truth`) ไม่ต้องเปิดอ่าน docs
   - งาน UX/UI: อ่านเฉพาะ `docs/skills/ui-scale-polish/SKILL.md`
   - งาน Schema / MongoModel: อ่านเฉพาะ `docs/skills/audit-mongomodel-sync/SKILL.md`
   - งานสถาปัตยกรรม / โดเมนเฉพาะ: เปิดสารบัญ `docs/kms/README.md` แล้วเลือกอ่านเฉพาะ **1 บทความที่ตรงกับเรื่อง**
   - สถานะงานค้าง / Handoff: เปิด `docs/handoff/HANDOFF-*.md` เฉพาะเมื่อลุงจืดสั่งหรือถามความเสี่ยง
   - แผนที่ไฟล์ใหญ่: ดู `docs/reference/CODE-MAP.md` เฉพาะเมื่อต้องแตะไฟล์ขนาดยักษ์ (>1,000 บรรทัด)

## การตัดสินใจหลัก (เรียงตามวันที่)

| วันที่ | การตัดสินใจ | เหตุผล / ผล |
|---|---|---|
| 2026-06-21 | **ชื่อ collection/field ใน Mongo = lowercase (อนุญาต underscore)** | ชื่อ camelCase ทำให้เกิด collection ผี (MongoDB case-sensitive); พบ 7 จุดใน goapi ตอน audit |
| 2026-06-21 | Holding admin จัดการด้วย email ผ่าน `shopusers` + Role (0 user / 1 admin / 2 owner); **owner+admin = สิทธิ์เต็ม**, owner ถูกปกป้อง | ไม่ใช้ collection ใน `admin-access-control.md` ที่ยังไม่ implement; **สถานะโค้ด 2026-09-09: เหลือ route เดียวคือ `/holding-member/list`** (`backend/internal/shop/shopuser_http.go:63`) ที่ whitelist ใน `exceptShopPath` (`backend/main.go:319`) ส่วน `/holding-member/add` และ `/holding-member/remove` ถูกเลิก register โดยตั้งใจ (`shopuser_http.go:64-65` — ต้องผ่าน invitation lifecycle) แม้ handler `AddHoldingMemberAdmin`/`RemoveHoldingMemberAdmin` ยังอยู่ และ frontend proxy ยังยิงไป 2 route นั้น (`frontend/src/app/api/holding-member/route.ts:21,29`) → ถ้าจะแก้ ต้องถามลุงจืดก่อนว่าจะทำ invitation flow หรือ register route กลับ |
| 2026-06-21 | `/googlelogin` ต้องมี Google ID token จริง (ไม่รับ email เปล่า) | security fix; test token ต้องมาจาก GIS จริง หรือใช้ demo/dev login |
| 2026-06-23 | **Checkpoint ก่อนงานใหญ่/เสี่ยง**: commit + `git push origin dev` ก่อนเริ่ม | มี restore point ("ป้องกันพัง"); งาน 1–2 บรรทัดไม่ต้อง; ห้าม push secret |
| 2026-06-28 | **ข้อมูลทุก env (รวม prod) disposable จน go-live** — ไม่ทำ migration/backfill, เปลี่ยน schema/ชื่อ topic/field ได้เลย; **ไม่ต้องขออนุญาต**เปลี่ยนโครง Mongo / จอ / field / โค้ด (ขยาย 2026-07-01) | pre-launch ไม่มีข้อมูลจริง; **หยุดกฎนี้ทันทีเมื่อมีข้อมูลลูกค้าจริง** (R0) — ถ้าไม่แน่ใจว่ามี ให้ถาม |
| 2026-06-28 | DEV = local ทั้งหมด, DEPLOY = .202 ทั้งหมด (2 ระบบแยก) | เร็ว + isolate |
| 2026-06-28 | viewport เป้าหมาย = **iPad ≥768px ขึ้นไป**; มือถือแค่ไม่พัง | ERP ใช้บนแท็บเล็ต/เดสก์ท็อป |
| 2026-06-29 | **ไม่มีพิธี migration ใน dev** — schema ผิด/ขาด = แก้ backend ให้สร้างของถูก แล้ว rebuild ("เอาตัวใหม่ที่ถูกต้องเลย") | ห้าม `DEV_API_MODE=3`/DDL มือเพื่อปะ binary เก่า |
| 2026-07-01 | ไม่ใช้ Magnitude เป็น browser agent; ใช้ Stagehand+DeepSeek | ต้องการ vision LLM ที่ไม่มี |
| 2026-07-09 | จบงานทุกครั้งต้องเช็คว่า `localhost:3000` ยังรัน | dev server เคยตายเองบ่อย |
| 2026-08-30 | **UX/UI ยึด "คนไทยอายุ 40+"** (ตัวหนังสือใหญ่, ไทยก่อน, ปุ่มใหญ่, contrast, ยืนยันก่อนทำลาย, feedback ทุก action) | ผู้ใช้หลัก = พนักงานบัญชี/เจ้าของกิจการ (รายละเอียด AGENTS.md) |
| 2026-08-30 | UAT ต้อง CRUD ครบ + ยืนยันใน MongoDB **ทีละ step** + ล้างด้วย id + seeded random | เคยเจอ API success แต่ DB มี orphan; เคยลบสาขาจริงด้วย regex |
| 2026-08-31 | รูปภาพห้ามเก็บใน Mongo — S3/MinIO เท่านั้น + **thumbnail คู่เสมอ** (`<field>thumb`) | list ใช้ thumb เป็นหลัก |
| 2026-09-02 | **UX/UI พรีเมี่ยม + ใช้ง่าย + เหมาะกับคนไทย ทั้งระบบ** (สีจาก `--primary` ผ่าน color-mix, light/dark เท่ากัน, พื้นผิวมีชั้น, motion น้อย, screenshot ตรวจรับ) + **อัปเดต skill ทุกครั้งที่แตะ UI** | หน้า login/holding = มาตรฐานอ้างอิง |
| 2026-09-02 | ผู้ช่วยคิด = Kimi K3 + GLM; Claude เป็นหัวหน้าและ verify ทุกอย่างเอง; DeepSeek/ChatGPT/OpenRouter ยังไม่เปิด | ประหยัด token; คำตอบผู้ช่วยไม่ใช่หลักฐาน |
| 2026-09-03 | Marketplace/listing model (`/goapi/product/v2/*`) ใช้รูปแบบตาม open platform ของตลาดออนไลน์รายใหญ่ แต่**ห้ามเอ่ยชื่อเจ้านั้นในโค้ด/เอกสาร** — ใช้คำกลาง ("marketplace", "ช่องทางขายออนไลน์", "ตัวเลือกสินค้า") | branding/IP hygiene; field บัญชี (unitcode, vattype, itemtype, cost) เป็นของเรา |
| 2026-09-03 | ลบ `docs/` เดิมทั้งหมดเพื่อออกแบบใหม่ → ไม่มี business-rule SoT จนกว่าจะเขียนใหม่; **requirement ไม่ชัด = ถาม ห้ามเดา** | ของเก่าอยู่ใน git history |
| 2026-09-05 | Product/Barcode/Unit ใช้ **transactional outbox** (`outboxevents`) + consumer แบบ ack-after-success + fences ใน PG; **ห้าม purge history ของ outbox** | แก้ Mongo/PG divergence จาก fire-and-forget |
| 2026-09-06 | **พัก ClickHouse บนเครื่อง dev**; ถอดถาวรรอตัดสิน | ไม่มีบทบาทจริง กิน RAM 5 GB (ADR `decisions/2026-09-06-pause-clickhouse-local.md`) |
| 2026-09-07 | **skill ส่วนตัวอยู่ที่ `docs/skills/`, ฐานความรู้อยู่ที่ `docs/kms/`, ทุกอย่างรวมใน `docs/`** เพื่อให้ AI หลายตัวเข้าใจตรงกัน | ลุงจืดใช้ AI หลายตัว; memory ส่วนตัวของ AI ตัวเดียวคนอื่นมองไม่เห็น |
| 2026-09-07 | **ปรับการอ่าน docs เป็น On-Demand (Lazy Loading) ไม่เปลือง context** | ห้ามโหลดเอกสารทั้งโฟลเดอร์หรือ handoff ล่วงหน้า; เปิดอ่านเฉพาะไฟล์ที่ตรงกับงานจริงเพื่อประหยัด Context Window |
| 2026-09-07 | **บังคับใช้กฎความเร็วสูงสุดและสุขอนามัย Context (Surgical Read/Patch/Terminal/Subagent)** | อ่านและแก้เฉพาะบรรทัด, ห้ามรัน full test suite โดยไม่จำเป็น, คุม output terminal, ใช้ subagent กัก context บวม, รักษา prompt cache |
| 2026-09-07 | **สถาปัตยกรรม 2-Tier: MongoDB เก็บย่อ (Storage) + PostgreSQL ประมวลผลเร็วแบบครบจบ (Processing Engine)** | ข้อมูลใน Mongo ต้องโคลนไปสร้างใน PG ทั้งหมด; Mongo เก็บแบบประหยัดขนาด; PG มีรายละเอียดครบถ้วนเพื่อประมวลผลจบในตัว ไม่ต้องต่อกลับมา Mongo อีก; ตอบคำถามค้างเดิมเรื่อง "PostgreSQL มีไว้เพื่ออะไร" (ถอดออกจากรายการแล้ว — **ไม่ใช่ข้อ 6 ในรายการปัจจุบัน**) (ADR `decisions/2026-09-07-mongodb-storage-postgres-processing-clone.md`) |
| 2026-09-08 | **เพิ่มเมนู 15 รายการให้เทียบเท่า FlowAccount + PEAK (185 → 200 เมนู)** และ **ไม่ทำระบบเงินเดือน** (ตัด payroll, สลิปเงินเดือน, **ภ.ง.ด.1 / ภ.ง.ด.1ก**, ไฟล์นำส่งเงินสมทบประกันสังคม (สปส. / กท.20 ก) ออกจากขอบเขต — ห้ามเพิ่มกลับเอง; กฎบังคับฉบับเต็มอยู่ใน `AGENTS.md` หัวข้อ "ขอบเขตผลิตภัณฑ์ — ไม่ทำระบบเงินเดือน") — แต่ **ภ.ง.ด.2 อยู่ในขอบเขต** เพราะเป็นภาษีหัก ณ ที่จ่ายของเงินได้ 40(3)/(4) (ดอกเบี้ย เงินปันผล ค่าสิทธิ) ต้นทางคือรายการจ่ายเงิน ไม่ใช่เงินเดือน จึงมีเมนู `/report/vatpnd2` ตั้งแต่ 2026-09-08 **ห้ามลบเพราะเข้าใจผิดว่าเป็นแบบยื่นของเงินเดือน**; ที่ยังอยู่ในขอบเขตตามปกติ: ภ.ง.ด.3/53 + 50 ทวิ ของคู่ค้า, เงินทดรองจ่ายพนักงาน, ทะเบียนพนักงาน `/employee` | เพิ่มเฉพาะรายการเมนู ยังไม่สร้างจอ/กฎธุรกิจ; รายการที่กฎยังไม่ชัด (XBRL ไฟล์ธนาคาร ระดับอนุมัติ เทมเพลตนำเข้าเอกสาร) เปิดเมนูไว้ก่อนและรอลุงจืดกำหนดกฎ (ADR decisions/2026-09-08-menu-parity-flowaccount-peak.md, รายละเอียด 19-menu-coverage-flowaccount-peak.md) |
| 2026-09-08 | **ปิดช่องว่างที่เหลือจากหลักฐานหน้าเว็บทางการ (200 → 218 เมนู)** — ล็อต/วันหมดอายุ + รายงานล็อต/สินค้าใกล้หมดอายุ, ใบรวมจ่าย/ใบเสร็จรวม, บัตรเครดิตกิจการ, เงินทดรองจ่ายกรรมการ, ซื้อสินทรัพย์/งานระหว่างก่อสร้าง/ตารางค่าเสื่อม, ภาษีเงินได้รอการตัดบัญชี, กลุ่ม "บุคลากรและผู้ใช้งาน" (`/employee` เข้าเมนูหลักแล้ว), เชื่อมต่อ LINE (`/line-oa`), สำรอง-ส่งออกข้อมูล, ส่งรายงานอัตโนมัติ | commit `3217fd04` — **ไม่มี ADR แยก บันทึกไว้ที่แถวนี้** รายละเอียดใน `19-menu-coverage-flowaccount-peak.md`; รอบนี้ตอบคำถามค้างเดิมข้อ 7 (ล็อตสินค้า/วันหมดอายุ, รายงานส่งอีเมลอัตโนมัติ, ส่งออก-สำรองข้อมูล) ว่า **ทำทั้งหมด** — เหลือเฉพาะกฎธุรกิจของแต่ละเมนู ซึ่งย้ายไปอยู่ในคำถามข้อ 6 |
| 2026-09-08 | **กวาดแหล่งนอกเว็บทางการเพิ่มอีก 6 เมนู + กลุ่มใหม่ "เชื่อมข้อมูลตลาดออนไลน์" (218 → 224 เมนู)** — ภ.ง.ด.2 (`/report/vatpnd2`), กำไรขาดทุนตามสาขาและแผนก (`/report/dimensionpnl`), ดึงคำสั่งซื้อจากร้านค้าออนไลน์ (`/transaction/marketplaceorder`), เชื่อม Shopee / Lazada / TikTok | commit `098107e1` (ADR `decisions/2026-09-08-menu-parity-social-sweep.md`); **จำนวนเมนูปัจจุบัน = 224** ต้องตรงกับ `frontend/src/lib/menu-icons.test.ts:10` (`toHaveLength(224)`) เสมอ; **ห้ามเขียนว่า "ครบ 100% เทียบ FlowAccount/PEAK" อีก** ให้เขียนว่า "ครบเท่าที่หลักฐานชุดนี้ครอบคลุม"; 3 จอ marketplace (`frontend/src/app/menu/marketplace-screen.tsx`) เขียนเสร็จอยู่ก่อนแล้วและ **เคย** เป็นจอกำพร้าเพราะ `onOpenRoute` เปิดได้เฉพาะ route ที่มีในเมนู (`frontend/src/app/menu/main-menu-screen.tsx:1453-1455`) — รอบนี้ต่อสายเมนูแล้วจึงเข้าถึงได้ ไม่ต้องแก้ `onOpenRoute` อีก |

## คำถามที่ยังไม่มีคำตอบ (ห้ามเดา — ถามลุงจืด)

1. currency / precision / rounding ของตัวเลขบัญชี
2. business contract ของ 15 package ใน `backend/.ci/test-quarantine.txt`
3. backup platform / RPO / RTO ของ prod
4. Unit contract: `unit_of_measure/code/businesscode` vs `units/unitcode` (MongoModel partial index)
5. ถอด ClickHouse ถาวรหรือพักต่อ
6. taxonomy XBRL ของ DBD, รูปแบบไฟล์โอนเงินธนาคาร, ระดับการอนุมัติรายจ่าย, เทมเพลตนำเข้าเอกสาร, **อัตราหัก/กำหนดยื่น ภ.ง.ด.2** (เงินได้ 40(3)/(4)), **มิติใดบังคับกรอกบนเอกสาร** สำหรับกำไรขาดทุนตามสาขาและแผนก (`/report/dimensionpnl`), **วิธีตัดสต็อกตามล็อต/วันหมดอายุ**, **รอบและช่องทางส่งรายงานอัตโนมัติ**, **รูปแบบ/ปลายทางไฟล์สำรองข้อมูล** — เมนูเปิดไว้แล้ว 2026-09-08 แต่ยังไม่มีกฎ ตอนทำจอต้องตรวจกับแหล่งทางการก่อนตาม Rule 17 (เงินเดือน/ภ.ง.ด.1/ภ.ง.ด.1ก/ประกันสังคม ตัดออกจากขอบเขตแล้ว ไม่ต้องถาม)
