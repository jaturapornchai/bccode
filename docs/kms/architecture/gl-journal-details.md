# หลักฐานลูกหนี้ เจ้าหนี้ และธนาคารในสมุดรายวัน GL

ตรวจโค้ด: 2026-09-20 — ผลการตรวจรับและ deploy แยกไว้ใน `docs/audits/GL-FIX-2026-09-20.md`.

## วัตถุประสงค์

ให้ห้องบัญชีบันทึกและตรวจหลักฐานในใบสำคัญรายวันเดียว ทั้งลูกหนี้ เจ้าหนี้ และธนาคาร โดยรับเอกสารจากภายนอกหรือกรอกเองได้ การตัดยอดคือการบันทึกผลรับจ่ายที่เกิดแล้ว ไม่ส่งคำสั่งโอนเงินและไม่สร้าง workflow หน้าบ้าน

## โครงสร้างและการทำงาน

- `Journal.details` รวม `partners`, `bank_accounts`, `documents`, `allocations`, `settlements`, `bank_lines`, `statement_lines`, `matches` และ `withdrawals`.
- `documents` คือเอกสารเพิ่ม/ลดหนี้ของคู่ค้า ฝั่ง `ar` หรือ `ap`. `allocations` เชื่อมหลายเอกสารกับหลายบรรทัด/หลายใบสำคัญ โดยลงเฉพาะส่วนบัญชีคุม.
- `settlements` เชื่อมบิลเพิ่มหนี้กับเอกสารลดหนี้/รับจ่าย หลายบิลและชำระบางส่วนได้ ตรวจยอดทั้งสองฝั่งใน transaction เดียว.
- `bank_lines` ระบุบัญชีธนาคารและทิศทางของบรรทัด GL. `statement_lines` มาจากหลักฐานธนาคารจริงที่ผู้ใช้ส่งเข้ามา; ไม่สร้าง Statement จาก GL. `matches` จัดสรรยอดหลายต่อหลายระหว่างสองฝั่ง.
- จำนวนเงินส่งเป็น decimal strings และคำนวณด้วย PostgreSQL NUMERIC/Go decimal. รายละเอียดต้องใช้สกุล THB และ scale ของปีบัญชี; ไม่ปัดยอดเกิน scale ให้เงียบ ๆ.
- Draft บันทึกรายวันและรายละเอียดด้วยคำสั่งเดียว. การผ่านรายการตรวจยอดจัดสรรครบเอกสารและไม่เกินบรรทัด GL. ยอดจัดสรรใน draft แยกจากยอดผ่านบัญชี; ไม่ถือเป็นยอดชำระยืนยันในรายงาน.
- Posted ใช้ `reconcile` พร้อม ID/version/reason เพื่อเพิ่มหลักฐานและตัดยอด/จับคู่. การเปลี่ยนคู่จัดสรรต้องถอนและสร้าง ID ใหม่ใน transaction เดียว พร้อมรักษายอดครบ. ยอดบัญชีเดิมแก้ไม่ได้.
- ถอนการตัดยอดหรือการจับคู่ด้วย `withdrawals` เก็บผู้ทำ เวลา และเหตุผล. เมื่อกระทบหลายใบสำคัญ จะเพิ่ม revision และให้ตรวจหลักฐานใหม่ทุกใบที่เกี่ยวข้อง.
- ก่อนกลับบัญชีต้องถอนการตัดยอดและจับคู่ที่ผูกอยู่; GL reversal รักษาสาขาและมิติของต้นฉบับ.

ฐานหลักฐาน `gl_subledger_*` แยกจาก cache `gl_records`/`gl_lines`; ไม่มี FK จากฐานหลักฐานไปยัง cache ที่ถูก rebuild. Internal FK, unique active pairs, finite positive NUMERIC และ immutable audit trigger ยังบังคับในฐานข้อมูล. Recalculate/reprocess replay ประวัติ GL ใน transaction และคงหลักฐาน/source registry เดิม.

## Workflow บนจอ

1. เปิดสมุดรายวันและเลือกใบสำคัญหรือสร้างใหม่ ใส่บัญชีเดบิต/เครดิตตามเอกสาร.
2. ในรายละเอียดด้านขวา เปิดหัวข้อหลักฐานลูกหนี้–เจ้าหนี้และธนาคาร เพิ่มคู่ค้า/บัญชีธนาคารเมื่อจำเป็น และเลือกเอกสารที่เกี่ยวข้อง.
3. จัดสรรยอดเอกสารลงบรรทัดบัญชีคุม และระบุธนาคารของบรรทัดเงินฝาก บันทึกพร้อมใบสำคัญ.
4. เมื่อได้ Statement ให้กรอกหรือนำเข้า CSV แล้วเลือกจับคู่ตามธนาคารและทิศทาง ระบบตรวจยอดที่ยังเหลือทั้งสองฝั่ง.
5. ดูเอกสารค้าง/Statement คงเหลือ/บรรทัดธนาคารรอจับคู่จากภายในหน้ารายวัน ใช้ค้นหาและเปลี่ยนหน้าได้.
6. หลังผ่านบัญชี กดบันทึกผลกระทบยอดเพื่อส่งเฉพาะหลักฐานใหม่ ใช้ผลตรวจแยกจากสถานะผ่านรายการ.

CSV ต้องมี `transaction_date,direction,amount`; วันที่เป็น YYYY-MM-DD, direction 1=เข้า/2=ออก, amount เป็นข้อความทศนิยม. เลือกธนาคารก่อนนำเข้า. รองรับคอลัมน์ `value_date,bank_reference,description,balance_after`. จำกัดไฟล์ 2 MB / 1,000 แถว. ใช้ hash ของไฟล์และเลขแถวเป็น source key; นำไฟล์เดิมซ้ำไม่เพิ่มรายการเดิม.

## API และ MCP

API และ MCP ใช้ token คนละชนิด ผูก Holding และ allowlist บริษัท. Readonly เลือกหลายบริษัทได้; readwrite ทุกคำขอเลือกบริษัทเดียว. Backend ตรวจ active user/member/Holding/company/role และ scope ล่าสุด ก่อนใช้ GL service เดียวกัน.

งาน `processes` (close, year-end, recalculate, reprocess) ทำงานทั้งบริษัท: backend ตรวจสิทธิ์เมนู/การกระทำก่อนใช้ `CompanyWide` ซึ่งคำนวณจาก membership grant ล่าสุดแยกจากสาขาที่ผู้ใช้เลือก. เมื่อมี grant ระดับบริษัทจึงล้าง branch สำหรับ process; การบันทึกและอ่านใบสำคัญตามปกติยังใช้สาขาที่เลือก. หน้า preview ส่ง `companywide=true` ให้รายงานใช้ขอบเขตเดียวกับ process โดยยังต้องมีสิทธิ์อ่านรายงาน. ผู้มีสิทธิ์เฉพาะสาขาถูกปฏิเสธทั้ง process และ preview ทั้งบริษัท.

สำหรับ API/MCP token ค่า CompanyWide อนุญาตเฉพาะ token ที่ไม่มี stored branch restriction และผ่าน company allowlist แล้ว. Legacy token ที่ยังเก็บสาขาจะไม่ถูกขยายเป็นทั้งบริษัท แม้ผู้ออก token เป็น OWNER/ADMIN. การใช้ `companywide=true` ไม่ข้ามข้อบังคับ readonly/readwrite หรือการเลือกบริษัทเดียวต่อคำขอเขียน.

| งาน | HTTP ภายใน | MCP |
|---|---|---|
| อ่านรายการหลักฐาน | `GET /gl/v2/journal-support?kind=documents&page=1&limit=50` | `gl_list`, resource=`journal-support`, query.kind=`documents` |
| อ่านยอดค้าง | `GET /gl/v2/reports/ar-outstanding?to=2026-12-31` | `gl_report`, report=`ar-outstanding` |
| อ่านเจ้าหนี้ค้าง | `GET /gl/v2/reports/ap-outstanding?to=2026-12-31` | `gl_report`, report=`ap-outstanding` |
| Statement ยังไม่จับคู่ | `GET /gl/v2/reports/bank-unmatched?to=2026-12-31` | `gl_report`, report=`bank-unmatched` |
| สร้าง/แก้/กระทบยอด | `POST /gl/v2/command` | `gl_command`, command ใช้ contract เดียวกัน |

API token เข้าทาง `/integration/gl/v2` (ผ่านเว็บ `/api/integration/gl`) พร้อม `X-BC-Company-Code` หรือ readonly `X-BC-Company-Codes`. MCP endpoint `/mcp/gl` และชื่อบริษัทอยู่ใน arguments. Support kinds: `partners`, `bank-accounts`, `documents`, `allocations`, `settlements`, `statements`, `bank-lines`, `matches`; รองรับ `q,page,limit,asof`. ค่าเงินในผลลัพธ์เป็นข้อความด้วย. รายงาน AR/AP ใช้ยอดสุทธิ: เพิ่มหนี้เป็นบวก ลดหนี้เป็นลบ; รายการ `journal-support` แสดงยอดบวกแยกแต่ละเอกสารสำหรับเลือกตัดยอด.

ตัวอย่าง arguments ของ MCP ที่อ่านเอกสารค้างจากบริษัทที่ token อนุญาต:

```json
{"companyCodes":["01","02"],"resource":"journal-support","query":{"kind":"documents","page":"1","limit":"50"}}
```

ตัวอย่าง payload เพื่อเพิ่มการจับคู่ในใบผ่านบัญชีแล้ว (แทน ID/version ด้วยค่าที่อ่านจากระบบ):

```json
{
  "resource":"journals","action":"reconcile",
  "requestid":"b4b56e18-cf67-4c19-9a28-a4e1ae77a5c5",
  "id":"journal-id","version":3,"reason":"ตรวจหลักฐานธนาคารแล้ว",
  "journal":{"details":{"matches":[
    {"id":"match-id","statement_line_id":"statement-id","journal_id":"journal-id","line_number":1,"amount":"600.00"}
  ]}}
}
```

ทุกการกระทำใหม่ใช้ requestid ใหม่; retry คำขอเดิมใช้ ID เดิม. การนำเข้าใบสำคัญใช้ `source_system`+`source_record_id` เป็นตัวตนคงที่ภายในบริษัท: เปลี่ยน requestid ยังได้ใบเดิมเมื่อ payload เหมือนเดิม; payload ต่างถูกปฏิเสธและต้องเปิดแก้ใบเดิมด้วย version. Source identity เปลี่ยนหลังสร้างไม่ได้.

## Config และ dependency

- ใช้ PostgreSQL กลางสำหรับสิทธิ์และ PostgreSQL ของ Holding สำหรับ GL ตามการตั้งค่าปัจจุบัน; ไม่มีบริการใหม่ ไม่มี environment flag ใหม่.
- Schema ถูก ensure ผ่าน PostgresStore ในฐาน Holding. ต้องมีสิทธิ์สร้างตาราง/index/trigger ตอนเริ่มใช้งานรุ่นนี้.
- HTTP Go :8888, frontend BFF Next.js และ MCP gateway เดิม; ไม่ใช้ MongoDB, ClickHouse, Kafka หรือ Redis ในเส้นทางนี้.
- ทดสอบแยกด้วย `BC_GL_TEST_POSTGRES_DSN`; tests สิทธิ์ใช้ `GL_AUTH_TEST_DSN`. ห้ามชี้ env ทดสอบไป production.

## ข้อจำกัดและแนวทางตรวจรับ

- รองรับ THB; ไม่ทำ FX, ไม่ net ลูกหนี้กับเจ้าหนี้, ไม่ส่งเงิน, ไม่ส่ง RD Prep/XBRL.
- จำกัดรายละเอียดรวม 2,000 แถวต่อคำสั่งและ body 2 MB; อ่านครั้งละไม่เกิน 1,000 แถว มี pagination และยอดรวมคำนวณจากฐานจริง.
- บริษัทเก่าที่ไม่มี `details` อ่านและทำงานตามเดิม; ระบบไม่เดาหรือสร้างคู่ค้า/เอกสาร/Statement ย้อนหลังจากคำอธิบายของใบสำคัญ.
- รายงานย้อนหลังต้องพิจารณาทั้งวันที่บัญชีและเวลาเกิด/ถอนหลักฐาน; ตรวจ cases บันทึกย้อนหลังและกลับรายการข้ามงวดกับหลักฐานจริงก่อนใช้อ้างอิงงบ.
- เงินตัวอย่างยืนยัน 0.1+0.2, scale boundaries, debit=credit, duplicate/concurrent request, rollback เมื่อเกินยอด, หลายบิล/หลายใบสำคัญ, source/Statement dedup และการ rebuild ที่รักษาหลักฐาน.

## แหล่งโค้ด

- `backend/internal/generalledger/subledger_types.go` — JSON contract.
- `backend/internal/generalledger/subledger.sql` — ตาราง/ข้อบังคับหลักฐาน.
- `backend/internal/generalledger/subledger.go`, `subledger_allocations.go`, `subledger_bank.go`, `subledger_reconcile.go` — atomic mutation และการถอน.
- `backend/internal/generalledger/subledger_queries.go` — ยอดค้างและขอบเขตการอ่าน.
- `backend/internal/generalledger/postgres_source.go` — stable source dedup.
- `backend/internal/generalledger/httpapi/subledger.go` — HTTP permission/snapshot.
- `backend/internal/generalledger/httpapi/scope.go`, `http.go`, `company_scope_test.go` — grant ระดับบริษัท, process/report preview และ legacy token branch guard.
- `frontend/src/app/gl/gl-journal-details.tsx`, `frontend/src/lib/gl-journal-details.ts` — หน้ารายวันและ CSV.
