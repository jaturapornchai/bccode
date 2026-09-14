---
date: 2026-09-11
status: accepted
tags: [bc-account, frontend, menu, champ, upgrade]
---

# เมนูอัปเกรดจาก Champ: งานเดิมเป็นหลัก ผสานงานใหม่ในระบบเดียวกัน

## วัตถุประสงค์และขอบเขต

ลุงจืดต้องการให้ผู้ใช้ Champ ย้ายมาใช้ BC แล้วหาเมนูงานเดิมได้ และใช้ฟังก์ชันใหม่ร่วมกันได้ใน 9 ระบบเดิม งานรอบนี้เป็นการจัดผังเมนูและการนำทาง ไม่ได้เพิ่มการบันทึกบัญชี การอนุมัติจริง หรือสูตรคำนวณทางธุรกิจ

ต้นทางที่ตรวจจริง: `D:/project-champ/champ/champ/menuconfigxml/menuconfig.xml` (UTF-16LE) และ `D:/project-champ/champ/champ/BC5Account.rc` (Windows-874; ส่วน PO/Bill/AP/AR/CHQSYS/Bank-Cash/IC/AS/GL เริ่มช่วง 20065–20266)

- XML มีเมนูปลายทางที่ไม่ใช่ตัวคั่น **490 occurrences / 482 resource IDs**: PO 96, Bill 15, AP 37, AR 53, CHQSYS 38, Bank/Cash 28, IC 68, AS 28, GL 48, รายงาน 79
- จำนวนนี้รวมรายงานรายวัน รายงานวิเคราะห์ รายงานซ้ำ และรายการใต้กลุ่มที่ source ตั้ง `disable` ไว้ จึงใช้เป็นจำนวนหน้าจอที่พัฒนาเสร็จไม่ได้
- resource `59011` ปรากฏทั้ง AP และ AR: ต้องพิจารณาโมดูลประกอบ ห้ามรวมการคำนวณบิลเจ้าหนี้กับลูกหนี้เพียงเพราะรหัสเหมือนกัน
- RC กับ XML มีความต่าง เช่น RC:20118 มีคูปอง `30026` เพิ่มจาก XML; การเทียบนี้ไม่ได้ยืนยันความสามารถของรายงานทุกแบบหรือคำสั่งทุกตัวใน RC ห้ามสรุปว่าเทียบเท่า Champ ทั้งหมด

## ผลการจัดผัง

**154 → 223 เมนู** (+69) ใน 9 ระบบ เพิ่มกลุ่มย่อย 20 กลุ่ม เก็บ id/route เดิมทั้ง 154 รายการครบโดย fixture จาก working tree ก่อนแก้ ไม่เอา baseline จาก HEAD ที่มีผังเก่ากว่า

| ระบบ | จำนวน | กลุ่มงาน |
|---|---:|---|
| po | 29 | งานจัดซื้อจัดหา · อนุมัติและยกเลิกการซื้อ · บันทึกซื้อและค่าใช้จ่าย · ต้นทุนแฝงและปรับปรุงใบรับสินค้า · เงินมัดจำและจ่ายล่วงหน้า · รายงานจัดซื้อ |
| bill | 30 | งานขายและออกบิล · อนุมัติและยกเลิกการขาย · สั่งจองและกำหนดส่งสินค้า · ใบลดหนี้และเพิ่มหนี้ · เงินมัดจำและรับล่วงหน้า · รายงานขายและวิเคราะห์การขาย |
| ap | 13 | ข้อมูลหลักเจ้าหนี้ · ตั้งหนี้อื่น รับวางบิล และตัดหนี้ · การเงินและการจ่ายชำระหนี้ · รายงานเจ้าหนี้ · ตรวจและประมวลผลเจ้าหนี้ |
| ar | 13 | ข้อมูลหลักลูกหนี้ · ตั้งหนี้อื่นและตัดหนี้ลูกหนี้ · งานวางบิลและแจ้งหนี้ · การเงินและการรับชำระหนี้ · รายงานลูกหนี้ · ตรวจและประมวลผลลูกหนี้ |
| cash-bank | 34 | บัญชีเงินฝากและสมุดบัญชี · การจัดการเงินสดและทดรอง · ธุรกรรมธนาคาร · เช็ครับ · เช็คจ่าย · รับและขึ้นเงินบัตรเครดิต · ตรวจและประมวลผลเงินฝากและเช็ค |
| ic | 45 | ข้อมูลหลักสินค้า · ขอเบิก ขอโอน และตรวจรับสินค้า · งานประจำสินค้าและคลัง · ตรวจนับและปรับปรุงสินค้า · สินค้าชุดและส่วนประกอบ · ทะเบียนเลขเครื่องสินค้า · ราคาขายและโปรโมชั่น · รายงานสินค้า · ตรวจและประมวลผลสินค้า |
| fa | 9 | งานสินทรัพย์ถาวร · ผ่านค่าเสื่อมเข้าบัญชี · รายงานสินทรัพย์ |
| vat | 15 | บันทึกภาษีและทะเบียนภาษี · รายงานภาษีและแบบยื่น |
| gl | 35 | ข้อมูลหลักและยอดยกมา · สมุดรายวันและงานประจำ · ผ่านบัญชีและประมวลผลสิ้นปี · รายงานการเงินและงบบัญชี |

- งานเดิม: แยกการอนุมัติ/ยกเลิก เช็ครับ เช็คจ่าย รับบัตรเครดิต สินค้าชุด ตรวจนับ และผ่านบัญชีตามงานที่พบใน Champ
- งานใหม่ยังอยู่ครบ: สแกนบิล/คลังเอกสาร/เอกสารระหว่างกิจการใน PO; e-Tax/เอกสารขายประจำใน BILL; กระทบยอด/สลิป/ไฟล์โอนเงินใน Cash; ล็อต/FIFO ใน IC; ภาพรวมและวิเคราะห์ธุรกิจใน GL
- เพิ่มชื่อค้นหาเดิม เช่น ใบเสนอซื้อ → ใบขอซื้อ, ใบสั่งจอง → ใบสั่งขาย, Weight cost → ต้นทุนแฝง, Serial Number → ทะเบียนเลขเครื่อง โดยไม่สร้างเมนูซ้ำและไม่เปลี่ยนชื่อไทยที่อนุมัติแล้ว (frontend/src/lib/menu-data.ts:37)
- ผังที่อ่านได้ทั้ง 223 รายการอยู่ที่ [รายการเมนูแยกตามระบบ](../snippets/champ-upgrade-menu-map.md)

## สถานะหน้าจอและข้อจำกัด

223 เมนูแบ่งเป็น **211 รอพัฒนา / 12 มีตัวเปิดหน้าจอเชื่อมแล้ว** ตัวเลขหลังไม่ใช่การรับรอง CRUD หรือการคำนวณของทุกจอ

- เมนูใหม่ 68 จาก 69 รายการเป็นรอพัฒนา รวมทะเบียนเลขเครื่องที่มี config แต่ทดสอบจริงได้ 404 จาก `/api/system-settings/productserialregistry`
- ใช้ route ทะเบียนเลขเครื่องเดิม `/productserialregistry` และเพิ่มข้อยกเว้น pending ใน frontend/src/lib/menu-screen-status.ts:14 แทนการเปิดจอที่ดึงข้อมูลไม่ได้; WorkTabPanel ตรวจสถานะเดียวกันก่อนเปิด config (frontend/src/app/menu/main-menu-screen.tsx:2828)
- โปรโมชั่นใช้จอเดิม `/promotionscreen`; ตรวจเปิดหน้าได้ ไม่ได้ทำ CRUD ในงานเมนูนี้
- ไฟล์ `product-set-screen.tsx` มีอยู่แต่ไม่ได้อยู่ใน dispatcher ของเมนูที่ตรวจ จึงไม่ถือว่า flow รวม/แยกชุดเชื่อมเสร็จหรือเท่ากับ Champ; เมนูสินค้าชุดที่เพิ่มเป็น pending
- หน้า pending บอกผู้ใช้เป็นภาษาไทยว่าเปิดดูแผนงานได้แต่ยังบันทึก/ประมวลผลไม่ได้ ไม่ใช้คำอธิบาย Flutter/Next.js หรือแสดง route ทางเทคนิค (frontend/src/app/menu/main-menu-screen.tsx:2841)
- ไม่เพิ่ม payroll / ภ.ง.ด.1 / ภ.ง.ด.1ก / ประกันสังคม และไม่คืน employee/user/permissiongroup/useraccessaudit เข้าเมนูสาขา; ภ.ง.ด.2/3/53/50 ทวิและเงินทดรองยังอยู่
- คงรายงานเดิมไว้ รอบนี้เพิ่มรายการคำสั่งงานตามตารางด้านล่าง ไม่ได้เพิ่มรายงานย่อยทุก variant จาก 490 occurrences หรือยืนยันว่ารายงานรวมของ BC ทดแทนทุก variant ได้
- คีย์ภาษาใหม่ครบ 13 คอลัมน์ตามโครง TSV; ไทย/อังกฤษกำหนดไว้ ภาษาอื่นใช้ English fallback อย่างเปิดเผยตามแบบแผนที่มีอยู่ ยังไม่ใช่คำแปลที่เจ้าของภาษารับรอง

[บัญชีรายการต้นทางครบ 490 occurrences](../snippets/champ-menu-source-audit.md) แสดงรายการที่จับคู่ชัดเจนและรายการที่ยังต้องตรวจเทียบต่อ โดยไม่ใช้การจับคู่ชื่อคล้ายเพื่ออ้างความครบ

## รายการเพิ่มที่อ้างกลับต้นทางได้

เลขบรรทัด menuconfig.xml อ้างไฟล์ใน D:/project-champ ด้านบน; menu-data.ts อ้าง frontend/src/lib/menu-data.ts

| Resource | เมนู BC | Route | หลักฐาน |
|---|---|---|---|
| 20003 | อนุมัติใบเสนอซื้อสินค้า | `/procurement/requisition-approval` | menuconfig.xml:16 · menu-data.ts:95 |
| 20025 | ยกเลิกใบสั่งซื้อสินค้า | `/procurement/order-cancellation` | menuconfig.xml:4 · menu-data.ts:96 |
| 20004 | ตารางเปรียบเทียบราคาซื้อ | `/procurement/price-comparison` | menuconfig.xml:19 · menu-data.ts:87 |
| 20016 | ประมวลผลใบสั่งซื้ออัตโนมัติ | `/procurement/generate-orders` | menuconfig.xml:20 · menu-data.ts:88 |
| 40008 | บันทึกต้นทุนแฝง | `/transaction/landedcost` | menuconfig.xml:25 · menu-data.ts:119 |
| 20020 | ปรับปรุงใบรับสินค้า | `/transaction/purchasereceiptadjustment` | menuconfig.xml:26 · menu-data.ts:120 |
| 30003 | อนุมัติใบเสนอราคาสินค้า | `/sales/quotation-approval` | menuconfig.xml:150 · menu-data.ts:167 |
| 30022 | ยกเลิกใบเสนอราคาสินค้า | `/sales/quotation-cancellation` | menuconfig.xml:151 · menu-data.ts:168 |
| 30012 | อนุมัติใบสั่งขาย/สั่งจองสินค้า | `/sales/order-approval` | menuconfig.xml:154 · menu-data.ts:169 |
| 30020 | ยกเลิกใบสั่งขาย/สั่งจองสินค้า | `/sales/order-cancellation` | menuconfig.xml:155 · menu-data.ts:170 |
| 30024 | ติดตามใบสั่งจองสินค้า | `/sales/reservations` | menuconfig.xml:156 · menu-data.ts:177 |
| 30025 | ตรวจสอบวันที่ใบสั่งขาย/สั่งจอง | `/sales/order-dates` | menuconfig.xml:157 · menu-data.ts:178 |
| 30011 | ปรับปรุงวันที่ส่งของให้ลูกค้า | `/sales/delivery-dates` | menuconfig.xml:167 · menu-data.ts:179 |
| 40004 | ตั้งเจ้าหนี้อื่นๆ | `/transaction/apotherdebt` | menuconfig.xml:173 · menu-data.ts:235 |
| 40005 | ใบรับวางบิลเจ้าหนี้ | `/transaction/apbillingreceipt` | menuconfig.xml:174 · menu-data.ts:236 |
| 40009 | ตัดหนี้สูญเจ้าหนี้ | `/transaction/apbaddebt` | menuconfig.xml:177 · menu-data.ts:237 |
| 49001 | คำนวณยอดเจ้าหนี้ใหม่ | `/tools/ap-recalculate` | menuconfig.xml:179 · menu-data.ts:260 |
| 59011 | คำนวณยอดคงเหลือบิลเจ้าหนี้ใหม่ | `/tools/ap-bill-balances` | menuconfig.xml:180 · menu-data.ts:261 |
| 50004 | ตั้งลูกหนี้อื่นๆ | `/transaction/arotherdebt` | menuconfig.xml:231 · menu-data.ts:283 |
| 50008 | ตัดหนี้สูญลูกหนี้ | `/transaction/arbaddebt` | menuconfig.xml:236 · menu-data.ts:284 |
| 50007 | ใบเสร็จชั่วคราว | `/transaction/temporaryreceipt` | menuconfig.xml:233 · menu-data.ts:300 |
| 59001 | คำนวณยอดลูกหนี้ใหม่ | `/tools/ar-recalculate` | menuconfig.xml:238 · menu-data.ts:315 |
| 59011 | คำนวณยอดคงเหลือบิลลูกหนี้ใหม่ | `/tools/ar-bill-balances` | menuconfig.xml:239 · menu-data.ts:316 |
| 60030 | นำฝากเช็ครับ | `/banking/cheques/deposit` | menuconfig.xml:314 · menu-data.ts:367 |
| 60040 | เช็ครับผ่าน | `/banking/cheques/received-clear` | menuconfig.xml:315 · menu-data.ts:368 |
| 60110 | เช็ครับคืน | `/banking/cheques/received-return` | menuconfig.xml:316 · menu-data.ts:369 |
| 60120 | นำเช็คเข้าใหม่ | `/banking/cheques/redeposit` | menuconfig.xml:317 · menu-data.ts:370 |
| 60050 | ยกเลิกเช็ครับ | `/banking/cheques/received-cancel` | menuconfig.xml:318 · menu-data.ts:371 |
| 60060 | ขายลดเช็ครับ | `/banking/cheques/discount` | menuconfig.xml:319 · menu-data.ts:372 |
| 60070 | เช็คจ่ายผ่าน | `/banking/cheques/issued-clear` | menuconfig.xml:321 · menu-data.ts:380 |
| 60080 | ยกเลิกเช็คจ่าย | `/banking/cheques/issued-cancel` | menuconfig.xml:322 · menu-data.ts:381 |
| 60023 | ทะเบียนรับชำระด้วยบัตรเครดิต | `/banking/cards/receipts` | menuconfig.xml:312 · menu-data.ts:388 |
| 60090 | ขึ้นเงินบัตรเครดิต | `/banking/cards/settlement` | menuconfig.xml:324 · menu-data.ts:389 |
| 60100 | ยกเลิกรายการรับบัตรเครดิต | `/banking/cards/cancellation` | menuconfig.xml:325 · menu-data.ts:390 |
| 70005 | นำฝากเงินสด | `/banking/cash-deposit` | menuconfig.xml:373 · menu-data.ts:356 |
| 70006 | ถอนเงินสด | `/banking/cash-withdrawal` | menuconfig.xml:374 · menu-data.ts:357 |
| 70007 | บันทึกค่าใช้จ่ายธนาคาร | `/banking/charges` | menuconfig.xml:375 · menu-data.ts:358 |
| 70008 | บันทึกรายได้จากธนาคาร | `/banking/income` | menuconfig.xml:376 · menu-data.ts:359 |
| 740008 | ผู้ติดต่อธนาคาร | `/banking/contacts` | menuconfig.xml:371 · menu-data.ts:330 |
| 70001 | กำหนดวงเงินสดย่อย | `/banking/petty-cash-limit` | menuconfig.xml:370 · menu-data.ts:343 |
| 70014 | คำนวณยอดคงเหลือเช็คใหม่ | `/tools/cheque-balances` | menuconfig.xml:327 · menu-data.ts:397 |
| 70011 | คำนวณยอดสมุดบัญชีใหม่ | `/tools/bank-balances` | menuconfig.xml:379 · menu-data.ts:398 |
| 80012 | ใบขอเบิกสินค้าและวัตถุดิบ | `/inventory/issue-request` | menuconfig.xml:414 · menu-data.ts:426 |
| 80013 | ใบขอโอนสินค้า | `/inventory/transfer-request` | menuconfig.xml:419 · menu-data.ts:427 |
| 80014 | ใบตรวจรับสินค้า | `/inventory/goods-inspection` | menuconfig.xml:434 · menu-data.ts:428 |
| 80009 | เอกสารเพื่อตรวจนับสินค้า | `/inventory/count-sheet` | menuconfig.xml:422 · menu-data.ts:448 |
| 80002 | สินค้าชุด | `/inventory/product-sets` | menuconfig.xml:427 · menu-data.ts:458 |
| 80015 | ตรวจสอบรวมสินค้าชุด | `/inventory/set-assembly` | menuconfig.xml:428 · menu-data.ts:459 |
| 80016 | ตรวจสอบแยกสินค้าชุด | `/inventory/set-disassembly` | menuconfig.xml:429 · menu-data.ts:460 |
| 80099 | รายการย่อยสินค้าชุดแบบที่ 2 | `/inventory/set-components` | menuconfig.xml:431 · menu-data.ts:461 |
| 80011 | ทะเบียนเลขเครื่อง | `/productserialregistry` | menuconfig.xml:433 · menu-data.ts:468 |
| 30010 | กำหนดราคาขายสินค้า | `/inventory/selling-prices` | menuconfig.xml:409 · menu-data.ts:475 |
| 30019 | โปรโมชั่น | `/promotionscreen` | menuconfig.xml:410 · menu-data.ts:476 |
| 80018 | ปรับปรุงราคาขายสินค้า | `/inventory/price-adjustment` | menuconfig.xml:439 · menu-data.ts:477 |
| 81099 | กำหนดลำดับรายวันสินค้า | `/inventory/daily-sequence` | menuconfig.xml:435 · menu-data.ts:499 |
| 90002 | บันทึกซ่อมบำรุงสินทรัพย์ | `/asset/maintenance` | menuconfig.xml:514 · menu-data.ts:521 |
| 110021 | ประเภทสินทรัพย์ | `/asset/types` | menuconfig.xml:516 · menu-data.ts:522 |
| 90007 | โอนค่าเสื่อมราคาเข้าบัญชีแยกประเภท | `/asset/post-gl` | menuconfig.xml:519 · menu-data.ts:529 |
| 100017 | ปรับปรุงภาษีซื้อ | `/transaction/purchasevatadjustment` | menuconfig.xml:571 · menu-data.ts:551 |
| 100003 | กำหนดงบประมาณประจำปี | `/gl/budget` | menuconfig.xml:564 · menu-data.ts:584 |
| 100010 | กลุ่มผังบัญชี | `/gl/account-groups` | menuconfig.xml:565 · menu-data.ts:585 |
| 100006 | รูปแบบการเชื่อมโยงบัญชีอัตโนมัติ | `/gl/account-mapping` | menuconfig.xml:561 · menu-data.ts:586 |
| 110019 | กลุ่มบัญชีสินค้า | `/gl/product-account-groups` | menuconfig.xml:566 · menu-data.ts:587 |
| 100021 | ยอดสะสมประจำปี | `/gl/annual-balances` | menuconfig.xml:568 · menu-data.ts:588 |
| 100009 | ผ่านรายการบัญชี | `/gl/posting` | menuconfig.xml:557 · menu-data.ts:610 |
| 100013 | ยกเลิกการผ่านรายการบัญชี | `/gl/unposting` | menuconfig.xml:558 · menu-data.ts:611 |
| 100016 | ประมวลผลข้อมูลบัญชีใหม่ | `/gl/reprocess` | menuconfig.xml:556 · menu-data.ts:612 |
| 100014 | คำนวณยอดผ่านรายการใหม่ | `/gl/recalculate-posted` | menuconfig.xml:559 · menu-data.ts:613 |
| 100015 | ประมวลผลสิ้นปี | `/gl/year-end` | menuconfig.xml:570 · menu-data.ts:614 |

## การใช้งานและ dependency

ไม่มี config, plugin หรือแพ็กเกจใหม่ ใช้ MENU_SECTIONS → เมนูซ้าย/เมนูบน/ทางลัดและสิทธิ์เดิม; ใช้ isMenuScreenPending → ป้ายและตัวเปิดหน้าให้ตรงกัน ไอคอนใหม่ทุก route อยู่ใน frontend/src/lib/menu-icons.ts และกลุ่มอยู่ใน MENU_GROUP_ICONS

ตัวอย่าง: เปิด Cash → เช็ครับ → นำฝากเช็ครับ จะเห็นหน้า “เมนูในแผนพัฒนา”; ค้น “Weight cost” จะเจอต้นทุนแฝงใน PO; ค้น “ใบสั่งจอง” จะพบใบสั่งขายและขั้นตอนที่เกี่ยวข้อง

## ตรวจรับ

- Vitest เฉพาะเมนู 5 ไฟล์: **46 passed**; ตรวจ baseline 154 id/routes, ไม่ซ้ำ, ขอบเขต payroll/Holding, ชื่อค้นหา Champ, ภาษาและไอคอน
- TypeScript: `npx tsc --noEmit` ผ่าน
- Playwright: menu-tree-master + menu-consistency + menu-champ-upgrade ผ่าน 3 tests; เปิดเมนูเก่า/ใหม่และจอโปรโมชั่น, ตรวจ pending, ค้นชื่อเก่า, console ไม่มี error ในขอบเขตเมนู
- ภาพจริง light/dark โดยกดปุ่มสลับธีม × 1600/1280/1024/768 portrait; ตรวจ hover/focus และ horizontal overflow; เมนู pending ไม่มีปุ่มบันทึก/ลบให้ทดสอบ disabled จึงไม่มีการเพิ่มธุรกรรมทดสอบ
- ไม่ทดสอบ CRUD/MongoDB เพราะไม่มีการเพิ่มหรือเปลี่ยนการเขียนข้อมูลในงานนี้; ต้องทำ UAT ตามกฎโปรเจ็กต์เมื่อต่อ business workflow จริง
- คำสั่งรันซ้ำ: `npx vitest run src/lib/menu-data.test.ts src/lib/menu-icons.test.ts src/lib/menu-screen-status.test.ts src/lib/menu-usage.test.ts src/app/menu/main-menu-tree-consistency.test.ts`; `npx playwright test e2e/menu-champ-upgrade.spec.ts e2e/menu-consistency.spec.ts e2e/menu-tree-master.spec.ts --workers=1` (จาก frontend โดยมี local server)

## การย้อนกลับ

ย้อนเฉพาะ diff ของงานเมนูรอบนี้พร้อมภาษาและเอกสาร ห้าม git reset/revert ไฟล์รวมทั้งก้อน เพราะมีการเปลี่ยนของลุงจืดค้างอยู่ก่อนเริ่มงานแล้ว ไม่มี schema/API/ข้อมูลธุรกิจที่ต้อง migrate กลับ

## ปรับชื่อหมวดตามคำสั่งเพิ่มเติม 2026-09-11

เอาข้อความอังกฤษในวงเล็บออกจาก `title.th` ของทั้ง 9 ระบบ (เช่น “ระบบบัญชีเจ้าหนี้”) โดย `title.en` ยังเป็นภาษาอังกฤษตามเดิม ใช้ catalog ร่วมกันจึงเปลี่ยนทั้งเมนูซ้าย เมนูบน และชื่อหมวดในผลค้นหา อ้าง `frontend/src/lib/menu-data.ts:76` และ exact-title assertions ใน `frontend/src/lib/menu-data.test.ts:239`; จำนวน 223 เมนูและ id/route ทุกตัวคงเดิม

## เมนูบนเป็นค่าเริ่มต้น (คำสั่งเพิ่มเติม 2026-09-11)

ตั้ง `menuLayout` เริ่มต้นและ fallback เป็น `top`; ยังคืนค่า `left` ที่ผู้ใช้เคยเลือกไว้จาก `bc_menu_layout_mode` รออ่าน preference ก่อนบันทึก เพื่อไม่ให้ค่าเริ่มต้นทับตัวเลือกเดิม และเปิด sidebar ให้ถูกเมื่อคืนค่า left อ้าง `frontend/src/app/menu/main-menu-screen.tsx:414` และ `frontend/src/app/menu/main-menu-screen.tsx:524`
