---
date: 2026-09-12
status: deployed
tags: [bc-account, ui, general-ledger, chart-of-accounts, tree-sort, search-icon, effective-level]
---

# แก้ไขการจัดเรียงผังบัญชีแบบต้นไม้และแก้ปัญหาไอคอนในช่องค้นหาโดนทับ

## บริบทและปัญหาที่พบ (Problem & Root Cause)

จากการทดสอบระบบจริงบน Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) ลุงจืดได้ส่งข้อผิดพลาดเข้ามา 2 จุด:
1. **ผังบัญชียังเรียงไม่ถูก**:
   - ในฐานข้อมูล MongoDB และ PostgreSQL ข้อมูลบัญชีตัวอย่างของกิจการ `demo` มีเฉพาะบัญชี `BM69-110101` ที่มีฟิลด์ `level: 2` ขณะที่บัญชีลูกอีก 32 บัญชีไม่มีฟิลด์ `level` (ถูกอ่านเป็น `level: 1`)
   - ค่าเริ่มต้นของ Dialog กรอง `onlyPosting: true` ทำให้บัญชีคุม (`BM69-1000`) ถูกซ่อนออกไป ส่งผลให้ `BM69-110101` ขึ้นเป็นบรรทัดแรกและเยื้องลอยๆ ใต้ความว่างเปล่า ขณะที่บัญชีลูกถัดไป (`BM69-110102`) ชิดซ้ายเป็น "ระดับ 1"
2. **ไอคอนโดนทับ (Search Icon Overlap)**:
   - ไอคอน `<Search />` และปุ่ม `<X />` ในช่องค้นหาขาด `top-1/2 -translate-y-1/2` และ `z-10`
   - คลาส `px-3` ในตัวแปร `control` ชนะคลาส `pr-9` ทำให้ข้อความหรือ placeholder ในช่องค้นหาซ้อนทับกับไอคอน

## การแก้ไขและการออกแบบ (Architecture & UI Pattern)

1. **การคำนวณระดับขั้นและจัดเรียงแบบต้นไม้ (`frontend/src/lib/general-ledger.ts`)**:
   - สร้างฟังก์ชัน `getEffectiveAccountLevel(account, accountsMap)`: หากบัญชีมี `parentaccountcode` จะคำนวณระดับจากบัญชีแม่ + 1 เสมอ แก้ปัญหาข้อมูล legacy ขาดฟิลด์ level
   - สร้างฟังก์ชัน `sortAccountsHierarchically(accounts)`:
     - เรียงลำดับ 5 หมวดบัญชีมาตรฐาน: สินทรัพย์ (1) $\to$ หนี้สิน (2) $\to$ ส่วนของเจ้าของ (3) $\to$ รายได้ (4) $\to$ ค่าใช้จ่าย (5)
     - นำบัญชีลูกไปต่อท้ายบัญชีแม่ตามสายสัมพันธ์โครงสร้างต้นไม้อย่างถูกต้อง
2. **ปรับปรุง `AccountSearchDialog` (`frontend/src/app/gl/account-search-dialog.tsx`)**:
   - กำหนด `onlyPosting` เริ่มต้นเป็น `false` เพื่อแสดงบัญชีคุมระดับ 1 ครบถ้วน
   - ในโหมดเลือกบัญชีเดี่ยว (`!all`): บัญชีคุมจะแสดง Badge "บัญชีคุม" และบล็อกการกด Enter หรือ Double-click ไม่ให้เลือกไปบันทึกรายการ
   - ปรับช่องค้นหาใช้ `pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 size-5 text-muted-foreground z-10` และ input ใช้ `!pl-11 !pr-10`
3. **ปรับปรุง `SearchInput` และ `AccountSelect` (`frontend/src/app/gl/gl-common.tsx`)**:
   - เพิ่ม Search icon ใน `SearchInput` พร้อม `!pl-9.5 !pr-9` และจัดกึ่งกลางด้วย `top-1/2 -translate-y-1/2 z-10`
   - `AccountSelect` ใช้ `sortAccountsHierarchically` และเยื้องข้อความใน `<select>` ตาม `effectiveLevel`
4. **อัปเดตฐานข้อมูล Production**:
   - ปรับปรุง MongoDB `bcai_account.chart_of_accounts`: บัญชีแม่ราก `level: 1` (5 รายการ), บัญชีย่อย `level: 2` (32 รายการ)
   - ปรับปรุง PostgreSQL `demo.gl_records`: บัญชีแม่ราก `level: 1` (5 รายการ), บัญชีย่อย `level: 2` (33 รายการ)
5. **การปล่อยระบบ Production (Release `r20260912-search-sort-1`)**:
   - Preflight Backups: `mongo.archive.gz` (71,811 bytes), `postgres-all.sql` (723,955 bytes), `runtime-config.tar.gz` (7,267 bytes)
   - Container Image: `bcai-account-frontend:r20260912-search-sort-1` และ `bcai-account-mainapi:r20260912-search-sort-1`
   - Live Chunk Verification: ตรวจพบโค้ดใหม่ใน chunk `3n8v2x3_82g2u.js` บน `https://account.bcaicloud.com/`
