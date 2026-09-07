---
tags: [bc-account, go, pdf, timezone, bug]
date: 2026-06-22
---

# PDF document date off-by-one (UTC not converted to branch timezone)

## Symptom
Server-rendered transaction PDFs printed the document date **one day early** for any
document transacted between 00:00–06:59 Thai local time. e.g. a doc stamped
`2026-06-22T18:30:00Z` (= 23 มิ.ย. 01:30 ICT) printed **22** instead of **23**.
Frontend `date-time.ts` could not fix it — the date is rendered on the backend.

## Root cause
`FormatDateWithFormat` in `backend/internal/goapi/handlers/gen-trans-pdf/common.go`
extracted `t.Year()/t.Month()/t.Day()` directly from a **UTC** `time.Time`
(`doc["docdatetime"]` = `primitive.DateTime.Time()` = UTC), with **no timezone
conversion**. `GenPDFPayload` carried `DateFormat` but no timezone field, so the
formatter had no branch-tz input. Violates the [[Timezone Iron Rule]]
(DB stores UTC+0; user-facing display uses the branch timezone).

## Fix
1. Added `Timezone string \`json:"timezone"\`` (IANA name) to `GenPDFPayload`.
2. Added `loadPDFLocation(tz)` helper: empty → `Asia/Bangkok`; if `time.LoadLocation`
   fails (tzdata missing) → `time.FixedZone("UTC+7", 7*3600)` so the off-by-one is
   still fixed.
3. In `FormatDateWithFormat(val, format, timezone)`: `t = t.In(loadPDFLocation(timezone))`
   **once**, before extracting Y/M/D — fixes every format branch (YYYY-MM-DD,
   DD MMM YYYY, Buddhist +543, etc.) at once.
4. Updated the 6 `base_pdf.go` header call sites to pass `payload.Timezone`, plus the
   2 internal callers (`FormatDateValue`, `GetCurrentDateTimeWithFormat`).
5. Defaulted `payload.Timezone` to `Asia/Bangkok` in `GenPDFHandler` (genpdf_handler.go),
   matching the existing DateFormat/Language default block. Clients (Flutter/Next.js)
   can send a real branch tz to override for non-Thai branches.

## Regression test
`backend/internal/goapi/handlers/gen-trans-pdf/common_test.go` —
`2026-06-22T18:30:00Z` under `Asia/Bangkok` → `2026-06-23` (YYYY-MM-DD),
`23/06/2569` (BE), `23 มิถุนายน 2569`, `23 Jun 2026`; both `time.Time` and
`primitive.DateTime` paths; plus a same-day (05:00Z → 22nd) no-over-correct case.
All pass. `go build ./internal/goapi/handlers/gen-trans-pdf/...` clean.

## Follow-up (not fixed here — scope)
`GetCurrentDateTime()` (PDF "Printed At" footer, base_pdf.go ~937/970) still uses
server-local `time.Now()` — same rule, separate function. Flagged as a spawned task.
