# Handoff ให้ GLM — ระบบบัญชีแยกประเภท BC Ai Account

บันทึก 20 กันยายน 2569 หลัง deploy `r20260920-gl-audit-fix-2` และตรวจรับแล้ว. สถานะด้านล่างอ้างอิงรอบตรวจที่ผ่านมา ไม่ใช่คำรับรองว่าไม่มีบั๊กอื่น. งานรับช่วงคือทบทวนและแก้จากโค้ดปัจจุบัน ไม่เริ่มระบบใหม่.

## 1. เริ่มอ่านและขอบเขตงาน

Workspace: `D:\bccode`, branch `dev`, HEAD ณ handoff `365ea413`. Working tree มีทั้ง modified และ untracked จำนวนมากจากหลายงาน; งาน GL รอบนี้ยังไม่ได้ commit/push. HEAD อย่างเดียวไม่มีงานล่าสุดครบ.

อ่านตามลำดับ เฉพาะส่วนที่เกี่ยวกับงาน:

1. `AGENTS.md` — กฎโปรเจ็กต์ล่าสุด รวมข้อจำกัด `mydocs/`, การตรวจรับและ deploy.
2. เอกสารนี้ — สถานะรับช่วงและลำดับงานที่เสนอ.
3. `docs/audits/GL-FIX-2026-09-20.md` — ผลแก้ไขและหลักฐานรอบล่าสุด.
4. `docs/kms/architecture/gl-journal-details.md` — runtime contract ของ AR/AP/Statement/source dedup.
5. `docs/kms/architecture/gl-mcp-tokens.md` และ `gl-journal-review.md` เมื่อแตะ token/การตรวจเอกสาร.
6. `docs/audits/GL-AUDIT-2026-09-20.md` — findings เดิม ใช้เป็นรายการ regression; ห้ามถือว่าทั้ง 13 ข้อยังไม่ได้แก้.
7. ตรวจ source/tests/schema ของเรื่องที่กำลังแก้; `docs/reference/CODE-MAP.md` ช่วยหาไฟล์ใหญ่.

`docs/handoff/HANDOFF-2026-09-20-GL-TAX.md` เป็นแผนก่อนหน้า มีหัวข้อ RD Prep/XBRL/ภาษีและชื่อตารางออกแบบที่ไม่ตรง runtime ปัจจุบันทั้งหมด. อย่าใช้แผนนั้นขยายงานเอง. Runtime GL ปัจจุบันใช้ `gl_records`, `gl_lines`, `gl_events`, `gl_projection_state` และ `gl_subledger_*`; ตรวจ `backend/internal/generalledger/schema.sql` และ `subledger.sql` โดยตรง.

## 2. ข้อกำหนดของลุงจืดที่ต้องรักษา

- GL เป็นระบบห้องบัญชี เน้นตรวจเอกสาร รับข้อมูลจากระบบอื่น ทำงานคนเดียวได้ ไม่ต้องพึ่ง workflow หน้าบ้านหรือโมดูลขาย/ซื้อ/คลัง/ลูกจ้าง.
- บันทึกรายวันและหลักฐานลูกหนี้ เจ้าหนี้ ธนาคารในใบสำคัญเดียว รองรับ many-to-many, ชำระ/จับคู่บางส่วน และดูยอดค้าง/ประวัติตรวจ.
- RD Prep และ XBRL พักไว้ ไม่เพิ่มเมนูหรือฟีเจอร์นอกคำสั่ง; เทียบ Champ เมื่อออกแบบสิ่งใหม่ตามกฎโปรเจ็กต์.
- ไม่เพิ่ม MongoDB, ClickHouse, Kafka หรือ Redis กลับมาในเส้นทางนี้. ข้อความ/โค้ดเก่าที่พูดถึงระบบเหล่านี้ไม่ใช่อนุญาตให้เปิดใช้อีก.
- API token และ MCP token คนละ credential. ผู้ดูแล Holding จัดการ token, เลือกบริษัทที่อนุญาตต่อ token; readonly อ่านหลายบริษัทได้, readwrite ทุกคำขอเลือกบริษัทเดียว แม้เป็นคำขออ่าน.
- บัญชีผู้พัฒนา `jaturapornchai@gmail.com` เคยได้รับคำสั่งให้เข้าได้ทุกจอ; กลไก identity/UID และ membership ฝั่ง server ถูกแก้แล้ว. ห้ามแก้ปัญหาด้วยการเชื่อ role/email ที่ browser ส่ง หรือคืน OWNER เมื่อหาสมาชิกไม่พบ.
- จำนวนเงินเป็น decimal strings/NUMERIC หรือ minor units เท่านั้น ห้าม float. ต้องสมดุล debit=credit, optimistic version, requestid และ audit trail.
- `mydocs/` เป็นข้อกำหนดของลุงจืด อ่านอย่างเดียวตามกฎล่าสุด. คำสั่งเก่าที่เคยให้แก้ mydocs ไม่ใช้ลบล้างข้อห้ามล่าสุดนี้. ข้อเสนอใหม่เขียนนอก mydocs.
- ห้าม pull/reset/checkout จาก remote หรือ stage/commit งานอื่นรวมเอง; local working tree คือของจริง. ห้าม push หากยังไม่ได้รับคำสั่ง.

## 3. สิ่งที่ทำแล้ว — อย่าแก้ซ้ำด้วยการถอด guard

| กลุ่ม | พฤติกรรมที่มีแล้ว | จุดเข้าโค้ด |
|---|---|---|
| A1–A4 สิทธิ์ | missing/inactive user/member/Holding/company/role ถูกปฏิเสธ, ไม่มี OWNER fallback; ตรวจสิทธิ์ล่าสุดทุกคำขอ | `backend/internal/shop/shopuser_postgres_repository.go`, `backend/internal/organization/rolepermission/postgres_authorization.go`, `backend/internal/generalledger/httpapi/scope.go` |
| G1–G4 รายการ | post เฉพาะ draft; reversed แก้/ลบ/post ซ้ำไม่ได้; บังคับ branch, ปีปิด, งวดล็อก, บัญชี active; reversal รักษาสาขา/มิติ | `backend/internal/generalledger/postgres_store.go`, `postgres_guards.go` |
| G5–G6 ประมวลผล | close/year-end สร้างรายการจริงแบบ draft, rebuild replay audit ใน transaction, lock/unlock ใช้งวดเดิม ID/version/reason | `backend/internal/generalledger/postgres_processes.go` |
| Report | gljournal SQL เรียงถูก, drill ตาม journalid, ย้อนด้วย applied filters | `backend/internal/generalledger/reports.go`, `frontend/src/app/gl/gl-reports.tsx` |
| Source dedup | unique company + source_system + source_record_id; requestid ใหม่/payload เดิมคืนใบเดิม, payload ต่างปฏิเสธ | `backend/internal/generalledger/postgres_source.go` |
| หลักฐาน | เอกสาร AR/AP, allocations, settlements, bank_lines, Statement และ matches หลายต่อหลาย มี withdrawal/audit | `backend/internal/generalledger/subledger*.go`, `subledger.sql` |
| หน้ารายวัน | details อยู่ในฟอร์มเดิม, dirty guard, picker อ่านชื่อ, CSV Statement, posted reconcile ส่งเฉพาะ delta | `frontend/src/app/gl/gl-journals.tsx`, `gl-journal-details.tsx`, `frontend/src/lib/gl-journal-details.ts` |
| ช่องเชื่อม | browser/API/MCP เข้า backend GL ชุดเดียว, ตรวจชนิด token และ company allowlist | `backend/internal/generalledger/httpapi/`, `backend/internal/mcptoken/`, `backend/internal/mcpgateway/` |

### Guard ที่สำคัญเป็นพิเศษ

- `CompanyWide` แยกจากสาขาที่เลือกบนจอ. Process ตรวจสิทธิ์ก่อนขยายเป็นทั้งบริษัท; preview ส่ง `companywide=true`. การอ่าน/แก้ใบสำคัญปกติยังจำกัดตามสาขา.
- Legacy token ที่มี BranchUID ห้ามขยายเป็นทั้งบริษัท แม้ผู้ออกเป็น OWNER. Regression: `TestCompanyScopePreservesLegacyTokenBranch`.
- Reconcile หลัง posted เปลี่ยนหลักฐานได้ แต่เปลี่ยนยอด GL ไม่ได้. ย้าย allocation ต้องถอน/สร้างใหม่ใน transaction เดียวและยอดยังครบ; ไม่เปิดให้ patch การเงินทั่วไป.
- ก่อน reverse ต้องถอน settlements/matches ที่เกี่ยวข้อง. การเปลี่ยนหลักฐานเพิ่ม version และให้ตรวจใหม่ทุกใบที่ได้รับผลกระทบ.
- Draft ที่เป็นเจ้าของ Statement ย้ายสาขาไม่ได้ แม้เอา inline details ออกจาก payload. การเพิ่ม allocation/settlement/match/Statement ต้องตรวจ master active. การถอน settlements/matches เดิมยังใช้ guard สิทธิ์/สถานะ/งวด/เหตุผล ไม่บล็อกเพียงเพราะ master ปิดใช้; allocation replacement ยังต้องตรวจ active ของรายการใหม่ตามเดิม.
- ฐาน `gl_subledger_*` และ source registry ไม่อิง FK ไป cache `gl_records/gl_lines` ที่ถูก rebuild; rebuild ต้องรักษาหลักฐานและ audit เดิม.
- Source identity เปลี่ยนหลังสร้างไม่ได้; hash สร้างจากสำเนา payload ก่อน normalization. อย่าแก้ slice ของ caller แล้วค่อย hash.
- year-end opening ที่ระบบสร้างป้องกันด้วย event/ID จริง ไม่เชื่อเพียงข้อความ reference ขึ้นต้น YEAR-END.
- AR/AP reports แสดงยอดสุทธิ เพิ่มหนี้บวก ลดหนี้ลบ. Support list แสดงยอดบวกแต่ละเอกสารสำหรับเลือกตัดยอด.
- `bank-unmatched` คือ Statement ที่ยังไม่จับคู่; GL ฝั่งธนาคารดู support kind `bank-lines`. อย่าสรุปว่า report นี้รวมสองฝั่งแล้ว.

## 4. API/MCP ที่ผู้รับงานควรใช้

| ทางเข้า | URL / contract |
|---|---|
| Browser session | BFF `/api/gl/...` → backend `/gl/v2/...` |
| REST API token | public `/api/integration/gl/...` → backend `/integration/gl/v2/...` |
| MCP token | `https://account.bcaicloud.com/mcp/gl` แบบ Streamable HTTP |
| หน้าจัดการ token | `https://account.bcaicloud.com/mcp-tokens` ระดับ Holding |

REST public path **ไม่มี `/v2`** ต่อท้าย `/api/integration/gl`. Token ส่งใน Authorization header ไม่ใส่ query/log/file. API ใช้ `X-BC-Company-Code` หรือ readonly `X-BC-Company-Codes`; MCP ใช้ companyCode/companyCodes ใน arguments.

- Support GET `/journal-support?kind=documents&page=1&limit=50`; kinds: partners, bank-accounts, documents, allocations, settlements, statements, bank-lines, matches. รองรับ q/asof.
- Reports เพิ่ม: ar-outstanding, ap-outstanding, bank-unmatched. Process preview workingpaper/trialbalance ใช้ companywide=true เมื่อมี grant.
- Write POST `/command`; `resource=journals`, `action=create/update/post/reverse/reconcile/review` ตาม service ที่รองรับ. ทุก action ยังต้องผ่านสิทธิ์/สถานะ/version ตาม backend.
- ตัวอย่าง payload และข้อจำกัดอยู่ใน `gl-journal-details.md`; อย่าสร้าง endpoint ใหม่เมื่อมี contract นี้แล้ว.

## 5. Production ณ รอบตรวจล่าสุด

ตรวจจริง 20 กันยายน 2569 ประมาณ 21:34–21:36 น. (Asia/Bangkok); ต้องอ่านสถานะใหม่ก่อน deploy ครั้งถัดไป.

- URL: `https://account.bcaicloud.com`.
- Host: `root@159.223.43.229`; ห้ามพิมพ์ secret/config ทั้งก้อนลง log.
- mainapi และ frontend: `r20260920-gl-audit-fix-2`; PostgreSQL 18 Alpine.
- Container names: `bcai-account-mainapi-1`, `bcai-account-frontend-1`, `bcai-account-postgres-1`.
- ฐานกลาง: `bcai_projection`; Holding DB: `rungrueng`.
- Release env: `/etc/bcai-account/release.env`; compose: `/opt/bcai-account/deploy/compose.yml` + `compose.8gb.yml`.
- ทั้งสาม healthy, restart 0; ไม่พบ error markers ที่ตรวจใน log ล่าสุด. Public root 200, GL ไม่มี session ได้ 401.

Baseline บริษัท 01 / Holding rungrueng / สาขา 00000 / ปี 2569:

| รายการ | ค่าที่ตรวจจริง |
|---|---:|
| ใบสำคัญ active | 144 |
| draft | 144 |
| posted GL lines | 0 |
| เดบิต | 6,889,798.48 |
| เครดิต | 6,889,798.48 |
| projection sequence | 240 |
| ตาราง subledger | 9 |

144 ใบเป็นข้อมูลสังเคราะห์ที่ผู้ใช้เคยสั่ง เดือนละ 12 ใบ ใช้คำอธิบายปกติ ไม่มีคำว่า “ข้อมูลตัวอย่าง” บนเอกสาร. ไม่ใช่หลักฐานธุรกรรมจริง; provenance อยู่ `docs/examples/gl-rungrueng-01-2569-20260920-v1.json`. ยังไม่มี details ผูก จึงไม่มี AR/AP/Statement support records และรายงาน posted ยัง 0 แถว. **ห้าม post/reverse/delete 144 ใบเพื่อทำ UAT และห้ามสร้าง Statement ให้ดูเหมือนหลักฐานจริงจากคำอธิบายเดิม.**

Backup ที่ตรวจ checksum แล้ว:

- `/opt/bcai-account/releases/r20260920-gl-audit-fix-2/backup/postgres-all.sql` (2,423,229 bytes).
- `/opt/bcai-account/releases/r20260920-gl-audit-fix-2/backup/runtime-config.tar.gz` (7,475 bytes).
- Manifest `preflight-backup.json`, `release.env.before` และ image รุ่นก่อน `r20260920-gl-audit-fix-1` ยังมี ณ เวลาตรวจ.
- Backup ก่อนเพิ่ม schema อยู่ release `r20260920-gl-audit-fix-1`; เก็บไว้ด้วย. Rollback image ไม่เท่ากับ rollback schema; อย่า drop ตารางหลักฐาน.

## 6. ผลทดสอบที่มีแล้ว และสิ่งที่ยังพิสูจน์ไม่ครบ

**ผ่านในรอบก่อน handoff:**

- Frontend Vitest 97 files / 723 tests, TypeScript และ Next production build; Docker build ทั้ง frontend/backend.
- ESLint ทั้ง repo 0 errors / 284 warnings เดิม; ไม่ได้เคลียร์ warnings ทั้งโปรเจ็กต์.
- PostgreSQL 18 integration: GL integrity/reports/source dedup/subledger/replay และ HTTP scope revocation; shop/role permissions integration; mcptoken/mcpgateway unit; `go build ./...`.
- Money tests ใช้ exact values, 0.1+0.2, scale/rounding boundary, debit=credit, duplicate/concurrent retry, over-allocation rollback, partial settlement และ rebuild.
- Playwright journal details: 600+400 allocation, partial match/settlement, dirty guard, modal focus trap/return, mobile 390 และ light/dark. **Browser layer ใช้ API fixtures**; persistence/totals พิสูจน์อีกชุดกับ PostgreSQL จริงแยก.
- Production ใช้ normal demo session อ่านรายการ/รายละเอียดและ preview รวมทุกสาขาสำเร็จ ไม่มี accounting mutation. `companywide=true` ได้ success payload และตารางจริง.
- CODE-MAP 54 ไฟล์ตรง source และ git diff whitespace check ผ่าน.

**ยังไม่อ้างว่าครบ:**

- `TestFreshInstall_FullUATCycle` ถูก skip ใน combined suite เพราะต้องเตรียม central Holding/tenant fixture เพิ่มจากฐาน GL แยก. ต้องทำ fixture ให้ครบก่อนอ้างว่าติดตั้งใหม่ทั้งระบบผ่าน.
- ยังไม่มีหลักฐาน browser → BFF → backend → PostgreSQL แบบเขียนจริงครบทุกวงจรในบริษัททดสอบเดียวกัน. ห้ามนำ fixture browser test มาอ้างเป็น full-stack UAT.
- ไม่ได้จำลองปิดปีจริงบน production และไม่ได้ตรวจผลกระทบย้อนหลังของบั๊กเดิมกับทุก Holding/บริษัท.
- Backup ผ่าน checksum แต่ยังไม่มีหลักฐาน restore rehearsal ในรอบนี้; การ restore ต้องทำในฐานแยก ไม่ทับฐานจริง.
- ภาพใน `.next/` และ `test-results/` เป็นไฟล์ชั่วคราว อาจถูก build/test รอบใหม่ล้าง; ใช้ผลรันทดสอบปัจจุบันเป็นหลัก. ภาพ preview รุ่นสุดท้าย ณ handoff คือ `frontend/.next/gl-production-companywide-preview.png`.

## 7. งานต่อที่แนะนำตามลำดับ

รายการนี้เป็นงานตรวจ/เสริมหลักฐาน ไม่ใช่ยืนยันว่ามีบั๊กใหม่แล้ว. เมื่อพบปัญหาให้ reproduce พร้อม test ก่อนแก้ขั้นต่ำ.

1. **ทำ fresh-install fixture ให้ครบ**: เตรียม central users/Holding/company/member/role และ Holding DB แยกให้ `TestFreshInstall_FullUATCycle` ใช้งานได้; ตรวจต้นเหตุจาก source ไม่ลบ assertion/guard เพื่อให้เขียว.
2. **Full-stack UAT บนบริษัททดสอบ**: สร้าง AR/AP invoice + payment/credit, หลายใบ/หลายบรรทัด, partial settlement, นำเข้า Statement ซ้ำ, match/withdraw/reallocate/reverse/rebuild. ตรวจทั้ง API totals และ DB persisted decimals แล้วทดสอบ retry/double-click และสอง session.
3. **ทบทวน as-of/ข้ามงวด**: accounting date กับเวลาสร้าง/ถอนหลักฐาน, Bangkok midnight, backdated post/reversal และ branch scopes. ใช้ existing tests เป็นฐาน เพิ่มเฉพาะช่องที่ยังไม่ครอบคลุม.
4. **ตรวจสิทธิ์หลัง revoke จากจอจริง**: readonly หลายบริษัท/readwrite บริษัทเดียว, token ผิดชนิด/หมดอายุ/revoke, issuer ถูกลดบทบาท, legacy branch token และ current membership grant; อย่าเปิดสิทธิ์จาก frontend อย่างเดียว.
5. **ทวนข้อความและเอกสารเฉพาะจุด**: new labels บางส่วนยังใช้ Thai/English fallback ไม่ได้ยืนยันแปลครบทุกภาษา. เอกสารเก่าบางไฟล์มีชื่อตาราง/บริการเก่า เช่น gl-journal-review.md กล่าวถึง gl_journal_lines ขณะที่ runtime schema ใช้ gl_lines; แก้เฉพาะเรื่องที่ verify source แล้วและอยู่ในขอบเขต.

6. **ทดลอง restore backup ในสภาพแวดล้อมแยก**: ตรวจว่ากู้ schema/audit/source registry/subledger ได้และยอดตรงก่อนอ้างว่าพิสูจน์การกู้คืนแล้ว. เลือกใช้สำเนาที่มีสิทธิ์เข้าถึงและไม่เผยข้อมูลจริงใน log.
7. **ตรวจข้อสงสัย CSV dedup**: source key ใช้ hash ไฟล์และเลขแถว. ทดลองข้อมูลเดิมที่สลับแถว/เปลี่ยน newline/ส่งออกใหม่ว่าจะเกิด Statement ซ้ำหรือไม่; ตอน handoff ยังไม่ reproduce จึงยังไม่ใช่บั๊กยืนยัน. อย่าทำ unique ตามวัน+ยอดอย่างเดียวเพราะธนาคารอาจมีรายการจริงที่เท่ากันหลายครั้ง.

ข้อจำกัดที่ตั้งใจไว้ ไม่แก้โดยเดากติกาธุรกิจ:

- THB เท่านั้น, scale ตามปีบัญชี; body 2 MB และรายละเอียดรวม 2,000 แถวต่อคำสั่ง; support page ไม่เกิน 1,000 แถว.
- บริษัทเดิมที่ไม่มี period rows ยังลงรายการได้เพื่อ compatibility; เมื่อมี locked period หรือ closed year จะบังคับ guard. ถ้าจะเปลี่ยนเป็นบังคับสร้างทุกงวด ต้องตกลงนโยบายกับลุงจืดก่อน.
- ปิดบัญชี/ขึ้นปีใหม่สร้าง draft ให้ตรวจและผ่านรายการ; ไม่ auto-post ข้ามการตรวจ.
- ไม่มี FX/net AR-AP/ส่งเงิน/เชื่อมธนาคารจริง/RD Prep/XBRL ในงานนี้.

## 8. คำสั่งตรวจซ้ำ (PowerShell)

ตัวอย่างด้านล่างต้องรันกับฐานทดสอบใหม่เท่านั้น. Test containers รอบก่อนถูกลบแล้ว; อย่าตั้ง DSN แล้วคิดว่าพอร์ต 55440 ยังมีบริการ. ตรวจชื่อ/พอร์ตว่างก่อนสร้าง container และอย่าใช้ container ของคนอื่น.

```powershell
Set-Location D:\bccode
docker run -d --name bc-gl-handoff-pg18 -e POSTGRES_HOST_AUTH_METHOD=trust -e POSTGRES_DB=gl_fix -p 127.0.0.1:55440:5432 postgres:18-alpine
docker exec bc-gl-handoff-pg18 pg_isready -U postgres -d gl_fix
```

รอ pg_isready สำเร็จก่อนชุดถัดไป. `trust` ใช้เฉพาะ disposable container ที่ bind loopback นี้ ไม่ใช้ production.

```powershell
Set-Location D:\bccode\backend
$env:BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:55440/gl_fix?sslmode=disable'
$env:GL_AUTH_TEST_DSN=$env:BC_GL_TEST_POSTGRES_DSN
go test -tags integration ./internal/generalledger/... -skip TestFreshInstall_FullUATCycle -count=1
go test -tags integration ./internal/shop ./internal/organization/rolepermission -run 'TestPostgresMembershipFailsClosed|TestPostgresRolePermissionAdministration' -count=1
go test ./internal/mcptoken ./internal/mcpgateway -count=1
go build ./...
```

คำสั่ง skip ข้างบนใช้เทียบ baseline เท่านั้น. งานลำดับ 1 ต้องสร้าง fixture แล้วรัน test นั้นโดยไม่ skip; ห้ามใช้ DSN production. ตรวจ exit code ทุกคำสั่งก่อนทำขั้นถัดไป.

```powershell
Set-Location D:\bccode\frontend
npm run test
npm run typecheck
npm run lint
npm run build
# เปิดอีก terminal ใน D:\bccode\frontend แล้วรัน npm run start หลัง build
# playwright.config.ts ไม่เริ่ม server ให้; รอ local server พร้อมก่อนคำสั่งนี้
$env:E2E_BASE_URL='http://localhost:3000'
npm run test:e2e -- e2e/gl-journal-details.spec.ts e2e/gl-report-drill.spec.ts
Set-Location D:\bccode
pwsh -NoProfile -File tools/gen-code-map.ps1 -Check
git diff --check
```

ถ้า CODE-MAP ไม่ตรง: รัน `pwsh -NoProfile -File tools/gen-code-map.ps1` แล้ว `-Check` อีกครั้ง. ก่อน cleanup ฐานทดสอบให้ตรวจ exact container ID/name/port ของตัวที่สร้างเอง; ไม่ลบกว้าง.

Windows เครื่องนี้เรียก Python ด้วย `py`, ตั้ง `PYTHONIOENCODING=utf-8` เมื่อต้องพิมพ์ไทย. อ่าน .env ได้เฉพาะเท่าที่จำเป็น ห้ามแสดง secret. ถ้าเครื่องมือ sandbox มีปัญหาให้ใช้กลไก permission ของ agent นั้นตามจริง ไม่คัดลอกวิธีข้ามข้อจำกัด.

## 9. Deploy หลังงานถัดไปผ่าน

ใช้กฎ AGENTS.md ล่าสุดและอ่าน `tools/fast-deploy.py` ก่อนรัน. อย่าใช้ tag เดิมซ้ำ เพราะ script อาจเขียนทับ backup ของ release นั้น.

```powershell
Set-Location D:\bccode
$env:PYTHONIOENCODING='utf-8'
# แทน <NEW_UNIQUE_TAG> ด้วยชื่อรุ่นใหม่ที่ไม่ซ้ำ ก่อนรัน
py tools/fast-deploy.py --all --tag <NEW_UNIQUE_TAG>
# ถ้า frontend-only จริง ใช้ --frontend แทน --all
```

ตรวจ backup/manifest ก่อน switch, image ที่รันจริง, health/restarts/logs, authenticated payload, หน้าจอจริง และจำนวน/ยอดก่อน–หลัง. หาก deploy script หลุดกลางทาง ให้ inspect release/images/services ก่อนทำต่อ ห้าม rerun โดยเดาว่ารุ่นยังไม่เปลี่ยน. ไม่แก้ข้อมูลบัญชี production เพื่อทำ smoke test.

## 10. Prompt ส่งให้ GLM

> อ่าน D:\bccode\AGENTS.md และ D:\bccode\docs\handoff\HANDOFF-2026-09-20-GL-GLM.md แล้วรับช่วงตรวจและแก้ระบบ GL จาก working tree ปัจจุบัน. เริ่มจากช่องว่างการตรวจรับในข้อ 7 โดยเฉพาะ fresh-install fixture และ full-stack UAT บนฐานแยก. ทวน findings เดิมกับแพตช์ที่แก้แล้ว ห้ามถือว่าอาการเดิมยังมีโดยไม่ reproduce. เมื่อพบข้อผิดพลาดให้เพิ่ม regression ที่พิสูจน์อาการ แก้ขั้นต่ำ และตรวจผลจริง. รักษา many-to-many, source dedup, exact decimal, audit, token API/MCP แยกชนิด และขอบเขตบริษัท/สาขา. ห้ามแก้ mydocs, ห้ามใช้/ผ่านรายการ 144 draft ของ rungrueng/01 เพื่อทดสอบ, ห้าม pull/reset/push เอง และอย่าเพิ่ม RD Prep/XBRL/หน้าบ้าน. รายงานสิ่งที่แก้ ผลทดสอบ ข้อจำกัด และหลักฐาน deploy ตามกฎโปรเจ็กต์เป็นภาษาไทยให้ลุงจืด.
