---
date: 2026-10-06
severity: medium
component: [frontend]
tags: [bc-account, ui, gl, chart-of-accounts, dark-mode, accessibility, toast]
fixed: true
---

# Symptom

พบระหว่างถ่ายวิดีโอสอนจอ "รายละเอียดผังบัญชี" (`/gl/chartofaccounts`) — 7 จุด + 1 จุดจากรีวิว:

1. ป้าย "บัญชีย่อย (บันทึกรายการ)" ในมุมมองผังต้นไม้ตัด 3 บรรทัด แถวสูงไม่เท่ากัน
2. ช่องรหัสบัญชีในฟอร์มแก้ไขดูเหมือนช่องปกติ ผู้ใช้ไม่รู้ว่าแก้ไม่ได้
3. ป้าย "มิติวิเคราะห์ (Champ)" โชว์ชื่อระบบต้นแบบให้ผู้ใช้เห็น
4. แบนเนอร์ "ห้ามลบเด็ดขาด" ขึ้นทุกบัญชี (รวมบัญชีที่ไม่มีรายการ) ผู้ใช้เลิกอ่าน
5. หัวรอง dialog ลบเขียน "1111 · ลบรายการ" (ซ้ำกับหัวข้อ ไม่บอกชื่อบัญชี)
6. ปุ่มยุบ/ขยายต้นไม้มี `aria-label` อังกฤษ "Collapse"/"Expand"
7. แถบ "บันทึกเรียบร้อยแล้ว" ค้างข้ามงาน
8. (รีวิว) โหมดมืด: ข้อความ "สำเร็จ" ทั้งระบบเป็นเขียวเข้ม `#13795b` บนพื้นมืด ≈ 2.5:1 อ่านไม่ออก

## Root Cause

- 1: คอลัมน์ป้าย `w-16` ตายตัว + ข้อความยาว; `whitespace-nowrap` ธรรมดาไม่ได้ผลเพราะ rule ที่ไม่อยู่ใน layer `p, small, label, span { text-wrap: pretty }` (`frontend/src/app/globals.css:5091-5096`) reset `text-wrap-mode` ชนะ utility ของ Tailwind
- 2: class `control` (`frontend/src/app/gl/gl-common.tsx:46`) ตั้งใจให้ช่อง disabled ดูปกติสำหรับโหมดดูข้อมูล; ใส่ `bg-*` บน input ก็แพ้ `input { background-color: transparent !important }`
- 3/4/5/6: ข้อความใน `languages.tsv` / โค้ดเดิม (แบนเนอร์ใน `AccountFields`, หัวรองใช้ชื่อการกระทำ, aria-label ฝังอังกฤษ)
- 7: state `message` ของจอไม่เคยถูกล้างหลังบันทึกสำเร็จ
- 8: `paletteVars()` เขียน `--success` ค่าเดียวทั้งสองโหมด inline บน `<html>` (`frontend/src/lib/theme-data.ts:611`) ทับค่า dark `#86efac` ใน globals.css (`:203`, `:280`) — `--danger` แยกโหมดถูกต้องอยู่แล้ว; บวกกับ tint ของ `.message.success` ในโหมดมืดเป็นสีโปร่ง (toast ลอยทับเนื้อหา ข้อความด้านล่างทะลุ)

## Fix

- `frontend/src/app/gl/gl-masters.tsx`: ป้ายสั้น `gl_posting_account_short`/`gl_control_account_short` + `whitespace-nowrap!` + title ชื่อเต็ม (`TreeNodeRow` :223-226); `CodeInput` (:1172) = พื้นเทาที่ wrapper + กุญแจ + `pr-9!` + `cursor-not-allowed` + คำใบ้ `gl_code_locked_hint` ใต้ช่อง (ซ่อนเป็น sr-only ในโหมดดูด้วย `in-data-[mode=view]:sr-only`, fieldset โหมดดู :1007); ถอดแบนเนอร์ (กฎการลบอยู่ใน dialog ลบ `gl_coa_referenced_no_delete`; backend `deletePGAccountGuard` ยังตัดสิน); `confirmSubject` (:184) = "รหัส · ชื่อ" จาก snapshot ที่บันทึกไว้ (+ เหตุผลถ้าพิมพ์); aria-label ไทย `gl_collapse_sub_accounts`/`gl_expand_sub_accounts` + `aria-expanded`; สำเร็จใช้ `toast.success/info` (หายเองใน 5 วินาที)
- `backend/assets/language/languages.tsv`: ตัด "(Champ)" ออกจาก `gl_dim_separate`/`gl_dim_separate_hint`, ข้อความ `gl_coa_referenced_no_delete` ใหม่, เพิ่ม 6 key
- `frontend/src/lib/theme-data.ts:611`: `"--success": isDark ? "#86efac" : "#13795b"` (มีผลทุกจอ)
- `frontend/src/components/toast-viewport.tsx:24` `opaqueSurface`: toast success/error = tint ทับ `var(--panel)`

**ยังค้าง (แจ้งลุงจืดแล้ว ไม่ได้แก้):** toast มุมขวาล่างบังปุ่ม "ลบรายการนี้" ของแผงขวา 5 วินาทีหลังบันทึก (ตำแหน่ง toast ใช้ร่วม ~12 จอ); `span { text-wrap: pretty }` ทำให้ `whitespace-nowrap` ใช้ไม่ได้ทั้งระบบ (เสนอเปลี่ยนเป็น `text-wrap-style: pretty`); ช่องติ๊กมิติแผนก/โครงการยังไม่มีผลกับการบันทึกรายวัน

## Regression Test

- `frontend/src/app/gl/gl-masters.test.ts` describe "ผังบัญชี: tree badges, toggle labels and delete confirm (2026-10-06)" — ป้าย `whitespace-nowrap!` + ข้อความมาจาก key, aria-label ไทย + aria-expanded, `confirmSubject` 4 กรณี
- `sh tools/verify.sh frontend` ผ่าน (vitest 1083/1083); `npm run verify:all` ผ่านครบ (codemap, frontend, frontend-build, backend, postgres)
- local (Demo → rungrueng/01/00000): ช่องรหัส padding-right 37px, cursor not-allowed, คำใบ้เห็นเฉพาะโหมดแก้ไข; แก้ชื่อค้างแล้วกดลบ → หัวรองเป็นชื่อเดิม; toast โหมดมืด (กดปุ่มสลับธีมจริง) `--success` = `#86efac` พื้นทึบ; PostgreSQL บัญชี 1111 แก้แล้วคืนค่าเดิม
- prod r20261006-coa-fixes-1 (Demo → rungrueng/01/00000 → ผังบัญชี → มุมมองผังต้นไม้): ป้ายบัญชีย่อย 172 ตัว สูง 21px `text-wrap-mode: nowrap`; ปุ่มยุบ/ขยาย 17 ปุ่มเป็น "ยุบบัญชีย่อยของ 1000…" ไม่มีอังกฤษ; ไม่มี "ห้ามลบเด็ดขาด"/"Champ"; ช่องรหัสล็อก + คำใบ้; dialog ลบ "1111 · เงินสดในมือ"; console มีแค่ 401 `/api/auth/refresh` ก่อน login (auth guard ปกติ)
- skill: `.agents/skills/ui-scale-polish/SKILL.md` §8.63 (+ §8.14 ทำเครื่องหมายยกเลิกแบนเนอร์)
