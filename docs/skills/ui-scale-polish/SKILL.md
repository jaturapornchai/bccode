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
          - Global CSS Single-line & Baseline overrides: [`frontend/src/app/globals.css`](../../../frontend/src/app/globals.css#L8430-L8475)
          - Header & Row implementation: [`frontend/src/app/system-settings/system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx#L3479-L3625)
          - Proportional column scaling & truncation: [`frontend/src/app/system-settings/system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx#L4085-L4120)

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
          - หน้าจอจัดการทางลัดเต็มจอ: [`frontend/src/app/menu/manage-shortcuts-screen.tsx`](../../../frontend/src/app/menu/manage-shortcuts-screen.tsx)
          - การเชื่อมแท็บและ Route `/shortcuts`: [`openManageShortcuts()`](../../../frontend/src/app/menu/main-menu-screen.tsx#L700-L717) และ [`WorkTabPanel`](../../../frontend/src/app/menu/main-menu-screen.tsx#L2678-L2718)
          - หน้า Dashboard ภาพรวมที่ถอด Modal ออก: [`frontend/src/app/menu/dashboard-home.tsx`](../../../frontend/src/app/menu/dashboard-home.tsx#L163-L175)

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
         - ตรรกะควบคุมความกว้าง/จดจำค่า/คีย์บอร์ด: [`frontend/src/app/menu/main-menu-screen.tsx`](../../../frontend/src/app/menu/main-menu-screen.tsx#L952-L1045) · โหลดค่าที่จดจำไว้ตอน mount: [`frontend/src/app/menu/main-menu-screen.tsx`](../../../frontend/src/app/menu/main-menu-screen.tsx#L484-L489) · Grip handle (markup + `role="separator"` + `onKeyDown`): [`frontend/src/app/menu/main-menu-screen.tsx`](../../../frontend/src/app/menu/main-menu-screen.tsx#L1092-L1133)
         - ชุดการทดสอบความกว้างและการจดจำ: [`frontend/src/app/menu/main-menu-sidebar-resize.test.ts`](../../../frontend/src/app/menu/main-menu-sidebar-resize.test.ts)

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
         - Global high-density pass: [`frontend/src/app/globals.css`](../../../frontend/src/app/globals.css)
         - Card & Table primitives: [`frontend/src/components/ui/card.tsx`](../../../frontend/src/components/ui/card.tsx) และ [`frontend/src/components/ui/table.tsx`](../../../frontend/src/components/ui/table.tsx)
         - Sidebar & Menu compaction: [`frontend/src/app/menu/main-menu-screen.tsx`](../../../frontend/src/app/menu/main-menu-screen.tsx)
         - Dashboard overview compaction: [`frontend/src/app/menu/dashboard-home.tsx`](../../../frontend/src/app/menu/dashboard-home.tsx)
    * **แบบแผน: ผังคลังสินค้า 2 ระดับ คลัง → ที่เก็บ (Warehouse → Location Master-Detail Inline Table) (ตั้งโดยลุงจืด 2026-09-10)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **โครงสร้าง 2 ระดับ คมชัด**: ยึดลำดับ `คลังสินค้า (Warehouse) → ที่เก็บสินค้า (Storage Location)` ตัดระดับชั้นวาง/ที่วางสินค้า (Bins) ออกทั้งหมด เพื่อลดความซับซ้อนและให้สอดคล้องกับพฤติกรรมการใช้งานจริง
         - **Master-Detail Layout**:
           * **ซ้าย (Master List)**: รายชื่อคลังสินค้าใน Sidebar ปรับความกว้างได้ (Resizable 220px–520px) แสดงรหัสคลัง, ชื่อคลัง, จำนวนที่เก็บสินค้า และปุ่มเพิ่มคลังสินค้า
           * **ขวา (Detail Table)**: ตารางที่เก็บสินค้าแบบแก้ไขข้อมูลและบันทึกพร้อมกันทีเดียว (Batch Save All) ประกอบด้วย 5 คอลัมน์คมชัด: ลำดับ (# พร้อม NEW/MOD badge), รหัสที่เก็บสินค้า (`e.g. ZONE-A`), ชื่อที่เก็บภาษาไทย, ชื่อที่เก็บภาษาอังกฤษ (ถ้าเปิดใช้ภาษา EN), และปุ่มลบ (Trash2)
         - **เข้าหน้าจอเริ่มดูข้อมูลเก่าได้ทันที (Zero Blocking Modal on Enter)**: เมื่อเปิดหน้าจอเข้ามา ให้เลือกคลังแรกเป็นค่าเริ่มต้นและแสดงตารางที่เก็บสินค้าทันที ห้ามเด้ง Dialog/Modal ขึ้นมาบังหน้าจอ
         - **Batch Save พร้อมสถานะ Dirty Guard**: มีปุ่ม "เพิ่มแถว" สำหรับสร้างแถวใหม่ในตาราง, แถบสรุปจำนวนรายการและแถวที่แก้ไข, ปุ่ม "คืนค่าเดิม" (Reset) และปุ่ม "บันทึกทั้งหมด" (Save All) ที่ส่งคำขอ POST/PUT/DELETE ไปยัง backend พร้อมกัน
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามเพิ่มระดับที่วางสินค้า (Bins / ชั้นวาง) กลับเข้ามาในหน้าจอนี้เด็ดขาด**: การมี 3 ระดับทำให้ตารางแน่นเกินไปและต้องเปิด Modal ซ้อน Modal ทำให้ผู้ใช้สับสน
         - **ห้ามเปิดหน้าต่าง Modal สร้างคลังขึ้นมาอัตโนมัติตอนเข้าหน้าจอ**: ผู้ใช้ต้องการเข้ามาดูหรือเลือกคลังเดิมก่อนเสมอ
         - **ห้ามใช้ระบบบันทึกแยกทีละแถวในตารางที่เก็บสินค้า**: ทำให้ต้องคลิกบันทึกซ้ำซ้อนหลายรอบ
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **Thai 40+ Accessibility & Usability**: การตัด Dialog ซ้อน Dialog ช่วยลด cognitive load และป้องกันสายตากระโดดตามกฎ UX 40+
         - **High Information Density**: ผู้ใช้สามารถกรอกและตรวจทานที่เก็บสินค้าได้หลายสิบรายการในหน้าจอเดียว
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - หน้าจอผังคลังและตารางที่เก็บสินค้า: [`frontend/src/app/system-settings/warehouse-tree-view.tsx`](../../../frontend/src/app/system-settings/warehouse-tree-view.tsx)
         - สเปกการตั้งค่าเมนูและฟิลด์: [`frontend/src/lib/system-setting-screens.ts`](../../../frontend/src/lib/system-setting-screens.ts#L1455-L1472)
         - การทดสอบ E2E อัตโนมัติ: [`frontend/e2e/product-warehouse-crud.spec.ts`](../../../frontend/e2e/product-warehouse-crud.spec.ts#L197-L265)
    * **แบบแผน: ตัวแบ่งและปรับความกว้างแนวตั้งมาตรฐานทั้งระบบ (Universal Resizable Splitter with Floating Pill Handle) (ตั้งโดยลุงจืด 2026-09-10)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **คอมโพเนนต์กลาง `ResizableSplitter`**: ทุกจุดที่มีการปรับขนาดความกว้างระหว่างคอลัมน์ซ้าย-ขวา (List-Detail Pane) ในระบบ ต้องใช้คอมโพเนนต์กลาง `<ResizableSplitter />` จาก `frontend/src/components/ui/resizable-splitter.tsx`
         - **Visual Aesthetics (รูปลักษณ์พรีเมี่ยมตามหน้าจอผังคลังสินค้า)**:
           * เส้นแกนแนวตั้งบางประณีต (`h-full w-0.5 transition-colors duration-150 rounded-full`, `bg-border/60` -> เมื่อชี้ hover เปลี่ยนเป็น `bg-primary/50` -> ขณะลาก resizing เปล่งสี `bg-primary`)
           * ปุ่มเม็ดยาลอยตรงกลาง (Floating Pill) พร้อมไอคอน `GripVertical`: กว้าง 14px สูง 32px (`h-8 w-3.5 rounded-full border border-border/80 bg-background/95 shadow-2xs`), เมื่อ hover ขยายเป็น `h-10 border-primary/50 text-primary`, ขณะลาก resizing ขยายเป็น `h-12 border-primary bg-primary text-primary-foreground shadow-md`
           * Hitbox กว้างพอสำหรับเมาส์และจอสัมผัส: `w-3 shrink-0 cursor-col-resize select-none touch-none -mx-1` ป้องกันการหลุดโฟกัสขณะลาก
         - **ฟังก์ชันครบวงจร**:
           * **Keyboard Accessibility**: รองรับ `ArrowLeft` / `ArrowRight` (ขยับทีละ 1%), `Home` (ขยับไปต่ำสุด min), `End` (ขยับไปสูงสุด max)
           * **Double-click Reset**: ดับเบิ้ลคลิกเพื่อคืนค่าความกว้างเริ่มต้น (Default Width)
           * **ARIA Standard**: มี `role="separator"`, `aria-orientation="vertical"`, `aria-valuenow`, `aria-valuemin`, `aria-valuemax`, `tabIndex={0}`
           * **Responsive Breakpoint**: รองรับพร็อพ `breakpoint="lg" | "xl"` ให้ซ่อนบนจอมือถือและแสดงบนจอเดสก์ท็อปอย่างถูกต้อง
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามใช้เส้นแบ่งเรียบๆ 1.5px/1px ไม่มีมือจับ (เช่น `hidden w-1.5 shrink-0 cursor-col-resize bg-border/70 transition hover:bg-primary/50 lg:block`)**: ทำให้ผู้ใช้มองไม่เห็นว่าเป็นจุดที่สามารถลากปรับขนาดได้ และดูไม่สวยงามไม่สม่ำเสมอกับส่วนอื่นของระบบ
         - **ห้ามขาดคีย์บอร์ดหรือดับเบิ้ลคลิกรีเซ็ต**: การปรับขนาดต้องรองรับทั้งเมาส์ สัมผัส และคีย์บอร์ดเสมอ
         - **ห้ามเขียนโค้ด markup มือจับซ้ำซ้อนกันในแต่ละหน้า**: รวมศูนย์ไว้ที่ `frontend/src/components/ui/resizable-splitter.tsx` ที่เดียว
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **Thai 40+ Accessibility & Ergonomics**: ผู้ใช้งานอายุ 40+ ต้องการ Visual Affordance ที่ชัดเจนว่าอะไรกดได้ ลากได้ มือจับเม็ดยาพร้อมไอคอนกริปบอกหน้าที่ทันทีโดยไม่ต้องเดา
         - **Zero Regression & Consistency**: รวมศูนย์สไตล์และการโต้ตอบเป็นหนึ่งเดียว ทำให้ทั้งระบบมีพฤติกรรมและความสวยงามเหมือนกัน 100%
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - คอมโพเนนต์กลาง: [`frontend/src/components/ui/resizable-splitter.tsx`](../../../frontend/src/components/ui/resizable-splitter.tsx)
         - การนำไปใช้ใน Master-Detail ตั้งค่า: [`frontend/src/app/system-settings/system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx#L3940)
         - การนำไปใช้ใน BOM Editor: [`frontend/src/app/system-settings/system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx#L2896)
         - การนำไปใช้ใน ผังคลังสินค้า: [`frontend/src/app/system-settings/warehouse-tree-view.tsx`](../../../frontend/src/app/system-settings/warehouse-tree-view.tsx#L366)
         - การนำไปใช้ใน สินค้า: [`frontend/src/app/menu/product-screen.tsx`](../../../frontend/src/app/menu/product-screen.tsx#L1774)
         - การนำไปใช้ใน บาร์โค้ด: [`frontend/src/app/menu/product-barcode-screen.tsx`](../../../frontend/src/app/menu/product-barcode-screen.tsx#L1152)
         - ชุดการทดสอบ Unit Tests: [`frontend/src/components/ui/resizable-splitter.test.ts`](../../../frontend/src/components/ui/resizable-splitter.test.ts)
    * **แบบแผน: การป้องกันไอคอนในช่องค้นหาโดนทับ และการจัดเรียงผังบัญชีแบบต้นไม้พร้อมตัวคุมสิทธิ์ลงรายการ (Chart of Accounts Tree & Input Icon Clearance) (ตั้งโดยลุงจืด 2026-09-12)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **การจัดตำแหน่งไอคอนค้นหาและปุ่ม Clean ใน Input**:
           * ไอคอน Search ซ้าย: วางเป็น absolute จัดกึ่งกลางแนวตั้งเสมอ `pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 size-5 text-muted-foreground z-10`
           * ปุ่ม Clean ขวา: `absolute right-3 top-1/2 -translate-y-1/2 z-10 grid size-7 place-items-center rounded-full hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer`
           * Input ช่องค้นหา: ต้องระบุ `!pl-11 !pr-10` (หรือ `!pl-9.5 !pr-9` สำหรับขนาดเล็ก) ที่มีเครื่องหมาย `!` เสมอ เพื่อชนะทั้ง `control` (`px-3`) และ global CSS `padding: 0px 4px !important;` ป้องกันตัวหนังสือหรือ placeholder ชนไอคอน
           * Stacking Context: ทั้งไอคอนซ้ายและปุ่มขวาต้องมี `z-10` เสมอ เพื่อไม่ให้ถูกพื้นหลังทึบของ Input (`bg-card`, `bg-background`) หรือ focus ring กลืนหายไป
         - **การจัดเรียงผังบัญชีแบบต้นไม้ (Hierarchical Tree Sorting & Effective Level)**:
           * จัดลำดับ 5 หมวดบัญชีมาตรฐาน: สินทรัพย์ (1) $\to$ หนี้สิน (2) $\to$ ส่วนของเจ้าของ (3) $\to$ รายได้ (4) $\to$ ค่าใช้จ่าย (5)
           * คำนวณ `effectiveLevel`: ราก = 1, ลูก = `parent.level + 1` อัตโนมัติ ป้องกันข้อมูล legacy ในฐานข้อมูลที่ขาดฟิลด์ level
           * เรียงต้นไม้สัมพันธ์: วางบัญชีย่อยต่อท้ายบัญชีแม่ (`parentaccountcode`) เสมอ พร้อมการเยื้อง `style={{ paddingLeft: `${Math.max(0, level - 1) * 18}px` }}` และสัญลักษณ์ `└─`
           * การคุมสิทธิ์เลือกบัญชี (Posting Guard): ในโหมดเลือกบัญชีเดียวสำหรับลงรายการ (`!multiSelect && !all`) หากเป็นบัญชีคุม (`!allowposting`) ต้องแสดง Badge "บัญชีคุม" และบล็อกการกดปุ่ม Enter หรือ Double-click
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามวางไอคอน Search หรือ Clean โดยไม่มี `top-1/2 -translate-y-1/2` และ `z-10`**: หาก input มี `bg-card` หรือ font ปรับขนาด ไอคอนจะถูกทับหรือเบี้ยวหลุดกึ่งกลาง
         - **ห้ามลืมใส่ `!` ใน padding ช่องค้นหา (เช่น ใช้แค่ `pl-11 pr-10`)**: จะถูก `.px-3` หรือ global input CSS ชนะ ทำให้ตัวหนังสือทับไอคอน
         - **ห้ามกำหนด `onlyPosting: true` เป็นค่าเริ่มต้นใน Dialog ค้นหาผังบัญชี**: จะทำให้บัญชีคุมถูกกรองทิ้งหมด และบัญชีย่อยระดับ 2 ปรากฏเดี่ยวๆ บนตารางโดยไม่มีหัวข้อแม่
         - **ห้ามพึ่งพาค่า `level` จากฐานข้อมูลดิบอย่างเดียว**: ข้อมูลใน MongoDB/PostgreSQL อาจไม่มีฟิลด์ level ให้คำนวณจากความสัมพันธ์ `parentaccountcode` เสมอ
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **Thai 40+ Ergonomics**: ผู้ใช้มองเห็นโครงสร้างบัญชีชัดเจนตามมาตรฐานบัญชีไทย บัญชีคุมชัดเจนไม่ทำให้กดผิด
         - **CSS Cascade Specificity**: Tailwind class `.px-3` ในตัวแปร `control` และ global stylesheet มักชนะ `.pr-9` ทำให้ต้องใช้ `!important` เพื่อรับประกันความปลอดภัยของ padding
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - ฟังก์ชันการคำนวณและจัดเรียงต้นไม้: [`frontend/src/lib/general-ledger.ts`](../../../frontend/src/lib/general-ledger.ts) (`sortAccountsHierarchically`, `getEffectiveAccountLevel`)
          - จอค้นหาผังบัญชีเต็มจอ: [`frontend/src/app/gl/account-search-dialog.tsx`](../../../frontend/src/app/gl/account-search-dialog.tsx)
          - คอมโพเนนต์ช่องค้นหา Baseline: [`frontend/src/app/gl/gl-common.tsx`](../../../frontend/src/app/gl/gl-common.tsx) (`SearchInput`, `AccountSelect`)
          - ชุดการทดสอบ Unit Tests: [`frontend/src/app/gl/search-input.test.ts`](../../../frontend/src/app/gl/search-input.test.ts) และ [`frontend/src/app/gl/account-search-dialog.test.ts`](../../../frontend/src/app/gl/account-search-dialog.test.ts)
    * **แบบแผน: ตารางรายการข้อมูลหลักแบบ CRUD มาตรฐานพร้อมปุ่มจัดการ แก้ไข/ลบ รายบรรทัด (Master CRUD List Table Pattern with Row-Level Actions) (ตั้งโดยลุงจืด 2026-09-12)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **แถบคอลัมน์ "จัดการ" (Actions Column) ประจำทุกแถว**:
           * มีคอลัมน์ `จัดการ` ทางด้านขวาสุดของตารางเสมอ (`th className="p-2.5 text-right pr-3 w-24"`)
           * ทุกแถวมีปุ่มจัดการ 2 ตัวเรียงชิดขวา:
             - ปุ่มแก้ไข (Pencil icon): `size-7 rounded-md bg-background text-primary border-primary/30 hover:bg-primary/15 hover:border-primary/50 shadow-none transition-colors shrink-0` พร้อม `title="แก้ไข (Edit)"` และ `aria-label="แก้ไข"`
             - ปุ่มลบ (Trash2 icon): `size-7 rounded-md bg-background text-red-600 border-red-200/70 hover:bg-red-50 hover:border-red-300 dark:text-red-400 dark:border-red-900/50 dark:hover:bg-red-950/40 shadow-none transition-colors shrink-0` พร้อม `title="ลบ (Delete)"` และ `aria-label="ลบ"`
           * กดปุ่มลบในแถวจะเรียก Dialog ยืนยัน (`useConfirmDialog` แบบ danger) พร้อมอธิบายรหัสและชื่อรายการ โดยไม่ต้องเปิดฟอร์มแก้ไขขึ้นมาก่อน
           * หยุด propagation (`event.stopPropagation()`) เพื่อไม่ให้ทริกเกอร์ row click เมื่อกดปุ่มจัดการ
         - **การจัดฟิลด์จำนวนเงินและตัวเลข (Numeric Column Alignment)**:
           * ฟิลด์จำนวนเงิน (เช่น งบประมาณ `budgets`, ประมาณการกระแสเงินสด `forecast`) ต้องแยกเป็นคอลัมน์เฉพาะ `จำนวนเงิน`
           * จัดข้อความชิดขวา `text-right font-mono font-semibold tabular-nums whitespace-nowrap` พร้อมจัดรูปแบบจุลภาคคั่นหลักพันและทศนิยม 2 ตำแหน่ง
         - **ป้ายสถานะมีสีสื่อความหมาย (Color-coded Status Badges)**:
           * ใช้งาน: `bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20`
           * ล็อกแล้ว: `bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20`
           * ปิดปีแล้ว / ปิดใช้งาน: `bg-muted text-muted-foreground border border-border`
         - **การโต้ตอบของแถว (Row Click & Active Ring Highlight)**:
           * แถวคลิกได้ทั้งแถว (`cursor-pointer`) เมื่อคลิกจะเปิดฟอร์มแก้ไขหรือดูรายละเอียด
           * แถวที่กำลังเลือกหรือแก้ไขแสดงไฮไลต์ชัดเจน `bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium`
         - **แผงฟอร์มแก้ไข (Editor Workbench Header & Footer)**:
           * มี Header แสดงไอคอนสถานะ (Pencil หรือ Plus), หัวข้อ (`แก้ไข <CODE>` หรือ `เพิ่มรายการใหม่`), ป้ายเตือนข้อมูลยังไม่บันทึก, และปุ่มปิด `X`
           * เมื่อยังไม่เลือกแถวใด ให้แสดง Empty state ที่มีการ์ดแนะนำพร้อมปุ่ม CTA `+ เพิ่มรายการใหม่`
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามทำตารางข้อมูลหลักที่ไม่มีปุ่มจัดการ (Actionless Rows)**: การบังคับให้ผู้ใช้คลิกเลือกแถวก่อนแล้วไปกดลบในฟอร์มด้านข้างเพียงทางเดียว ขัดกับความคุ้นเคยของคนไทย 40+ ที่คุ้นเคยกับหน้าจอตั้งค่ามาตรฐาน (เช่น `/productunit`)
         - **ห้ามรวมจำนวนเงินเข้าไปในข้อความชื่อรายการ (e.g. `งบกันยายน · 100,000.00`)**: ทำให้อ่านยากและเปรียบเทียบตัวเลขในตารางไม่ได้ ต้องแยกเป็นคอลัมน์จำนวนเงินชิดขวาเฉพาะ
         - **ห้ามแสดงสถานะเป็นข้อความธรรมดาไม่มีสี**: สื่อสารความแตกต่างได้ยาก คน 40+ ต้องการ Badge ชัดเจน
         - **ห้ามลบข้อมูลโดยไม่มี Confirm Dialog และการป้องกันข้อมูลผังบัญชี**: ผังบัญชีที่มีการลงรายการแล้วต้องมีข้อความเตือนเด็ดขาดห้ามลบ
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **Thai 40+ Usability & Familiarity**: ผู้ใช้งานคุ้นเคยกับรูปแบบตาราง CRUD ของหน้าจอตั้งค่าพื้นฐานในระบบ (เช่น หน่วยนับสินค้า `/productunit`, หมวดสินค้า) การทำให้หน้าจอ Master ในระบบบัญชีมี Action Column เหมือนกันสร้างความสม่ำเสมอทั้งระบบ (System Consistency)
         - **Accidental Click Prevention**: การแยกปุ่มแก้ไขและปุ่มลบพร้อมกล่องยืนยันช่วยป้องกันความผิดพลาดของพนักงานบัญชี
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
          - หน้าจอจัดการข้อมูลหลักบัญชี: [`frontend/src/app/gl/gl-masters.tsx`](../../../frontend/src/app/gl/gl-masters.tsx)
          - มาตรฐานตารางรายการตั้งค่า: [`SettingDataList` ใน `frontend/src/app/system-settings/system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx#L4074-L4135)
          - ชุดการทดสอบ Unit Tests: [`frontend/src/app/gl/gl-masters.test.ts`](../../../frontend/src/app/gl/gl-masters.test.ts)

    * **แบบแผน: การแก้ไขตัวเลขและจำนวนเงิน — โฟกัสค่า 0 เป็นช่องว่าง (Focus-Zero-as-Empty), ป้องกันพิมพ์ 1 กลายเป็น 10, แทนที่ค่าเดิม และทำความสะอาดเลข 0 นำหน้า (Formatted Numeric Input & Zero Focus Pattern) (ตั้งโดยลุงจืด 2026-09-14)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **สถานะปกติเมื่ออยู่นิ่ง (Idle / Blur Mode)**:
           * แสดงผลตัวเลขพร้อม Thousands Comma และทศนิยมตาม `scale` (หรือ `decimals`) เสมอ เช่น `70,000.00`, `0.00`
           * จัดข้อความชิดขวาเสมอ (`text-right font-mono tabular-nums`)
         - **เมื่อเลื่อนหรือคลิกเข้าไปแก้ไข (Focus / Edit Mode)**:
           * สลับเป็น plain text ไม่มี comma เพื่อให้แก้ไขสะดวก
           * **กรณีค่าเดิมเป็น 0 หรือว่าง (Zero Focus)**: ต้องตั้งค่าข้อความในช่องกรอกให้เป็นสตริงว่าง (`""`) ทันที เพื่อให้เบราว์เซอร์แสดง placeholder สีเทาอ่อน (เช่น `0.00` หรือ `0`) เมื่อผู้ใช้กดตัวเลข เช่น กด `1` จะได้ค่า `1` ทันที **ไม่มีทางกลายเป็น `10` หรือ `01`**
           * **กรณีค่าเดิมไม่ใช่ 0 (Non-zero Focus)**: คงค่าตัวเลขดิบไว้ (เช่น `500.00`) และทำการเลือกข้อความทั้งหมด (`target.select()`) ภายใน `requestAnimationFrame` เพื่อให้พร้อมพิมพ์ตัวเลขใหม่ทับค่าเดิมได้ทันที
           * ดักจับ `onMouseUp` ด้วย `event.preventDefault()` ในจังหวะโฟกัส เพื่อไม่ให้การปล่อยเมาส์ล้างแถบไฮไลต์ที่ถูกเลือก
         - **การทำความสะอาดขณะพิมพ์ (Active Typing / Leading Zero Normalization)**:
           * ใน `onChange` ต้องตัดเลข 0 ที่อยู่หน้าตัวเลข 1–9 ทันทีด้วย Regex `cleaned.replace(/^(-?)0+([1-9])/, "$1$2")` เช่น หากมีเคอร์เซอร์แทรกแล้วพิมพ์ `"01"` จะแปลงกลับเป็น `"1"` อัตโนมัติ ป้องกันตัวเลขกลายพันธุ์
         - **เมื่อออกจากช่องกรอก (Blur / Commit Mode)**:
           * นำข้อความที่พิมพ์มาจัดรูปแบบ คืนค่าศูนย์ตามสเกล (เช่น ว่างเปล่า $\to$ `0.00`), ใส่ Thousands Comma, และ emit ค่าที่ถูกต้องกลับไปยัง state
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามใส่ข้อความ `"0"` ค้างไว้ในช่องกรอกขณะ Focus**: เมื่อผู้ใช้เลื่อนคีย์บอร์ด (Tab / ลูกศร) เข้ามาในช่อง เคอร์เซอร์ของเบราว์เซอร์จะไปอยู่ที่ตำแหน่งแรก (index 0) ก่อนเลข 0 การกดเลข 1 จะแทรกข้างหน้ากลายเป็น `"10"` ทันที ทำให้ยอดเงินผิดพลาดไป 10 เท่า
         - **ห้ามเรียก `target.select()` แบบ Synchronous ธรรมดาใน React**: เมื่อ React Re-render จะล้าง cursor selection ให้ใช้ `requestAnimationFrame(() => target.select())` เสมอ
         - **ห้ามแทรก Comma ขณะกำลังพิมพ์ (Active Typing)**: การแทรก comma ทุก keystroke ทำให้ตำแหน่งเคอร์เซอร์กระโดด พิมพ์ทศนิยมผิด และลบตัวเลขยาก
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **พนักงานบัญชีและความถูกต้องของการป้อนข้อมูล (Data Accuracy for Thai 40+)**: ในการลงบัญชีหรือคีย์ยอดเงิน ความผิดพลาดจาก `1` กลายเป็น `10` สร้างความเสียหายทางบัญชี การเคลียร์เป็นช่องว่างเมื่อค่าเป็น 0 และมี placeholder กำกับ ทำให้การคีย์ข้อมูลไหลลื่นและถูกต้อง 100%
         - **React Controlled Component Re-render Lifecycle**: การแยก `editText` ชั่วคราวระหว่างโฟกัส และ commit กลับเป็น formatted เมื่อ blur ป้องกัน React controlled loop ที่ล็อคค่าเลข 0 ไว้
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - คอมโพเนนต์จำนวนเงิน GL: [`frontend/src/app/gl/gl-common.tsx`](../../../frontend/src/app/gl/gl-common.tsx) (`AmountInput`)
         - คอมโพเนนต์ตัวเลขกลางระบบ: [`frontend/src/components/ui/numeric-input.tsx`](../../../frontend/src/components/ui/numeric-input.tsx) (`NumericInput`)
         - สเปกการตัดสินใจ: [`docs/kms/decisions/2026-09-12-formatted-numeric-input.md`](../../../docs/kms/decisions/2026-09-12-formatted-numeric-input.md)
         - การทดสอบ Unit Tests: [`frontend/src/app/gl/amount-input.test.ts`](../../../frontend/src/app/gl/amount-input.test.ts) และ [`frontend/src/components/ui/numeric-input.test.ts`](../../../frontend/src/components/ui/numeric-input.test.ts)
         - การทดสอบ Browser จริง: Playwright script [`scratch/test_numeric_focus.cjs`](../../../scratch/test_numeric_focus.cjs)

    * **แบบแผน: การป้องกันขอบโฟกัสถูกตัดขาด และขอบเขต Focus Ring สมบูรณ์รอบด้าน (Container Focus Clearance & Primary Border Pattern) (แก้ปัญหาขอบซ้ายหาย/ไม่ครบ — ตั้งโดยลุงจืด 2026-09-14)**:
      1) **แบบแผนใหม่ (New Standard Pattern)**:
         - **พื้นที่เผื่อ Focus Ring ใน Scroll Container (`p-1` หรือ `px-1.5 py-1`)**:
           * ในฟอร์มที่มีการเลื่อน (`overflow-y-auto` หรือ `overflow-auto`) **ต้องมี padding รอบด้านอย่างน้อย `p-1` (4px)** เสมอ
           * ห้ามใช้ `pr-1` เดี่ยวๆ โดยไม่มี `pl-1` เพราะ CSS `overflow-y: auto` บังคับให้ `overflow-x: hidden` อัตโนมัติ ทำให้เงาโฟกัส (`box-shadow: 0 0 0 2px`) ฝั่งซ้ายที่ยื่นออกไปพิกัดติดลบ (`x < 0`) ถูกขอบซ้ายของคอนเทนเนอร์ตัดขาด (Clipped) กลายเป็นเส้นตรงแบน ไม่มีวงแหวน
         - **ขอบเขตตัวจริงต้องเปลี่ยนเป็นสี Primary (`focus-visible:border-primary`)**:
           * คลาสกลาง `control` และคอมโพเนนต์ `<Input />` ต้องมี `focus-visible:border-primary` เสมอ
           * เมื่อโฟกัส เส้นขอบ 1px ของตัว Input เองจะเปลี่ยนเป็นสี `--primary` ทันที ร่วมกับ `focus-visible:ring-2 focus-visible:ring-ring` (หรือ `focus-visible:ring-primary/25`)
           * ทำให้เส้นขอบมีความคมชัด สมบูรณ์ 4 ด้าน 360 องศา โค้งมนตามรัศมี `rounded-xl` อย่างสวยงาม ไม่หลุดหายหรือซีดจาง
         - **แอนิเมชันนุ่มนวล**: ใส่ `transition-colors` เพื่อให้สีของขอบและเงาเปลี่ยนผ่านอย่างประณีต ไม่กระตุก
      2) **กับดัก/สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
         - **ห้ามใส่ `overflow-y-auto pr-1` บนฟอร์มหรือ fieldset โดยไม่มี padding ด้านซ้าย (`pl`)**: จะทำให้ขอบและเงาฝั่งซ้ายของช่องกรอกถูกตัดทิ้ง 100% จนเห็นเป็นขอบขาดในแนวตั้ง
         - **ห้ามพึ่งพาแค่ `ring-2` โดยไม่เปลี่ยน `border-color`**: หากไม่เปลี่ยนสี border เมื่อเกิดการ clip ที่เงาภายนอก เส้นขอบจริงจะค้างเป็นสีเทาอ่อน `#dcc0bc` ทำให้ผู้ใช้เห็นว่าขอบซ้ายหาย
      3) **เหตุผลทางเทคนิค (Root Cause & Rationale)**:
         - **CSS Overflow Specification Interaction**: ตามมาตรฐาน W3C CSS เมื่อกำหนด `overflow-y: auto/scroll` เบราว์เซอร์จะคำนวณ `overflow-x` เป็น `auto` หรือ `hidden` เสมอ หากไม่มี padding ซ้าย วัตถุที่ชิด `left: 0` จะถูกตัดเงาภายนอก (`box-shadow` offset นอกขอบเขต element) ออกทั้งหมด
         - **ความพรีเมี่ยมและความคมชัด (Thai 40+ Contrast & Aesthetics)**: เส้นขอบโค้งมนที่สมบูรณ์ทุกมุมช่วยให้ผู้ใช้รับรู้ขอบเขตของช่องที่กำลังพิมพ์ได้อย่างชัดเจน ไม่รู้สึกว่า UI บกพร่องหรือเส้นขาด
      4) **ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
         - นิยามคลาส `control`: [`frontend/src/app/gl/gl-common.tsx`](../../../frontend/src/app/gl/gl-common.tsx#L13)
         - คอมโพเนนต์ Input กลาง: [`frontend/src/components/ui/input.tsx`](../../../frontend/src/components/ui/input.tsx#L9)
         - คอนเทนเนอร์ฟอร์ม GL Master: [`frontend/src/app/gl/gl-masters.tsx`](../../../frontend/src/app/gl/gl-masters.tsx#L377) และ [`gl-masters.tsx`](../../../frontend/src/app/gl/gl-masters.tsx#L453)
         - คอนเทนเนอร์ฟอร์ม GL Journal: [`frontend/src/app/gl/gl-journals.tsx`](../../../frontend/src/app/gl/gl-journals.tsx#L343) และ [`gl-journals.tsx`](../../../frontend/src/app/gl/gl-journals.tsx#L490)
         - การทดสอบยืนยันผลด้วย Playwright: [`scratch/capture_padded.cjs`](../../../scratch/capture_padded.cjs)

---

## 5. CSS Architecture & การแก้ไข

* **Non-destructive Skin Pass**: เมื่อปรับแต่ง skin หรือแก้ UI ใหม่ ให้เขียนเป็นบล็อกต่อท้าย `frontend/src/app/globals.css` พร้อมระบุวันที่และเหตุผล
* **Scope Selector**: ใช้ class เฉพาะเจาะจงนำหน้า (เช่น `.login-shell`, `.workspace-page`, `.bc-list-*`) เพื่อชนะ cascade โดยไม่แตะต้องโครงสร้าง layout เดิม
* **ตัวแปรสี**: ใช้ `var(--primary)`, `color-mix(in srgb, var(--primary) X%, transparent)` และ `--text-*` เสมอ

---

## 6. Checklist ตรวจรับงาน UI ก่อนบอกเสร็จ

ก่อนส่งมอบงาน UI ต้องตรวจหลักฐานจริงครบทุกข้อ:
- [ ] **Viewport Check**: ตรวจครบ 4 ขนาดตามกฎพรีเมี่ยม `AGENTS.md` ข้อ 8 (กว้าง 1600, 1280, 1024 และ 768 แนวตั้ง = iPad ขึ้นไป) ทั้ง Light และ Dark — Layout ไม่แตก ไม่ตกขอบ
- [ ] **Dual Theme**: ทดสอบทั้ง Light Mode และ Dark Mode จริง (กดปุ่มสลับธีม) คอนทราสต์อ่านออก
- [ ] **State Coverage**: ตรวจ hover / focus (ring ชัด) / disabled / error ของ control หลักครบทุกสถานะ ทั้ง Light และ Dark ตามกฎพรีเมี่ยม `AGENTS.md` ข้อ 8
- [ ] **Thai Text Safety**: ตรวจดูสระบน/ล่างและวรรณยุกต์ไทย ต้องไม่ทับซ้อนกับขอบหรือบรรทัดอื่น
- [ ] **No Text Clip**: ไม่มีตัวหนังสือหรือปุ่มใดถูกตัดขาดหรือล้นขอบจอ (`scrollHeight <= clientHeight`)
- [ ] **Popover Safety**: Dropdown, Dialog, Datepicker เปิดแล้วไม่ถูกตัดหรือจมหายไปใต้ Card
- [ ] **Clean Console**: ไม่มี Runtime errors หรือ Warning ค้างใน Dev Console
- [ ] **Code Verification**: รัน `npm run typecheck` และ Unit tests ที่เกี่ยวข้องผ่าน 100%

---

## 7. เอกสารประวัติและบทเรียนย้อนหลัง (Archive)

รายละเอียดเชิงลึกและบันทึกประวัติการแก้บักเฉพาะกรณี (เคส 4.1 ถึง 4.40) ดูได้ที่:
* 👉 [references/case-studies-and-gotchas.md](references/case-studies-and-gotchas.md)

---

## 8. เพิ่มเมนูใหม่ในเมนูหลัก (ต้องแก้ครบ 4 จุด เสมอ)

**แบบแผนใหม่ (New Standard Pattern)** — เพิ่ม 1 เมนู = แก้ 4 ไฟล์ ถ้าขาดข้อใดข้อหนึ่ง unit test จะ fail ทันที:

1. `frontend/src/lib/menu-data.ts` — เพิ่มบรรทัดใน `MENU_SECTIONS` ด้วยเฮลเปอร์ `tx(id, th, en, route, category, languageKey?)` (ค่า default: `category = "transaction"`, `languageKey = menuKey(id)`); ถ้าจอนั้นมีคีย์ภาษาอยู่แล้วใน `languages.tsv` ให้ส่งพารามิเตอร์ที่ 6 เช่น `tx("marketplace-shopee", "เชื่อม Shopee", "Shopee Connection", "/marketplace/shopee", "master", "shopee_mappings")` แล้วข้ามข้อ 2 ได้ (ห้ามเพิ่มแถวซ้ำ); กลุ่มใหม่ใช้ `ml(id, th, en)` เป็น title
   ```ts
   tx("sales-by-customer", "ยอดขายตามลูกค้า", "Sales by Customer", "/report/salesbycustomer", "report"),
   ```
2. `backend/assets/language/languages.tsv` — เพิ่ม 1 แถวต่อ 1 key **ครบ 13 คอลัมน์** (`key th en cn ja km ko lo my vi ms id fil`) คั่นด้วย TAB; key = id ที่แทน `-` ด้วย `_` (`menuKey()` ใน `menu-data.ts:31`) **ยกเว้น** เมนูที่ส่ง `languageKey` เป็นพารามิเตอร์ที่ 6 ของ `tx()` ให้ยึดคีย์นั้นแทน; ไฟล์นี้ EOL ผสม (ท้ายไฟล์เป็น CRLF) → เขียนต่อท้ายด้วยสคริปต์ที่คุม newline เอง อย่าใช้ Edit tool
3. `frontend/src/lib/menu-icons.ts` — เพิ่ม `"<route>": "<iconKey>",` ใน `ROUTE_ICON_KEYS` โดยเลือกจาก union `MenuIconKey` ที่มีอยู่ ห้ามคิด key ใหม่
4. `frontend/src/lib/menu-icons.test.ts` — อัปเดตจำนวนใน `expect(items).toHaveLength(N)` ให้เท่าจำนวนเมนูใหม่ทั้งหมด

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern)**

* ห้ามเพิ่มเมนูโดยไม่เพิ่มแถวภาษา — เทสต์ `has backend language keys for every menu item` และ `has all supported language cells...` จะ fail และผู้ใช้ภาษาอื่นจะเห็น slug
* ห้ามเติมแค่ th/en แล้วปล่อยคอลัมน์อื่นว่าง — เทสต์ตรวจครบทุกภาษา
* ห้ามตั้ง id ซ้ำ — id ถูกใช้เป็นรหัสสิทธิ์ (`keeps menu item ids unique for permission codes`) และถูกอ่านโดย `role-screen-matrix.tsx` / `permission-editors.tsx`
* ห้ามสร้าง section id `settings` ในเมนูหลัก; `/employee`, `/user`, `/permissiongroup`, `/useraccessaudit` อยู่ระดับ Holding/Workspace เท่านั้นตามกฎ 2026-09-10 ห้ามคืนกลุ่ม `organization-people` ในเมนูสาขา (แทนข้อยกเว้นเก่า 2026-09-08)

**เหตุผลทางเทคนิค (Root Cause & Rationale)**

`menuText()` ยึดป้ายไทยจาก `label.th` (ตั้งแต่ 2026-09-09 เพื่อกันแคชเก่าเปลี่ยนความหมาย); ภาษาอื่นดึง dictionary ของ backend ด้วย key ก่อน แล้วค่อย fallback มาที่ป้ายในโค้ด — เมนูที่ไม่มี key จะโชว์ slug ให้ผู้ใช้เห็นตอน backend ตอบ dictionary มาแล้ว; ส่วนไอคอนใช้ `menuIconKeyForRoute()` ที่ fallback เป็นไอคอนประจำหมวด ทำให้เมนูหลายตัวหน้าตาเหมือนกันจนแยกไม่ออก (ผิดกฎคนไทย 40+ ที่ต้องแยกจุดคลิกได้ด้วยตา) จึงบังคับให้ทุก route มีไอคอนของตัวเองด้วยเทสต์

**ไฟล์อ้างอิงจริง (Reference Implementation)**

* ผังเมนูครอบคลุมมาตรฐานระบบบัญชี 2026-09-08 (รวม 224 รายการ): `frontend/src/lib/menu-data.ts` รายการ `business-dashboard` / `purchase-tax-invoice-register` / `import-partner` / `vat-pnd2` / `dimension-pnl` และกลุ่ม `marketplace`
* **ระบบเงินเดือนอยู่นอกขอบเขต** — ห้ามเพิ่มเมนู payroll / ภ.ง.ด.1 / ประกันสังคมกลับเอง (กฎใน `AGENTS.md`); แต่ **ภ.ง.ด.2 อยู่ในขอบเขต** เพราะเป็นภาษีหัก ณ ที่จ่ายเงินได้ 40(3)/(4) ที่มีต้นทางจากรายการจ่ายเงิน ไม่ใช่เงินเดือน
* เหตุผลและข้อกำหนดผังเมนู: `docs/kms/19-menu-coverage-market-standard.md`, ADR `docs/kms/decisions/2026-09-08-menu-parity-market-standard.md`
* วิธีตรวจ: `npx vitest run src/lib/menu-data.test.ts src/lib/menu-icons.test.ts src/lib/menu-usage.test.ts` + `npx tsc --noEmit` + เปิดเมนูจริง ค้นชื่อไทยที่เพิ่ม แล้วดูทั้ง light/dark ด้วยการกดปุ่มสลับธีม

---

## 8.1 ก่อนเพิ่มเมนู ให้เช็ค "จอกำพร้า" ก่อนเสมอ (บทเรียน 2026-09-08)

**แบบแผนใหม่** — ก่อนจะเขียนจอใหม่หรือสรุปว่า "ฟีเจอร์นี้ยังไม่มี" ต้องเทียบ route ที่ `main-menu-screen.tsx` เปิดได้ กับ route ที่มีในเมนูก่อน:

```bash
grep -ohE '"/[a-z0-9/-]+"' frontend/src/app/menu/main-menu-screen.tsx frontend/src/lib/system-setting-screens.ts | tr -d '"' | sort -u > /tmp/screens.txt
grep -oE '"/[^"]+"' frontend/src/lib/menu-data.ts | tr -d '"' | sort -u > /tmp/menu.txt
comm -23 /tmp/screens.txt /tmp/menu.txt | grep -vE '^/api/'
```

**กับดัก / สิ่งที่ห้ามทำซ้ำ** — เขียนจอเสร็จแล้วไม่เพิ่มรายการเมนู แล้วคิดว่า "เดี๋ยวค่อยต่อ" · สรุปว่าระบบไม่มีฟีเจอร์นั้นทั้งที่จอมีอยู่ แล้วเขียนจอซ้ำ · เชื่อว่า unit test จะจับให้ (ไม่จับ — เทสต์ตรวจจากเมนูไปหาไอคอน/ภาษา ไม่ได้ตรวจย้อนกลับ)

**เหตุผลทางเทคนิค** — `onOpenRoute` (`frontend/src/app/menu/main-menu-screen.tsx:1453-1455`) เปิดแท็บได้เฉพาะ route ที่หาเจอใน `allMenuItems` เท่านั้น จอที่ dispatch ไว้แล้วแต่ไม่มีรายการเมนูจึงเป็นโค้ดตายในสายตาผู้ใช้ ทั้งที่ไอคอนและ language key อาจเตรียมไว้ครบแล้ว

**ไฟล์อ้างอิงจริง** — `frontend/src/app/menu/marketplace-screen.tsx` (3 จอ Shopee/Lazada/TikTok ทำงานได้เต็มรูปแบบ แต่เข้าไม่ถึงจนถึง 2026-09-08 เพราะไม่มีเมนู; ไอคอนอยู่ที่ `menu-icons.ts:221-223` และ language key `shopee_mappings`/`lazada_mappings`/`tiktok_mappings`/`marketplace_connectors` มีใน `languages.tsv` มาก่อนแล้ว) · route ที่อยู่นอกเมนูโดยตั้งใจและไม่ต้องแก้: `/currency`, `/datamodelgraph`, `/shortcuts`, `/menu` (แท็บภาพรวม), `/workspace` (หน้าเลือกกิจการ) และ route ตั้งค่าองค์กรทั้งหมดใน `system-setting-screens.ts`

---

## 8.2 การจัดวาง Selector Grid ให้เต็มแนวนอน (Full-width Grid) ห้ามใช้ JS ResizeObserver

**แบบแผนใหม่ (New Standard Pattern)** — ตัวเลือกกลุ่ม / การ์ดตัวเลือกจำนวนคงที่ (เช่น ปุ่มกลุ่มหมวด 1-20) ต้องใช้ Tailwind CSS Grid ที่กระจายคอลัมน์สม่ำเสมอและกว้างเต็มพื้นที่ (`w-full`) โดยแบ่งตาม breakpoint เพื่อให้ลงตัวในทุกหน้าจอ:

```tsx
<main className="grid w-full grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5" data-testid="product-category-group-grid">
  {groups.map((item) => (
    <button
      type="button"
      key={item.id}
      className="group flex min-h-14 w-full items-center gap-2 rounded-xl border border-border bg-card px-2.5 py-2 text-left shadow-sm ..."
    >
      ...
    </button>
  ))}
</main>
```

- ปุ่มหรือการ์ดภายในใช้ `w-full` เพื่อยืดเต็มช่องกริดเสมอ
- กำหนด breakpoint ให้หารจำนวนการ์ดลงตัวสวยงาม (เช่น 20 รายการ: `xl:grid-cols-5` = 4 แถวพอดี, `lg:grid-cols-4` = 5 แถวพอดี, `sm:grid-cols-2` = 10 แถวพอดี)

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามใช้ JS `useLayoutEffect` / `ResizeObserver` เพื่อคำนวณความกว้างการ์ดแบบ dynamic**: เสี่ยงต่อ mount-timing bug (เช่น ขณะ initial mount ติด `loading` element ยังไม่ render ใน DOM ทำให้ ref เป็น `null` และคำนวณไม่ได้) ส่งผลให้การ์ด fallback ไปใช้ขนาด intrinsic แคบๆ
- **ห้ามใช้ `flex-wrap` คู่กับ `flex-none` หรือ inline style `width/flexBasis`**: ทำให้เกิดช่องว่างสีดำ/พื้นหลังโล่งขนาดใหญ่ฝั่งขวา (unused horizontal space) เมื่อการ์ดขึ้นบรรทัดใหม่แล้วแบ่งพื้นที่ไม่เต็มความกว้างคอนเทนเนอร์
- **ห้ามฮาร์ดโค้ดค่า minimum width (เช่น `const GROUP_CARD_MIN_WIDTH_PX = 280`)**: ทำให้จำนวนคอลัมน์ที่ได้กระโดดและเหลือเศษพื้นที่ขวาสุดเสมอ

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
CSS Grid เป็น native browser layout engine ที่คำนวณพื้นที่แบบ sub-pixel accuracy ทันทีใน render tree เดียว โดยไม่มี layout shift (CLS), ไม่ต้องรอ JS hydration หรือ lifecycle hooks, และไม่มีปัญหา ref timing เมื่อมี conditional loading state ยิ่งไปกว่านั้น CSS Grid ยังรับประกันว่าการ์ดทุกใบจะขยายเต็มความกว้างของ grid cell 100% เสมอ ทำให้ไม่มีช่องว่างว่างเปล่าทางขวาของหน้าจอ

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- `frontend/src/app/system-settings/product-category-tree-view.tsx` (ปุ่มเลือกกลุ่มหมวด 1-20 ในหน้าจัดหมวดสินค้า `data-testid="product-category-group-grid"`)

---

## 8.3 แบบแผน UX/UI สำหรับ Tree View ลากวางและปุ่มจัดการบนแถว (Unified Tree UX/UI)

**แบบแผนใหม่ (New Standard Pattern)** — หน้าจอจัดการโครงสร้างแบบต้นไม้ (เช่น กลุ่มสินค้า `productgroup` และ จัดหมวดสินค้า `productcategorygroupselectscreen`) ต้องมี UX/UI ที่เป็นมาตรฐานเดียวกันทั้งระบบ:

1. **Row Actions Toolbar บนแถวรายการ**:
   - `GripVertical`: ตัวจับลากทางซ้ายสุด สำหรับจัดลำดับและเปลี่ยนระดับ
   - `ChevronUp` / `ChevronDown`: ปรับลำดับขึ้น-ลงระหว่างพี่น้อง (Sibling reorder) สำหรับผู้ใช้ที่ไม่ถนัดลากเมาส์
   - `FolderPlus` (สีเขียว `text-emerald-600`): เพิ่มรายการย่อยใต้รายการนั้นโดยตรง โดยไม่ต้องกวาดสายตาขึ้น Header
   - `Edit3` (สีน้ำเงิน `text-blue-600`): เปิดฟอร์มแก้ไขรายการนั้นทันที
   - `Trash2` (สีแดง `text-destructive`): ลบรายการ พร้อม Confirm Dialog ภาษาไทยก่อนทำลาย
2. **ระดับสีลำดับชั้น (Multi-level Hierarchy Styles)**:
   - ใช้ 5 ระดับสีที่ตัดกันชัดเจน (`GROUP_LEVEL_STYLES` / `CATEGORY_LEVEL_STYLES`) แทนการใช้สีเดียวทั้งต้นไม้: Level 0 (Primary), Level 1 (Sky), Level 2 (Emerald), Level 3 (Amber), Level 4 (Violet)
3. **การลากวาง (Drag & Drop Hierarchy)**:
   - **วางเป็นลูก (Drop Inside)**: มีชิปป้ายเขียวเด่นชัด `data-*-inside-drop-guid` แสดงตอนลากผ่าน
   - **วางเป็นหลัก (Drop as Root)**: มีแถบสีเขียวด้านล่างสุดของรายการต้นไม้ `data-*-root-drop="true"` สำหรับดึงรายการลูกกลับมาเป็นระดับหลัก
   - **Undo / Redo Toolbar**: มีปุ่ม "เลิกทำ" และ "ทำซ้ำ" ด้านบนเสมอ เพื่อให้กู้คืนการลากผิดพลาดได้ทันที

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามบังคับให้ผู้ใช้ต้องคลิกเลือกรายการก่อน แล้วกวาดสายตาขึ้นไปหาปุ่ม "เพิ่มย่อย" บน Header**: ทำให้คนไทย 40+ สับสน ไม่เข้าใจว่าทำไมปุ่มถึง disabled และหาจุดกดยาก
- **ห้ามทำ Tree View ที่ขาดปุ่มเพิ่มย่อย/แก้ไข/ลบ บนแถว**: การซ่อน Action ไว้หลังการคลิกแถวทำให้การทำงานช้าและไม่เป็นธรรมชาติ
- **ห้ามใช้สีระดับชั้นเหมือนกันทุกระดับ**: ทำให้แยกไม่ออกระหว่างหมวดหลักและหมวดย่อย

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
การมี Action buttons บนแถวโดยตรง (Direct Row Manipulation) ลด Cognitive Load ของผู้ใช้ โดยผู้ใช้เห็นเป้าหมายและจุดกดในตำแหน่งสายตาเดียวกัน (Fitts's Law) และการมี Drag & Drop ควบคู่กับปุ่ม Up/Down รองรับทั้งผู้ใช้เมาส์ จอสัมผัส และคีย์บอร์ด (WCAG 2.1 Accessibility)

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- `frontend/src/app/system-settings/product-group-tree-view.tsx` (ต้นแบบของระบบ)
- `frontend/src/app/system-settings/product-category-tree-view.tsx` (ยกระดับให้เหมือนกันครบถ้วน)

---

## 8.4 รวมการจัดหมวดสินค้าและบาร์โค้ดในหมวดไว้ในหน้าจอเดียว (Master-Detail Category & Barcodes Consolidation)

**แบบแผนใหม่ (New Standard Pattern)** — สำหรับ Master Data ที่มีโครงสร้างหมวดหมู่และรายการบาร์โค้ดผูกในหมวด (เช่น จัดหมวดสินค้า):
1. **ตำแหน่งเมนูในกลุ่มข้อมูลหลัก (Master Data Placement)**:
   - "จัดหมวดสินค้า" (`/productcategorygroupselectscreen`) จัดอยู่ในส่วน **"ข้อมูลหลัก" (Master Data)** ภายใต้กลุ่ม **"สินค้าและบาร์โค้ด" (Product Catalog)** โดยวางต่อท้าย "บาร์โค้ด" (`/productbarcode`) ทันที เพื่อให้สอดคล้องกับขั้นตอนการทำงานจริง (ต้องกำหนดรหัสสินค้าและบาร์โค้ดก่อน จึงจะนำบาร์โค้ดมาจัดเข้าหมวดหมู่สำหรับ POS/ขายหน้าร้านได้)
   - กลุ่ม "จัดกลุ่มสินค้า" ในส่วน "ค่าเริ่มต้น" (Defaults) จะคงเหลือเฉพาะ "หน่วยนับสินค้า" (`/productunit`) และ "กลุ่มสินค้า" (`/productgroup`)
2. **ผูกรายการระดับ "บาร์โค้ด" ไม่ใช่ระดับสินค้าทั่วไป (Barcode-Level Items)**:
   - ใช้ศัพท์ **"บาร์โค้ดในหมวด"** (Barcodes in Category) และปุ่ม **`+ เพิ่มบาร์โค้ด`** (Add Barcode) แทนคำว่า "สินค้า" เพราะในระบบขายหน้าร้าน/แคชเชียร์ หมวดสินค้าจะจัดกลุ่มหน่วยขายย่อยระดับบาร์โค้ด (SKU Barcodes)
   - ค้นหาผ่าน API `POST /api/product-barcode/list` (พร้อม `{ holdingcode, keyword, limit }`) เพื่อดึงบาร์โค้ดจริงจาก MongoDB พร้อมข้อมูลประกอบ: บาร์โค้ด, ชื่อสินค้า/บาร์โค้ด, รหัสสินค้า, หน่วยนับ, และราคาขาย
   - บันทึกลงฟิลด์ `codelist: [{ code: barcode, xorder: index, names: [...] }]` ของคอลเลกชัน `productcategories`
3. **รวมเป็นหน้าจอเดียว (Consolidated Master-Detail Layout)**:
   - หน้าจอเดียวจบที่ "จัดหมวดสินค้า" (`/productcategorygroupselectscreen`) ไม่แยกเมนู "สินค้าในหมวด" ออกไปเป็นเมนูโดดเดี่ยวที่ทำให้ผู้ใช้สับสน
   - **ฝั่งซ้าย**: `ProductCategoryTreeView` เลือกกลุ่ม 1–20 และผังหมวดหมู่แบบลากวาง (Drag & Drop Hierarchy)
     - **ตัดปุ่ม "เพิ่มหมวดย่อย" ออกทั้งหมด**: ทั้งปุ่มบน Header และปุ่ม `FolderPlus` บนแถวรายการ (Row Action) โดยให้ผู้ใช้สร้างหมวดด้วยปุ่ม "เพิ่มหมวดสินค้า" แล้วใช้วิธีลากวาง (Drag & Drop) จัดระดับเข้าเป็นลูกแทน เพื่อลดความรกรุงรังของหน้าจอ
   - **ฝั่งขวา**: เมื่อเลือกหมวดหมู่ แสดงแท็บ 2 แท็บ:
     - 🏷️ **บาร์โค้ดในหมวด** (Default Tab): ปุ่มหลักชัดเจน **`+ เพิ่มบาร์โค้ด`** ค้นหาบาร์โค้ด, เพิ่ม, ลากจัดลำดับ (Drag & Drop), ลบออก, พร้อมปุ่ม "แก้ไขข้อมูลหมวด" และ Badge นับจำนวนบาร์โค้ด (รวมถึงปุ่มเพิ่มในสถานะกล่องว่าง Empty State)
     - ⚙️ **ข้อมูลหมวดสินค้า**: รายละเอียดหมวดหมู่ (รหัส, ชื่อ, ระดับชั้น) พร้อมปุ่มแก้ไข/ลบ
     - **แบนเนอร์เชื่อมโยงในฟอร์มแก้ไข**: ในหน้าฟอร์มแก้ไขหมวด (`SettingFormDialog`) จะมีแบนเนอร์ด้านบนระบุจำนวนบาร์โค้ดที่ผูกอยู่ พร้อมปุ่ม **`+ เพิ่มบาร์โค้ด`** ที่กดแล้วสลับไปแท็บรายการบาร์โค้ดและเปิดหน้าต่างค้นหาเพิ่มบาร์โค้ดทันที
4. **Backward Compatibility & Menu Cleanup**:
   - หน้า `/productcategorylist` ทำ Next.js `redirect("/productcategorygroupselectscreen")` เพื่อรองรับ bookmark/url เดิม
   - ตัดเมนูย่อยซ้ำซ้อน `product-category-list` ออกจาก `MENU_SECTIONS` (คงเหลือเมนูทั้งหมด 223 เมนู)
5. **Dirty State Guards & Responsive Switching**:
   - ตรวจสอบ `categoryUnsavedChanges` ก่อนให้ผู้ใช้สลับรายการหมวดใน Tree ป้องกันรายการที่เพิ่งเพิ่ม/จัดลำดับสูญหาย
   - เมื่อผู้ใช้อยู่ในฟอร์มแก้ไขแล้วคลิกแถวหมวดเดิมใน Tree ให้ปิดฟอร์มและสลับกลับสู่แท็บ "บาร์โค้ดในหมวด" ทันที ไม่ติดค้างในหน้าจอแก้ไข

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามเอา "จัดหมวดสินค้า" ไปวางก่อนบาร์โค้ด หรือทิ้งไว้ใน Defaults**: ในทางปฏิบัติ ผู้ใช้ไม่สามารถจัดหมวดสินค้าได้หากยังไม่ได้สร้างบาร์โค้ด
- **ห้ามใช้คำว่า "สินค้า" กับรายการในหมวด**: การระบุว่า "เพิ่มสินค้า" ทำให้ผู้ใช้เข้าใจผิดว่าจะได้สินค้าหลักมาแทนที่จะเป็นบาร์โค้ดที่ขายจริง
- **ห้ามใส่ปุ่ม "เพิ่มหมวดย่อย" แยกต่างหาก**: เมื่อระบบมี Drag & Drop แล้ว การมีทั้งปุ่ม "เพิ่มหมวดหลัก" และ "เพิ่มหมวดย่อย" ทำให้ผู้ใช้ 40+ สับสนว่าต้องกดปุ่มไหน
- **ห้ามซ่อนปุ่มเพิ่มรายการเมื่อผู้ใช้อยู่ในหน้าจอแก้ไขหมวด**: ผู้ใช้มองว่าการแก้ไขหมวดรวมถึงการจัดการบาร์โค้ดในหมวดด้วย การไม่มีปุ่มเพิ่มในหน้าแก้ไขทำให้เข้าใจผิดว่าระบบไม่มีฟังก์ชันนี้

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
โครงสร้างข้อมูล MongoDB สำหรับหมวดสินค้าเก็บทั้ง Tree Metadata (`guidfixed`, `parentguid`, `groupnumber`, `names`) และ `codelist` (รายการบาร์โค้ดที่ผูกในหมวด) อยู่ใน collection เดียวกัน (`productcategories`) โดย `codelist.code` เก็บค่า barcode string การเชื่อมต่อด้วย `/api/product-barcode/list` และจัดวาง UI ในกลุ่ม Master Data สอดคล้องกับ Business Process และ Data Model จริง 100%

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- `frontend/src/app/system-settings/system-settings-screen.tsx` (`data-testid="product-category-detail-pane"`, `SettingFormDialog`)
- `frontend/src/app/system-settings/product-category-tree-view.tsx` (`ProductCategoryTreeView`)
- `frontend/src/app/system-settings/product-category-items-editor.tsx` (`ProductCategoryItemsEditor`)
- `frontend/src/app/[systemSetting]/page.tsx` (Route redirect)
- `frontend/src/lib/menu-data.ts` (`MENU_SECTIONS` -> `master` -> `products`)



## 2026-09-09 — ชื่อเมนูและป้ายรอพัฒนา

- **New Standard Pattern:** ชื่อไทยในเมนูและหน้าจอต้องตรงกัน; `menuText` ยึด `label.th` และซิงค์ TSV เสมอ ภาษาอื่นใช้ dictionary ตามเดิม ใช้ `<MenuPendingBadge route={item.route} language={language} backendLanguage={backendLanguage} />` ทุกช่องทางเปิดเมนู
- **Anti-pattern:** ห้ามให้ dictionary เก่าทับชื่อไทยจนธุรกรรมเปลี่ยนความหมาย ห้ามใช้คำว่าประวัติเมื่อหน้าจอแสดงสิทธิ์ปัจจุบัน ห้ามถือว่ามีเมนูแปลว่าพัฒนาหน้าจอแล้ว และห้ามซ่อน/ปิดเมนูรอพัฒนาเอง
- **Root Cause & Rationale:** คีย์ภาษาเดิมใช้ร่วมหลายบริบทและมีแคชต่างรุ่น ป้ายสถานะจึงต้องตรวจจาก custom dispatcher และ settings config จริง ไม่ใช่มี route หรือไอคอนอย่างเดียว ป้ายใช้ `bg-muted text-foreground border-border text-[0.9rem] leading-normal` เพื่ออ่านได้ในทั้งสองธีม
- **Reference Implementation:** `frontend/src/lib/menu-data.ts:485`, `frontend/src/lib/menu-screen-status.ts:13`, `frontend/src/app/menu/menu-pending-badge.tsx:6`, `frontend/e2e/menu-consistency.spec.ts:3`
- **วิธีตรวจ:** เทสต์เทียบชื่อ TSV/หน้าจอและแคชเก่า เทียบ registry กับ dispatcher; E2E บัญชี Demo กดเมนูจริง ตรวจ 4 viewport × light/dark โดยกดปุ่มธีม (รอ transition 350ms) ตรวจ hover/focus และไม่มี error ในหน้าเมนู ป้ายไม่เปลี่ยนสิทธิ์เดิม

---

## 8.5 แบบแผน Master-Detail 3 คอลัมน์ สำหรับคลังสินค้าและที่เก็บสินค้าจำนวนมาก (Warehouse & Large Location Hub — ตั้งโดยลุงจืด 2026-09-09)

**แบบแผนใหม่ (New Standard Pattern)**:
1. **โครงสร้าง 3 คอลัมน์ (Master-Detail Workspace)**:
   - **คอลัมน์ 1 (ซ้าย ~260px): คลังสินค้า (Warehouses)**: แสดงรายชื่อคลังสินค้าแบบการ์ดกระชับ พร้อม Badge สรุปจำนวนที่เก็บ (`ลูก X`), ปุ่ม `+ เพิ่มคลังสินค้า`, และปุ่มจัดการในแถว (`title="เพิ่มที่เก็บสินค้า"`, `"แก้ไขคลังสินค้า"`, `"ลบคลังสินค้า"`)
   - **คอลัมน์ 2 (กลาง 1fr กว้างสุด): ศูนย์จัดการที่เก็บสินค้าและที่วางสินค้า (Location Workspace)**:
     - ส่วนหัวแสดงรหัสและชื่อคลังที่เลือก พร้อมสถิติจำนวนที่เก็บสินค้าและที่วางสินค้าทั้งหมด
     - **ช่องค้นหาด่วนเฉพาะเจาะจง (Instant Location Search)**: ใช้ `Input` พร้อมไอคอนนำทาง `!pl-10` กรองที่เก็บสินค้าในคลังแบบ Real-time ทันทีที่พิมพ์
     - **ปุ่มคุมการแสดงผล**: `[ กางทั้งหมด ]` และ `[ ยุบทั้งหมด ]` สำหรับเปิด/ปิดดูที่วางสินค้า (Bins)
     - **ตารางรายการที่เก็บสินค้า Single-line Baseline Rhythm**: ลำดับ, Badge รหัสที่เก็บ (`text-xs font-mono font-bold`), ชื่อที่เก็บ (ตัวหนังสือไทย $\ge 0.9\text{rem}$), Badge จำนวนที่วางสินค้า (`📦 X ที่วาง`), และปุ่ม Action ประจำแถว (`title="เพิ่มที่วางสินค้า"`, `"แก้ไข"`, `"ลบที่เก็บสินค้า"`)
     - **ตารางย่อยที่วางสินค้า (Bins Sub-table)**: กางออกใต้แถวที่เก็บสินค้าอย่างเป็นระเบียบ แสดงรหัส, ชื่อ, บาร์โค้ด, พิกัด (Aisle/Rack/Level/Position) และปุ่มจัดการ (`title="แก้ไข"`, `"ลบที่วางสินค้า"`)
   - **คอลัมน์ 3 (ขวา ~380-400px): ฟอร์มจัดการข้อมูล (Active Form Panel)**:
     - แสดงหัวข้อและไอคอนระบุโหมดชัดเจน (เพิ่มคลัง, แก้ไขคลัง, เพิ่มที่เก็บ, แก้ไขที่เก็บ, เพิ่มที่วาง, แก้ไขที่วาง)
     - ปุ่ม "ยกเลิก" และ "บันทึก" ตรึงอยู่มุมขวาบน พร้อมรองรับ Keyboard Submit และ E2E Selectors
2. **การปรับสัดส่วนตามขนาดหน้าจอ (Responsive Grid Layout)**:
   - ใช้ `grid w-full grid-cols-1 lg:grid-cols-[240px_minmax(0,1fr)_340px] xl:grid-cols-[260px_minmax(0,1fr)_390px]` เพื่อให้จอระดับ iPad / Laptop (1024px+) และ Desktop (1280px+) แสดงผลได้ครบ 3 คอลัมน์โดยไม่ล้นและไม่มี horizontal scrollbar

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามรวม คลังสินค้า $\to$ ที่เก็บสินค้า $\to$ ที่วางสินค้า ทั้ง 3 ระดับไว้ใน Tree View แคบๆ ช่องเดียว**: เมื่อที่เก็บสินค้ามี 50–200+ รายการ ต้นไม้จะยาวเป็นพันแถว เลื่อนหายากมากและตาลาย
- **ห้ามปล่อยให้พื้นที่ด้านขวาเป็นแค่การ์ดฟอร์มว่างเปล่า**: ทำให้เสียพื้นที่หน้าจอแนวนอน 60% ไปโดยเปล่าประโยชน์ ควรใช้พื้นที่ส่วนกลางเป็นพื้นที่หลักแสดงและค้นหาที่เก็บสินค้า
- **ห้ามขาดช่องค้นหาเฉพาะเจาะจงสำหรับที่เก็บสินค้า**: ผู้ใช้ 40+ ไม่ควรต้องเลื่อนไล่หาที่เก็บสินค้าเป็นร้อยรายการด้วยสายตา

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
การแยกคลังสินค้าออกเป็น Selector ทางซ้าย และมอบพื้นที่หลักตรงกลางให้กับรายการที่เก็บสินค้า ทำให้ผู้ใช้สามารถสแกนรายการนับร้อย ค้นหาด้วยคีย์เวิร์ด และเจาะลึกดูที่วางสินค้า (Drill-down) ได้อย่างสะดวกรวดเร็วตามหลัก Cognitive Information Hierarchy โดยที่ฟอร์มด้านขวาพร้อมทำงานทันทีโดยไม่ต้องสลับหน้าจอไปมา และคงความเข้ากันได้ 100% กับ E2E Regression Contract

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- `frontend/src/app/system-settings/warehouse-tree-view.tsx`
- `frontend/e2e/product-warehouse-crud.spec.ts`

---

## 8.6 แบบแผน Editable Table Grid พร้อม Concurrent Batch Save สำหรับที่เก็บสินค้า (Location Editable Table Grid & Save All Pattern — ตั้งโดยลุงจืด 2026-09-09)

**แบบแผนใหม่ (New Standard Pattern)**:
1. **โครงสร้าง 2 ฝั่งพร้อม Draggable Resizable Splitter (ปรับความกว้างได้)**:
   - **ฝั่งซ้าย (ตัวแปร `--warehouse-sidebar-width` ปรับได้ 220px - 520px, default 280px)**: คลังสินค้า แสดงรายการคลังแบบการ์ดกระชับ พร้อม Badge สรุปจำนวนที่เก็บ (`ลูก X`), รหัสคลังแสดงใน Badge ครั้งเดียว (ห้ามพิมพ์รหัสซ้ำข้างชื่อ), ปุ่ม `+ เพิ่มคลังสินค้า`, และปุ่มจัดการในแถว (`title="เพิ่มที่เก็บสินค้า"`, `"แก้ไขคลังสินค้า"`, `"ลบคลังสินค้า"`)
   - **ตัวแบ่งปรับขนาดความกว้าง (Draggable Splitter)**:
     - ใช้ `role="separator"` พร้อม `aria-orientation="vertical"`, `aria-valuenow`, `aria-valuemin={220}`, `aria-valuemax={520}`
     - รองรับการลากด้วยเมาส์/สัมผัส (Pointer capture) พร้อมจดจำค่าลง `localStorage` (`bc_warehouse_sidebar_width`)
     - รองรับ Double-click เพื่อคืนค่าเริ่มต้น (280px)
     - รองรับการควบคุมผ่านคีย์บอร์ด (`ArrowLeft`, `ArrowRight`, `Home`, `End`)
     - มี Grip Handle ตรงกลาง พร้อมไอคอน `GripVertical` ที่ตอบสนองต่อการ hover และ active state
   - **ฝั่งขวา (1fr)**: ตารางแก้ไขข้อมูลที่เก็บสินค้าโดยตรง (Inline Editable Table Grid):
     - **Header Bar**: แสดงชื่อคลังสินค้าที่เลือก, Badge สถิติจำนวนที่เก็บสินค้า, Badge เตือน `ยังไม่ได้บันทึก` สีส้มเมื่อมีการแก้ไข, ปุ่ม `คืนค่าเดิม` (Reset), ปุ่ม `+ เพิ่มแถว` (Add Row), และปุ่ม CTA เด่น `บันทึกทั้งหมด` (Save All)
     - **Instant Filter**: ช่องค้นหารหัสหรือชื่อที่เก็บสินค้าแบบ Real-time พร้อมไอคอน `!pl-10`
     - **Sticky Table Header**: `thead` ติดตรึงด้านบนด้วย `sticky top-0 z-10 bg-muted/90 backdrop-blur border-b shadow-2xs` ไม่เลื่อนหายตามเนื้อหา
     - **Single-line Baseline Rhythm Table Rows**:
       - ลำดับที่ (`font-mono text-center` พร้อม Badge เล็ก `NEW` สีเขียว หรือ `MOD` สีส้ม เมื่อมีการเปลี่ยนแปลง)
       - รหัสที่เก็บ (`Input` ขนาดกะทัดรัด `h-8 font-mono font-bold uppercase placeholder="e.g. ZONE-A"`)
       - ชื่อที่เก็บสินค้าภาษาไทย (`Input` แถวเดียว `placeholder="ชื่อที่เก็บสินค้า (ไทย)"`)
       - ชื่อที่เก็บสินค้าภาษาอังกฤษ (`Input` แถวเดียว `placeholder="Location Name (EN)"` แยกเป็นอีกคอลัมน์เมื่อระบบเปิดใช้งานภาษาอังกฤษ)
       - ที่วางสินค้า (`📦 X ที่วาง` pill badge พร้อมปุ่มเปิดหน้าต่างจัดการ Bins และปุ่ม `+` เพิ่มที่วาง `title="เพิ่มที่วางสินค้า"`)
       - ปุ่มจัดการแถว (`title="แก้ไข"`, `title="ลบที่เก็บสินค้า"`)
       - สถานะแถวใช้เส้นขอบซ้ายเบาๆ (`border-l-2 border-l-emerald-500` สำหรับแถวใหม่, `border-l-2 border-l-amber-500` สำหรับแถวแก้ไข) ห้ามย้อมสีพื้นหลังทั้งแถวหนาเตอะ
     - **Pinned Bottom Footer (ตรึงขอบล่างเสมอ)**:
       - แยก Footer ออกมาอยู่นอก Container เลื่อนของตาราง (`shrink-0 border-t bg-secondary/5`)
       - ฝั่งซ้าย: แสดงยอดรวมแถวทั้งหมด (`รวม X รายการ`) และสรุปจำนวนแถวใหม่/แถวแก้ไข
       - ฝั่งขวา: ปุ่ม `+ เพิ่มแถว` และปุ่ม `บันทึกทั้งหมด` ตรึงอยู่กับที่เสมอ ไม่เลื่อนหลุดสายตา
2. **กลไก Concurrent Batch Save (เซฟพร้อมกันทีเดียว)**:
   - ผู้ใช้สามารถกด `+ เพิ่มแถว` ได้ต่อเนื่องโดยไม่ต้องรอบันทึกทีละรายการ (ระบบสร้าง `tempId`)
   - แก้ไขข้อมูลในช่องตารางได้ทันที ระบบจะ flag สถานะ `isNew`, `isModified`, `isDeleted`
   - เมื่อกด `บันทึกทั้งหมด`:
     - Validate ทุกแถวพร้อมกัน (รหัสและชื่อห้ามว่าง, รหัสห้ามซ้ำกันเอง)
     - รวมคำขอยิงแบบ Concurrency ด้วย `Promise.all` แยกตามประเภท (`DELETE`, `POST`, `PUT`)
     - แสดงสรุปผลสำเร็จหรือแจ้งเตือน error ชัดเจนผ่าน Banner ด้านบนตาราง (`CheckCircle2` / `AlertTriangle`)
3. **การจัดการระดับที่ 3 (Bins / ที่วางสินค้า) ด้วย Modal Dialog**:
   - ใช้ `.dialog-backdrop` ตามมาตรฐาน `globals.css` (ไม่ใช้ Radix หรือ external package ที่ไม่มีอยู่)
   - หน้าต่างย่อยมีตาราง Bins ภายใน พร้อมฟอร์มเพิ่ม/แก้ไข และปุ่มลบ เพื่อไม่ให้รบกวนหน้าตารางหลัก

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามใส่ปุ่ม Action Bar (เพิ่มแถว/บันทึก) ไว้ใต้ `<table>` ภายใน `overflow-auto`**: เมื่อมีแถวข้อมูลเยอะ ปลายตารางจะดันปุ่มหลุดจอไปอยู่ข้างล่าง ทำให้ผู้ใช้ต้อง scroll ลงไปลึกมากเพื่อกดบันทึก
- **ห้ามซ้อน Input หลายภาษา (TH / EN) ใน Cell เดียวกันแนวตั้ง**: ทำให้ความสูงของแถวนั้นโป่งเป็น 2-3 เท่า ทำลาย Single-line Baseline Rhythm ของตาราง ให้แยกเป็นคอลัมน์เฉพาะ
- **ห้ามย้อมสีพื้นหลังทั้งแถว (`bg-emerald-500/5` / `bg-amber-500/5`) หนาเตอะ**: ทำให้สีตีกับ hover state และดูเลอะเทอะ ให้ใช้ขอบซ้าย `border-l-2` และ badge ในเซลล์ลำดับแทน
- **ห้ามแสดงรหัสคลังซ้ำซ้อนในหัวการ์ด**: เช่น `[ 00000 ] 00000 - คลังหลัก` ให้แสดงเพียง `[ 00000 ] คลังหลัก`
- **ห้ามล็อคความกว้าง Side pane ให้ตายตัว**: หน้าจอผู้ใช้มีหลายขนาด ควรมี Resizable Splitter ให้ปรับขนาดและจดจำค่าลง localStorage เสมอ

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
- ช่วยให้ผู้ใช้กลุ่มคนไทย 40+ ใช้งานได้อย่างราบรื่นเหมือน Spreadsheet มองเห็นปุ่มดำเนินการหลักได้ตลอดเวลา (Pinned Footer) และอ่านข้อมูลได้ต่อเนื่องไม่สะดุดเส้นสายตา (Single-line Baseline Rhythm)
- การมี Splitter ช่วยให้ผู้ใช้ที่มีชื่อคลังยาว หรือต้องการเน้นดูตารางฝั่งขวาสามารถปรับแต่งพื้นที่ทำงานตามความสะดวกของตนเอง

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- `frontend/src/app/system-settings/warehouse-tree-view.tsx`
- `frontend/e2e/product-warehouse-crud.spec.ts`

---

## 8.7 มาตรฐาน Universal Resizable Splitter สำหรับทุกหน้าจอ Master-Detail / Split Workbench (Universal Resizable Splitter Standard — ตั้งโดยลุงจืด 2026-09-10)

**แบบแผนใหม่ (New Standard Pattern)**:
1. **คอมโพเนนต์กลาง `<ResizableSplitter>` (`frontend/src/components/ui/resizable-splitter.tsx`)**:
   - หน้าจอสองฝั่ง (Master-Detail, Tree-Detail, Catalog-Selection, List-Editor) ทุกจอในระบบ ต้องใช้คอมโพเนนต์กลางนี้เป็นตัวคั่น ห้ามเขียนเส้นคั่นดิบขึ้นมาเอง
   - รูปลักษณ์: Floating pill กึ่งกลางแนวตั้ง พร้อมไอคอน `GripVertical`, รองรับ Dark/Light mode, ขอบคมชัด, แสงเงาละมุน (hairline border + shadow-sm), hover glow ด้วย primary accent
   - มี Breakpoint อัตโนมัติ (`breakpoint="md" | "lg" | "xl"`) ซ่อนตัวแบ่งบนหน้าจอมือถือและแสดงเป็น Flex Column ซ้อนกัน เมื่อขยายถึง breakpoint จะสลับเป็น Flex Row / Grid 2 คอลัมน์พร้อมตัวแบ่ง
2. **ระบบการควบคุมและการเข้าถึง (Full Accessibility & Persistence)**:
   - **Pointer / Touch Drag**: ลากปรับความกว้างได้ทันทีด้วย pointer capture
   - **Keyboard Navigation**: ปุ่ม `ArrowLeft` / `ArrowRight` ขยับทีละ 2px หรือ 2%, `Home` ปรับไปแคบสุด, `End` ปรับไปกว้างสุด
   - **Double-click Reset**: ดับเบิ้ลคลิกเพื่อคืนค่าความกว้างเริ่มต้น (Default width/percent)
   - **Persistence**: บันทึกค่าลง `localStorage` ทันที ผู้ใช้ปิดแท็บหรือรีเฟรชค่าจะยังอยู่
   - **Zero SSR Layout Glitch**: ควบคุมขนาดผ่าน CSS Variable บน Parent Container เสมอ เช่น `style={{ ["--tree-split-basis" as any]: `${treeSplitPercent}%` }}` และใส่คลาส Tailwind เช่น `xl:w-[var(--tree-split-basis)]` เพื่อป้องกัน SSR Hydration mismatch
3. **การประยุกต์ใช้ 2 รูปแบบ**:
   - **Pixel-width Sidebar**: ใช้กับ Master List ที่ต้องการความกว้างคงที่ (เช่น 260px - 620px, default 340px)
   - **Percentage-basis Split**: ใช้กับหน้าจอ Workbench ที่ต้องแบ่งสัดส่วนเนื้อหาให้สมดุลทั้งจอ เช่น 60%/40% หรือ 65%/35%

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามใช้ตัวคั่นเส้นทื่อแบบเดิม**: เช่น `<div className="hidden w-1.5 shrink-0 cursor-col-resize bg-border/70 ... lg:block" />` เส้นเรียบไม่มี Grip handle คนมองไม่ออกว่าลากได้
- **ห้ามใช้ Grid แข็งที่ปรับขนาดไม่ได้**: เช่น `xl:grid-cols-[minmax(0,1fr)_420px]` หรือ `lg:grid-cols-[1fr_360px]` เมื่อผู้ใช้มีหน้าจอขนาดใหญ่/กว้าง จะเสียพื้นที่เปล่าประโยชน์
- **ห้ามใส่ `onMouseDown` อย่างเดียว**: ต้องรองรับ `PointerEvent` หรือ Touch เสมอ เพื่อให้ใช้บน iPad / Touch Screen ได้
- **ห้ามขาด ARIA Accessibility**: ต้องมี `role="separator"`, `aria-orientation="vertical"`, `aria-valuenow`, `aria-valuemin`, `aria-valuemax` และ `aria-label` ภาษาไทยอธิบายชัดเจน

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
- ผู้ใช้กลุ่มคนไทย 40+ ใช้งานอุปกรณ์หลากหลาย (iPad แนวนอน 1024px, โน้ตบุ๊ก 1366px, จอสำนักงาน Full HD 1920px, จอ Ultrawide) การบังคับคอลัมน์กว้างตายตัวทำให้บางจอแน่นเกินไป หรือบางจอกว้างจนฟอร์มยืด
- การมี Grip Vertical pill ที่ชัดเจนช่วยสร้าง Affordance ให้ผู้ใช้รู้ทันทีว่าสามารถลากปรับแต่งได้ตามใจชอบ และดับเบิ้ลคลิกคืนค่าเดิมได้โดยไม่ต้องกลัวทำพัง

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- คอมโพเนนต์กลาง: `frontend/src/components/ui/resizable-splitter.tsx`
- ชุดทดสอบความครอบคลุม: `frontend/src/components/ui/resizable-splitter.test.ts`
- หน้าจอที่นำไปใช้งานครบทุกจุด (11 หน้าจอ):
  1. `frontend/src/app/system-settings/warehouse-tree-view.tsx` (ผังคลังสินค้า)
  2. `frontend/src/app/system-settings/company-branch-tree-view.tsx` (โครงสร้างองค์กร บริษัท-สาขา)
  3. `frontend/src/app/system-settings/system-settings-screen.tsx` (`SettingMasterDetail`)
  4. `frontend/src/app/system-settings/system-settings-screen.tsx` (`BOM` สูตรการผลิต)
  5. `frontend/src/app/system-settings/system-settings-screen.tsx` (`ProductGroupTreeView` & `ProductSubgroupTreeView`)
  6. `frontend/src/app/system-settings/system-settings-screen.tsx` (`ProductCategoryTreeView`)
  7. `frontend/src/app/menu/product-screen.tsx` (ข้อมูลสินค้า)
  8. `frontend/src/app/menu/product-barcode-screen.tsx` (บาร์โค้ดสินค้า)
  9. `frontend/src/app/menu/product-set-screen.tsx` (สินค้าชุด)
  10. `frontend/src/app/menu/product-barcode-shelf-screen.tsx` (พิมพ์ป้ายสินค้า)
  11. `frontend/src/app/menu/manage-shortcuts-screen.tsx` (จัดการทางลัด)

---

## 8.8 มาตรฐานความสม่ำเสมอของโครงสร้าง Tree เมนูหลัก — ห้าม Bypass กลุ่มที่มี 1 รายการเป็น Flat Item (Uniform Menu Tree Structure Standard — ตั้งโดยลุงจืด 2026-09-10)

**แบบแผนใหม่ (New Standard Pattern)**:
1. **เรนเดอร์ทุกกลุ่มด้วย `MenuTreeGroup` เสมอ (Zero Bypass Policy)**:
   - ในทุก Section ที่มีหลายกลุ่ม (เช่น ข้อมูลหลัก `master`, งานประจำ `transactions`, รายงาน `reports`) ทุก `MenuGroup` ต้องเรนเดอร์ผ่าน `<MenuTreeGroup>` เสมอ ไม่ว่ากลุ่มนั้นจะมีสมาชิก 1 รายการหรือหลายรายการ
   - แถวหัวกลุ่มต้องมีโครงสร้างมาตรฐานเดียวกัน:
     - ไอคอนลูกศรพับ/ขยาย (`ChevronDown`) ขนาด `h-3.5 w-3.5` พร้อม transition หมุน 180 องศาเมื่อเปิด
     - ไอคอนประจำกลุ่ม (`MenuGroupIcon`) ขนาด 14px ในกรอบ `size-6 rounded-md bg-muted/60`
     - ชื่อกลุ่ม (`text-xs font-semibold text-foreground/90 truncate`)
     - ป้ายจำนวนรายการ (`h-5 min-w-5 rounded-full bg-muted/80 text-[11px] font-medium text-muted-foreground`)
     - ความสูงมาตรฐานแถว `min-h-8` พร้อมกรอบ `rounded-xl border border-border/70 bg-card/60 shadow-2xs`
2. **การพับ/ขยายที่สม่ำเสมอ (Uniform Collapsible Tree)**:
   - เริ่มต้นทุกกลุ่มจะพับอยู่ (Collapsed) เพื่อความเป็นระเบียบและประหยัดพื้นที่แนวตั้ง
   - เมื่อคลิกหัวกลุ่ม จะกางรายการย่อยออกมาพร้อมเส้นไกด์ไลน์นำสายตาฝั่งซ้าย (`ml-3.5 border-l border-border/50 pl-2`)
   - เมนูย่อยภายในจะเยื้องเข้าด้านในและมีเส้นเชื่อมแนวนอน (`-left-2 top-1/2 h-px w-2 bg-border/50`) ชี้ไปยังไอคอนของเมนูนั้น

**กับดัก / สิ่งที่ห้ามทำซ้ำ (Anti-pattern / Deprecated)**:
- **ห้ามใส่ Bypass พิเศษสำหรับกลุ่มที่มี 1 รายการเด็ดขาด**:
  ```tsx
  // ❌ กับดักเดิมที่ห้ามทำ:
  if (group.items.length === 1 && nodes.length === 1 && nodes[0].type === "item") {
    return <MenuTreeItemButton nested={false} ... />;
  }
  ```
- **ผลเสียร้ายแรงที่เกิดจาก Anti-pattern นี้**:
  1. **ความสูงผิดจังหวะ**: เมนูกลายเป็น `min-h-10` ขณะที่กลุ่มอื่นรอบตัวเป็น `min-h-8`
  2. **ขนาดตัวอักษรไม่เข้าพวก**: ตัวอักษรโป่งเป็น `text-sm font-medium` ขณะที่กลุ่มอื่นเป็น `text-xs font-semibold`
  3. **ขาด Chevron นำทาง**: ไม่มีลูกศรพับ/ขยาย ทำให้ดูเหมือนองค์ประกอบที่โหลดไม่เสร็จหรือหลุดกลุ่ม
  4. **มีปุ่มแท็บใหม่ (`ExternalLink ↗`) โผล่มาแทน Badge นับจำนวน**: ทำให้จังหวะสายตาด้านขวาของรายการเสียสมดุล
  5. **พฤติกรรมการคลิกไม่เหมือนเพื่อน**: คลิกกลุ่มอื่นจะเป็นการพับ/ขยาย แต่คลิกปุ่มนี้จะเปิดหน้าจอทันที ทำให้ผู้ใช้ 40+ สับสนและตกใจ

**เหตุผลทางเทคนิค (Root Cause & Rationale)**:
- สายตาของผู้ใช้กลุ่มคนไทย 40+ กวาดอ่านรายการเมนูจากบนลงล่างตามแบบแผนของ Tree Folder หากมีรายการใดรายการหนึ่งเปลี่ยนรูปร่างกลางคัน จะถูกตีความว่าเป็น Bug ของหน้าจอทันที
- ด้านการเข้าถึง (Accessibility / Screen Reader): การคง `role="treeitem"` ที่มี `aria-expanded` และ `aria-controls` จับคู่กับ `id` ของกลุ่มลูกอย่างสมบูรณ์ ช่วยให้ระบบนำทางด้วยคีย์บอร์ดและ Screen Reader ทำงานได้อย่างถูกต้อง ไม่หลุดลำดับชั้น

**ไฟล์และบรรทัดอ้างอิง (Reference Implementation)**:
- การเรนเดอร์กลุ่ม Tree กลาง: `frontend/src/app/menu/main-menu-screen.tsx:2120-2160`
- การกำหนดไอคอนกลุ่ม: `frontend/src/app/menu/main-menu-screen.tsx:2415-2460`
- Unit Test ป้องกันการ Bypass: `frontend/src/app/menu/main-menu-tree-consistency.test.ts`
- E2E Playwright ตรวจสอบทั้ง 9 กลุ่มของข้อมูลหลัก: `frontend/e2e/menu-tree-master.spec.ts`





## 8.9 เมนูอัปเกรด Champ และสถานะหน้าใช้งานจริง (2026-09-11)

- **New Standard Pattern:** เก็บ baseline id/route ก่อนจัดผัง; ใช้ 9 ระบบ PO → BILL → AP → AR → Cash → IC → FA → VAT → GL; ขั้นตอนอนุมัติ/ยกเลิกและเช็คแยกกลุ่มตาม workflow เดิม งานใหม่อยู่ในระบบที่เกี่ยวข้อง ชื่อเดิมค้นผ่าน `MenuLabel.aliases` เช่น `aliases: ["บันทึกใบเสนอซื้อสินค้า"]` โดยป้ายไทยและ permission id เดิมคงอยู่
- **Anti-pattern / Deprecated:** ห้ามยกจำนวนรายการใน XML มาอ้างเป็นความสามารถที่เสร็จ ห้ามเพิ่มเมนูซ้ำเพียงเพื่อให้ค้นชื่อเดิมเจอ ห้ามถือว่ามี config แล้ว API พร้อม และห้ามหน้า pending โชว์ Flutter/Next.js/route ให้ผู้ใช้อ่านแทนคำอธิบายงาน
- **Root Cause & Rationale:** source Champ มี 490 occurrences รวมรายงานย่อย/รายการซ้ำ; resource 59011 ใช้ทั้ง AP/AR ต้องแยกบริบท ตรวจ runtime พบทะเบียนเลขเครื่องมี config แต่ API 404 จึงมี pending override พร้อมใช้สถานะเดียวกันกับ dispatcher เพื่อไม่เปิดจอเสีย; ถอน override เมื่อเชื่อม API และตรวจ UAT ครบ
- **Reference Implementation:** `frontend/src/lib/menu-data.ts:37`, `frontend/src/lib/menu-data.ts:670`, `frontend/src/lib/menu-screen-status.ts:14`, `frontend/src/app/menu/main-menu-screen.tsx:2828`; fixture `frontend/src/lib/__fixtures__/menu-before-champ-upgrade.json`; ADR `docs/kms/decisions/2026-09-11-champ-upgrade-menu-workflows.md`
- **วิธีตรวจ:** Vitest 5 ไฟล์เฉพาะเมนู + tsc; Playwright `menu-champ-upgrade.spec.ts` ร่วมกับ `menu-tree-master`/`menu-consistency`; กด light/dark จริงทุกขนาด 1600/1280/1024/768, hover/focus, ไม่มี overflow/console error; ไม่มีการเขียนข้อมูลสำหรับการทดสอบเมนู
- คีย์ใหม่เพิ่ม TSV ครบ 13 ช่องพร้อมไอคอน route/group; ภาษาอื่นที่ยังไม่แปลใช้ English fallback และต้องแจ้งข้อจำกัด ไม่อ้างว่าผ่านตรวจเจ้าของภาษา


## 8.10 ชื่อหมวดเมนูหลักเป็นไทยล้วน (2026-09-11)

- **New Standard Pattern:** `MENU_SECTIONS[].title.th` ใช้ชื่อไทยโดยไม่มีอังกฤษในวงเล็บ เช่น `{ th: "ระบบบัญชีเจ้าหนี้", en: "Account Payable (AP)" }`; ใช้ชื่อจาก catalog เดียวกันทั้งเมนูซ้าย เมนูบน และข้อความบอกหมวด
- **Anti-pattern / Deprecated:** ห้ามเติม `(Account Payable : AP)` หรืออังกฤษ/ตัวย่อท้ายชื่อหมวดไทยกลับมา และห้ามตัดวงเล็บของทุกเมนูแบบเหมา เช่น `(50 ทวิ)` ซึ่งยังสื่อความหมายทางบัญชี
- **Root Cause & Rationale:** ลุงจืดให้เอาวงเล็บออกเพื่อให้หมวดไทยอ่านสั้นและไม่เบียดพื้นที่เมนู แก้ข้อมูล `title.th` โดยตรงแทนการซ่อนด้วย CSS หรือแปลงข้อความตอน render
- **Reference Implementation:** `frontend/src/lib/menu-data.ts:76` และ `frontend/src/lib/menu-data.test.ts:239`; ตรวจ exact Thai title ทั้ง 9 หมวดพร้อม baseline id/routes และ Playwright menu-tree-master/menu-consistency ทั้งสองธีม


## 8.11 เมนูบนเป็นค่าเริ่มต้นและจำรูปแบบที่เลือก (2026-09-11)

- **New Standard Pattern:** เริ่ม `menuLayout` เป็น `top`; ถ้า `bc_menu_layout_mode` เป็น `left` ให้คืนค่าเมนูซ้าย นอกนั้นใช้เมนูบน รอ `menuLayoutLoaded` ก่อนเขียน preference และตั้ง `sidebarHidden` ให้ตรงกับรูปแบบเมนู ปุ่มทั้งสองมี `aria-pressed` บอกสถานะที่เลือก
- **Anti-pattern / Deprecated:** ห้าม fallback เป็น `left` เมื่อยังไม่เคยเลือก; ห้าม effect เขียนค่าเริ่มต้นทับค่าที่บันทึกไว้ก่อน restore เสร็จ และห้ามคืนค่า left แต่ปล่อย sidebar ซ่อนค้างจาก top
- **Root Cause & Rationale:** ลุงจืดกำหนดเมนูบนเป็นค่าเริ่มต้น การอ่าน localStorage กับ effect บันทึกเกิดบน mount เดียวกัน ต้องมีสถานะโหลดเสร็จเพื่อรักษาตัวเลือกผู้ใช้ รวมถึงตอน React ตรวจ effects ซ้ำ
- **Reference Implementation:** `frontend/src/app/menu/main-menu-screen.tsx:414`, `frontend/src/app/menu/main-menu-screen.tsx:485`, `frontend/src/app/menu/main-menu-screen.tsx:524`; Playwright `frontend/e2e/menu-consistency.spec.ts` ตรวจ default top และ `frontend/e2e/menu-tree-master.spec.ts` ตรวจสลับ left → reload → sidebar ยังแสดง

## 8.12 หน้าบัญชีแยกประเภทและการเก็บฟอร์มในแท็บ (2026-09-11)

- **New Standard Pattern:** หน้าบัญชีใช้ `gl-workbench`, `panel`, `control`, `actionClass` จาก `frontend/src/app/gl/gl-common.tsx:9`; control/action `min-h-[2.6em] text-[0.95rem]`, panel `rounded-2xl border-border bg-card`, สีจาก theme tokens เท่านั้น ใช้ ResizableSplitter และตัวเลือกย่อ/ขยายบรรทัดตาม §8.7 จำนวนเงินเก็บ decimal strings และคำนวณด้วย BigInt ไม่อ่านค่าที่จัดรูปแบบกลับไปคำนวณ
- **New Standard Pattern — ฟอร์มค้าง:** ส่ง `window.dispatchEvent(new CustomEvent("bc-gl-dirty", { detail: { route, dirty } }))` พร้อม beforeunload; main menu ฟัง dirty ตาม route และใช้ GL หนึ่งแท็บต่อ route ปิดแท็บ/เปลี่ยน workspace/ออกจากระบบต้องยืนยันก่อนทิ้งข้อมูล สลับแท็บหรือค้นหาให้คง component mounted และซ่อนด้วย container ไม่ unmount
- **New Standard Pattern — โหลด lookup ใหม่:** ปุ่มโหลดใหม่เรียก `list.reload(); refs.reload();` คู่กัน โดยไม่ reset record/original ของฟอร์ม เพื่อให้แท็บที่ยัง mounted เลือกบัญชี/ปีที่เพิ่มจากอีกแท็บได้ ห้ามบังคับ remount เพื่อรีเฟรชข้อมูลอ้างอิง (`frontend/src/app/gl/gl-common.tsx:36`, `frontend/src/app/gl/gl-masters.tsx:65`); ทดสอบเพิ่มบัญชีจากอีกแท็บแล้วโหลดกลับมาเลือกโดยข้อมูลและสถานะยังไม่บันทึกคงอยู่
- **New Standard Pattern — รอรายงานหลัง Kafka:** `glRequest` รอซ้ำเฉพาะ GET ที่ตอบ HTTP 409 และ `errorcode: "GL_PROJECTION_PENDING"` โดยเว้น 100 → 250 → 500 → 1000 ms สูงสุดประมาณ 15 วินาที ยกเลิก backoff และ fetch ที่กำลัง retry เมื่อหมดเวลาหรือ caller abort; แสดงข้อความไทยให้รอแล้วโหลดใหม่เมื่อยังไม่พร้อม ฟอร์มและ lookup ใช้ component เดิมต่อโดยไม่ remount (`frontend/src/lib/general-ledger-api.ts:17`)
- **Anti-pattern — รอ projection:** ห้าม retry ทุก 409 เพราะ version/snapshot conflict ต้องให้ผู้ใช้ตรวจข้อมูลใหม่; ห้ามส่ง POST command ซ้ำหรือสร้าง requestid ใหม่จากกลไกรอ GET และห้ามตั้ง polling ไม่สิ้นสุด
- **Root Cause & Rationale — รอ projection:** MongoDB รับคำสั่งแล้วส่งผ่าน Kafka ก่อน PostgreSQL พร้อมอ่าน จึงอาจมีช่วงสั้นที่ GET หลังบันทึกยังอ่านรุ่นล่าสุดไม่ได้ การ retry ต้องผูก typed error เฉพาะนี้เพื่อไม่ซ่อนความผิดพลาดอื่นหรือส่งรายการบัญชีซ้ำ
- **วิธีตรวจ / Reference — รอ projection:** fake timers ใน `frontend/src/lib/general-ledger-api.test.ts:28` ตรวจ delayed readiness/backoff, timeout 15 วินาทีรวม fetch ระหว่าง retry, other 409 ไม่ retry, POST ไม่ retry/ไม่เปลี่ยน requestid และ caller abort; รัน focused Vitest และ tsc ก่อนส่งงาน
- **New Standard Pattern — คำแสดงผลรายงานไทย:** แปลค่าที่คอลัมน์ `accounttype/bookcode/status/direction/category` เฉพาะตอน render ด้วย mapping ตาม source; เติมชื่อยอด `currentearnings` ว่า “กำไรขาดทุนที่ยังไม่ปิด”, `cash` ว่า “เงินสดและเงินฝากธนาคาร” และ `unclassifiedlines` ว่า “บรรทัดที่ยังไม่ระบุประเภท” โดยแสดงจำนวนเป็นสตริงตามด้วย “รายการ” แถว `__current_earnings__` แสดงรหัสเป็น “—” และคงชื่ออธิบายบัญชีไว้
- **Anti-pattern — รายงานไทย:** ห้ามแสดง enum ภายในหรือ synthetic key เป็นข้อมูลบัญชีที่ผู้ใช้ต้องตีความ ห้ามใช้คำว่า “ยอดรวม” แทนยอดที่มีความหมายเฉพาะ และห้ามแก้ payload/จำนวนเงิน/API/CSV เพื่อเปลี่ยนภาษาแสดงผล
- **Root Cause & Rationale — รายงานไทย:** รายงาน generic รับ enum และยอดเสริมจาก backend; UAT ข้อมูลว่างไม่เห็นข้อความเหล่านี้ เมื่อมีข้อมูลจริงจึงต้องตรวจทุกคอลัมน์และป้ายยอด ภาษาไทยเลือกจาก key ของคอลัมน์เพื่อไม่แปลรหัสธุรกิจหรือคำอธิบายอิสระโดยบังเอิญ
- **วิธีตรวจ / Reference — รายงานไทย:** `frontend/src/app/gl/gl-reports.tsx:16` และ `gl-reports.test.ts:12` ตรวจ HTML ที่ render จริงสำหรับ enum, ป้ายยอด, จำนวนบรรทัด, synthetic key และจำนวนเงิน `0.30000001` พร้อมยืนยัน payload/CSV ไม่เปลี่ยน; ตรวจภาพรายงานที่มีข้อมูลจริง light/dark ด้วยปุ่ม theme และงบทดลอง 1600/1280/1024/768 ก่อนปิดงาน
- **New Standard Pattern — split pane:** ใช้ `<div>` เป็นกรอบรายการ/ฟอร์มของ SplitWorkbench และ `min-w-0` ที่ form/fieldset; ห้ามใช้ `<section>` แล้วหวังว่า `xl:w-[var(--gl-list-width)]` จะชนะกฎ CSS กลาง `main,section,form,fieldset,... { width:100%;max-width:100% }` ซึ่งอยู่นอก Tailwind layer (`frontend/src/app/globals.css:383`) ต้องตรวจ bounding rect ของสอง pane และภาพจริงว่าฟอร์มไม่ถูกบีบ ไม่อาศัยแค่ document overflow
- **Anti-pattern / Deprecated:** ห้ามเอา `topSearchResults ? results : workTabs` มาครอบฟอร์มที่มี state; ห้ามใช้ query loading หลังสลับภาษาเป็นเงื่อนไข unmount แท็บเดิม; ห้ามใช้ Number/parseFloat กับจำนวนเงิน; ห้ามจำกัดยอดรายงานเท่ากับ precision ของหนึ่งรายการ; ห้ามปลดป้ายรอพัฒนาเพียงเพราะมีหน้าเตรียมข้อมูล
- **Root Cause & Rationale:** การค้นหาเมนูแบบ ternary เดิมทำลายฟอร์มที่เปิดอยู่ และ GL หลายแท็บ route เดียวทำให้ event dirty ชนกัน จำนวนเงิน MongoDB Decimal128 อาจอ่านกลับเป็นเลขยกกำลัง แต่ UI/API ต้องคงสตริงทศนิยม plain และไม่ตัดทศนิยม ยอดส่งออกหลายหน้าต้องใช้ sequence เดียวกันเพื่อไม่ปะปนคนละ snapshot
- **Reference Implementation:** `frontend/src/app/menu/main-menu-screen.tsx:663`, `frontend/src/app/menu/main-menu-screen.tsx:2865`, `frontend/src/app/gl/gl-common.tsx:76`, `frontend/src/lib/general-ledger.ts:59`, `frontend/src/lib/general-ledger-api.ts:64`, `frontend/src/lib/menu-screen-status.ts:19`
- **วิธีตรวจ:** Vitest ชุด general-ledger/BFF/menu-screen-status และ tsc; Playwright `frontend/e2e/general-ledger-uat.spec.ts` ตรวจ dirty ค้นหา/สลับแท็บ/ปิดแท็บ, CRUD พร้อม query Mongo ทีละขั้น, light+dark กด theme toggle จริง ×1600/1280/1024/768 และ console/overflow; ผลจริงอ้าง `docs/kms/architecture/2026-09-11-general-ledger-v2.md` ห้ามอ้างว่าผ่านก่อนรันจบ


## 8.13 ผังบัญชีรองรับระดับ (Level 1–12) และการแสดงผลลำดับชั้น (2026-09-11)

- **New Standard Pattern:**
  - **ฟิลด์ระดับบัญชี (Level):** ฟิลด์ `level` ใน `Account` เป็นจำนวนเต็ม 1 ถึง 12 (ตามแบบแผน Champ `AccLevel` ใน `glfrmchartofaccountformview.cpp:257` และมาตรฐานบัญชีไทย) ค่าเริ่มต้นของบัญชีใหม่คือ 1 (`emptyAccount(): { level: 1 }`)
  - **ฟอร์มกรอกข้อมูล (`AccountFields`):** มีตัวเลือก "ระดับบัญชี (1–12)" ให้ผู้ใช้เลือก และเมื่อเลือกบัญชีแม่ (`parentaccountcode`) ระบบจะคำนวณระดับที่แนะนำอัตโนมัติ `(parent.level || 1) + 1` โดยจำกัดไม่เกิน 12
  - **การแสดงผลในตาราง (`GLMasters`):** เมื่อเป็นรายการผังบัญชี (`resource === "accounts"`) ให้เพิ่มคอลัมน์ "ระดับ" พร้อม Badge `ระดับ {level}` ใช้โทน `bg-primary/10 text-primary border-primary/20` (หนึ่งจอหนึ่งตระกูลสี) และในคอลัมน์ชื่อบัญชี ให้เยื้องตามระดับชั้น `style={{ paddingLeft: (level - 1) * 16 + "px" }}` พร้อมสัญลักษณ์กิ่งไม้ `└─` (`font-mono text-muted-foreground`) สำหรับ `level > 1`
  - **การแสดงผลใน Dropdown (`AccountSelect`):** ในตัวเลือก `<option>` ให้ใส่ non-breaking space `\u00A0\u00A0` ตามระดับชั้นและตามด้วย `└─ ` เพื่อให้นักบัญชีเห็นโครงสร้างผังบัญชีได้ทันทีในทุกจอที่มีการเลือกบัญชี
- **Anti-pattern / Deprecated:**
  - ห้าม hardcode สี badge เช่น `bg-blue-100 text-blue-800` ให้ derive จาก `--primary` เสมอ
  - ห้ามตัดทิ้งหรือละเลย `level` เมื่อสร้างหรือแก้ไขบัญชี
  - ห้ามอนุญาตให้บัญชีลูกมีระดับน้อยกว่าหรือเท่ากับบัญชีแม่
  - ห้ามใช้ `padding-left` กว้างเกินไปจนดันข้อความหลุดตาราง ให้ใช้หน่วย 16px ต่อระดับและคง `truncate` พร้อม `title` tooltip เสมอ
- **Root Cause & Rationale:**
  - นักบัญชีและผู้สอบบัญชีไทย (วัย 40+) ต้องการมองเห็นระดับการสรุปยอด (หมวดบัญชีคุม > บัญชีย่อย > บัญชีลงรายการ) ในพริบตา การมีระดับ 1–12 พร้อมการเยื้องภาพช่วยลดความสับสนและข้อผิดพลาดในการลงบัญชี
  - การเยื้องใน `<option>` ต้องใช้ `\u00A0` เนื่องจากเบราว์เซอร์ตัด space ปกติทิ้งใน native `<select>`
- **Reference Implementation:**
  - Backend: `backend/internal/generalledger/models.go:47`, `backend/internal/generalledger/mutations.go:158`, `backend/internal/generalledger/account_level_test.go`
  - Frontend: `frontend/src/lib/general-ledger.ts:9`, `frontend/src/app/gl/gl-masters.tsx:71`, `frontend/src/app/gl/gl-masters.tsx:112`, `frontend/src/app/gl/gl-common.tsx:30`
- **วิธีตรวจ:**
  - `go test -v ./internal/generalledger -run "TestAccountLevel"`

## 8.14 การป้องกันการลบผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวันเด็ดขาด (2026-09-11)

- **New Standard Pattern:**
  - **Backend Strict Guard (`mutations.go`):** เมื่อลบผังบัญชี (`cmd.Action == "delete"` ใน `accounts`) ระบบต้องตรวจว่ารหัสบัญชีดังกล่าวถูกอ้างอิงใน `lines.accountcode` ของ `gl_journals` หรือไม่ โดยห้ามกรอง `isdeleted: false` ออก (แม้สมุดรายวันนั้นจะถูกยกเลิก/void แล้ว ประวัติการทำรายการก็ยังคงอยู่เพื่อการตรวจสอบ Audit Trail) หากพบข้อมูลอ้างอิง ต้องปฏิเสธการลบด้วยข้อความภาษาไทยที่ชัดเจน:
    `"บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ"`
  - **Frontend Warning Notice (`AccountFields`):** ในฟอร์มแก้ไขผังบัญชีเดิม (`value.id`) ให้แสดง `<Notice text="ผังบัญชีที่มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบเด็ดขาด หากไม่ต้องการใช้งานให้ปิดใช้งานแทน" />` เพื่อแจ้งผู้ใช้ล่วงหน้า
  - **Frontend Confirmation Dialog (`GLMasters`):** ในกล่องยืนยันการลบ (`runAction("delete")`) ให้ระบุคำอธิบายเพิ่มเติม (`details`) ชี้แจงชัดเจนว่าหากมีรายการเคลื่อนไหวหรืออ้างอิงจากสมุดรายวัน ระบบจะไม่อนุญาตให้ลบเด็ดขาด
- **Anti-pattern / Deprecated:**
  - ห้ามใส่เงื่อนไข `isdeleted: false` ในการค้นหาการอ้างอิงของสมุดรายวัน เพราะหากมีผู้ใช้ยกเลิกรายวันหรือลบฉบับร่าง บัญชีที่เคยผ่านรายการจะไม่ถูกป้องกัน ทำให้เกิด dangling references ในประวัติศาสตร์บัญชี
  - ห้ามลบผังบัญชีจริงเมื่อมีข้อมูลอ้างอิง ให้ปิดการใช้งาน (`isactive: false`) แทนเสมอ
  - ห้ามแสดงข้อความ technical error หรือรหัสข้อผิดพลาดภาษาอังกฤษ ให้ใช้ภาษาไทยที่แนะนำทางออกแก่ผู้ใช้ทันที
- **Root Cause & Rationale:**
  - ตามหลักมาตรฐานการบัญชีและข้อกำหนดกรมสรรพากร ผังบัญชีที่มีรายการเคลื่อนไหวทางบัญชีหรือถูกอ้างอิงในสมุดรายวัน ห้ามถูกลบทิ้งจากระบบบัญชีโดยเด็ดขาด การอนุญาตให้ลบจะทำให้งบการเงินย้อนหลัง รายงานแยกประเภท และรายงานการตรวจสอบ (Audit Trail) เสียหาย
- **Reference Implementation:**
  - Backend: `backend/internal/generalledger/mutations.go:79`, `backend/internal/generalledger/account_crud_integration_test.go:217`
  - Frontend: `frontend/src/app/gl/gl-masters.tsx:55`, `frontend/src/app/gl/gl-masters.tsx:124`
- **วิธีตรวจ:**
  - Integration Test: `go test -v -tags=integration ./internal/generalledger -run "TestLedgerMongoPostgresAccountLevelCRUD"` ตรวจว่าเมื่อสร้างบัญชีและมีสมุดรายวันอ้างอิง การลบบัญชีต้องล้มเหลวด้วยข้อความระบุชัดเจน และบัญชียังคงอยู่ครบทั้งใน MongoDB และ PostgreSQL แม้หลัง void รายวันแล้ว

## 8.15 มาตรฐานการออกแบบงบการเงินและปรับแต่งฟอนต์ (Financial Statement Designer Standard — 2026-09-11)

- **New Standard Pattern:**
  - **การปรับแต่งรูปแบบตัวอักษรระดับงบ (Global Font & Style Control):**
    - ตัวเลือกแบบอักษร (Font Family): มีฟอนต์มาตรฐานภาษาไทยที่อ่านง่ายสำหรับผู้ใช้ 40+ เช่น Sarabun (สารบรรณ), Prompt (พร้อม), Kanit (คณิต), Noto Sans Thai, Inter และ Monospace
    - มีระบบโหลดฟอนต์ Google Fonts อัตโนมัติ (`ensureFontLoaded`) เมื่อผู้ใช้เลือกใช้งาน
    - ตัวปรับขนาดอักษร (Base Font Size: 13–18px), Line Height (1.4–2.0) และความเข้มของเส้นตาราง (Border Color / Weight)
  - **การจัดโครงสร้างแถวงบการเงิน (Statement Rows Structure):**
    - ประเภทแถว: `header` (หัวข้อ), `account` (ผูกบัญชี), `total` (ยอดรวม/สูตรคำนวณ), `text` (ข้อความประกอบ), `blank` (เว้นวรรค)
    - สไตล์เฉพาะแถว: ตัวหนา (`bold`), ตัวเอียง (`italic`), ขีดเส้นใต้เดี่ยว/คู่ตามหลักบัญชี (`single` / `double`), ระดับการเยื้อง (Indent 0–5 ระดับ)
    - การกำหนดเลขแถว (`rowno` เช่น 10, 20, 30) เพื่อใช้เป็นตัวแปรอ้างอิงในสูตรคำนวณ
  - **เครื่องคำนวณยอดและสูตรคณิตศาสตร์ (Live Calculation Engine):**
    - รองรับการอ้างอิงตัวแปรแถว เช่น `R10 + R20 - R30`
    - รองรับฟังก์ชันช่วงรวม เช่น `SUM(R10:R50)`
    - ใช้ Recursive Descent Parser พร้อมทศนิยมแบบ Fixed-Point BigInt (สเกล $10^8$) เพื่อป้องกัน floating point error ในยอดเงิน
    - มีระบบ Circular Dependency Detection (`visited` set) ป้องกันสูตรวนซ้ำไม่รู้จบ
  - **โหมดสลับการทำงาน (Designer & Live Preview Tabs):**
    - แท็บ "ออกแบบผังงบ" สำหรับจัดลำดับแถว (เลื่อนขึ้น/ลง, เพิ่มแถว, กำหนดสไตล์, ผูกบัญชี)
    - แท็บ "พรีวิวผลลัพธ์จริง" ดึงยอดจากงบทดลองตามงวด/ปีบัญชีที่เลือก มาคำนวณและแสดงผลในแบบอักษรและสไตล์ที่กำหนด พร้อมปุ่ม "พิมพ์งบการเงิน" (Print CSS) และ "ส่งออก CSV"
  - **แม่แบบตั้งต้น (Starter Templates):** มีปุ่ม "โหลดแม่แบบตั้งต้น" (งบดุลตาม DBD, งบกำไรขาดทุนจำแนกตามหน้าที่, งบต้นทุนการผลิต, งบกระแสเงินสดวิธีทางอ้อม) ให้นักบัญชีเริ่มใช้งานได้ทันทีโดยไม่ต้องสร้างจากศูนย์
- **Anti-pattern / Deprecated:**
  - ห้ามใช้ `eval()` หรือ JavaScript function evaluation ในการคำนวณสูตรยอดเงิน เพราะเสี่ยงต่อช่องโหว่ความปลอดภัยด้าน Code Injection และ Floating Point precision
  - ห้ามใช้สี hardcode นอก palette ระบบ ให้ใช้ตัวแปร CSS และ `--primary`
  - ห้ามลืมกลไก Dirty Guard ในการสลับแท็บหรือเปลี่ยนแม่แบบ
  - ห้ามบังคับผู้ใช้พิมพ์รหัสบัญชีด้วยมือ ให้มี Account Picker Dialog ช่วยค้นหาและเลือกผังบัญชี
- **Root Cause & Rationale:**
  - ธุรกิจแต่ละแห่งมีรูปแบบการรายงานงบการเงินที่แตกต่างกันตามประเภทธุรกิจและข้อกำหนดของผู้บริหาร/ผู้สอบบัญชี การมี Financial Statement Designer ที่ผู้ใช้สร้างเองได้ไม่จำกัดและปรับแบบอักษรได้อิสระ ช่วยให้ซอฟต์แวร์ยืดหยุ่นสูงเทียบชั้นกับระบบ ERP ชั้นนำ
- **Reference Implementation:**
  - Backend: `backend/internal/generalledger/models.go:76`, `backend/internal/generalledger/mutations.go:121`, `backend/internal/generalledger/statement_template_test.go`
  - Frontend: `frontend/src/lib/general-ledger.ts:184`, `frontend/src/app/gl/gl-statement-designer.tsx`, `frontend/src/app/gl/general-ledger-screen.tsx:37`
- **วิธีตรวจ:**
  - Frontend Unit Tests: `npx vitest run src/lib/general-ledger.test.ts`
  - TypeScript Typecheck: `npx tsc --noEmit`
  - Go Unit Tests: `go test -v ./internal/generalledger -run TestStatementTemplateModel`

## 8.16 มาตรฐานระบบค้นหาผังบัญชีแบบเต็มจอ (Full-Screen Chart of Accounts Search Dialog Standard — 2026-09-12)

- **New Standard Pattern:**
  - **Full-Screen Account Search Dialog (`AccountSearchDialog`):**
    - เมื่อผู้ใช้คลิกเลือกผังบัญชี หรือกดปุ่มค้นหา `[ 🔍 ]` หรือกดปุ่มลัด `F2` / `Space` บนช่องเลือกบัญชี (`AccountSelect`) ระบบจะเปิด Dialog ค้นหาผังบัญชีแบบเต็มจอทันที
    - มีปุ่มสลับ "เต็มจอ (Fullscreen)" และ "หน้าต่างใหญ่ (Large Modal)" เพื่อความสบายตาในการทำงาน
    - ช่องค้นหาหลักขนาดใหญ่ (`text-lg`) พร้อม Auto-focus ทันทีเมื่อเปิด ค้นหาได้ทั้งรหัสบัญชี, ชื่อบัญชีไทย/อังกฤษ, และหมวดหมู่
    - แท็บหมวดบัญชี 5 หมวดหลัก (สินทรัพย์, หนี้สิน, ส่วนของเจ้าของ, รายได้, ค่าใช้จ่าย) พร้อม Badge นับจำนวนบัญชีในแต่ละหมวด
    - ตารางแสดงผลที่อ่านง่ายสำหรับผู้ใช้คนไทยอายุ 40+: แสดงรหัสบัญชีชัดเจนใน font-mono, ชื่อบัญชีพร้อมการเยื้องตาม Level `(level - 1) * 18px` และสัญลักษณ์ `└─ `, Badge หมวดบัญชีตามสีที่ถูกต้อง, ด้านปกติ (เดบิต/เครดิต), และสถานะ (ลงรายการได้ / บัญชีคุม)
    - รองรับการนำทางด้วยแป้นพิมพ์ครบวงจร: `↑` / `↓` เลื่อนแถว, `Enter` เลือกบัญชี, `Esc` ปิดหน้าต่าง, และดับเบิ้ลคลิกแถวเพื่อเลือกทันที
    - รองรับทั้ง Single-Select (สำหรับสมุดรายวัน, รายงาน, ผังบัญชี) และ Multi-Select (สำหรับแถวในงบการเงิน)
  - **การทำงานร่วมกับ Playwright E2E และความเข้ากันได้ย้อนหลัง (Zero Regression):**
    - `AccountSelect` ยังคงคง `<select aria-label={label}>` ไว้ใน DOM เพื่อให้คำสั่ง `locator.selectOption(...)` ใน Playwright E2E Tests (เช่น `general-ledger-uat.spec.ts`) ทำงานได้ราบรื่น 100%
    - ดักจับ `onMouseDown` และ `onKeyDown` เพื่อเปิด Full-Screen Dialog เมื่อผู้ใช้จริงกดคลิกหรือกดแป้นพิมพ์
- **Anti-pattern / Deprecated:**
  - ห้ามปล่อยให้ผู้ใช้ต้องเลื่อนหาผังบัญชีใน native `<select>` dropdown ที่แคบและไม่มีช่องค้นหา
  - ห้ามลบ `<select aria-label={label}>` ออกจนทำให้ E2E test `.selectOption()` ล้มเหลว
  - ห้ามใช้สี hardcoded นอกเหนือจากตัวแปร CSS ของระบบ
  - ห้ามลืมการจัดการโฟกัสและปุ่ม ESC สำหรับการปิดหน้าต่าง
- **Root Cause & Rationale:**
  - ผังบัญชีขององค์กรจริงมีจำนวนหลายสิบถึงหลายร้อยรายการ การใช้ native dropdown แบบดั้งเดิมทำให้ค้นหาชื่อภาษาไทยได้ยากมากและมองไม่เห็นโครงสร้างระดับชั้น (Level) หรือด้านบัญชี การเปิดเป็น Full-Screen Search Dialog ช่วยให้นักบัญชีค้นหาและเลือกบัญชีได้อย่างรวดเร็ว แม่นยำ และสบายตา
- **Reference Implementation:**
  - Dialog Component: `frontend/src/app/gl/account-search-dialog.tsx`
  - Control Integration: `frontend/src/app/gl/gl-common.tsx:27`
  - Statement Designer Integration: `frontend/src/app/gl/gl-statement-designer.tsx:949`
  - Unit Tests: `frontend/src/app/gl/account-search-dialog.test.ts`
- **วิธีตรวจ:**
  - `npx vitest run src/app/gl/account-search-dialog.test.ts`
  - `npx tsc --noEmit`
  - `npm test`

## 8.17 มาตรฐานเครื่องมือช่วยตรวจสอบและคัดลอก DOM (DOM Inspector & Copy DOM Standard — 2026-09-12)

- **New Standard Pattern:**
  - **DOM Inspector Widget (`DevDomInspector`):**
    - แสดงปุ่มลอยอยู่ที่มุมล่างซ้าย (`fixed bottom-4 left-4 z-[999999]`) มีไอคอน `Code2`, ข้อความ "Copy DOM" และปุ่มลัด `Alt+คลิก`
    - ทำงานได้ทั้งบน Localhost (Dev) และ Production (`account.bcaicloud.com`) เพื่อสนับสนุนการ Pair Programming, Review, ตรวจสอบโครงสร้าง และทำ Visual Feedback
    - โหมดการทำงาน 2 แบบ:
      1. **Toggle Mode:** คลิกปุ่มเพื่อเปิด/ปิดโหมดตรวจจับอย่างต่อเนื่อง (ปุ่มแสดงสถานะเปิด พร้อมปุ่มปิด X และคำแนะนำ Toast)
      2. **Shortcut Mode:** กดปุ่ม `Alt` บนแป้นพิมพ์ค้างไว้ จะเข้าสู่โหมดพร้อมคลิกคัดลอกทันที ปล่อย `Alt` จะกลับสู่โหมดปกติ
    - **Visual Overlay & Feedback:**
      - เมื่อชี้เมาส์ที่องค์ประกอบใด จะมีกรอบไฮไลต์สีน้ำเงิน (`border: 2px dashed #3b82f6`) พร้อม Tooltip แสดง Tag Name, ID และ Classes
      - เมื่อคลิก จะสกัด `target.outerHTML` และคัดลอกลง Clipboard ทันทีผ่าน `copyTextSafely` พร้อมเปลี่ยนสีกรอบเป็นสีเขียว (`#22c55e`) และแจ้งเตือน Toast สำเร็จ
    - **Clipboard Fallback Mechanism:**
      - เรียกใช้ `navigator.clipboard.writeText` ก่อน
      - หากล้มเหลว (เช่น ใน iframe หรือ non-secure context) ให้ fallback ไปยัง `<textarea>` + `document.execCommand('copy')` อัตโนมัติ
- **Anti-pattern / Deprecated:**
  - ห้ามใส่เงื่อนไข `process.env.NODE_ENV !== "development"` บล็อกไม่ให้วิดเจ็ตทำงานบน Production ตามคำสั่งลุงจืด 2026-09-12
  - ห้ามให้ click event ของโหมดคัดลอก DOM ไปกระตุ้น action จริงของหน้าจอ (ต้องใช้ `e.preventDefault()`, `e.stopPropagation()`, `e.stopImmediatePropagation()` ใน capture phase `useCapture = true`)
- **Root Cause & Rationale:**
  - ลุงจืดและทีมงานต้องการความสะดวกในการชี้องค์ประกอบบนหน้าจอจริงทั้งบนเครื่อง Localhost และ Live Production เพื่อคัดลอกโค้ด DOM ไปสั่งปรับแต่งหรือส่งต่อให้ AI วิเคราะห์แก้ไขได้อย่างรวดเร็ว แม่นยำ ไม่ต้องเปิด DevTools (F12) หาแถวเอง
- **Reference Implementation:**
  - Component: `frontend/src/components/dev-dom-inspector.tsx`
  - Layout: `frontend/src/app/layout.tsx:94`
  - Unit Tests: `frontend/src/components/dev-dom-inspector.test.ts`
- **วิธีตรวจ:**
  - `npx vitest run src/components/dev-dom-inspector.test.ts`

## 8.18 มาตรฐานแถบค้นหาหลัก (Baseline Search Toolbar Standard — Auto Debounce 2s & Clean Icon — 2026-09-12)

- **New Standard Pattern:**
  - **Auto Search พร้อม Debounce 2 วินาที (`useDebouncedSearch`):**
    - เมื่อผู้ใช้พิมพ์ข้อความในช่องค้นหา (`SearchInput`) ระบบจะหน่วงเวลา 2 วินาที (2000ms) หลังจากการพิมพ์ตัวอักษรสุดท้าย หากผู้ใช้ไม่กดปุ่ม "ค้นหา" หรือ Enter ระบบจะเริ่มค้นหาให้อัตโนมัติ (`onSearch(query)`)
    - หากผู้ใช้กดปุ่ม "ค้นหา" หรือกดปุ่ม Enter ก่อนครบ 2 วินาที ระบบจะยกเลิกตัวนับเวลาและดำเนินการค้นหาทันที (`searchNow()`)
  - **ปุ่มและไอคอน Clean ล้างคำค้นหา (`Clean Icon`):**
    - ในช่องค้นหา เมื่อมีข้อความ (`value.length > 0`) จะมีปุ่มไอคอน Clean (`X`) ปรากฏที่ขอบขวาของช่องค้นหา (`absolute right-2`)
    - ช่อง input มีการหลบระยะขวาด้วย `pr-9` เพื่อป้องกันข้อความทับซ้อนกับปุ่ม
    - เมื่อคลิกปุ่ม Clean: ระบบจะยกเลิกตัวนับเวลา debounce, ล้างข้อความในช่องค้นหาเป็นค่าว่างทันที, สั่งค้นหาด้วยค่าว่างทันที (`onSearch("")`), และดึงโฟกัส (`focus()`) กลับมาที่ช่อง input
    - รองรับการกดปุ่ม `Escape` ขณะที่ช่องค้นหามีข้อความเพื่อทำหน้าที่เดียวกับปุ่ม Clean
  - **โครงสร้าง Baseline Form Toolbar:**
    - คอมโพเนนต์มาตรฐาน: `SearchInput` ใช้งานคู่กับ `useDebouncedSearch`
    - ช่องค้นหา: `<SearchInput value={query} onChange={setQuery} onClear={clear} onSearch={searchNow} ... />`
    - ปุ่มค้นหาหลัก: `<Button type="submit"><Search className="size-4 mr-1.5" />ค้นหา</Button>`
- **Anti-pattern / Deprecated:**
  - ห้ามใช้ native `<input>` เปล่าที่ไม่มีปุ่ม Clean (ล้างคำค้นหา)
  - ห้ามปล่อยให้ผู้ใช้ต้องกดปุ่ม "ค้นหา" หรือ Enter ทุกครั้งโดยไม่มีระบบ Auto Search
  - ห้ามยิง request ค้นหาทุกๆ ตัวอักษรที่พิมพ์โดยไม่มี Debounce (ทำให้ server รับภาระหนักและข้อมูลกระตุก)
  - ห้ามให้การกดปุ่ม "ค้นหา" ซ้ำซ้อนกับตัวนับเวลาที่ยังค้างอยู่ (ต้องเคลียร์ timer ทันทีที่สั่ง searchNow)
- **Root Cause & Rationale:**
  - ผู้ใช้งานคนไทย 40+ มักพิมพ์ข้อความค้นหาทีละตัวและรอให้ผลลัพธ์ปรากฏ การมี Auto Search 2 วินาทีช่วยให้ไม่ต้องเอื้อมมือไปกดปุ่มค้นหาทุกครั้ง ขณะเดียวกันปุ่ม Clean ช่วยให้ล้างข้อความได้ทันทีโดยไม่ต้องกด Backspace ทีละตัวอักษร
- **Reference Implementation:**
  - Hook & Component: `frontend/src/app/gl/gl-common.tsx:SearchInput`, `frontend/src/app/gl/gl-common.tsx:useDebouncedSearch`
  - Usage in Masters: `frontend/src/app/gl/gl-masters.tsx:75`
  - Usage in Journals: `frontend/src/app/gl/gl-journals.tsx:69`
  - Usage in Statement Designer: `frontend/src/app/gl/gl-statement-designer.tsx:321`
  - Unit Tests: `frontend/src/app/gl/search-input.test.ts`
- **วิธีตรวจ:**
  - `npx vitest run src/app/gl/search-input.test.ts`
  - `npx tsc --noEmit`
  - `npm test`

## 8.19 มาตรฐานช่องกรอกตัวเลขและการแสดงผลจำนวนเงิน (Formatted Numeric Input Standard — Decimal, Thousands Comma, Right-Aligned & Clean Edit Mode — 2026-09-12)

- **New Standard Pattern:**
  - **การแสดงผลเมื่ออยู่นิ่ง / ไม่ได้โฟกัส (Idle / Blur Mode):**
    - ตัวเลขทุกจำนวนในระบบต้องจัดรูปแบบให้มี **เครื่องหมายคั่นหลักพัน (Thousands Comma)** และ **ทศนิยมตามที่กำหนด (Fixed Scale Decimals)** เสมอ (ค่าเริ่มต้น 2 ตำแหน่ง เช่น `70,000.00`, `1,234,567.89`)
    - ต้องจัดข้อความ **ชิดขวาเสมอ (`text-right`)** พร้อมระบุ `tabular-nums` เพื่อให้ตัวเลขในแต่ละหลักมีความกว้างเท่ากัน จัดเรียงตรงตามจุดทศนิยมอย่างเป็นระเบียบ
  - **โหมดแก้ไขเมื่อโฟกัส (Focus / Edit Mode):**
    - เมื่อผู้ใช้คลิกหรือแท็บเข้ามาในช่องกรอก ให้สลับการแสดงผลเป็น **ข้อความธรรมดาที่ไม่มีเครื่องหมายจุลภาค (Plain Text without Commas)** เช่น `70000.00` หรือ `70000` ทันที
    - ทำการเลือกข้อความทั้งหมด (`event.target.select()`) ทันทีที่โฟกัส พร้อมดัก `onMouseUp` (`e.preventDefault()`) เพื่อไม่ให้การปล่อยเมาส์ล้างการเลือกข้อความ ทำให้ผู้ใช้สามารถพิมพ์ตัวเลขใหม่แทนที่ได้ทันที
    - ขณะที่ผู้ใช้กำลังพิมพ์ (Active Typing) **ห้ามใส่เครื่องหมายจุลภาค (Comma) แทรกระหว่างการพิมพ์เด็ดขาด** เพื่อป้องกันปัญหาเคอร์เซอร์กระโดดหรือเลื่อนตำแหน่งผิดพลาด
  - **การปรับข้อมูลเมื่อหลุดโฟกัส (Blur / Commit Mode):**
    - เมื่อหลุดโฟกัส (`onBlur`) หรือกดปุ่ม Enter (`onKeyDown` Enter $\to$ `blur()`): ให้ตัดช่องว่าง, ปรับทศนิยมให้ครบตาม Scale ด้วยการปัดเศษครึ่งหนึ่งขึ้น (Half-up Rounding) อย่างแม่นยำ, จัดรูปแบบด้วย Thousands Comma, และส่งค่า (Emit) ตัวเลขที่สะอาด (Clean Numeric String/Number) กลับไปยัง State ของระบบ
  - **คอมโพเนนต์มาตรฐาน:**
    - สำหรับระบบบัญชีแยกประเภท / การเงิน (GL & Financial): ใช้ `AmountInput` จาก `frontend/src/app/gl/gl-common.tsx`
      ```tsx
      <AmountInput
        value={value.amount}
        onChange={(amount) => set({ amount })}
        scale={2}
        required
      />
      ```
    - สำหรับระบบสินค้า / บาร์โค้ด / ทั่วไป (Product & Inventory): ใช้ `NumericInput` จาก `frontend/src/components/ui/numeric-input.tsx`
- **Anti-pattern / Deprecated:**
  - ห้ามใช้ `<input type="number">` เมื่อต้องการแสดง Thousands Comma เพราะมาตรฐาน HTML จะบล็อกเครื่องหมายจุลภาค ทำให้เกิด validation error หรือค่าว่าง
  - ห้ามจัดรูปแบบแทรกเครื่องหมายจุลภาคแบบ Real-time ขณะที่ผู้ใช้กำลังพิมพ์ เพราะจะทำให้ตำแหน่งเคอร์เซอร์เลื่อนและพิมพ์ลำบากมาก
  - ห้ามปล่อยให้ช่องกรอกตัวเลขการเงินจัดชิดซ้าย (`text-left`) หรือใช้ฟอนต์ตัวเลขที่ไม่มี `tabular-nums`
  - ห้ามใช้ `parseFloat()` หรือ `Number()` กับจำนวนเงินขนาดใหญ่มากจนสูญเสียความแม่นยำ (Precision Loss)
- **Root Cause & Rationale:**
  - ลุงจืดกำหนดข้อกำหนดชัดเจน: *"ตัวเลขทั้งหมด ในระบบ ต้องมี ทศนิยม และ comma และต้องชิดขวา ยกเว้นตอนเข้าไปแก้ไข ให้เป็น text ธรรมดา ยังไม่ต้องมี comma"*
  - ผู้ใช้งานบัญชีอายุ 40+ อ่านจำนวนเงินได้ง่ายเมื่อมี Thousands Comma และทศนิยมครบถ้วน แต่ขณะแก้ไข การมี Comma จะขัดขวางการลบและแก้ไขตัวเลข การแยกโหมดระหว่าง Idle (มี Comma + ทศนิยม) และ Focus (Plain text ไม่มี Comma + Select All) จึงเป็นวิธีที่ตอบโจทย์ทั้งความชัดเจนและความสะดวกรวดเร็วในการทำงานจริง
- **Reference Implementation:**
  - Component: `AmountInput`, `formatAmountValue`, `cleanAmountValue` ใน `frontend/src/app/gl/gl-common.tsx:156-326`
  - Component: `NumericInput` ใน `frontend/src/components/ui/numeric-input.tsx:72-350`
  - Usage: `frontend/src/app/gl/gl-masters.tsx:166` (Budget & Forecast Amount)
  - Usage: `frontend/src/app/gl/gl-journals.tsx:98` (Journal Debit & Credit)
  - Usage: `frontend/src/app/menu/product-tab-shared.tsx:125`
  - Unit Tests: `frontend/src/app/gl/amount-input.test.ts`
  - Unit Tests: `frontend/src/components/ui/numeric-input.test.ts`
- **วิธีตรวจ:**
  - `npm test -- src/app/gl/amount-input.test.ts`
  - `npm test -- src/components/ui/numeric-input.test.ts`
  - `npm test`
  - `npm run typecheck`

## 8.20 มาตรฐานการแยกโหมดแสดงข้อมูลและโหมดแก้ไขใน CRUD Table (CRUD List-Detail View Mode vs Edit Mode Standard — 2026-09-13)

- **New Standard Pattern:**
  - **การคลิกแถวในตารางรายการ (Row Click = View Mode / โหมดแสดงข้อมูล):**
    - เมื่อผู้ใช้คลิกบรรทัด/แถวใดๆ ในตารางรายการ (CRUD list table) **ให้ถือเป็นการเปิดแสดงข้อมูลในโหมดแสดงข้อมูล (View Mode / Read-only) เสมอ**
    - ในโหมดแสดงข้อมูล:
      1) ฟิลด์ข้อมูลทั้งหมดต้องถูกปิดการแก้ไข / เป็น `readOnly` หรือครอบด้วย `<fieldset disabled className="grid gap-3 opacity-95">` ป้องกันการพิมพ์หรือแก้ไขข้อมูลโดยไม่ได้ตั้งใจ
      2) **ห้ามมีปุ่ม "บันทึก" (Save) หรือ "ปรับปรุง" (Update) หรือ `<button type="submit">` โผล่มาให้กดเด็ดขาด** ("การแสดงข้อมูล จะกด Save หรือ Update ไม่ได้")
      3) ส่วนหัว (Header) ต้องแสดงหัวข้อและ Badge ระบุสถานะชัดเจน เช่น `แสดงข้อมูล: <CODE>` พร้อม Badge `โหมดแสดงข้อมูล`
      4) มีแบนเนอร์แจ้งเตือนแบบกระชับ: `โหมดแสดงข้อมูล (Read-only) — หากต้องการแก้ไข ให้กดปุ่ม "แก้ไข"`
      5) มีปุ่มเด่นสำหรับเข้าสู่โหมดแก้ไข: ปุ่ม **"แก้ไข" / "แก้ไขข้อมูล"** (`Pencil` icon) และปุ่ม "ปิด" (`X` icon)
  - **การเข้าสู่โหมดแก้ไข (Explicit Edit Mode):**
    - ผู้ใช้จะเข้าสู่โหมดแก้ไข (`isEditing = true`) ได้ผ่าน 2 วิธีอย่างชัดแจ้งเท่านั้น:
      1) กดปุ่ม "แก้ไข" (`Pencil` icon) ในคอลัมน์จัดการของแถวนั้นๆ
      2) กดปุ่ม "แก้ไข" / "แก้ไขข้อมูล" ในหน้าต่างรายละเอียด View Mode
      3) หรือกดปุ่ม "+ เพิ่มรายการ" เพื่อสร้างรายการใหม่
  - **สีพื้นหลังแถวในตารางด้านซ้าย (Datalist Row Background — View vs Edit Mode):**
    - เมื่ออยู่ในโหมดแสดงข้อมูล (View Mode: `isSelected && !isEditing`):
      `bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium` (สีประจำ Palette ระบบ)
    - **เมื่ออยู่ในโหมดแก้ไข (Edit Mode: `isSelected && isEditing`):**
      แถวของรายการที่กำลังถูกแก้ไขด้านซ้าย **ต้องเปลี่ยนสีพื้นหลังเป็นโทนสีส้ม/อำพัน (Amber Active Editing State) ทันที** เพื่อให้ผู้ใช้ทราบชัดเจนว่ากำลังแก้ไขรายการใด:
      `bg-amber-100/70 hover:bg-amber-100/90 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 ring-1 ring-inset ring-amber-500/50 font-medium`
      พร้อมตัวหนังสือรหัส `text-amber-900 dark:text-amber-300`
  - **Dirty Guard สอดคล้องกับโหมดแก้ไข:**
    - ตัวตรวจจับข้อมูลค้างบันทึก (`dirty`) ต้องผูกกับ `isEditing` เสมอ:
      `const dirty = isEditing && record !== null && JSON.stringify(record) !== original;`
      ทำให้ขณะเปิดดูข้อมูลใน View Mode ค่า `dirty` จะเป็น `false` เสมอ จึงสามารถคลิกดูรายการอื่นๆ ในตารางได้อย่างรวดเร็วโดยไม่มี popup เตือนกวนใจ
- **Anti-pattern / Deprecated:**
  - ห้ามเปิดเข้าสู่หน้าฟอร์มที่แก้ไขได้ทันทีเมื่อคลิกแถวในตารางรายการ
  - ห้ามแสดงปุ่ม Save หรือ Update ค้างไว้ในหน้าจอที่เปิดแสดงข้อมูล
  - ห้ามปล่อยให้การคลิกดูแถวในตารางทำให้เกิดสถานะ dirty จนผู้ใช้คลิกดูแถวอื่นต่อไม่ได้
- **Root Cause & Rationale:**
  - กฎจากลุงจืด 2026-09-13: *"CRUD เมื่อกดบรรทัดนั้นๆ ให้ถือเป็นการแสดงข้อมูล ถ้าต้องการแก้ไข ต้องกดปุ่ม แก้ไข การแสดงข้อมูล จะกด Save หรือ Update ไม่ได้ ตรวจให้หมด"*
  - ผู้ใช้งานบัญชีมักเปิดดูข้อมูลเพื่อตรวจสอบความถูกต้อง (Review/Inspection) เป็นประจำ การเปิดหน้าแก้ไขทันทีและมีปุ่ม Save เสี่ยงต่อการเผลอกดแก้ไขหรือกดบันทึกทับข้อมูลเดิมโดยไม่ตั้งใจ การแยก View Mode ออกจาก Edit Mode อย่างเด็ดขาดช่วยป้องกันข้อผิดพลาดทางบัญชีได้อย่างสิ้นเชิง
- **Reference Implementation:**
  - `frontend/src/app/gl/gl-masters.tsx` (`GLMasters` — accounts, fiscal-years, budgets, periods, account-groups, mappings, product-account-groups, forecast)
  - `frontend/src/app/gl/gl-journals.tsx` (`GLJournals` — Daily journals, Opening balances)
  - `frontend/src/app/gl/gl-statement-designer.tsx` (`GLStatementDesigner` — Statement templates)
  - `frontend/src/app/system-settings/system-settings-screen.tsx` (`SettingDataList` & `SettingDetailPanel`)
  - `frontend/src/app/system-settings/company-branch-tree-view.tsx` (Tree view company & branch edit state)
  - `frontend/src/app/menu/product-screen.tsx` (Product catalog master-detail edit state)
  - `frontend/src/app/menu/product-barcode-screen.tsx` (Product barcode master-detail edit state)
  - `frontend/src/app/menu/product-set-screen.tsx` (Product combo set master-detail edit state)
  - Unit Tests: `frontend/src/app/gl/gl-masters.test.ts`
  - Unit Tests: `frontend/src/app/gl/gl-journals.test.ts`
- **วิธีตรวจ:**
  - `npm test -- src/app/gl/gl-masters.test.ts`
  - `npm test -- src/app/gl/gl-journals.test.ts`
  - `npm test`
  - `npx tsc --noEmit`
  - `npm run build`

## 8.21 มาตรฐานแถบเมนูนำทางหลักตัดขึ้นบรรทัดใหม่ (Top Navigation Bar Flex Wrap vs No Horizontal Scroll — 2026-09-14)

- **New Standard Pattern:**
  - **แถบเมนูนำทางด้านบน (Top Navigation Bar):**
    - ใช้คอนเทนเนอร์ `flex flex-wrap items-center gap-1` เสมอ
    - **ห้ามใส่ `overflow-x-auto` หรือแถบเลื่อนแนวนอน**
    - เมื่อความกว้างหน้าจอไม่พอ (เช่น หน้าจอ 1024px, 1280px หรือเปิดหน้าต่างแบบ Split screen) ปุ่มเมนูจะตัดขึ้นบรรทัดใหม่อย่างเป็นธรรมชาติ (`flex-wrap`) ทำให้ผู้ใช้มองเห็นเมนูครบทุกโมดูลโดยไม่ต้องเลื่อนหน้าจอ
  - **การคำนวณตำแหน่ง Popover เมนูย่อย (`openSectionLeft` Clamping):**
    - คอนเทนเนอร์แม่ต้องเป็น `<div className="relative" ref={menuRootRef}>` ครอบปุ่มทั้งหมด เพื่อให้ Popover เมนูลอย (`absolute top-[calc(100%+6px)]`) ลอยอยู่ใต้แถบเมนูรวมทุกบรรทัดเสมอ ไม่ทับซ้อนปุ่มในแถวใดๆ
    - ต้องคำนวณ Clamp ค่า `openSectionLeft` ไม่ให้พาเนลเมนูย่อยหลุดออกนอกขอบขวาของจอ:
      ```tsx
      const containerWidth = menuRootRef.current?.clientWidth ?? window.innerWidth;
      const maxLeft = Math.max(0, containerWidth - 300);
      const clampedLeft = Math.min(Math.max(0, target.offsetLeft), maxLeft);
      setOpenSectionLeft(clampedLeft);
      ```
- **Anti-pattern / Deprecated:**
  - ห้ามใช้ `overflow-x-auto` บนแถบเมนูนำทางหลัก เพราะทำให้เมนูโมดูลท้ายๆ ถูกซ่อนอยู่หลังแถบเลื่อนแนวนอน
  - ห้ามปล่อยให้ `openSectionLeft` อิงตาม `target.offsetLeft` ดิบๆ โดยไม่ Clamp เพราะจะทำให้พาเนลย่อยล้นขอบจอขวาเมื่อปุ่มอยู่ชิดริมขวา
- **Root Cause & Rationale:**
  - กฎจากลุงจืด 2026-09-14: *"ไม่ต้องการให้มีตัวเลื่อน ให้ wrap ลงมาเลย ถ้าล้น"*
  - ผู้ใช้งานบัญชีต้องการมองเห็นเมนูสำคัญทั้งหมดในคราวเดียว การมีตัวเลื่อนแนวนอนทำให้กดใช้งานลำบากและมองไม่เห็นสถานะของโมดูลอื่นๆ
- **Reference Implementation:**
  - `frontend/src/app/menu/main-menu-screen.tsx`
  - ADR: `docs/kms/decisions/2026-09-14-menu-bar-flex-wrap.md`
- **วิธีตรวจ:**
  - `npx tsc --noEmit`
  - `npm test`
  - `npm run build`

## 8.22 มาตรฐานการจัดวางไอคอนในช่องเลือกข้อมูลและ Textbox Action Buttons (Input & Select Addon Buttons Standard — 2026-09-14)

- **New Standard Pattern:**
  - **ช่องเลือกข้อมูล (`<select>`) ที่มีปุ่ม Action ด้านใน (เช่น `AccountSelect`):**
    - ต้องใส่ `appearance-none [&::-ms-expand]:hidden` เสมอ เพื่อปิดการแสดงผลลูกศรดั้งเดิม (`▼`) ของบราวเซอร์อย่างสิ้นเชิง ไม่ให้มีติ่งลูกศรโผล่แลบออกมาทับซ้อนกับปุ่ม
  - **การจัดวางกลุ่มปุ่มด้านขวา (Right Action Addon Group):**
    - คอนเทนเนอร์: `absolute right-2 top-1/2 -translate-y-1/2 z-10 flex items-center gap-1.5 pointer-events-auto`
    - ปุ่ม Clean ล้างค่า (`X`): ใช้รูปทรงวงกลม `inline-flex !size-6 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-full text-muted-foreground/70 hover:bg-muted hover:text-foreground transition-colors cursor-pointer shrink-0`
    - ปุ่ม Action ค้นหา (`🔍`): ใช้รูปทรงสี่เหลี่ยมโค้งมน `inline-flex !size-7 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-[8px] border border-primary/20 bg-primary/10 hover:bg-primary/20 text-primary transition-all active:scale-95 disabled:pointer-events-none disabled:opacity-50 cursor-pointer shadow-2xs shrink-0`
    - เว้นระยะห่างระหว่างปุ่ม `gap-1.5` ชัดเจน และห่างจากขอบขวา `right-2` (8px) ไม่เบียดมุมมน `rounded-xl`
  - **การป้องกัน Global CSS ทับซ้อน (Anti-Selector Clashing):**
    - **ห้ามใช้คลาส `rounded-lg` หรือ `h-8` บนปุ่มขนาดเล็ก/ปุ่มใน input เด็ดขาด**: เนื่องจาก `globals.css:6507` มี selector `button[class*="rounded-lg"]` และ `button[class*="h-8"]` ซึ่งบังคับ `height: 36px !important; padding-inline: 0.7em !important;` ส่งผลให้ปุ่มยืดกลายเป็นวงรีและดันไอคอน SVG เบี้ยวหลุดจุดกึ่งกลาง (Not Centered)
    - **วิธีที่ถูกต้อง**: ให้ใช้ `rounded-[8px]` (Arbitrary value) แทน `rounded-lg` พร้อมใส่ `!min-h-0 !max-h-none !min-w-0 !p-0` เพื่อล้างค่าความสูงและ padding ทั้งหมด
    - **การจัดกึ่งกลางเรขาคณิต (Mathematical Center)**: ใช้ `inline-flex items-center justify-center` ร่วมกับ `!p-0` และจัดขนาดไอคอน SVG ให้สัมพันธ์กับปุ่ม เช่น ปุ่ม `!size-7` (28px) ใช้ไอคอน `size-3.5` (14px) ช่องว่างรอบไอคอนจะเท่ากันสมบูรณ์แบบทั้ง 4 ทิศทาง (บน=ล่าง, ซ้าย=ขวา)
  - **การกำหนด Padding ฝั่งขวาของกล่องข้อความ:**
    - กรณีมี 2 ปุ่ม (Clear + Search): ต้องกำหนด `!pr-20` (80px) เสมอ เพื่อให้ข้อความตัดจบ (`truncate`) ก่อนถึงปุ่ม ไม่ซ้อนทับหรือลอดใต้ปุ่ม
    - กรณีมี 1 ปุ่ม (Search เท่านั้น): กำหนด `!pr-12` (48px)
- **Anti-pattern / Deprecated:**
  - ห้ามใช้คลาส `rounded-lg` บนปุ่มลูกเล่นใน input เพราะจะโดน `button[class*="rounded-lg"]` ใน `globals.css` ครอบงำและยืดเป็น 36px พร้อมใส่ `padding-inline: 0.7em` ทำให้ไอคอนเอียงไปทางขวา ไม่ center
  - ห้ามใช้ `grid place-items-center` เดี่ยวๆ โดยไม่เคลียร์ `padding-inline` เพราะ padding จะดัน SVG หลุดแนวแกน
  - ห้ามลืมใส่ `appearance-none` บน `<select>` ที่มี Custom Action Buttons เพราะบราวเซอร์บน Windows จะวาดลูกศร `▼` ดั้งเดิมซ้อนทับอยู่ใต้ปุ่ม
  - ห้ามวางปุ่มชิดขอบเกินไป (`right-1.5`) ในกล่องที่มีมุมมน `rounded-xl`
  - ห้ามใช้ `!pr-14` (56px) เมื่อมี 2 ปุ่ม (ปุ่มรวมกว้าง 66px) เพราะข้อความจะวิ่งเข้าไปซ้อนทับใต้ปุ่ม Clean
- **Root Cause & Rationale:**
  - คำสั่งจากลุงจืด 2026-09-14: *"textbox icon ทับกัน แก้ให้ด้วย ให้สวยๆ"* และ *"ยังไม่ center แก้ใหม่"*
  - เกิดจากการที่ selector `button[class*="rounded-lg"]` ใน `globals.css:6507` บังคับ `height: 36px` และ `padding-inline: 0.7em` ทับปุ่ม `size-7 rounded-lg` ทำให้ไอคอนขยับไปทางขวา 7.53px และความสูงยืดกลืนกับกรอบ select
- **Reference Implementation:**
  - `frontend/src/app/gl/gl-common.tsx` (`AccountSelect`, `SearchInput`)
  - ADR: `docs/kms/decisions/2026-09-14-input-addon-icons-no-overlap.md`
- **วิธีตรวจ:**
  - `npx tsc --noEmit`
  - `npm test`
  - `npm run build`

## 8.23 มาตรฐานความสูงเต็มหน้าจอและการขยายส่วนที่เหลืออัตโนมัติ (Full-Height Auto-Expanded Workbench Standard — 2026-09-14)

- **New Standard Pattern:**
  - **การตั้งค่า Viewport ในหน้าจอหลัก (`main-menu-screen.tsx`):**
    - เส้นทางหน้าจอที่เป็น Workbench หรือ Data Table แบบแยกส่วน (เช่น `/gl/*`, `/product`, `/productbarcode`, `/datamodelgraph`) ต้องรวมอยู่ในเงื่อนไข `activeTabNeedsFixedViewport` และคลาสของ `<section>`:
      ```tsx
      const activeTabNeedsFixedViewport = 
        activeWorkTab.route === "/productbarcode" || 
        activeWorkTab.route === "/product" || 
        activeWorkTab.route === "/datamodelgraph" || 
        isGeneralLedgerRoute(activeWorkTab.route);
      ```
      และในแท็บคอนเทนเนอร์:
      ```tsx
      <section
        aria-hidden={tab.id !== activeTabId}
        className={cn("min-w-0", (tab.route === "/productbarcode" || tab.route === "/product" || tab.route === "/datamodelgraph" || isGeneralLedgerRoute(tab.route)) && "lg:h-full lg:min-h-0 lg:overflow-hidden")}
        hidden={tab.id !== activeTabId}
        key={tab.id}
        role="tabpanel"
      >
      ```
  - **คอนเทนเนอร์แม่ของ Workbench (`GeneralLedgerScreen`):**
    - เมื่อเป็นโหมดฝัง (`embedded = true`) ต้องให้ความสูงเต็มและตัด overflow ชั้นนอก:
      ```tsx
      <main className={`gl-workbench flex min-w-0 flex-1 flex-col gap-2 text-[0.95rem] leading-relaxed ${embedded ? "p-2 h-full min-h-0 overflow-hidden" : "mx-auto max-w-[1800px] p-3 min-h-[calc(100dvh-2rem)]"}`} data-gl-route={cleanRoute}>
        {/* Navigation Tabs (Top) */}
        <nav className="shrink-0 flex flex-wrap gap-2">...</nav>
        {/* Content Wrapper */}
        <div className="flex-1 min-h-0 flex flex-col">{content}</div>
      </main>
      ```
  - **การแยกส่วนซ้ายขวา (`SplitWorkbench`):**
    - คอนเทนเนอร์และพาเนลซ้าย-ขวาต้องยืดเต็มความสูงแนวตั้ง โดยอาศัยธรรมชาติ `align-items: stretch` ของ Flex Row:
      ```tsx
      <div ref={container} className={`flex min-w-0 flex-1 flex-col gap-2 xl:flex-row xl:h-full xl:min-h-0 ${className ?? ""}`} style={{ "--gl-list-width": `${width}%` } as CSSProperties}>
        <div data-gl-pane="list" className={`${panel} flex flex-col min-h-0 w-full xl:w-[var(--gl-list-width)] xl:shrink-0`}>{list}</div>
        <ResizableSplitter ... />
        <div data-gl-pane="editor" className={`${panel} flex flex-col min-h-0 flex-1`}>{editor}</div>
      </div>
      ```
  - **การจัดโครงสร้างฝั่งรายการ (List Side):**
    - ส่วนหัวค้นหา (Search toolbar): `shrink-0`
    - แถวสรุปจำนวนรายการ (Stats row): `shrink-0`
    - ตารางข้อมูล (Data Table): **ใช้ `flex-1 min-h-[300px] overflow-auto rounded-xl border border-border`** (ขยายเต็มพื้นที่ว่างแนวตั้งที่เหลือทั้งหมด ไม่จำกัดด้วย max-h)
    - แถบเปลี่ยนหน้า (Pager bar): `shrink-0`
  - **การจัดโครงสร้างฝั่งแก้ไข/แสดงข้อมูล (Editor Side):**
    - ส่วนหัว (Header): `shrink-0`
    - แถบแจ้งเตือนสถานะ (Notice/Banner): `shrink-0`
    - ฟอร์มกรอกข้อมูลหรือตารางรายละเอียด (Fieldset/Lines): **ในนี้ไม่ต้อง expanded ให้ใช้ความสูงตามธรรมชาติ พร้อม `content-start overflow-y-auto pr-1`** (ห้ามยืดกระจายช่องกรอก) เลื่อนเฉพาะภายในเมื่อเนื้อหายาว
    - ส่วนท้ายปุ่มควบคุม (Footer actions): **ให้อยู่ด้านล่างสุดเสมอด้วย `shrink-0 mt-auto`**
    - สถานะว่าง (Empty state): ใช้ `flex flex-col flex-1 h-full min-h-64 items-center justify-center content-center gap-3 text-center p-6` เพื่อให้ไอคอนและข้อความอยู่กึ่งกลางพื้นที่พาเนลพอดี

- **Anti-pattern / Deprecated:**
  - **ห้ามขยายช่องกรอกฟอร์มจนกระจายตัว:** ห้ามใส่ `flex-1` บน Fieldset ใน Editor Pane เพราะจะทำให้ CSS Grid ยืดแถวช่องกรอกและ Checkboxes ห่างกันเป็นช่องว่างมหึมาผิดธรรมชาติ ให้ใช้ความสูงตามธรรมชาติและจัดกลุ่มกระชับ (`content-start`) แล้วใช้ `mt-auto` ผลัก Footer ไปตรึงขอบล่างแทน
  - **กับดัก CSS Flexbox:** ห้ามใส่ `h-full` หรือ `xl:h-full` บน Flex Item ภายใน Flex Row ที่คอนเทนเนอร์แม่ได้ความสูงจาก `flex: 1` เพราะตามสเปก CSS เมื่อ flex item มี `height: 100%` เบราว์เซอร์จะปิดการทำงานของ `align-self: stretch` และเมื่อ parent ไม่มี explicit height ที่ระบุเป็น px ค่า `height: 100%` จะคำนวณไม่ออกและตกกลับไปเป็น `auto` ทำให้ความสูงหดเหลือเท่ากับเนื้อหาภายในทันที! ให้ใช้ `flex flex-col min-h-0` โดยไม่ต้องใส่ `h-full` เพื่อปล่อยให้ `align-items: stretch` ยืดพาเนลเต็มความสูงแนวตั้งตามธรรมชาติ
  - ห้ามใช้ `max-h-[62vh]` หรือ `max-h-[65vh]` กับตารางในหน้าจอ Workbench เพราะบนจอขนาดใหญ่หรือจอที่มี Header สั้น จะทำให้เกิดช่องว่างสีพื้นโล่ง (Dead Space) ขนาดใหญ่ถึง 300–500px ที่ด้านล่างของหน้าจอ
  - ห้ามลืมใส่ `min-h-0` บน Flex child ที่ต้องการให้หดหรือยืดพอดี (`flex-1 min-h-0`) เพราะตามค่าเริ่มต้นของ CSS Flexbox `min-height` จะเป็น `auto` ทำให้ Flex item ขยายเกินกรอบจนหลุดจอ
  - ห้ามลืมใส่ `isGeneralLedgerRoute(activeWorkTab.route)` ใน `activeTabNeedsFixedViewport` เพราะจะทำให้แท็บพาเนลของเมนูใช้ `lg:overflow-y-auto` ซึ่งเลื่อนทั้งหน้าจอแทนที่จะตรึง Viewport แล้วให้ตารางภายในเลื่อนเอง
  - ห้ามปล่อยให้ Empty State (`เลือกรายการเพื่อแสดงข้อมูล`) ใช้แค่ `min-h-64` โดยไม่ใส่ `h-full items-center justify-center` เพราะกล่องจะกองอยู่ด้านบนแล้วปล่อยพื้นที่ด้านล่างว่างโล่ง

- **Root Cause & Rationale:**
  - คำสั่งจากลุงจืด 2026-09-14: *"ความสูง ต้อง expanded ส่วนที่เหลือด้วย auto ด้วย"*
  - ผู้ใช้งานระบบบัญชีต้องการเห็นจำนวนแถวข้อมูลมากที่สุดเท่าที่ความสูงของหน้าจอจะเอื้ออำนวย (Viewport Height Optimization) โดยไม่ต้องคอยเลื่อนดูตารางสั้นๆ ที่มีช่องว่างโล่งไร้ประโยชน์กว่า 50% ของหน้าจอด้านล่าง
  - การตรึง Search Bar และ Pager ไว้ที่บน-ล่าง แล้วให้ Table ขยายเต็มพื้นที่ส่วนที่เหลือ (`flex-1`) ช่วยเพิ่มประสิทธิภาพการทำงานของนักบัญชีอย่างมาก

- **Reference Implementation:**
  - `frontend/src/app/menu/main-menu-screen.tsx` (บรรทัด 658, 1469)
  - `frontend/src/app/gl/general-ledger-screen.tsx` (บรรทัด 46–50)
  - `frontend/src/app/gl/gl-common.tsx` (`SplitWorkbench`, บรรทัด 459–470)
  - `frontend/src/app/gl/gl-masters.tsx` (บรรทัด 160–186, 308–488)
  - `frontend/src/app/gl/gl-journals.tsx` (บรรทัด 153–186, 295–600)
  - `frontend/src/app/gl/gl-statement-designer.tsx` (บรรทัด 310–350, 418–855)
  - `frontend/src/app/gl/gl-reports.tsx` (บรรทัด 33–88)
  - ADR: `docs/kms/decisions/2026-09-14-gl-workbench-full-height-expanded.md`
- **วิธีตรวจ:**
  - `npx tsc --noEmit`
  - `npm test`
  - `npm run build`
  - ตรวจวัดขนาด DOM จริงบนหน้าจอ (เช่น `listPaneBox.height` ขยายเต็มพื้นที่ เหลือช่องว่างด้านล่าง < 10px)
## 8.24 มาตรฐานแจ้งข้อผิดพลาดจาก API แบบไทยในหน้าต่างแก้ไข (Thai Inline API-Error Standard — 2026-09-14)

- **New Standard Pattern:**
  - **หนึ่งหน้าจอ หนึ่งกล่องข้อความ Thai ** — เมื่อบันทึกไม่ผ่าน ให้แสดงข้อความไทยในกรอบของหน้าต่าง/ส่วนแก้ไขที่ผู้ใช้กำลังกรอกอยู่ (ไม่ใช่แถบด้านบนสุดของหน้า) ใช้คอมโพเนนต์ `Notice` ที่ให้ `role="alert"` + `text-[0.95rem]` อยู่แล้ว → `frontend/src/app/gl/gl-masters.tsx:384, :460` (ทั้งโหมดดูและโหมดแก้ไข)
  - **ข้อความมาจากเซิร์ฟเวอร์เสมอ** — อ่านฟิลด์ `message` (ถ้าไม่มี อ่าน `error.message`) ถ้าเป็นข้อความไทยให้ใช้ทันที ถ้าไม่ใช่/ว่าง ให้แปลจากรหัสเครื่อง `code` ผ่านตารางไทยสำรอง แล้วจึงค่อยใช้ประโยคกลาง → `frontend/src/lib/general-ledger-api.ts:14-45` (`commandCodeMessages`) + `:73` (`commandErrorInfo`, `GLCommandError`)
  - **ห้ามโชว์ข้อความอังกฤษ/รหัสเทคนิค** — ตัดข้อความที่ไม่มีอักษรไทยทิ้งทั้งหมด (regex `[ก-๛]`) แล้วแทนด้วยประโยคไทยกลาง `frontend/src/lib/general-ledger-api.ts:67-71` (`commandFailure`)
  - **ค่าในฟอร์มต้องอยู่ครบ และโฟกัสไปช่องที่ผิด** — ตอนบันทึกไม่ผ่าน อนุญาตให้แก้เฉพาะ state `{ error, errorField }` เท่านั้น (ไม่แตะ record) → `frontend/src/app/gl/gl-masters.tsx:24` (`errorStatePatch`) และย้ายโฟกัสด้วย `data-field` ที่วางบนคอนโทรลจริง + `focus({ preventScroll: true })` เพื่อไม่ให้หน้าจอกระโดด → `frontend/src/app/gl/gl-masters.tsx:60-65`; ตัวคอนโทรล: `gl-masters.tsx:534` (`data-field="accountcode"`), `:535` (`accountnameth`, `accountnameen`), `:539`+ `gl-common.tsx:342,349` (`AccountSelect field=...`)
  - **BFF: ข้อผิดพลาดที่ผู้ใช้แก้เองได้ ไม่ต้องทำให้ console ขึ้น error** — ถ้า 4xx และ body มี `code` (ไม่ใช่ 401/403) ให้ส่งต่อเป็น HTTP 200 + `success:false` แล้วให้หน้าจอเช็ค `success` เอง ส่วน 401/403/5xx คงสถานะจริง → `frontend/src/lib/workspace-api.ts:61-91` และเปิดใช้เฉพาะคำสั่ง GL ที่ `frontend/src/app/api/gl/[...glPath]/route.ts:19`
  - **สีของสถานะผิดพลาดต้องมาจากโทเคนธีม** — `text-destructive`, `border-border`, `hover:bg-destructive/10`, `hover:border-destructive/40`, `focus-visible:ring-destructive/30` → `frontend/src/app/gl/gl-masters.tsx:314, 428, 499`
- **Anti-pattern / Deprecated:**
  - ปล่อยให้บันทึกไม่ผ่านแล้วเงียบ (มีแต่ `console.error`) หรือให้ผู้ใช้เดาเองจากปุ่มที่กดไม่ติด
  - แสดงข้อความจากผู้ให้บริการดิบ ๆ เช่น `E11000 duplicate key`, `Unexpected token <`, `Failed to load resource`, เลข HTTP หรือ stack trace
  - แสดงกล่องข้อผิดพลาดสองที่พร้อมกัน (แถบบนสุดของหน้า + ในหน้าต่าง) — ผู้ใช้จะอ่านไม่รู้ว่าอันไหนคือสาเหตุ; ถ้ามี record เปิดอยู่ ให้ส่วนแก้ไขเป็นเจ้าของข้อความ และแถบหน้าเงียบ (`editorAlert` → `frontend/src/app/gl/gl-masters.tsx:15-19`, ใช้ที่ `:193, :384, :460`)
  - วาง `data-field` ไว้ที่ `<label>`/ตัวห่อ แล้วโฟกัสไม่ได้จริง — ต้องวางบน `input/select` (หรือให้คอมโพเนนต์ส่งต่อ prop ให้คอนโทรลของตัวเอง) หรือมี fallback `root.querySelector("input,select,textarea,button")`
  - `text-red-600`, `border-red-200`, `hover:bg-red-50` และคู่ dark mode ที่เขียนตายตัว — สีจะไม่ตรงกับพาเลตที่ผู้ใช้เลือก
  - ใช้ `response.json()` ตรง ๆ กับทุก response — ถ้า backend ตอบ HTML/ข้อความเปล่า จะได้ error อังกฤษหลุดถึงผู้ใช้ (ต้อง `.catch(() => ({}))` → `frontend/src/lib/general-ledger-api.ts:108`)
- **Root Cause & Rationale:**
  - การล้มเหลวแบบเงียบคือกับดักใหญ่ของผู้ใช้อายุ 40+ (กฎข้อ 8: ทุก action ต้องมี feedback ภาษาไทย) และการ์ดข้อความที่อยู่ไกลจุดกดทำให้ผู้ใช้คิดว่า “กดไม่ติด” — จึงต้องวางกล่องข้อความในบริเวณที่ตากำลังมอง (หน้าต่างแก้ไข)
  - เบราว์เซอร์ (Chrome) จะบันทึก console error ให้ทุก response ที่สถานะ 4xx/5xx ของ fetch → ถ้าอยากให้ “console error = 0” ระหว่างที่ผู้ใช้แก้ฟอร์ม ต้องให้ BFF แปลงข้อผิดพลาดที่คาดได้ (มี `code`) เป็น 200 + `success:false`; แต่ 401/403/5xx ต้องคงสถานะจริงเพื่อไม่บัง auth failure
  - ข้อความไทยต้องเป็นของเซิร์ฟเวอร์ (แหล่งเดียว) เพื่อไม่ให้ข้อความสองฝั่งไม่ตรงกัน ส่วนตารางไทยสำรองมีไว้กันกรณี message หาย ไม่ใช่แทนที่
- **Reference Implementation:**
  - หน้าจอ: `frontend/src/app/gl/gl-masters.tsx` (ผังบัญชี `/gl/chartofaccounts`) — `showFormError` `:53-58`, `editorAlert` `:15-21`, โฟกัส `:60-65`, กล่องข้อความ `:384, :460`
  - ชั้น API: `frontend/src/lib/general-ledger-api.ts` (`commandErrorInfo` `:73`, `commandFailure` `:67-71`), `frontend/src/lib/workspace-api.ts:61-91`, `frontend/src/app/api/gl/[...glPath]/route.ts:19`
  - ฝั่งเซิร์ฟเวอร์ (สัญญาข้อผิดพลาด): `backend/internal/generalledger/errors.go` (`duplicate_code` = 409 + `message` ไทย), `backend/internal/generalledger/httpapi/http.go:159`
  - ทดสอบ: `frontend/src/app/gl/gl-masters.test.ts:180-235` (หนึ่ง alert, ค่าคงอยู่, mutual exclusivity), `frontend/src/lib/workspace-api.test.ts`, `frontend/src/lib/general-ledger-api.test.ts`; E2E `frontend/e2e/gl-chart-of-accounts-error.spec.ts`
