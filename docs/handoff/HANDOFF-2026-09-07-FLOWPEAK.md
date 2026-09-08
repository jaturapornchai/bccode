# HANDOFF: FlowAccount & PEAK Account Benchmark to Unified FlowPEAK Engine

> **เอกสารประวัติ — ไม่ใช่ขอบเขตงานปัจจุบัน:** วันที่ 2026-09-07 ลุงจืดกำหนดให้โฟลเดอร์ features-flowaccount-peak มีเฉพาะข้อมูล FlowAccount และ PEAK Account ไม่รวมเงินเดือนและสเปก BC/ระบบออกแบบเอง เอกสารที่ลิงก์ด้านล่างเป็นสำเนาประวัติซึ่งย้ายออกแล้ว ให้เริ่มจาก [ดัชนีปัจจุบัน](../features-flowaccount-peak/README.md) แทนคำสั่งทำงานต่อใน handoff นี้

**Date**: 2026-09-07  
**Author**: Gemini / Antigravity  
**Target Recipient**: OpenAI Codex (or Peer Agent)  
**Branch**: `dev`  
**Latest Git Commit**: `0b35f791` (ค่าตอนเขียน — **ล้าสมัยแล้ว** ตรวจของจริงด้วย `git log -1`)  

---

## 1. Executive Summary & Objective

ลุงจืดได้มอบหมายให้ศึกษาเจาะลึกระบบบัญชีชั้นนำของไทย 2 แพลตฟอร์ม:
1. **FlowAccount**: เด่นด้านความง่ายของผู้ใช้ (UX), AutoKey OCR, FlowPayroll, Bank Rules & Feeds, E-Commerce Integrations
2. **PEAK Account (PeakEngine)**: เด่นด้านความลึกเชิงบัญชีคู่ (Double-Entry GL), การคำนวณต้นทุนสินค้า FIFO Layering, PEAK Asset (ทะเบียนสินทรัพย์+คิดค่าเสื่อม), มิติข้อมูล (Tags/Classification), และการล็อกงวดบัญชี (`LockDate`)

และได้สั่งการให้ **"นำ FlowAccount ผสมผสานกับ PEAK Account โดยตรง"** กลายเป็นพิมพ์เขียวระบบบัญชีมาตรฐานเดียว (**FlowPEAK Unified Accounting Engine**) ภายใต้กฎเหล็กทางสถาปัตยกรรม 2-Tier Data Model

---

## 2. กฎเหล็กทางสถาปัตยกรรม (Core Architectural Rules)

1. **2-Tier Data Model (ตั้งโดยลุงจืด 2026-09-07)**:
   - **Tier 1 (MongoDB)**: Ultra-Lean Storage เก็บเฉพาะ ID, Code, Qty, Price; ไม่เก็บชื่อ/ข้อความซ้ำซ้อน; ไม่เก็บรูปภาพ Binary (ใช้ MinIO S3 URI + Thumbnail เท่านั้น)
   - **Tier 2 (PostgreSQL)**: Self-Contained Processing Engine มีรายละเอียดครบถ้วน (Enriched) ทั้งชื่อสินค้า, ข้อมูลลูกค้า, สาขา, หน่วยนับ โดย **ห้าม Query ข้ามกลับไปหา MongoDB เด็ดขาด (Zero Cross-DB Join)** เพื่อการออกงบการเงินและคำนวณ FIFO ด้วยความเร็วสูงสุด
2. **On-Demand Context Rule**:
   - เปิดอ่านเฉพาะไฟล์ที่จำเป็น ห้ามสแกนอ่านทั้งโฟลเดอร์เพื่อประหยัด Context Window

---

## 3. สรุปไฟล์เอกสารและพิมพ์เขียวที่สร้างไว้ (ย้ายไป `docs/archive/flowpeak-legacy-2026-09-07/` แล้วเมื่อ 2026-09-07)

| ไฟล์ | สาระสำคัญ |
| :--- | :--- |
| [`README.md`](../archive/flowpeak-legacy-2026-09-07/README.md) | ดัชนีและภาพรวมเปรียบเทียบ FlowAccount vs PEAK |
| [`01-flowaccount-in-depth.md`](../archive/flowpeak-legacy-2026-09-07/01-flowaccount-in-depth.md) | เจาะลึกฟีเจอร์ FlowAccount (Sales, Purchase, 5 Journals, AutoKey OCR, FlowPayroll, e-Tax, Open API) |
| [`02-peak-account-in-depth.md`](../archive/flowpeak-legacy-2026-09-07/02-peak-account-in-depth.md) | เจาะลึกฟีเจอร์ PEAK (AllInOne API, FIFO Costing, PEAK Asset, PEAK Tax, PEAK Board, Webhooks) |
| [`03-comparative-matrix.md`](../archive/flowpeak-legacy-2026-09-07/03-comparative-matrix.md) | ตารางเปรียบเทียบ Feature-by-Feature 10 หมวดหมู่ |
| [`04-bc-account-alignment-and-gaps.md`](../archive/flowpeak-legacy-2026-09-07/04-bc-account-alignment-and-gaps.md) | การวิเคราะห์จุดแข็ง-จุดที่ควรเสริมเมื่อนำมาใช้กับระบบ |
| [`05-flowaccount-data-model.md`](../archive/flowpeak-legacy-2026-09-07/05-flowaccount-data-model.md) | โครงสร้าง ER Diagram & JSON Schemas ของ FlowAccount |
| [`06-peak-account-data-model.md`](../archive/flowpeak-legacy-2026-09-07/06-peak-account-data-model.md) | โครงสร้าง ER Diagram, FIFO Layer Table & Payload ของ PEAK Account |
| [`07-bc-account-2tier-data-model.md`](../archive/flowpeak-legacy-2026-09-07/07-bc-account-2tier-data-model.md) | สถาปัตยกรรม 2-Tier: MongoDB (Lean) + PostgreSQL (Enriched DDL) + Sync Pipeline Matrix |
| [`08-unified-flowpeak-specification.md`](../archive/flowpeak-legacy-2026-09-07/08-unified-flowpeak-specification.md) | **พิมพ์เขียวระบบผสมผสาน FlowPEAK**: รวม 8 โมดูลหลัก + PostgreSQL DDL ฉบับสมบูรณ์ |

---

## 4. โครงสร้างระบบผสมผสาน (FlowPEAK Masterpiece Specification)

ระบบรวมศูนย์ 8 โมดูลหลัก:
1. **Sales & AR**: หน้าตากระดาษบิลแบบ FlowAccount + AllInOne API บันทึก/ชำระเงิน/ตัดสต็อกใน 1 รอบแบบ PEAK
2. **Purchase & AP**: AutoKey OCR สแกนบิลแบบ FlowAccount + ออกหนังสือรับรอง 50 ทวิ e-Withholding Tax แบบ PEAK
3. **Inventory & FIFO**: สต็อกหลายคลังแบบ FlowAccount + เครื่องยนต์ตัดต้นทุนตาม Lot (FIFO Cost Layers) แบบ PEAK
4. **Fixed Assets**: ทะเบียนสินทรัพย์ + เครื่องคิดค่าเสื่อมราคาเส้นตรงรายวัน/เดือน + ออก Draft JV สิ้นเดือน (PEAK Asset)
5. ~~**Integrated Payroll**~~ **(ยกเลิก 2026-09-08 — ลุงจืดสั่งไม่เอาระบบเงินเดือน ห้ามทำ รวม ภ.ง.ด.1 และไฟล์ประกันสังคม)**: จัดการเงินเดือน, คำนวณหัก ปกส., ภ.ง.ด.1, สลิปเงินเดือน, ไฟล์ส่งธนาคาร, ลงสมุดรายวัน PV/JV อัตโนมัติ (FlowPayroll)
6. **Banking & Reconciliation**: Bank Feed เชื่อม Statement ธนาคารไทย + Bank Rules จับคู่อัตโนมัติ (FlowAccount)
7. **Double-Entry General Ledger**: สมุดรายวัน 5 เล่ม (`UV`, `SV`, `RV`, `PV`, `JV`) + บัญชีแยกประเภท + ฟังก์ชันล็อกงวดบัญชี (`LockDate` จาก PEAK)
8. **Multi-dimensional Dimensions**: แท็กโครงการ (Project), แผนก (Department), สาขา (Branch) ออกงบ Project P&L ได้ทันที

---

## 5. แนะนำงานที่ส่งต่อให้ Codex ทำต่อ (Action Items for Codex)

Codex สามารถเลือกหยิบหัวข้อต่อไปนี้ไปลงมือทำต่อได้ทันที:

### Option A: ออกแบบ RESTful / OpenAPI 3.0 Specification ของ FlowPEAK
- กำหนด Schema request/response สำหรับ:
  - `POST /api/v1/transactions/sales/all-in-one` (Header, Items, Payments, WHT, Tags)
  - `POST /api/v1/expenses/ocr-scan` (AutoKey Payload $\rightarrow$ Form Draft)
  - `POST /api/v1/payroll/runs` (คำนวณและปิดยอดเงินเดือนประจำเดือน) — **ยกเลิก 2026-09-08 ไม่ต้องทำ**
  - `POST /api/v1/assets/{id}/depreciate` (รันคำนวณค่าเสื่อมราคาและสร้าง JV)

### Option B: สร้าง Go Code Scaffolding & SQL Migrations
- แปลง DDL ใน `docs/archive/flowpeak-legacy-2026-09-07/08-unified-flowpeak-specification.md` เป็นไฟล์ Migration (เช่น `backend/migrations/*.sql`)
- สร้าง Go Structs สำหรับ Domain Entities และ FIFO Engine Logic:
  - `InventoryCostLayerService`: ฟังก์ชัน `ConsumeFIFOLayers(ctx, productID, warehouseID, qty)` ใน PostgreSQL

### Option C: ออกแบบ Mockup หน้าจอ UI ตามกฎ UX คนไทย 40+
- ออกแบบหน้าจอ Sales Invoice Form ที่ผสานความเรียบง่ายของ FlowAccount เข้ากับกล่องระบุ Project Tag และ Payment / WHT ในตัว

---

## 6. Prompt สำหรับส่งให้ Codex ทันที (Ready-to-Paste Codex Prompt)

```text
คุณคือ Codex กำลังทำงานในโปรเจกต์ D:\bccode (branch: dev)
กรุณาอ่าน HANDOFF ล่าสุดที่ docs/handoff/HANDOFF-2026-09-08.md ก่อน (ไฟล์ HANDOFF-2026-09-07-FLOWPEAK.md เป็นเอกสารประวัติเท่านั้น)
เอกสาร docs/archive/flowpeak-legacy-2026-09-07/08-unified-flowpeak-specification.md เป็นสเปกเก่าที่เก็บถาวรแล้ว ใช้อ่านเป็นบริบทได้ ห้ามใช้เป็นข้อกำหนดปัจจุบัน (ขอบเขตปัจจุบันอยู่ที่ AGENTS.md + docs/kms/19-menu-coverage-flowaccount-peak.md)

สถานะปัจจุบัน:
- ได้ศึกษาและสร้างสเปก FlowAccount + PEAK รวมเป็น FlowPEAK แล้ว (สเปกย้ายไป docs/archive/flowpeak-legacy-2026-09-07/ เมื่อ 2026-09-07) — ตรวจ commit ปัจจุบันด้วย git log -1 อย่าอ้างตัวเลขในเอกสารนี้
- ยึดกฎ 2-Tier Data Model: MongoDB เก็บ Lean Data, PostgreSQL ทำการประมวลผลทั้งหมด (FIFO, GL, Tax, Reports) โดยห้าม Query ข้ามกลับไป MongoDB
- ระบบเงินเดือน (payroll), ภ.ง.ด.1 / ภ.ง.ด.1ก และไฟล์นำส่งเงินสมทบประกันสังคม อยู่นอกขอบเขตผลิตภัณฑ์ (ลุงจืดสั่ง 2026-09-08 ดู AGENTS.md) ห้ามทำ — ส่วน ภ.ง.ด.2 ยังอยู่ในขอบเขต
- ให้เลือกทำงานต่อในส่วน: [ระบุ Option A: OpenAPI / Option B: Go Backend Migrations & FIFO Service / Option C: UI Mockups]
```
