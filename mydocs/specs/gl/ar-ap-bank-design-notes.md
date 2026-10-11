# ตรวจสอบบัญชีรายวัน: ลูกหนี้ เจ้าหนี้ และ Statement แบบ many-to-many

เพิ่มร่างตามคำอนุมัติลุงจืด 2026-09-20 ให้ตรวจแก้ต่อ ไม่ใช่ migration ของฐานจริง

## ไฟล์และลำดับสร้าง

ไฟล์ SQL ที่กล่าวถึงอยู่ใน `../../datamodels/gl/` ส่วนเอกสาร Markdown อยู่ใน `mydocs/specs/gl/`

chartofaccount.sql → journalbook.sql → period.sql → gltrans.sql → partners.sql → ar.sql → ap.sql → bank.sql; vat.sql และ wht.sql รันหลัง gltrans.sql เช่นเดิม

- partners.sql: gl_partners คู่ค้าหนึ่งรายเป็นทั้งลูกหนี้และเจ้าหนี้ได้ เป็นทะเบียนภายในห้องบัญชี รับข้อมูลจากภายนอกหรือคีย์ตรงได้ ไม่พึ่งโมดูลคู่ค้าหน้าบ้าน
- ar.sql: ar_documents, ar_journal_allocations, ar_settlements
- ap.sql: ap_documents, ap_journal_allocations, ap_settlements
- bank.sql: gl_bank_accounts, bank_statement_lines, bank_journal_lines, bank_statement_matches

## ขั้นตอนรับข้อมูลและตรวจสอบ

รับข้อมูลจากภายนอกหรือคีย์ใบสำคัญ → ตรวจคู่ค้า/หลักฐานหนี้หรือการชำระที่เกิดขึ้นแล้ว → กระจายยอดบรรทัดลูกหนี้หรือเจ้าหนี้ไปหลายเอกสาร → เลือกบิลและยอดตัดชำระ → ระบุบัญชีธนาคารและจับคู่ Statement ที่มี → VAT/WHT ตามรายการ → บันทึกทุกส่วนใน PostgreSQL transaction เดียวผ่าน Go sql.Tx

Statement อาจเข้าภายหลัง จึงไม่บังคับต้องจับคู่ก่อนบันทึกใบสำคัญ และไม่สร้าง Statement ปลอมจาก GL

## ความหมายยอดและกรณีตรวจรับ

- เอกสาร AR/AP มี ID ภายในแยกจากเลขเอกสาร คู่ค้า/บริษัทต้องตรงกันทั้งสองฝั่งตัดชำระโดย composite FK
- ar/ap_journal_allocations เชื่อมเอกสารกับบรรทัด GL แบบ many-to-many; amount เป็นส่วนของบรรทัดบัญชีคุม ห้ามรวมทุกบรรทัดเดบิตเครดิตจนยอดซ้ำ
- เอกสารเพิ่มหนี้ 1,000 และ 500 รับชำระ 800 และ 700: จัดสรร 600+200 และ 400+300 รวมทั้งสองบิลหมดพอดี
- settlements เป็นการตัดยอดเอกสารเพิ่มหนี้กับเอกสารลดหนี้ ไม่สร้าง GL ซ้ำจากการจับคู่ มี ID ต่อการตัดยอดเพื่อป้องกัน retry ซ้ำ
- รับชำระรวมภาษีถูกหักใช้ยอดเต็มตัดลูกหนี้ ส่วนธนาคารใช้ยอดเข้าจริง; ค่าธรรมเนียมและภาษีเป็นบรรทัด GL แยก ห้ามเพิ่มยอดจับคู่ Statement เพื่อกลบส่วนต่าง
- ยอดคงเหลือคำนวณจากยอดเอกสารที่ผ่านรายการ ลบยอดตัดชำระที่ยังไม่ถอน ไม่เก็บยอดคงเหลือซ้ำ; ยอดชำระยังไม่จัดสรรเป็นยอดรอตัดบิล
- bank_statement_matches จับคู่หลายบรรทัด GL กับหลาย Statement แต่บริษัท บัญชีธนาคาร และทิศทางต้องตรงกัน
- เงินเป็น NUMERIC ตาม GL เดิม รองรับทศนิยมที่ผู้ใช้กำหนด ไม่ใช้ float/JSON number; ร่างนี้จำกัด THB ยังไม่รองรับ FX

## ข้อบังคับที่ต้องทำใน Backend ก่อนนำไปใช้งานจริง

DDL นี้บังคับ FK/ประเภท/ยอดบวก/จำนวน finite และ key เท่านั้น ยังไม่ได้บังคับยอดสะสมหรือสิทธิ์แก้ไข

1. ใช้ PostgreSQL เพียงตัวเดียว ไม่ใช้ MongoDB, ClickHouse, Kafka หรือ Redis ตามคำสั่งลุงจืด 2026-09-20
2. ภายใน sql.Tx ล็อกเอกสารและบรรทัดที่เกี่ยวข้องด้วย SELECT FOR UPDATE ตามลำดับคงที่ ตรวจ SUM ของรายการที่ยังไม่ถอนทั้งสองฝั่งก่อนเพิ่มยอด ห้ามตรวจยอดนอก transaction
3. ตรวจการผ่านรายการ งวดเปิด สิทธิ์บริษัท/สาขา คู่ค้า active/บทบาท และบัญชีคุมลูกหนี้เจ้าหนี้; ตรวจ bank_journal_lines.gl_account_code ตรง journal_lines.account_code และ direction ตรงเดบิต/เครดิตจริง
4. ผลรวม allocations ต้องตรง amount ของเอกสารก่อนผ่านรายการ และไม่เกินยอดบรรทัด GL; settlements ไม่เกินทั้งบิลและยอดรับจ่าย; matches ไม่เกินทั้ง Statement และยอดบรรทัดธนาคาร
5. API ส่ง decimal string; ตรวจ scale ก่อนลง DB ใช้นโยบาย rounding เดียวกับบริษัท ห้ามคำนวณใหม่จากข้อความจัดรูปแบบ
6. ID การบันทึกซ้ำต้องคืนผลเดิมเมื่อ payload เดิม และปฏิเสธเมื่อ ID เดิมแต่ payload ต่าง
7. หลังยืนยัน ห้ามลบหรือเปลี่ยนคู่/ยอดของ allocation โดยตรง ถอนด้วย reversed_at/reversed_by/reversal_reason เพียงครั้งเดียวพร้อม audit แล้วสร้าง ID ใหม่; การกลับบัญชีต้องทำ reversal GL ไม่ใช่แค่ถอนจับคู่
8. รายงานย้อนหลังใช้ created_at/reversed_at ประกอบวันที่บัญชี; ป้องกันแก้เอกสารที่มีการตัดชำระ/จับคู่แล้ว
9. ทดสอบ round-trip API/DB, 0.1+0.2, ขอบเขตปัดเศษ, debit=credit, retry, concurrent allocation, partial settlement, reversal, การนำเข้า Statement ซ้ำ และการกระทบยอดทั้งระบบ

## ขอบเขตที่ยังต้องยืนยัน

ร่างนี้ยังไม่ครอบคลุมเงินต่างประเทศ การหักกลบลูกหนี้กับเจ้าหนี้ ส่วน workflow สั่งจ่าย/คืนเงิน/เช็คเป็นหน้าที่ระบบต้นทาง ไม่เพิ่มเมนูหรือเงินเดือน ไม่รวม RD Prep/XBRL

ต้นแบบที่ตรวจ: D:/project-champ/champ-report/BC5REP_AR/RepStatementByArView.cpp (PayAmount/BillBalance) และ D:/project-champ/champ-report/BC5REP_CASH/CASHRepBankStatementView.cpp (BookNo/DocNo/TransBankDate/MyAmount) ส่วน many-to-many เป็นข้อกำหนดเพิ่มเติมที่ลุงจืดยืนยัน

## การตรวจเอกสาร

ใช้ขอบเขตใน [spec.md](spec.md) เป็นหลัก ตัดยอดในตาราง settlements เพื่อกระทบยอดบัญชี ไม่ออกคำสั่งชำระเงินจริง และไม่สร้าง workflow หน้าบ้าน

รัน review.sql หลัง gltrans.sql เพื่อเก็บประวัติรอตรวจ/พบข้อแตกต่าง/ตรวจแล้วแยกตาม revision ของใบสำคัญ การจับคู่หรือถอนจับคู่ที่กระทบผลตรวจต้องเปิดตรวจใหม่ใน transaction เดียว ไม่มีระบบอนุมัติหลายแผนก

