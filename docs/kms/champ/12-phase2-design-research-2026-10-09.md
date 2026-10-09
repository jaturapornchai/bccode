# ผลสำรวจก่อนทำเฟส 2/3 ของเมนู Champ (ยกเลิก/อนุมัติ, จอใหม่, รายงานใน submenu) — 2026-10-09

ผลงานของ workflow สำรวจ (sonnet อ่านโค้ด+ภาพ Champ, Opus สรุป) เก็บไว้ให้คนรับงานต่อ — **เป็นข้อมูลสำรวจ ไม่ใช่การตัดสินใจ**: ข้อที่เขียนว่า "ต้องให้ลุงจืดอนุมัติ" ยังไม่ได้อนุมัติ ห้ามลงมือจนกว่าลุงจืดจะตอบ ตัวเลข file:line ตรวจ ณ commit `f66f059b` (บรรทัดอาจเลื่อน)
ภาพ Champ = `D:project-champcspecหน้าจอเมนู*` (โฟลเดอร์จัดซื้อสะกด "จัดชื้อ"), ซอร์ส C++ = `D:project-champ`

## ก. จอ ยกเลิก/อนุมัติ 11 จอ (6 จอเดิมรอ API + 5 จอเมนูใหม่ที่เพิ่ม 2026-10-09)

### ของที่มีอยู่ฝั่ง backend
The generic ERP doc API has NO approve/unapprove/cancel endpoint. Routes are only list/get/create/update/delete (backend/internal/transaction/httpapi/http.go:38-44). Auth pattern: authorize() (http.go:64-104) = lookupDocType -> session -> allowsDocType (http.go:113-120, uses doctype.Screens(menu id)) -> PgSqlFastConnect -> ensureSchema -> gl.EnsureSchema only when non-read AND PostsToGL. Actions: "" / create / update / delete (http.go:47-52); write needs both <menu> and <menu>:<action>. Errors via gl.UserError + docFailure (http.go:235-246) -> BFF userErrorStatusOk relays them as HTTP 200 + success:false. Delete handler (http.go:300-318) is the template to copy.

Columns in table doc (schema.go:13-70): iscancel, cancelreason, isref, islocked, isclosed, approval_status text DEFAULT '' (partial index schema.go:79), isclosedmanual + closedmanual_*. REALITY: SaveDoc (save.go:44-219) never writes iscancel / cancelreason / approval_status; DeleteDoc (save.go:223-265) only sets isdelete=true. Only isref+isclosed are written, at save.go:191-194, when a downstream doc references this doc via docref (DDL schema.go:125-135: docno = downstream, docnoref = source, with transflag columns; written save.go:183-188). That makes isclosed "referenced downstream", not "approved", and it goes stale: the ADR says reopening the source when a downstream doc is deleted is not done (docs/kms/decisions/2026-10-09-erp-posting-reverses-on-edit-delete.md:29). docdetail.iscancel exists (schema.go:81-113) but is never written; there are no remain/cancel-qty columns. Edit/delete have no guard on iscancel / approval / isref. The Status shown to the UI is derived in service.go:78-84 and 139-145 (iscancel->2, isclosed->1 mislabelled "Approved", else 0); ListDocs filters only NOT isdelete and searches docno/custcode (ILIKE) only, with no approval filter. ErpTransactionDoc (models.go:25-55) has Status/IsClosed/IsCancel but no approval or cancel-reason fields, and ErpDetailItem has no remain/cancel qty. No audit/event table exists (only erp_doc_journals, schema.go:189-201). ensureSchema (schema.go:204-232) already uses advisory lock + CREATE IF NOT EXISTS + ALTER TABLE doc ADD COLUMN IF NOT EXISTS (schema.go:71-73) = the pattern for new columns.

Registry backend/internal/doctype/doctype.go:55-126: purchase-requisition(2), purchase-order(6), quotation(34), sale-order(36), stock-issue-request(430), stock-transfer-request(431) are all StockNone and not in postingFlags (doctype.go:201-203; frontend mirror erp-transaction-modules.ts:80-93). So approve/cancel on these six types touches NO GL journal and NO stock_ledger; serialized(dt) is false so no company lock is needed. The GL/stock reversal machinery of the ADR is only relevant for the downstream docs (purchase, sale-invoice ...).

Champ truth (docs/kms/champ/01-purchase.md:73-158, 281-338, 925-1011; 02-sales.md:250-275, 363-420, 752-800; source POFrmApprovePurchaseOrder.cpp:235-236,373,407 and ChmFrmPurchaseOrder.cpp:403,550,1457): QT/PO/SO "approve" screens = header-level batch IsConfirm=1 (+ConfirmCode/Time; SO also releases HoldingStatus 1->0). PR approval (5805) is itself a document (BCReqConfirm) with partial ConfirmQty that recomputes PR RemainQty/BillStatus. The "cancel" menu screens (PO 5808, SO 5308, QT, PR, PR-approval, issue-request, transfer-request) are line-level cancel DOCUMENTS (BCPOVoid/BCSOVoid/...) with VoidQty <= RemainQty and a BCCancelReason master; whole-doc cancel is Ctrl+F5 (IsCancel=1; PO PoStatus=4) and is blocked when IsConfirm=1. The BC KMS itself says to decide with ลุงจืด before following (02-sales.md:4040).

### ฝั่ง frontend
The 6 existing screens (menu-data.ts:190, 205, 249, 250, 252, 253) all route to the operations workbench and are PENDING by construction. frontend/src/lib/erp-operations.ts:208 has `APPROVAL_KINDS = {}` (empty) and isApprovalApiReady (210-212) therefore returns false for every kind. The fetchers (239-311) call a non-existent `/api/goapi/api/approval/<kind>-status/{pending|approve|reject}`, which no backend registers. operations-workbench.tsx:89 `approvalReady`; lines 179-184 show the "not connected" notice; the pending table and approve/reject buttons (221-352) render only when ready, with no reason input and no cancel action. Tests lock the pending state: erp-operations.test.ts:66-69, menu-screen-status.ts:95-96, menu-screen-status.test.ts:102-104 (must be flipped in the same change). The wiring that DOES work is the BFF catch-all frontend/src/app/api/erp-transaction/[...erpPath]/route.ts: module whitelist ERP_TRANSACTION_MODULES + segment sanitizer, GET allows only the query params q,page,limit,offset,fromdate,todate,custcode,branchcode,status,lang (route.ts:30); POST/PUT/DELETE relay to /transaction/<path> with userErrorStatusOk (POST 42-80, requires a JSON object/array body, so approve/cancel POSTs must send `{}` or {reason}). So new sub-path POSTs (/:module/:id/approve) relay with NO BFF change; a new GET filter (`status=pending`) already passes the allowlist. helpers in erp-transaction.ts (saveErpTransaction/deleteErpTransaction: apiFetch + Accept-Language + success:false check + statusErrorKey) are the template for new approve/cancel helpers. Three of the 11 Champ items have no menu entry in BC yet (listed in docs/kms/champ/11-menu-parity-audit-2026-10-09.md).

### ต่อจอ (Tier A = ระดับหัวเอกสาร, ใช้คอลัมน์เดิม)
| เมนู | doctype | action | อนุญาตจากสถานะ | ผลลัพธ์ | ผลข้างเคียง |
|---|---|---|---|---|---|
| ยกเลิกใบเสนอซื้อสินค้า (screenshot: ยกเลิกใบเสนอชื้อสินค้า.jpg) | purchase-requisition (transflag 2) | cancel | not deleted, not cancelled, approval_status='' (not approved) and no live downstream doc (PO/RFQ referencing it via docref) | iscancel=true, cancelreason set, cancelled_by/at; list Status=2 | None to GL/stock (StockNone). Append erp_doc_events row. Champ original is a line-level void doc (VoidQty<=RemainQty) - header-level in Tier A. |
| บันทึกอนุมัติใบเสนอซื้อสินค้า (screenshot: บันทึกอนุมัติใบเสนอชื้อสินค้า.jpg; exists in BC menu) | purchase-requisition (transflag 2) | approve | not deleted, not cancelled, approval_status='' | approval_status='approved', approved_by/at; Status=1 | None to GL/stock. Edit/delete of the PR blocked afterwards (Champ IsConfirm=1 blocks save). Champ original is a partial-ConfirmQty document (BCReqConfirm) that recomputes PR RemainQty via UpdateDocumentFlow - Tier B. |
| ยกเลิกใบอนุมัติเสนอซื้อสินค้า (screenshot: ยกเลิกใบอนุมัติเสนอชื้อสินค้า.jpg) | purchase-requisition (transflag 2) | unapprove | approval_status='approved', not cancelled, no live downstream doc | approval_status='' (back to pending), approved_by/at cleared; Status=0 | None to GL/stock. Event row records the reversal with reason. This screen is menu-missing in BC today (menu-parity audit). |
| อนุมัติใบสั่งซื้อสินค้า (screenshot: อนุมัติใบสั่งชื้อสินค้า.jpg) | purchase-order (transflag 6) | approve | not deleted, not cancelled, approval_status='' | approval_status='approved' (Champ: UPDATE BCPurchaseOrder SET IsConfirm=1, POFrmApprovePurchaseOrder.cpp:373); Status=1 | None to GL/stock. Header-level batch (multi-select) like Champ. |
| ยกเลิกใบสั่งซื้อสินค้า (screenshot: ยกเลิกใบสั่งชื้อสินค้า.jpg; exists in BC) | purchase-order (transflag 6) | cancel | not deleted, not cancelled, approval_status in ('', 'approved'); REJECT if any live (not deleted/cancelled) downstream doc exists (goodsreceive/purchase referencing it) | iscancel=true, cancelreason, cancelled_by/at (Champ whole-doc cancel: IsCancel=1, PoStatus=4); Status=2 | No GL/stock for the PO itself. PO already received (goods receive / purchase exists): Tier A refuses with gl_err_erp_doc_has_downstream - user must cancel/delete the downstream docs first; remaining-qty cancel of a partly received PO is Tier B (needs remain_qty). Do NOT touch the downstream docs' journals/stock. |
| อนุมัติใบเสนอราคา (screenshot: อนุมัติใบเสนอราคาสินค้า.jpg) | quotation (transflag 34) | approve | not deleted, not cancelled, approval_status='' | approval_status='approved' (Champ IsConfirm=1 + ConfirmCode/Time); Status=1 | None to GL/stock. Locks edit/delete (guard). |
| ยกเลิกใบเสนอราคา (screenshot: ยกเลิกใบเสนอราคาสินค้า.jpg) | quotation (transflag 34) | cancel | not deleted, not cancelled, no live downstream sale-order | iscancel=true, cancelreason, cancelled_by/at; Status=2 | None to GL/stock. |
| อนุมัติใบสั่งขาย/สั่งจองสินค้า (screenshot: อนุมัติใบสั่งขาย สั่งจองสินค้า.jpg) | sale-order (transflag 36) | approve | not deleted, not cancelled, approval_status='' | approval_status='approved' (Champ IsConfirm=1); Status=1 | None to GL/stock. Champ also flips HoldingStatus 1->0 (credit-hold release) - BC has no credit-hold concept, so NOT modelled; flag for ลุงจืด. |
| ยกเลิกใบสั่งขายสินค้า (screenshot: ยกเลิกใบสั่งขายสินค้า.jpg) | sale-order (transflag 36) | cancel | not deleted, not cancelled, approval_status in ('', 'approved'); REJECT if a live downstream doc (sale-invoice, receivedeposit referencing it by docref) exists | iscancel=true, cancelreason, cancelled_by/at (Champ: BCSOVoid 5308); Status=2 | Deposits: a receivedeposit referencing the SO is a downstream doc, so Tier A blocks cancel until the deposit doc is cancelled/refunded (receiveadvancerefund) - never auto-reverses the deposit's GL. No GL/stock change on the SO itself. |
| ยกเลิกใบขอเบิกใช้สินค้า,วัตถุดิบ (screenshot: ยกเลิกใบขอเบิกใช้สินค้า,วัตถุดิบ.jpg) | stock-issue-request (transflag 430) | cancel | not deleted, not cancelled, no live downstream stock-pickup-product | iscancel=true, cancelreason, cancelled_by/at; Status=2 | None (StockNone request doc: no stock_ledger movement). No approve step exists in Champ for this type, so approve is not exposed. |
| ยกเลิกใบขอโอนสินค้า (screenshot: ยกเลิกใบขอโอนสินค้า.jpg) | stock-transfer-request (transflag 431) | cancel | not deleted, not cancelled, no live downstream stock-transfer | iscancel=true, cancelreason, cancelled_by/at; Status=2 | None (StockNone). No approve step in Champ. |

### Endpoint ที่เสนอ (Tier A)
- GET /transaction/:module/list?status=pending|approved|cancelled (new optional filter on existing list handler; module in {purchase-requisition,purchase-order,quotation,sale-order}) - pending = NOT isdelete AND NOT iscancel AND approval_status=''; approved = approval_status='approved'; cancelled = iscancel. Adds fromdate/todate/custcode filters already allowed by the BFF. Response = existing DocListResponse with new fields approvalstatus, approvedby, approvedat, cancelreason, cancelledby, cancelledat. BFF already forwards `status`.
- POST /transaction/:module/:id/approve - body {} (BFF requires a JSON object). Allowed only for the 4 approvable modules. One sql.Tx: SELECT ... FOR UPDATE the doc row, check transitions (from approval_status='' and not iscancel/isdelete), UPDATE approval_status='approved', approved_by, approved_at; INSERT erp_doc_events. 200 {success:true,data:doc}; illegal state -> gl.UserError gl_err_erp_doc_not_pending (relayed as 200 success:false). Permission: <menu> and <menu>:approve.
- POST /transaction/:module/:id/unapprove - body {reason?}. Only purchase-requisition exposed in the menu (generic for the 4 approvable types). Requires approval_status='approved', not cancelled, no live downstream doc; sets approval_status=''; event row stores the reason. Permission <menu>:approve.
- POST /transaction/:module/:id/cancel - body {reason} (required, trimmed, max length checked; Champ BCCancelReason master is NOT replicated, free text in Tier A). Modules: all six types. Tx: FOR UPDATE, guards (not deleted/cancelled, no live downstream via docref join, status rule per screen row), UPDATE iscancel=true, cancelreason, cancelled_by/at; event row. No GL/stock call (all six are StockNone/non-posting); if a future cancellable type has PostsToGL or moves stock the handler must call the existing reversal path under LockCompanyTx (serialized(dt)) - not needed now. Permission <menu> and <menu>:cancel.
- POST /transaction/:module/approve-batch - body {ids:[docno...]} for the Champ-style multi-select approve grids (QT/PO/SO/PR); loops the same per-doc function in ONE tx and returns per-id result [{id,ok,code}] (all-or-nothing is NOT used: continue and report each, matching Champ batch screens).
- GET /transaction/:module/:id/events - read the append-only erp_doc_events audit rows (action, from_status, to_status, reason, actor, at). Permission read.
- Guards on EXISTING handlers (no new route): update (PUT) and delete (DELETE) must reject when iscancel=true or approval_status='approved' (gl_err_erp_doc_locked), mirroring Champ IsConfirm blocking save; DeleteDoc should also reject a doc that has live downstream docs.
- Frontend wiring (no new BFF route): add helpers in erp-transaction.ts (listPendingErpDocs / approveErpDoc / unapproveErpDoc / cancelErpDoc / approveBatchErpDocs) copying saveErpTransaction (apiFetch, Accept-Language, success:false -> detail); fill APPROVAL_KINDS in erp-operations.ts:208 per screen and drop the dead /api/goapi/api/approval/... fetchers (239-311); operations-workbench.tsx:89,179-184,221-352 render the table + approve button, plus a reason dialog (required) for cancel/unapprove; flip the pending expectations in erp-operations.test.ts:66-69, menu-screen-status.ts:95-96 and menu-screen-status.test.ts:102-104; add the 3 missing menu entries only after ลุงจืด confirms (menu-parity audit).

### SQL ที่เสนอ (ใส่ใน `ensureSchema` ของ `backend/internal/transaction/httpapi/schema.go` — ห้าม DDL มือ)
```sql
-- Dev rule: no manual DDL, code creates structure. Put these in ensureSchema (httpapi/schema.go) next to the existing ALTER ... ADD COLUMN IF NOT EXISTS lines (schema.go:71-73), all lowercase snake_case, nothing dropped.
-- 1) doc: approval_status (exists, '' default) takes only '' or 'approved'; iscancel and cancelreason already exist. Add who/when:
ALTER TABLE doc ADD COLUMN IF NOT EXISTS approved_by text NOT NULL DEFAULT '';
ALTER TABLE doc ADD COLUMN IF NOT EXISTS approved_at timestamptz;
ALTER TABLE doc ADD COLUMN IF NOT EXISTS cancelled_by text NOT NULL DEFAULT '';
ALTER TABLE doc ADD COLUMN IF NOT EXISTS cancelled_at timestamptz;
ALTER TABLE doc ADD CONSTRAINT doc_approval_status_chk CHECK (approval_status IN ('', 'approved')) NOT VALID; -- guard in a DO $$ ... IF NOT EXISTS block, since ADD CONSTRAINT has no IF NOT EXISTS
-- 2) append-only audit trail (Rule 20: corrections are new rows, never UPDATE/DELETE)
CREATE TABLE IF NOT EXISTS erp_doc_events (
  id bigserial PRIMARY KEY,
  businesscode text NOT NULL,
  transflag integer NOT NULL,
  docno text NOT NULL,
  action text NOT NULL CHECK (action IN ('approve','unapprove','cancel')),
  from_status text NOT NULL DEFAULT '',
  to_status text NOT NULL DEFAULT '',
  reason text NOT NULL DEFAULT '',
  actor text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_erp_doc_events_doc ON erp_doc_events (businesscode, transflag, docno, id);
-- 3) pending-list index for the 4 approvable types
CREATE INDEX IF NOT EXISTS idx_doc_pending ON doc (businesscode, transflag, docdatetime) WHERE NOT isdelete AND NOT iscancel AND approval_status = '';
-- 4) live-downstream check used by cancel/unapprove/delete guards (parameterised, NOT via the stale isref flag):
-- SELECT 1 FROM docref r JOIN doc d ON d.businesscode=r.businesscode AND d.docno=r.docno AND d.transflag=r.docnotransflag
--  WHERE r.businesscode=$1 AND r.docnoref=$2 AND r.docnoreftransflag=$3 AND NOT d.isdelete AND NOT d.iscancel LIMIT 1;
-- 5) Tier B only (needs ลุงจืด approval): ALTER TABLE docdetail ADD COLUMN IF NOT EXISTS cancel_qty numeric(18,4) NOT NULL DEFAULT 0 and remain_qty, plus a reason master table erp_cancel_reasons (code, name) for BCCancelReason parity.
-- i18n: add rows in backend/assets/language/languages.tsv (13 columns key th en cn ja km ko lo my vi ms id fil, others = English placeholder): gl_err_erp_doc_not_pending, gl_err_erp_doc_not_approved, gl_err_erp_doc_already_cancelled, gl_err_erp_doc_locked, gl_err_erp_doc_has_downstream, gl_err_erp_cancel_reason_required, erp_msg_approved, erp_msg_unapproved, erp_msg_cancelled, plus the UI keys for reason dialog / pending table headers.
```

### ความเสี่ยง
- Champ parity gap (big decision for ลุงจืด): Champ's cancel screens are line-level cancel DOCUMENTS (VoidQty <= RemainQty, reason master BCCancelReason) and PR approval is a partial-ConfirmQty document. Tier A (header-level approve/cancel with reason) is smaller and covers whole-doc workflows, but is NOT identical to Champ; Tier B (remain_qty/cancel_qty per line + void documents + reason master) is the true parity and is much larger. AGENTS.md says big deviations need approval first and docs/kms/champ/02-sales.md:4040 says decide with ลุงจืด - recommend shipping Tier A and asking. Not verified: whether any Champ user relies on partial-qty void.
- Downstream detection: isref/isclosed (save.go:191-194) are set when a child references the doc and never reset when the child is deleted, so using them would block cancels forever; the design uses a live docref JOIN doc query instead. docref only has the global header-level ref (save.go:176-188), line-level doc references (docdetail.docref) are not in docref, so a downstream doc that references the PO only per-line would be missed - unverified how the screens fill docref.
- PO cancelled after goods received / SO with deposits: Tier A hard-blocks (gl_err_erp_doc_has_downstream) instead of cascading; cascading would need reversing downstream GL/stock and conflicts with the ADR rule that settled docs cannot be edited. Users who want to cancel the unreceived remainder of a partly received PO cannot until Tier B.
- Behavior change on existing handlers: edit/delete guards for approved/cancelled docs will make currently editable PO/QT/SO/PR fail after approval (matches Champ IsConfirm blocking save, but any UAT data already approved by hand in DB would be locked). service.go Status mapping (isclosed -> 1) must change, otherwise referenced docs keep showing 'Approved'; frontend/other code reading Status or IsClosed must be checked.
- approval_status text had no constraint and no writer so far; adding a CHECK to an existing table needs the NOT VALID/DO-block form. Deleting a cancelled doc (DeleteDoc sets isdelete) must keep the event rows; docno is never reused (ADR), so events stay unambiguous.
- Concurrency: two users approving/cancelling the same doc - use SELECT ... FOR UPDATE on the doc row inside the tx and re-check state; no company-wide advisory lock needed because none of these 6 types are serialized(dt) (doctype.go postingFlags). If a future type with PostsToGL/stock gets approve/cancel it must reverse via the existing ExecuteSourceTx path under LockCompanyTx.
- Permissions: new actions <menu>:approve and <menu>:cancel are not in any role/permission-set catalogue yet; owner/demo have '*' so Demo works, but other roles and the Developer screen-access metadata (memory: developer-screen-access) need the new action names or these screens are 403 for them. Admin-defined role UI not checked.
- SO approval in Champ also releases HoldingStatus (credit hold) and QT/PO/SO may have approval-level/limit rules in Champ not found in the code I read; BC has no credit-hold concept. Not modelled; flag rather than invent (Rule 1).
- BFF: POST requires a JSON object body (route.ts:64-66) so the frontend must send `{}` for approve; GET list `status` filter passes the allowlist (route.ts:30), but if a different query name is chosen it will be silently dropped. Sub-path POST relay is by reading the code, not by running it - verify with a real call.
- Cancelled docs stay in the table with iscancel=true; ListDocs currently shows them (only NOT isdelete filter) and Status=2, so list/report screens that sum totals must exclude iscancel - report queries not audited.
- Four of the screens (PR approval unapprove, QT/PO/SO approve via menu id) have Champ screenshots but I only viewed them, I did not trace each Champ C++ dialog source for stored-procedure side effects other than PO approve (POFrmApprovePurchaseOrder.cpp) and the KMS notes; side effects of QT/SO approve beyond IsConfirm are taken from the KMS, not from source.

### Test ที่ต้องมี
- Go unit (httpapi, no DB): transition table test for approve/unapprove/cancel per module (allowed vs rejected states, approvable-module whitelist, cancel allowed on the 6 types, approve rejected on stock-issue-request/stock-transfer-request), reason required/trimmed, allowsDocType with <menu>:approve and <menu>:cancel (read-only user gets 403 with erp_err_action_forbidden).
- Go integration (tag integration, BC_GL_TEST_POSTGRES_DSN, like save_integration_test.go): create PR/PO/QT/SO -> approve -> row has approval_status='approved', approved_by/at and one erp_doc_events row; unapprove -> back to ''; cancel -> iscancel=true, cancelreason saved, cancelled_by/at, Status=2; verify with psql/SELECT after each step (UAT rule).
- Integration: PO with a live goodsreceive/purchase referencing it -> cancel returns gl_err_erp_doc_has_downstream and the PO row is unchanged; cancel/delete the downstream then cancel PO succeeds; SO with receivedeposit same.
- Integration: PUT and DELETE on an approved or cancelled doc are rejected (gl_err_erp_doc_locked) and the doc/docdetail rows are unchanged; DELETE of a non-approved doc still works and keeps events.
- Integration: concurrency - two parallel approve calls on the same doc: exactly one succeeds, one gets not_pending, exactly one event row.
- Integration: approve/cancel on these 6 types create NO gl_journals / stock_ledger / erp_doc_journals rows (count before/after equals).
- Integration: list ?status=pending|approved|cancelled returns the right sets and total/pagination; deleted docs never appear; approve-batch returns per-id results and a partial failure does not roll back the successes.
- Audit trail: erp_doc_events is append-only (no UPDATE/DELETE statements in code - grep test), actor and reason stored, GET events returns rows in order.
- languages.tsv guard test: every new key has all 13 columns (missing columns => key skipped, docs/kms/17-dev-gotchas.md); Thai text shown when lang=th, English when lang=en.
- Frontend vitest: erp-operations.test.ts / menu-screen-status.test.ts updated so the 6 screens are no longer pending; helper tests for success:false relay (detail shown) and 403 -> erp_err_action_forbidden; reason dialog blocks empty reason; workbench renders pending table, approve and cancel actions.
- Browser UAT via Demo button (owner ['*']): login Demo -> rungrueng -> company 01 -> branch -> each of the 11 screens: list pending, approve one, cancel one with reason, check network payload (not just HTTP 200), console errors, then psql check of doc + erp_doc_events; verify th and en text; run npm run verify (tsc/vitest/build) and go test ./internal/transaction/... before reporting done.

### ขนาดงาน
Tier A (recommended first): about 1.5-2.5 days. Backend 3 handlers + batch + events (about 1 day: schema/ALTER, Store methods approve/unapprove/cancel with FOR UPDATE and downstream guard, edit/delete guards, service.go Status + list filter + new response fields, permissions actions, about 10 languages.tsv rows), frontend helpers + workbench reason dialog + APPROVAL_KINDS wiring + test updates (about 0.5-1 day), tests/UAT/verify (about 0.5 day). Tier B (Champ-exact line-level void documents, remain_qty, reason master, PR partial-ConfirmQty approval document, UpdateDocumentFlow equivalent) is about 5-8 days and needs ลุงจืด's explicit approval first.

## ข. จอ/เครื่องมือใหม่ที่ยังเป็น `/pending/*` (ขนาดงานและคำถามเปิด)

### ใบสั่งจองสินค้า — document, ขนาด M
- ภาพ Champ: หน้าจอเมนูจัดขาย\ใบสั่งจองสินค้า.jpg , หน้าจอเมนูจัดขาย\ใบสั่งขายสินค้า.jpg
- หน้าตา/ที่มา: ฟอร์มเดียวกับใบสั่งขาย (billfrmsaleorder.cpp) แต่ preset SOStatus=1 (จอง) vs 0 (สั่งขาย) ที่ billfrmsaleorder.cpp:357-361,1167,1755: หัวเอกสาร เลขที่/วันที่/ลูกค้า/พนักงานขาย + กริดรายการสินค้า + ท้ายบิล; ปุ่มมาตรฐานเอกสารขาย
- จอ/โค้ด BC ที่ใกล้สุด: backend/internal/doctype/doctype.go:70 (sale-order flag 36 SO); frontend/src/lib/erp-transaction.ts:133-141 (route saleorder); menu-data.ts:251,254 (sale-order tx + reservation flow)
- ลง GL/สต็อก: ไม่ลง GL; เป็นการจองสต็อก (ไม่ตัดสต็อกจริง) - ยืนยันจากโค้ด billfrmsaleorder.cpp ได้แค่ว่าเป็น SOStatus flag ยังไม่ได้ตามถึง stock engine
- คำถามที่ยังเปิด: ทำเป็น doctype ใหม่ (key 434+) หรือ type flag บนใบสั่งขายเดิม; prefix SR ชนกับใบรับคืนขาย (doctype.go:72 sale-invoice-return ใช้ SR); ต้องตรวจว่า Champ ใช้ยอดจองไปลดยอดพร้อมขายหรือไม่

### ใบสั่งจ่ายสินค้า — document, ขนาด M
- ภาพ Champ: หน้าจอเมนูจัดขาย\ใบสั่งจ่ายสินค้า.jpg , หน้าจอเมนูจัดขาย\ปรับปรุงวันที่ส่งของให้ลูกค้า.jpg
- หน้าตา/ที่มา: เอกสารใบสั่งจ่ายสินค้า (Delivery Order, BCDeliveryOrder ใน MyDebtReCalc.cpp:787,928-932,3074-3112) อ้างใบสั่งขาย/ใบกำกับ หัวเอกสาร+กริดสินค้า+วันส่ง; เชื่อมกับจอปรับปรุงวันที่ส่งของ
- จอ/โค้ด BC ที่ใกล้สุด: backend/internal/doctype/doctype.go:70 (แม่แบบเอกสารขายแบบ doc-only); frontend erp-transaction.ts:133-141; menu-data.ts:274 (delivery-date)
- ลง GL/สต็อก: ไม่ทราบ - ไม่เห็นหลักฐานชัดว่า DO ตัดสต็อก (MyDebtReCalc.cpp ใช้คำนวณยอดค้าง/สถานะส่ง) ยังไม่ได้ verify
- คำถามที่ยังเปิด: DO ตัดสต็อกหรือไม่ และเชื่อมสถานะส่งกลับใบสั่งขายยังไง (docref); ต้องเพิ่ม doctype ใหม่ใน doctype.go แล้ว mirror ฝั่ง frontend 4 จุด + languages.tsv + tests

### บันทึกเตรียมจ่ายชำระหนี้ — document, ขนาด M
- ภาพ Champ: หน้าจอเมนูเจ้าหนี้\บันทึกเตรียมจ่ายจ่ายชำระหนี้.jpg , หน้าจอเมนูเจ้าหนี้\บันทึกจ่ายชำระหนี้.jpg
- หน้าตา/ที่มา: หน้าตาคล้ายใบจ่ายชำระหนี้: หัวเอกสาร เจ้าหนี้/วันที่ + กริดใบแจ้งหนี้ค้างจ่าย + ช่องทางจ่าย; สถานะเตรียมจ่าย (PREPAIR_PAYMENT ใน cqfrmcheqout.cpp:153) ก่อนออกเช็ค/จ่ายจริง
- จอ/โค้ด BC ที่ใกล้สุด: backend/internal/doctype/doctype.go:90 (pay flag 29 PV); frontend erp-transaction.ts:319 (pay route); doc-only ไม่อยู่ใน postingFlags doctype.go:201
- ลง GL/สต็อก: น่าจะไม่ลง GL (เป็นขั้นเตรียม - รายการจ่ายจริงเป็นตัวลง); ยืนยันจากสถานะ cheque-out เท่านั้น ยังไม่ verify
- คำถามที่ยังเปิด: เตรียมจ่ายแล้วโยงไปออกเป็นใบจ่ายจริงยังไง (docref/copy); ผูกกับเช็คเตรียมจ่ายหรือไม่

### บันทึกเช็คเตรียมจ่าย — tool, ขนาด M
- ภาพ Champ: หน้าจอเมนูเช็ค บัตร\บันทึกเช็คเตรียมจ่าย.jpg
- หน้าตา/ที่มา: จอสร้างเช็คจ่ายเป็นชุด (batch generator เลขที่เช็ค): เลือกสมุดเช็ค/ธนาคาร, เลขเริ่มต้น, จำนวนใบ, วันที่ -> ได้แถวเช็คสถานะเตรียมจ่าย (PREPAIR_PAYMENT cqfrmcheqout.cpp:153)
- จอ/โค้ด BC ที่ใกล้สุด: backend/internal/doctype/doctype.go:114 (chequeissued 418 CHQ-OUT); frontend erp-transaction.ts:371; menu-data.ts:388-418 (เมนูเช็ค)
- ลง GL/สต็อก: ไม่ลง GL (สร้างเลขเช็คสถานะเตรียมจ่าย); ยังไม่ verify ละเอียด
- คำถามที่ยังเปิด: สร้างเป็น chequeissued หลายแถวใน doc เดิมหรือแยกตาราง; สถานะเตรียมจ่ายต้องมี column/สถานะใหม่หรือไม่

### ข้อมูลการเคลื่อนไหวเช็ค — inquiry, ขนาด M
- ภาพ Champ: หน้าจอเมนูเช็ค บัตร\ข้อมูลการเคลื่อนไหวเช็ค.jpg
- หน้าตา/ที่มา: จอสอบถามอย่างเดียว (CQFrmChqFlow.cpp): ค้นเลขที่เช็ค -> กริดไล่สถานะ/วันที่/เอกสารที่เกี่ยวข้องของเช็คนั้นตามลำดับเวลา
- จอ/โค้ด BC ที่ใกล้สุด: ยังไม่มีแม่แบบ inquiry ใน BC: frontend/src/lib/erp-tools.ts:147 (TOOL_ENDPOINTS ว่าง) และเมนูเช็ค menu-data.ts:388-418; ข้อมูลมาจาก doc/docref ของ cheque doctypes (doctype.go:114+)
- ลง GL/สต็อก: ไม่ - อ่านอย่างเดียว
- คำถามที่ยังเปิด: ต้องมี API query ใหม่ฝั่ง backend (รวมประวัติเช็คข้ามเอกสาร chequedeposit/clear/return/cancel); ไม่มีหน้าจอ inquiry ที่ทำงานจริงให้ reuse

### บริการธนาคาร — setting, ขนาด M
- ภาพ Champ: หน้าจอเมนูธนาคาร เงินสด\บันทึกรายละเอียดธนาคาร.jpg , หน้าจอเมนูธนาคาร เงินสด\เมนู-ประกอบ.jpg
- หน้าตา/ที่มา: ไม่มี jpg ชื่อ บริการธนาคาร (ใช้บันทึกรายละเอียดธนาคารแทน); จอ Champ มี 3 แท็บ/ตาราง: ดอกเบี้ย BCBANKINTERESTS, สกุลเงิน BCBANKCURRENCY, ค่าธรรมเนียม BCSERVICEFEE (CashFrmBankService*.cpp)
- จอ/โค้ด BC ที่ใกล้สุด: frontend/src/lib/system-setting-screens.ts:892-920 (bookbankscreen, kind main-crud); field types :13-40 ไม่มีชนิดกริดลูก; menu-data.ts:421-458
- ลง GL/สต็อก: ไม่ - เป็นข้อมูลหลัก
- คำถามที่ยังเปิด: framework main-crud ไม่มี child-rows grid ต้องเพิ่มชนิดฟิลด์หรือทำเป็นจอเฉพาะ; ยืนยันว่า jpg ที่ตรงคือบันทึกรายละเอียดธนาคาร

### ปรับปรุง ข้อมูล Supplier ตามสินค้า — tool, ขนาด M
- ภาพ Champ: หน้าจอเมนูสินค้าคงคลัง\ปรับปรุง ข้อมูล Supplier ตามสินค้า.jpg
- หน้าตา/ที่มา: จอเครื่องมือ: ตัวกรองสินค้า/กลุ่ม + กริดสินค้าพร้อมคอลัมน์ Supplier หลัก/รอง แก้ทีละหลายแถวแล้วบันทึก
- จอ/โค้ด BC ที่ใกล้สุด: frontend/src/lib/erp-operations.ts:170-176 (price-adjustment placeholder) และ menu-data.ts:494; erp-tools.ts:147 TOOL_ENDPOINTS ว่าง; menu-screen-status.ts:60-74 (pending card)
- ลง GL/สต็อก: ไม่ - แก้ master สินค้า/ผู้จำหน่ายเท่านั้น
- คำถามที่ยังเปิด: BC ไม่มีรูปแบบ filter+grid tool ที่ใช้งานได้จริง ต้องสร้าง pattern ใหม่; ตาราง supplier-by-product ฝั่ง BC มีหรือยัง (ไม่ได้ตรวจ)

### ตรวจนับสต็อกโดย Hand Held — tool, ขนาด M
- ภาพ Champ: หน้าจอเมนูสินค้าคงคลัง\Stock to count by Hand held.jpg
- หน้าตา/ที่มา: จอเตรียมข้อมูลนับสต็อกส่งออก/นำเข้าเครื่อง Hand Held: เลือกคลัง/ที่เก็บ/กลุ่มสินค้า -> สร้างไฟล์ส่งเครื่อง และรับไฟล์ผลนับกลับเป็นใบตรวจนับ
- จอ/โค้ด BC ที่ใกล้สุด: backend/internal/doctype/doctype.go:66 (stock-count 433 CNT); frontend erp-transaction.ts:509 (stockcount), :549 (count-sheet); menu-data.ts:513-517
- ลง GL/สต็อก: ตัวใบตรวจนับ (433) เป็น doc-only ใน BC; Champ จอนี้เตรียม/นำเข้าข้อมูลนับ ไม่ตัดสต็อกเอง (ยังไม่ verify ปลายทาง)
- คำถามที่ยังเปิด: ไม่มีฮาร์ดแวร์ Hand Held ในเว็บ - ต้องตัดสินว่าทำเป็นนำเข้า/ส่งออกไฟล์ (CSV) หรือข้าม; รูปแบบไฟล์ของ Champ ยังไม่ได้ตรวจ

### Transaction Lock — tool, ขนาด M
- ภาพ Champ: หน้าจอเมนูสินค้าคงคลัง\Transaction Lock.jpg
- หน้าตา/ที่มา: จอเครื่องมือล็อกรายการ: เลือกช่วงวันที่/สินค้า -> ล็อกให้ต้นทุนไม่ถูกคำนวณใหม่ (อนุมานจาก islockcost ใน stockprocess_class.cs:4420-4421)
- จอ/โค้ด BC ที่ใกล้สุด: backend/internal/transaction/httpapi/schema.go:43 (doc.islocked มีแต่ไม่ถูกใช้) และ :137-170 (stock_ledger ไม่มี column ล็อก); menu-data.ts สินค้าคงคลัง
- ลง GL/สต็อก: ไม่ลง GL; กระทบการคำนวณต้นทุนสต็อก (freeze recost) - เป็นข้อสรุปจากชื่อ flag ยังไม่ verify หน้าจอ
- คำถามที่ยังเปิด: semantics จริงของ lock (ล็อกต้นทุนหรือล็อกแก้ไขเอกสาร); ถ้า cost engine ของ BC ต้องเคารพ lock ขนาดงานเป็น L; ใช้ doc.islocked ที่ยังว่างอยู่ได้หรือไม่

### ปรับปรุงวันที่รับสินค้าจากเจ้าหนี้ — tool, ขนาด M
- ภาพ Champ: หน้าจอเมนูจัดชื้อ\ปรับปรุงวันที่รับสินค้าจากเจ้าหนี้.jpg
- หน้าตา/ที่มา: จอเครื่องมือ: ค้นเอกสารซื้อ/รับสินค้าตามช่วง -> กริดแก้วันที่รับสินค้า/วันครบกำหนดทีละแถว แล้วบันทึก (คู่ฝั่ง AR: BILLDlgOverhaulTransport.cpp:130-131 อัปเดต DeliveryDate/DueDate ของ BCARINVOICE)
- จอ/โค้ด BC ที่ใกล้สุด: frontend/src/lib/erp-operations.ts:99-105 (delivery-dates placeholder); menu-data.ts:225,274; backend/internal/generalledger/subledger_types.go:51 (due_date)
- ลง GL/สต็อก: ไม่ลง GL/สต็อก - แก้วันที่ใน doc และ due_date ของ sub-ledger เท่านั้น
- คำถามที่ยังเปิด: ใช้ tool เดียวร่วมกับ ปรับปรุงวันที่ส่งของให้ลูกค้า (AR) ได้หรือไม่; การแก้ due_date กระทบรายงานอายุหนี้ที่คำนวณไปแล้วหรือไม่

### ยกเลิกเอกสารที่ผ่านรายการแล้ว — tool, ขนาด S
- ภาพ Champ: หน้าจอเมนูบัญชี\ยกเลิกการผ่านรายการ.jpg , หน้าจอเมนูบัญชี\เมนู-ประกอบ.jpg
- หน้าตา/ที่มา: ไม่มี jpg ชื่อ ยกเลิกเอกสารที่ผ่านรายการแล้ว ตรง ๆ (ใช้ ยกเลิกการผ่านรายการ.jpg): dialog GLDlgCancelPost ยกเลิกตามเลขชุดผ่านรายการ (BCPostTrans) GLDlgCancelPost.cpp:164,261,417,435-468; resource IDD_GL_DLG_CANCEL_POST bc5rc.utf8.txt:5594
- จอ/โค้ด BC ที่ใกล้สุด: S ถ้าแค่ชี้เมนูไป /gl/unposting: frontend/src/app/gl/general-ledger-screen.tsx:47 (/gl/unposting -> GLJournals mode=reverse), gl-journals.tsx:600-607 (reverse ต่อใบ); menu-data.ts:634 gl-post (alias :97-98 มีชื่อนี้); gl-post-batch.tsx:80-96 (tab post/reverse เป็น useState)
- ลง GL/สต็อก: Champ: ยกเลิกชุดผ่านรายการ (ถอด GL ทั้งชุด); BC: reversal เป็นรายการกลับต่อใบเท่านั้น backend post_batches.go ไม่มี action cancel (มีแค่ check/create, nextPostBatchNo :320), journals เก็บ PostBatchNo ไว้ models.go:258
- คำถามที่ยังเปิด: /gl/unposting ไม่อยู่ใน GL_MENU_ITEMS (general-ledger.ts:145-146) น่าจะตกไป not-found ที่ general-ledger-screen.tsx:52 (ยังไม่ได้รันทดสอบ); /gl/posting?tab=reverse เปิดตรงไม่ได้ (gl-post-batch.tsx:80-96 ไม่อ่าน query string) ต้องแก้เล็กน้อย; ถ้าต้องการ batch-cancel เหมือน Champ จริง = M (เพิ่ม action ยกเลิกชุดใน post_batches.go ที่สร้าง reversal ครบชุดตาม PostBatchNo)

## ค. submenu รายงานที่ยังไม่ได้ทำ

- **จัดขาย › รายงานการขายสินค้าชุด(สินค้าชุดแบบที่ 2)** — สมาชิก: ไม่ทราบ (ไม่มีหลักฐาน); ความมั่นใจ low
  - หลักฐาน: The only jpg evidence (หน้าจอเมนูจัดขาย/เมนู-ประกอบด้วย.jpg) shows the third report header 'รายงานการขายสินค้าชุด(สินค้าชุดแบบที่ 2) ▸' as the last line after 'รายงานข้อมูลรายวัน ▸' and 'รายงานเชิงวิเคราะห์ ▸'; no cascade image exists. The string appears nowhere in menuconfig.xml dump, BC5Account.rc dump, any Report_ระบบ* PDF file name (0 of 59 in Report_ระบบจัดขาย contain 'ชุด'; 0 in all 9 Report folders), or the Manual (pdftotext output is garbled). Only unproven leads in source: BillRepSaleItemSet (ID_BILLSYS_TRANS_38001 = resource.h:4093 id 5324, MainFrm.cpp:1000/9340, 2003 comment 'sale item set report') and BillRepItemSetProfit (ID_BILLSYS_TRANS_38002 = id 5325, handler MainFrm.cpp:9348 but no ON_COMMAND) in champ-report/BC5REP_IC2. They are not in menuconfig.xml and nothing ties them to the type-2 submenu, so they are NOT listed as members.

หมายเหตุจากการสำรวจ:
CHEQUE: BC (frontend/src/lib/menu-bars.ts:171) has only one report submenu labelled รายงานรายวัน (REPORTS_DAILY, menu-bars.ts:69) holding cheque-received-report / cheque-issued-report / credit-card-report (menu-data.ts:462-464). In Champ those 3 correspond to the 3 nested groups under รายงานข้อมูลหลัก (รายงานเช็ครับ 5-6 detail variants, รายงานเช็คจ่าย 6-7, รายงานบัตรเครดิต 5), so BC puts them under the wrong head; Champ's รายงานรายวัน holds 8 transaction reports BC does not have at all, and BC has no รายงานข้อมูลหลัก submenu. PDF-only names with unverified placement (do not invent): รายงานรายละเอียดเช็ครับ-ตามพนักงาน (no handler in source, no resource id), รายงานประวัติเช็ครับ (ID_MNU_CHQIN_STORY = 5683, resource.h:4452, OnMnuChqinStory MainFrm.cpp:10111; not in menuconfig.xml or .rc menus, may belong to the ข้อมูลการเคลื่อนไหวเช็ค flow rather than any report submenu), and PDF name 'รายงานรายละเอียดบัตรเครดิต-ตามวันที่รับ' vs XML name '...-ตามวันที่จ่าย' (same MASTERREP_3_1, label differs between stale XML and the 5.5 PDF). Champ XML marks both cheque report popups disable=yes, so presence in the real 5.5 build is confirmed only by the dropdown headers + PDF names. BILL: BC menu-bars.ts:122-129 has รายงานข้อมูลรายวัน (sales, sales-daily, sales-by-document, sale-return-report) and รายงานเชิงวิเคราะห์ (sales-by-seller, gross-profit-by-document, gross-profit-by-product); both heads exist in Champ's dropdown so those items are in the right heads (per stale XML champ-menuconfig.txt:672-735 รายงานยอดขายสุทธิ-เรียงตามพนักงานขาย is under เชิงวิเคราะห์, รายงานรายวันขายสินค้าและบริการ under ข้อมูลรายวัน). BC has no submenu for the type-2 set sales reports and no BC item belongs in it. memory-hub search found no earlier record of the type-2 report members. To resolve the BILL gap, ลุงจืด would need a screenshot of the opened cascade of that submenu in Champ 5.5."
