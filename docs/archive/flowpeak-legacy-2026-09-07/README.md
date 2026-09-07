# การศึกษาและเปรียบเทียบคุณสมบัติ FlowAccount vs PEAK Account สำหรับ BC Ai Account

> **ขอบเขตล่าสุด 2026-09-07:** รวมเมนู FlowAccount + PEAK แบบละเอียด **ไม่รวมเงินเดือนและ HR ที่เกี่ยวข้อง** ตามคำสั่งลุงจืด ให้เริ่มจาก [09 — รายการเมนูรวม](09-unified-menu-catalog.md): 18 หมวด 243 รายการความสามารถ พร้อมหลักฐานทางการและรายการรอยืนยัน เนื้อหาเงินเดือนในบทความเก่าเป็นประวัติ benchmark ไม่ใช่ขอบเขตพัฒนา

> รวบรวมและวิเคราะห์อย่างละเอียด: 2026-09-07  
> วัตถุประสงค์: เพื่อเป็นเกณฑ์มาตรฐาน (Benchmark) ในการออกแบบโมดูลงานบัญชี สต็อก ภาษี และ API ของระบบ **BC Ai Account** ให้ครอบคลุม เหนือกว่า และตอบโจทย์ธุรกิจไทย

---

## 1. บทนำและภาพรวมของทั้งสองโปรแกรม

ในตลาดซอฟต์แวร์บัญชีออนไลน์บนระบบคลาวด์ (Cloud Accounting) ของประเทศไทย **FlowAccount** และ **PEAK Account (PeakEngine)** ถือเป็น 2 ผู้นำหลักที่มีส่วนแบ่งการตลาดสูงสุดและมีแนวคิดการออกแบบระบบที่น่าศึกษา:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CLOUD ACCOUNTING THAILAND                        │
├──────────────────────────────────────┬──────────────────────────────────────┤
│             FLOWACCOUNT              │             PEAK ACCOUNT             │
├──────────────────────────────────────┼──────────────────────────────────────┤
│ ปรัชญา: "บัญชีออนไลน์ ใช้ง่ายสำหรับ SME" │ ปรัชญา: "ระบบบัญชีครบวงจรเพื่อการเติบโต" │
│ จุดเน้น: User-Friendly, รวดเร็ว, เจ้าของกิจการ│ จุดเน้น: ความลึกเชิงบัญชี, ผู้บริหาร, นักบัญชี │
│ UI: เรียบง่าย ออกแบบคล้ายกระดาษเอกสารจริง │ UI: เน้นฟังก์ชัน ควบคุมการลงบัญชี ป้ายแท็ก   │
│ AI: AutoKey (สแกนบิล/ใบเสร็จ)         │ AI: AI ตรวจสอบความถูกต้อง, แนะนำผังบัญชี │
│ จุดแข็ง: FlowPayroll, เชื่อม Shopee/Lazada │ จุดแข็ง: PEAK Board (BI), PEAK Asset, FIFO │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

---

## 2. โครงสร้างเอกสารในโฟลเดอร์นี้

เอกสารในโฟลเดอร์นี้แบ่งเป็น 9 บทความ เปิดเฉพาะเรื่องที่ทำ โดยใช้บทที่ 09 เป็นขอบเขตเมนูล่าสุด:

| ลำดับ | ไฟล์เอกสาร | รายละเอียดเนื้อหา |
| :---: | :--- | :--- |
| **01** | [`01-flowaccount-in-depth.md`](01-flowaccount-in-depth.md) | เจาะลึกคุณสมบัติ FlowAccount: ระบบขาย, ซื้อ, สต็อก, คลังสินค้า, สมุดรายวัน 5 เล่ม, AutoKey OCR, e-Tax Invoice, FlowPayroll และ Open API |
| **02** | [`02-peak-account-in-depth.md`](02-peak-account-in-depth.md) | เจาะลึกคุณสมบัติ PEAK Account (PeakEngine): เอกสาร All-In-One, ต้นทุน FIFO อัตโนมัติ, PEAK Asset (คิดค่าเสื่อม), PEAK Tax, PEAK Board (BI) และ Open API & Webhook |
| **03** | [`03-comparative-matrix.md`](03-comparative-matrix.md) | ตารางเปรียบเทียบคุณสมบัติต่อคุณสมบัติ (Feature-by-Feature Matrix) รวม 10 หมวดหมู่หลักแบบละเอียด |
| **04** | [`04-bc-account-alignment-and-gaps.md`](04-bc-account-alignment-and-gaps.md) | แผนที่การประยุกต์ใช้กับ **BC Ai Account**: การผสานกับสถาปัตยกรรม 2-Tier (MongoDB Storage + PostgreSQL Processing Engine), จุดที่ BC เหนือกว่า และจุดที่ควรพัฒนาเสริม |
| **05** | [`05-flowaccount-data-model.md`](05-flowaccount-data-model.md) | Data Model & Schema ของ FlowAccount: Entity-Relationship, JSON Payload, Master Data, Sales/Purchase, Banking, 5 Journals, Payroll |
| **06** | [`06-peak-account-data-model.md`](06-peak-account-data-model.md) | Data Model & Schema ของ PEAK Account: ER Diagram, FIFO Cost Layers, AllInOne Transaction, PEAK Asset Register, Tags/Dimensions, Period Lock |
| **07** | [`07-bc-account-2tier-data-model.md`](07-bc-account-2tier-data-model.md) | **BC Ai Account 2-Tier Architecture Blueprint**: MongoDB Lean Storage Layer + PostgreSQL Self-Contained Processing Engine (Zero Cross-DB Join), FIFO & GL DDL, CDC Pipeline |
| **08** | [`08-unified-flowpeak-specification.md`](08-unified-flowpeak-specification.md) | **FlowPEAK Unified Masterpiece Blueprint**: สเปกระบบและการผสมผสาน FlowAccount + PEAK Account เป็นระบบเดียว (UX ง่าย + AutoKey OCR + Payroll + FIFO Layers + ทะเบียนสินทรัพย์ + AllInOne + Double-Entry GL + Tags + DDL ครบวงจร) |
| **09** | [`09-unified-menu-catalog.md`](09-unified-menu-catalog.md) | เมนูรวมล่าสุด 18 หมวด 243 รายการ ไม่รวมเงินเดือน/HR แยกเมนู งาน รายงาน ตั้งค่า และการเชื่อมต่อ พร้อมแหล่งอ้างอิงและข้อจำกัด |

---

## 3. สรุปเปรียบเทียบสถาปัตยกรรมและกลุ่มผู้ใช้

| มิติการเปรียบเทียบ | FlowAccount | PEAK Account | โอกาสของ BC Ai Account |
| :--- | :--- | :--- | :--- |
| **กลุ่มเป้าหมายหลัก** | พ่อค้าแม่ค้าออนไลน์, ฟรีแลนซ์, Micro-SME, Startups เริ่มต้น | SME ขนาดกลาง-ใหญ่, บริษัทที่ต้องตรวจบัญชี, สำนักงานบัญชี | กิจการที่มีหลายสาขา/หลายบริษัท (Holding), ค้าปลีก-ส่ง, ธุรกิจวัสดุ/ผลิต |
| **การลงรายการบัญชี** | ซ่อนความซับซ้อน บันทึกลงสมุดรายวัน 5 เล่มเบื้องหลังอัตโนมัติ | แสดงการลงบัญชีชัดเจน ปรับปรุงเดบิต/เครดิตได้อิสระ ล็อคงวดบัญชีได้ | ใช้สถาปัตยกรรม 2-Tier: หน้าร้านบันทึกง่าย (Mongo) ประมวลผลลึกใน PG |
| **ระบบสินค้าและสต็อก** | สต็อกเฉลี่ย ไม่เน้นต้นทุนแบบล็อตหรือ FIFO ลึกซึ้ง | สต็อก FIFO แบบ Perpetual บันทึกต้นทุนขาย Real-time พร้อมคำนวณย้อนหลังได้ | สต็อกแยกคลัง/โซน/ชั้นวาง + คำนวณต้นทุน FIFO/Average บน PostgreSQL |
| **โครงสร้างสิทธิ์** | ระดับผู้ใช้งานพื้นฐาน (Admin, Staff, Accountant) | สิทธิ์ละเอียดตามโมดูลและสาขา | สิทธิ์ระดับ Holding / Company / Branch / Screen Actions ลึกที่สุด |
| **โมเดลการคิดราคา API** | มี Open API ให้เชื่อมต่อตามแพ็กเกจ | คิดค่าบริการตาม Transaction ที่ยิง API (Billable POST) | เปิด API แบบบูรณาการ ไม่เก็บค่า Transaction ซ้ำซ้อน |
