# UI Scale & Polish — Case Studies, Change History & Gotchas (Archive)

> ข้อมูลอ้างอิงประวัติการแก้บั๊ก บทเรียน และเคสเฉพาะกิจ (ข้อ 4.1 – 4.40) แยกออกจาก SKILL.md หลักเพื่อประหยัด Token


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
- **ภาษาที่ใช้งาน ก็ย้าย**เป็นแท็บแรกของขั้น "ข้อมูลบริษัทและสาขา" (ขั้น 1 เดิมถูกถอด, `companyTab` เริ่มที่
  `/activelanguages`, ใช้ `effectiveAccessRoute` แทน `activeAccessRoute` ทุกที่ที่เช็คหน้าภาษา) → เมนูหลัก
  **ไม่มีหมวด ตั้งค่า อีกแล้ว** (section ถูกลบจาก MENU_SECTIONS; test ยืนยัน) การตั้งค่าทั้งหมดอยู่ที่ ตั้งค่าระบบ ใน workspace
- **กับดัก**: `SystemSettingsScreen route=…` รับเฉพาะ slug ใน `system-setting-screens.ts` — จอที่ไม่ใช่ system-setting config
  (เช่น /currency = `CurrencyScreen`, /line-oa = `LineOaLinkScreen`) จะขึ้น "ไม่พบหน้าจอ" ต้อง branch render component
  ของมันเองเหมือนที่ `WorkTabPanel` ใน main-menu ทำ (ทำแล้วสำหรับแท็บ สกุลเงิน ใน workspace-screen)

## 4.30) ตั้งค่าระบบ 4 ขั้น: ธุรกิจ → คน → สิทธิ์ → ตรวจสอบ (2026-09-02)

โครง wizard ใหม่ (workspace-screen.tsx): `accessSettingNavItems` 4 ขั้นชื่อเป็นคำถามชาวบ้าน + `STEP_TABS` แท็บย่อยต่อขั้น
(`stepTabs` state, `effectiveAccessRoute` = แท็บที่เลือก) — ใช้ `effectiveAccessRoute` ทุกที่ที่เคยเช็ค `activeAccessRoute`
(manual, banner ภาษา, mount) · ปุ่ม "ถัดไป: <ขั้นถัดไป>" ท้ายทุกขั้น · ✔ ในเมนูซ้ายเมื่อขั้นมีข้อมูล
(`stepProgress` นับจาก list API: employee/user limit 1000, permissiongroup limit 1; บริษัทใช้ `flatCompanies` เพราะ slug company ตอบเป็นต้นไม้)
- **กับดัก API list**: `/api/system-settings/<slug>` ต้องส่ง `page=1&q=` และ header `x-bc-backend-url` (ไม่งั้น 400 "กรุณากรอก Backend URL")
- ขั้น 2 "คนในองค์กร": แท็บ พนักงาน/บัญชีเข้าระบบ + แถบสรุป จำนวนพนักงาน/บัญชี/คนที่มีทั้งสอง (จับคู่ email ↔ username normalize)
  — การรวมเป็น entity เดียวยังไม่ทำ (ต้องเปลี่ยน schema) นิยาม "พนักงาน" เคยเขียนไว้ใน `docs/organization.md` ซึ่งถูกลบไปพร้อม business-rule docs เมื่อ 2026-09-03 (ยังไม่มีเอกสารแทน — ต้องถามลุงจืด)
- ขั้น 3: ตารางสิทธิ์มีปุ่ม **ใช้ค่าแนะนำ** (`RoleScreenMatrix role=`; USER = เข้า+เพิ่ม+แก้ไข, ADMIN/OWNER = ทั้งหมด, กับจอที่แสดงอยู่)
  + แท็บ รายการจอทั้งหมด (permissiondefinition) แทนการเป็นขั้นแยก
- MongoModel: workflow `system_setup_steps` (rev 1355) · เมนูหลักไม่มีหมวด ตั้งค่า

## 4.31) อัปโหลดรูป local ล้ม "Failed to upload image to storage" ทั้งที่ env ถูก → ดู bootstrap.json (2026-09-02)
- **อาการ:** upload PNG (โลโก้สาขา ฯลฯ) ตอบ `Failed to upload image to storage`; `docker logs mainapi` มี `S3 … 403 InvalidAccessKeyId` ทั้ง GetObject/PutObject แม้ `docker exec mainapi env` และ `/proc/1/environ` แสดง `S3_ACCESS_KEY_ID` ที่ถูกต้อง และ `mc` ด้วย creds เดียวกัน (ผ่าน `--network container:mainapi`) ใช้ได้
- **สาเหตุ:** `backend/internal/goapi/setupconfig/loader.go` (`applyBootstrapSection` → `os.Setenv`) เอาค่าใน `bootstrap.json` (local = `bootstrap.local.json` mount เป็น `/app/bootstrap.json`) **ทับ env ของ compose ตอน start** — `s3accesskeyid: minioadmin` + `s3bucketname: app-images` ไม่มีใน MinIO local → client singleton (`GetR2Client`) ถูก cache ด้วย creds ผิดตลอดอายุ process. env ที่เห็นจาก `docker exec`/`/proc/*/environ` คือค่าตอน spawn จึงหลอกตา
- **วิธีตรวจให้ไว:** `docker logs mainapi | grep "Object storage client initialized"` → ถ้า `bucket:` ไม่ตรง `S3_BUCKET_NAME` ใน `storage.local.env` แปลว่า bootstrap ทับ; ดู `[Bootstrap] override S3_*` บรรทัดก่อนหน้า
- **แก้:** ให้ `s3accesskeyid/s3secretaccesskey/s3bucketname` ใน `bootstrap.local.json` ตรงกับ `storage.local.env`/`minio.local.env` แล้ว `docker restart mainapi`; verify ด้วย POST `/api/upload/image` (category `branch`) จากหน้าเว็บที่ login แล้ว → 200 + `mc ls t/bcai-account/bc001/branch/` เห็น `.png` + `.png.thumb.webp` + GET `/goapi/s3/file/<key>.thumb.webp` = 200 image/webp
- **กติกา:** ไฟล์ทั้งคู่เป็น local-only (gitignored) — เวลา rotate key MinIO ต้องแก้ 3 ที่: `minio.local.env`, `storage.local.env`, `bootstrap.local.json` (ต่อยอด §4.21)

## 4.32) ชุดสิทธิ์ = โปรไฟล์สำเร็จรูป เลือกได้หลายชุดต่อคน (2026-09-02)
- **แนวคิด (ลุงจืด):** ชุดสิทธิ์มีไว้เพื่อความสะดวก — ตั้งชุดตามแผนก (บัญชี/ขาย/คลัง) แล้วคนในองค์กรเลือกได้หลายชุด สิทธิ์จริง = union ทุกชุด + ชุดมาตรฐานของระดับสิทธิ์ (USER/ADMIN/OWNER)
- **ขั้น 3 ชุดสิทธิ์ (`permissiongroup`):** `rolecode` เป็น `businessCodeField` (พิมพ์อะไรก็ได้ → proxy ทำ uppercase+ตัดช่องว่าง; backend regex `^[A-Z0-9_-]{2,30}$` กันเรียกตรง) · ตาราง `RoleScreenMatrix` เดิม · ปุ่ม "ใช้ค่าแนะนำ" โผล่เฉพาะ USER/ADMIN/OWNER
- **ขั้น 2 ผู้ใช้งาน (`user`):** ฟิลด์ `permissionsets` (jsonField) เรนเดอร์ด้วย `PermissionSetsEditor` (`field-editors/permission-sets-editor.tsx`): การ์ด checkbox หลายคอลัมน์ + ค้นหาเมื่อ >6 ชุด + badge จำนวน · ซ่อน USER/ADMIN/OWNER ออกจากตัวเลือก (มาจาก radio ระดับสิทธิ์อยู่แล้ว) · empty state ชี้ไปขั้น 3
- **กับดัก:** จอ user มี renderer 2 ชุดที่ hard-code `sections[].keys` (`system-settings-screen.tsx` ~4780 และ `setting-form-dialog.tsx` ~249) **และ** `FieldEditor` ซ้ำ 2 ที่ (`components/system-settings/field-editor.tsx` + local ใน `system-settings-screen.tsx` ~8598) — เพิ่มฟิลด์ใหม่ต้องแตะทั้ง 4 จุด ไม่งั้นได้ textarea JSON ดิบ
- **Backend:** `shopusers.permissionsets []string` (normalize uppercase/dedup ตอน save; copy ลง profile ทั้ง list+info) · `/organization/role-permission/me` คืน `permissions` ที่ union แล้ว + `permissionsets`; ADMIN/OWNER ที่ไม่มี record ของ role ตัวเองยังได้ `*` เสมอ (บั๊กที่เจอตอน verify: union แล้ว owner หลุด `*`) · unit test `union_test.go` · bcmodel diagram "การ Login" (shopusers.permissionsets, role_permission.rolecode)
- **Verify:** สร้างชุด `sales` ผ่าน UI → list แสดง `SALES` · ผู้ใช้ติ๊ก/ปลดชุดในการ์ดได้ · API `/me` owner = `["*"]` + sets · `permtest`: username ว่าง (Google-only owner) กดบันทึกจอ user ไม่ผ่าน "กรุณากรอกช่องที่จำเป็น" = ปัญหาเดิมนอกขอบเขต (flag แล้ว)

## 4.33) ลำดับตั้งค่าระบบ: ธุรกิจ → สิทธิ์การใช้งาน → คน → ตรวจสอบ (2026-09-02)
- **ทำไมสลับ (ลุงจืดยืนยัน + Kimi K3 เห็นตรงกัน):** wizard ต้องเดินหน้าอย่างเดียว — จอบัญชีเข้าระบบเลือกชุดสิทธิ์ ถ้าชุดอยู่ขั้นหลังผู้ใช้ต้องข้ามไปสร้างแล้วย้อน · SME ไทยคิด "กำหนดหน้าที่ก่อน แล้วจับคนลง" · ขั้นสิทธิ์ข้ามได้ (ชุดมาตรฐานมีเสมอ)
- **คำที่ใช้ทั้งระบบ:** "สิทธิ์การใช้งาน" (ไม่ใช้ "ชุดสิทธิ์"/"Permission Set" ในจอ) — แท็บขั้น 2 = "ชุดสิทธิ์การใช้งาน", ช่องในจอบัญชี = "สิทธิ์การใช้งานเพิ่มเติม"
- **จุดแก้เมื่อสลับขั้น:** `workspace-screen.tsx` `STEP_TABS` + `accessSettingNavItems` (ลำดับ object = ลำดับขั้น, `stepProgress` ใช้ route key จึงไม่ต้องแก้) · ข้อความ "ขั้น N" ที่ hard-code ใน `permission-sets-editor.tsx`, `system-settings-screen.tsx` ~4784, `setting-form-dialog.tsx` ~252 · bcmodel workflow `system_setup_steps` (rev 1358)
- **Verify:** เปิด /workspace → ตั้งค่าระบบ → sidebar 1 ธุรกิจของฉัน / 2 สิทธิ์การใช้งาน / 3 คนในองค์กร / 4 ตรวจสอบ, ปุ่ม "ถัดไป: สิทธิ์การใช้งาน" · tsc + vitest ผ่าน
- **ระยะถัดไป (Kimi เสนอ):** ปุ่ม "สร้างชุดใหม่" แบบ modal ในจอบัญชีเข้าระบบ เป็น safety net สำหรับคนที่ข้ามขั้น 2

## 4.34) ปุ่ม "ทดลองใช้ระบบ (Demo)" แทน Dev Login + ข้อมูลตัวอย่าง SME ไทย (2026-09-02)
- **สิ่งที่ทำ:** login-screen ปุ่ม Dev (loopback-only) → ปุ่ม Demo แสดงทุก environment เรียก `/api/auth/demo-login` → backend `POST /demo-login` (public, whitelist ใน `main.go` + `cmd/*/main.go` exceptShopPath, blocked ที่ `/backend/demo-login` ใน next.config เหมือน dev-login) · เปิดด้วย `BCAI_DEMO_LOGIN_ENABLED=true` (+`BCAI_DEMO_USERNAME`, default `demo`) — local: `docker-compose.yml` default true, prod: `provision-server.sh` backend.env
- **Backend:** `internal/demo` (env helper) · `DemoLoginByUsername` สร้าง user `demo` อัตโนมัติครั้งแรก (RegisterByUsername, password สุ่ม) audit `DEMO_LOGIN` · `creator_access.go` ยกเว้นเงื่อนไข Google Identity ให้ demo user ตอนสร้างองค์กร (เฉพาะเมื่อ demo เปิด)
- **Seed:** `node scripts/seed-demo.mjs` (`SEED_BASE=https://…` สำหรับ server) อ่าน `scripts/demo-data.json` (Kimi K3 ร่าง, Claude verify EAN-13 checksum + dedup ข้ามบริษัท) — 1 holding `demo` / 3 บริษัท (วัสดุก่อสร้าง 2 สาขา, คาเฟ่, หจก.ซื้อมาขายไป) / 6 แผนก / 4 ชุดสิทธิ์ / 12 พนักงาน / 15 สินค้า+8 ลูกค้า+6 ผู้ขาย ต่อบริษัท · idempotent (dup = skip)
- **กับดักที่เจอ:** (1) `/backend/:path*` rewrite ชี้ mainapi root — ห้ามใช้ `/backend/goapi/...` สำหรับ auth routes (2) create-holding/company/branch ตอบแค่ success ต้อง GET list กลับมาเอา uid (3) เจ้าของต้องมี `accessscopes` ครอบทุกบริษัท ไม่งั้น `select-holding` รายบริษัท 403 — PUT `/holding/permission` ต้อง key ด้วย `editusername=useruid` + `username:""` ไม่งั้น upsert เป็น membership ซ้ำ (4) ลบข้อมูล holding ทิ้งต้องลบ `organizationcodeclaims.normalizedcode` ด้วย ไม่งั้น create-holding เงียบแต่ไม่สร้าง (5) barcode `prices[].price` เป็น number, branch `addresses[].address` เป็น string
- **Verify:** local กดปุ่ม → หน้าเลือกกลุ่มกิจการ "กลุ่มกิจการรุ่งเรือง (สาธิต) 3 บริษัท 4 สาขา" → เลือกบริษัทได้ทั้ง 3 · Go tests organization/services ผ่าน (builder stage)

## 4.35) หน้าแรก (ภาพรวม) = งานของฉันวันนี้ ตามสิทธิ์ (2026-09-02)
- **แนวคิด (ลุงจืดยืนยัน):** ไม่ทำ dashboard แยกต่อบทบาท — widget ผูกกับรหัสจอ แสดงเฉพาะที่ `allowedMenuIds` (union ชุดสิทธิ์จาก `/permissiongroup/me`) มี; ไม่มีสิทธิ์ = ไม่แสดง (ไม่ใช่เลข 0)
- **ไฟล์:** `src/app/menu/dashboard-home.tsx` (แทน `DashboardHome` เปล่าใน main-menu-screen) props: auth, workspace, allowedMenuIds, allMenuItems, frequentMenuEntries, onOpenItem, mainApiUrl (`deriveMainApiUrl(auth.backendUrl)` → `/backend` root)
- **3 แถว:** เอกสารที่ดูแล (การ์ดตัวเลข `total` จาก `/transaction/<doc>/list?limit=5` ต่อประเภท, กดเปิดแท็บจอ) · ทางลัด (`getFrequentMenuEntries` ≥4 อัน ไม่งั้นเติมจอที่มีสิทธิ์) · ความเคลื่อนไหวล่าสุด (รวม 5 รายการล่าสุดทุกประเภท เรียง `docdatetime`, แสดง docno/คู่ค้า `custnames|creditornames`/`totalamount`)
- **กับดัก:** menu id ≠ ชื่อ route: ขายสินค้า = `sale` (ไม่ใช่ sale-invoice), ปรับปรุงสต็อก = `stock-adjust`; list endpoint คืน `total` ระดับบน ไม่ใช่ใน pagination; ยังไม่มี filter สถานะ/ช่วงวัน (param "-" = docdatetime range ยังไม่ได้ใช้)
- **ยังไม่ทำ (เฟส 2):** แถวตัวเลขเงิน (ยอดขาย/ซื้อ/กำไร — ไม่มี summary endpoint), รออนุมัติ (`/api/approval/*-status/pending` เป็น POST payload ยังไม่ได้ไล่), สินค้าใกล้หมด, ประกาศจากเจ้าของ
- **Verify:** demo owner เห็นการ์ด 8 ประเภท (0 เอกสาร) + ทางลัด 8 + empty state; tsc/eslint ผ่าน

## 4.36) กับดัก: sidebar โชว์แต่ตัวเลข (ไม่มีไอคอน/label) หลัง restart dev server = tab ค้าง ไม่ใช่ CSS bug (2026-09-04)

ลุงจืดรายงานสกรีนช็อตจอ `/product`: sidebar ซ้ายเหลือแต่ pill ตัวเลขลอย ๆ (4,3,3,10,2,4,3,4,3,3,2,6,3,3,2,3 —
ตรงกับ count ของกลุ่มย่อยใน "ข้อมูลหลัก" เป๊ะ) ไม่มีไอคอน/label/chevron เลย ข้างขวาเนื้อหา `/product` ก็ไม่ครบ —
พร้อมสั่ง "ต้องมี base line แก้ skill ด้วย"

**สืบสวนแล้ว ไม่ใช่ CSS bug จริง — reproduce ไม่ได้เลย:**
- ไล่ทดสอบจอ `/product` ทุก breakpoint ที่เกี่ยวข้อง (`showLeftMenu` ที่ `lg:` 1024 ใน main-menu-screen.tsx,
  `product-workbench` ที่ `xl:` 1280 ใน product-screen.tsx): 360/480/620/768/850/1024/1280 px ทุกจอ label+icon+badge
  ครบเสมอ (ต่ำกว่า 1024 = sidebar เต็มจอแบบ mobile overlay อ่านง่าย, ≥1024 = 280px column โชว์ครบ)
- ลองบวก app zoom-control (`~/zoom-control.tsx`, ค่าสูงสุด 150%) ที่ 1024px ด้วย — label ห่อบรรทัดสอง
  แต่ไม่หายไปไหน ไม่ตรงกับที่รายงาน (badge ล้วน ไม่มี label เลย)
- โครง JSX ของแถวกลุ่ม (`MenuSectionAccordion`/`MenuTreeGroup`) ผูก badge+chevron ด้วย `shrink-0` เสมอ
  คู่กับ icon+label ที่ `min-w-0 truncate` — ไม่มี path ไหนที่ badge โผล่ได้โดยไม่มี chevron ติดมาด้วย
  (ในสกรีนช็อตที่รายงานไม่มี chevron เลยสักตัว) → โครงสร้าง DOM ปัจจุบันสร้างอาการนี้ไม่ได้

**ข้อสงสัยหลัก — ตรงกับที่เพิ่งเจอเรื่อง 403/HMR ในเซสชันเดียวกัน**: ก่อนหน้านี้ผมเพิ่ง `taskkill` +
restart `next dev` (เพื่อให้ `allowedDevOrigins` มีผล ดู memory เครื่อง) — ถ้า tab ของลุงจืดเปิดค้างอยู่ตอนนั้น
WebSocket HMR หลุดกลางคัน แล้ว React patch เพี้ยนจนจอค้างสภาพผสม (บาง component re-render บาง component ไม่)
ก็เป็นไปได้สูงว่าเป็นสาเหตุ ไม่ใช่ CSS ผิดจริง — สังเกตแบบเดียวกับที่ลุงจืด paste console 403 ซ้ำหลัง restart
ไปแล้วในเทิร์นก่อนหน้า (ของเก่าที่ค้างอยู่ใน tab ไม่ใช่ของสด)

**วิธีตรวจก่อนเชื่อว่าเป็น CSS bug ครั้งหน้า**: กด hard refresh (Ctrl+Shift+R) ที่ tab ที่รายงานปัญหาก่อนเสมอ —
ถ้าหายหลัง refresh = tab ค้าง ไม่ใช่โค้ด; ถ้ายังเป็นอยู่ ให้ขอ **ความกว้างหน้าต่างจริง + browser zoom% จริง**
(ไม่ใช่ resize จำลอง) จากลุงจืดมาก่อน แล้ว reproduce ด้วยค่านั้นเป๊ะ ๆ ก่อนแก้ CSS — ห้ามแก้ CSS แบบเดา
เมื่อ reproduce ไม่ได้ (Rule 2/7 ห้าม fabricate ว่า "แก้แล้ว" ทั้งที่ไม่เห็นบั๊กจริง)

## 4.37) เมนู "สินค้า" (งานบัญชี) vs "สินค้า (ส่วนขยาย)" — จอเดียว 2 โหมด ไม่ก็อปคอมโพเนนต์ (2026-09-04)

**แบบแผน**: ลุงจืดสั่งแยกฟอร์มสินค้าให้งานบัญชีง่ายขึ้น — เมนู **สินค้า** เหลือ 3 tab
(ข้อมูลหลัก / หน่วยนับและบาร์โค้ด / คงคลัง) ส่วนอีก 8 tab (หมวดหมู่, ส่วนประกอบ BOM, ภาพและสี,
การจัดส่ง/โลจิสติกส์, ร้านอาหาร/POS, เวลาขาย, สาขา/ธุรกิจ, อื่นๆ) ย้ายไปเมนูใหม่ **สินค้า (ส่วนขยาย)**
ที่วางต่อจาก "สินค้า" ใน `products` group. ทำด้วย prop `mode: "core" | "extension"` บน `ProductScreen`
(`frontend/src/app/menu/product-screen.tsx` — `CORE_PRODUCT_TABS` / `EXTENSION_PRODUCT_TABS` ใกล้บรรทัด 135)
ไม่ก็อปจอ 2,600 บรรทัดออกเป็นไฟล์ใหม่ (Rule 15) — list ซ้าย + panel ขวา + save handler ใช้ของเดิมทั้งหมด.

**เหตุผลที่ส่วนขยาย = edit-only**: ซ่อน "เพิ่ม" / "คัดลอก" (และ "บันทึกและเพิ่มใหม่" หายเองเพราะ
`editorMode` ไม่เคยเป็น create) — สร้างสินค้าจากจอที่ไม่มี tab ข้อมูลหลักจะได้สินค้าไร้ชื่อ/หน่วยนับ.
default tab ของโหมดส่วนขยาย = `classification` (จอเปิดแล้วต้องมี tab active เสมอ).

**สิ่งที่ต้องทำทุกครั้งเมื่อเพิ่ม menu item ใหม่** (เจอจาก test แดง 2 ตัวใน `menu-data.test.ts`):
1. `menu-data.ts` เพิ่ม `tx(id, th, en, route, category)`; `menu-icons.ts` เพิ่ม route → icon
2. `main-menu-screen.tsx` เพิ่ม route ใน switch ของ `WorkTabPanel` **และ** ในสองรายการ
   `activeTabNeedsFixedViewport` (บรรทัด ~603 และ ~1229) ถ้าเป็นจอ list+detail แบบ fixed viewport
3. `backend/assets/language/languages.tsv` **ต้องมีแถว key = id (ขีดกลาง→ขีดล่าง) ครบ 12 คอลัมน์ภาษา**
   ไม่งั้น test "has backend language keys" + "has all supported language cells" ล้ม — ไฟล์เป็น CRLF ล้วน
   แก้ด้วย Python byte-preserving; ภาษาที่แปลไม่ได้ให้ใส่ English ตาม precedent แถว `productset`
   (อย่าแต่งคำแปลเขมร/พม่า/ลาวเอง — Rule 1/7). UI ไม่พังระหว่างรอ backend รับ TSV ใหม่ เพราะ `menuText`
   fallback ไป `label.th` เมื่อ dictionary พร้อมแต่ไม่มี key
4. อัปเดต `menu-data.test.ts` รายการ id ของ group นั้น
5. **สิทธิ์**: `canAccessMenuItem = allowedMenuIds.has(item.id)` — role ที่ได้ `"*"` (OWNER/ADMIN) เห็นเมนูใหม่ทันที
   แต่ role ที่มีรายการสิทธิ์ชัดเจนจะเห็นเมนูใหม่เป็น **ล็อก** จนกว่า admin จะติ๊กเพิ่มในกลุ่มสิทธิ์ — ไม่ใช่ bug
   ต้องบอกลุงจืดเวลาเพิ่มเมนู

**กับดัก HMR ตอนแก้**: แก้ค่าคงที่ (ลบ `PRIMARY_PRODUCT_TABS`) กับ JSX ที่อ้างถึงคนละ save → Turbopack
compile กลางทางแล้วโยน `ReferenceError ... is not defined` ค้างใน console (read_console_messages สะสมของเก่า
ไม่ล้างตอน reload) — ตรวจว่า chunk ที่เสิร์ฟจริงยังมี identifier เก่าไหมด้วย `fetch(script.src)` แล้ว `includes()`
ก่อนจะไล่หา bug ในซอร์สที่ไม่มี.

**วิธีตรวจรับ**: เปิดทั้ง 2 เมนูใน work tab → กด "แก้ไข" → อ่านปุ่ม tab ที่ **มองเห็นจริง**
(`offsetParent !== null` เพราะ WorkTabPanel ที่ไม่ active ยังอยู่ใน DOM) ต้องได้ 3 / 8 tab ตามสเปก;
บันทึกจากจอส่วนขยายแบบไม่แก้อะไร → `PUT /api/product/<guid>` 200 แล้ว query `appdb.products` ยืนยัน
doc เดิม (guid เดียว, names/itemtype/vattype เท่าเดิม, `updatedat` เลื่อน) — พิสูจน์ว่า save จากจอที่ไม่ render
tab ข้อมูลหลักไม่ทำให้ field หลักหาย. ไม่มี CSS ใหม่ในงานนี้ (ใช้ class เดิมทั้งหมด) จึงไม่ได้รัน light/dark sweep ซ้ำ.

## 4.38) ปรับปรุง Sidebar Tree Menu & Polish เมนูสินค้าและบาร์โค้ด (2026-09-04)

**ปัญหาที่พบและแก้ไข**:
1. **ไอคอนซ้ำกัน**: เมนู "สินค้า" (`/product`) และ "สินค้าชุด" (`/productset`) เดิมใช้ไอคอน `grid` ซ้ำกัน ทำให้กวาดสายตาแยกไม่ออก — ปรับเป็น `boxes` (กล่องซ้อนกัน) ใน `frontend/src/lib/menu-icons.ts` และ `frontend/src/app/menu/menu-icon.tsx` เพื่อสื่อความเป็น Bundle / Set ชัดเจน
2. **คำศัพท์ภาษาไทยสำหรับคน 40+**: เมนู "สินค้า (ส่วนขยาย)" ฟังเป็นศัพท์คอมพิวเตอร์/เทคนิค — ปรับป้ายเมนูเป็น **"ข้อมูลเสริมสินค้า"** (`Product Extended Data`) ใน `frontend/src/lib/menu-data.ts` และ `frontend/src/lib/product-barcode/language.ts`
3. **ปุ่ม `+` ในเมนูชวนเข้าใจผิด**: เดิมปุ่มทางขวาสุดของแต่ละรายการเมนูใช้ไอคอน `Plus` สำหรับการ "เปิดแท็บใหม่" (`onOpenNewItem` / `newTabLabel`) ส่งผลให้คนทำบัญชี/ร้านค้า 40+ เข้าใจผิดคิดว่าเป็นปุ่ม "สร้างรายการใหม่ด่วน" และยิ่งสับสนเมื่อกดที่เมนูส่วนขยาย — ปรับเป็น `ExternalLink` ทั้งใน `MenuTreeItemButton` (sidebar) และ `topMenuItemRow` (top menu popover)
4. **ยกระดับ Visual Hierarchy & Soft Tree**:
   - เพิ่ม container ล้อมไอคอนเมนู (`bg-muted/60 text-muted-foreground group-hover:bg-primary/20 group-hover:text-primary`) ให้มีมิติ ช่วยให้สมองกวาดสายตาจำแนกหมวดหมู่ได้เร็ว
   - ปรับเส้น Tree Guide จากขีดแข็งกระด้างให้เป็นเส้นนุ่มนวล (`border-border/50`) และปรับ Badge นับจำนวนเป็น pill นุ่มมนตา
   - แถวเมนูและปุ่มรองรับ hover ย้อมสีจาก `--primary` ตามกฎ "หนึ่งจอ หนึ่งตระกูลสี"

**วิธีตรวจรับ**:
- รัน unit test: `npx vitest run src/lib/menu-icons.test.ts src/lib/menu-data.test.ts` (ยืนยันไอคอน route ครบ 119 เมนู)
- รัน typecheck: `npm run typecheck`
- ตรวจสอบผ่าน browser ทั้ง Light Mode และ Dark Mode จริง (ยืนยัน contrast, text readability, และ layout ไม่เพี้ยน)

## 4.39) ปรับปรุงชุดไอคอนทั้งระบบให้ตรงบริบท (Icon Overhaul) (2026-09-04)

**ปัญหาที่พบและแก้ไข**:
1. **Section Icons ระดับหมวดหลัก**:
   - `master` (ข้อมูลหลัก): เดิมใช้ `<Command>` (สัญลักษณ์ ⌘ บนคีย์บอร์ด Mac) ซึ่งคนไทย 40+ ไม่เข้าใจ — เปลี่ยนเป็น `<Database>` (ฐานข้อมูลหลักของระบบ)
   - `reports` (รายงาน): เดิมใช้ `<LayoutDashboard>` ซ้ำกับ "ภาพรวม" — เปลี่ยนเป็น `<BarChart3>` (กราฟแท่งรายงาน)
   - `transactions` (งานประจำ): เดิมใช้ `<Store>` (ร้านค้า) ซึ่งแคบเกินไปสำหรับงานเอกสารประจำวัน — เปลี่ยนเป็น `<BriefcaseBusiness>` (กระเป๋าทำงาน/ธุรกรรมประจำ)
2. **Group Accordion Headers (หัวข้อกลุ่มเมนูทุกกลุ่ม)**:
   - เดิม: Accordion ทุกกลุ่มไม่มีไอคอนเลย มีเพียงลูกศร Chevron และข้อความเปลือยๆ
   - แก้ไข: เพิ่ม `MenuGroupIcon` คลุมด้วย container มนนุ่มย้อม primary (`bg-muted/60 text-muted-foreground group-hover:bg-primary/20 group-hover:text-primary`) จับคู่ไอคอนตรงความหมาย เช่น `products` → `Package`, `product-sku-options` → `Palette`, `partners` → `UsersRound`, `sales-payment-banking` → `Landmark`, `sales-pos` → `MonitorCog`, `approval` → `CheckCheck`, `restaurant` → `ChefHat` ฯลฯ
3. **Menu Route Item Icons**:
   - `/product`: ปรับจาก `"grid"` (ตารางสี่ช่อง) เป็น `"package"` (กล่องสินค้าเดี่ยว) เพื่อให้สอดคล้องคู่กับ `/productset` ที่เป็น `"boxes"` (กล่องซ้อน)
   - `/masterbrandscreen`: ปรับจาก `"settings"` เป็น `"badge"` (ตราสัญลักษณ์/ยี่ห้อ)
   - `/mastermodelscreen`: ปรับจาก `"activity"` เป็น `"design"` (รุ่น/โมเดลสินค้า)
   - `/masterpatternscreen`: ปรับจาก `"grid"` เป็น `"table"`
   - `/knowledgebasescreen`: ปรับจาก `"upload"` เป็น `"archive"` (คลังความรู้)
   - `/cashinginthedrawer`: ปรับจาก `"check"` เป็น `"money"` (รับ-ส่งเงินสด)
   - `/checkdaily/dailyinfoscreen`: ปรับจาก `"check"` เป็น `"calendarCheck"` (ตรวจสอบประจำวัน)
   - `/transaction/rfq`: ปรับจาก `"activity"` เป็น `"fileText"` (เอกสารสืบราคา)
   - `/transaction/stocktransfer`: ปรับจาก `"activity"` เป็น `"truck"` (ขนย้าย/โอนสินค้า)
   - `/zonegroupselectscreen`: ปรับจาก `"grid"` เป็น `"map"` (โซน/ผังพื้นที่)
   - `/rebuildstockscreen`: ปรับจาก `"settings"` เป็น `"cloudSync"`
   - `/datamodelgraph`: ปรับจาก `"table"` เป็น `"network"` (แผนภาพกราฟข้อมูล)

**วิธีตรวจรับ**:
- Unit test: `npx vitest run src/lib/menu-icons.test.ts` ผ่าน 100%
- Typecheck: `npm run typecheck` ผ่าน 0 errors
- Browser inspection ทั้ง Light Mode และ Dark Mode ยืนยันไอคอนคมชัด ตรงความหมาย และสม่ำเสมอทุกจุด



## 4.40) ปรับปรุงหน้าเลือกกลุ่มกิจการ, หน้าเลือกบริษัท, และหน้าตั้งค่าระบบและการเข้าถึง (2026-09-04)

**ปัญหาที่พบ**:
1. หน้า /holding และ /workspace มีการ hardcode สีส้ม (\mber-500\, \order-amber-500\, \	ext-amber-600\) สำหรับริบบิ้น, ไอคอน Avatar, และ Badge "เจ้าของ/OWNER" ซึ่งขัดกับกฎ "หนึ่งจอ หนึ่งตระกูลสี" เมื่อผู้ใช้เลือกธีมสีอื่น
2. Eyebrow ใน Card Header ถูกคลาสเดิมใน globals.css (\.card-header .eyebrow { color: var(--teal); }\) บังคับเป็นสีเขียวหลุดธีม
3. หน้าต่าง Wizard "ตั้งค่าระบบและการเข้าถึง" (\SystemSettingsScreen\) ขาด \<AppHeaderControls>\ ใน Header ทำให้ผู้ใช้ไม่สามารถปรับ Zoom, เปลี่ยนพาเลตสี, หรือสลับโหมดมืด/สว่างได้
4. Sidebar Rail ในหน้าตั้งค่ามีขนาดตัวเลขขั้นตอนเล็กเกินไป (\size-5 text-[10px]\) และข้อความขั้นตอนที่ไม่ได้เลือกดูจมซีดเกินไป ไม่เหมาะกับผู้ใช้อายุ 40+
5. Subtabs และ Master-detail layout ในแท็บย่อย (ภาษา, บริษัท/สาขา, สกุลเงิน, ประเภทธุรกิจ, สิทธิ์, พนักงาน, บัญชี, ตรวจสอบ) ยังขาดมิติและการเชื่อมต่อสีกับธีมหลัก

**การแก้ไข**:
1. แปลง hardcoded amber ทั้งหมดใน \holding-screen.tsx\ และ \workspace-screen.tsx\ ให้ derive จาก \--primary\ (เช่น \g-primary/15 text-primary border-primary/30\)
2. Override \.login-shell .login-card .card-header .eyebrow\ ใน \globals.css\ ให้ผูกกับ \ar(--primary)\
3. เพิ่ม \<AppHeaderControls>\ ใน Header ของ Wizard ตั้งค่าระบบ ให้ผู้ใช้ปรับแต่งการแสดงผลได้ครบถ้วน
4. ขยายวงกลมขั้นตอนใน Sidebar Rail เป็น \size-7 text-xs font-black\ พร้อมปรับคอนทราสต์ตัวหนังสือให้ชัดเจนอ่านง่าย
5. ปรับ Subtabs ให้เป็น interactive pill buttons (\min-h-[2.6em] rounded-xl font-bold\) และเติม ambient depth ให้พื้นหลังเชื่อมโยงกับธีมหลักในทั้ง Light Mode และ Dark Mode

**วิธีตรวจรับ**:
- Typecheck: \
pm run typecheck\ ผ่าน 0 errors
- Browser inspection และบันทึกภาพหน้าจอจริง (Playwright) ทั้ง Light และ Dark mode ตรวจสอบครบทุกแท็บย่อยใน \SystemSettingsScreen\ (Step 1-4)
