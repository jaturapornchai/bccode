---
date: 2026-09-24
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, go, postgres, gl, permissions]
---

# สมุดรายวันเป็นข้อมูลหลักที่ผู้ใช้กำหนดเอง (มีประเภทสมุด 1–6) — โค้ดไม่ยึดรหัส JV/PV/RV/SV/UV

## Context

- ลุงจืดสั่ง 2026-09-24: "สมุดรายวันไม่ตายตัว" — ผู้ใช้เพิ่ม ลบ แก้สมุดได้ตลอด และตั้งรหัสเป็นภาษาไทยได้ (เช่น `สมุดซื้อ`)
- โค้ดเดิมมีรายการสมุดตายตัว (JV/PV/RV/SV/UV) ทั้งฝั่ง backend และจอ และหลายจุดตัดสินใจจากรหัส: การปิดบัญชี/สินทรัพย์ถาวรใช้ `"JV"`, จอเลือก VAT ขาย/ซื้อจากรหัสสมุด, สิทธิ์ผูกกับรหัสสมุด (`jv-journal`, `sv-journal` …) — บริษัทที่ตั้งรหัสเองจึงใช้ไม่ได้
- ความหมาย SV/UV ในระบบเดิม **สลับกัน**; มาตรฐานโปรแกรมบัญชีไทยทั่วไปใช้ **SV = สมุดรายวันขาย, UV = สมุดรายวันซื้อ**
- ต้นแบบ Champ: ตาราง `BCGLBook` (`D:\project-champ\champ\champ\Script\SQLSERVER_Script.sql`) เป็น master ที่ผู้ใช้กำหนดเอง — `Code VARCHAR(15)`, `Name`, `TitleName`, `PrintSlip`, `FomFileName`; ไม่มีประเภทสมุด
- สเปกของลุงจืด `mydocs/datamodels/gl/journalbook.sql`: `code VARCHAR(15)`, `name_th VARCHAR(100) NOT NULL`, `name_en VARCHAR(100)`, `book_type SMALLINT CHECK (1..6)` (1 ทั่วไป, 2 จ่าย, 3 รับ, 4 ขาย, 5 ซื้อ, 6 ยอดยกมา), `is_active`
- regexp รหัสเดิม `^[\p{L}\p{N}]...` ไม่รับสระ/วรรณยุกต์ไทย (เป็น combining mark `\p{M}`) — ดู `docs/kms/17-dev-gotchas.md` §กับดัก regex ภาษาไทย

## Decision

1. **Master `journal-books`** = `{code, name, nameen, booktype, isactive}` ตามสเปก: รหัส 1–15 ตัวอักษร (นับ rune, NFC + trim ที่ server), ชื่อไทยบังคับ ≤ 100, ชื่ออังกฤษ ≤ 100, `booktype` 1–6 บังคับตอนสร้าง (`backend/internal/generalledger/journal_books.go:18-28`, `validateJournalBookMaster` :110)
2. **โค้ดตัดสินใจจาก `booktype` เท่านั้น ไม่ดูรหัส**:
   - ปิดบัญชีและสินทรัพย์ถาวรใช้สมุดประเภท 1 ที่เปิดใช้งาน (รหัสน้อยสุด); ยกยอดข้ามปีใช้ประเภท 6 ก่อน ไม่มีจึงใช้ 1 (`processBookCode` :251, `ChooseBookCode` :210); ไม่มีเลย → `journal_book_general_missing` บอกให้ไปกำหนดสมุด
   - จอ: ค่าเริ่มต้นประเภท VAT = ขายเมื่อ `booktype` 4 (`defaultVatTaxType` `frontend/src/lib/gl-journal-details.ts:50`); สมุดเริ่มต้นของใบใหม่ = สมุดที่เลือก → ประเภท 1 (`defaultJournalBookCode` `frontend/src/lib/general-ledger.ts:107`)
3. **บันทึกเอกสาร**: สมุดต้องมีอยู่ ไม่ถูกลบ และมีประเภท (`journal_book_not_found` / `journal_book_type_missing`); เอกสารใหม่และเอกสารที่ระบบสร้างต้องใช้สมุดที่เปิดใช้งาน (`journal_book_inactive`); แก้/ผ่าน/กลับรายการใบเดิมไม่บังคับ active (`checkJournalBook` :185); สมุด ประเภทรายการ และเลขที่ของใบที่บันทึกแล้ว **เปลี่ยนไม่ได้** (`journal_book_immutable` / `journal_kind_immutable` / `journal_docno_immutable`) — จอล็อกช่องเลือกสมุดของใบที่บันทึกแล้ว
4. **ลบ vs ปิดใช้งาน**: สมุดที่มีใบสำคัญ **หรือรูปแบบการเชื่อมบัญชี (`mappings.bookcode`)** ที่ไม่ถูกลบอ้างถึง = "มีการใช้งาน" (`journalBookInUse` :132 — ตามสเปก FK `ON DELETE RESTRICT`) → **ลบไม่ได้** ให้ปิดใช้งานแทน (`journal_book_in_use_delete`), เปลี่ยนรหัส/ประเภทไม่ได้ (`journal_book_in_use_code` / `_type`) — ข้อความบอกทั้งสองสาเหตุ ("มีเอกสารบันทึกอยู่แล้ว หรือมีรูปแบบการเชื่อมบัญชีเลือกสมุดนี้อยู่"); แก้ชื่อ ชื่ออังกฤษ และเปิด/ปิดใช้งานได้เสมอ; สมุดเก่าที่ยังไม่มีประเภทกำหนดประเภทได้แม้มีเอกสารแล้ว และ **เปลี่ยนชื่อหรือปิดใช้งานได้โดยไม่ต้องเลือกประเภท** (`validateJournalBookMaster(…, allowUntyped)` :110, `guardJournalBookUpdate` :140, `guardJournalBookDelete` :154)
5. **สมุดเริ่มต้น 5 เล่ม** JV(1) PV(2) RV(3) SV(4) UV(5) สร้างพร้อมปีบัญชีแรกของบริษัท เฉพาะเมื่อบริษัทยังไม่เคยมี record `journal-books` เลย (รวมที่ลบแล้ว) **และยังไม่มีใบสำคัญเลย** — ไม่ปลุกสมุดที่ผู้ใช้ลบกลับมา และไม่สร้างให้บริษัทเดิมที่มีใบสำคัญแต่ไม่มี record สมุด เพราะใบเดิมอาจใช้ SV/UV ความหมายสลับ การสร้างสมุดรหัสเดียวกันพร้อมประเภทคือการเดาประเภทจากรหัส (`defaultJournalBookChanges` :271)
6. **ระบบไม่เดาประเภทของสมุดเก่า** (prod มีสมุดไม่มีประเภท และ SV/UV เดิมความหมายสลับ) → **ผู้ใช้กำหนดประเภทเองตามชื่อสมุดที่ตนตั้งไว้ ไม่ใช่ตามรหัส**: หน้า "กำหนดสมุดรายวัน" มีแถบเตือน (`UntypedJournalBooksNotice` `frontend/src/app/gl/gl-masters.tsx:1162`) แสดงสมุดที่ยังไม่มีประเภททุกเล่มเป็นปุ่ม "รหัส · ชื่อ" กดเปิดไปเลือกประเภทได้ทันที และบอกว่าปิดใช้งานได้โดยไม่ต้องเลือก; ช่องประเภทของสมุดเก่ามีคำอธิบาย `gl_book_type_untyped_hint`; ข้อความ backend ชี้ไปหน้าเดียวกัน (`journal_book_type_missing`, `journal_book_not_found` บอกให้เลือกสมุดที่มีหรือสร้างสมุดนั้นก่อน); `cmd/glseed` หยุดพร้อมข้อความ "ระบบไม่เดาประเภทจากรหัส"
7. **สิทธิ์** (`backend/internal/generalledger/httpapi/http.go:91,127-137`): เลิกผูกสิทธิ์กับรหัสสมุด — ใบสำคัญใช้ `gl-journals` (ประเภท opening รับ `gl-opening-balance` ด้วย; อ่านใบร่าง/ผ่านรายการรับ `gl-post`), ผ่านรายการ **และกลับรายการ** ต้องมี `gl-post:update` (แทน `gl-unpost` ซึ่งเป็นเมนูเดียวกัน "ผ่านรายการและกลับรายการ"), แก้สมุดต้องมี `gl-journal-books:create/update/delete`, อ่าน `journal-books`/`accounts`/`fiscal-years` ได้ทุกคนที่มีสิทธิ์ GL (ใช้เป็นตัวเลือก); รหัสสิทธิ์เก่า `jv/uv/sv/rv/pv-journal` และ `gl-unpost` ไม่ให้สิทธิ์อะไรแล้ว (ตรวจแล้ว: local ไม่มีแถวสิทธิ์เหล่านี้, prod ทุก role เป็น `*`)
8. **สาขาหัวเอกสาร** (`checkJournalBranch` `backend/internal/generalledger/httpapi/branch.go:17`): session ระดับสาขา เว้นว่างหรือสาขาตัวเอง; session ระดับบริษัทต้องส่ง `branchcode` ของสาขาที่ active (`journal_branch_required` / `journal_branch_not_found` / `journal_branch_outside_session`)

## Alternatives

- **คงรายการสมุดตายตัว 5 เล่ม** — ไม่เลือก: ขัดคำสั่งลุงจืดและต้นแบบ Champ ที่เป็น master
- **เดาประเภทจากรหัส (SV→ขาย, UV→ซื้อ) ตอนอ่านสมุดเก่า** — ไม่เลือก: ข้อมูลเดิมใช้ SV/UV สลับกัน เดาแล้วภาษีขาย/ซื้อผิดฝั่ง; ผู้ใช้กำหนดเองครั้งเดียวปลอดภัยกว่า
- **สิทธิ์รายสมุด (หนึ่งสมุดหนึ่งรหัสสิทธิ์)** — ไม่เลือก: สมุดเพิ่ม/ลบได้ตลอด รายการสิทธิ์จะไม่นิ่ง และเมนูปัจจุบันมีรายการเดียว "สมุดรายวัน"
- **ลบสมุดที่มีเอกสารได้แล้วย้ายเอกสาร** — ไม่เลือก: เปลี่ยนหลักฐานที่ผ่านบัญชีแล้ว ผิดหลัก audit trail; ปิดใช้งานแทน

## Consequences

- tenant เดิมบันทึกเอกสารไม่ได้จนกว่าจะมีสมุดที่มีประเภท: local `rungrueng` บริษัท 01 มีใบสำคัญ 206 ใบแต่ไม่มี record `journal-books`; prod มีสมุด AP/AR/GJ/JV/PV/RV/SV/UV ที่ยังไม่มีประเภท → หลัง deploy ต้องกำหนดประเภทที่หน้า "กำหนดสมุดรายวัน" (ลุงจืดตัดสิน SV/UV เอง) หรือ `go run ./cmd/glseed -apply` สร้างสมุดเริ่มต้นที่ขาด
- ข้อความ Thai ของ error มาจาก backend (ระบุรหัสสมุด), ภาษาอื่นมาจากแถว `gl_err_<code>` ใน `languages.tsv` (`httpapi/http.go:205-212`)
- บริษัทเดิมที่มีใบสำคัญแต่ไม่มี record สมุด (เช่น local `rungrueng`/01) ไม่ได้สมุดเริ่มต้นอัตโนมัติ — ต้องสร้างสมุดเองที่หน้า "กำหนดสมุดรายวัน" (ตั้งรหัสให้ตรงกับที่ใบเดิมใช้แล้วเลือกประเภทตามชื่อ)
- Tests: `journal_books_test.go`, `journal_books_integration_test.go` (`TestJournalBooksLifecycleIntegration`), `subledger_new_row_rules_integration_test.go` (`TestJournalBookInUseByMapping`), `httpapi/permissions_test.go`, `httpapi/branch_test.go`, `frontend/src/lib/general-ledger.test.ts`, `frontend/src/app/gl/gl-masters.test.ts`
