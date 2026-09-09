# BC Ai Account — Codex Router

For every task under `D:\bccode`:

1. ฐานความรู้ระบบ **สร้างเสร็จและใช้งานได้แล้ว** ที่ `docs/kms/` (บทความ `00`–`19` รวม 20 ไฟล์ + `README.md` เป็นดัชนี + `decisions/` (ADR) + `bugs/` + `architecture/` + `snippets/`) — แต่ docs อธิบายว่า **โค้ดทำอะไร** ไม่ใช่ข้อกำหนดทางธุรกิจ: requirement/business rule ที่ไม่ชัด = ห้ามเดา ต้องถามลุงจืด.
2. **On-Demand Context Rule**: ฐานความรู้อยู่ที่ `docs/kms/` และ skill ส่วนตัวอยู่ที่ `docs/skills/` — **เปิดอ่านเฉพาะไฟล์ที่จำเป็นกับงานนั้นเท่านั้น (ดูผังเลือกอ่านใน `docs/README.md`)** ห้ามกวาดอ่านทั้งโฟลเดอร์ หรือเปิด handoff ล่วงหน้าโดยไม่จำเป็น เพื่อประหยัด Context Window ของ AI.
3. Use `D:\bccode\docs\kms\00-source-router.md` only to locate implementation evidence.
4. Inspect the exact source, tests, schema, configuration, and runtime evidence required by the task.

ไฟล์นี้เป็นทั้งจุดเข้าเส้นทาง (routing) และกฎบังคับของโปรเจ็กต์ — รายละเอียดว่าระบบทำงานอย่างไรอยู่ที่ `docs/kms/` (code = truth: ถ้า docs ขัดกับโค้ด ให้ยึดโค้ดแล้วแก้ docs ใน commit เดียวกัน)

## กฎ: ขอบเขตผลิตภัณฑ์ — ไม่ทำระบบเงินเดือน (ตั้งโดยลุงจืด 2026-09-08)

BC **ไม่ทำระบบเงินเดือน (payroll)** และไม่ทำสิ่งที่เป็นผลจากเงินเดือน คือ **ภ.ง.ด.1 / ภ.ง.ด.1ก** และ **ไฟล์นำส่งเงินสมทบประกันสังคม (สปส. / กท.20 ก)** — ห้าม AI ตัวใดเพิ่มเมนู จอ สเปก หรือ API เหล่านี้กลับเข้ามาเอง แม้จะเห็นว่า FlowAccount/PEAK มี (เอกสารของคู่แข่งใน `docs/features-flowaccount-peak/` เป็นหลักฐานเปรียบเทียบเท่านั้น)

ยังอยู่ในขอบเขตตามปกติ: ภาษีหัก ณ ที่จ่ายของคู่ค้า — **ภ.ง.ด.2** (เงินได้ 40(3)/(4) ดอกเบี้ย/เงินปันผล/ค่าสิทธิ ต้นทางคือรายการจ่ายเงิน ไม่ใช่เงินเดือน; เมนู `vat-pnd2` → `/report/vatpnd2` ใน `frontend/src/lib/menu-data.ts` **ห้ามลบทิ้งเพราะเข้าใจผิดว่าเป็นเรื่องเงินเดือน**) + **ภ.ง.ด.3/53** + หนังสือรับรอง 50 ทวิ, เงินทดรองจ่ายพนักงาน (งานการเงิน), ทะเบียนพนักงาน (`/employee`) ซึ่งอยู่ **ทั้ง** ในหน้าตั้งค่า (`frontend/src/lib/system-setting-screens.ts` slug `employee`) **และ** ในเมนูหลัก ข้อมูลหลัก › บุคลากรและผู้ใช้งาน ตั้งแต่ 2026-09-08 (ส่วน `/line-oa` เป็นรายการเมนูหลักตั้งแต่ 2026-09-08 เช่นกัน แต่อยู่กลุ่ม ข้อมูลหลัก › ผู้ช่วย AI และคลังความรู้ และไม่ใช่จอในหน้าตั้งค่า)

เหตุผลและรายละเอียดการเทียบเมนูกับ FlowAccount/PEAK: `docs/kms/19-menu-coverage-flowaccount-peak.md` + ADR `docs/kms/decisions/2026-09-08-menu-parity-flowaccount-peak.md` (รอบเว็บทางการ: ตัดระบบเงินเดือนออก + เพิ่ม 15 เมนู เป็น 200) และ `docs/kms/decisions/2026-09-08-menu-parity-social-sweep.md` (รอบแหล่งนอกทางการ: เพิ่มอีก 6 เมนู 218 → 224 — ที่มาของ ภ.ง.ด.2 และกลุ่มเชื่อมข้อมูลตลาดออนไลน์ พร้อมข้อห้ามเขียนว่า "ครบ 100%")

## กฎ: skill ส่วนตัวอยู่ที่ `docs/skills/` และฐานความรู้อยู่ที่ `docs/kms/` (ตั้งโดยลุงจืด 2026-09-07)

1. **skill ส่วนตัวของลุงจืดทุกตัวเก็บใน `docs/skills/<name>/SKILL.md`** (ย้ายจาก `.agents/skills/` แล้ว 2026-09-07) — ห้ามสร้าง/คัดลอกไปที่ `.agents/skills/`, `.claude/skills/` หรือที่อื่น เพื่อให้ตรวจง่ายที่เดียว
2. **ต้องใช้ skill จากที่นี่จริง ๆ** — ก่อนทำงานที่ skill ครอบคลุม (เช่น UI → `docs/skills/ui-scale-polish/SKILL.md`, MongoModel → `docs/skills/audit-mongomodel-sync/SKILL.md`) ให้เปิดอ่านไฟล์ล่าสุดจาก disk ทุกครั้ง **ห้ามใช้เวอร์ชันที่จำได้/cache** เพราะลุงจืดอาจแก้ด้วยมือ; ถ้า AI ตัวใดโหลด skill ผ่านกลไกอัตโนมัติจากที่อื่นได้ ก็ยังต้องยึดไฟล์ใน `docs/skills/` เป็นตัวจริง
3. **บทเรียน/กับดัก/ความรู้ที่ต้องไม่ลืม → เขียนลง `docs/kms/`** (ไม่ใช่แค่ memory ส่วนตัวของ AI ตัวใดตัวหนึ่ง) เป็นไฟล์ Markdown หัวข้อละไฟล์ อ้าง `file:line` ของโค้ดจริง และเพิ่มบรรทัดใน `docs/kms/README.md`; docs ต้องตามโค้ด (code = truth) — ถ้าโค้ดเปลี่ยนให้แก้ docs ใน commit เดียวกัน
4. commit ที่แก้ skill/kms ให้รวมไปกับ commit งานที่ทำให้เกิดการเปลี่ยนแปลงนั้น (เหมือนกฎ Mandatory Skill Upgrade ด้านล่าง)

## กฎ: ผู้ช่วยคิด = Kimi K3 + GLM (ตั้งโดยลุงจืด 2026-09-02; DeepSeek ถอดออก 2026-09-03)

- ใช้ตาม Orchestration Rule ใน ~/.claude/CLAUDE.md: py ~/.claude/tools/kimi-ask.py หรือ glm-ask.py (ต้องใช้ py launcher + PYTHONIOENCODING=utf-8)
- คำตอบผู้ช่วยเป็นความเห็นเท่านั้น Claude ต้อง verify กับ source/build/test ก่อนใช้; ห้ามส่ง secret/PII; R0/R1 ตัดสินโดย Claude + ลุงจืด
- DeepSeek/ChatGPT/OpenRouter ยังไม่เปิดใช้ (ถามก่อน)

## กฎ: UX/UI ยึด "คนไทย อายุ 40+" เป็นบุคลิกหลัก (ตั้งโดยลุงจืด 2026-08-30)

ผู้ใช้หลักของระบบคือคนไทยอายุเกิน 40 ปี (พนักงานบัญชี/เจ้าของกิจการ) — ทุกงาน UX/UI (ทั้งปรับของเดิมและสร้างใหม่) ต้องออกแบบให้กลุ่มนี้อ่านออก ใช้ได้ ไม่กลัวกดผิด ก่อนความสวย/ทันสมัยเสมอ:

1. **ตัวหนังสือใหญ่อ่านง่าย** — ข้อความที่ผู้ใช้ต้องอ่านเพื่อตัดสินใจ ≥ 0.9rem; ห้ามต่ำกว่า 0.8rem กับข้อความสำคัญ (ระบบ root ladder 15–21px อยู่แล้ว) ตัวเล็กสุดใช้ได้เฉพาะ metadata ที่อ่านเมื่อต้องการ
2. **ภาษาไทยเป็นหลัก (Thai-First)** — ป้าย/ปุ่ม/คำอธิบายไทยก่อน; ศัพท์บัญชีตามธรรมเนียมไทย; ศัพท์อังกฤษที่จำเป็นต้องมีคำอธิบายไทย/title ประกอบ
3. **ภาษาไทยซับซ้อน** — line-height ≥ 1.45 สำหรับข้อความไทย, ห้ามตัดกลางคำ, token ยาวต้อง `overflow-wrap: anywhere`
4. **ปุ่ม/จุดคลิกใหญ่พอ** — action หลัก ≥ 2.6em (≈44px+); **ห้ามใช้ icon เปล่า ๆ กับ action สำคัญ** ต้องมีข้อความไทยกำกับ หรืออย่างน้อย `aria-label` + `title` ไทย (คน 40+ ไม่จำความหมาย icon)
5. **คอนทราสต์สูง** — ข้อความ 4.5:1 / ข้อความใหญ่-UI 3:1 (WCAG AA); **ห้ามสื่อสถานะด้วยสีอย่างเดียว** ต้องมีข้อความหรือไอคอนประกอบ
6. **หนึ่งจอ ทำงานเรื่องเดียว** — CTA หลักเดียวต่อจอ มองเห็นชัด; ลำดับอ่าน บน→ล่าง ซ้าย→ขวา ตามลำดับทำงานจริง
7. **กันพลาด + ยืนยันก่อนทำลาย** — ลบ/ทับข้อมูลต้องมี dialog ไทยอธิบายผลกระทบ; validate ก่อนส่งเสมอ; ห้ามทิ้งข้อมูลที่ผู้ใช้กรอกโดยไม่เตือน (dirty guard)
8. **Feedback ทุก action** — สำเร็จ/ล้มเหลวต้องแจ้งด้วยภาษาไทยง่าย ๆ ทันที; error ห้ามโชว์รหัสเทคนิคดิบ ๆ ต้องแปลเป็นภาษาที่ผู้ใช้ทำอะไรต่อได้
9. **อย่าพาสายตากระโดด** — ห้าม auto-scroll ฉับพลัน, hover/focus ring ชัดเจน, ตำแหน่ง dialog ใกล้จุดกด
10. **Motion น้อยและช้าพอเห็น** — transition ≤ 300ms, ไม่ใช้ parallax/เอฟเฟกต์เร็ว, เคารพ `prefers-reduced-motion`

รายละเอียด operational + บทเรียจากงานจริงอยู่ใน skill `ui-scale-polish` (§0)

## กฎ: UX/UI ต้อง "พรีเมี่ยม + ใช้ง่าย + เหมาะกับคนไทย" ทั้งระบบ (ตั้งโดยลุงจืด 2026-09-02)

ต่อยอดจากกฎ "คนไทย 40+" ด้านบน (ใช้ง่าย = กฎนั้น) — กฎนี้กำหนดมาตรฐาน **ความพรีเมี่ยม** ให้ทุกจอเท่ากับหน้า login/holding ที่ approve แล้ว ทุกจอใหม่หรือจอที่แก้ ต้องผ่านทั้ง 2 กฎ ห้ามเลือกอย่างใดอย่างหนึ่ง:

1. **หนึ่งจอ หนึ่งตระกูลสี** — accent ทุกจุด (CTA, focus ring, icon หัวข้อ, eyebrow, glow, hover) derive จาก `--primary` ผ่าน `color-mix`; **ห้าม hard-code สี** (teal/indigo/#fff/#dadce0 ฯลฯ) เพราะระบบมี 10 พาเลต + dark mode ที่ผู้ใช้เลือกได้ — สีที่ไม่ตามพาเลตคือ bug
2. **Light + Dark ต้องสวยเท่ากัน** — ทุก surface/control ใช้ตัวแปร `--login-*`/`--panel`/`--text` ให้ปรับตาม theme เอง; ก่อนบอกเสร็จต้องเปิดดูทั้ง 2 โหมด **โดยกดปุ่มสลับธีมจริง** (ตัวแปรพาเลตถูกเขียน inline บน `<html>` — แค่แก้ `data-theme` ไม่ใช่การทดสอบที่ถูกต้อง)
3. **พื้นผิวมีชั้น ไม่แบน** — card/panel: มุมโค้ง 16–28px, ขอบ hairline โปร่ง, เงาย้อมสี primary แบบนุ่ม (offset ใหญ่ blur ใหญ่ alpha ต่ำ), inner highlight 1px ด้านบน; พื้นหลังมี glow/vignette เบา ๆ ให้จอมีความลึก แต่ข้อความทุกคำต้องนั่งบนพื้นทึบพอ (กฎ 40+ ข้อ 5)
4. **Control เป็นภาษาเดียวกันทั้งจอ** — input/ปุ่ม/ปุ่ม social ใช้ radius, ความสูง, hairline ชุดเดียว; hover ยกตัว ≤ 1px + เงาเพิ่มนิดเดียว; focus ring 3–4px สี primary 14–20%; ปุ่มรอง (เช่น Dev Login) ต้องดูรองจริง (dashed/muted) ไม่แย่งสายตา CTA หลัก
5. **Motion = บรรยากาศ ไม่ใช่ลูกเล่น** — อนุญาต ambient glow ช้ามาก (≥ 15s) + stagger เข้าจอ ≤ 0.5s เท่านั้น; ห้ามใช้ `filter: blur()` เคลื่อนไหว (แพงบน iPad) ใช้ radial-gradient แทน; ทุก animation ต้องนิ่งใต้ `prefers-reduced-motion`
6. **ภาพ/ภาพประกอบต้องมีคุณภาพ** — hero ใช้รูปถ่ายบริบทธุรกิจไทยจริง (ร้าน/คลัง/บัญชี) เป็น webp กว้าง ≥ 1600px มี vignette ให้ข้อความอ่านออก; ห้าม SVG/clipart/stock ต่างชาติชัดเจน (ดูกฎรูปภาพใน `~/.claude/CLAUDE.md`)
7. **Popover/Dialog ห้ามโดนตัด** — อย่าใส่ `overflow: hidden` บน panel ที่มี popover ลูก (font/palette picker, dropdown); ถ้าต้อง clip effect ให้ clip ที่ shell ชั้นนอกสุด และเปิด popover ทุกตัวทดสอบหลังแก้ CSS ทุกครั้ง
8. **ตรวจรับพรีเมี่ยมด้วย screenshot จริง** — ก่อนบอกเสร็จ: light+dark × 1600 / 1280 / 1024 / 768-portrait (iPad ขึ้นไปตาม [[viewport-target-ipad-up]]) + hover/focus/disabled/error state + ไม่มี console error; "น่าจะสวย" ไม่นับ
9. **CSS แบบไม่ทำลายของเดิม** — skin pass ใหม่ = block เดียวต่อท้าย `globals.css` มี comment วันที่+เหตุผล, selector prefix `.login-shell`/`.workspace-page` ฯลฯ ให้ชนะ cascade, ไม่แตะ layout/type scale ที่ approve แล้ว, ค่าใช้ตัวแปรล้วน; แก้ไฟล์นี้ด้วย Node byte-preserving (EOL ผสม) ไม่ใช้ Edit tool
10. **อัปเดต skill ทุกครั้ง (ตั้งโดยลุงจืด 2026-09-02)** — จบงาน UX/UI ใด ๆ (ใหม่/แก้/บทเรียน/กับดัก) ต้อง**สะท้อนกลับเข้า `docs/skills/ui-scale-polish/SKILL.md`** เป็นหัวข้อใหม่ (แบบแผน + เหตุผล + วิธีตรวจ + ไฟล์/บรรทัด) และ **commit skill พร้อมงาน** — เพื่อให้ AI ตัวอื่น/เครื่องอื่นทำต่อแล้วได้ผลลัพธ์เหมือนกัน; ก่อนแตะ UI ต้องโหลด skill นี้ก่อนเสมอ ถ้ากฎใน AGENTS.md กับ skill ขัดกัน ให้ AGENTS.md ชนะแล้วแก้ skill ให้ตรง

ตัวอย่างที่ผ่านมาตรฐาน: หน้า login + holding หลัง pass 2026-09-02 (block "Login premium pass 3" ท้าย `frontend/src/app/globals.css`)


## กฎ: แก้ไข UX/UI ระดับทั้งระบบ ต้อง Upgrade Skill เสมอ — ห้ามวนกลับไปใช้แบบเดิม (ตั้งโดยลุงจืด 2026-09-06)

ทุกการแก้ไขหรือสร้าง UX/UI ที่เป็นมาตรฐานกลางหรือใช้ทั้งระบบ (เช่น Icon ประจำปุ่ม, สไตล์ทางลัด, การหลบ Icon ใน Input ด้วย `!pl-10`, ความสูงปุ่ม, สี Palette, Density, การจัดวาง ฯลฯ):

1. **ต้อง Upgrade Skill ทันที (Mandatory Skill Upgrade)**:
   - ต้องสะท้อนการเปลี่ยนแปลงเข้าสู่ `docs/skills/ui-scale-polish/SKILL.md` ทันทีเสมอ
   - ต้องระบุชัดเจนทั้ง 4 ส่วน:
     1) **แบบแผนใหม่ (New Standard Pattern)**: โค้ดตัวอย่าง คลาส CSS และคุณสมบัติที่ถูกต้อง
     2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**: รูปแบบเดิมที่ผิดพลาดหรือทำให้เกิดปัญหา
     3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**: ทำไมต้องทำแบบนี้ (เช่น CSS `!important` ทับซ้อน, ความคมชัด, จอสัมผัส)
     4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**: ตัวอย่างที่ใช้งานจริงในระบบ
2. **ห้ามวนกลับไปใช้แบบเดิมเด็ดขาด (Zero Regression Guarantee)**:
   - ก่อนแตะไฟล์ UI ใดๆ AI ทุกตัว (Gemini, Claude, Codex) **ต้องอ่าน `ui-scale-polish` ก่อนเสมอ**
   - ห้ามลอกเลียนแบบโค้ดเก่าในจุดที่ยังไม่ได้อัปเกรด — หากพบหน้าจอเดิมที่ยังใช้แพทเทิร์นเก่า ต้องปรับปรุงให้เป็นมาตรฐานล่าสุดตาม Skill ทันที
3. **Commit Skill พร้อมโค้ดเสมอ**:
   - ทุก Git Commit ที่มีการเปลี่ยนแปลง UX/UI มาตรฐาน ต้องรวมไฟล์ `SKILL.md` ไว้ใน Commit เดียวกันเสมอ เพื่อให้เครื่องอื่นและ AI ตัวถัดไปได้รับมาตรฐานใหม่อย่างต่อเนื่อง


## กฎ: UAT ต้องตรวจ CRUD + MongoDB เสมอ (ตั้งโดยลุงจืด 2026-08-30)

การทดสอบ UAT ทุกครั้ง (ทุก entity/หน้าจอที่มีการเขียนข้อมูล) ต้อง:

1. **CRUD ครบ** — ไม่หยุดที่อ่านข้อมูล/happy path: ต้อง Create → Read → Update → Delete ครบวงจรของ entity นั้น รวม edge cases (ค่าว่าง, ซ้ำ, อักขระพิเศษ, double-click)
2. **ยืนยันใน MongoDB จริง — ทีละ step** — หลังแต่ละ operation ต้อง query ตรวจใน `appdb` (mongosh ใน container `mongodb`) **ทันทีเป็นขั้นตอน**: Create → ตรวจว่า doc ถูกสร้างถูก field/ค่าทันที → Update → ตรวจว่าค่าเปลี่ยนจริงใน doc เดิม (ไม่ใช่ doc ใหม่/orphan) → Delete → ตรวจว่าหายจริง — **ห้ามรอจบชุดทดสอบแล้วตรวจรวดเดียวทีหลัง** (ตรวจทีเดียวจะไม่รู้ว่า step ไหนเขียนผิด; เคยเจอ API ตอบ success แต่ DB มีขยะ/orphan)
3. **ข้อมูลต้นทางต้องรอด** — ตรวจว่าการทดสอบไม่กระทบ record ตัวอื่น/กลุ่มอื่น (เช่น ลบ membership ที่ทดสอบแล้ว membership เดิมในกลุ่มอื่นต้องอยู่ครบ)
4. **เคลียร์ข้อมูลทดสอบหลังจบ** — ลบด้วย id/code ที่ระบุเป้าหมายเท่านั้น **ห้ามใช้ regex กว้างกับชื่อ/ข้อความ** (เคยโดนลบสาขาจริงไป 2 ตัว)
5. **สุ่มข้อมูลแบบมี seed** — ใช้ pattern `tests/uat-crud.spec.ts` (seeded random + บันทึก seed ลง metrics) เพื่อ reproduce ได้
6. **รายงานผลตรงไปตรงมา** — ผ่าน/ไม่ผ่าน + หลักฐาน (ภาพ + query result) ไม่ใช่แค่ "ทดสอบแล้ว"


## กฎ: รูปภาพห้ามเก็บใน MongoDB — เก็บใน S3 (MinIO) + ต้องมี thumbnail เสมอ (ตั้งโดยลุงจืด 2026-08-31)

ทุกฟีเจอร์ที่มีรูปภาพ (พนักงาน, ผู้ใช้, สินค้า ฯลฯ) ต้องทำตามนี้เสมอ:

1. **ห้ามเก็บไฟล์รูป (binary/base64 ขนาดใหญ่) ใน MongoDB** — Mongo เก็บได้แค่ URI/object key ของรูป (เช่น `/goapi/s3/file/<key>`)
2. **ไฟล์รูปเก็บใน S3/MinIO เท่านั้น** — ใช้ช่องทางอัปโหลดที่มีอยู่ (`POST /api/upload/image` → `/goapi/image/upload` → PutObject ลง bucket จาก env S3_*)
3. **ต้องมี thumbnail เสมอ** — ทุกรูปต้องมีรูปย่อ (editor สร้างอัตโนมัติและอัปโหลดเป็นอีก object) — เก็บ uri ของ thumb ใน field `<field>thumb` คู่กับ field หลักเสมอ (เช่น `avatar`/`avatarthumb`, `profilepicture`/`profilepicturethumb`) และจอ list ต้องใช้ thumb เป็นตัวแสดงหลัก
4. **ตอน UAT ตรวจตามกฎ UAT + Mongo** — ต้องยืนยันว่า Mongo เก็บแค่ URI, ไฟล์จริงอยู่ใน bucket (ตรวจผ่าน minio client/mc หรือ GET ผ่าน endpoint), และมี object thumbnail คู่กัน


## กฎ: ความเร็วสูงสุด + ประหยัด Context และ Token (ตั้งโดยลุงจืด 2026-09-07)

ทุก AI agent (Gemini, Claude, Codex) ต้องปฏิบัติตามกฎนี้อย่างเคร่งครัดเพื่อรักษาความเร็วและไม่เปลือง token:

1. **Surgical Read (อ่านตรงจุด)**:
   - ห้ามเปิดอ่านทั้งไฟล์ขนาดใหญ่ (>300 บรรทัด) โดยไม่จำเป็น; ให้ระบุเลขบรรทัด StartLine/EndLine เสมอ
   - สำหรับไฟล์ขนาดยักษ์ (>1,000 บรรทัด เช่น `frontend/src/app/system-settings/system-settings-screen.tsx`) ให้ดูตำแหน่งฟังก์ชันจาก `docs/reference/CODE-MAP.md` ก่อนเปิดอ่าน — **CODE-MAP เป็นไฟล์ auto-generated และเก่าได้**: ก่อนใช้เลขบรรทัด ให้เทียบจำนวนบรรทัดในหัวข้อกับ `wc -l` จริงก่อน ถ้าไม่ตรงแปลว่าเลขคลาดเคลื่อน ให้ `grep -n` ชื่อฟังก์ชันยืนยัน หรือ regenerate ด้วย `pwsh -NoProfile -File tools/gen-code-map.ps1`
   - **มีตัวกันแล้ว แต่ต้องติดตั้งเอง (2026-09-09)**: git hook `.githooks/pre-commit` รัน `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ทุกครั้งที่ commit แตะไฟล์ ≥ 950 บรรทัด หรือไฟล์ที่อยู่ในแผนที่อยู่แล้ว (ครอบคลุมการลบ/ย้าย/ทำให้เล็กลงด้วย) — **ทุก clone ต้องสั่ง `npm run hooks:install` ครั้งหนึ่ง ไม่งั้น hook ไม่ทำงาน**
     (ติดตั้งแบบ copy เข้า `.git/hooks/` โดยตั้งใจ — **ห้ามใช้ `git config core.hooksPath`** เพราะมันปิด hook เดิมใน `.git/hooks/` ทิ้งทั้งหมด รวมถึง `post-commit` ที่ refresh Obsidian vault ของลุงจืด; แก้ `.githooks/` แล้วต้องรัน `npm run hooks:install` ซ้ำ)
     ข้ามรอบเดียวใช้ `SKIP_CODE_MAP_CHECK=1 git commit ...`
   - **ไม่มี CI แล้ว**: `.github/workflows/ci.yml` ถูกลบ 2026-09-09 ตามมติลุงจืด "GitHub เก็บ code อย่างเดียว" — ทุกอย่างที่ CI เคยตรวจย้ายมาที่ `tools/verify.sh` ซึ่ง**รันเมื่อคนสั่งเท่านั้น** (`npm run verify` ก่อน push ทุกครั้ง, `npm run verify:all` ก่อน deploy); กู้ workflow เดิมได้ด้วย `git show 1799b069:.github/workflows/ci.yml`
2. **Surgical Patch (แก้เฉพาะจุด)**:
   - ใช้ targeted replace/patch แก้เฉพาะ block ที่จำเป็น ห้าม rewrite หรือ print ทั้งไฟล์ซ้ำ
   - Token ขาออก (Output Token) แพงและช้ากว่าขาเข้า 3–5 เท่า — ยิ่งแก้ตรงจุด AI ยิ่งทำงานเร็ว
3. **Command Output Hygiene (คุม Output ใน Terminal ไม่ให้บวม)**:
   - **ห้ามรัน full test suite ทั้งระบบโดยไม่จำเป็น**: รันเฉพาะไฟล์เทสต์ที่เกี่ยวข้องตรง ๆ เช่น `npm test -- <path>.test.ts`
   - **Git สรุปสั้น**: ใช้ `git status -s` และ `git diff --stat` เป็นหลัก
   - **จำกัดบรรทัด**: คำสั่งที่อาจพ่น log เกิน 50 บรรทัด ต้องครอบด้วย filter หรือ head เสมอ (เช่น `| Select-Object -First 30` หรือ `| head -n 30`)
4. **Context Isolation ด้วย Subagent**:
   - งานสำรวจ/วิจัยที่ต้องเปิดอ่านไฟล์จำนวนมาก (>5 ไฟล์) หรือค้นหากว้างขวาง ให้ delegate ให้ Subagent (เช่น `research`) ทำใน sandbox แยก แล้วส่งกลับมาเฉพาะสรุปสั้น 5-10 บรรทัด เพื่อไม่ให้ context หลักบวม
5. **Prompt Caching Discipline**:
   - ไม่แก้ไขสลับไปมาใน system instructions / rules บ่อย เพื่อให้ backend ของโมเดลติด Prompt Cache สูงสุด (ประหยัด token 90% และตอบเร็วกว่าปกติ 2-4 เท่า)


## กฎ: สถาปัตยกรรม 2-Tier — MongoDB เก็บย่อ (Storage) + PostgreSQL ประมวลผลเร็วแบบครบจบ (Processing Engine) (ตั้งโดยลุงจืด 2026-09-07)

ระบบกำหนดบทบาทและข้อตกลงการจัดการข้อมูลระหว่าง MongoDB และ PostgreSQL ไว้อย่างเคร่งครัด ดังนี้:

1. **Clone จาก MongoDB ไปสร้างใน PostgreSQL เสมอ (Single Direction of Clone / Projection)**:
   - ข้อมูลทุกอย่างที่ถูกบันทึกลงใน MongoDB ต้องมีกลไก (Outbox / Kafka Consumer / Worker Sync) ไปสร้างสำเนา (Clone / Read Model) ใน PostgreSQL (per-holding database `<holdingcode>`) เสมอ
   - ทุก entity / collection ที่มีการเขียนในระบบ ต้องมีคู่ตารางใน PostgreSQL รองรับ
2. **MongoDB = Storage Layer (เน้นเก็บข้อมูล ประหยัดขนาด ไม่บวม)**:
   - MongoDB ทำหน้าที่เป็น Store หลักในการรับเข้าและบันทึกข้อมูล (Intake & Persistence)
   - **ต้องประหยัดขนาดข้อมูล (Compact & Slim Storage)**: เก็บเฉพาะฟิลด์ที่จำเป็น ไม่เก็บข้อมูลบวมซ้ำซ้อน (No Redundant Data) และห้ามเก็บไฟล์ binary หรือรูปภาพใน MongoDB เด็ดขาด (เก็บแค่ URI / Object Key ตามกฎรูปภาพ)
3. **PostgreSQL = Processing & Computation Engine (การประมวลผลทั้งหมดเพื่อความเร็วสูงสุด)**:
   - การประมวลผลทางธุรกิจ การคำนวณซับซ้อน งานรายงาน และการสืบค้นทั้งหมด ต้องเกิดขึ้นและประมวลผลใน PostgreSQL เท่านั้น (เช่น งานตัดสต็อก, คำนวณต้นทุน FIFO/Average, สรุปยอดขาย, ภาษี, บัญชีแยกประเภท, งบการเงิน, Search & Filter ขั้นสูง)
   - ใช้ความสามารถเชิงสัมพันธ์ (Relational, B-Tree/GIN Indexing, CTE, Window Functions, Views) เพื่อให้ระบบทำงานได้เร็วที่สุด
4. **PostgreSQL ต้องมีรายละเอียดครบถ้วนในตัว (Self-Contained — ไม่พึ่งพา MongoDB อีก)**:
   - โครงสร้างตาราง / แถวข้อมูลใน PostgreSQL ต้อง Denormalize และ Enrich ข้อมูลที่จำเป็นในการคำนวณและออกรายงานให้ครบถ้วนในตัว (เช่น ชื่อภาษาต่างๆ, รหัสบาร์โค้ด, ข้อมูลอ้างอิง, สถานะ, หน่วยนับ)
   - **ตอนดึงข้อมูลจาก PostgreSQL ไปประมวลผล จะต้องจบในตัว 100% ห้ามมีการ query ข้ามกลับมาต่อหรือ join กับ MongoDB อีกเด็ดขาด** (Zero Cross-DB Runtime Dependency) เพื่อรักษาความเร็วสูงสุดและความเป็นอิสระของ Processing Engine


## กฎ: GitHub เก็บโค้ดอย่างเดียว — ไม่มี CI ต้องตรวจเองก่อน push (ตั้งโดยลุงจืด 2026-09-09)

ลุงจืดตัดสินใจว่า **จะไม่จ่ายเงินให้ GitHub อีก** และให้ GitHub ทำหน้าที่เดียวคือเก็บ/แชร์โค้ด (`origin`) — `.github/workflows/ci.yml` ถูกลบทิ้งถาวรแล้ว (บัญชีถูกล็อกเรื่อง billing ตั้งแต่ 2026-09-02 ทำให้ทุก run ตายก่อนเริ่ม job อยู่แล้ว)

1. **ห้าม AI ตัวใดสร้าง workflow ใหม่ใน `.github/`** หรือเสนอให้ "เปิด CI กลับมา" โดยไม่ได้ถามลุงจืดก่อน — รวมถึงห้ามย้าย `backend/.github/workflows/*.yaml` (ของค้างยุค repo เก่า ที่มี `git push origin main` และ push image ไป `ghcr.io/smlsoft/*`) ขึ้นมาที่ root เด็ดขาด
2. **ตัวตรวจจริงมีสามอย่างเท่านั้น** (ทุก clone ต้อง `npm run hooks:install` ครั้งหนึ่ง ไม่งั้น hook ทั้งสองตัวไม่ทำงาน):
   - `.githooks/pre-commit` — อัตโนมัติตอน commit แต่ตรวจแค่ `docs/reference/CODE-MAP.md` (ข้าม: `SKIP_CODE_MAP_CHECK=1`)
   - `.githooks/pre-push` — อัตโนมัติตอน push: รัน `tools/verify.sh codemap` เสมอ และเพิ่ม `frontend` เมื่อ commit ที่จะ push แตะไฟล์ใต้ `frontend/`; ไม่รัน target `backend` เพราะช้าเกินไป (เตือนให้รัน `npm run verify:all` เองแทน) — ข้าม: `SKIP_VERIFY=1 git push`
   - `tools/verify.sh` (คนสั่งเอง) — `npm run verify` = codemap + frontend lint/typecheck/test ก่อน push ทุกครั้ง, `npm run verify:all` = รวม backend/outbox/projection ก่อน deploy
3. **ห้ามเขียนในเอกสารหรือรายงานว่า "CI จะจับให้"** — ไม่มีอะไรตรวจให้อัตโนมัติหลัง push แล้ว ถ้าไม่ได้รัน `verify` เอง ให้บอกตรง ๆ ว่ายังไม่ได้ตรวจ (กฎ VERIFY BEFORE DONE)
4. คำสั่งใน `tools/verify.sh` คัดลอกจาก workflow เดิมแบบคำต่อคำ — ถ้าแก้คำสั่งทดสอบ ต้องแก้ที่นี่ที่เดียว และอัปเดต `docs/kms/11-testing-quality.md` §5 ในคอมมิตเดียวกัน

เหตุผลเต็ม + ทางเลือกที่พิจารณาแล้ว: `docs/kms/decisions/2026-09-09-github-storage-only.md`
