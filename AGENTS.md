# BC Ai Account — Codex Router (ฉบับย่อ)

ฉบับเต็ม (เหตุผล ตัวอย่าง ประวัติ) อยู่ที่ `docs/kms/rules/agents-full.md` — เปิดอ่านเฉพาะหัวข้อที่ต้องใช้

## Router (`D:\bccode`)
- ความรู้ระบบ=`docs/kms/` (`00`–`22` ไม่มี `12`;`README.md` ดัชนี;`decisions/` ADR;`bugs/`;`architecture/`;`snippets/`)(โค้ดทำอะไร ไม่ใช่ business rule);requirement ไม่ชัด→ถามลุงจืด ห้ามเดา;code=truth(docs ขัดโค้ด→แก้ docs ใน commit เดียวกัน)
- อ่านเฉพาะไฟล์ที่จำเป็น(ผัง `docs/README.md`)ห้ามกวาดทั้งโฟลเดอร์/เปิด handoff ล่วงหน้า;`docs/kms/00-source-router.md` หาหลักฐานโค้ดเท่านั้น;`mydocs/` อ่านอย่างเดียว

## กฎ/skill อยู่ที่เดียว (2026-09-07,09-20)—ADR `2026-09-20-single-source-ai-rules-skills.md`
- กฎ=`AGENTS.md` (Claude ผ่าน `CLAUDE.md` `@AGENTS.md`;Gemini CLI ผ่าน `.gemini/settings.json` `context.fileName`);skill=`.agents/skills/<name>/SKILL.md` (Claude:junction `.claude/skills` จาก `npm run ai:link`)
- ห้ามไฟล์กฎแยกต่อเครื่องมือ(`GEMINI.md`,`.agents/rules/*`,`.cursorrules`,`.codex/AGENTS.md`,กฎซ้ำใน `CLAUDE.md`)→ทำเป็น pointer/import;ห้ามคัดลอก skill ไป `.claude/skills/` จริง,`.gemini/skills/`,`docs/skills/`,`~/.agents/skills/`
- กฎเครื่อง=`~/.agents/AGENTS.md` (สำเนา `node ~/.agents/sync-rules.mjs`)ห้ามฝังกฎโปรเจ็กต์ใน global
- อ่าน SKILL.md จาก disk ห้าม cache:UI→`ui-scale-polish`;บัญชี/ภาษี→`thai-accounting-tax`+`docs/kms/21-thai-tax-form-references.md` (ห้ามเดา ต้องอ้างอิงทางการ,2026-09-23);แก้ skill→commit พร้อมงาน;บทเรียน→`docs/kms/` หัวข้อละไฟล์ อ้าง `file:line`+เพิ่มใน `docs/kms/README.md`

## mydocs/ Read-Only (2026-09-20)
- `mydocs/`=ข้อกำหนดของลุงจืด(`mydocs/specs/rules.md` ฯลฯ)อ่านใหม่เมื่อเปลี่ยน;AI ทุกตัวห้ามสร้าง/แก้/ลบไฟล์หรือโฟลเดอร์ใน `mydocs/`;ข้อเสนอ→แชท/เอกสารนอก `mydocs/`;GL:md `mydocs/specs/gl/`,SQL `mydocs/datamodels/gl/`

## Deploy อัตโนมัติ (2026-09-12)
- จบงาน+verify ผ่าน(test/tsc/build 100%)→deploy production(https://account.bcaicloud.com/)ไม่ถาม
- Frontend only→ไม่ rebuild `mainapi` (tag จาก image เดิม);`docker save`|`ssh -C root@159.223.43.229 "docker load"` ไม่เขียน `images.tar`
- Preflight backup `release.env.before`+DB;สลับ `release.env` atomic+`docker compose up -d --no-deps frontend` (หรือ mainapi);ตรวจ HTTP(200/401)+smoke test URL จริง

## ยึด `D:\project-champ` เป็นต้นแบบ (2026-09-15,09-19)
- Upgrade Windows→Web ไม่บวม:ฟังก์ชัน/workflow/logic/ฟิลด์/การคำนวณ/เมนู(`champ/champ/menuconfigxml/menuconfig.xml`,`BC5Account.rc`,C++/MFC,DDL,รายงาน)ต้องเทียบเท่า ห้ามตัด เว้นลุงจืดสั่งเป็นลายลักษณ์อักษร(เช่น payroll);ตรวจโค้ดจริงก่อนออกแบบจอ/เมนู/ฟิลด์/API
- Strict Parity:ห้ามคิดฟังก์ชัน/เมนู/จอใหม่เอง—Champ ไม่มี+ไม่มีคำสั่งลุงจืด+ไม่ใช่ข้อบังคับกฎหมายบัญชี/ภาษีไทย=ห้ามเพิ่ม;Champ แยกจอซ้ำ→ยุบเป็นจอเดียวมี Tab/Action;ความสามารถใหม่(ตลาดออนไลน์,AI ผู้ช่วย,LINE OA,e-Tax Invoice,Dashboard ผู้บริหาร)→เสนอข้อดี-ข้อเสียให้ลุงจืดอนุมัติก่อน
- เมนูตาม `menuconfig.xml`;`goods-inspection`,`low-stock-alert`,`sale-reservation-flow` มีใน Champ ห้ามตัด;ห้าม engine คำนวณฝั่ง browser—ADR `2026-09-19-champ-parity-menu-cut.md`

## Backend-first+REST API+MCP Server (2026-09-15)
- Business logic/คำนวณบัญชี-ภาษี/aggregation/ปิดงวด/year-end/ปันส่วนต้นทุน/งบ-รายงาน→Backend(Go+PostgreSQL) 100%;Frontend=presentation/input/validation เบื้องต้น ห้าม business logic/heavy calc
- ทุกธุรกรรม/ฟังก์ชัน/การสืบค้นมี REST API+schema+error contract
- Backend มี MCP Server ให้ AI agent/Vibe Coding:Read(ผังบัญชี,ยอดคงเหลือ,สมุดรายวัน,งวด,รายงาน),Write(journal,post/unpost,งวด,ประมวลผล),Knowledge & Schema;ใช้ได้ทั้ง MCP และ REST

## PostgreSQL อย่างเดียว (2026-09-19,09-20,09-23)
- ห้ามเพิ่ม Kafka,Zookeeper,Worker,MongoDB,Redis,ClickHouse กลับ;Go Pure `CGO_ENABLED=0` ไร้ `librdkafka`/CGO
- ธุรกรรมบัญชี(GL,Journals,Invoices,Receipts,Master Data)=synchronous ACID `BEGIN ... COMMIT` ผ่าน `sql.Tx` ไม่ใช้ eventual consistency/queue;pagination `LIMIT $1 OFFSET $2`+`ORDER BY` เสมอ
- Prod:`postgres` (`postgres:18-alpine` ตาม `deploy/account/compose.yml`),`minio`;session/token ตาราง `cache_entries`;จอที่ API เดิมบน Mongo="รอพัฒนา"—ADR `2026-09-23-remove-mongo-kafka-redis-clickhouse.md`

## ชื่อใน PostgreSQL lowercase (2026-09-19)
- ทุกชื่อ(database `bcai_projection`,schema,table `gl_records`,column `holding_code`,index,FK,PK,sequence)=`[a-z0-9_]` ห้าม uppercase/camelCase และ quoted identifiers(`holding_code` ไม่ใช่ `"holdingCode"`);Go tag `db:`/`json:` และ TS API/DTO ตรง snake_case

## ขอบเขต:ไม่ทำ payroll (2026-09-08)
- ห้ามเพิ่มเมนู/จอ/สเปก/API:ระบบเงินเดือน,ภ.ง.ด.1/ภ.ง.ด.1ก,ไฟล์สมทบประกันสังคม(สปส./กท.20 ก)
- ยังอยู่ในขอบเขต:ภ.ง.ด.2(เมนู `vat-pnd2`→`/report/vatpnd2` ใน `frontend/src/lib/menu-data.ts` ห้ามลบ),ภ.ง.ด.3/53,50 ทวิ,เงินทดรองจ่ายพนักงาน,ทะเบียนพนักงาน `/employee`
- อ้างอิง `docs/kms/19-menu-coverage-market-standard.md`,ADR `2026-09-08-menu-parity-market-standard.md` (ห้ามเขียนว่า "ครบ 100%")

## ขอบเขต GL (2026-09-20,09-23)
- ภาษีทุกเมนูอยู่ในหมวด "บัญชีแยกประเภท";ผู้ใช้ GL ตัวเดียว ใช้ได้เลย ระบบอื่นยังไม่ต้องทำ—ADR `2026-09-23-tax-inside-general-ledger.md`
- ระบบห้องบัญชี:รับข้อมูล ตรวจสอบ กระทบยอด ปิดงบ ไม่ใช่ระบบหน้าบ้าน/workflow ระบบอื่น;สเปกหลัก `mydocs/specs/gl/spec.md`
- รายวันครั้งเดียวเชื่อมลูกหนี้/เจ้าหนี้/Statement ธนาคารแบบ many-to-many(เก็บยอดจัดสรรในตารางเชื่อม);AR/AP/Statement=รายละเอียดตรวจสอบ ไม่ใช่คำสั่งรับจ่ายเงินจริง

## บทบาท DevOps/Product Engineer+นักบัญชีไทย (2026-09-18)
- ห้ามเชื่อแค่ HTTP 200/401/build ผ่าน—ตรวจ payload จริง(ไม่ใช่ empty JSON/จอขาว/stale chunks);หลัง deploy `docker compose ps`+`docker logs --tail 50`;เทียบ local vs prod(timezone `Asia/Bangkok`,env,`X-Forwarded-For`);ผิดปกติ→สืบด้วย evidence
- ประเมิน blast radius;backup ก่อน switch;เก็บ release เก่า rollback ได้ ≥72 ชม.;ทุกฟังก์ชันมี evidence(test,DDL,API response,log)
- Defensive:fallback null/undefined(`?? ""`,`|| []`);ปุ่มบันทึก enabled ทันทีที่แก้จริง+dirty guard;ปัดเศษ/หาร 0/ติดลบ/`trim()`;ดัก `ChunkLoadError`
- ปลอดภัย:OWASP,HttpOnly Secure cookies,parameterized SQL 100%,sanitize input
- บัญชี:TFRS,ประมวลรัษฎากร,DBD;Debit=Credit 100%;VAT(ภ.พ.30/36)แยกฐาน+ภาษีซื้อต้องห้าม;WHT(ภ.ง.ด.2/3/53,50 ทวิ)ตามอัตราจริง;ทุกธุรกรรมมี audit trail ลง GL;audit log เหตุผลยกเลิก/แก้ไข;ห้ามเดากฎหมาย/ภาษี/สูตร—ค้นข้อกฎหมาย/ประกาศสรรพากร/เทียบ Champ/SML
- คีย์งานลื่น(Tab/Enter/Numpad,Smart Paste);error ไทยบอกวิธีแก้;จบงานสรุปสถานะ+เสนอ 2–3 ข้อขั้นต่อไป

## รูปแบบนำเสนอ `[1]`,`[2]`,`[3]` (2026-09-18)
- สรุป/แนะนำขั้นต่อไป ห้ามพรรณนายาว→จัดเป็น `[1]`,`[2]`,`[3]` ลุงจืดตอบเลขเดียว(`1`,`1,2,3`)ทำต่อได้ทันที
- แต่ละข้อ:Action+ปัญหาที่แก้/คุณค่า+Scope(ไฟล์/โมดูล);ตัวแนะนำสุดก่อน ระบุ `(แนะนำ)`;แยกหมวด Bug/Fix,Enhancement UX,Value Add

## ข้อมูลสมจริง SMEs ไทย ห้ามคำว่า demo (2026-09-19)
- ข้อมูลทุกชนิด(seed/test/ผังบัญชี/สมุดรายวัน/สินค้า/ลูกค้า/เจ้าหนี้/องค์กร)เหมือนใช้งานจริง 100%;ห้าม "demo","ข้อมูลตัวอย่าง","(สาธิต)","Mock","Dummy","Test" ในข้อมูลที่ผู้ใช้เห็น
- ผังบัญชี multi-level ตาม TFRS/DBD ห้าม flat;JSON มี `"level"`+`"parentaccountcode"`:L1 `10000`–`50000` (สินทรัพย์/หนี้สิน/ส่วนของเจ้าของ/รายได้/ค่าใช้จ่าย);L2 `11000`/`12000`/`21000`/`22000`;L3 บัญชีคุม `11100`/`11200`/`11300`/`12100`;L4 ลงรายการ `11110`–`12120`;L1–L3 `allowposting:false`,L4 `true`
- ชื่อธุรกิจ/สาขา/รายการค้าไทยจริง

## ห้ามชื่อแบรนด์คู่แข่ง (2026-09-09)
- ห้ามชื่อ/อ้างอิง FlowAccount หรือ PEAK Account(หรือบุคคลภายนอกใด)ในโค้ด(คอมเมนต์,ชื่อไฟล์/ตัวแปร,routes,UI,commit message)และเอกสาร `docs/`;ใช้คำกลาง เช่น "มาตรฐานโปรแกรมบัญชีไทยทั่วไป";`.githooks/pre-commit` สแกน staged→พบ=reject

## ภาษา: key อังกฤษ+th/en เท่านั้น (2026-09-14,09-16,09-19)
- ห้าม hard-code ข้อความไทย/ภาษาใดในโค้ด:key snake_case(`gl_post_journal`)+`backendText(dictionary,key,fallback)` (`frontend/src/lib/backend-language.ts`);ข้อความอยู่ `backend/assets/language/languages.tsv` (13 คอลัมน์ `key th en cn ja km ko lo my vi ms id fil`)
- ทุกจอรับ prop `language`+`useBackendLanguage(language,backendUrl)`;ห้ามอ่าน `localStorage` เมิน prop;component กลาง(`confirm-dialog`,`numeric-input`)รับป้ายเป็น prop;ห้าม `t(th,en)`
- เพิ่มข้อความ=เพิ่ม 1 แถวครบ 13 คอลัมน์ใน commit เดียวกัน(ไม่ครบ→ถูกข้าม คืน key ดิบ—`docs/kms/17-dev-gotchas.md`);th=ต้นฉบับ,en=คำแปลจริง,ภาษาอื่นใส่อังกฤษ placeholder ห้ามว่าง
- Backend คืนข้อความผ่าน key+`language.Text(key,lang)` ตาม `Accept-Language`/`lang` ไม่ใช่ `fmt.Errorf("ข้อความไทย")`
- เปิดแค่ th+en:`ACTIVE_LANGUAGE_CODES = ["th", "en"]` (`frontend/src/lib/i18n.ts`);ห้ามเปิดภาษาที่ 3/เริ่มรอบแปล/เพิ่ม-ลดคอลัมน์ tsv จนลุงจืดสั่ง;คงโครง 12 ภาษา(`src/locales/*.json`)ซ่อนไม่ลบ;ตรวจรับ th+en(ไทยค้างตอนเลือก en=bug)

## UX/UI (2026-08-30,09-02,09-06)
**40+**—ผู้ใช้หลัก=คนไทยอายุ 40+ มาก่อนความสวย
- ตัวอักษร ≥0.9rem(ห้าม <0.8rem กับข้อความสำคัญ);ไทยเป็นหลัก;line-height ≥1.45 ห้ามตัดกลางคำ `overflow-wrap: anywhere`
- ปุ่ม action หลัก ≥2.6em(≈44px+);ห้าม icon เปล่ากับ action สำคัญ(ต้องมีข้อความไทยหรือ `aria-label`+`title`);WCAG AA 4.5:1(ใหญ่/UI 3:1)ห้ามสื่อสถานะด้วยสีอย่างเดียว
- หนึ่งจอ CTA เดียว;ลบ/ทับ→dialog ไทยอธิบายผล+validate+dirty guard;feedback ไทยง่าย ไม่โชว์รหัสเทคนิค;ห้าม auto-scroll ฉับพลัน;transition ≤300ms เคารพ `prefers-reduced-motion`

**พรีเมี่ยม**—เท่าหน้า login/holding
- accent จาก `--primary` ผ่าน `color-mix` ห้าม hard-code สี;light/dark สวยเท่ากัน;card มุม 16–28px ขอบ hairline เงาย้อมสี primary นุ่ม;hover ยก ≤1px;focus ring 3–4px
- Motion:ambient glow ≥15s+stagger ≤0.5s เท่านั้น ห้าม `filter: blur()` เคลื่อนไหว;hero=รูปถ่ายธุรกิจไทยจริง webp ≥1600px ห้าม SVG/clipart;ห้าม `overflow: hidden` บน panel ที่มี popover
- ตรวจรับ screenshot light+dark×1600/1280/1024/768-portrait+hover/focus/disabled/error,ไม่มี console error;dark กดปุ่มสลับธีมจริง
- CSS skin pass=block เดียวท้าย `frontend/src/app/globals.css` (มี comment วันที่+เหตุผล);แก้ด้วย Node byte-preserving(EOL ผสม)ไม่ใช้ Edit tool

**Upgrade skill**—ก่อนแตะ UI อ่าน `ui-scale-polish`;มาตรฐานกลาง(icon ปุ่ม,`!pl-10`,ความสูงปุ่ม,palette,density)→อัปเดต `.agents/skills/ui-scale-polish/SKILL.md` ทันที(แบบแผนใหม่,anti-pattern,root cause,ไฟล์:บรรทัด)commit พร้อมโค้ด;AGENTS.md ขัด skill→AGENTS.md ชนะ;ห้ามลอกโค้ดเก่าที่ยังไม่อัปเกรด

## UAT: CRUD+PostgreSQL ทีละ step (2026-08-30,09-23)
- CRUD ครบ Create→Read→Update→Delete+edge cases(ว่าง,ซ้ำ,อักขระพิเศษ,double-click);ตรวจ PostgreSQL ทันทีหลังแต่ละ operation(`psql` ใน container `postgres` ฐานของ holding นั้น):Create→แถว/ค่าถูก;Update→แถวเดิมเปลี่ยน(ไม่ใช่แถวใหม่/orphan);Delete→หายจริง;ห้ามรอจบชุดแล้วตรวจรวด
- record อื่นต้องรอด;เคลียร์ข้อมูลทดสอบด้วย id/code เป้าหมายเท่านั้น ห้าม regex กว้าง;สุ่มมี seed ตาม `tests/uat-crud.spec.ts`;รายงานผ่าน/ไม่ผ่าน+หลักฐาน

## ทดสอบจอด้วยปุ่ม Demo สิทธิ์ all (2026-09-23)
- ปุ่ม "ทดลองใช้ระบบ(Demo)" เปิดทุก env ช่วง dev:`/api/auth/demo-login`→backend `/demo-login`;env `BCAI_DEMO_LOGIN_ENABLED=true`+`BCAI_DEMO_USERNAME` (ค่าเริ่มต้น `demo`)
- Demo=owner holding `rungrueng` `permission_sets`=`["*"]` `access_scopes`=`{}` ใน `holding_members` ของ `bcai_projection`;จอปฏิเสธ Demo=bug metadata สิทธิ์ ต้องแก้ ห้ามเลี่ยงไปบัญชีอื่น
- AI ทดสอบเองหลังแก้ UI/API(Browser pane/Playwright):Demo→`rungrueng`→บริษัท `01`→สาขา `00000`→ตรวจ payload จริงใน network,console error,screenshot;ห้ามขอลุงจืดทดสอบแทน ห้ามกรอกรหัสผ่านจริง
- ก่อน go-live ปิด(`BCAI_DEMO_LOGIN_ENABLED=false`)หรือย้าย Demo ไป holding แยก;ห้าม Demo เป็นสมาชิก holding ที่มีข้อมูลลูกค้าจริง

## รูปภาพ:S3/MinIO+thumbnail (2026-08-31)
- ห้ามเก็บรูป(binary/base64)ใน PostgreSQL—เก็บแค่ URI/object key;อัปโหลด `POST /api/upload/image`→`/goapi/image/upload`→PutObject bucket จาก env `S3_*`
- ต้องมี thumbnail เสมอ เก็บใน `<field>thumb`;จอ list ใช้ thumb

## ความเร็ว+ประหยัด token (2026-09-07)
- ห้ามอ่านทั้งไฟล์ >300 บรรทัดโดยไม่จำเป็น;ไฟล์ >1,000 บรรทัดดู `docs/reference/CODE-MAP.md` ก่อน—เก่าได้:เทียบ `wc -l` ไม่ตรง→`grep -n` หรือ `pwsh -NoProfile -File tools/gen-code-map.ps1`
- `.githooks/pre-commit` รัน `tools/gen-code-map.ps1 -Check` เมื่อแตะไฟล์ ≥950 บรรทัด/ในแผนที่;ทุก clone `npm run hooks:install` (ห้าม `git config core.hooksPath` เพราะปิด `post-commit`);ข้ามรอบเดียว `SKIP_CODE_MAP_CHECK=1 git commit ...`
- patch เฉพาะจุด ห้าม rewrite ทั้งไฟล์;ห้ามรัน full test suite(`npm test -- <path>.test.ts`);output >50 บรรทัดครอบ `| head -n 30`;สำรวจ >5 ไฟล์→delegate subagent(`research`)รับสรุป 5–10 บรรทัด

## GitHub เก็บโค้ดอย่างเดียว ไม่มี CI (2026-09-09)
- `.github/workflows/ci.yml` ลบถาวร(กู้จาก commit `1799b069`);ห้ามสร้าง workflow ใหม่ใน `.github/` หรือเสนอเปิด CI โดยไม่ถามลุงจืด;ห้ามย้าย `backend/.github/workflows/*.yaml` ขึ้น root
- ตัวตรวจ(ทุก clone `npm run hooks:install`):`.githooks/pre-commit`;`.githooks/pre-push` รัน `tools/verify.sh codemap`+`frontend` เมื่อแตะ `frontend/` (ข้าม `SKIP_VERIFY=1 git push`);`npm run verify` ก่อน push ทุกครั้ง;`npm run verify:all` (+backend)ก่อน deploy
- ห้ามเขียนว่า "CI จะจับให้";ไม่ได้รัน verify=บอกตรง ๆ;แก้คำสั่งทดสอบที่ `tools/verify.sh` ที่เดียว+อัปเดต `docs/kms/11-testing-quality.md` §5 ใน commit เดียวกัน

## ประวัติแก้ไข=git log ไม่ใช่ README (2026-09-23)
- ห้ามเพิ่ม Activity Log ใน `README.md` (แก้เมื่อภาพรวมระบบเปลี่ยนจริง);commit บรรทัดแรกอังกฤษ conventional(`feat(gl): ...`)+body สรุปไทย(แก้อะไร ได้อะไร ผลทดสอบ);เหตุผล→ADR;บั๊ก→`docs/kms/bugs/`

## commit อัตโนมัติ (2026-09-25)
- จบงาน verify ผ่าน→commit ไม่ถาม(ก่อนสรุป/ก่อน deploy);งานคนละเรื่องแยก commit;verify ไม่ผ่าน/ไม่จบ=ห้าม commit เป็นงานเสร็จ รายงานตรง ๆ
- `git add <path>` เฉพาะไฟล์ของงาน ห้าม `git add -A`/`git add .` (AI อื่นแก้ tree เดียวกัน);ดู `git status -s`+`git diff --cached --stat`;grep secret(`sk-`,`Bearer `,`password=`,`api_key=`)ใน diff—เจอ=หยุดถาม
- commit≠push(push เมื่อลุงจืดสั่ง);ห้าม `--no-verify`;hook ไม่ผ่าน→แก้แล้ว commit ใหม่ ห้าม `--amend` commit เดิม
- ใน worktree(`.claude/worktrees/*`)commit บน branch ของ worktree ไม่แตะ `dev`;ท้ายรายงานบอก hash+branch+สิ่งที่ยังไม่ commit

## Opus คุม+agent ลูก (2026-10-04)
- Opus 5.5 คุม;ตารางเลือกรุ่น(haiku/sonnet/opus/fable)ใน Orchestration Rule ของ `~/.agents/AGENTS.md`
- บัญชี/ภาษี/เงิน(Rule 20),security/auth/สิทธิ์,verify ผลลูก→`opus`;สำรวจโค้ด/แก้ UI/unit+test→`sonnet`;`haiku` เฉพาะงานกลไก ห้ามเขียนคอลัมน์ th ของ `languages.tsv`/ข้อความบนจอ/KMS ไทย ห้าม audit ที่ต้องครบ(Champ parity,secret scan,ลบ flag);`fable` เฉพาะปัญหายากที่ Opus ติด
- ผลลูกต้องผ่าน Opus:รัน `npm run verify`/test/query PostgreSQL เอง;ลูกห้าม commit/push/deploy/แตะ `mydocs/`;ห้ามใช้ AI ค่ายอื่นเว้นลุงจืดสั่ง

## Dev: เดินหน้าอย่างเดียว ไม่สนใจข้อมูลเก่า (2026-06-28,09-15,09-23)
- PostgreSQL ทุก env(local/.202/production)ทิ้งได้:ไม่ทำ migration/backfill/dual-read-write/compat shim;แก้ schema/API contract ตรง ๆ;โค้ดสร้างโครงสร้างเอง(`CREATE ... IF NOT EXISTS`)แล้ว rebuild+deploy;ห้าม DDL มือกับ DB ที่รัน
- ของใหม่พร้อม→ลบของเก่าใน change เดียวกัน ไม่มี feature flag/ตาราง-route-wrapper เก่า "เผื่อไว้"
- แก้ DB/ฟิลด์/หน้าจอ/โค้ดได้เลยไม่ต้องถาม แต่ verify ครบก่อนบอกเสร็จ;ไม่ครอบคลุม git push/force-push/R0 อื่น;ต้องล้างข้อมูล→รายงานว่าล้างอะไร เท่าไร(backup ก่อน deploy ของ `tools/fast-deploy.py` คงไว้)
- Go-live=กฎนี้หยุดสำหรับข้อมูลจริง(ลุงจืดประกาศ go-live หรือ tenant prod มีข้อมูลลูกค้าจริง)→ต้อง migration จริง(backup,นับก่อน/หลัง,ทางถอย)หยุดถามทุกครั้ง;ไม่แน่ใจว่ามีข้อมูลจริงหรือยัง→ถามลุงจืด ห้ามเดา
