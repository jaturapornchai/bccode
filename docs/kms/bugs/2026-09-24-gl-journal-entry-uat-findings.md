---
date: 2026-09-24
severity: high
component: [frontend, backend]
tags: [bc-account, postgres, gl, tax]
fixed: true
---

# Symptom

Found while running a prod UAT (Demo › rungrueng/01/00000) that recorded a receipt with 3% withheld by the customer (RV: Dr 1111 9,700 / Dr 1153 300 / Cr 4122 10,000 + withholding detail direction 2). There are 4 related problems:

1. **The search box in "ค้นหาและเลือกบัญชี" does not work.** Typed text is cleared on every keystroke, so users can only find an account by scrolling.
2. **Smart Guard misreads the WHT account as VAT.** It warns "ภาษีมูลค่าเพิ่มในเอกสาร 300.00 ต่างจาก 7% ของฐานภาษีที่คาดการณ์ (679.00)" and shows the button "ปรับเป็น 7% พอดี (679.00)". Clicking it would change the withheld tax to 679 and make the entry wrong.
3. **A new journal cannot be saved.** The screen shows "ไม่พบรายการบัญชีในบริษัทหรือสาขานี้" (backend 404 `gl_err_not_found`). It saves only after the user types a branch code into the "รหัสสาขา" field.
4. **The WHT received report still shows a reversed document.** After RV6909-9901 was reversed with RV6909-9902 (net balance of 1153 = 0.00), the report still showed RV6909-9901 with 300.00.

## Root Cause

1. `frontend/src/app/gl/account-search-dialog.tsx`: the "reset on open" effect depended on `selectedCodes`, whose default is `[]`, a new array on every render. The effect ran after every render and called `setSearch("")` each time.
2. `frontend/src/lib/gl-smart-guard.ts` `isVatAccount` used `code.startsWith("115")`/`"214"` to mean input/output VAT. In the real chart, 1153 is ภาษีเงินได้ถูกหัก ณ ที่จ่าย and 1154 is ภ.ง.ด.51 จ่ายล่วงหน้า. Also, any account name containing "ภาษีมูลค่าเพิ่ม" was treated as VAT, including income accounts named "ได้รับยกเว้นภาษีมูลค่าเพิ่ม". `appendVatLine` also fell back to hard-coded codes 1151/2141. **Account codes must never be used as conditions**, because every company has its own chart (ลุงจืด 2026-09-24).
3. `emptyJournal()` sends `branchcode: ""`. `validateJournalLines` rejects the journal when `scope.Branch != "" && j.BranchCode != scope.Branch`, and a session that has selected a branch always has `scope.Branch`. Every new journal created from the screen was therefore rejected.
4. `buildWithholdingReport` and its recorded-entry reader (then `journalsWithRecordedWithholdings`, now `recordedWithholdingRows` in `backend/internal/goapi/handlers/tax_report.go`) read `gl_lines` + `details.withholdings` without checking the journal status. After a reversal the original gets status `reversed`, but its lines and details stay (the reversing document offsets them). The VAT register already filtered `status = 'posted'` (`generalledger/subledger_vat.go`).

## Fix

1. Reset the dialog only when `open` changes false→true (tracked with a `wasOpenRef` ref).
2. `isVatAccount(accountName, lineDescription, accountType)` uses the account name, account type and line description only:
   - Names containing หัก ณ ที่จ่าย / ภาษีเงินได้ / withholding are never VAT.
   - Income/expense accounts are never the VAT line.
   - A generic "ภาษีมูลค่าเพิ่ม" account is classified input or output by accounttype (asset/liability).
   - `appendVatLine` looks the account up by name among posting accounts; if none is found it leaves accountcode empty for the user to choose.
3. `mutateJournal` create: a blank `branchcode` defaults to `scope.Branch`. An explicit different branch is still rejected.
4. Both WHT report queries JOIN `gl_records` and require `payload->>'status' = 'posted'`, the same rule as the VAT register.

## Regression Test

- `frontend/src/lib/gl-smart-guard.test.ts`: covers WHT not being VAT, income accounts that mention VAT, the UAT journal having no VAT line, VAT detected by name in a chart with unusual codes, and appendVatLine without a hard-coded code.
- `backend/internal/generalledger/postgres_integrity_integration_test.go` `TestPostgresIntegrityBlankBranchDefaultsToSession`.
- `backend/internal/goapi/handlers/tax_withholding_recorded_integration_test.go` adds RV4 with status `reversed`; the report must still show 3 rows / 570.00.
- Prod UAT evidence, checked in PostgreSQL at each step:
  - draft → `gl_records` withholdings base "10000" tax "300" direction 2;
  - post → 3 `gl_lines` rows;
  - report 1 row, `taxbasesource: recorded`;
  - reverse → original `reversed` + RV6909-9902, net 1153 = 0.00.
- Cleanup: `gl_subledger_audit` has an append-only trigger (`gl_reject_audit_mutation`), so posted documents cannot be deleted directly. Test data is cleared by a reversal with a stated reason, not by bypassing the trigger.
