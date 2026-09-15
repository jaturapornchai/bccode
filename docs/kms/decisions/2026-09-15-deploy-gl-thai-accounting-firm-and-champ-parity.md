---
date: 2026-09-15
status: deployed
tags: [bc-account, deployment, general-ledger, accounting-firm, trial-balance, mongodb, postgresql, project-champ]
---

# Deploy ระบบบัญชีแยกประเภท วงจรบัญชีมาตรฐานไทย และแก้ไขระบบรายงาน สู่ Production (account.bcaicloud.com)

## วัตถุประสงค์และผลการปล่อยระบบ

ปล่อยรีลีส `r20260915-1` สู่ Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) วันที่ 15 กันยายน 2026 ผ่านระบบ Fast Streamed Zero-Disk Deployment ตามคำสั่งลุงจืด โดยครอบคลุมทั้งการทดสอบวงจรบัญชีครบวงจรแบบสำนักงานบัญชีไทย (CPA Full-Cycle), การตรวจสอบความสมบูรณ์ของฐานข้อมูล 2-Tier (MongoDB + PostgreSQL), และการแก้ไขข้อบกพร่องในระบบรายงานทางการเงิน (Financial Reporting Query Engine)

## สรุปรายการฟังก์ชันและข้อแก้ไขที่ปล่อยขึ้น Production

1. **การทดสอบวงจรบัญชีครบวงจรแบบสำนักงานบัญชีไทย (7 Phases E2E)**:
   - Phase 1: กำหนดปีบัญชี 2569 พร้อมเงื่อนไข Scale และการล็อกสถานะงวด
   - Phase 2: ผังบัญชีมาตรฐานไทย 5 หมวด (Assets, Liabilities, Equity, Revenue, Expenses) พร้อมระบบ Tree View, Parent-Child และ AllowPosting Guard
   - Phase 3: บันทึกยอดยกมาต้นงวด (Opening Balance) เดบิต 1,000,000 = เครดิต 1,000,000 บาท ดุลสมบูรณ์
   - Phase 4: สมุดรายวันเฉพาะ 5 เล่มตามมาตรฐานไทย:
     - `SV` (รายวันซื้อ): ซื้อเชื่อ + ภาษีซื้อ 7%
     - `UV` (รายวันขาย): ขายเชื่อ + ภาษีขาย 7%
     - `PV` (รายวันจ่าย): จ่ายเจ้าหนี้ + ค่าเช่าหัก ณ ที่จ่าย 3%
     - `RV` (รายวันรับ): รับชำระหนี้จากลูกหนี้การค้า
     - `JV` (รายวันทั่วไป): ตัดต้นทุนขายสินค้า (COGS แบบ Perpetual)
   - Phase 5: งบทดลอง (Trial Balance): ผลรวมเดบิต 1,849,000.00 = เครดิต 1,849,000.00 บาท (ผลต่าง 0.00 บาท)
   - Phase 6: งบกำไรขาดทุน (P&L): รายได้ 250,000 - ต้นทุน 80,000 = กำไรขั้นต้น 170,000 - ค่าเช่า 20,000 = กำไรสุทธิ 150,000.00 บาท
   - Phase 7: สมการงบดุล (Balance Sheet Equation Proof): สินทรัพย์ 1,168,100.00 = หนี้สิน 18,100.00 + ส่วนของเจ้าของและกำไรสะสม 1,150,000.00 บาท (ดุล 100% ผลต่าง 0.00 บาท)

2. **การแก้ไข Backend Reporting Query Engine (`reports.go`, `reports_operations.go`)**:
   - แก้ไข syntax error ในคิวรีงบกำไรขาดทุน (`profitLoss`) และงบแสดงฐานะการเงิน (`balanceSheet`) โดยเปลี่ยนจากการใช้ `FILTER(WHERE ...)` ภายใน `SUM()` ที่ผิดไวยากรณ์ SQL มาเป็นมาตรฐาน `SUM(CASE WHEN ... THEN ... ELSE 0 END)`
   - ปรับปรุงการเรียงลำดับ `ORDER BY` ในรายงานสมุดบัญชีแยกประเภท (`ledger`) และพยากรณ์เงินสด (`cashflowforecast`) ให้เรียงลำดับตามคอลัมน์จริงของตารางรายงาน
   - ส่งผลให้ Integration Tests ผ่านครบทั้ง 15 รายงาน (`ledger`, `trialbalance`, `workingpaper`, `pnl`, `balancesheet`, `annual-balances`, `daily-check`, `cashflow`, `cashflowforecast`, `project-pnl`, `dimensionpnl`, `projectsummary`, `dashboard`, `executivesummary`, `financialgraphs`)

3. **เทียบเคียงคุณสมบัติตามต้นแบบ `D:\project-champ`**:
   - รองรับไปป์ไลน์การลงบัญชีจากเอกสารซื้อขาย/การเงิน/เช็คลงสมุดรายวัน
   - รองรับ 7 มิติข้อมูล (Cost Allocation, Branch, Department, Project ฯลฯ)
   - รองรับกระดาษทำการ 8 ช่อง และการปิดงวดบัญชีโอนปิดหมวด 4-5 เข้าหมวด 3

## Production Artifacts & Release Details

| บริการ | Image | สถานะ |
|---|---|---|
| mainapi | `bcai-account-mainapi:r20260915-1` | Up (healthy) |
| worker | `bcai-account-mainapi:r20260915-1` | Up (healthy) |
| frontend | `bcai-account-frontend:r20260915-1` | Up (healthy) |

- **Release Tag**: `r20260915-1`
- **Release Directory**: `/opt/bcai-account/releases/r20260915-1`
- **Preflight Backups**:
  - `mongo.archive.gz` (สำรอง MongoDB ด้วย `--archive --gzip --oplog`)
  - `postgres-all.sql` (สำรอง PostgreSQL ด้วย `pg_dumpall`)
  - `runtime-config.tar.gz` (สำรองคอนฟิก `/etc/bcai-account`)
  - `release.env.before`: บันทึกสถานะ image ก่อนหน้า (`r20260914-2`)
- **Verification Evidence**:
  - Live HTTP Status: 200 OK บน `https://account.bcaicloud.com/`
  - Live Auth Guard: 401 Unauthorized บน `https://account.bcaicloud.com/api/gl/accounts`
  - Vitest Suite: ผ่าน 504/504 การทดสอบ (100%)
  - TypeScript Check: ผ่าน 0 errors
  - ESLint: ผ่าน 0 errors
