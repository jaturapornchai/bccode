---
date: 2026-09-24
severity: high
component: [frontend, bff, backend]
tags: [bc-account, session, gl, i18n]
fixed: true
---

# Symptom

Found in the Thai UAT round on 2026-09-24 (S7). Steps:

1. Press Demo, then select rungrueng › 01 › 00000.
2. Open Settings › Login Accounts (`/user`).
3. Go back to the journal list (`/gl/journals`) or to the trial balance.

Every accounting screen then shows "กรุณาเลือกบริษัทก่อนใช้งานบัญชี" and lists 0 items. The header still shows company [01]. In PG, `cache_entries` for the session has `businesscode=''` and `branchuid=''`.

A second bug was found in the same round (S3). When the browser's Accept-Language is `en-US`, every GL error falls back to the generic text "ทำรายการไม่สำเร็จ กรุณาลองใหม่…". Affected codes include `journal_book_in_use_delete` and `partner_tax_id_checksum`. The response actually carried `message_th`, which explains the cause in Thai.

## Root Cause

- **Session:** the settings screen calls `GET /api/workspace/holdings?management=true&activeholdingcode=…` without `businesscode`. To enrich the company list, the BFF calls `POST /select-holding` with `{holdingcode}` only. That call rewrites the server-side session and leaves only the Holding selected.
- **GL errors:** `glRequest` did not send the app language as `Accept-Language`, which differs from `taxRequestHeaders` in `thai-tax.ts`. The backend therefore answered with the English `message`. `commandFailure` accepts only a Thai `message`, so it discarded it and never looked at `message_th`.

## Fix

- **Session:**
  - New backend endpoint `GET /session/selection` (`backend/internal/authentication/authentication_http.go`, `SessionSelection`). It returns the session's current holding, company and branch.
  - Before the BFF switches the Holding temporarily, it reads that selection (`readSessionSelection` in `frontend/src/app/api/workspace/[...workspacePath]/route.ts`). Afterwards it restores both the company and the branch (`workspaceRestoreSelection`).
  - The branch is restored only while the company is unchanged.
- **GL errors:**
  - GL requests now send the app language.
  - When the text received is not Thai, `commandFailure` uses `message_th`.

## Regression Test

- `TestSessionSelectionPayloadCarriesCompanyAndBranch` (backend).
- `frontend/src/app/api/workspace/[...workspacePath]/route.test.ts` (restore selection).
- `general-ledger-api.test.ts` (Accept-Language + `message_th`).
- **Manual retest:**
  1. Press Demo, then open `/user`.
  2. Save a user.
  3. Check that `cache_entries` in PG is still 01/00000.
  4. Open `/gl/journals` and check that all 98 items load.
  5. In an `en-US` browser, enter a wrong tax ID and check that the Thai message appears.
