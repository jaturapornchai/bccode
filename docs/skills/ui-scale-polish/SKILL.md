---
name: ui-scale-polish
description: ระบบสเกล/density และ premium skin ของ BC Ai Account (token, fluid scale, typography คนไทย 40+, premium standard, checklist ตรวจรับ) — ใช้เมื่อสร้างหรือปรับปรุง UI ทุกหน้าจอ (login, workspace, settings, menu) เพื่อให้สัดส่วนสวย อ่านง่าย สม่ำเสมอทุกจอทุก viewport โหลด skill นี้ก่อนแก้ไข CSS และสร้าง UI
---

# UI Scale & Polish System — BC Ai Account

ระบบสไตล์และสเกลหลักอยู่ที่ `frontend/src/app/globals.css` และ `frontend/src/app/zoom-control.tsx`
ออกแบบตามระบบนี้เท่านั้น ห้าม hardcode ค่า px ขนาดใหญ่ในคอมโพเนนต์ใหม่

---

## 0. กฎเหล็ก: Upgrade Skill เสมอเมื่อแก้ UX/UI ทั้งระบบ (ตั้งโดยลุงจืด 2026-09-06)

> [!IMPORTANT]
> **ห้ามวนกลับไปใช้แบบเดิมเด็ดขาด (Zero Regression Guarantee)**:
> ทุกการแก้ไข ปรับปรุง หรือค้นพบบทเรียน/กับดักใน UX/UI ที่เป็นมาตรฐานกลางหรือใช้ทั้งระบบ (เช่น การใส่ Icon ในปุ่มทางลัด, การใช้ `!pl-10` หลบ Icon ในช่องกรอก, การคุมความสูง 42px+, สี Palette, ความหนาแน่น Density) **ต้องอัปเกรดไฟล์ SKILL.md นี้ทันที และ commit พร้อมงานเสมอ**
> 
> AI ทุกตัว (Gemini, Claude, Codex) **ต้องอ่าน Skill นี้ก่อนเริ่มงาน UI ทุกครั้ง** ห้ามลอกเลียนแบบโค้ดเก่าในจุดที่ยังไม่ได้อัปเกรด — หากพบหน้าจอเดิมที่ยังใช้แพทเทิร์นเก่า ต้อง refactor ให้ตรงตามมาตรฐานล่าสุดนี้ทันที เพื่อให้ระบบเหมือนกันทั้งระบบและไม่วนกลับไปใช้แบบเดิม

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
   * **กับดัก global padding `!important`**: ใน `globals.css:379` มีกฎ `input:not(...) { padding: 0px 4px !important; }` ดังนั้น input ที่มี leading icon ซ้อนอยู่ข้างใน **ต้องใช้คลาส `!pl-10` เสมอ** (มีเครื่องหมาย `!`) เพื่อชนะ `!important` ของ global CSS มิฉะนั้น padding-left จะถูกทับเหลือ 4px ทำให้ icon ทับตัวหนังสือไทย และต้องใส่ `pointer-events-none` บนไอคอนเสมอเพื่อไม่ให้ดักจับคลิก
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
     * **แบบแผน: ตารางรายการข้อมูล Master Data — Single-line Baseline Rhythm & Pinned Actions (แก้ปัญหา Icon ตกบรรทัด — ตั้งโดยลุงจืด 2026-09-06)**:
       1) **แบบแผนใหม่ (New Standard Pattern)**:
          - **บังคับ Single-line แถวเดียว ไม่ตกบรรทัด (`flex-nowrap items-center`)**: แถวข้อมูล `.bc-list-row` และหัวตาราง `.bc-list-header` ต้องใช้ `flex-nowrap items-center` หรือ CSS Grid คุมตายตัว ห้ามแถวตารางข้อมูลตกบรรทัดเด็ดขาด
          - **Baseline Rhythm Alignment**: ควบคุมความสูงแถวและช่องไฟด้วย `--baseline: 0.25rem` (หน่วยกริด 4px)
            * หัวตาราง: `min-height: calc(var(--baseline) * 8)` (32px), `padding: 4px 8px`, `gap: 6px`
            * แถวข้อมูลมาตรฐาน: `min-height: calc(var(--baseline) * 9)` (36px), `padding: 4px 8px`, `gap: 6px` (หรือ 28–32px ในโหมด Compact)
            * ปุ่ม Action ในแถว: ขนาดมาตรฐาน `size-7` (28px), icon `size-3.5` (14px), อยู่กึ่งกลางความสูง 36px พอดีอย่างสมดุล (บน-ล่างเหลือ 4px พอดี) ขนานกับเส้นฐานของข้อความในแถว 100%
          - **ตรึงคอลัมน์ปุ่มจัดการขวาสุดเสมอ (Pinned Right Actions)**:
            * คลาส `.bc-list-actions`: `shrink-0 ml-auto w-16 sm:w-20 flex items-center justify-end gap-1`
            * หัวตารางคอลัมน์ "จัดการ" และปุ่ม Action แถวข้อมูลต้องใช้ความกว้างเท่ากัน (`w-16 sm:w-20 shrink-0 text-right`) เพื่อให้ตรงแนวกันแบบ Pixel-perfect
          - **Responsive Fluid Columns & Ellipsis Truncation**:
            * คอลัมน์รหัส (Code): ใช้ `basis-28 grow-[1.2] min-w-[64px]` (ตัวหนา / Badge กะทัดรัด)
            * คอลัมน์ชื่อ (Name): ใช้ `basis-36 grow-[2] min-w-[80px]` ยืดหยุ่นรับพื้นที่ส่วนใหญ่
            * คอลัมน์รอง (เช่น สิทธิ์บริษัท/สถานะ): ใช้ `hidden md:inline-flex` หรือ `hidden lg:inline-flex` เพื่อหลบให้คอลัมน์หลักเมื่อหน้าจอหรือ Split Pane แคบ และขยายตัวเมื่อผู้ใช้ลากแบ่ง Pane กว้างขึ้น
            * ทุกเซลล์ข้อความใช้ `.bc-cell-text` หรือคลาส truncation พร้อมใส่ `title={value}` เพื่อให้อ่านข้อความเต็มผ่าน hover tooltip ได้เสมอ
       2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
          - **ห้ามใส่ `flex-wrap: wrap` บน `.bc-list-header` หรือ `.bc-list-row`เด็ดขาด**: เมื่อพื้นที่แนวนอนจำกัด (เช่น Master-Detail Split Pane กว้าง 30% หรือ ~300px) คอลัมน์สุดท้ายจะถูกบีบให้ตกบรรทัดไปอยู่ชั้นที่ 2 ด้านล่างขวา ทำให้ icon โดนผลักไปก้นช่อง และความสูงแถวบวมขึ้นเป็น 2 เท่า (~64px) เสียพื้นที่โดยเปล่าประโยชน์
          - **ห้ามล็อก `shrink-0` บนทุกคอลัมน์พร้อมกัน**: การล็อก `shrink-0` กับ `min-w-[100px]+` หลายๆ คอลัมน์จะทำให้ความกว้างรวมล้น Container เสมอ
          - **กับดัก `.truncate` ใน `globals.css:27-31`**: ในโปรเจกต์นี้มีกฎ `.truncate { white-space: normal !important; }` บังคับไว้เพื่อไม่ให้ข้อความทั่วไปซ่อนหลัง `...` ดังนั้นในตารางข้อมูลที่ต้องการตัดคำบรรทัดเดียว **ต้องใช้คลาส `.bc-cell-text` หรือกฎ `.bc-list-row span { white-space: nowrap !important; overflow: hidden !important; text-overflow: ellipsis !important; }` เท่านั้น**
       3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
          - **High Information Density & Visual Rhythm**: ผู้ใช้ 40+ และนักบัญชีต้องการสแกนข้อมูลจำนวนมากอย่างเป็นระเบียบ การที่ปุ่ม Action ตกไปอยู่บรรทัดที่สองทำให้ตารางดูเบี้ยวและสายตากระโดด การจัดเป็นบรรทัดเดียวทำให้สแกนได้เร็วและแสดงผลได้หลายสิบรายการพร้อมกัน
          - **Mathematical Balance**: การคุมขนาดด้วย `--baseline` rhythm (ปุ่ม 28px ในแถว 36px พร้อม padding 4px) ทำให้จุดศูนย์กลางของ Icon ตรงกับจุดกึ่งกลางของฟอนต์ภาษาไทยขนาด 15–18px อย่างแม่นยำ ไม่เอียง ไม่จม
       4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
          - Global CSS Single-line & Baseline overrides: [`frontend/src/app/globals.css`](frontend/src/app/globals.css#L8430-L8475)
          - Header & Row implementation: [`frontend/src/app/system-settings/system-settings-screen.tsx`](frontend/src/app/system-settings/system-settings-screen.tsx#L3479-L3625)
          - Proportional column scaling & truncation: [`frontend/src/app/system-settings/system-settings-screen.tsx`](frontend/src/app/system-settings/system-settings-screen.tsx#L4085-L4120)

    * **แบบแผน: ทางลัดส่วนตัวผู้ใช้ (Personalized Shortcuts) — แยกจอเต็ม ไม่ใช้ Popup (ตั้งโดยลุงจืด 2026-09-05; ปรับปรุงจอเต็ม High Density 2026-09-06)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **แยกเป็นหน้าจอเต็ม (Dedicated Full Screen / Tab)**: เมื่อผู้ใช้กด "จัดการทางลัด" หรือ "+ เพิ่มทางลัด" ให้เปิดเป็นแท็บหน้าจอเต็ม `⭐ จัดการทางลัด` (`route: "/shortcuts"`) เพื่อใช้พื้นที่หน้าจอเดสก์ท็อปอย่างเต็มที่ ไม่ใช้ Modal Popup ขนาดเล็ก
         - **โครงสร้าง Split View 2 คอลัมน์ (High Information Density)**:
           * **ซ้าย: คลังเมนูทั้งหมด (Catalog Grid)**: กริดการ์ด 2-3 คอลัมน์ (`grid sm:grid-cols-2 xl:grid-cols-3 gap-2.5`) แสดงไอคอน `MenuRouteIcon`, ชื่อไทยหนา, เส้นทาง route, สถานะ `✓ เพิ่มแล้ว` หรือปุ่ม `+ เพิ่มเป็นทางลัด`
           * **ขวา: รายการทางลัดปัจจุบัน (Sticky Reorder Panel)**: ตรึงสายตาด้านขวา (`sticky top-3`) แสดงลำดับ `1..N`, ไอคอน, ปุ่มเลื่อน `[▲]` `[▼]` (พร้อม tooltip/aria-label) และปุ่ม `[ลบออก]`, พร้อมข้อความยืนยันการบันทึกอัตโนมัติ
           * **บน: แถบจำลองผลจริง (Live Preview Strip)**: แสดงตัวอย่างปุ่มทางลัดแบบ Real-time ตามที่ผู้ใช้ปรับแต่งทันที
         - **ตัวกรองหมวดหมู่และค้นหาด่วน**:
           * แถบปุ่มกรองหมวดหมู่งานแนวนอน (Horizontal Pills): ทั้งหมด, งานประจำ (ซื้อ/ขาย/คลัง), ข้อมูลหลัก (สินค้า/คู่ค้า), รายงาน, การเงินและบัญชี, ตั้งค่าระบบ พร้อมตัวเลข badge แสดงจำนวนเมนูในหมวดนั้น
           * ช่องค้นหาด่วนใช้คลาส `!pl-10 !pr-10 [&::-webkit-search-cancel-button]:appearance-none` เพื่อหลบ icon ซ้าย และซ่อนปุ่ม cancel ซ้ำซ้อนของเบราว์เซอร์
           * ตัวเลือกกรอง `[ ] แสดงเฉพาะเมนูที่ยังไม่ได้เพิ่มเข้าทางลัด` (Checkbox toggle)
         - **การซิงก์ข้อมูลข้ามแท็บแบบ Real-time**: บันทึกลง `localStorage` และยิง Custom Event `window.dispatchEvent(new CustomEvent("bc_shortcuts_updated"))` เพื่อให้หน้าภาพรวม (Dashboard Home) อัปเดตรายการทันทีโดยไม่ต้องรีโหลดหน้าเว็บ
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามใช้ Modal Popup (`role="dialog"`) กับการจัดการคลังเมนูขนาดใหญ่**: พื้นที่ Popup ที่ถูกจำกัดความกว้าง/ความสูงทำให้แสดงผลได้ทีละไม่กี่เมนู ผู้ใช้ 40+ มองภาพรวมยากและต้องเลื่อน scroll ซ้ำซ้อน
         - **ห้ามลืมซ่อน native cancel button**: ใน `input[type="search"]` หากมีปุ่ม clear `[✕]` แบบ custom ต้องใส่ `[&::-webkit-search-cancel-button]:appearance-none` เสมอ เพื่อป้องกัน icon กากบาทซ้อนกัน 2 ตัว
         - **ห้ามบังคับ Drag-and-Drop เพียงอย่างเดียว**: ต้องมีปุ่มลูกศร `[▲]` และ `[▼]` ขนาดสัมผัสชัดเจน ($\ge 30\text{px}$) สำหรับผู้ใช้จอสัมผัสและเมาส์
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **High Information Density บนจอกว้าง**: พนักงานบัญชีและเจ้าของกิจการต้องการเห็นตัวเลือกทั้งหมดในพริบตา การแยกเป็นหน้าจอเต็มทำให้สามารถสแกนเมนูเป็นร้อยรายการและค้นหาได้เร็วกว่า Popup หลายเท่า
         - **Zero Context Loss**: การเปิดเป็นแท็บในระบบ ช่วยให้ผู้ใช้สลับกลับไปดูหน้าภาพรวม หรือเปิดหน้าจออื่นคู่ขนานได้โดยที่สถานะการค้นหาหรือคลังเมนูไม่สูญหาย
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
          - หน้าจอจัดการทางลัดเต็มจอ: [`frontend/src/app/menu/manage-shortcuts-screen.tsx`](frontend/src/app/menu/manage-shortcuts-screen.tsx)
          - การเชื่อมแท็บและ Route `/shortcuts`: [`frontend/src/app/menu/main-menu-screen.tsx`](frontend/src/app/menu/main-menu-screen.tsx#L683-L702) และ [`WorkTabPanel`](frontend/src/app/menu/main-menu-screen.tsx#L2540-L2555)
          - หน้า Dashboard ภาพรวมที่ถอด Modal ออก: [`frontend/src/app/menu/dashboard-home.tsx`](frontend/src/app/menu/dashboard-home.tsx#L163-L175)

    * **แบบแผน: แถบเมนูข้างปรับความกว้างได้ (Resizable Navigation Sidebar with Grip Handle & Persistence) (ตั้งโดยลุงจืด 2026-09-07)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **ปรับความกว้างได้อิสระ (Resizable Sidebar Width)**: แถบเมนูนำทางด้านซ้ายใน `main-menu-screen.tsx` ปรับความกว้างได้ตั้งแต่ 260px ถึง 540px (ค่าเริ่มต้น 320px) โดยใช้ CSS variable `--menu-sidebar-width` กับ Tailwind arbitrary grid `lg:grid-cols-[var(--menu-sidebar-width)_minmax(0,1fr)]` บน desktop และคง `grid-cols-[minmax(0,1fr)]` บน mobile
         - **มือจับลากชัดเจน (Prominent Floating Grip Handle)**:
           * บริเวณมือจับกว้าง 16px (`w-4 -right-2 top-0 z-30`) ซ้อนกึ่งกลางเส้นขอบขวาของ `<aside>` พอดี จับง่ายทั้งเมาส์และจอสัมผัส
           * เส้นไฮไลต์แนวตั้งตลอดความสูง (`inset-y-0 left-1/2 w-0.5 group-hover:bg-primary/50`) และปุ่มเม็ดยาตรงกลางพร้อมไอคอนกริป `GripVertical` (`h-10 w-3 rounded-full border bg-background/95 shadow-sm group-hover:h-12 group-hover:border-primary/50`)
           * ขณะลาก (`isResizingSidebar`): เม็ดยาขยายเป็น `h-14 bg-primary text-primary-foreground`, แสดงคอร์เซอร์ `col-resize` ทั่วทั้งจอ และระงับ text selection (`select-none`)
         - **จดจำค่าอัตโนมัติและทางลัดรีเซ็ต**:
           * บันทึกค่าลง `localStorage` คีย์ `bc_menu_sidebar_width` ทันทีที่ปล่อยมือ
           * **ดับเบิ้ลคลิก (Double-click)**: คืนค่าความกว้างเริ่มต้น (320px) ทันที
           * **Keyboard Accessibility**: รองรับปุ่ม `ArrowLeft` / `ArrowRight` (ขยับทีละ 16px), `Home` (ต่ำสุด 260px), `End` (สูงสุด 540px), และ `Enter`/`Space` (รีเซ็ตเป็น 320px)
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้าม hardcode ความกว้างแถบเมนูด้านซ้ายตายตัว เช่น `lg:grid-cols-[280px_...]`**: เมนูภาษาไทยของระบบ ERP บัญชีมีข้อความยาว (เช่น "รายงานการเงิน", "ตรวจและซ่อมข้อมูล", "แท็กโครงการและแผนก") หากความกว้างแคบเกินไป ข้อความจะถูกตัดตกบรรทัด ทำให้อ่านยากและรกสายตา
         - **ห้ามใส่ inline `gridTemplateColumns` ทับโดยตรงบน root grid**: เพราะจะทำให้ layout จอมือถือ (< lg) ถูกบีบเป็น 2 คอลัมน์พัง ต้องใช้ตัวแปร CSS `--menu-sidebar-width` ผ่าน Tailwind responsive class `lg:grid-cols-[var(--menu-sidebar-width)_...]` เท่านั้น
         - **ห้ามทำมือจับลากเป็นเส้นบาง 1px ที่เล็งยาก**: คนไทย 40+ เล็งเมาส์ยาก ต้องมี hit area กว้างอย่างน้อย 16px และมี visual grip pill แสดงสถานะชัดเจน
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **Thai-First 40+ Accessibility**: ป้ายเมนูภาษาไทยต้องการความกว้างอย่างน้อย 300–320px เพื่อให้ข้อความส่วนใหญ่เรียงตัวในบรรทัดเดียว (Single-line) ไม่ตกบรรทัด
         - **Universal Display Compatibility**: หน้าจอผู้ใช้มีตั้งแต่แล็ปท็อป 13 นิ้ว ไปจนถึงจอเดสก์ท็อป 4K การให้ผู้ใช้ลากปรับความกว้างได้เองและจดจำค่าถาวรช่วยให้ทุกคนปรับให้เหมาะกับสายตาและขนาดจอของตนเองได้สมบูรณ์แบบ
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - โค้ดควบคุมและ Drag handle: [`frontend/src/app/menu/main-menu-screen.tsx`](frontend/src/app/menu/main-menu-screen.tsx#L950-L1085)
         - ชุดการทดสอบความกว้างและการจดจำ: [`frontend/src/app/menu/main-menu-sidebar-resize.test.ts`](frontend/src/app/menu/main-menu-sidebar-resize.test.ts)

    * **แบบแผน: ปรับ Padding & Margin ให้น้อยลงเพื่อแสดงข้อมูลได้เยอะที่สุด (Ultra-High Density Spacing) (ตั้งโดยลุงจืด 2026-09-07)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **ลด Padding & Margin ทั้งระบบ (System-wide Compaction)**:
           * `:root` Density Tokens: `--density-page-pad: clamp(4px, 1vw, 8px)`, `--density-card-pad: clamp(6px, 1.2vw, 8px)`, `--density-modal-pad: clamp(8px, 1.5vw, 12px)`, `--density-gap: 4px`, `--density-gap-lg: 6px`, `--density-control-h: 34px`, `--density-control-compact-h: 30px`
           * Card & Panels: `CardHeader` และ `CardContent` ลด padding เหลือ `p-2.5` และ `gap-1`
           * Tables & Data Lists: `table th` สูง `26px` padding `3px 6px`, `table td` padding `2px 6px text-xs`, `.bc-list-row` min-height `26px` padding `2px 6px`
           * Sidebar Navigation: เมนูซ้ายใช้ `p-1.5 gap-1`, ปุ่มหมวดหมู่และเมนูย่อยปรับความสูงกระชับ `min-h-7` ถึง `min-h-8` padding `px-2 py-1` (แสดงเมนูได้มากกว่าเดิม 35% โดยไม่ต้องเลื่อน)
           * Dashboard Home: กริดเอกสาร `p-2 gap-1.5`, ทางลัด `px-2.5 py-1.5 text-xs`, ประวัติความเคลื่อนไหว `px-2 py-1 text-xs` (หน้าจอ 1080p แสดงครบทุกส่วนในสายตาเดียว)
         - **ยกเว้นหน้าจอ Login และ Holding**: ใช้ selector `main:not(.login-shell):not(.holding-shell)` เพื่อรักษาสเกลและ layout hero ที่ผ่านการ approve แล้ว
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามใช้ padding ขนาดใหญ่ (เช่น `p-5`, `p-6`, `gap-4`) บนหน้าจอจัดการข้อมูลบัญชี/ERP**: ทำให้เกิดพื้นที่ว่างเปล่าแนวตั้ง (Vertical Sprawl) นักบัญชีต้อง scroll ตลอดเวลา
         - **ห้ามทำแถวตารางสูงเกิน 36px ในโหมดข้อมูลปกติ**: ทำให้แสดงผลได้ทีละ 5-8 รายการ การย่อลงมาเหลือ 26-28px ช่วยให้มองเห็นข้อมูลได้ 15-20 รายการพร้อมกัน
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **High Information Density**: ผู้ใช้งานระบบบัญชีต้องการกวาดสายตาตรวจข้อมูลจำนวนมากได้อย่างรวดเร็วในหน้าจอเดียว
         - **Screen Real Estate Optimization**: การลด padding จาก 16-24px เหลือ 6-10px คืนพื้นที่หน้าจอให้กับตารางและฟอร์มข้อมูลจริงถึง 30–50%
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - Global high-density pass: [`frontend/src/app/globals.css`](frontend/src/app/globals.css)
         - Card & Table primitives: [`frontend/src/components/ui/card.tsx`](frontend/src/components/ui/card.tsx) และ [`frontend/src/components/ui/table.tsx`](frontend/src/components/ui/table.tsx)
         - Sidebar & Menu compaction: [`frontend/src/app/menu/main-menu-screen.tsx`](frontend/src/app/menu/main-menu-screen.tsx)
         - Dashboard overview compaction: [`frontend/src/app/menu/dashboard-home.tsx`](frontend/src/app/menu/dashboard-home.tsx)

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

---

## 8. เพิ่มเมนูใหม่ในเมนูหลัก (ต้องแก้ครบ 4 จุด เสมอ)

**แบบแผนใหม่ (New Standard Pattern)** — เพิ่ม 1 เมนู = แก้ 4 ไฟล์ ถ้าขาดข้อใดข้อหนึ่ง unit test จะ fail ทันที:

1. `frontend/src/lib/menu-data.ts` — เพิ่มบรรทัดใน `MENU_SECTIONS` ด้วยเฮลเปอร์ `tx(id, th, en, route, category)` (ค่า default ของ category คือ `"transaction"`); กลุ่มใหม่ใช้ `ml(key, th, en)` เป็น title
   ```ts
   tx("sales-by-customer", "ยอดขายตามลูกค้า", "Sales by Customer", "/report/salesbycustomer", "report"),
   ```
2. `backend/assets/language/languages.tsv` — เพิ่ม 1 แถวต่อ 1 key **ครบ 13 คอลัมน์** (`key th en cn ja km ko lo my vi ms id fil`) คั่นด้วย TAB; key = id ที่แทน `-` ด้วย `_` (`menuKey()` ใน `menu-data.ts:31`) ไฟล์นี้ EOL ผสม (ท้ายไฟล์เป็น CRLF) → เขียนต่อท้ายด้วยสคริปต์ที่คุม newline เอง อย่าใช้ Edit tool
3. `frontend/src/lib/menu-icons.ts` — เพิ่ม `"<route>": "<iconKey>",` ใน `ROUTE_ICON_KEYS` โดยเลือกจาก union `MenuIconKey` ที่มีอยู่ ห้ามคิด key ใหม่
4. `frontend/src/lib/menu-icons.test.ts` — อัปเดตจำนวนใน `expect(items).toHaveLength(N)` ให้เท่าจำนวนเมนูใหม่ทั้งหมด

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern)**

* ห้ามเพิ่มเมนูโดยไม่เพิ่มแถวภาษา — เทสต์ `has backend language keys for every menu item` และ `has all supported language cells...` จะ fail และผู้ใช้ภาษาอื่นจะเห็น slug
* ห้ามเติมแค่ th/en แล้วปล่อยคอลัมน์อื่นว่าง — เทสต์ตรวจครบทุกภาษา
* ห้ามตั้ง id ซ้ำ — id ถูกใช้เป็นรหัสสิทธิ์ (`keeps menu item ids unique for permission codes`) และถูกอ่านโดย `role-screen-matrix.tsx` / `permission-editors.tsx`
* ห้ามใส่หน้าจอตั้งค่าองค์กร (บริษัท/สาขา/ผู้ใช้/สิทธิ์/พนักงาน/ภาษา/สกุลเงิน) ลงเมนูหลัก — อยู่ใน workspace wizard และมีเทสต์ห้ามไว้ (`has no settings section in the main menu`)

**เหตุผลทางเทคนิค (Root Cause & Rationale)**

`menuText()` ดึงป้ายจาก dictionary ของ backend ด้วย key ก่อน แล้วค่อย fallback มาที่ป้ายในโค้ด — เมนูที่ไม่มี key จะโชว์ slug ให้ผู้ใช้เห็นตอน backend ตอบ dictionary มาแล้ว; ส่วนไอคอนใช้ `menuIconKeyForRoute()` ที่ fallback เป็นไอคอนประจำหมวด ทำให้เมนูหลายตัวหน้าตาเหมือนกันจนแยกไม่ออก (ผิดกฎคนไทย 40+ ที่ต้องแยกจุดคลิกได้ด้วยตา) จึงบังคับให้ทุก route มีไอคอนของตัวเองด้วยเทสต์

**ไฟล์อ้างอิงจริง (Reference Implementation)**

* 15 เมนู parity กับ FlowAccount/PEAK 2026-09-08: `frontend/src/lib/menu-data.ts` รายการ `business-dashboard` / `purchase-tax-invoice-register` / `import-partner` (ระบบเงินเดือนถูกตัดออกจากขอบเขตวันเดียวกัน — ห้ามเพิ่มเมนู payroll/ภ.ง.ด.1/ประกันสังคมกลับเอง)
* เหตุผลและช่องว่างที่ยังไม่ตัดสิน: `docs/kms/19-menu-coverage-flowaccount-peak.md`, ADR `docs/kms/decisions/2026-09-08-menu-parity-flowaccount-peak.md`
* วิธีตรวจ: `npx vitest run src/lib/menu-data.test.ts src/lib/menu-icons.test.ts src/lib/menu-usage.test.ts` + `npx tsc --noEmit` + เปิดเมนูจริง ค้นชื่อไทยที่เพิ่ม แล้วดูทั้ง light/dark ด้วยการกดปุ่มสลับธีม

---
