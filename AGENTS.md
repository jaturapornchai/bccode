# BC Ai Account — Codex Router

For every task under `D:\bccode`:

1. Read `D:\bccode\docs\README.md` in full.
2. Read every matching Jead-authored file under `D:\bccode\docs\**`.
3. Use `D:\bccode\AI_INDEX.md` only to locate implementation evidence.
4. Inspect the exact source, tests, schema, configuration, and runtime evidence required by the task.

This file is only a routing entrypoint. Source-of-truth boundaries are defined only in `D:\bccode\docs\README.md`.

## กฎ: DeepSeek เป็นที่ปรึกษา (ตั้งโดยลุงจืด 2026-08-29)

- เมื่อต้องการความเห็นที่สอง ติดขัด หรือก่อนตัดสินใจงานด้านดีไซน์/สถาปัตยกรรมที่ไม่มีกฎครอบ ให้ปรึกษา DeepSeek ผ่าน:
  `node scripts/ask-deepseek.mjs "โจทย์/บริบท"` (หรือ `--model deepseek-reasoner` สำหรับโจทย์ซับซ้อน)
- Key อยู่ใน `deepseek.env` (git-ignored ผ่าน pattern `*.env`) — **ห้าม commit key หรือคัดลอก key ลงไฟล์/log อื่น**
- **ห้ามส่ง** secret, password, token, ข้อมูลส่วนบุคคล หรือข้อมูลลับของระบบไปปรึกษา — ส่งเฉพาะโจทย์/โครงสร้างที่จำเป็น
- คำตอบของ DeepSeek เป็น **ความเห็นที่ปรึกษาเท่านั้น ไม่ใช่กฎ** — ยังต้องยึดลำดับความน่าเชื่อถือใน `docs/README.md` และการตัดสินใจสุดท้ายเป็นของลุงจืด

## กฎ: Kimi K3 Max เป็นที่ปรึกษาเช่นกัน (ตั้งโดยลุงจืด 2026-08-30)

- ใช้ผ่าน: ไม่มีโจทย์อะไรมา — มีแค่คำว่า "โจทย์" กับเครื่องหมายผิดพลาด

ถามมาให้ชัด:
- วิเคราะห์โจทย์ธุรกิจ SME ไทย 40+ ?
- รีวิวโจทย์เทคนิค (Next.js/Go/MongoDB) ?
- ตั้ง requirements/UI ให้ ?

ไม่ระบุมา = ตอบไม่ได้ อย่าเสียเวลากัน (default model kimi-k3-max; --model เปลี่ยนได้)
- Key อ่านจาก ~/.kimi/kimi-claw/openclaw.json ตอน runtime — **ห้าม copy key ลงไฟล์/log/commit**
- กฎความปลอดภัย/ขอบเขตเหมือน DeepSeek ทุกประการ: ห้ามส่ง secret/PII, คำตอบเป็นความเห็นที่ปรึกษาเท่านั้น

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

ตัวอย่างที่ผ่านมาตรฐาน: หน้า login + holding หลัง pass 2026-09-02 (block "Login premium pass 3" ท้าย `frontend/src/app/globals.css`)


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
