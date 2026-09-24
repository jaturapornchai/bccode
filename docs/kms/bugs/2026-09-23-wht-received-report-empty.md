---
date: 2026-09-23
severity: medium
component: [backend]
tags: [bc-account, postgres, tax, wht]
fixed: true
---

# Symptom

The report "รายงานภาษีถูกหัก ณ ที่จ่าย" (`/report/whtreceive`, direction `received`) came back empty on production (`account.bcaicloud.com`, holding `rungrueng`). It showed the note "ไม่พบบัญชีภาษีหัก ณ ที่จ่ายในผังบัญชี".

This happened even though the chart of accounts has account `1153` "ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)".

## Root Cause

`buildWithholdingReport` (`backend/internal/goapi/handlers/tax_report.go`) looked up the withheld-tax account by name. It required the name to contain the word "ภาษีถูกหัก" as one unbroken string.

The standard account name "ภาษี**เงินได้**ถูกหัก ณ ที่จ่าย" has "เงินได้" in between, so it never matched. The function also returned early whenever no account was found. That meant withholding entries the user had recorded in the journal (`details.withholdings`, `wht_direction=2`) were never read either.

## Fix

- A new function (then `journalsWithRecordedWithholdings`; since 2026-09-24 `recordedWithholdingRows` in `backend/internal/goapi/handlers/tax_report.go`, which reads each recorded entry by its payment date) pulls every posted journal in the period that has a recorded withholding entry in the requested direction. Those journals are merged into the report even when the tax line was posted to an account whose name does not match. Recorded evidence beats guessing from the account name.
- For `received`, the name lookup now accepts both "ภาษีถูกหัก" and "ถูกหัก ณ ที่จ่าย". Rows found this way are still labelled `inferred`.
- The "no account found" note now appears only when there is no matching account **and** no recorded entry.
- The PND50/51 tax credit (`generalledger.WithheldFromCompanyTotal`) still counts recorded entries only. This fix does not change it.

## Regression Test

- `TestWithholdingReceivedReport` in `backend/internal/goapi/handlers/tax_withholding_recorded_integration_test.go` runs under `tools/verify.sh postgres`. It covers three journals:
  - one recorded entry posted to account 1153
  - one unrecorded entry posted to 1153 (inferred)
  - one recorded entry posted to an account with a different name
- Against the old code the test gets `rows = []` plus the note (reproduces the bug). Against the new code it gets 3 rows totalling 570.00.
- The existing `TestWithholdingReportUsesRecordedBase` still passes. It covers PND3 not seeing entries recorded as PND53.
