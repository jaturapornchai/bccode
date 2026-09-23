---
date: 2026-09-23
severity: high
component: [backend, db]
tags: [bc-account, postgres, tax, vat]
fixed: false
---

# Symptom

รายงานภาษีขาย/ภาษีซื้อ (`POST /api/report/tax/vat-register`) และสรุป ภ.พ.30 (`POST /api/report/tax/pp30-summary`) บน production ตอบ 500 `QUERY_ERROR` ทุก holding — จอภาษีมูลค่าเพิ่มจึงขึ้น "โหลดข้อมูลไม่สำเร็จ" มาตลอด

## Root Cause

1. handler ต่อฐาน holding (`mypg.PgSqlFastConnect(holdingCode)` — `backend/internal/goapi/handlers/tax_report.go`) แต่ query อ่าน `public.saleinvoicetransaction` / `public.purchasetransaction` / `public.debtor` / `public.creditor` ซึ่ง **ไม่มีในฐาน holding ใดเลย** — ตรวจ prod 2026-09-23: มีเฉพาะใน `bcai_projection` (ของค้างยุค Kafka projection) และมี 0 แถว
2. query รายการใช้คอลัมน์ `m.name0` แต่ DDL จริงของ `debtor`/`creditor` มี `names` ไม่มี `name0`; ฝั่งขายยัง join `debtor` ด้วย `t.creditorcode`
3. ตาราง ERP มี `holdingcode` แต่ query ไม่กรองบริษัท (company) — ถ้าวันหนึ่งตารางมีข้อมูล รายงานจะรวมทุกบริษัท

## Fix

ยังไม่แก้ (อยู่นอกขอบเขตงานแปลงยอดเงินเป็น decimal 2026-09-23) — ทางที่ถูกคืออ่านภาษีมูลค่าเพิ่มจากบัญชีแยกประเภทที่ผ่านรายการแล้วตามสเปก `mydocs/datamodels/gl/vat.sql` (งาน GL migration) แทนตารางเอกสาร ERP แล้วกรองตาม company ที่ผู้ใช้เลือก

ระหว่างรอ (2026-09-23 รอบถอด Mongo): เมนู ภาษีซื้อ/ภาษีขาย/ภ.พ.30/ภ.พ.36 ขึ้น "รอเชื่อมข้อมูล" แล้ว (`LIVE_TAX_FORMS` ใน `frontend/src/lib/menu-screen-status.ts` เหลือเฉพาะ ภ.ง.ด./50 ทวิ ที่อ่านจาก GL จริง) เพราะลุงจืดสั่งให้ผู้ใช้ GL อย่างเดียวใช้ได้เลย และจอเอกสาร ERP ที่เคยสร้างข้อมูลลงตารางเหล่านี้ถูกปลดไปพร้อม MongoDB

สิ่งที่แก้แล้วในรอบนี้: ยอดเงินทุกช่องคำนวณด้วย decimal ปัดต่อใบใน SQL (`moneySQL`) และส่งเป็น string — พิสูจน์กับ DDL จริงของ `saleinvoicetransaction` (double precision) ว่า 0.1/100.005 ได้ยอดรวม 100.11 ไม่มีเศษ float

## Regression Test

- `TestTaxVatQueriesRunOnRealSchema` (`backend/internal/goapi/handlers/tax_withholding_integration_test.go`, build tag `integration`) — ยิง query ยอดรวมกับ DDL จริง (dump schema จาก `bcai_projection`); เมื่อย้ายไปอ่านจาก GL ต้องเปลี่ยน test นี้ให้ยิงกับฐาน holding ที่ restore มาจริง ไม่ใช่ DDL ที่ยกมาแยก
- หลังแก้: smoke test บน prod ต้องได้ 200 + `summary` จริงจาก holding `rungrueng` ไม่ใช่แค่ดู HTTP status
