---
date: 2026-09-25
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, go, gl, tax, wht, postgres]
---

# ที่อยู่สำหรับภาษี (สำนักงานใหญ่) + โทรศัพท์ เก็บในทะเบียนบริษัทแบบแยกช่อง แล้วเติมหัวแบบภาษีและ 50 ทวิ ให้อัตโนมัติ

## Context

- หัวแบบ ภ.พ.30/36, ภ.ง.ด.50/51/53, ภ.ธ.40 และหนังสือรับรอง 50 ทวิ ต้องมีที่อยู่ของกิจการ แต่ทะเบียนบริษัท (`companies` ใน `bcai_projection`) ไม่มีที่อยู่เลย:
  - จอแบบยื่นยกที่อยู่มาจาก "ฉบับก่อน" (`copyProfile`) เท่านั้น แบบแรกของบริษัทจึงต้องพิมพ์ที่อยู่เองทุกครั้ง
  - 50 ทวิ ฝั่งผู้มีหน้าที่หักภาษี (บริษัทเรา) ไม่มีที่อยู่จากระบบ ผู้ใช้ต้องพิมพ์เองทุกใบ ถ้าเว้นว่าง backend ตอบ `wht_cert_payer_address_required`
- แหล่งทางการ (รหัสตาม `docs/kms/21-thai-tax-form-references.md` §1):
  - **RD-50TAWI-FORM**: ใต้ช่องที่อยู่ให้ระบุ "ชื่ออาคาร/หมู่บ้าน ห้องเลขที่ ชั้นที่ เลขที่ ตรอก/ซอย หมู่ที่ ถนน ตำบล/แขวง อำเภอ/เขต จังหวัด" ระบบใช้ลำดับนี้ต่อบรรทัดที่อยู่
  - **TH-BMA-ACT-2528**: พ.ร.บ.ระเบียบบริหารราชการกรุงเทพมหานคร พ.ศ. 2528
    - ม.7 แบ่งพื้นที่เป็นเขตและแขวง
    - ม.8 ให้กฎหมายที่อ้าง จังหวัด/อำเภอ/ตำบล หมายถึง กรุงเทพมหานคร/เขต/แขวง
    - ที่อยู่ในกรุงเทพฯ จึงใช้คำนำหน้า แขวง/เขต และไม่มีคำว่า "จังหวัด"
- **เทียบ Champ (`D:\project-champ`, อ่านอย่างเดียว)**:
  - `BCConfigurations` เก็บที่อยู่เป็นข้อความเดียว `Address VARCHAR(255)` / `AddressEng` / `Telephone` / `Fax` (`champ/champ/Script/SQLSERVER_Script.sql:664`, `:675–678`)
  - struct `champ-lib/BC5FRMWRK/companyinfo.h:33–36` เป็นระดับบริษัท ไม่แยกสาขา
- หัวแบบ rdform ใช้ที่อยู่ 13 ช่อง (`addr_building` … `addr_postcode`) ข้อความก้อนเดียวแบบ Champ จึงแยกลงช่องไม่ได้โดยไม่เดา

## Decision

1. **ใช้ที่อยู่แบบแยกช่องตาม key ของ rdform** (ต่างจาก Champ ที่เก็บเป็นข้อความเดียว: ลุงจืดเลือก 2026-09-25):
   - 13 ช่อง `addr_building`, `addr_room`, `addr_floor`, `addr_village`, `addr_no`, `addr_moo`, `addr_soi`, `addr_junction`, `addr_road`, `addr_subdistrict`, `addr_district`, `addr_province`, `addr_postcode` + `phone`
   - เก็บเฉพาะระดับบริษัท = ที่อยู่สำนักงานใหญ่
   - **ไม่มี** fax / email / website / ที่อยู่รายสาขา
2. **แพ็กเกจ `backend/internal/taxaddress/`** เป็นกติกาเดียวของที่อยู่ ใช้ทั้ง API บริษัทและงานภาษี:
   - `Keys` (`taxaddress.go:15`), `Normalize` (`:64`)
   - `Validate` (`:105`) คืนข้อผิดพลาดของช่องแรกที่ผิด ตรวจ:
     - เพดานตัวอักษรนับ rune: อาคาร 40, ห้อง/ชั้น/เลขที่/หมู่ 20, หมู่บ้าน/ซอย/แยก/ถนน 100, ตำบล/อำเภอ/จังหวัด 50
     - ห้ามขึ้นบรรทัดใหม่หรือแท็บ
     - รหัสไปรษณีย์ 5 หลักหรือว่าง
     - โทรศัพท์ไม่เกิน 50 ตัวอักษร
   - `Line()` (`line.go:66`) ต่อบรรทัดเดียวตามลำดับ RD-50TAWI-FORM:
     - ผู้ใช้กรอกเฉพาะชื่อ/เลข ระบบเติมคำนำหน้าตามถ้อยคำทางการ
     - ไม่เติมซ้ำถ้าพิมพ์คำนำหน้ามาแล้ว เช่น "ถนนสาทรใต้", "ซ.", "ตรอก…"
     - กรุงเทพฯ (`IsBangkok` `:38`) ใช้ แขวง/เขต และไม่มี "จังหวัด"
3. **ตาราง `companies`**:
   - เพิ่ม 14 คอลัมน์ `TEXT NOT NULL DEFAULT ''` ผ่าน `ALTER TABLE … ADD COLUMN IF NOT EXISTS` ใน `backend/internal/centraldb/centraldb.go:125–138` (ตามกฎ "โค้ดใหม่คือ migration")
   - สำเนา DDL ตามไปที่ `schema_full.sql`, `fresh_provision_all.sql` และ fixture `generalledger/uat_fixture_test.go`
4. **API บริษัท** (`backend/internal/organization/company/company_http.go`) ใช้ `Company.Address *taxaddress.Address` + `Phone *string` (`models/company.go:17–20`):
   - GET คืนเสมอ
   - POST ไม่ส่ง = ว่าง
   - PUT ไม่ส่ง (nil) = คงค่าเดิม ถ้าส่งมาจะแทนทั้งชุด (`mergeCompanyTaxAddress` `:472`)
   - ข้อผิดพลาดตอบเป็น `VALIDATION_FAILED` + `field` (`address.<key>` / `phone`) + ข้อความจาก key `company_tax_addr_err_*` ตามภาษาที่ผู้ใช้เลือก (`companyTaxAddressError` `:534`)
   - SQL เป็น parameterized ทั้งหมด
5. **เติมหัวแบบภาษีอัตโนมัติ**:
   - `CompanyHeader` มี `Address` / `Phone` / `AddressLine` อ่านจาก `queryCompanyHeader` (`backend/internal/goapi/handlers/tax_money.go:26`, `:53`)
   - `applyRegistryAddress` (`tax_form_fill.go:215`) ทำงานหลัง `copyProfile`:
     - ทะเบียนมีที่อยู่ → ล้าง `addr_*` ที่ยกจากฉบับก่อนทั้งชุด แล้วใส่ของทะเบียน เฉพาะ key ที่แบบนั้นมี
     - ทะเบียนมีโทรศัพท์และแบบมีช่องโทรศัพท์ → ใส่โทรศัพท์
     - ฉบับก่อนยื่นในนามสาขา (`branch_no` ไม่ใช่ 00000) → ไม่ทับ เพราะทะเบียนเก็บเฉพาะสำนักงานใหญ่
     - ทุกช่องยังแก้ได้บนจอ และฉบับที่บันทึกแล้วพิมพ์ตามที่บันทึก
6. **50 ทวิ ฝั่งบริษัทเรา** (`wht_certificate.go:156`, `:202`, `companyAddressParty` `:259`) ใช้ที่อยู่ตามลำดับ:
   1. snapshot ที่บันทึกในรายการภาษีหัก
   2. ค่าที่ส่งมาจากจอ
   3. `AddressLine` จากทะเบียน
   ชื่อ/เลขผู้เสียภาษีของบริษัทยังมาจากทะเบียนเหมือนเดิม
7. **จอ**:
   - ส่วน "ที่อยู่สำหรับแบบภาษี (สำนักงานใหญ่)" อยู่ในฟอร์มบริษัท (`frontend/src/app/system-settings/company-branch-tree-view.tsx` `CompanyTaxAddressSection`)
     - ตรวจแบบเดียวกับ backend (`taxAddressProblems` ใน `frontend/src/lib/thai-tax.ts`) และเตือนใต้ช่อง
     - ไม่ตัดข้อความด้วย `maxLength`
   - แผง 50 ทวิ เติมที่อยู่ผู้หักภาษีจาก `company.addressline` (แก้ได้) — เป็นค่าแสดง/พิมพ์เท่านั้น: ปุ่ม "บันทึกลงใบสำคัญ" เขียนเฉพาะช่องที่ผู้ใช้แก้และต่างจากค่าในใบสำคัญ พร้อมแสดงค่าเดิมในใบสำคัญในหน้ายืนยัน (`whtSnapshotChanges`)
   - ฝั่งภาษีถูกหักแสดง "ที่อยู่ของกิจการเรา (ตามทะเบียนบริษัท)" (`frontend/src/app/tax/wht-certificate-panel.tsx`)

## Alternatives

- **ข้อความเดียวแบบ Champ (`Address VARCHAR(255)`)** ไม่เลือก:
  - หัวแบบ rdform แยก 13 ช่อง ถ้าจะแตกข้อความเดียวลงช่องต้องเดาว่าคำไหนคือถนน/ตำบล ผิดกฎ NO MAGIC
  - คำนำหน้ากรุงเทพฯ (แขวง/เขต) ก็ตรวจจากข้อความก้อนเดียวไม่ได้
  - ในทางกลับกัน จากช่องแยกต่อเป็นบรรทัดเดียว (50 ทวิ) ได้เสมอ
- **ที่อยู่รายสาขาใน `branches.settings` / ตารางสาขา** ไม่ทำตอนนี้ (ลุงจืดตัดสินใจ): Champ ก็เก็บระดับบริษัทเท่านั้น และแบบที่ยื่นในนามสาขายังไม่มีในขอบเขตรอบนี้
- **เพิ่ม fax / email / website** ไม่ทำ: แบบภาษีที่ระบบรองรับไม่มีช่องเหล่านี้ (Champ มี Fax แต่ไม่มีแบบไหนใช้) ตามกฎ Zero-Bloat

## Consequences

- บริษัทที่มีอยู่แล้วได้ค่าว่างทุกช่อง (ไม่ใช่ NULL):
  - พฤติกรรมเดิมยังอยู่ (ยกจากฉบับก่อน) จนกว่าผู้ใช้จะกรอกที่อยู่ในทะเบียน
  - ไม่มี backfill ตามกฎ "ช่วง dev เดินหน้าอย่างเดียว"
- เมื่อกรอกทะเบียนแล้ว ที่อยู่ที่ยกจากฉบับก่อนของสำนักงานใหญ่จะถูกแทนด้วยของทะเบียนทั้งชุด ป้องกันหัวแบบที่ปนที่อยู่สองแห่ง ผู้ใช้ยังแก้ในแบบนั้นได้
- **งานในอนาคต (ยังไม่ทำ)**:
  - ที่อยู่รายสาขาสำหรับแบบที่ยื่นแยกสาขา (เช่น ภ.พ.30 แยกยื่นรายสถานประกอบการ): ตอนนี้ฉบับก่อนที่เป็นสาขาจะคงที่อยู่ที่ยกมา และฉบับแรกของสาขาต้องพิมพ์เอง
  - เก็บ snapshot ที่อยู่ฝั่งบริษัทตอนบันทึกรายการภาษีหัก (ตอนนี้ 50 ทวิ ที่ออกซ้ำภายหลังจะใช้ที่อยู่ทะเบียนปัจจุบัน ถ้าตอนบันทึกไม่ได้เก็บไว้)
  - ที่อยู่ภาษาอังกฤษ (Champ มี `AddressEng`)
- PDF และไฟล์ยื่นด้วยสื่อ (`rdfile`) ของฉบับที่บันทึกแล้วไม่เปลี่ยนพฤติกรรม:
  - ทับจากทะเบียนเฉพาะชื่อและเลขผู้เสียภาษี (`applyCompanyHeader` `tax_form.go:268`)
  - ที่อยู่/โทรศัพท์ใช้ตามที่บันทึกในฉบับ แก้ทะเบียนทีหลังจึงไม่เปลี่ยนแบบที่ยื่นไปแล้ว
- ทดสอบ:
  - `go test ./internal/taxaddress/ ./internal/organization/company/ ./internal/goapi/handlers/`
  - integration (`-tags integration`, PostgreSQL จริง): `TestCompanyTaxAddressAgainstCentralSchema`, `TestQueryCompanyHeaderTaxAddress`
  - `npx vitest run src/lib/thai-tax.test.ts src/app/tax/wht-certificate-panel.test.ts src/app/system-settings/company-branch-tree-view.test.ts src/app/system-settings/settings-language-keys.test.ts`

เกี่ยวข้อง: [[2026-09-23-tax-inside-general-ledger]], [[2026-09-24-rd-wht-file-export]], `docs/kms/21-thai-tax-form-references.md` §1 (RD-50TAWI-FORM, TH-BMA-ACT-2528) §6 §9, skill `ui-scale-polish` §8.45
