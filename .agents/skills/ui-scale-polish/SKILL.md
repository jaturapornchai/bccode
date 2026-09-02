---
name: ui-scale-polish
description: ระบบสเกล/density และ premium skin ของ BC Ai Account (token, cascade traps, GIS/motion gotchas, ตารางสิทธิ์, checklist ตรวจรับ) — ใช้เมื่อสร้าง/แก้ UI ทุกจอ (login, workspace, settings, menu) เพื่อให้สัดส่วนสวย อ่านง่าย สม่ำเสมอทุกจอทุก viewport และเวลาผู้ใช้บ่นว่า "ตัวหนังสือเล็ก/ใหญ่ ไม่สมส่วน ช่องไฟเยอะ จอไม่ตรงกัน ย้ายจอแล้วเพี้ยน" ให้โหลด skill นี้ก่อนแก้ CSS เสมอ
---

# UI Scale & Density System — BC Ai Account

ระบบทั้งหมดอยู่ที่ `frontend/src/app/globals.css` (tail) + `frontend/src/app/zoom-control.tsx`.
แก้ตามระบบนี้เท่านั้น ห้าม hardcode px ขนาดใหญ่ใน component ใหม่

## 0) บุคลิกหลัก: คนไทย อายุ 40+ (กฎหลัก 2026-08-30)

ผู้ใช้หลักคือคนไทยอายุเกิน 40 ปี (พนักงานบัญชี/เจ้าของกิจการ) — อ่านออก ใช้ได้
ไม่กลัวกดผิด มาก่อนความสวย/ทันสมัย (บัญญัติเต็มใน AGENTS.md):

- ตัวหนังสือที่ต้องอ่านตัดสินใจ ≥ 0.9rem; จิ๋ว (<0.8rem) ได้เฉพาะ metadata
- ไทยก่อน: ป้าย/ปุ่ม/feedback เป็นไทย; ศัพท์อังกฤษมี title/คำอธิบายไทย
- action สำคัญห้าม icon เปล่า — ต้องมีข้อความไทย หรือ aria-label+title ไทย
- จุดคลิก ≥ 2.6em; คอนทราสต์ AA; สถานะห้ามสื่อด้วยสีเดี่ยว
- ลบ/ทับ = dialog ไทยอธิบายผล; ทุก action มี feedback ไทยทันที
- transition ≤ 300ms ไม่เอฟเฟกต์เร็ว; ห้ามพาสายตากระโดด (auto-scroll)
- แก้ UI แล้วใช้ Playwright วัด (ฟอนต์/ปุ่ม/overlap) — อย่าเชื่อตาเปล่า

## 1) สเกลแบบ fluid (auto ทุกจอ)

Root font-size เดียวคุมทุกอย่าง (ทุก UI ใช้ rem):

```css
html { font-size: clamp(15px, calc(0.46875vw + 9px), 21px); }
/* 15px @≤1280 → 18px @1920 → 21px @≥2560 — คือของเดิม (10/12/14) ×1.5
   ตามกฎ "ทั้งระบบ 100% = 150%" (2026-08-29) */
```

- **หน้า login ยกเว้น**: `html[data-login-scale="1"]` (LoginWrapper **และ holding-screen**
   ตั้ง/ถอด attribute — /holding ใช้ class ชุด login ทั้งหมดจึงต้อง exempt ด้วย
   ไม่งั้น ×1.5 ทำปุ่ม .primary-button โตเกินของที่อนุมัติ) คืนสูตรเดิม
   `clamp(10px, 0.3125vw + 6px, 14px)` !important — หน้า login อนุมัติขนาด (×2 pass)
   ไปแล้ว ห้ามขยายตาม; !important จำเป็นเพื่อชนะ inline style ที่ zoom-control
   ตั้งค้างไว้ตอน SPA navigation จาก /menu
- `zoom-control.tsx → applyUiZoom()` ต้องใช้สูตรเดียวกัน:
  `fontSize = calc(clamp(15px, 0.46875vw + 9px, 21px) * zoom/100)` (ป้ายแสดง nominal 100%)
- control ติด listener `resize` + `focus` เพื่อ re-apply ตอนย้ายจอ/DPI
- **ห้าม**กลับไปใช้ media query เป็นขั้น (ต้นเหตุ "จอใน ZCode ไม่เท่า browser จริง")
- **บทเรียน rem ไม่ cascade**: ตั้ง font-size ที่ wrapper element ไม่กระทบ `rem`
  ข้างใน (rem ผูก html เสมอ) — จะ exempt หน้าใดต้องแก้ที่ระดับ html เท่านั้น
  (attribute + CSS !important) แนวตั้ง wrapper font-size ใช้ไม่ได้
- **บทเรียน grid track แน่นอน vs minmax**: คอลัมน์ที่ต้องการสัดส่วนตามจอ
  ให้ใช้ track เดี่ยว `clamp(520px, 34vw, 800px)` — `minmax(clamp(...), 800px)`
  ปล่อยให้เนื้อหาดันไปที่ cap เสมอ (เคส holding @1280 ฟอร์มยึด 800px จน
  split กลับด้าน 38/63)

**ความเท่ากันข้ามหน้า (holding ↔ login, 2026-08-29)**: หน้าที่ใช้ class ชุด login
(ตอนนี้มี /login และ /holding) ต้อง**จับคู่ขนาดตามบทบาท** ไม่ใช่แค่ไม่ใหญ่กว่า —
ช่องค้นหา = ช่องกรอก (1.8rem), ชื่อการ์ดรายการ = label (1.8rem), chip สรุป/code
badge/badge ในการ์ด = help text (1.25rem), CTA = ปุ่ม login (1.75rem) — วัดแล้วตรง
กันเป๊ะที่ 1280/1920/2560 (18/21.6/25.2 ฯลฯ) ดู block "Holding page type parity
with login" (scope `.holding-shell` เพื่อไม่โดนกริดของ /workspace)
**วิธีตรวจ**: วัด computed font-size ของ element คู่บทบาทที่เดียวกันระหว่างสองหน้า
ที่ทุก viewport ต้องออกมาเท่ากันเป๊ะ
- chrome ที่ล็อก px (ปุ่ม 36px, work-tab 34px) ไม่โตตาม root — ยอมรับโดยเจตนา
  (เป้าหมายคือข้อความใหญ่ขึ้น) ตรวจ clip หลังแก้สเกล:
  `button.scrollHeight > clientHeight` ทุกปุ่ม (ผลลัพธ์ล่าสุด = 0 clip)

## 2) Baseline rhythm (vertical grid)

```css
:root { --baseline: 0.25rem; --space-1..6 = baseline×1..6; --leading-tight/normal = ×5/×6 }
```

กฎ: line-height และช่องไฟแนวตั้งเป็นจำนวนเต็มเท่าของ `--baseline` (4/8/12/16px @root16) — ข้อความเรียงตารางตรงกันทุกจอ

**⚠️ กฎไทย (สำคัญสุด)**: ภาษาไทยมีสระบน/ล่างและวรรณยุกต์ซ้อน — **line-height ต่ำกว่า 1.4 จะทับกัน** (เช่น h1 "เข้าสู่ระบบ" ทับแท็กไลน์, หัวข้อทับเนื้อความ)
- หัวข้อ/ข้อความไทย: line-height ≥ **1.45** + `margin-block: 0.2em 0.3em` กันสระบนชนขอบบรรทัดก่อนหน้า
- `--leading-tight` (1.25) ใช้ได้เฉพาะ input เดี่ยว/single-line Latin — **ห้าม**ใช้กับหัวข้อไทย

## 3) Density (ข้อมูลให้มากที่สุด)

- ตาราง: `th/td padding 0.45em 0.7em`
- ช่องกรอก/select: `min-height 2.6em` (login lock 40px เฉพาะ .login-card)
- `.field-group gap = --space-2`
- ปุ่ม/ไอคอนใน `.input-with-icon` เป็น **em ทั้งหมด** (ห้าม px ตายตัว — px จะทับตัวหนังสือตอนสเกลเปลี่ยน)

**Density round 4 (2026-08-28)** — control set 40→**36px** ทั้งชุด (ปลดล็อกได้เพราะหน้าตาปุ่ม Google เป็น CSS ของเราแล้ว iframe จึงสูงเท่าไรก็ได้);
icon ในช่องกรอกเล็กลง `1.1em` ชิดขอบ `left 0.55em` + `padding-left 2.05em` → ตัวอักษรได้พื้นที่มากขึ้น

**Density round 3 (2026-08-28)** — บีบฟอร์ม/login อีกขั้น: `.login-card gap 0.55rem / padding 0.9-1.1rem`,
`.field-group gap 0.125rem`, heading padding 0.3rem, section gap 0.5rem;
global: `gap-3/4 → 0.45/0.55rem`, `space-y-3/4`, `mt/mb-3/4 → 0.4rem` (exact-token เท่านั้น)

**กฎแยก chrome กับ content (สำคัญ)**: density pass มีเป้าคือ CHROME (toolbar/nav/tab/ปุ่ม) —
**การ์ดเนื้อหาที่มีข้อความอ่านจริง** (การ์ดเลือกบริษัท/สาขา `group/company`, `group/branch` ฯลฯ)
ต้องคืนอากาศ: padding 1rem+, line-height 1.45-1.5, badge มี margin — ไม่งั้นข้อความแน่นจนอ่านลำบาก
(block "Content cards breathe again" ใน globals.css)

**Vertical padding reduction (ทั้งระบบ)** — ลด padding แนวตั้งของ Tailwind utilities
(block "Global vertical padding reduction" ท้าย globals.css): `pt/pb/py/p 4-6 → 0.6rem`,
`pt/pb-2 → 0.35rem` — ใช้ exact-token `[class~="pt-4"]` เท่านั้น, **แนวนอนไม่แตะ** (ความยาวบรรทัดอ่านง่าย)
และ hero login มี rhythm ของตัวเองอยู่แล้ว

**Utility-level density (ทั้งระบบ)** — Tailwind padding utilities เรนเดอร์กระชับลงหนึ่งสเต็ป
(block "Global density via Tailwind utilities" ท้าย globals.css): `p-4→0.8rem`, `p-3→0.65rem`,
`px-3→0.7rem`, `py-2→0.45rem`, `rounded-2xl→0.9rem` — ใช้ exact-token match เท่านั้น
(`[class~="p-3"]`) เพื่อไม่โดน responsive variant (md:p-3) และไม่ทับ padding ที่กำหนดเอง
(.login-card ฯลฯ) — จอใหม่/หน้าใหม่จะได้ density เดียวกันโดยอัตโนมัติ

## 3.1) กฎถาวร: ทุกครั้งที่แก้ UX/UI ต้องสะท้อนกลับเข้า skill นี้

เมื่อผู้ใช้ให้แก้ UX/UI ใด ๆ: หลังแก้ของจริงเสร็จ ให้**สร้างหรือปรับปรุง skill นี้ทุกครั้ง**
(เพิ่มกฎ/ค่าที่ใช้/กับดักที่เจอ) เพื่อให้ทุกหน้าจอทำงานคล้ายกันและแก้ซ้ำไม่ต้องค้นใหม่ —
สิ่งที่ต้องสะท้อน: ค่าที่ใช้ (px/rem/em), pattern ที่ได้ผล, กับดัก, และ checklist การตรวจ

## 4) ปุ่มต้องเหมือนกันเป๊ะ

ปุ่ม Google ถูกวาดใน iframe สูงตายตัว 40px → ทุก control ในกลุ่มเดียวกัน lock 40px/6px radius/14px-600 (ดู block "Uniform controls" ใน globals.css) ห้ามให้ปุ่มใดสูงต่างกันในแถวเดียวกัน

**เทคนิคบังคับทั้งแถวด้วย :has()** — อย่าไล่ override ทีละ class (component บางตัวครอบ wrapper ตัวเอง ทำให้ selector ลงไม่ครบ):

```css
div[class*="ml-auto"]:has(.zoom-control) button {
  height: 36px !important; min-height: 36px !important; max-height: 36px !important;
}
```

จับที่ container ของแถวที่มี marker ชัด (เช่น .zoom-control) แล้วบังคับปุ่มทุกปุ่มในนั้น

## 4.1) หลีกเลี่ยง iframe (กฎ)

**พยายามไม่ใช้ iframe ถ้าไม่จำเป็น** — ใช้ได้เฉพาะเมื่อ third-party บังคับ (เช่น ปุ่ม Google/GIS) เพราะ:

- ความกว้าง/สูงเป็น px ตายตัว ไม่ยืดตาม root scale → ต้อง override CSS หลายชั้น (wrapper div + iframe) และพังเมื่อสเกลเปลี่ยน
- ซ่อน DOM จากการตรวจสอบ/จัดสไตล์, theme (dark mode) ไม่ตาม, และ render ซ้ำเมื่อย้ายจอ

ถ้าจำเป็นต้องใช้ (เคส GIS ปุ่ม Google):

1. ส่ง width ที่ตรงกับ container จริงเข้า `renderButton`
2. override ทั้ง **wrapper div และ iframe** ให้ยืด 100% (ดู block `.gis-button-host` ใน globals.css)
3. จำกัดความสูงให้เท่าปุ่มข้างเคียง (uniform-height rule หัวข้อ 4)

## 4.2) ปุ่มทุกปุ่มใช้ CSS ชุดเดียวกัน (global button system)

อย่าไล่แก้ปุ่มเป็น class ๆ — สร้างชุดกฎกลางใน globals.css ที่ครอบทุก variant ของ shadcn/Tailwind
พร้อมกันในหน้าเดียว (outline/ghost/secondary/icon) แล้วทุกหน้าได้เหมือนกันอัตโนมัติ:

```css
button[class*="rounded-lg"], button[class*="h-8"], button[class*="h-9"],
.header-control-button, .icon-button, .zoom-control-trigger {
  font-size: 0.8rem; font-weight: 600; line-height: 1.4;
  height: 36px; min-height: 36px; border-radius: 6px;
  padding-inline: 0.7em; gap: 0.45em;
}
```

กฎเสริม: variant ต่างกันได้ที่ "สี/พื้น" เท่านั้น (outline=ขอบ+พื้นขาว, ghost=โปร่ง,
primary=เขียวเข้ม) — เงื่อนไขบังคับ: **ความสูง รัศมี ฟอนต์ ต้องเท่ากันทุกปุ่มในหน้าเดียวกันเสมอ**

**ห้ามล็อก font-size เป็น px (2026-08-29)** — block "uniform controls" เดิมตั้ง `font-size: 14px`
ให้ปุ่ม Google/Dev/CTA + ป้าย Google + ข้อความ error banner = ตัวหนังสือกลุ่มเดียวบนหน้า login ที่
ไม่ขยายตาม fluid root (10px@1280 → 14px@2560) แก้เป็น rem หมดแล้ว: 14px → `0.875rem`, banner
0.875/0.8125/0.75/0.6875rem, svg Google 18px → `1.2857em` — ดู block "Login fluid text" ท้าย
globals.css **วิธี audit**: tree-walker หา element ที่มี text node แล้วเทียบ computed font-size
ราย element ระหว่าง 1280/1920/2560 (resize โดยไม่ reload) — **ห้าม group ตาม selector**
(span ไร้คลาสจากหลายปุ่ม merge กันแล้วปิดบังค่าที่ล็อก) ค่าใดเท่ากันทุก viewport = ล็อก px ต้องแก้

**Login ×2 type scale (2026-08-29): ผู้ใช้ขอ "ขนาดอีก 200%"** — หน้า login ใช้ type scale
สองเท่าของระบบปกติ (hero + card ทั้งหน้า): h1 hero `clamp(3.8rem, 5.2vw, 8rem)`, feature
strong/span 1.8/1.56rem, status 1.56rem, การ์ด login: label/input 1.8rem, ปุ่มทุกปุ่ม 1.75rem
(h2 3.3rem) — ดู block "Hero type scale ×2" และ "Login card type ×2" ท้าย globals.css
**กฎติดตาม**: (1) ปุ่ม Google ที่มองเห็นอยู่บน `.gis-wrap .google-face` ไม่ใช่
`.gis-button-host` — ขยายผิดที่แล้วป้าย Google เล็กกว่าปุ่มข้าง ๆ; (2) control height
ต้องเป็น **em/min-height 2.6em** ไม่ใช่ px (36px เดิม) เพราะฟอนต์ใหญ่แล้ว height ตายตัวจะบีบข้อความ;
(3) **ความสูงเท่ากัน (กับดัก .gis-wrap)**: `min-height` แบบ em คำนวณจาก font-size **ของตัว element
เอง** — ถ้า `.gis-wrap` ไม่ได้ตั้ง font-size เดียวกับปุ่มพี่น้อง (1.75rem) มันจะเตี้ยกว่า
(เคสจริง: Google ~38px ข้าง Dev/CTA 64px) ต้องตั้ง font-size ที่ wrap + `align-items: stretch`
ที่ grid แล้ววัดยืนยันทั้ง 3 ปุ่มด้วย getBoundingClientRect

**สแกนตัวหนังสือทับกัน (overlap scan)** — เทียบ text line boxes (Range.getClientRects)
ทุกคู่ที่ไม่ใช่ ancestor กัน ถ้า intersect > 3px = ทับจริง; ใช้ตอนผู้ใช้บ่น "ทับกัน"
และใช้ยืนยันหลังแก้ (เคสจริง: badge chip ทับ 7px จาก leading คู่กับ font ใหญ่ แก้ด้วย
line-height ชัด + gap แถว) — **ระวัง false positive**: label โปร่งใสใน iframe ของ GIS
(`nsm7Bb-*`) ทับกับป้ายที่เราวาดทับมันโดยตั้งใจ ไม่ใช่บั๊ก; ตรวจ clip ด้วย
`scrollHeight > clientHeight` ของกล่องข้อความ
**บั๊ก "ทับเมื่อกด" ต้องไล่กดจริงทุก state (2026-08-29 บทเรียน settings wizard)**:
scan ตอนเปิดหน้าอย่างเดียวไม่พอ — ทับส่วนใหญ่โผล่หลังกด step/ย่อเมนู (พบ 3 จุด:
helper sidebar ล้นปุ่มที่โดนล็อกสูง 36px ไปทับปุ่มถัดไป 92x5px, route ยาว
`/transaction/...` ใน `.bc-list-row` ไม่ตัดบรรทัดทับ cell ข้าง ๆ 28x14px, h2.text-xl
สระบนชนบรรทัดบน 4px) ต้อง click-through ทุก step + collapse + หลาย viewport แล้ว
scan ทุกจังหวะ

**กับดัก Tailwind `break-words` กับ token ยาวไร้ช่องว่าง**: `overflow-wrap: break-word`
**ไม่ตัด** string ยาวต่อเนื่อง (เช่น `/transaction/purchaserequisition`) ใน flex cell
ที่แคบ — ข้อความอยู่บรรทัดเดียวล้นกล่องทับ cell ข้าง ๆ ต้องใช้
`overflow-wrap: anywhere` (มีผลกับ min-content ด้วย ยอมให้ shrink+wrap)

**กับดัก chrome-height lock กับปุ่มที่ wrap ได้**: ปุ่ม nav ที่ label เป็นไทยยาว wrap
หลายบรรทัด (sidebar settings) ห้ามตกอยู่ใต้กฎ lock 36px — ให้ `height:auto;
max-height:none` + min-height em เฉพาะ aside นั้น ไม่งั้นบรรทัดที่ 2 ล้นออกนอกปุ่ม

## 4.3) Two-tier + บทเรียน text-xs!

- ระบบสองชั้น: action=36px / navigation row=30px — ฟอนต์ 0.8rem+radius 6 เท่ากันทุกปุ่ม
- **กับดัก Tailwind `text-xs!` (important)**: CSS override ไม่ชนะ (important + layer กลับลำดับ)
  แม้ specificity สูงสุด — **ต้องแก้ markup ตรง ๆ** (เปลี่ยนเป็น text-sm หรือลบ !) เช่นเคส
  section button ใน main-menu-screen.tsx
- ตรวจผลด้วย: `fontSize distribution` ของทุก button ในหน้า ต้องเหลือค่าเดียว (เช่น 19/19 = 0.8rem)

## 4.4) Icon–text pairing + baseline (สองชุด)

- **alignment ใช้ `align-items: center` ทุกปุ่ม** — ห้ามใช้ `first baseline` (ทำให้ icon ในปุ่มไร้ข้อความกระโดดขึ้น เพราะไม่มี baseline ให้เกาะ; line-height 1.4 ที่สม่ำเสมอทำให้ text อยู่แนวเดียวกันเอง) — วัด: padding บน=ล่างของ icon ในปุ่ม (เช่น 12/12)
- **ปุ่มมีข้อความ**: icon `1em` คู่ขนาดตัวอักษร (≈10px ที่ 0.8rem)
- **ปุ่ม icon ล้วน** (bell/T/palette/flag/w-8/w-9/.icon-button): icon `1.25em` (≈12px) — อย่าใช้ 1em เพราะจะเล็กเกินจนปุ่มดูโหว่ (บทเรียนจาก round แรก)
- วัดผล: text-pair iconH ≈ fontSize (10≈9.69) / icon-only iconH ≈ 1.25×fontSize (12)

## 4.5) Sidebar labels: wrap ไม่ truncate

sidebar ตั้งค่า (`aside[class*="md:w-60"]`) — ป้ายขั้นยาว (ไทย) **wrap ลงบรรทัดใหม่** แทน `truncate`
(ตัด+จุดไข่ปลา): `white-space: normal + overflow-wrap: anywhere + line-height 1.45` และปุ่ม
`min-height 2.6em + align-items: flex-start` ให้จังหวะสองบรรทัดเรียบร้อย — ใช้กับ sidebar/wizard nav
ทุกที่ (chrome แถบบนหรือ tab สั้นยัง truncate ได้ถ้าไม่กระทบความหมาย)

## 4.6) การ์ดเลือก = แถวรายการเต็มความกว้าง (list rows) ไม่ใช่การ์ดลอย

การ์ดเลือกรายการ (เลือกบริษัท/สาขา บน /workspace): ใช้ **แถวเต็มความกว้างเรียงคอลัมน์เดียว**
(`flex flex-col gap-3 max-w-3xl mx-auto` บน container) — การ์ดลอยเล็ก 350px กลางพื้นที่ว่างมหาศาล
ทำให้ "เลือกยาก" โครงสร้างแถว: [bar accent 4px] [logo 44px] [ชื่อ+badges flex-1] [code badge] [ArrowRight]
hover: border เขียว + พื้น accent + icon เลื่อนขวา — อ่านสแกนง่าย คลิก target เต็มแถว

**ตำแหน่ง CTA (2026-08-29)**: ปุ่ม action ของรายการ (เช่น "เพิ่มกลุ่มกิจการ" บน /holding)
วาง**ท้าย section หลังรายการ** เป็นแถบเต็มความกว้างการ์ด (`.holding-add-row` + `width: 100%`)
— อยู่ตรง content ที่มันสร้างเพิ่ม ไม่แย่งพื้นที่แถวค้นหา และไล่ลำดับอ่าน: ดูรายการ → เพิ่ม

## 4.7) Login design tokens (cinematic hero + refined card)

หน้า login: hero เป็นแผงเข้มไล่เฉด navy→teal ตัวอักษรขาว (`.brand-hero` + `.brand-copy`),
การ์ด feature/status เป็น dark glass (ขาว 9-12% + ขอบขาว 16-22%) ไม่ใช่ขาวจางบนภาพสว่าง,
การ์ดฟอร์มขาว radius 1.1rem เงาสองชั้น, CTA solid teal gradient (#0d9488→#0f766e) ไม่จาง,
icon row บนการ์ดเป็น ghost squares — ดู block "Login redesign v2" ท้าย globals.css
**(อัปเดต 2026-08-29: ghost squares ถูกแทนด้วยปุ่มจริงตาม §4.9 เพราะผู้ใช้บอกว่า "icon เล็กไป
และต้องมีลักษณะเป็นปุ่ม")**

## 4.9) Card-header controls: ปุ่มจริง สเกล fluid (2026-08-29)

กติกาจากผู้ใช้: กลุ่มปุ่ม icon บนหัวการ์ด (ฟอนต์/พาเลตต์/ธีม/ธงภาษา/ตั้งค่า) ต้อง "เหมือนปุ่ม"
มีขอบ+พื้น+hover และไอคอนใหญ่พอ — block "Login-family card header controls" ท้าย globals.css:

- **Scope `.card-header .header-controls`** (= การ์ด login + holding เท่านั้น ไม่โดนเมนูอื่น)
  `font-size: 1.5rem !important` บนปุ่ม, กล่อง **3em** (45px@1280 → 54px@1920 ตาม ladder),
  svg **1.4em** (ไม่ใช่ 18px attr เดิม), ธง `2em × 1.45em`, ขอบ `var(--line)` +
  พื้น `var(--panel-soft)` + `shadow-card` + hover `translateY(-1px)`
- **กับดัก 3 ชั้นที่ต้อง !important/ลำดับถูก:**
  1. บล็อก "Menu chrome uniformity" ล็อก `.header-controls *` เป็น 36px !important —
     ต้อง specificity สูงกว่า (`(0,3,0)`) + !important + อยู่ท้ายไฟล์ จึงหลุดได้
  2. บล็อก two-tier text ตั้ง `.header-controls .language-dialog-trigger` เป็น
     `font-size: 0.8rem !important` — **important ชนะ non-important เสมอไม่สน specificity/
     ลำดับ** → font-size ใน block ใหม่ต้อง !important เท่านั้น (ไม่งั้นกล่อง 3em กลายเป็น 24px)
  3. บล็อก "ghost squares" ตั้ง `border: 1px solid transparent` specificity เท่ากัน —
     block ท้ายกว่าชนะ แต่ต้องตรวจ **สีขอบ** จริง (assert `borderColor !== 'rgba(0,0,0,0)'`
     ไม่ใช่แค่ border-width เพราะ transparent ก็มี width 1)
- **จุดสีธีม (.theme-palette-dot) ชนไอคอนพาเลตต์**: ห้ามหดไอคอนพาเลตต์ลง (1.3em แล้วเป็น
  ตัวเตี้ยสุดเอง ตามคำผู้ใช้ "ความสูงต้องเหมือนกัน") — ไอคอนคง **1.4em เท่ากันทุกตัว** +
  `translate(-0.14em,-0.14em)` เลื่อนพ้นมุม + จุด 0.56em ที่ right/bottom 0.26em
- **ธงชาติ**: `2.1em × 1.4em` — สูง 1.4em เท่าไอคอน svg เป๊ะ (2.1em = อัตราส่วนธง 3:2);
  ธงเป็นสีทึบจะ "ดู" ใหญ่กว่าไอคอนลายเส้น ขนาดกล่องพอ ๆ กัน — วัด DOM ยืนยันความสูงเท่ากัน
- **วัดยังไงไม่ให้โกหก**: ปุ่มภาษามี svg ลูกโลก display:none อยู่ก่อน img ธง —
  ต้องเลือก "ไอคอนตัวแรกที่ rect.width > 0" ไม่ใช่ querySelector ตัวแรก
- **Regression**: `tests/login-header-controls.spec.ts` (HC-01..04): ขนาด @1280/1920/700,
  fluid ratio = root ratio (10→12px = 1.2), สูงเท่ากัน ±1px, ขอบมองเห็น, hover เปลี่ยน,
  holding parity กับ login, และหน้าอื่น (non-card-header) ยังล็อก ≤40px

## 4.10) ปุ่ม clear (X) ในช่องกรอก (2026-08-29)

กติกาจากผู้ใช้: ช่องกรอกต้องมี X ล้างค่าด้านใน textbox — block "Login card: clear (X)
buttons" ท้าย globals.css + pattern ใน login-screen.tsx:

- **X mount เฉพาะเมื่อมีข้อความ** และใส่ class `has-clear` บน `.input-with-icon` พร้อมกัน
  → input reserve `padding-right` 3.2em เฉพาะตอน X มีอยู่ (ไม่มี X = padding ปกติ)
- **ช่องที่มี trailing control (ตา) อยู่แล้ว**: X ใส่ class `before-trailing` จอดซ้ายของตา
  (`right: calc(0.4em + 2.6rem + 0.2em)`) และ password ใช้ padding-right พิเศษคลุมทั้ง X+ตา
- ขนาด/ตำแหน่งเลียน `.input-trailing-icon` เป๊ะ (login-card 2.6rem, right 0.4em) —
  em/rem ตาม fluid scale, hover พื้นเทา + focus ring
- **aria-label ต่อช่อง**: `${t(lang,"clearField")} ${t(lang,label)}` → "ล้าง ชื่อผู้ใช้";
  key ใหม่ `clearField` เพิ่มตามกฎ Thai-First (th.json ก่อน → copy เป็น placeholder
  ทุก locale; i18n drift test เฝ้าอยู่แล้ว)
- **Regression**: `tests/login-clear-buttons.spec.ts` (CB-01..03): mount/unmount,
  aria-label, X ไม่ทับตา (boundingBox เทียบกัน), ตายัง toggle ได้, ล้างแล้ว CTA กลับ
  disabled
- กับดัก Playwright: `boundingBox()` คืน plain object `{x,y,width,height}` —
  ไม่มี `.right()` แบบ DOMRect ต้องบวกเอง

## 4.11) Login card: กล่องใหญ่ขึ้นด้วย padding บางลง (2026-08-29)

"ลด padding ลง อีก กล่อง login จะได้ใหญ่ขึ้น" — block "Login card: slimmer
padding, wider box" ท้าย globals.css:

- `.login-shell:not(.holding-shell) .login-card` — **:not(.holding-shell) จำเป็น**
  เพราะ holding ใช้ .login-card เหมือนกันแต่ geometry ของมันถูกอนุมัติไปแล้ว
  (คอลัมน์ clamp(520px,34vw,800px) การ์ดต้องเต็มคอลัมน์)
- ค่าที่ใช้: `max-width: min(100%, 50rem, 532px)` (จาก 44rem), padding
  `clamp(0.85rem,1.1vw,1.25rem) clamp(1rem,1.4vw,1.7rem)` (จาก 1.2-1.8/1.4-2.1rem),
  gap `0.8rem` (จาก 1.05rem) → @1280 กล่อง 446→496px, ช่องกรอก 410→460px
- ตรวจหลังแก้: scrollHeight ยังพอดี viewport ทุกจุด (800/800, 1080/1080) +
  overlap 0 + vision pass — ลด padding แล้ว Thai line-height (§2) ห้ามลดตาม


## 4.8) Login เต็มจอ (full-viewport split-screen) + กฎคอนทราสต์ตัวหนังสือบนภาพ (2026-08-29)

**Layout เต็มจอ (≥1024px)** — ดู block "Login full-viewport layout" ท้าย globals.css:

- `.login-shell`: hero panel = `minmax(0, 1fr)` full-bleed (shell `padding: 0`,
  brand-panel `border-radius: 0; border: 0`, `min-height: 100dvh` ทั้งสองแผง),
  form column = `minmax(clamp(400px, 28vw, 560px), 560px)` ชิดขวา
- **ห้ามกลับไปใช้คอลัมน์ตายตัว + `justify-content: center`** — เคยทำให้เป็นแถบกลางจอ
  เหลือพื้นว่างข้างละ ~320px @1920 และ ~500px @2560 (ที่มาของ "ไม่เต็มจอ")
- เนื้อหาใน hero (`.brand-hero`, `.brand-feature-grid`, `.brand-status-strip`)
  **ยืดเต็ม 100% ของ interior แผง** (`max-width: 100%`, padding แผงคือ inset เดียว)
  — วัดจริง 100% ที่ 1280/1920/2560; ข้อความชิดซ้ายในกล่องกว้าง = อ่านเป็น cinematic
  banner (cap เดิม 820→1040px ถูกบ่นว่า "background เยอะไป / ขนาด 100%" คือ
  ยืดเต็มแผง); card ฟอร์ม `max-width: min(100%, 44rem)` = เต็มคอลัมน์ยกเว้น
  padding ขอบ (วัดจริง 79%/83%/83% ของคอลัมน์ที่ 1280/1920/2560 — เดิม 55%
  ที่ 1280 คือจุดที่ผู้ใช้บ่น) banner error กว้างตาม card (`44rem`)
- **901–1023px (แท็บเล็ตแนวตั้ง)**: brand panel ถูกซ่อนอยู่แล้ว → ต้อง collapse
  `.login-shell` เหลือ `minmax(0, 1fr)` คอลัมน์เดียว (block fluid ≥901 เก่าจอง
  คอลัมน์ที่สองไว้ ทำให้ฟอร์มถูกหดเหลือ ~525px มีคอลัมน์ผีข้างหลัง)

**กฎคอนทราสต์ (บังคับทุกครั้งที่วางตัวหนังสือบนภาพ/glass)** — ดู block
"Login text contrast on photo backgrounds":

- ตัวหนังสือทุกตัวต้องมีพื้นหลัง**ทึบพอ** — glass โปร่งบนภาพสว่างทำข้อความจาง
  (เคสจริง: feature card glass 0.55 + ตัวอักษร alpha 0.75 บนภาพร้านค้าสว่าง = อ่านแทบไม่ออก)
- ค่าที่ผ่านตา: hero gradient จุดท้าย (ฝั่งสว่าง) ≥ 0.64, feature card bg
  `rgba(8,24,40,0.8)`, description alpha 0.92, status strip bg 0.84 + ตัวอักษร
  `#aef0e2`
- **ตัวหนังสือเข้ม (เช่น `#0b1720`) ห้ามวางบนภาพตรง ๆ** ต้องมี chip ขาวรอง —
  `.brand-badge-row` ใช้ `background: rgba(255,255,255,0.8)` + blur + radius 0.9rem
- ตรวจ: เปิดหน้าจริงบนภาพพื้นหลังสว่างสุดของชุดภาพ แล้วกวาดตาทุกข้อความบนแผงเข้ม

## 5) Workflow แก้ UI หน้าใดหน้าหนึ่ง

1. แก้เฉพาะ scoped class ของหน้านั้น (`.login-card`, `.form-panel` ฯลฯ) — ค่าเป็น rem/em/clamp เท่านั้น
2. ถ้าต้องขยายเฉพาะจอใหญ่ ใช้ clamp ไม่ใช่ media query ขั้น
3. build+deploy:
   `cd frontend && docker build --build-arg BCAI_LOCAL_BACKEND_URL=http://mainapi:8888 --build-arg NEXT_PUBLIC_GOOGLE_CLIENT_ID=212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com -t bcai-account-frontend:rYYYYMMDD-N .`
   แล้ว recreate container `bcai-account-frontend-local` ด้วย env เดิม (ดู command เดิมในประวัติ)
   **และต้อง `docker network connect backend_app-network bcai-account-frontend-local` ด้วยทุกครั้ง** —
   ถ้าลืม DNS จะหา `mainapi` ไม่เจอ → หน้า login ขึ้น banner "เชื่อมต่อ Backend ไม่ได้" ทันที
4. **ตรวจ 3 ขนาด viewport โดยไม่ reload**: 1280×800 → 1920×1080 → 2560×1440 (setViewportSize + วัด `getComputedStyle(document.documentElement).fontSize` = 10/12/14px) แล้วถ่ายภาพเทียบ
5. ตรวจภาพด้วยตา: หัวข้อไทยต้องไม่มีสระ/วรรณยุกต์ทับบรรทัดถัดไป, ปุ่มแถวเดียวกันสูงเท่ากัน, ไอคอนไม่ทับตัวหนังสือ

## 6) กับดักที่เจอแล้ว

- `.dockerignore` ระดับ root (`/models/`, `/role_permission_http.go`) — อย่าเพิ่ม pattern ไม่ anchor เพราะมันตัดซอร์สซ้อนและ build cache ทำให้ deploy โค้ดเก่าเงียบ ๆ
- `docker build` ล้มเงียบ: อย่าอ่านแค่ `tail -2` — ต้อง grep "ERROR" หรือ `--no-cache` เมื่อสงสัย
- ไอคอน px ตายตัว (left:12px size:18px) + padding rem = ทับกันเมื่อสเกลเปลี่ยน → ใช้ em เสมอ
- evaluate ใน browser tool: ใช้ expression string อ่านอย่างเดียว (function form ถูก reject)
- **ตัวหนังสือไทย + line-height tight = ทับกัน** (เคสจริง: `.brand-copy h1` line-height 1.25 ทับแท็กไลน์ — แก้เป็น 1.45 + margin-block)

## 4.12) Modal บนหน้า login-family: สเกล "เหมือนชาวบ้าน" (2026-08-29)

"ปรับ font ให้เหมือนชาวบ้าน" + "ปุ่มด้วย" — modal (`.holding-modal-panel` สร้าง/แก้
กลุ่มกิจการบน /holding) เคยสับสนขนาด: หัวเรื่อง/caption 0.78-0.94rem (ย่อสุด),
label+input 1.8rem (ติดบล็อก ×2 ของ .login-card), ปุ่มคู่เดียวกัน = 9px (--text-sm)
กับ 17.5px (1.75rem) บล็อก "Holding create/edit modal: ordinary form scale" ท้าย
globals.css แก้เป็นสเกลฟอร์มธรรมดา:

- หัวเรื่อง 1.5rem/800, คำอธิบาย+help ยาว 1.2rem/600, label 1.4rem/700, input 1.4rem,
  field-help 1.15rem/600, เลขยืนยัน 1.5rem
- **ปุ่มทุกปุ่มใน modal = 1.4rem/700 + min-height 3em เท่ากัน** (`.secondary-button`
  เดิมใช้ --text-sm จึงจิ๋วกว่าปุ่ม primary ที่โดนบล็อก ×2 คนละเท่า)
- **กับดัก grid-column**: media block ของหน้า login ตั้ง `.primary-button
  { grid-column: 1 / -1 }` — ใน modal (grid 2 คอลัมน์ .holding-create-actions) มัน
  ดันปุ่มสร้างลงแถวเต็มกว้างทิ้ง ยกเลิก ไว้ลำพัง → ต้อง restore
  `.holding-modal-panel .primary-button { grid-column: auto }`
- กติกาทั่วไป: modal/dialog บนหน้าพิเศษ (login-family) อย่าปล่อยสืบทอดสเกล ×2 ของ
  หน้า — กำหนดสเกลฟอร์มปกติของตัวเอง + ปิดทางที่ media query ของหน้าหลักรั่วเข้ามา
- แก้ CSS แล้วต้อง rebuild image ก่อนวัด (container คือ build output ไม่ใช่ source)

## 4.13) เมนูหลัก: workspace context เป็น breadcrumb ใต้ header (2026-08-30)

ผู้ใช้ขอย้ายแถบบริบทออกจากเนื้อหา → ลอง footer ติดหมุดท้าย sidebar (3 chips
ซ้อนแนวตั้ง ~150px) ผู้ใช้บอก "ไม่สวย ขอไอเดีย" → เสนอ 4 ตัวเลือก (breadcrumb ใต้
header / switcher หัว sidebar / status bar ล่างเต็มจอ / ย่อ sidebar ล่าง 1 บรรทัด)
ผู้ใช้เลือก **breadcrumb ใต้ header** (DeepSeek แนะนำตรงกัน):

- `<nav class="menu-breadcrumb">` วาง**ถัดจาก `</header>` ใน section เนื้อหา** (ไม่
  อยู่ใน header เพราะ header auto-hide กับ topChromeHidden — breadcrumb ต้องค้าง)
- โครง 1 บรรทัด: icon(Crown/Building2/GitBranch) + ชื่อ คั่น ChevronRight,
  `flex-wrap` + ชื่อ `truncate max-w-[18rem]` (ชื่อยาวตัดไม่ล้น), py-1 → สูง ~27px
- **ห้ามซ้ำรหัส**: display name มีรหัสอยู่แล้ว ("บ้านเชียง (bc001)", "[TST03] …")
  อย่าเพิ่ม chip รหัสแยก — เคยใส่แล้วซ้ำสองชั้น (bc001bc001)
- sidebar คืนตำแหน่งเดิม (ปุ่มซ่อนเมนูลอย sticky bottom-3 แบบเดิม) + ลบ
  WorkspaceContextPanel ที่ตายแล้วทิ้ง (dead code)
- ก่อนย้าย element ออกจาก scope ต้อง grep CSS hooks ของ scope นั้นก่อนทุกครั้ง
  (`.menu-top-chrome ...` ผูกแค่ปุ่ม controls จึงย้ายได้)
- **บทเรียน build trap**: `docker build .` ต้องรันจาก `frontend/` เสมอ — รันจาก
  root แล้ว chain `&& rm; run` จะแอบ reuse image tag เก่าใน local ทำให้เห็นของเก่า
  ทั้งที่ build เหมือนสำเร็จ (tail -1 พิมพ์บรรทัดเดียวกันทั้งสำเร็จ/ล้ม — เช็ค
  exit หรือดู "naming to" line ด้วย)


## 4.14) เมนูหลัก: ช่องค้นหาอยู่ที่เดียว — sidebar เป็นหลัก (2026-08-30)

"ค้นหาเมนูเอาออก เพราะมีอยู่แล้ว ใน เมนู section" — ช่องค้นหาใน header
(menu-top-chrome) ซ้ำกับของใน sidebar (ผูก globalSearch ค่าเดียวกัน):
**เรนเดอร์ header search เฉพาะเมื่อ sidebar ไม่ visible**
(`{!showLeftMenu && <label .../>}` เมื่อ `showLeftMenu = menuLayout === "left"
&& !sidebarHidden`) — sidebar อยู่ = header สะอาด (สูง 92→50px ด้วย), ซ่อน
sidebar/โหมดเมนูบน = ช่องค้นหา header กลับมาเป็น fallback (ห้ามลบทิ้งตรง ๆ
เพราะบางสถานะจะไม่มีทางค้นหาเลย)

## 4.15) ค้นหาเมนูแบบ full-text ทุกภาษา (2026-08-30)

"ค้นหาไม่เจอ ต้องแบบ full text search ด้วย ได้ทุกภาษา" — เดิมจับแค่
`menuText(label, ภาษาปัจจุบัน)` ทำให้พิมพ์อังกฤษใน UI ไทยไม่เจอเลย

- **lib/menu-data.ts**: `normalizeMenuSearchText()` = ตัด zero-width +
  **ตัดวรรณยุกต์/สระบนล่างไทย (U+0E31, U+0E34-3A, U+0E47-4E) ทั้งสองฝั่ง** +
  toLocaleLowerCase + collapse space — พิมพ์ผิดวรรณยุกต์/ตัวพิมพ์ใหญ่ก็เจอ
- `menuSearchHaystack(label, dict)` = th + en + ทุกภาษาที่ label มี + key +
  dictionary override → **query ภาษาไหนก็ match ได้**; item รวม route+id ด้วย
- main-menu-screen: getVisibleGroups/Items/getMenuTreeNodes ใช้ matcher ร่วม
- ผลวัดจริง: "sale" บน UI ไทย 0→45 รายการ, "รา้ยงาน" (ผิดวรรณยุกต์) = 24 เท่า
  "รายงาน" ปกติ, REPORT พิมพ์ใหญ่เจอ
- กับดัก TS: `Object.keys(label) as LanguageCode[]` แล้วเทียบ `"key"` →
  TS2367 no-overlap — iterate string keys ธรรมดา + cast Record<string,unknown>
- เคส query ไม่เจอเลย: หัว section 5 ตัวยัง render (พฤติกรรมเดิม) — ถ้าจะทำ
  empty-state "ไม่พบเมนู" ให้ทำภายหลัง

## 4.16) ค้นหาในโหมดเมนูบน: ผลค้นหาเป็นการ์ดในพื้นที่เนื้อหา (2026-08-30)

"เลือกเมนูบน แล้ว ค้นหา จะให้แสดงยังไง" — โหมดเมนูบน/ซ่อน sidebar ไม่มี tree
ให้กรอง พิมพ์ค้นหาแล้วไม่มีที่แสดงผลเลย:

- **เมื่อ query  active และไม่มี sidebar** (`topSearchResults !== null` คือ
  `menuLayout !== "left" || sidebarHidden` + needle ไม่ว่าง) → พื้นที่เนื้อหา
  สลับเป็น `[role="search"]` results grid ชั่วคราว **แทน TopMenuChrome + tabs**
  (คืนทุกอย่างเมื่อล้างการค้นหา)
- การ์ด = ชื่อเมนู + "section · group" ตัวจาง, กดเปิดงานเลย (openMenuItem) +
  **ล้าง query อัตโนมัติ**; header มี query + count badge + ปุ่ม "ล้างการค้นหา"
- empty state: SearchX + hint "ค้นหาได้ทั้งไทย/อังกฤษ/route ไม่สน case+วรรณยุกต์"
- โหมดซ้ายปกติไม่ทำอะไรเพิ่ม (tree กรองเองอยู่แล้ว — อย่าซ้ำซ้อนสองที่)
- กับดัก hooks: useMemo ที่คำนวณจาก `showLeftMenu` ซึ่งประกาศทีหลังบรรทัด
  นั้น → TDZ/TS error — ใช้ expression `menuLayout === "left" && !sidebarHidden`
  ตรง ๆ ใน memo + deps [globalSearch, menuLayout, sidebarHidden, ...]

## 4.17) หน้าสกุลเงิน: กันระบบไม่มีสกุลเงิน + ตั้งสกุลหลักจากหน้าเดียว (2026-08-30)

กฎผลิตภัณฑ์ "อย่างน้อยต้องมี 1 สกุลเงิน" — ปรับทั้ง Go + UI:

- **Backend guard** (`currency_http_service.go`): lastActiveCurrencyGuard ใน
  UpdateCurrency (case disable) / DeleteCurrency / DeleteCurrencyByGUIDs —
  นับ active ที่จะ "รอด" หลังกระทบ (active ที่ guid ไม่อยู่ในชุดที่ลบ) เหลือ 0
  = ปฏิเสธ ตอบ error ไทย (UI โชว์ข้อความนั้นตรง ๆ)
- **Frontend**: ปุ่มลบ disabled + title เหตุผลไทยเมื่อเป็นสกุลหลักหรือ active
  ตัวสุดท้าย; ปุ่ม "ตั้งเป็นสกุลเงินหลัก" บนการ์ด (PUT /organization/branch/:guid
  ด้วย branch doc เต็ม + basecurrency ใหม่ — base เดิมซ่อนในฟอร์มสาขา) —
  ตั้งแล้ว set baseCurrencyOverride ทันทีไม่ต้องรีโหลด
- **auto-seed**: เปิดหน้าแล้ว list ว่าง → POST THB + set base อัตโนมัติ (ref
  กัน seed ซ้ำ; reset ref ถ้า POST fail เพื่อ retry ได้)
- **quick-add**: ปุ่ม THB/USD/EUR/JPY/CNY/GBP/MYR/SGD — **ต้องหา preset ด้วย
  code อย่างเดียว** (findCurrencySymbolPreset ต้องการ name+symbol ตรง ไม่งั้น
  validate ตอนแก้ไขภายหลังจะปฏิเสธของที่ระบบสร้างเอง)
- **rate vs base**: ช่องอัตราในฟอร์มแก้ไข (ซ่อนกับสกุลหลัก) → append ลง
  exchangerates history (วันเดียวรวมเป็น entry เดียว) + โชว์ "1 USD = 35.5 THB"
  บนการ์ด
- **UAT กับดัก**: การ์ดหาด้วย xpath h2 exact + ancestor rounded [last()] (Card
  ใช้ rounded-lg; quick-add card มี code ค้างในปุ่มด้วย — .first() หลอก);
  disabled ปุ่มลบมี aria-label เป็นข้อความ guard ไม่ใช่ "ลบ"; ฟอร์ม**ไม่มี**
  toggle disable (status read-only ใน UI — disable ได้ทาง API เท่านั้น)
- **env traps**: mainapi recreate แล้ว dev-login 401 = secret ไม่ตรงกับ frontend
  (เทียบ hash ไม่พิมพ์ค่า) · node dev server หลงเหลือยึด :3000 ทำ HTML/chunks
  ไม่ตรง build — netstat เช็คแล้ว kill

## 4.18) เมนูหลัก: feedback เมื่อโหลดสิทธิ์ล้ม + chip จำนวนเซสชันออนไลน์ (2026-08-31)

เคสจริง: แท็บผู้ใช้ถือเซสชันตาย (หลัง backend restart + refresh rotation) →
ดึงสิทธิ์ /permissiongroup/me ล้ม → frontend เดิม catch เงียบ ๆ คืน Set() ว่าง →
เมนูล็อกหมด "ไม่มีสิทธิ์" โดยผู้ใช้ไม่รู้สาเหตุ (ผิดกฎ 40+ ข้อ 8)

- **แก้ที่ fetchAllowedMenuIds (main-menu-screen.tsx)**: คืน
  `{allowed, failure: "session-expired"|"http-error"|"network-error"|"no-permission-record", status}`
  แทน Set เปล่า — effect แสดง toast ไทยทุกกรณี:
  - 401 (หลัง authFetch ลอง refresh ครั้งเดียวแล้ว) = เซสชันตายจริง → toast
    "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" + clearAuthSession + ลบ workspace keys +
    router.replace("/") — อย่าล็อกเมนูเงียบ ๆ
  - network/http/no-record → toast warning + อยู่หน้าเดิม (fail-closed ยังใช้;
    refetch เกิดเองที่ window focus ซึ่ง effect มีอยู่แล้ว)
- **chip จำนวนเซสชัน** (ผู้ใช้ขอ "กี่เครื่องกำลังใช้ระบบ"): backend
  `GET /sessions/active-count` (authentication_http.go + AuthService.
  ActiveSessionStats ใน auth.go — SCAN Redis session-* ตัด session-revoked-*,
  active = lastseenat ≤ 30 นาที, แยก holdingcode) → proxy Next
  `/api/auth/sessions` → chip ชิดขวาใน breadcrumb "ผู้ใช้งานออนไลน์: N เซสชัน"
  + title รายละเอียดต่อ holding; โหลดตอน mount + ทุก focus; ล้มเงียบ (ข้อมูลประกอบ)
  - สร้าง slice Holdings จาก map ตอนท้าย (append *stat ตอนประกาศ = สำเนาค่าเก่า 0 หมด)
  - path ต้องลง exceptShopPath ของ main.go (จับแบบ exact match)

## 4.20) chip เซสชัน → หน้าต่างรายชื่อผู้ใช้ (distinct) (2026-08-31)

ผู้ใช้ขอต่อ: "กด chip แล้วแสดงรายชื่อ เวลาเข้าใช้ ใช้ล่าสุด" และ "distinct ด้วย"

- **backend**: CreateSession เก็บ `username`+`name` ลง session hash เพิ่ม
  (dev-login เดิมทิ้งว่าง → จอแสดง "ไม่ทราบชื่อ (เซสชันเก่า)"); ตัวเก่า resolve
  จาก bearer cache ผ่าน accesskey ถ้ายังไม่หมดอายุ
- **distinct ตามผู้ใช้**: entries รวมด้วย map key `username|name|holding` —
  Sessions=จำนวน, CreatedAt=เข้าใช้ล่าสุด, LastSeenAt=max, Active=any;
  session ไม่รู้ชื่อจับกลุ่ม "ไม่ทราบชื่อ" ต่อ holding (136 เซสชัน → 8 แถวจริง)
- **frontend**: chip เป็น `<button>` เปิด dialog (pattern dialog-backdrop +
  line-login-dialog ตาม password/LINE dialog) — แถว: จุดเขียว=active, ชื่อ,
  badge "N เซสชัน" (ถ้า >1), badge holding, "เข้าใช้ล่าสุด: HH:MM · ใช้งานล่าสุด:
  x นาทีที่แล้ว"; เรียงใช้ล่าสุดก่อน; ปุ่ม × และ ปิด
- เวลาแสดง: วันนี้ = เวลาอย่างเดียว (toLocaleString th-TH), วันอื่น = วัน+เวลา;
  relative = นาที/ชม./วัน ที่แล้ว — อ่านรวดเดียวรู้เรื่องตามกฎ 40+

## 4.19) Test infra: spec ใหม่ต้องประกาศ serial + execSync ผ่าน cmd.exe (2026-08-31)

- **ทุก spec ที่ใช้ storageState เดียวกันต้อง `test.describe.configure({mode:"serial"})`**
  — config หลักเป็น fullyParallel:true; ลืมประกาศแล้ว 3 test พุ่งพร้อมกันด้วย
  refresh cookie เดียวกัน → /refresh 3 ครั้งใน 2ms → reuse → revoke ทั้ง session
  (เห็นใน log mainapi) — อาการ: test แรก ๆ ผ่าน ตัวหลังโดนเตะไป login
- **รันหลาย suite พร้อมกันต้อง --workers=1 เสมอ** (บันทึกแล้วแต่ลืมได้ — serial
  ในไฟล์กันได้เฉพาะในไฟล์ ข้ามไฟล์ยังชนกันที่ cookie)
- **ทุก suite ที่เรียก API ด้วย session (ไม่ใช่ชุด login-page) ต้องมี
  afterEach cookie chain** (`storageState({path:'.auth/user.json'})`) —
  suite ที่ไม่มี (currency-uat/uat เดิม) กิน cookie ของ setup แล้วทิ้งของเก่า
  ไว้ ทำ suite ถัด ๆ ไปเด้ง login ทั้งชุด (battery 2026-08-31: 6× /refresh 401
  กระจุก) — ชุด login-page (test.use storageState: empty) ไม่ต้อง chain
- ชุดที่เปิด /menu ตรง ๆ (menu-session-feedback) ต้องมี openMenu self-heal
  (เจอ Dev Login = cookie ตาย → login ใหม่ → กลับ /menu) กัน cookie เก่าตั้งแต่ตัวแรก
- **execSync บน Windows = cmd.exe**: ห้าม pipe/while/$( ) ในคำสั่ง — ถ้าต้อง
  รวมข้อมูลฝั่ง Redis ให้เขียนเป็น Lua EVAL คำสั่งเดียว
  (`redis-cli EVAL "local ks=redis.call('keys','session-*') ..." 0` — ใน Lua ใช้
  single-quote เท่านั้น cmd ไม่ตีความ)
- **IAB จอ /workspace**: ปุ่มการ์ดบริษัท getByRole กับชื่อไทย timeout ทั้งที่มีจริง
  (DOM re-render ตลอด) — ใช้ `page.evaluate` หา `[...document.querySelectorAll('button')].find(x=>x.textContent.includes('TST03')).click()` แทน

## 4.21) รูปพนักงาน: อัปโหลด → S3 + thumbnail (กฎรูปภาพ 2026-08-31)

กฎใหม่ (AGENTS.md): รูปห้ามเก็บ Mongo — เก็บ S3/MinIO + ต้องมี thumbnail เสมอ

- **เปิดฟิลด์รูป**: config จอ (system-setting-screens.ts) เพิ่ม
  `imageUploadField("profilepicture", "รูปพนักงาน", "Employee photo", undefined, "profilepicturethumb")`
  — thumbnailKey = ชื่อ field thumb (บังคับตามกฎ) — editor เดิม
  (image-upload-editor.tsx) ย่อรูป + อัปโหลด 2 ไฟล์อัตโนมัติผ่าน /api/upload/image
- **backend**: models.Employee (ทั้ง internal/models และ shop/employee/models)
  เพิ่ม `ProfilePictureThumb json/bson:"profilepicturethumb"` — ไม่งั้น PUT
  รับค่าแล้วเงียบ ๆ ทิ้ง field thumb (พฤติกรรม binding unknown field = drop)
- **แสดงผล**: list + detail header ใช้ LogoAvatar (authenticated image → blob:
  src ปกติ) — ลำดับ uri: `profilepicturethumb → profilepicture → avatarthumb → avatar`
  (thumb ก่อนเสมอตามกฎ)
- **UAT ที่ต้องระวัง**: (1) input[type=file] ในจอพนักงานมีหลายตัว (CSV import!)
  — ต้อง scope ใน section ที่มี text "รูปพนักงาน" (2) PNG test ต้อง valid —
  PNG เสียโดน handler ปฏิเสธตอน decode สร้าง thumbnail ("zlib: invalid checksum")
  (3) waitForResponse ต้อง match `/api/upload/image` (Next proxy ที่จอยิง) ไม่ใช่
  `/goapi/image/upload` (server-side) (4) ตรวจ S3 ด้วย `docker exec minio ls
  /data/bcai-account/<key>` (5) execSync mongosh: ใช้ single-quote ใน JS query
  กัน cmd.exe กิน double-quote
- ผลจริง: Mongo = URI คู่ (ต้นฉบับ+thumb) ไม่มี binary · MinIO มี object จริง ·
  list/detail แสดงรูปผ่าน authenticated image
- **ตามมา (2026-08-31 เย็น)**: แทรกรูปไว้ fields[0] ทำ "รหัส:" ใน detail โชว์ URI —
  เพราะ recordDisplayCode → recordBusinessLookup อ่าน `config.fields[0].key` ก่อน
  record.code — แก้ด้วย `businessCodeField("code", ...)` (mark businessCode:true,
  uppercase บันทึก) ไม่ต้องย้ายลำดับ field · ผู้ใช้รายงาน "รูปไม่ขึ้นหลัง save"
  จริง ๆ = save ช่วง backend เก่า (ทิ้ง thumb) + หน้า cache — หลัง deploy
  backend ใหม่ + refresh แสดงครบ; MinIO มี .thumb.webp สร้างอัตโนมัติอีกชั้น
- **เรียง list (2026-08-31)**: main-crud สร้าง sort จาก `fields[0].key` (2 จุดใน
  system-settings-screen ~942/~1090) — แทรกรูปไว้ fields[0] ทำให้ list เรียงด้วย
  URI → save แล้วแถวเด้ง แก้แล้ว: เลือก `fields.find(f=>f.businessCode)?.key ??
  fields[0]?.key` (ตอนเพิ่ม field ใหม่ชนิดสื่อ ให้ mark businessCode ที่ code
  และเช็คจุดที่อ่าน fields[0] ทั้งหมด)
- **รูปเล็กแตกหลัง save (2026-08-31 ค่ำ)**: มี image cache สองระบบแยกกัน —
  authenticated-image.tsx (แก้แล้ว) และ **logo-avatar.tsx** ที่ list row ใช้ —
  LogoAvatar: cacheKeyFor ใส่ token + syncCacheOwner revoke ทั้ง cache ทันทีเมื่อ
  token หมุน → รูปแถวแตกเป็น alt text ทันทีหลัง save (token rotate จาก PUT/refetch)
  — แก้เหมือนเดิม: key ตัด token + revokeObjectUrlLater (delay 2 นาที) แทน revoke
  ทันที · บทเรียน: **ก่อนแตะรูป ต้องเช็ค image cache ให้ครบทุกไฟล์** (มี 2 ระบบ)
  และ session นี้ token หมุนตลอด (single-use refresh) อะไรที่ผูก token ใน key/lifetime
  จะพังเป็นจังหวะ
- **เปลี่ยน combobox → popup (2026-09-02, ลุงจืดขอ)**: Company/BranchScopeSearchPicker
  (holding-scope-editor.tsx — ใช้ทั้งจอพนักงาน/ผู้ใช้/สิทธิ์) เปลี่ยนจาก input+dropdown
  เป็น **ปุ่ม "เพิ่มบริษัท/เพิ่มสาขา" → popup dialog** (ScopePickerPopup — dialog-backdrop
  + line-login-dialog pattern): ช่องค้นหา + รายการแถวใหญ่ กดรายการ = เลือกทันที +
  ปุ่มปิด. local duplicates ใน system-settings-screen ถูกลบ (import จาก
  holding-scope-editor แทน — กันโค้ดซ้ำ 2 ชุดแก้ไม่ครบ) · UAT: scope-save ALL OK
- **เลือกแล้วเพิ่มเลย (2026-09-02)**: ตัดปุ่ม "เพิ่ม"/"เพิ่มสาขา" ข้าง picker ออก —
  คลิกรายการใน popup = เพิ่มเข้า list ทันที (Company/BranchScopeSearchPicker props
  เปลี่ยน: value/onChange → onPick; ผู้เรียก wire ตรงเข้า addCompanyScope/addBranchScope;
  ลบ pending state companyAddCode/branchAddCodes ทิ้ง) — flow เหลือ 2 จังหวะ:
  ปุ่ม → popup เลือก → รายการโชว์ใน list → กด บันทึก ของฟอร์มเพื่อ persist

## 4.22) Login premium round 2: full-bleed ชน radius + accent เดียวตาม palette (2026-09-02)

เจอ 2 กับดักจากบล็อก "Login premium restyle" ที่ append ท้ายไฟล์:

- **restyle block ที่ไม่ scope media query ทับ layout ที่อนุมัติแล้ว** — บล็อก
  premium ตั้ง `.login-shell .brand-panel { border-radius: 28px + shadow }` ทุก
  ความกว้าง แต่ ≥1024 แผงเป็น full-bleed (x=0 y=0 100dvh ตาม §4.8) → พื้นหลัง
  shell โผล่ตรงมุมจอโค้ง + เงาลอยไร้ประโยชน์ (วัด rect+computed ยืนยัน)
  **กฎ: ทุก restyle ที่แตะ radius/shadow/border ของ panel ต้องเช็คก่อนว่า
  breakpoint ไหน panel เป็น full-bleed — แล้ว scope หรือ override คืน 0 ใน
  media นั้นเสมอ**
- **accent ต้องตระกูลเดียวและไหลตาม palette** — CTA เป็น teal ตายตัว
  (#0d9488→#0f766e + glow teal), focus ring var(--teal), hover ปุ่มหัวการ์ด
  var(--indigo) ทั้งที่ eyebrow/icon/brand-mark ตาม --primary แล้ว → จอมี 3
  เฉด accent พร้อมกัน แก้เป็น derive จาก --primary หมด (gradient
  `color-mix(in srgb, var(--primary) 76%, #000)`, glow primary 55%, ring
  primary 18%, hover border/color primary) — ทุก palette + dark mode เปลี่ยน
  ตามอัตโนมัติ **กฎ: หน้าจอที่มี palette switcher ห้าม hardcode สี accent
  เป็น hex — ใช้ var(--primary)/color-mix เสมอ**
- audit ที่ต้องรันหลัง restyle หน้า login: rect+radius+shadow ของ .brand-panel
  ที่ 1280/1920/2560, overlap scan (มี false positive เดิม: label ใน iframe
  GIS `nsm7Bb-*` ทับ google-face โดยเจตนา — ไม่ใช่บั๊ก), ปุ่ม clip = 0,
  ความสูงแถวปุ่ม Google/Dev/CTA เท่ากัน ±1px

## 4.23) Premium skin ทั้งระบบ: token + block เดียวท้ายไฟล์ (2026-09-02)

กฎ AGENTS.md "UX/UI พรีเมี่ยม + ใช้ง่าย + เหมาะกับคนไทย" ถูก implement เป็น block
`/* ===== Global premium skin (2026-09-02) */` ท้าย `globals.css` (ประมาณบรรทัด 7900+):

- **radius ladder**: `:root { --radius-md: 8px; --radius-lg: 12px; --radius-xl: 16px;
  --radius-2xl: 20px }` — ทุก Tailwind `rounded-*` ใช้ `var(--radius-*)` จึงมนตามทันที
  (ยกเว้น `[class~="rounded-2xl"] { .9rem !important }` ~บรรทัด 6266 ที่ cap ไว้ให้ control
  40px ไม่กลายเป็น pill — **อย่าลบ**)
- **surface tokens** ใช้ซ้ำได้ทุกจอ: `--surface-hairline(-strong)`, `--surface-glass(-strong)`,
  `--surface-shadow(-hover)` — การ์ด/panel ใหม่ให้ใช้ชุดนี้ ไม่เขียน rgba ใหม่
- **เงา shadcn** (`--shadow-card`, `--shadow-popover`) ถูกเขียน **inline บน `<html>`** โดย
  `applyVisualTheme()`/`layout.tsx` จาก `src/lib/theme-data.ts` → override ใน stylesheet `:root`
  ไม่มีผล ต้องแก้สูตรใน theme-data.ts (ทำแล้ว: layered + ย้อม primary)
- ทุก accent = `color-mix(in srgb, var(--primary) N%, …)`; ปุ่ม `bg-primary` ทุกตัวได้ gradient
  จาก rule `[class~="bg-primary"] { background-image: linear-gradient(...) }`
- **ห้ามใส่ `overflow: hidden` บน panel ที่มี popover ลูก** (font/palette picker) — clip ที่
  `.login-shell` ชั้นนอกแทน
- **Tailwind bridge** (~บรรทัด 410-520, มี light+dark 2 ชุด) remap utility เก่า เช่น
  `.bg-muted\/30`, `.rounded-xl`, `.bg-card`, `.shadow-sm` — เมื่อ class Tailwind แสดงผลแปลก
  ให้ตรวจที่นี่ก่อน (วิธี: JS วน `document.styleSheets` แล้ว `el.matches(rule.selectorText)`)
- **ตรวจรับ**: light+dark (กดปุ่มสลับธีมจริง — ตัวแปรพาเลตอยู่ inline บน html การแก้
  `data-theme` เฉย ๆ ไม่ใช่การทดสอบ) × 1600/1280/1024/768-portrait + hover/focus/disabled/error
  + ไม่มี console error + เปิด popover ทุกตัวหลังแก้ CSS

## 4.24) ปุ่ม Sign in with Google: ห้ามซ่อน iframe ของ GIS (2026-09-02, prod bug)

pattern "หน้ากากปุ่มของเราทับ iframe GIS ที่ `opacity: 0`" ทำให้ **Chrome จริงเงียบ** —
GIS ใช้ IntersectionObserver v2 (clickjacking guard) ถ้า iframe ไม่ visible เต็มจะไม่เรียก
`window.open` เลย (browser ฝังตัว/เก่าที่ไม่มี IOv2 ยังทำงาน จึงหลอกตอนทดสอบ)
- ทำอย่างเดิม: แสดงปุ่ม GIS ตัวจริง (`theme: outline, size: large, text: continue_with`)
  ให้ wrapper div ของ GIS เต็ม host (`position:absolute; inset:0; width/height:100%`) —
  layout นี้ผ่านการวัดว่า clickable
- ขนาด: GIS ล็อก 40px สูง/กว้าง ≤400 → ขยายด้วย `container.style.zoom` (ไม่ใช่ transform
  ที่ทำให้ guard ล้ม) ให้สูงเท่าปุ่มข้างเคียง และ**หารความกว้างที่ขอจาก GIS ด้วย zoom** ก่อน
  (`login-screen.tsx renderGoogle`)
- **วิธีวัดใน Chrome จริง** (claude-in-chrome): stub `window.open` ให้บันทึก URL แล้วคลิกปุ่ม —
  ต้องได้ 2 URL (`o/oauth2/v2/auth`, `gsi/select`); ถ้าปุ่มเป็น personalized "Continue as …"
  จะใช้ FedCM ไม่เรียก window.open → ทดสอบใน localhost แทน
- §4.22 ข้อ "label ใน iframe GIS ทับ google-face โดยเจตนา" **ยกเลิก** — face ถูกลบแล้ว

## 4.25) Entrance animation ห้ามใช้ motion/react บน panel ทั้งจอ (2026-09-02, prod bug)

motion ขับ mount animation ด้วย requestAnimationFrame ซึ่งไม่ทำงานในแท็บที่ `document.hidden`
→ หน้าที่โหลดตอนอยู่แท็บพื้นหลังค้าง `opacity:0` ทั้งใบ (ฟอร์ม login หาย) ใช้ CSS
`@keyframes … animation-fill-mode: both` แทน (block `login-panel-enter`) และคง motion ไว้แค่
stagger ของลูก ผ่าน `initial={false} animate="animate"` ตรวจ: `document.querySelector(".form-panel")
.getAttribute("style")` ต้องเป็น null

## 4.26) Data list (`.bc-list-*`) และจอ master data (2026-09-02)

- zebra ใช้ `bg-muted/30` → bridge เคย map เป็น `--icon-soft-bg` (ชมพู #ffdad5) แก้เป็น
  `--muted-bg` 55%; `.rounded-xl/.rounded-2xl` เคยถูกบีบ 4px → ตอนนี้ตาม token
- `.bc-list-header` glass + เส้น primary บาง, แถวที่เลือก `box-shadow: inset 3px 0 0 var(--primary)`,
  hover primary 5% — ไอคอน "แก้ไข" ใช้ `text-primary` (ไม่ใช่ sky), "ลบ" คงแดง (semantic)
- ปุ่มใน list/detail ถูกซ่อนตามสิทธิ์ (§4.27)

## 4.27) สิทธิ์ต่อจอ เข้า/เพิ่ม/แก้ไข/ลบ + ตารางเลือกสิทธิ์ (2026-09-02)

- ค่าเก็บใน `role_permission.permissions`: `รหัสจอ` = เข้า, `รหัสจอ:create|update|delete`,
  `*` = ทั้งหมด (helper: `src/lib/permission-actions.ts`, backend validate ใน
  `rolepermission/models.NormalizeRequest`)
- จอ data: `useScreenActions(auth, workspace, route)` (`src/lib/use-screen-actions.ts`) map
  route → รหัสจอ ด้วย `flattenMenuItems()` แล้วซ่อน เพิ่ม/คัดลอก (create), แก้ไข (update),
  ลบ (delete) ใน `SettingDataList`/`SettingDetailPanel` (prop `actions`, default เปิดหมด =
  จอนอกเมนู/โหลดไม่ได้ไม่ล็อก; การเข้าจอคุมที่เมนูหลัก; backend enforcement = งานระยะ 2)
- **UX ตัวเลือกจำนวนมาก (100+ จอ) = ตาราง ไม่ใช่การ์ด**: `RoleScreenMatrix`
  (`components/system-settings/field-editors/role-screen-matrix.tsx`) — แถวจัดกลุ่มตามเมนู,
  คอลัมน์ต่อ action, ช่องค้นหา + กรอง เลือกแล้ว/ยังไม่เลือก, หัวตารางติ๊กกับ "จอที่แสดงอยู่",
  คลิกทั้งแถว = เข้า, ติ๊ก action → เข้าให้เอง, เอาเข้าออก → ล้างแถว, ช่องติ๊ก 20px แถว ≥44px
  ตัวหนังสือ ≥0.9rem (กฎ 40+) — ใช้แบบแผนนี้กับตัวเลือกชุดใหญ่อื่น ๆ (สาขา, สิทธิ์ผู้ใช้)
- **กับดัก**: input ซ้อนใน `<label>` ใหญ่ — คลิกข้อความในการ์ดจะ toggle checkbox ตัวแรก
  ของ label; แยกเป็น `<div>` + `<label>` ต่อ control เสมอ

## 4.28) Deploy ให้เห็นผล (2026-09-02)

งาน UI ที่ "เสร็จ" ต้องขึ้น https://account.bcaicloud.com ตามขั้นตอนใน memory
`account-bcaicloud-deploy` (build image ด้วย build-arg 2 ตัว, `docker save | ssh docker load`,
สลับ `release.env`, `compose up -d --no-deps frontend`, curl 200 + เปิดดูจริง) — deploy = R0
ต้องได้คำยืนยันจากลุงจืดก่อนทุกครั้ง

## 4.29) ตั้งค่าที่ "เป็นเรื่องบริษัท" อยู่ในขั้น 2 ของ ตั้งค่าระบบ (2026-09-02)

ลุงจืดตัดสินใจ: สกุลเงิน / ประเภทธุรกิจ / พนักงาน ย้ายจากเมนูหลัก ตั้งค่า › ตั้งค่าบริษัท ไปเป็น
**แท็บในขั้น "ข้อมูลบริษัทและสาขา"** (workspace-screen.tsx: `companyStepTabs` + state `companyTab`;
`SystemSettingsScreen route={companyTab}`) แม้ข้อมูลยังเป็นระดับ Holding (holdingcode) — ผู้ใช้มองว่า
"เรื่องบริษัท" ควรอยู่ที่เดียวกับการสร้างบริษัท
- เมนูหลัก: ลบ 3 item ออกจาก `MENU_SECTIONS` (menu-data.ts) และลบโฟลเดอร์ `company-settings`
  (main-menu-screen.tsx) → ตั้งค่า เหลือ ภาษาที่ใช้งาน + ตั้งค่าทั่วไป
- ผลข้างเคียงที่รู้: `useScreenActions` map route→รหัสจอผ่านเมนู — จอที่ไม่อยู่ในเมนูจะไม่ล็อกปุ่ม
  (ADMIN/OWNER เท่านั้นที่เข้า wizard ได้อยู่แล้ว); รหัสสิทธิ์ `currency/company-type/employee`
  ยังอยู่ใน permissiondefinition ของ backend
- MongoModel MCP: workflow `company_settings_hub` (project "BC Ai Account") บันทึกโครงนี้ไว้
  — เวลาย้าย/รวมจอ ให้บันทึก workflow ที่นั่นด้วยเสมอ (ลุงจืดใช้เป็น bcmodel)
- แบบแผน: "hub ขั้นตอน + แท็บย่อย" ใช้ปุ่ม tab `min-h-10 rounded-lg` active = bg-primary; key ของ
  SystemSettingsScreen ต้องรวม route ของแท็บเพื่อ remount ตอนสลับ
- **ต่อมาในวันเดียวกัน**: ลุงจืดสั่งถอด "ตั้งค่าทั่วไป" (LINE OA, ออกแบบฟอร์ม, แจ้งเตือน LINE, ผู้ให้บริการ AI,
  คัดลอกข้อมูลทดสอบ) และ "โครงสร้างข้อมูล (สมอง)" ออกจากเมนูหลักด้วย → หมวด ตั้งค่า เหลือ ภาษาที่ใช้งาน
  ตัวเดียว (`systemTreeFolders` ว่าง แต่คงโครงไว้) โค้ดจอยังอยู่ เปิดผ่าน route ตรงได้ ถ้าจะเอากลับให้เพิ่ม `tx()` ใน menu-data
