# ผลตรวจและข้อแตกต่างในหน้ารายวัน

## วัตถุประสงค์

ให้ห้องบัญชีบันทึกผลตรวจเอกสารและข้อแตกต่างได้ในหน้ารายวัน โดยแยกจากสถานะผ่านรายการ ไม่ต้องสร้าง workflow อนุมัติหลายแผนก

## Workflow

1. เปิดใบสำคัญที่บันทึกแล้วในโหมดแสดงข้อมูล เลื่อนไปส่วนผลตรวจและข้อแตกต่าง
2. กดบันทึกผลตรวจ เลือกรอตรวจ/พบข้อแตกต่าง/ตรวจแล้ว หากพบข้อแตกต่างต้องกรอกหมายเหตุ
3. บันทึกผ่าน API พร้อม version ใบสำคัญและหมายเลขผลตรวจล่าสุด หากข้อมูลถูกแก้พร้อมกันให้โหลดเอกสารใหม่ หมายเหตุที่ยังบันทึกไม่ได้ยังแสดงให้คัดลอกได้
4. เปิดประวัติเพื่อดูผู้ตรวจ เวลา และรุ่นเอกสาร การเปลี่ยน version ใบสำคัญทำให้ผลตรวจปัจจุบันกลับเป็นรอตรวจ แต่ประวัติเก่ายังอยู่

## Config และ dependencies

ไม่เพิ่ม env หรือ service ใหม่ ใช้การเชื่อม PostgreSQL และ auth/session ของ GL เดิม ตาราง runtime คือ `gl_journal_review_events` ใน `backend/internal/generalledger/schema.sql` สร้างผ่านตัวโหลด schema เดิมเมื่อ backend เริ่มใช้ฐานนั้น ข้อมูลบัญชีจริงยังอยู่ใน gl_records/gl_journal_lines ไม่ใช่ journal_entries ในร่าง mydocs

Frontend ใช้ GLJournalReviewPanel, useGLCommand และ BFF `/api/gl`; Backend ใช้ Go sql.Tx, company lock และ idempotency เดิม ไม่มี MongoDB, ClickHouse, Kafka หรือ Redis เพิ่มในงานนี้

## API example

อ่าน `GET /gl/v2/journal-reviews/{journal-id}` หลังยืนยันตัวตนและเลือกบริษัท คืน journalid, version, status, eventno และ events

ส่ง payload ต่อไปนี้ไป `POST /gl/v2/command` ด้วย session ที่มีสิทธิ์แก้ไขสมุดรายวัน เปลี่ยน id/version/expectedEventNo ให้ตรงผลอ่านล่าสุด และใช้ requestid ใหม่ต่อการทำรายการใหม่:

```json
{
  "resource": "journals",
  "action": "review",
  "id": "journal-id-from-get",
  "version": 4,
  "requestid": "2dc52000-603a-40bf-9800-374d73cfd6ce",
  "review": { "status": 2, "note": "ยอดธนาคารไม่ตรง", "expectedEventNo": 0 }
}
```

Backend ใช้ตัวตนผู้ตรวจจาก session ไม่รับตัวตนจาก payload การตรวจไม่แก้ยอดบัญชีหรือ journal version; trigger ป้องกัน UPDATE/DELETE/TRUNCATE ของประวัติ

## ข้อจำกัด

- แสดงประวัติล่าสุด 100 รายการ ไม่มี pagination ประวัติในรอบนี้
- การผ่านรายการที่เพิ่ม version ทำให้ต้องตรวจใหม่เช่นเดียวกับการแก้ไข
- อ่านผลตรวจผ่าน `gl_get` resource `journal-reviews` และบันทึกผ่าน `gl_command` action `review` ได้ด้วย MCP token แบบ readwrite; API token เป็นคนละประเภท ดู `gl-mcp-tokens.md`
- หน้าจอจำกัดหมายเหตุ 2,000 ตัวอักษร Backend รองรับ 4,000
- การยืนยันสิทธิ์ใช้ fail-closed: ถ้าเชื่อม DB ไม่ได้ สมาชิกไม่ active/ไม่พบ หรืออ่าน/แปลงสิทธิ์ไม่สำเร็จ คืน HTTP 403 พร้อมสิทธิ์ว่าง ไม่ใช้สิทธิ์ที่อ่านได้เพียงบางส่วน OWNER/ADMIN ได้ wildcard เฉพาะเมื่อพบสมาชิก active แล้วเท่านั้น

## Verification

- Frontend typecheck/build ผ่าน; Vitest ทั้งชุด 685 tests ผ่าน รวมชุด BFF 6 tests
- Backend `go test ./internal/generalledger/...` และ `go build ./...` ผ่าน
- ทดสอบสิทธิ์แบบ fail-closed ผ่าน 8 กรณี พร้อม OWNER/ADMIN/STAFF ที่มีสิทธิ์ถูกต้อง
- PostgreSQL integration ผ่าน lifecycle, retry, payload mismatch, concurrent reviewer, stale version, company/branch isolation และ append-only
- Playwright ใช้หน้าจอจริงกับ HTTP fixtures ตรวจหมายเหตุบังคับ ประวัติ ป้องกันทิ้งข้อมูล โหลด version ใหม่ และคงข้อความเมื่อบันทึกชนกันผ่าน ตรวจภาพ light/dark แล้ว
- Production รุ่น `r20260920-gl-review-deny-1`: สำรองฐานข้อมูล/config ก่อนสลับ ทั้งสองบริการ healthy หน้าเว็บคืน HTML 200 และ GL API ไม่มี token คืน JSON 401; ยังไม่ได้ทดสอบ session ผู้ใช้จริงบน production
- ถอด MongoDB startup initializer, การสร้าง index อัตโนมัติของสินค้า/บาร์โค้ด/คลัง, Mongo-to-Kafka outbox workers และ coupon cleanup scheduler แล้ว เส้นทาง GL ใช้ PostgreSQL เช่นเดิม; legacy API ที่อ้าง MongoDB ยังอยู่ จึงไม่ใช่การย้ายฐานข้อมูลทั้งระบบ
