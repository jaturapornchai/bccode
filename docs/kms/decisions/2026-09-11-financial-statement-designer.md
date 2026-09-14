---
date: 2026-09-11
status: implemented
tags: [bc-account, general-ledger, financial-statement-designer, menu]
---

# สถาปัตยกรรมและระบบออกแบบงบการเงิน (Financial Statement Designer)

## บริบทและโจทย์ความต้องการ (Context & Requirement)

ลุงจืดได้สั่งการ:
> "เพิ่มเมนู ออกแบบงบการเงิน สามารถเปลี่ยน font ฯลฯ
> รองรับงบต่างๆ เช่น งบดุล งบกำไรขาดทุน ฯลฯ
> ให้ผู้ใช้สามารถสร้างเองได้ไม่จำกัด 
> คิดให้ด้วย"

ในระบบ ERP มาตรฐาน งบการเงินแต่ละองค์กรมีความต้องการจัดรูปแบบที่แตกต่างกัน ทั้งตามข้อกำหนดของ DBD/สรรพากร, รายงานสำหรับผู้บริหาร, รายงานสำหรับธนาคาร หรือรูปแบบเฉพาะกิจการ การมีระบบ Financial Statement Designer ช่วยให้ผู้ใช้สามารถกำหนดผังงบ ผูกรหัสบัญชี ใส่สูตรคำนวณ ปรับแต่งฟอนต์ และสั่งพิมพ์หรือส่งออก CSV ได้อย่างอิสระและไม่จำกัด

## สถาปัตยกรรมการออกแบบ (Architectural Design)

### 1. Go Backend (CQRS / Event Sourcing / Master-Detail Storage)
- **Collection / Table Mapping**:
  - MongoDB: `gl_statement_templates` (`MasterCollections["statement-templates"]`)
  - PostgreSQL: โปรเจกต์ลง `gl_records` (`recordtype = "statement-templates"`)
- **Data Models**:
  - `StatementGlobalStyle`: `FontFamily`, `FontSize`, `LineHeight`, `BorderColor`, `BorderWeight`, `IsCompact`
  - `StatementStyle`: `Bold`, `Italic`, `Underline` (`none` | `single` | `double`), `Align` (`left` | `center` | `right`), `Indent` (0–5), `Color`
  - `StatementRow`: `RowID`, `RowNo` (เลขอ้างอิงสูตร), `RowType` (`header` | `account` | `total` | `text` | `blank`), `Description`, `AccountCodes` (`[]string`), `Formula` (สูตรคณิตศาสตร์), `Style`, `IsHidden`
  - `Master`: เพิ่มฟิลด์ `StatementType` (`balance-sheet` | `profit-loss` | `cost-of-goods` | `cash-flow` | `custom`), `GlobalStyle`, `Rows`

### 2. Live Calculation Engine (เครื่องคำนวณยอดและสูตรคณิตศาสตร์)
- รองรับการคำนวณผลลัพธ์จริงแบบ Live โดยดึงยอดเงินจากงบทดลอง (`/api/gl/reports/trialbalance`) ในงวด/ปีบัญชีที่เลือก
- **การประเมินผลสูตรคำนวณ (`evaluateStatementFormula`)**:
  - อ้างอิงตัวแปรแถว: `R10 + R20 - R30`
  - รองรับช่วงผลรวม: `SUM(R10:R50)`
  - ป้องกัน Circular Dependency ด้วย `visited` Set
  - คำนวณด้วย Recursive Descent Parser บน Fixed-Point BigInt (สเกล $10^8$) เพื่อความถูกต้องระดับทศนิยม ปราศจาก Floating Point Error

### 3. Frontend Designer Workbench (`gl-statement-designer.tsx`)
- **Dual Mode**:
  - **ออกแบบผังงบ**: เพิ่ม/ลบ/จัดลำดับแถว (เลื่อนขึ้น/ลง), สไตล์ (หนา/เอียง/ขีดเส้นใต้เดี่ยว-คู่/เยื้อง), ผูกรหัสบัญชีด้วย Modal Account Picker
  - **พรีวิวผลลัพธ์จริง**: สลับดูตัวอย่างงบจริงตามปีบัญชีที่เลือก พร้อมปุ่มสั่งพิมพ์ (Print CSS จัดหน้าสวยงาม) และดาวน์โหลด CSV
- **Global Font & Style Toolbar**:
  - ฟอนต์ภาษาไทยมาตรฐาน: Sarabun, Prompt, Kanit, Noto Sans Thai, Inter, Monospace
  - โหลด Google Fonts อัตโนมัติ (`ensureFontLoaded`)
  - ปรับขนาดตัวอักษร 13–18px, Line Height, ความเข้มของเส้นตาราง
- **Starter Templates**:
  - แม่แบบงบดุลตามมาตรฐาน DBD
  - แม่แบบงบกำไรขาดทุนจำแนกตามหน้าที่
  - แม่แบบงบต้นทุนการผลิตและต้นทุนขาย
  - แม่แบบงบกระแสเงินสด (วิธีทางอ้อม)

### 4. Menu & Integration Points
- Route: `/gl/statement-designer`
- เมนู ID: `financial-statement-designer` ในกลุ่ม `gl-reports`
- จำนวนเมนูรวม: 224 เมนู (GL 34 เมนู)
- รองรับทุกภาษาใน `backend/assets/language/languages.tsv`
