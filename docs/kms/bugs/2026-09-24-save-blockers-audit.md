---
date: 2026-09-24
severity: high  # low | medium | high | critical
component: [backend, frontend, db]
tags: [bc-account, go, postgres, gl, auth, i18n]
fixed: true
---

# ตรวจรอบ "บันทึกไม่ได้" 2026-09-24 — GL, ภาษี, บัญชีเข้าระบบ และ master

รอบนี้ไล่หาทุกจุดที่ผู้ใช้กดบันทึกแล้วไม่ผ่าน (หรือผ่านแต่ข้อมูลผิดเงียบ ๆ) แยกเป็นรายข้อ: อาการ → ต้นเหตุ → วิธีแก้ → เทสต์กันถอยหลัง
สัญญา API ที่เปลี่ยน: ADR `decisions/2026-09-24-user-defined-journal-books.md` และ `decisions/2026-09-23-wht-tax-base-editable.md` §3

## A. สมุดรายวันและรหัส (backend GL)

1. **สร้างสมุด/รหัสภาษาไทยไม่ได้** (เช่น `สมุดซื้อ`, `ค่าน้ำ/ไฟ`)
   - ต้นเหตุ: regexp `^[\p{L}\p{N}]…` — สระ/วรรณยุกต์ไทยเป็น combining mark (`\p{M}`) ไม่ใช่ `\p{L}` (กับดักใน `17-dev-gotchas.md`)
   - แก้: `checkCode` (`backend/internal/generalledger/models.go:268`) — trim + NFC, ตัวแรกต้องเป็นตัวอักษร/ตัวเลข, ตัวถัดไปรับ `\p{L}\p{M}\p{N}` และ `_ . - / # ( ) :`, ห้ามช่องว่าง/อักขระมองไม่เห็น, นับความยาวเป็น rune ตาม DDL ใน mydocs (สมุด 15, เลขที่เอกสาร 30, รหัสบัญชี 20, ชื่อบัญชี 150); error แยกรายช่อง `code_*` + `field`
   - ผลพลอยได้: request id 64 ตัวอักษรของสินทรัพย์ถาวร (SHA-256 hex) ที่เคยถูกปฏิเสธ ผ่านแล้ว (`validRequestID` :299)
   - เทสต์: `TestCheckCodeAcceptsThaiCodes`, `TestCheckCodeRejectsWithFieldAndReason`, `TestCodeLimitsCountRunes`, `TestNormalizeCodeTrimsAndComposes` (`journal_books_test.go`)
2. **สมุดตายตัว + SV/UV ความหมายสลับ + ประมวลผลยึดรหัส `JV`** — บริษัทที่ตั้งรหัสเองปิดบัญชี/ลงสินทรัพย์ถาวรไม่ได้
   - แก้: master ผู้ใช้กำหนดเองมี `booktype` 1–6, โค้ดเลือกสมุดตามประเภท (ADR ข้างบน)
   - เทสต์: `TestValidateJournalBookMaster`, `TestDefaultJournalBooksFollowSpecMeaning`, `TestChooseBookCodeByType`, `TestJournalBooksLifecycleIntegration`
3. **ข้อความมี NUL (U+0000) → HTTP 500** — PostgreSQL `jsonb` ไม่รับ `\u0000` (มักติดมาจากการคัดลอก PDF)
   - แก้: `rejectNUL` (`models.go:466`) ตรวจทุกช่องข้อความก่อนเขียน ตอบ 400 `text_contains_nul` พร้อม `field` เช่น `lines[2].description`
   - เทสต์: `TestRejectNULNamesFieldAndReturns400`
4. **ใบใหม่จาก session ระดับบริษัทส่ง `branchcode: ""`** — ไม่มีการตรวจสาขา / สาขาปลอมผ่าน
   - แก้: `checkJournalBranch` (`httpapi/branch.go:17`) หลังตรวจสิทธิ์: ระดับสาขา = ว่างหรือสาขาตัวเอง, ระดับบริษัทต้องส่งสาขาที่ active
   - เทสต์: `httpapi/branch_test.go`
5. **`postgres_store_test.go` ต่อฐาน `appdb` ที่ 127.0.0.1:5432 เมื่อไม่ตั้ง DSN** (เป็น postgres ของ tenant จริงบนเครื่อง dev) — แก้เป็น `t.Skip`

## B. รายละเอียดภาษีและคู่ค้า (backend GL subledger)

1. **ใบร่างที่ฝังข้อมูลคู่ค้าฉบับเก่าบันทึก/ผ่านรายการไม่ได้** ("ถูกแก้ไขแล้ว กรุณาโหลดใหม่") แม้ผู้ใช้ไม่ได้แตะแถวคู่ค้า
   - ต้นเหตุ: optimistic lock เทียบ `version` ของทุกแถวกับทะเบียน รวมแถวที่ไม่เปลี่ยน
   - แก้: แถวที่เหมือน snapshot เดิมของใบ = ไม่เขียนทะเบียน (`subledgerMutation.previousPartners`); `version 0` = บันทึกเป็นค่าล่าสุด; ชนจริงได้ `partner_version_conflict` ที่บอกวิธีแก้; เพิ่มบทบาท/แก้เลขภาษีได้หลังมีเอกสาร แต่ยกเลิกบทบาทที่ยังมีเอกสารไม่ได้
   - เทสต์: `TestPostgresPartnerSnapshotAndTaxRowRoundTrip` (ปิดโค้ดข้อนี้ชั่วคราวแล้วเทสต์ล้มที่ขั้นที่ 4 จริง)
2. **ชื่อภาษาไทยยาว ~85 ตัวอักษรถูกปฏิเสธ** — นับความยาวเป็น byte (ไทย 3 byte/ตัว) → นับเป็น rune ทุกช่อง (`TestStatementAndWithdrawalLengthsCountRunes`, `TestPostgresThaiTextLengthsCountedInRunes`)
3. **เลขผู้เสียภาษีแบบมีขีด (`0-1055-12345-67-8`) และสาขา `0` ถูกปฏิเสธ** → ตัดขีด/ช่องว่างเหลือ 13 หลัก, สาขาเติม 0 ซ้ายให้ครบ 5 หลัก ทั้งคู่ค้า VAT และ snapshot ภาษีหัก (`TestNormalizeTaxIDAndBranch`, `TestValidatePartnerReportsEachField`)
4. **ไม่ส่ง `wht_rate` → กลายเป็นหัก 0% ภาษี 0 เงียบ ๆ** (การ clone ผ่าน JSON แปลงค่าว่างเป็น `"0"`) → บังคับส่ง `wht_rate_required` (`TestCloneJournalDetailsRejectsMissingWithholdingRate`, `TestPostgresWithholdingWithoutRateRejected`)
5. **ลบแถว VAT/ภาษีหักแถวสุดท้ายของใบที่ผ่านบัญชีไม่ได้** — `[]` ถูกตีความเหมือน "ไม่ส่ง" → อ่านคีย์ก่อน clone: `[]` = ล้าง, ไม่ส่งคีย์ = คงเดิม, audit `withholding_replace`/`vat_replace` ที่ after=[] (`TestExplicitEmptyTaxArraysSurviveDecode`)
6. **50 ทวิ และรายงาน ภ.ง.ด. ใช้ทะเบียนคู่ค้าปัจจุบัน** — แก้ทะเบียนภายหลัง เลขภาษีของใบเก่าเปลี่ยนตาม
   - แก้: snapshot `payer_*`/`payee_*` ตาม `mydocs/datamodels/gl/wht.sql` ในรายการภาษีหัก (ฝั่งคู่ค้าเติมจากทะเบียน ณ วันบันทึกเมื่อว่างทั้งชุด); ใบรับรองเลือกค่าทีละช่อง snapshot → ทะเบียน → ค่าจากจอ; รายงานภาษีหักใช้ snapshot ของคู่ค้าก่อนทะเบียน (`applyWithholdingPartySnapshot` `backend/internal/goapi/handlers/tax_report.go`) และส่ง `withholdingid` ให้จอจับคู่รายการตรงตัว
   - เทสต์: `TestWithholdingSnapshotNormalizedAndValidated`, `TestFillPartnerSnapshotBySide`, `TestApplyRecordedWithholdingPrefersSnapshotThenRegisters`, `TestWhtCertificateHandler_UsesRecordedWithholding`, `TestApplyWithholdingPartySnapshotPrefersRecordedParty`

## C. หน้าจอ GL (frontend)

1. **กด Tab ผ่านช่องยอด 0.00 แล้วค่ากลายเป็น "" → backend ปฏิเสธทั้งใบ**; ช่องยอดปัดเศษ/ตัดเครื่องหมายลบเงียบ ๆ
   - แก้: `AmountInput` (`frontend/src/app/gl/gl-common.tsx`) คงค่า `"0"`, ไม่ปัดเศษ ไม่ตัดลบ แต่แจ้งใต้ช่อง, แปลงเลขไทย ๐–๙; `normalizeJournalLines` เปลี่ยนช่องว่างเป็น `"0"` ก่อนส่ง
   - เทสต์: `src/app/gl/amount-input.test.ts` (27), `src/lib/general-ledger.test.ts`
2. **ช่องเลขผู้เสียภาษี `maxLength=13` ตัดเลขที่วางแบบมีขีด** และส่งยอดภาษีเป็น `""` → ตัดขีดเหลือตัวเลข บอกจำนวนหลักที่ขาด, ยอดว่างไม่ส่ง (ให้ backend คำนวณ), อัตราบังคับกรอก (`src/lib/gl-journal-details.test.ts`)
3. **วางจาก Excel ทิ้งแถวที่มีแต่เดบิต และแปลง `(1,500.00)`/`-1500` เอง** → เก็บฝั่งที่มี ฝั่งว่าง = `"0"`, ยอดติดลบ/วงเล็บถูกปฏิเสธพร้อมบอกแถว-ช่อง (`src/lib/clipboard-journal-parser.test.ts` ใช้ TSV จริงจาก Excel)
4. **ช่องเลือกบัญชีแม่ให้เลือกตัวเองหรือบัญชีลูกหลาน → ผังบัญชีวนเป็นวง** → เฉพาะบัญชีคุมที่เปิดใช้งาน ไม่รวมตัวเองและลูกหลาน (`src/app/gl/gl-masters.test.ts`)
5. **ใบใหม่ไม่เลือกปีบัญชีตามวันที่** → เลือกปีที่เปิดอยู่และครอบคลุมวันที่ เปลี่ยนตามเมื่อแก้วันที่
6. **จอ 50 ทวิ เก็บที่อยู่ผู้หักใน localStorage** → อ่าน snapshot จากใบสำคัญ แก้แล้ว "บันทึกลงใบสำคัญ" ผ่าน update (ร่าง) / reconcile (ผ่านบัญชีแล้ว) พร้อม dialog ยืนยัน (`src/app/tax/wht-certificate-panel.test.ts`)
7. **BFF GL ตอบ 404 กับ path ที่มีรหัสไทยมีสระ/วรรณยุกต์** — regex segment `[\p{L}\p{N}_.-]` ขาด `\p{M}` (`frontend/src/app/api/gl/[...glPath]/route.ts`, `api/integration/gl/[...path]/route.ts`) → เพิ่ม `\p{M}` (`route.test.ts` "accepts Thai codes with vowel signs and tone marks")

## D. บัญชีเข้าระบบ / master บทบาท-ประเภทธุรกิจ

1. **วันหมดอายุการเข้าใช้งานไม่เคยมีผล** — ค่าไม่ถูกอ่านจาก DB; วันที่ผิดรูปแบบ → 500
   - แก้: `holding_members.access_expiry_date DATE` = **วันสุดท้ายที่ใช้งานได้ (นับรวมวันนั้น)** — ปิดสิทธิ์ 00:00 ของวันถัดไปตาม timezone ของ holding (`authmodels.AccessEndsAt`/`AccessExpired`, `centraldb.AccessExpiryInstant`; รอบ E แก้จากเดิมที่ปิดตั้งแต่ 00:00 ของวันนั้น ซึ่งขัดกับป้ายจอ "ใช้งานได้ถึงวันที่"), ตอบ 403 `user_access_expired`; วันที่ผิด = 400 `ss_err_access_expiry_date_invalid`; ผู้สร้าง holding ตั้งวันหมดอายุไม่ได้
   - เทสต์: `TestAccessShopRefusesMembershipOnItsExpiryDate`, `TestAccessExpiryInstant`, `TestShopUserCreatorCannotGetAccessExpiry`, `TestPostgresSaveFullProfilePersistsMembershipFields`
   - รอบ E ปิดครบทุกทางที่อ่านสมาชิกภาพ: `FindActiveMembership` (+ `FindActiveHoldingManager` ที่ MCP token ใช้ทั้งตอนออกและทุกครั้งที่ใช้), shop/AccessShop, `rolepermission`, GL `httpapi/scope.go` (`resolveScope`)
2. **ตำแหน่ง แผนก รูป (avatar/avatarthumb) ไม่ถูกบันทึก** → คอลัมน์ใหม่ใน `holding_members` (ไม่ส่งรูป = คงรูปเดิม)
3. **เพิ่มผู้ใช้ที่มีอยู่แล้วทับข้อมูลเงียบ ๆ** — ทับอีเมล/ชื่อบัญชีกลางของ holding อื่น และทับ role/สิทธิ์สมาชิกเดิม (`ON CONFLICT DO UPDATE`)
   - แก้: แยกเพิ่ม/แก้ด้วย `editusername`/`useruid`; 409 `ss_err_user_already_member` / `ss_err_user_code_taken` (เว้นแต่ `addexistinguser: true`) / `ss_err_user_code_locked` / `ss_err_user_email_locked`; แก้สมาชิกที่ลบแล้ว = 404
   - เทสต์: `TestShopUserAddRejectsExistingMember`, `TestShopUserEditOfMissingMemberIsNotFound`, `TestPostgresSaveFullProfileDuplicateAndExistingAccount`
4. **Google login สร้างบัญชีใหม่เสมอ** — บัญชีที่ admin สร้างไว้ล่วงหน้าใช้ไม่ได้ → ผูกเมื่ออีเมลยืนยันแล้วตรงบัญชีเดียวที่ยังไม่ถูก claim (`TestPostgresGoogleLinksPrecreatedAccountByEmail`, `TestPostgresGoogleNeverTakesOverOrMergesAccounts`, `TestPostgresGoogleConcurrentPrecreatedLink`)
5. **สร้างประเภทธุรกิจ/ชุดสิทธิ์รหัสซ้ำ = upsert ทับของเดิม; แก้/ลบโดนหลายแถวที่ต่างแค่ตัวพิมพ์; ตอบ 200 แม้ไม่มีแถวตรง** → INSERT ธรรมดา + 409 เมื่อซ้ำ (ไม่สนตัวพิมพ์), revive แถวที่ลบแล้วแถวเดียว, แก้/ลบเลือกแถวเดียว ไม่งั้น 409 ambiguous / 404 (`TestPostgresBusinessTypeCreateNeverOverwrites`, `TestPostgresRolePermissionLegacyCaseDuplicates`)
6. **บันทึกแบบยื่นภาษีล้มเมื่อรหัสบริษัท/ชื่อผู้ใช้ยาวเกินคอลัมน์** — `tax_filings.company_code VARCHAR(20)`, `created_by/updated_by/saved_by VARCHAR(100)` → TEXT (DO block ตรวจ `information_schema` ก่อน ALTER, ข้อมูลเดิมอยู่ครบ) (`TestTaxFilingColumnsWidenToText`)

## E. รอบรวมงาน 2026-09-24 (GL backend / ภาษี-ไฟล์ยื่น-ฐานกลาง / frontend)

1. **ใบลดหนี้ฝั่งซื้อถูกบล็อกด้วยกรอบ 6 เดือน** — ใบลดหนี้ต้องลดภาษีซื้อ "ในเดือนภาษีที่ได้รับ" (ม.82/10, ป.80/2542 ข้อ 5) ไม่ใช่สิทธิที่เลื่อนได้
   - แก้: `checkPurchaseClaimWindow` แยกตาม `document_type` — ใบลดหนี้ตรวจแค่ "งวดไม่ก่อนเดือนของใบลดหนี้" (`vat_credit_note_period_before_note`); ใบกำกับ/ใบเพิ่มหนี้ใช้ 6 เดือน; จอ `vatClaimTiming` ตามเดียวกัน (`credit_note_before`) — ทะเบียน `docs/kms/21` §10
   - เทสต์: `TestVatNewRowRulesByDocumentType`, `frontend/src/lib/gl-journal-details.test.ts`
2. **ปีงวดภาษีกรอกเป็น พ.ศ. (2569) ผ่านเข้าไปเป็นงวดในอนาคต** → ปี ≥ 2400 ปฏิเสธ `vat_period_year_buddhist` ไม่แปลงให้เอง; จอเตือน `buddhist_year`
3. **กฎใหม่ทำให้ใบร่างเก่าแก้/ลบ/กลับรายการไม่ได้** → ตรวจเฉพาะแถวใหม่หรือแถวที่แก้ (`vatRowNeedsRules`, ความยาวข้อความเฉพาะข้อความใหม่), ผ่านรายการตรวจทุกแถว (`TestVatUnchangedRowSkipsNewRules`, `TestJournalTextLimitsCheckOnlyNewText`)
4. **ใบกำกับเดียวกันหลายแถวในใบสำคัญเดียวไม่ถูกเตือน** → `DuplicateDocNos` รวมเลขของใบนี้เอง (สถานะใช้สิทธิเดียวกัน) (`TestPostgresVatSameVoucherDuplicate`)
5. **เลขผู้เสียภาษีพิมพ์ผิดหลักผ่านได้** → หลักตรวจสอบ mod 11 ทุกจุดที่รับเลข 13 หลัก (แถวใหม่/แก้เท่านั้น); ว่างยังผ่าน; ภ.ง.ด.2 ใช้เลขศูนย์ได้ (`TestPostgresWithholdingTaxIDChecksum`, `TestPostgresPartnerRDFieldsAndChecksum`) — `docs/kms/21` §12
6. **สมุดที่ถูกอ้างใน "รูปแบบการเชื่อมบัญชี" ลบ/เปลี่ยนรหัสได้** → `journalBookInUse` นับ `mappings` ด้วย ข้อความบอกทั้งสองสาเหตุ (`TestJournalBookInUseByMapping`); สมุดเก่าไม่มีประเภทเปลี่ยนชื่อ/ปิดใช้งานได้โดยไม่ต้องเลือกประเภท + แถบเตือนรายชื่อในหน้า "กำหนดสมุดรายวัน"; บริษัทที่มีใบสำคัญแต่ไม่มี record สมุดไม่ได้สมุดเริ่มต้นอัตโนมัติ (ห้ามเดาประเภทจากรหัส) — ADR `decisions/2026-09-24-user-defined-journal-books.md`
7. **รายงานภาษีหัก: ใบที่กลับรายการเดือนหลังหายจากเดือนที่ยื่นไปแล้ว** → คงไว้ในเดือนที่จ่าย + `reversedmonth` + หมายเหตุแถว (จอรายงานแสดงหมายเหตุใต้ประเภทเงินได้); บัญชีที่ใช้ได้ทั้ง ภ.ง.ด.3/53 ไม่เดาแบบ (`unknownform`) (`TestWithholdingReportReversalsAndForms2026_09_24`, `frontend/src/lib/thai-tax.test.ts`) — `docs/kms/21` §9
8. **50 ทวิ ของรายการที่บันทึกแก้ยอด/วันที่ได้ และ backend ปฏิเสธด้วย `wht_cert_record_mismatch`** → อ้างรายการแล้ว backend พิมพ์ค่าที่บันทึกเสมอ, จออ่านอย่างเดียว, มีค่าผู้จ่าย/ผู้รับแก้ค้างต้องบันทึกลงใบสำคัญก่อนสร้าง PDF; ตรวจสิทธิ์บริษัท/สาขาจากสมาชิกภาพจริงทุกคำขอ (403)
9. **สร้างไฟล์ยื่นด้วยสื่อ** (ใหม่) — ADR `decisions/2026-09-24-rd-wht-file-export.md`: ยอดหัวแบบ ภ.ง.ด.2 นับเฉพาะรายการแรกของบรรทัด, ACC_NO บังคับเฉพาะเงินปันผล, ยื่นหัวแบบอย่างเดียวเมื่อไม่มีใบแนบและยอดศูนย์, ตัดช่องว่างหัวท้ายทุกช่อง (`backend/internal/rdfile/rdfile_test.go`, `handlers/tax_rdfile_test.go`)
10. **ผู้ใช้หมดอายุยังเรียก API ของ GL ได้** → `resolveScope` ตรวจ `access_expiry_date` ตอบ 403 `user_access_expired` (`TestScopeAccessExpiry`, integration "access expired" ใน `scope_postgres_integration_test.go`)
11. **ชุดสิทธิ์แก้ทับกันได้ / ประเภทธุรกิจเขียนได้ทุกคน / เพิ่มผู้ใช้ที่เคยถูกลบกลับไม่ได้** → `role_permissions.version` (409 `ss_err_permission_set_changed`), `business_types` เขียนได้เฉพาะ OWNER/ADMIN, เพิ่มบัญชีที่ถูกลบออกจากกลุ่มนี้กลับได้ (บันทึกการลบใน `organization_audits`); จอเพิ่มผู้ใช้ที่มีบัญชีอยู่แล้วขึ้น dialog ยืนยันก่อนส่ง `addexistinguser`
12. **authentication log body/รหัสผ่าน** → log เฉพาะชนิดข้อมูล
13. **ช่องยอดเงิน: คำเตือน "ตัดตัวอักษร" ค้างใต้ค่าที่ถูกต้องแล้ว** → ผูกคำเตือนกับค่าที่ทำให้เกิด (`droppedWarningShown`) (`src/app/gl/amount-input.test.ts`)
14. **ข้อความ/แถวภาษาที่ตายเพราะรอบนี้** ลบออกจาก `languages.tsv`: `ss_f_access_expiry_date`, `ss_h_auto_disable_access_on_this`, `tax_form_drift_more`, `gl_err_partner_code_invalid`, `wht_cert_record_mismatch`

## Regression Test (รวม)

- `cd backend && go test ./internal/generalledger/... ./internal/goapi/handlers/... ./internal/shop/... ./internal/authentication/... ./internal/centraldb/... ./internal/organization/...`
- integration กับ PostgreSQL ชั่วคราว: `npm run verify:all` (target `postgres`)
- frontend: `cd frontend && npx vitest run src/app/gl src/app/tax src/lib/general-ledger.test.ts src/lib/gl-journal-details.test.ts src/lib/clipboard-journal-parser.test.ts "src/app/api/gl"`
- ข้อความทุก error มีแถว `gl_err_<code>` / `ss_err_*` ใน `languages.tsv` — `TestGLUserErrorCodesHaveEnglishRows`, `gl-language-keys.test.ts`, `wht-certificate-language-keys.test.ts`
