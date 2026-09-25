---
date: 2026-09-25
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, go, gl, budget, postgres]
---

# งบประมาณรายเดือนต่อบัญชี (Champ 5500) + รายงานเปรียบเทียบงบประมาณกับยอดจริง (Champ 5530)

## Context

- **Champ (ต้นแบบ, อ่านอย่างเดียว `D:\project-champ`)**
  - เมนู `menuconfig.xml:564` id 5500 resource 100003 "กำหนดงบประมาณ", `:580` id 5525 resource 104012 "รายงานรายละเอียดงบประมาณ", `:593` id 5530 resource 104017 "รายงานเปรียบเที่ยบงบประมาณ"
  - ตาราง `BCGLBudget` (`champ/champ/Script/SQLSERVER_Script.sql:9207–9234`): แถวละ 1 รหัสงบ = 1 บัญชี, `StartDate`/`StopDate`, `Amount MONEY`, มิติ Branch/Depart/Project/Allocate/Job/Part/Side, `Remark`, `Status` (combo ปิด/เปิด ใน `GLFrmBudget.cpp:212–218` — ไม่มีขั้นอนุมัติ ไม่ล็อกการแก้)
  - จอ `GLFrmBudget.cpp:130–171`: grid แถวละงบ (รหัส ชื่อ บัญชี มิติ วันเริ่ม วันสิ้นสุด จำนวน สถานะ หมายเหตุ)
  - รายงานเปรียบเทียบ `champ-report/BC5REP_GLFS/GLRepBudgetCompView.cpp`:
    - เรียงตามรหัสบัญชี รวมย่อยต่อบัญชี + รวมทั้งหมด (`:252–282`, `:398–414`)
    - ยอดจริง = `SUM(DEBIT)-SUM(CREDIT)` ของรายการที่ผ่านบัญชีแล้ว (`IsConfirm = 1`) ของบัญชีนั้นในช่วงวันที่ของงบ แล้วเอาค่าสัมบูรณ์ (`GetPeriodAmount` `:444–537`, abs ที่ `:536`)
    - มิติกรองเฉพาะเมื่อบริษัทเปิดแยกมิตินั้น: ตรงรหัส หรือ `IS NULL` (`:286–360`)
    - ผลต่าง = งบ − ยอดจริง แสดงวงเล็บเมื่อติดลบ (`:377–387`)
  - รายงานรายละเอียด `GLRepBudgetView.cpp`: รายการงบเรียงตามบัญชี + จำนวนและยอดรวมต่อบัญชี
- **`mydocs/`** ไม่มีนิยามงบประมาณ (ค้น `budget`/`งบประมาณ` ทั้งโฟลเดอร์ไม่พบ); `mydocs/datamodels/gl/period.sql` กำหนดงวดบัญชี `period_no` 1–12 ต่อปี — ใช้เป็นหน่วยงวดของงบรายเดือน
- **BC ก่อนงานนี้**: งบเป็น master ทั่วไปใน `gl_records` (`kind='budgets'`: บัญชีเดียว + ยอดเดียว + วันเริ่ม/สิ้นสุด) ผ่านฟอร์ม master กลาง `/gl/budget`; รายงาน `budgetcomparison` รวมยอดตามบัญชี ไม่แยกมิติ และคิดยอดจริงด้วยการเช็ค `accounttype='income'`
- ไม่มีข้อเท็จจริงกฎหมาย/ภาษีไทยเกี่ยวข้อง จึงไม่ต้องลงทะเบียนใน `docs/kms/21-thai-tax-form-references.md`

## Decision

1. **ตารางของตัวเอง (forward-only, backend สร้างเองด้วย `CREATE ... IF NOT EXISTS`)** — `backend/internal/generalledger/budget.sql`, โหลดใน `EnsureSchema` (`postgres.go:73`)
   - `gl_budgets` หัวงบ: `company, code` (PK), `name`, `fiscal_year`, `branch_code`/`department_code`/`project_code` (`''` = ทุกสาขา/แผนก/โครงการ), `status` `open|closed` (ธง Champ), `remark`, `version`, ผู้สร้าง/แก้ + เวลา
   - `gl_budget_lines` ยอดรายเดือน: `(company, budget_code, account_code, period_no 1–12)` PK, `amount numeric(18,2) >= 0`, FK → `gl_budgets` `ON DELETE CASCADE`
   - งวด 1 = เดือนของวันเริ่มปีบัญชี; ปีบัญชีสั้นกว่า 12 เดือนใส่ยอดได้เฉพาะงวดที่มีอยู่ (`fiscalPeriodStarts` `budgets.go:157`)
2. **คำสั่งผ่าน `Execute` เหมือนคำสั่ง GL อื่น** — requestid ซ้ำคืนผลเดิม, company lock, ตรวจ version, บันทึก audit ใน `gl_events` (projection ข้าม `kind='budgets'` ที่ `postgres.go:169`)
   - `resource: "budgets"`, payload `budget` = `{code,name,fiscalyear,branchcode,departmentcode,projectcode,status,remark,lines:[{accountcode,periods:[12 decimal strings]}]}`
   - `create` / `update` (แทนที่บรรทัดทั้งชุด, เปลี่ยนรหัสไม่ได้) / `delete` — `mutateBudget` `budgets.go:245`
   - ตรวจ (`validateBudget` `budgets.go:175`): ชื่อ ≤ 200 ตัวอักษร, สถานะ open/closed, ปีบัญชีมีจริง, บัญชีมีจริง + ลงรายการได้ + ใช้งานอยู่ + ไม่ซ้ำในงบเดียว, 12 งวด, ไม่ติดลบ, ทศนิยม ≤ 2, ต่ำกว่า 10^16; session ระดับสาขาทำงบได้เฉพาะสาขาตัวเอง; session ระดับบริษัทที่ระบุสาขาต้องเป็นสาขาที่เปิดใช้งานในทะเบียน (`checkBudgetBranch` `httpapi/branch.go:42`, เรียกที่ `httpapi/http.go:655`)
   - `spread` = คำนวณแบ่งยอดทั้งปีเป็น 12 งวด (ตัดทศนิยม 2 ตำแหน่ง เศษไปงวดสุดท้าย รวมกลับได้เท่าเดิม) ไม่บันทึกอะไร ต้องการแค่สิทธิ์เปิดจอ `gl-budget` (`SpreadAnnual` `budgets.go:95`, `spreadBudget` `budgets.go:120`) — ให้คำนวณที่ backend ตามกฎ backend-first
   - อ่าน: `GET /gl/v2/budgets` (ค้นรหัส/ชื่อ/หมายเหตุ, `ORDER BY fiscal_year DESC, code LIMIT/OFFSET`, กรองสาขาตาม session) และ `GET /gl/v2/budgets/{code}` (บรรทัด 12 งวด + รวมต่อบัญชี + รวมทั้งงบ) — `listBudgets` `budgets.go:423`, `readBudget` `budgets.go:385`
   - สิทธิ์: จอ `gl-budget` + action create/update/delete เหมือน master อื่น (`resourceScreens` ใน `httpapi/http.go`)
3. **รายงาน `budgetcomparison`** (`reports.go:437`)
   - แถว = งบ × บัญชี เรียงตามบัญชีแล้วรหัสงบ (ลำดับแบบ Champ)
   - คอลัมน์: บัญชี, ชื่อ, หมวด, รหัส/ชื่อ/สถานะงบ, สาขา/แผนก/โครงการของงบ, งบประมาณ, ใช้จริง, ผลต่าง (งบ − จริง), ร้อยละที่ใช้; รวมท้ายรายงาน งบ/จริง/ผลต่าง
   - ยอดงบ = งวดที่ "วันเริ่มงวด" อยู่ในช่วงรายงาน (`budgetPeriodsInRange` `reports.go:498`)
   - ยอดจริง = บรรทัด `gl_lines` (ผ่านบัญชีแล้วเท่านั้น) ของบัญชีนั้นในช่วงรายงาน ไม่รวมยอดยกมาและรายการปิดบัญชี ภายในมิติของงบ (`''` = ไม่กรอง); เครื่องหมายตาม `normal_balance` ของบรรทัด (ด้านเครดิต = เครดิต − เดบิต) — เป็นคุณสมบัติของผังบัญชี ไม่ใช้รหัสบัญชีเป็นเงื่อนไข
   - ตัวกรองเพิ่ม `budgetcode`; รายงานกรองสาขาแล้วจะเห็นเฉพาะงบของสาขานั้น (งบทุกสาขาไม่ถูกเทียบกับยอดของสาขาเดียว)
4. **Frontend**
   - BFF `/api/gl/*` รับ payload `budget`, action `spread` (เฉพาะ `budgets`), query `budgetcode` และตรวจว่า `periods[]`/`total` เป็นข้อความทศนิยม
   - จอ `/gl/budget` ขึ้น "ยังไม่พร้อม" (`pendingRoutes` ใน `frontend/src/lib/menu-screen-status.ts`) เพราะฟอร์ม master เดิมส่ง `master` ยอดเดียวซึ่ง API ใหม่ไม่รับแล้ว; ตัด `budgets` ออกจากฟอร์ม master กลาง (`gl-masters.tsx`) และ `masterRoutes`
5. **`cmd/glseed`** สร้างงบ 4 รายการเป็นรายเดือน (ยอดทั้งปีผ่าน `SpreadAnnual`) ของสาขาที่ seed โดยบัญชีมาจากการค้นตามประเภท + ชื่อ (ไม่ใช้รหัสตายตัว)

## ต่างจาก Champ โดยตั้งใจ / สิ่งที่ไม่ทำ

| Champ | BC | เหตุผล |
|---|---|---|
| วันเริ่ม/สิ้นสุดอิสระต่อแถว + ยอดเดียว | 12 งวดรายเดือนต่อบัญชีในปีบัญชี | งานนี้คืองบรายเดือน; งวดตาม `mydocs/.../period.sql`; ช่วงวันที่ = ผลรวมเดือนที่ครอบคลุม |
| 1 รหัสงบ = 1 บัญชี | 1 รหัสงบ มีหลายบัญชี | กรอกงบทั้งชุดในเอกสารเดียว; รายงานยังแยกแถวต่อบัญชีเหมือน Champ |
| มิติ Allocate/Job/Part/Side | ไม่ทำ | `gl_lines` ของ BC มีแค่สาขา/แผนก/โครงการ ไม่มีมิติอื่นให้เทียบ |
| มิติว่าง = `IS NULL` (เมื่อเปิดแยกมิติ) | มิติว่าง = ทุกค่า | งบระดับบริษัทไม่ต้องแยกมิติ; ถ้าต้องการเฉพาะให้ระบุรหัส |
| ยอดจริง = ค่าสัมบูรณ์ของ เดบิต − เครดิต | ตามด้านปกติของบัญชี | ค่าสัมบูรณ์ซ่อนยอดกลับด้าน (เช่น ค่าใช้จ่ายสุทธิด้านเครดิต) ทำให้อ่านเป็นการใช้งบ |
| รวมย่อยต่อบัญชี | มีเฉพาะรวมทั้งรายงาน | framework รายงาน GL รวมท้ายรายงานเท่านั้น; แถวเรียงตามบัญชีอ่านต่อกันได้ |
| กรองช่วงรหัส จาก–ถึง | กรองค่าเดียว (บัญชี/งบ/สาขา/แผนก/โครงการ) | เหมือนรายงาน GL อื่นใน BC |
| รายงานรายละเอียด 5525 | ไม่มีรายงานแยก | ข้อมูลเดียวกันได้จาก `GET /gl/v2/budgets` + `/{code}`; BC ยังไม่มีเมนู 5525 (`docs/kms/snippets/champ-menu-source-audit.md:395` "ตรวจเทียบต่อ") — เพิ่มเมนูต้องให้ลุงจืดตัดสิน |
| ไม่มี | ไม่มีขั้นอนุมัติ/ล็อกงบ | WIP เดิมมี draft/approved + approve/unapprove — ตัดออกตามกฎ Champ parity เหลือธงเปิด/ปิดแบบ Champ |
| ไม่มี | ไม่มีคอลัมน์ "เป็นไปตามงบ/ไม่เป็นไปตามงบ" | WIP เดิมมี — ตัดออก (Champ ไม่มี และความหมายกำกวมกับบัญชีงบดุล) |

## Consequences

- งบเดิมที่เป็น master ใน `gl_records` (`kind='budgets'`) ไม่ถูกอ่านอีก — ข้อมูลช่วง dev ทิ้งได้ตามกฎเดินหน้าอย่างเดียว ไม่ทำ migration
- ผู้ใช้ยังกรอกงบผ่านจอไม่ได้จนกว่าจะสร้างจองบรายเดือน (ตาราง บัญชี × 12 เดือน + ปุ่มแบ่งยอดทั้งปีที่เรียก `spread`) — งานถัดไป
- ช่วงรายงานที่ไม่ตรงต้นเดือน: งบนับทั้งงวดเมื่อวันเริ่มงวดอยู่ในช่วง แต่ยอดจริงนับตามวันที่จริง — ควรเลือกช่วงตามเดือนเต็ม
- ทดสอบ: `budgets_test.go` (แบ่งยอด, งวดปีสั้น, สถานะ), `budgets_integration_test.go` (CRUD + ตรวจ PostgreSQL ทีละขั้น, แยกสาขา, รายงานเทียบยอดจริงรายได้/ค่าใช้จ่าย, requestid ซ้ำ), `httpapi/branch_test.go` (`TestCheckBudgetBranch`), BFF `frontend/src/app/api/gl/[...glPath]/route.test.ts`
