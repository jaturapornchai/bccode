---
date: 2026-09-25
severity: high
component: [backend]
tags: [bc-account, go, postgres, fixed-asset, depreciation]
fixed: true
---

# Symptom

แก้ **อัตราค่าเสื่อมราคา (%)**, **% ค่าเสื่อมปีแรก** หรือ **ค่าเสื่อมสะสมยกมา** ของสินทรัพย์ในจอทะเบียนสินทรัพย์ (`frontend/src/app/asset/fixed-assets-screen.tsx` ช่องอัตรา ~บรรทัด 624, ช่อง % ปีแรก ~643) แล้วกดบันทึก → ระบบตอบ "แก้ไขสินทรัพย์เรียบร้อยแล้ว" แต่ตารางค่าเสื่อมที่ยังไม่ผ่านรายการ (`fa_records` kind `depreciations`) ยังเป็นค่าเดิม งวดที่ผ่านรายการเข้า GL ครั้งถัดไปจึงใช้ค่าเสื่อมตามอัตราเก่า

พิสูจน์ด้วยเทสต์ (PostgreSQL ชั่วคราว) กับโค้ดก่อนแก้: สินทรัพย์ทุน 120,000 อัตรา 20% (60 งวด) แก้อัตราเป็น 25% → ยังเก็บ 60 งวด ต้องเป็น 49; แก้ % ปีแรกเป็น 40 → 60 งวด ต้องเป็น 36; ค่าเสื่อมสะสมยกมา 24,000 → 60 งวด ต้องเป็น 48

# Root Cause

- `Store.UpdateAsset` (`backend/internal/fixedasset/store.go`) คำนวณตารางใหม่เฉพาะเมื่อ ราคาทุน / อายุการใช้งาน / ราคาซาก / วันเริ่มคิด เปลี่ยน — แต่ `Calculator.CalculateSchedule` (`calculator.go`) อ่านอีก 3 ช่องด้วย คือ `deprecpercent`, `firstyearpercent`, `beginaccumdeprec` ที่ไม่อยู่ในเงื่อนไข
- จอส่งแค่ `action: "update"` (`handleSaveAsset` ใน `fixed-assets-screen.tsx`) ไม่เรียก `recalculate` ตามหลัง และจอไม่มีปุ่มคำนวณใหม่ จึงไม่มีทางให้ผู้ใช้แก้ตารางเองได้
- ช่อง `method` ไม่ถูกใช้โดยตัวคำนวณ (คำนวณแบบเส้นตรงเสมอ) จึงไม่นับเป็นช่องที่ต้องคำนวณใหม่

# Fix

- เพิ่ม `scheduleInputsChanged(old, asset)` ใน `store.go` ครอบคลุมทุกช่องที่ `CalculateSchedule` อ่าน: `cost`, `scrapvalue`, `deprecpercent`, `firstyearpercent`, `beginaccumdeprec`, `usefullifeyears`, `startcalcdate` (ตัวเลขเทียบด้วยค่า decimal: "20" = "20.00" ไม่นับว่าเปลี่ยน)
- ถ้าเปลี่ยน → เรียก `recalculateSchedule` ตัวเดียวกับคำสั่ง `recalculate` ภายใน transaction เดียวกับการบันทึกสินทรัพย์ (`sql.Tx`) — ลบตารางเดิมแล้วสร้างใหม่ทั้งหมดจากสินทรัพย์ที่แก้แล้ว
- **มีงวดผ่านรายการเข้า GL แล้ว = ปฏิเสธ** (แก้รอบ review 2026-09-25): `recalculateSchedule` คำนวณใหม่ตั้งแต่ `beginaccumdeprec` จึงต่อจากยอดที่ผ่านรายการจริงไม่ได้ — พิสูจน์แล้ว (ทุน 120,000 ซาก 1 อัตรา 20%): ผ่านงวด 1 แล้วแก้ % ปีแรกเป็น 40 → ค่าเสื่อมตลอดอายุรวม 71,999.00 แทน 119,999.00; ผ่าน ม.ค.–มิ.ย. แล้วแก้เป็น 10%/10 ปี → รวม 125,949.65 เกินราคาทุน; แก้เป็น 25%/4 ปี → ขาด 2,975.29 และ `PostDisposal` อ่าน `AccumDeprec` ของแถวล่าสุด ทำให้กำไร/ขาดทุนจากการจำหน่ายผิดเท่ากัน จึงให้ทั้ง `UpdateAsset` (เมื่อช่องคำนวณเปลี่ยน) และคำสั่ง `recalculate` คืน `gl.UserError` code `fa_schedule_posted` (HTTP 409, ข้อความ key `gl_err_fa_schedule_posted` ใน `languages.tsv`) และ rollback ทั้งรายการ — ผู้ใช้ต้องยกเลิกการผ่านรายการค่าเสื่อมของสินทรัพย์นั้นก่อน (`reverse-gl`) แล้วแก้ใหม่
- `httpapi/http.go` คำสั่ง `assets/update` และ `assets/recalculate` ส่ง error ผ่าน `failure()` เพื่อคง code + 409 ไว้ (เดิม `fail()` ทิ้ง code เป็น "error" 400)
- แก้ชื่อ / สถานที่ / หมายเหตุ / รหัสบัญชี ฯลฯ ไม่แตะตารางค่าเสื่อม และบันทึกได้แม้มีงวดผ่านรายการแล้ว

# Regression Test

`backend/internal/fixedasset/store_update_recalc_test.go` (ต้องมี `BC_GL_TEST_POSTGRES_DSN`; ไม่มีจะ skip):

- `TestUpdateAssetRegeneratesScheduleWhenInputsChange` — แก้ทีละช่องทั้ง 7 ช่อง แล้วงวดที่เก็บต้องเท่ากับ `CalculateSchedule` ของสินทรัพย์ที่แก้แล้ว + `assertScheduleTotal`: ค่าเสื่อมสะสมแต่ละงวดต่อจากงวดก่อน และรวมทั้งอายุ = ทุน − ซาก − ค่าเสื่อมสะสมยกมา (3 ช่อง `deprecpercent`/`firstyearpercent`/`beginaccumdeprec` fail กับโค้ดเดิม)
- `TestUpdateAssetNonScheduleEditKeepsScheduleRows` — แก้ชื่อ/สถานที่/หมายเหตุ + อัตราเดิมเขียนต่างรูป ("20" แทน "20.00") แถวค่าเสื่อมต้องเหมือนเดิมทุกไบต์ (id, คอลัมน์ `updated_at`, payload)
- `TestScheduleRebuildRefusedWhilePeriodPosted` — ผ่านรายการงวด 1 แล้ว: แก้ทั้ง 7 ช่องทีละช่อง และสั่ง recalculate ต้องได้ `fa_schedule_posted` 409, แถวค่าเสื่อมเหมือนเดิมทุกไบต์, version สินทรัพย์ไม่ขยับ; แก้ชื่อ/สถานที่/หมายเหตุยังบันทึกได้; หลัง `ReverseDepreciation` แก้อัตราได้และตารางใหม่ผ่าน `assertScheduleTotal` (fail กับโค้ดก่อนรอบ review: update/recalculate ถูกรับไว้)

# Open (ต้องให้ลุงจืดตัดสิน)

- ตอนนี้เลือกทาง "ปลอดภัยขั้นต่ำ" คือห้ามแก้ช่องคำนวณเมื่อมีงวดผ่านรายการแล้ว — ยังไม่ได้ยืนยันว่า Champ (`D:project-champchampchampDepreciation.cpp` / `FAFrmAssetsMaster.cpp`) ทำอย่างไรเมื่อแก้สินทรัพย์หลังผ่านรายการ ทางเลือกอีกทางคือคำนวณเฉพาะงวดที่ยังไม่ผ่านรายการต่อจากงวดล่าสุดที่ผ่านแล้ว (เริ่มจาก `AccumDeprec` ของงวดนั้น ไม่บวกค่าเสื่อมปีแรกซ้ำ) — ต้องให้ลุงจืดตัดสินตาม Champ ก่อนเปลี่ยน
- จอ `fixed-assets-screen.tsx` ยังเปิดให้แก้ช่องเหล่านี้ได้ตามปกติ ผู้ใช้จะเห็นข้อความปฏิเสธหลังกดบันทึก (ยังไม่ได้ล็อกช่องล่วงหน้าบนจอ)
- FA httpapi ยังไม่แปลข้อความตามภาษา (ส่งข้อความไทยของ `gl.UserError` เสมอ) — แถว `gl_err_fa_schedule_posted` เตรียมไว้สำหรับตอนเปิดแปล
