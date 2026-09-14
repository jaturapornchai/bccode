---
date: 2026-09-10
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, frontend, menu, holding, tenancy, deduplication]
---

# ตัดกลุ่ม "บุคลากรและผู้ใช้งาน" ออกจากเมนูหลัก (กำหนดสิทธิ์และบุคลากรที่ระดับ Holding)

หักล้าง (supersede) ข้อความใน [2026-09-08-menu-parity-market-standard.md](2026-09-08-menu-parity-market-standard.md) ที่เคยเขียนว่า "กลับคำ... ย้ายเข้าเมนู... อย่าลบทิ้ง"

## Context

วันที่ 2026-09-08 มี AI ในรอบ menu parity พยายามไล่จับคู่ฟังก์ชันระบบบัญชีในตลาดให้ได้ 218–224 เมนู โดยนำเอา:
1. ทะเบียนพนักงาน (`/employee`)
2. ผู้ใช้งานระบบ (`/user`)
3. กลุ่มสิทธิ์การใช้งาน (`/permissiongroup`)
4. ประวัติการเข้าใช้งาน (`/useraccessaudit`)

ไปสร้างกลุ่มใหม่ชื่อ **"บุคลากรและผู้ใช้งาน"** (`id: organization-people`) ใต้หมวด "ข้อมูลหลัก" (`master`) ของเมนูหลักประจำวัน (`frontend/src/lib/menu-data.ts`) และเขียนกำกับไว้ใน ADR เดิมว่า "อย่าลบทิ้ง"

แต่ในโครงสร้างสถาปัตยกรรมจริงของ BC Ai Account:
- การบริหารจัดการผู้ใช้ พนักงาน สิทธิ์การเข้าถึง และการตรวจสอบ **ถูกจัดการที่ระดับ Holding / Workspace Wizard (`/holding`, `/workspace`)** อยู่แล้ว
- API ฝั่ง backend อยู่ใต้ `/holding/users`, `/holding/permission`, `/organization/role-permission`
- การนำมาวางไว้ในเมนูหลักประจำวันของสาขาทำให้เกิดความซ้ำซ้อน สับสน และผิดระดับชั้นการปกครองข้อมูล (Tenancy level)

วันที่ 2026-09-10 ลุงจืดตรวจพบและสั่งการ:
> *"อันนี้ซ้ำหรือไม่ ถ้าซ้ำลบออก เพราะกำหนดรหัส holding"*
> *"ทำอะไรเสร็จ พยายามจำด้วย รู้สึกว่า วนไปมา ของเก่ากลับมา เคยบอกให้ทำไปแล้ว"*

## Decision

1. **ตัดกลุ่ม "บุคลากรและผู้ใช้งาน" (`organization-people`) ออกจากเมนูหลัก (`frontend/src/lib/menu-data.ts`) อย่างถาวร** (ลด 4 เมนู จาก 170 เหลือ 166 รายการ)
2. **ขอบเขตการจัดการบุคลากรและสิทธิ์ผูกขาดที่ระดับ Holding เท่านั้น**:
   - `/employee`, `/user`, `/permissiongroup`, `/useraccessaudit` รวมถึงการเชิญผู้ใช้ จัดกลุ่มสิทธิ์ และดูประวัติ ต้องทำผ่านหน้า `/holding` และ `/workspace` (Wizard ขั้นตอน "คนในองค์กร", "สิทธิ์การใช้งาน", "ตรวจสอบ")
   - **ข้อห้ามเด็ดขาด**: ห้าม AI ตัวใดนำ `/employee`, `/user`, `/permissiongroup`, `/useraccessaudit` หรือกลุ่ม `organization-people` เพิ่มกลับเข้ามาใน `menu-data.ts` หรือเมนูหลักประจำวันของสาขาอีก
3. **เพิ่มกลไกป้องกันการวนซ้ำ (Zero Regression Guarantee)**:
   - อัปเดตข้อห้ามชัดเจนใน `AGENTS.md` (หัวข้อขอบเขตผลิตภัณฑ์)
   - ล็อกยอดเมนู 166 รายการใน `frontend/src/lib/menu-icons.test.ts`
   - เพิ่ม Automated Test ใน `frontend/src/lib/menu-data.test.ts` (`has no organization-people group in the master section`) เพื่อ fail ทันทีหากมีใครพยายามเพิ่มกลับเข้ามา
   - บันทึกใน `docs/kms/18-decisions-and-agreements.md` และปรับปรุง `19-menu-coverage-market-standard.md`

## Alternatives

- เก็บไว้ในเมนูหลักแต่กดแล้ว redirect ไปหน้า Holding: ปฏิเสธ เพราะทำให้เมนูสาขารกและปะปนเรื่องระดับองค์กรกับงานประจำวัน
- เปลี่ยนชื่อเป็น "ข้อมูลพนักงานสาขา": ปฏิเสธ เพราะระบบ BC Ai Account พนักงานและสิทธิ์ผูกที่ระดับ Holding/Company ไม่ใช่ entity ย่อยเฉพาะสาขา

## Consequences

- ✅ เมนูหลักหมวดข้อมูลหลัก (`master`) สะอาดเหลือ 8 กลุ่ม (30 รายการ) ไม่มีความซ้ำซ้อนกับระดับ Holding
- ✅ เมนูหลักรวมลดลงเหลือ 166 leaf items (รายงาน 54 รายการ รวม 220 รายการ)
- ✅ ป้องกันปัญหา AI ตัวถัดไปอ่านเจอ ADR เก่าแล้ววนลูปนำของเดิมกลับมาด้วย test และ ADR หักล้างฉบับนี้
