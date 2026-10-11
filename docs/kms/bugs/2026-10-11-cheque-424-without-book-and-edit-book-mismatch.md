---
date: 2026-10-11
severity: high
component: [backend, db]
tags: [bc-account, go, postgres, cheque, gl]
fixed: true
---

# Symptom

ผลตรวจ (opus review) ของโมดูลเช็ค พบ 4 จุด
1. บันทึกขายลดเช็ค (424) โดยไม่ระบุสมุดเงินฝาก ได้ HTTP 500 `pq: new row for relation "cheque_registers" violates check constraint` — Champ ไม่บังคับสมุดใน 424
2. แก้เอกสาร 420/424 ของเช็คที่ก่อนเอกสารอยู่ในมือพร้อมสมุด บันทึกไม่ได้ ขึ้น `cheque_err_book_mismatch` ทั้งที่ไม่ได้เปลี่ยนสมุด
3. ยกเลิก/แก้เอกสารสถานะสองใบพร้อมกันที่มีเช็คชุดเดียวกัน อาจล้มด้วย deadlock (40P01)
4. แก้เลขบัญชี/สกุลเงินของสมุดเงินฝากผ่านรายละเอียดบัญชีธนาคารในใบสำคัญได้ แม้สมุดนั้นถูกเช็คอ้างอยู่ (จอสมุดเงินฝากห้ามแล้ว)

# Root Cause

1. CHECK ของเช็ครับเดิม `status = 0 OR book_code IS NOT NULL` (ไม่มีชื่อ → PG ตั้งเป็น `cheque_registers_check3`) และ `registerBook` ยกเว้นแค่สถานะ 4 แต่ 424 ตั้งเช็คเป็นสถานะ 5 ได้โดยไม่มีสมุด (Champ comment การตรวจสมุดออกที่ `cqfrmcheqinsale.cpp:374-382`)
2. `UpdateDoc` ถอนผลรอบเดิมผ่าน `releaseDoc` → `restoreStates` คืนเช็ครับเป็นในมือและล้างสมุด (ตาม Champ `CQFrmBaseTransStatus.cpp:188-190`) แล้ว `lockDocCheques` → `checkSelectable` เทียบสมุดของเอกสารกับสมุดของเช็คที่เพิ่งถูกล้าง จึงไม่ตรง
3. `restoreStates` อัปเดตเช็คตามลำดับที่ `voidDocEvents ... RETURNING` คืนมา (ไม่เรียง) ขณะที่ `lockDocCheques`/`ApplyRecalculate` ล็อกตาม id → สองคำขอล็อกแถวสลับลำดับกัน
4. `subledger_records.go` ตรวจการใช้สมุดด้วย query ของตัวเองที่ไม่นับทะเบียน/เอกสารเช็ค ต่างจาก `bankBookUsed` (`bank_accounts.go:404`) ของจอสมุดเงินฝาก

# Fix

1. `backend/internal/cheque/schema.go:92` CHECK มีชื่อ `cheque_registers_in_book_check` = `status IN (0,4,5)`; DO block `schema.go:106-112` ลบ CHECK เดิมที่ไม่มีชื่อของตารางเก่า (หาโดยนิยาม) แล้วเพิ่มตัวใหม่แบบ idempotent; `register.go:262` `registerBook` ยกเว้น `InDiscounted` ด้วย ให้ตรงกับ CHECK — ต่างจาก Champ จอทะเบียนที่บังคับสมุดเมื่อ ≠ 0 (`cqfrmcheqin.cpp:393-401`) เพื่อให้แก้ทะเบียนของเช็คขายลดที่ไม่มีสมุดได้; frontend `cheque-doc-forms.ts` `registerBookRequired` ยกเว้น 5 ตาม
2. `docs.go:502` `releaseDoc` คืน id เช็คที่ถอนผล; `UpdateDoc` ส่งต่อให้ `lockDocCheques` (`docs.go:203`, `:558`) ซึ่งเฉพาะเช็คที่เพิ่งถอนผลและกลับเป็นในมือไม่มีสมุด ให้ถือว่าสมุดตรงกับเอกสาร (`docs.go:251`)
3. `events.go:159` `sort.Strings(ids)` ก่อนวนอัปเดตใน `restoreStates`
4. `subledger_records.go:218` ใช้ `bankBookUsed` ตัวเดียวกับจอสมุดเงินฝาก (อยู่ใน package `generalledger` เดียวกัน ไม่เกิด import cycle); `postgres_store.go:120` เรียก `cheque.EnsureSchema` ก่อนเปิด tx (เหมือน `executeBankAccount`) เพราะกลุ่มกิจการที่ยังไม่เคยเปิดโมดูลเช็คจะยังไม่มีตาราง `cheque_registers`

# Regression Test

- `backend/internal/cheque/register_test.go:131` (`TestRegisterStatusAndBookRules`) — ก่อนแก้: `register_test.go:132: discounted cheque without a book:` ได้ error
- `backend/internal/cheque/glposting/glposting_integration_test.go:641` `TestEnsureSchemaRelaxesReceivedBookCheck` — ใส่ CHECK เดิมกลับแล้วรัน `EnsureSchema` สองรอบ; ก่อนแก้: `glposting_integration_test.go:673: discounted received cheque without a book must be allowed: pq: new row for relation ...`
- `glposting_integration_test.go:519-550` (`TestChequeLifecycleOnRealLedger` ขั้น 10.1/10.2) — แก้ 424 ของเช็คในมือพร้อมสมุด และ 424 ไม่มีสมุดแล้วแก้ทะเบียนต่อ; ก่อนแก้: `glposting_integration_test.go:526: edit the discount of an on-hand cheque with a book: cheque_err_book_mismatch` และ `:535: doc 424 on 2026-10-22: pq: new row for relation ...`
- `backend/internal/cheque/events_test.go:12` `TestRestoreStatesUpdatesChequesInIDOrder` (sqlmock ตรวจลำดับ UPDATE; deadlock จริงทดสอบให้เกิดแน่นอนไม่ได้) — ก่อนแก้: `events_test.go:31: restoreStates: ExecQuery 'UPDATE cheque_registers ...` ไม่ตรงลำดับที่คาด
- `backend/internal/generalledger/bank_accounts_integration_test.go:258` `TestPostgresJournalBankDetailsKeepNumberOfBookUsedByCheques` — ก่อนแก้: `:276: journal changing the number of a book used by cheques: <nil>`
- `bank_accounts_integration_test.go:284` `TestPostgresJournalBankDetailsBeforeChequeTablesExist` — กันการแก้ข้อ 4 ทำใบสำคัญพังในกลุ่มกิจการที่ยังไม่มีตารางเช็ค; ก่อนเพิ่ม `EnsureSchema`: `:301: change details of an unused book: pq: relation "cheque_registers" does not exist`

ออกแบบ: `docs/kms/champ/design-2026-10-11/cheque-bank.md`
