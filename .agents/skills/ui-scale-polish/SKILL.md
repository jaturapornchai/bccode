---
name: ui-scale-polish
description: ระบบสเกล/density และ premium skin ของ BC Ai Account (token, fluid scale, typography คนไทย 40+, premium standard, checklist ตรวจรับ) — ใช้เมื่อสร้างหรือปรับปรุง UI ทุกหน้าจอ (login, workspace, settings, menu) เพื่อให้สัดส่วนสวย อ่านง่าย สม่ำเสมอทุกจอทุก viewport โหลด skill นี้ก่อนแก้ไข CSS และสร้าง UI
---

# UI Scale & Polish System — BC Ai Account

ระบบสไตล์และสเกลหลักอยู่ที่ `frontend/src/app/globals.css` และ `frontend/src/app/zoom-control.tsx`
ออกแบบตามระบบนี้เท่านั้น ห้าม hardcode ค่า px ขนาดใหญ่ในคอมโพเนนต์ใหม่

---

## 1. กฎหลัก: บุคลิกคนไทย อายุ 40+ (Thai-First & High Accessibility)

ผู้ใช้หลักของระบบคือพนักงานบัญชีและเจ้าของกิจการอายุเกิน 40 ปี:
1. **ตัวหนังสือใหญ่อ่านง่าย**: ข้อความที่ต้องอ่านเพื่อตัดสินใจ $\ge 0.9\text{rem}$ (ห้ามต่ำกว่า $0.8\text{rem}$ กับข้อความสำคัญ)
2. **ภาษาไทยเป็นหลัก (Thai-First)**: ป้าย, ปุ่ม, feedback เป็นไทยก่อน; ศัพท์อังกฤษที่จำเป็นต้องมี title/คำอธิบายไทย
3. **ภาษาไทยซับซ้อน**: `line-height >= 1.45` ป้องกันสระบน/ล่างและวรรณยุกต์ทับกัน, ห้ามตัดกลางคำ, ข้อความยาวต้อง `overflow-wrap: anywhere`
4. **ปุ่มและจุดคลิกใหญ่พอ**: Action หลัก $\ge 2.6\text{em}$ ($\approx 44\text{px}+$); **ห้ามใช้ icon เปล่ากับ action สำคัญ** ต้องมีข้อความไทยกำกับ หรืออย่างน้อย `aria-label` + `title` ไทย
5. **คอนทราสต์สูง**: ข้อความ 4.5:1 / UI ใหญ่ 3:1 (WCAG AA); **ห้ามสื่อสถานะด้วยสีเดี่ยว** ต้องมีข้อความหรือไอคอนประกอบ
6. **หนึ่งจอ ทำงานเรื่องเดียว**: CTA หลักเดียวต่อจอ ลำดับอ่าน บน $\to$ ล่าง, ซ้าย $\to$ ขวา
7. **กันพลาดและยืนยันก่อนทำลาย**: ลบ/ทับข้อมูลต้องมี dialog ไทยอธิบายผลกระทบ; มี dirty guard เตือนก่อนทิ้งข้อมูลที่กรอก
8. **Feedback ทุก Action**: แจ้งสถานะสำเร็จ/ล้มเหลวด้วยภาษาไทยที่เข้าใจง่ายทันที ห้ามโชว์ error code ดิบ
9. **อย่าพาสายตากระโดด**: ห้าม auto-scroll ฉับพลัน; focus ring ชัดเจน; ตำแหน่ง dialog ใกล้จุดกด
10. **Motion ช้าและนิ่งพอ**: Transition $\le 300\text{ms}$; เคารพ `prefers-reduced-motion`

---

## 2. กฎความพรีเมี่ยม (System-wide Premium Standards)

1. **หนึ่งจอ หนึ่งตระกูลสี**: Accent ทุกจุด (CTA, focus ring, icon, glow, hover) derive จาก `--primary` ผ่าน `color-mix`; **ห้าม hard-code สี** (เพราะระบบมี 10 พาเลต + dark mode)
2. **Light + Dark ต้องสวยเท่ากัน**: ใช้ตัวแปร `--login-*` / `--panel` / `--text` ให้ปรับตาม theme เอง; ทดสอบทั้ง 2 โหมดโดยการกดปุ่มสลับธีมจริง
3. **พื้นผิวมีมิติ ไม่แบน**: Card/Panel มุมโค้ง 16–28px, ขอบ hairline โปร่ง, เงาย้อมสี primary นุ่มนวล, inner highlight 1px บนพื้นหลังที่มี ambient glow
4. **Control เป็นภาษาเดียวกัน**: Input, ปุ่ม, selector ใช้ radius, ความสูง, hairline ชุดเดียวกัน
5. **ภาพประกอบบริบทธุรกิจไทย**: ใช้รูปถ่ายร้านค้า/คลัง/บัญชีบริบทไทยจริง (.webp $\ge 1600\text{px}$) มี vignette เหมาะสม ห้ามใช้ clipart/stock ฝรั่ง
6. **Popover/Dialog ห้ามโดนตัด**: ห้ามใส่ `overflow: hidden` บน panel ที่มี popover ลูก (dropdown, color picker) ให้ clip ที่ shell ชั้นนอกสุดเท่านั้น

---

## 3. สเกลแบบ Fluid & Baseline Rhythm

### Root Font-size Engine
```css
html { font-size: clamp(15px, calc(0.46875vw + 9px), 21px); }
/* 15px @<=1280px -> 18px @1920px -> 21px @>=2560px (ทั้งระบบ 100% = 150%) */
```
* **หน้า Login และ Holding ยกเว้น**: ใช้ attribute `html[data-login-scale="1"]` คืนสูตรเดิม `clamp(10px, 0.3125vw + 6px, 14px) !important` เพื่อรักษาสเกลที่ approve แล้ว
* `zoom-control.tsx`: คำนวณด้วยสูตร fluid เดียวกันคูณด้วย `zoom / 100`

### Baseline & Typography
* Baseline rhythm: `:root { --baseline: 0.25rem; }`
* ช่องไฟและ line-height ต้องเป็นจำนวนเต็มเท่าของ `--baseline`
* หัวข้อไทย: `line-height >= 1.45` พร้อม `margin-block: 0.2em 0.3em` เพื่อกันสระลอยชนบรรทัดบน

---

## 4. กฎการจัดวางคอมโพเนนต์และฟอร์ม (Form Controls & Density)

1. **Radio และ Checkbox ต้องใช้ Wrap**:
   * หากมีตัวเลือก Radio หรือ Checkbox หลายข้อ **ต้องจัดวางแบบแนวนอนและห่อบรรทัด (`flex flex-wrap gap-x-4 gap-y-2`)** เสมอ เพื่อประหยัดพื้นที่ความสูงหน้าจอ (Vertical Space)
   * ต้องใส่คลาส `w-auto` กำกับที่ `<label>` เพื่อป้องกัน global CSS `label { width: 100% }` บีบให้ตกบรรทัด
2. **ความสูงปุ่มในแถวเดียวกันต้องเท่ากัน (Uniform Height)**:
   * ปุ่มใน toolbar / action bar แถวเดียวกันต้องสูงเท่ากันเสมอ (ความสูงมาตรฐาน `36px` หรือ `min-height: 2.6em`)
3. **การแยก Chrome กับ Content**:
   * Toolbar, Nav, Header บีบให้กระชับได้ (High density)
   * Card และส่วนอ่านเนื้อหา (Content area) ต้องโปร่งสบาย (`padding 1rem+`, `line-height 1.45-1.5`)
4. **Input ไอคอน**:
   * ขนาดและตำแหน่งไอคอนในช่องกรอกต้องใช้หน่วย `em` เสมอ เพื่อขยายตาม font-size โดยไม่ทับตัวหนังสือ
5. **High Information Density (ทุกจอแสดงข้อมูลให้มากในคราวเดียว ไม่ต้องเลื่อนเยอะ — ตั้งโดยลุงจืด 2026-09-04)**:
   * ผู้ใช้ ERP และนักบัญชีต้องการสแกนข้อมูลจำนวนมากอย่างรวดเร็ว หลีกเลี่ยงการเสียพื้นที่แนวตั้ง (Vertical Sprawl)
   * **โหมด `listdata` (ตาราง/รายการข้อมูล)**:
     - **แถวกระชับ (Compact Rows)**: ความสูงแถว $\approx 28\text{--}32\text{px}$ (`py-0.5` ถึง `py-1`, คลาส `.bc-list-row.is-compact`) บรรจุข้อมูลสำคัญใน 1 บรรทัด (Single-line)
     - **Single-line Truncation**: ใช้ `.bc-cell-text` เพื่อตัดข้อความยาวด้วย ellipsis (`white-space: nowrap !important; overflow: hidden !important; text-overflow: ellipsis !important;`) พร้อมใส่ attribute `title={value}` สำหรับ hover tooltip เสมอ
     - **Density Toggle**: มีปุ่มสลับ "ย่อบรรทัด" / "ขยายบรรทัด" ใน Toolbar เสมอ โดยกำหนดค่าเริ่มต้นเป็นย่อบรรทัด (`compactRows = true`) และจดจำค่าลง `localStorage`
     - **รวม Toolbar & Stats**: Action buttons, search box, และ stats badges รวมอยู่ในแถวเดียวกัน หลีกเลี่ยงการแยกเป็นหลายชั้นจนกินพื้นที่ความสูง
   * **โหมด `view` (ส่วนแสดงรายละเอียด / ข้อมูลเดี่ยว)**:
     - **Multi-column Card Layout**: จัดวาง Card เป็น 2 คอลัมน์บนจอใหญ่ (`lg:grid-cols-2` หรือ `xl:grid-cols-[1fr_1.5fr]`) แทนที่จะเรียงยาวต่อกันเป็นแถวเดี่ยว
     - **Wide Field Grids**: กางฟิลด์ข้อมูลภายใน Card แนวนอน 3–4 คอลัมน์ (`grid-cols-2 sm:grid-cols-3 xl:grid-cols-4`) เพื่อใช้ประโยชน์จากความกว้างหน้าจอเดสก์ท็อป
     - **Compact Field Cells**: ช่องไฟฟิลด์ `px-2.5 py-1.5`, ป้ายชื่อ `text-[10px]` หรือ `text-[11px]`, ค่า `text-xs` หรือ `text-sm`
     - **ซ่อนกล่องว่างโดยอัตโนมัติ (Zero Empty Placeholders)**: ถ้าไม่มีข้อมูลในหมวดนั้น (`visibleFields.length === 0 && !showEmptyFields`) ให้ซ่อนการ์ดทั้งใบ (คืนค่า `null`) ห้ามแสดง Card เปล่า "ยังไม่มีข้อมูลในส่วนนี้" มาดันข้อมูลจริงลงล่าง
     - **Header Banner กะทัดรัด**: แบนเนอร์หัวข้อข้อมูลสรุปลด padding เหลือ `p-3` ถึง `p-3.5` (ห้ามใช้ `p-6`) พร้อมวางปุ่ม Action สำคัญต่อท้ายในบรรทัดเดียวกัน
   * **โหมด `edit` (หน้าจอและฟอร์มบันทึก/แก้ไข)**:
     - **จัดฟอร์ม 2–3 คอลัมน์**: ฟิลด์กรอกข้อมูลจัดเป็นกริดแนวนอน (`grid sm:grid-cols-2 lg:grid-cols-3`) ช่องกรอกสูงมาตรฐาน 36px (`h-9`)
     - **Sticky Horizontal Tabs**: แบ่งหมวดหมู่ฟอร์มด้วย Tab แนวนอนที่ยึดติดด้านบน (`sticky top-0`) พร้อมปุ่ม Tab กะทัดรัด (`py-1.5 px-3 text-xs`) และใช้ `flex-wrap` ห้ามซ่อนหรือตัด Tab
     - **ช่องไฟและ Padding กะทัดรัด**: การ์ดและเซกชันใช้ `p-3` และ `space-y-2.5` ถึง `space-y-3` (ห้ามใช้ `p-6` หรือ `space-y-6`)
     - **Radio/Checkbox Wrap**: ตัวเลือก radio และ checkbox ห่อแนวนอน (`flex flex-wrap gap-2`)
     - **จำกัดความสูงพรีวิวสื่อ**: บาร์โค้ด SVG, พรีวิวรูปภาพ, หรือกล่อง Media จำกัดความสูงเหมาะสม (เช่น บาร์โค้ด `max-h-16`) เพื่อไม่ให้ดันปุ่ม Save/Cancel หลุดขอบล่างของหน้าจอ
     - **Internal Scroll Only**: Modal หรือ Pane ของฟอร์มต้องกำหนดความสูงสูงสุดสัมพันธ์กับ viewport (`max-h-[85vh]` ถึง `[90vh]`) โดยส่วนหัว (Header) และปุ่มบันทึก (Footer) อยู่กับที่ และให้เลื่อนเฉพาะส่วนเนื้อหาฟอร์ม (`overflow-y-auto`) เท่านั้น
    * **แบบแผน: ทางลัดส่วนตัวผู้ใช้ (Personalized Shortcuts) — เพิ่ม/ลบ/ย้าย (ตั้งโดยลุงจืด 2026-09-05)**:
      - **แยกตามตัวตนผู้ใช้ (Per-user Storage)**: บันทึกรายการ ID เมนูลง `localStorage` แยกคีย์ตาม user (`bc_user_shortcuts_v1:{backend}:{username}`)
      - **Fail-closed ตามสิทธิ์**: กรองแสดงเฉพาะเมนูที่อยู่ใน `allowedMenuIds` เสมอ
      - **ออกแบบปุ่มเพื่อคนไทย 40+**: จัดการลำดับด้วยปุ่มลูกศร `[▲ เลื่อนขึ้น]` และ `[▼ เลื่อนลง]` พร้อมปุ่ม `[ลบ]` ชัดเจน ไม่บังคับ Drag-and-Drop เพียงอย่างเดียว เพื่อความแม่นยำบนจอสัมผัสและเมาส์
      - **ค้นหาภาษาไทยยืดหยุ่น**: ช่องค้นหาเมนูต้องใช้ `menuSearchMatches` ซึ่งตัดวรรณยุกต์/สระไทย (`normalizeMenuSearchText`) และค้นหาได้หลายภาษาพร้อมกัน
      - **มีปุ่มรีเซ็ต**: มีปุ่ม "รีเซ็ตเป็นค่าเริ่มต้น" (Smart Fallback) ให้ผู้ใช้ย้อนกลับได้เสมอหากจัดลำดับพลาด

---

## 5. CSS Architecture & การแก้ไข

* **Non-destructive Skin Pass**: เมื่อปรับแต่ง skin หรือแก้ UI ใหม่ ให้เขียนเป็นบล็อกต่อท้าย `frontend/src/app/globals.css` พร้อมระบุวันที่และเหตุผล
* **Scope Selector**: ใช้ class เฉพาะเจาะจงนำหน้า (เช่น `.login-shell`, `.workspace-page`, `.bc-list-*`) เพื่อชนะ cascade โดยไม่แตะต้องโครงสร้าง layout เดิม
* **ตัวแปรสี**: ใช้ `var(--primary)`, `color-mix(in srgb, var(--primary) X%, transparent)` และ `--text-*` เสมอ

---

## 6. Checklist ตรวจรับงาน UI ก่อนบอกเสร็จ

ก่อนส่งมอบงาน UI ต้องตรวจหลักฐานจริงครบทุกข้อ:
- [ ] **Viewport Check**: ตรวจสอบทั้ง 3 ขนาดหน้าจอ (1280, 1920, 2560) Layout ไม่แตก ไม่ตกขอบ
- [ ] **Dual Theme**: ทดสอบทั้ง Light Mode และ Dark Mode จริง (กดปุ่มสลับธีม) คอนทราสต์อ่านออก
- [ ] **Thai Text Safety**: ตรวจดูสระบน/ล่างและวรรณยุกต์ไทย ต้องไม่ทับซ้อนกับขอบหรือบรรทัดอื่น
- [ ] **No Text Clip**: ไม่มีตัวหนังสือหรือปุ่มใดถูกตัดขาดหรือล้นขอบจอ (`scrollHeight <= clientHeight`)
- [ ] **Popover Safety**: Dropdown, Dialog, Datepicker เปิดแล้วไม่ถูกตัดหรือจมหายไปใต้ Card
- [ ] **Clean Console**: ไม่มี Runtime errors หรือ Warning ค้างใน Dev Console
- [ ] **Code Verification**: รัน `npm run typecheck` และ Unit tests ที่เกี่ยวข้องผ่าน 100%

---

## 7. เอกสารประวัติและบทเรียนย้อนหลัง (Archive)

รายละเอียดเชิงลึกและบันทึกประวัติการแก้บักเฉพาะกรณี (เคส 4.1 ถึง 4.39) ดูได้ที่:
* 👉 [references/case-studies-and-gotchas.md](references/case-studies-and-gotchas.md)