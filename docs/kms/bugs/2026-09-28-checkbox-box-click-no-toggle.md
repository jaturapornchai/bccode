---
date: 2026-09-28
severity: medium
component: [frontend]
tags: [bc-account, ui, checkbox, gl, financial-statements, accessibility]
fixed: true
---

# Symptom

หน้าต่างเลือกบัญชีแบบหลายรายการ (เปิดจากจอออกแบบงบการเงิน `/gl/statement-designer`): คลิกที่ **สี่เหลี่ยมช่องติ๊ก** ของแถวบัญชีแล้วไม่มีอะไรเกิดขึ้น ต้องคลิกที่ส่วนอื่นของแถวถึงจะติ๊ก — ผู้ใช้ 40+ คลิกสี่เหลี่ยมก่อนเสมอจึงคิดว่าระบบค้าง อาการเดียวกันเกิดกับช่องติ๊กในตารางอื่นที่ไม่มี `<label>` ห่อ

ระหว่างแก้พบ regression 2 จุด (แก้ในรอบเดียวกันก่อน commit):
1. input ล่องหนสูงกว่ากล่อง → คลิกขอบบนของ**แถวถัดไป**ในจอออกแบบงบ ไปติ๊ก "+/-" (กลับเครื่องหมาย) ของแถวก่อนหน้า = เครื่องหมายบรรทัดงบเปลี่ยนเงียบ ๆ
2. หลังคลิกช่องติ๊กของแถว กด Space แล้วไปติ๊กบัญชีที่ไฮไลต์อยู่เดิม ไม่ใช่แถวที่เพิ่งคลิกและเห็นวงโฟกัส

## Root Cause

- `Checkbox` (`frontend/src/components/ui/checkbox.tsx`) เดิมเป็น `<input class="peer sr-only">` + กล่องที่วาด `<span aria-hidden>` ซึ่งพึ่ง `<label>` ห่อให้คลิกกล่องแล้วส่งต่อไป input — นอก label คลิกกล่องไม่โดนทั้ง input และ label
- จุดที่ตายจริง: `frontend/src/app/gl/account-search-dialog.tsx:517-529` `<td>` ของโหมดหลายรายการเรียก `e.stopPropagation()` (กัน `onClick` ของแถว :496-499 ด้วย) และ `Checkbox` ไม่อยู่ใน label; อีกสองจุดที่อาการเดียวกัน: `frontend/src/app/system-settings/company-branch-tree-view.tsx:375` (div ไม่มี onClick) และ `frontend/src/app/tax/tax-form-fields.tsx:57` (label เป็นพี่น้องแบบ `htmlFor` ในกริด / `<td>` เปล่าในตารางแถว — เดิม toggle ได้ทางคีย์บอร์ดเท่านั้น)
- Regression 1: rule unlayered `input, select, textarea { min-height: 2.6em }` (`frontend/src/app/globals.css:5083-5087`) ชนะ `size-full` ที่อยู่ใน layer ของ Tailwind v4 → input ล่องหนสูง 29–31px บนกล่อง 18.8px (root 15px) ห้อยใต้กล่อง ~8px ทับแถวถัดไป — การทดสอบด้วยหน้าจำลองที่ไม่โหลด `globals.css` จับไม่ได้ (reviewer เจอบน localhost:3000 จริง)
- Regression 2: keydown ของหน้าต่าง (`account-search-dialog.tsx:232` Enter, :242 Space) `preventDefault()` แล้ว toggle **แถวที่ไฮไลต์**; `<td>` หยุด `onClick` ของแถวจึงไม่เคยเรียก `setHighlightedIndex(index)` ไฮไลต์ค้างที่แถวเดิม

## Fix

- `frontend/src/components/ui/checkbox.tsx:30` input จริงวางทับกล่องแบบโปร่งใส: `peer absolute inset-0 m-0 size-full min-h-0! cursor-pointer appearance-none opacity-0 disabled:cursor-default` (มาก่อนกล่องเพื่อให้ `peer-*` ทำงาน); กล่อง `pointer-events-none` (:36); hover ย้ายเป็น `peer-hover:border-primary/60` เฉพาะยังไม่ติ๊กและไม่ disabled (:41); ลบคลาสที่ไม่มีผลแล้ว (`cursor-pointer` บนกล่อง, `cursor-default pointer-events-none` ใน disabled) — API ของ `Checkbox`/`CheckboxCard` ไม่เปลี่ยน ไม่ต้องแก้ผู้เรียก
- `min-h-0!` ต้องเป็นแบบ important (`min-h-0` ธรรมดาแพ้ rule global ที่ไม่อยู่ใน layer — กับดักเดียวกับ `text-xs!`); Tailwind 4.3.3 ออก `.min-h-0\! { min-height: 0px !important }`
- `frontend/src/app/gl/account-search-dialog.tsx:520-523` `<td>` ยัง `stopPropagation()` (แถวไม่ toggle ซ้ำ) แต่เพิ่ม `setHighlightedIndex(index)` ให้ Space/Enter ทำกับแถวที่เพิ่งคลิก
- ทำไมไม่ toggle สองรอบใน `<label>`: คลิกที่ input เอง = คลิก interactive content ใน label → label ไม่ส่งคลิกซ้ำ (กฎ label activation ของ HTML); React `onChange` ของ checkbox แยกจาก `onClick` จึง `stopPropagation` ใน `onClick` ของพ่อไม่บล็อกการติ๊ก

### ตรวจการใช้งานทั้งระบบ (25 จุด / 14 ไฟล์)

path ในตารางอยู่ใต้ `frontend/src/app/` (ยกเว้น `components/…` อยู่ใต้ `frontend/src/`)

| กลุ่ม | จุดใช้งาน | ผล |
|---|---|---|
| อยู่ใน `<label>` ไม่มี onClick พ่อที่ toggle ซ้ำ | `gl/account-search-dialog.tsx:404` (เฉพาะบัญชีลงรายการ); `gl/gl-statement-designer.tsx:789, 800, 807, 861, 1050, 1126`; `gl/gl-statement-set.tsx:328` (เลือกแม่แบบ), `:341` (`includeNotes`, disabled ได้); `gl/gl-statement-suggestions.tsx:147` (`mt-0.5` — input คลุม wrapper รวม margin); `menu/manage-shortcuts-screen.tsx:388`, `menu/product-set-screen.tsx:1499`, `menu/product-tab-shared.tsx:87` (label disabled มี `pointer-events-none`), `menu/tab-product-marketplace.tsx:75`, `menu/tab-product-timeforsale.tsx:86`; `tax/wht-certificate-panel.tsx:574` (พิมพ์สำเนาคู่ฉบับ), `:578` (เป็นใบแทน) | ปกติ toggle ครั้งเดียว |
| `CheckboxCard` (label ในตัว) | `menu/tab-product-logistics.tsx:74`, `system-settings/company-branch-tree-view.tsx:2071, 2147, 2222` (มี `id`), `components/system-settings/field-editor.tsx:221` | ปกติ |
| ไม่มี label ห่อ (เคยคลิกกล่องไม่ติด) | `gl/account-search-dialog.tsx:525` (td หยุด propagation; แถวมี onClick/onDoubleClick toggle), `system-settings/company-branch-tree-view.tsx:375`, `tax/tax-form-fields.tsx:57` | **แก้แล้ว** ด้วยการแก้ component เดียว |

- ดับเบิลคลิกที่ช่องติ๊กของแถวในหน้าต่างเลือกบัญชี = native 2 ครั้ง + `onDoubleClick` ของแถว 1 ครั้ง = สุทธิ 1 ครั้ง เท่ากับดับเบิลคลิกส่วนอื่นของแถว (พฤติกรรมเดิม)
- **ยังค้าง (นอกขอบเขต):** ช่อง "เฉพาะบัญชีลงรายการ" (`onlyPosting`, `account-search-dialog.tsx:404`) ในโหมดหลายรายการ — โฟกัสอยู่แล้วกด Space จะไป toggle แถวที่ไฮไลต์แทน เพราะ keydown ยกเว้นแค่ช่องค้นหา (:242); ส่วนที่เหลือของเซลล์ช่องติ๊ก (นอกกล่อง 18.8px ในแถวสูง 41px) ยังคลิกไม่ติด — reviewer เสนอ `if (e.target === e.currentTarget) toggleMultiCheck(...)` ใน `onClick` ของ td (ยังไม่ได้ทำ)

## Regression Test

- `frontend/src/components/ui/checkbox.test.ts:21-46` — input ไม่มี `sr-only`; มีครบ `peer absolute inset-0 size-full min-h-0! opacity-0 m-0 appearance-none cursor-pointer`; กล่องมี `pointer-events-none` + `peer-focus-visible:ring-2` + `peer-hover:border-primary/60`; input มาก่อนกล่อง; disabled render `disabled=""` และไม่มี `peer-hover`
- `npx vitest run src/components/ui/checkbox.test.ts src/app/gl/account-search-dialog.test.ts src/app/gl/gl-statement-suggestions.test.ts` → 3 ไฟล์ 20/20 ผ่าน; `npx tsc --noEmit -p .` 0 error
- ตรวจสดบน `next dev` localhost:3000 (Demo → rungrueng/01/00000 → `/gl/statement-designer`): input ทุกตัวสูง 18.8px เท่ากล่อง, `min-height` computed 0px; `elementFromPoint` ใต้กล่อง 1/3/6/9px ไม่โดน input; คลิกกล่อง "ซ่อนรายการที่ไม่มียอด" toggle เฉพาะตัวนั้น; หน้าต่างเลือกบัญชีหลายรายการ: ไฮไลต์อยู่ 1000 → คลิกกล่อง 1131 = ติ๊กเฉพาะ 1131 ไฮไลต์+โฟกัสย้ายมา → Space ยกเลิก 1131, 1000 ไม่ถูกติ๊ก; ไม่มี console error; ไม่ได้บันทึกแม่แบบ (ไม่มีข้อมูลเขียนลง DB)
- prod r20260928-1 (2026-09-28, Demo → rungrueng/01/00000 → `/gl/statement-designer` → แม่แบบมาตรฐานที่ยังไม่บันทึก → "+ เลือกผังบัญชี" 188 บัญชี): input 19×19px ตรงกับกล่องพอดี; คลิกเมาส์จริงที่กล่องแถว 1111 → "เลือกอยู่ 1 บัญชี" เฉพาะแถวนั้น + โฟกัสอยู่ที่ checkbox; คลิกซ้ำ → 0; ปิดหน้าต่างและออกจากจอโดยไม่บันทึก — PG `rungrueng` kind `statement-templates` ที่ไม่ถูกลบ = 0 (18 แถวเป็น `isdeleted=true` แก้ล่าสุด 2026-09-27)
- skill: `.agents/skills/ui-scale-polish/SKILL.md` §8.61
