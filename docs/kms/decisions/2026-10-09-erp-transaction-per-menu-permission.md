---
date: 2026-10-09
status: accepted
tags: [bc-account, erp, transaction, permission, go, security]
---

# `/transaction/:module` บังคับสิทธิ์รายเมนูที่ backend

## Context
ทุกจอ ERP เปิดพร้อมกันใน commit `6c4447f8` แต่ `/transaction/:module` ตรวจแค่ session + บริษัท — ผู้ใช้ที่มีสิทธิ์เมนูอื่น (หรือไม่มีสิทธิ์เลย) ยิง API อ่าน/สร้าง/แก้/ลบเอกสารทุกประเภทได้ และ `ErpCrudWorkbench` แสดงปุ่มเพิ่ม/แก้/ลบให้ทุกคน ขณะที่ GL ตรวจสิทธิ์รายเมนูอยู่แล้ว (`generalledger/httpapi`) ทำให้ ERP เป็นช่องทางอ้อมไปเขียนใบสำคัญลง GL

## Decision
- **รหัสสิทธิ์ = menu id** เหมือน GL: `<menu>` = เข้าจอ/อ่าน, `<menu>:create|update|delete` = กระทำ; เขียนต้องมีทั้งสิทธิ์เข้าจอและ `<menu>:<action>`; `*` (OWNER/ADMIN) ผ่านทั้งหมด; resolve จาก DB ทุกคำขอ (`holding_members` + `role_permissions`) fail-closed
- **ทะเบียน `doctype` เป็นแหล่งเดียวของ "ประเภทเอกสาร → menu id"** (`DocType.Screen`, `doctype.Screens(key)` คืนจอหลักก่อนแล้วจอสำรอง) — `stock-count` เปิดได้จากสองจอ (`stock-count`, `stock-count-sheet`) สิทธิ์จอใดจอหนึ่งพอ แต่ action ต้องมาจากจอเดียวกับที่ให้สิทธิ์เข้า; ประเภทที่ไม่มี menu id ถูกปฏิเสธ
- ทุก handler (`list/get/create/update/delete`) เรียก `Http.authorize(request, action)` ที่เดียว: module → session (`glhttp.ResolveSessionAccess`) → สิทธิ์ → เชื่อม DB → `ensureSchema`; ปฏิเสธด้วย **403** `forbidden` + ข้อความ key `erp_err_read_forbidden` / `erp_err_action_forbidden` (th/en ใน `languages.tsv`) — ไม่ใช้ 401 เพราะ UI จะวนขอ token ใหม่โดยไม่จำเป็น
- Frontend: `ErpCrudWorkbench` รับ `auth`/`workspace` แล้วใช้ `useScreenActions` ซ่อนปุ่มสร้าง/แก้/ลบและกันใน handler; `statusErrorKey` แยก 403 อ่าน/เขียน — UI เป็นแค่ความสะดวก การบังคับจริงอยู่ backend

## Consequences
- ผู้ใช้ที่เคยใช้งานได้เพราะไม่มีการตรวจ อาจถูก 403 จนกว่าผู้ดูแลจะติ๊กสิทธิ์เมนู/action ที่หน้า "ตั้งค่าระบบและการเข้าถึง › กลุ่มสิทธิ์"
- เพิ่มประเภทเอกสารใหม่ต้องใส่ `Screen` (เทสต์ `TestEveryTypeHasItsOwnPermissionScreen` และเทสต์ฝั่ง frontend เทียบ route → menu id ทุก route จะล้มถ้าลืม/ไม่ตรง)
- ยังไม่ทำ: สิทธิ์ย่อยระดับ option flag ของ Champ (~25 ตัว) และขอบเขตคลังต่อ role; เส้นทางเขียน GL ของ ERP (`post*ToGL`) ยังไม่ผ่านสิทธิ์ GL — ติดตามใน `docs/kms/05-transaction-sales-purchase.md` §4
