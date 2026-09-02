# กฎ Holding, Company และ Branch

## เป้าหมาย

กำหนดพื้นที่ข้อมูลของระบบให้แยกจากกันอย่างปลอดภัยตามลำดับ:

```
Holding
└── Company
    └── Branch
```

Holding ใช้สำหรับบริหารและดูข้อมูลรวมของบริษัทที่ผู้ใช้มีสิทธิ์ ไม่ใช่เจ้าของข้อมูลธุรกรรมของทุกบริษัทโดยอัตโนมัติ

## ความหมาย

- Holding: กลุ่มกิจการและ Tenant ระดับ Application สำหรับ Membership และสิทธิ์
- Company: บริษัทที่เป็นเจ้าของและเป็นขอบเขตแยกสินค้า ลูกค้า เอกสาร และข้อมูลธุรกิจ
- Branch: สาขาที่อยู่ภายใต้ Company
- Membership: ความสัมพันธ์ระหว่าง User กับ Holding พร้อม Role สถานะ และวันหมดอายุ โดย Company และ Branch เป็น Scope ภายใน Membership
- Scope: Allow-list ที่กำหนดขอบเขตสูงสุดของข้อมูลธุรกิจที่ Membership อาจเข้าถึงได้ โดยต้องใช้รหัสถาวรและอยู่ใน Holding เดียวกับ Membership Scope ไม่ให้สิทธิ์โดยลำพัง ทุกคำขอข้อมูลธุรกิจยังต้องผ่าน Organization effective status และ Permission ของ Domain
- Holding Owner: User ที่มี Membership Role OWNER ผู้สร้าง Holding ได้ Role นี้โดยอัตโนมัติ และ Holding หนึ่งมี OWNER ได้มากกว่าหนึ่งคน
- Holding Admin: User ที่มี Membership ภายใน Holding และ Role เป็น ADMIN
- Organization stored status: สถานะ active หรือ inactive ที่บันทึกไว้กับ Holding, Company หรือ Branch นั้นโดยตรง
- Organization effective status: สถานะที่ใช้อนุญาตการทำงานจริงหลังรวม Organization stored status ของพื้นที่นั้นกับพื้นที่แม่
- OWNER ที่ใช้งานได้: User ที่ active และมี Membership Role OWNER ซึ่ง active และยังไม่หมดอายุ โดยไม่ขึ้นกับว่า Holding กำลัง active หรือ inactive เพื่อให้ยังมีผู้มีสิทธิ์บริหารและ activate กลับ

## รหัสประจำตัว

- Holding, Company และ Branch ต้องมีรหัสถาวรที่ Backend สร้าง ห้ามเปลี่ยน และใช้สำหรับ Relation กับ Database routing ส่วนโครงสร้างจัดเก็บให้ยึด [ลำดับความน่าเชื่อถือของเอกสาร](README.md#ลำดับความน่าเชื่อถือของเอกสาร)
- แต่ละระดับมีรหัสธุรกิจ โดย Holding ห้ามซ้ำทั้งระบบ Company ห้ามซ้ำภายใน Holding และ Branch ห้ามซ้ำภายใน Company
- รหัสธุรกิจต้องตัดช่องว่างหัวท้ายและเปรียบเทียบแบบไม่แยกตัวพิมพ์เล็กใหญ่ การเปลี่ยนรหัสต้องเป็นไปตาม [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว) ตรวจความซ้ำแบบ Atomic บันทึก Audit และห้ามนำรหัสที่เคยใช้กลับมาใช้ซ้ำภายในขอบเขตเดิม
- Company ต้องอยู่ใต้ Holding เดิมและ Branch ต้องอยู่ใต้ Company เดิมตลอดอายุข้อมูล ปัจจุบันไม่มี Workflow ย้าย Parent หรือย้ายข้อมูลข้าม Tenant จึงห้ามเปลี่ยน Parent ด้วยการแก้รหัส Relation โดยตรง

## สิทธิ์ที่ยืนยันแล้ว

- เงื่อนไขด้านตัวตนสำหรับการสร้าง Holding, Company และ Branch คือ User ต้อง active และมี Google Identity ที่ใช้งานได้ตาม [วิธีเข้าสู่ระบบ](login.md#วิธีเข้าสู่ระบบ)
- User ที่ผ่านเงื่อนไขด้านตัวตนสร้าง Holding ได้ และผู้สร้างได้รับ Role OWNER ภายในงาน Atomic เดียวกับการสร้าง Holding
- OWNER และ ADMIN สร้าง Company และ Branch ได้เฉพาะภายใน Holding ที่ Organization effective status active ส่วนการบริหารพื้นที่ inactive และการเปลี่ยนสถานะให้ยึด [สถานะการใช้งานขององค์กร](#สถานะการใช้งานขององค์กร)
- สิทธิ์ตามบทบาท (`role_permission.permissions`) เก็บเป็นรายการข้อความต่อจอ: `รหัสจอ` = เข้าจอได้, `รหัสจอ:create` / `รหัสจอ:update` / `รหัสจอ:delete` = เพิ่ม/แก้ไข/ลบ ในจอนั้น และ `*` = ทุกจอทุกการกระทำ (ค่าเริ่มต้นของ ADMIN/OWNER ที่ยังไม่กำหนด) รายการเดิมที่มีแต่ `รหัสจอ` ยังใช้ได้และหมายถึงเข้าจอได้อย่างเดียว Backend ต้องปฏิเสธรูปแบบอื่น ส่วน Frontend ซ่อนปุ่ม เพิ่ม/แก้ไข/ลบ ตามรายการนี้ (ตั้งโดยลุงจืด 2026-09-02) การบังคับที่ Backend API ต่อการกระทำเป็นงานระยะถัดไป
- OWNER และ ADMIN ไม่ได้รับสิทธิ์อ่านหรือแก้ข้อมูลธุรกรรมโดยอัตโนมัติ การเข้าถึงธุรกรรมต้องผ่าน [Membership](#membership) และ Permission ตามเอกสาร Domain ที่เกี่ยวข้อง หาก Domain นั้นยังไม่มี Permission Source of Truth ให้ปฏิเสธการเข้าถึงและหยุดถาม ห้ามอนุมานสิทธิ์จาก Role หรือ Scope อย่างเดียว
- Organization Audit ที่ OWNER หรือ ADMIN ดูได้โดยไม่ต้องมี Business Scope จำกัดเฉพาะการเปลี่ยน Holding, Company, Branch, Membership, Invitation, Organization status และ timezone ส่วน Audit ของข้อมูลธุรกรรมหรือ Domain อื่นเป็น Business Data และต้องผ่าน Organization effective status, Scope กับ Domain Permission ของข้อมูลนั้น
- Membership ทุก Role ยกเว้น OWNER ผู้สร้าง Holding ต้องเริ่มจาก Invitation
- OWNER และ ADMIN เชิญ USER รวมถึงเปลี่ยนสถานะ วันหมดอายุ และ Scope Company/Branch ของ Membership ที่ Role ปัจจุบันเป็น USER ได้ โดย ADMIN ห้ามเปลี่ยน Role
- OWNER เท่านั้นที่เปลี่ยน Role ของ Membership ใด ๆ ได้ รวมถึงเลื่อน USER เป็น ADMIN หรือ OWNER และลด OWNER หรือ ADMIN เป็น Role อื่น OWNER ยังมีสิทธิ์เชิญและเปลี่ยนสถานะ วันหมดอายุ หรือ Scope ของ OWNER และ ADMIN การเปลี่ยน Role ของ Membership ที่มีอยู่แล้วไม่ต้องออก Invitation ใหม่
- Membership มีสถานะ active หรือ inactive การถอน Membership หมายถึงเปลี่ยนเป็น inactive โดยคงข้อมูลและ Audit ผู้มีสิทธิ์สามารถ activate Membership รายการเดิมกลับได้ ห้ามสร้าง Membership รายการที่สองใน Holding เดิมหรือ Hard delete
- Invitation ต้องระบุ Holding และ Role โดย Scope เริ่มต้นอาจว่างได้ตาม [Membership](#membership) การตอบรับต้องมี Session ของ User ที่ active พร้อม Google Identity ที่ active และ verified email ตรงกับอีเมลผู้รับหลังตัดช่องว่างและเปรียบเทียบแบบไม่แยกตัวพิมพ์เล็กใหญ่ ห้ามตีความ dot หรือ plus alias เอง ไม่ต้องยืนยัน Google ใหม่ถ้า Identity ดังกล่าวยัง active แต่ Invitation token อย่างเดียวไม่พอ แล้วจึงสร้าง Membership ที่ active แบบ Atomic
- Invitation status เป็น `PENDING`, `ACCEPTED`, `REVOKED` หรือ `EXPIRED` ใช้ได้ครั้งเดียว และรับได้เฉพาะขณะ `PENDING` กับเวลาปัจจุบัน UTC ยังไม่ถึงเวลาหมดอายุ โดยต้องกำหนดเวลาหมดอายุเป็น 72 ชั่วโมงหลังสร้าง หากถึงเวลาแล้วให้ถือว่าใช้ไม่ได้ทันทีแม้ status ยังเป็น `PENDING` และ Backend ต้องเปลี่ยนจาก `PENDING` เป็น `EXPIRED` เพียงครั้งเดียวด้วย Atomic compare-and-set เมื่อ Request หรือ Background job พบรายการนั้น
- OWNER เพิกถอน Invitation ทุก Role ภายใน Holding ได้ ส่วน ADMIN เพิกถอนได้เฉพาะ Invitation Role USER โดยไม่ขึ้นกับว่าใครเป็นผู้สร้าง เมื่อออก Invitation ใหม่ให้ผู้รับและ Holding เดิม ภายใน MongoDB transaction เดียวต้องเปลี่ยนรายการ `PENDING` ที่หมดเวลาเป็น `EXPIRED` ก่อน แล้วเพิกถอนเฉพาะ `PENDING` ที่ยังไม่หมดเวลาและสร้างรายการใหม่ หากผู้กระทำไม่มีสิทธิ์เพิกถอนรายการเดิมที่ยังไม่หมดเวลาให้ปฏิเสธทั้งหมด และต้องรับประกันแบบ Atomic ว่ามี `PENDING` ที่ยังไม่หมดเวลาไม่เกินหนึ่งรายการต่อ Holding กับอีเมลที่ normalize แล้ว
- หากผู้รับมี Membership ใน Holding อยู่แล้ว ห้ามสร้าง Membership ซ้ำ ให้ใช้ Workflow เปลี่ยนสิทธิ์ Membership และห้ามออก Invitation ให้อีเมลดังกล่าวใน Holding นั้น ไม่ว่า Membership เดิมจะ active หรือ inactive โดยให้ใช้การเปลี่ยนสิทธิ์หรือ activate Membership เดิมแทน
- Membership ใช้งานได้เมื่อสถานะ active และไม่มีวันหมดอายุหรือเวลาปัจจุบัน UTC ยังไม่ถึงวันหมดอายุ Membership ที่ inactive หรือหมดอายุห้ามปรากฏใน Workspace ที่เลือกใช้งาน แต่อาจแสดงแบบ disabled ในหน้าประวัติสิทธิ์
- ห้ามถอน ลดสิทธิ์ ทำ inactive กำหนดวันหมดอายุ หรือทำ User inactive จนทำให้ Holding ไม่มี OWNER ที่ใช้งานได้อย่างน้อยหนึ่งคน การกำหนดวันหมดอายุให้ OWNER ทำได้เมื่อหลังการเปลี่ยนยังมี OWNER ที่ใช้งานได้อย่างน้อยหนึ่งคนซึ่งไม่มีวันหมดอายุ การตรวจและการเปลี่ยนต้องอยู่ในงาน Atomic เดียวกัน หากระบบต้องระงับ OWNER คนสุดท้ายด้วยเหตุความปลอดภัย ต้องมี Owner recovery Workflow ที่ลุงจืดยืนยันก่อน
- ทุกการเปลี่ยน Role, สถานะ, วันหมดอายุ หรือ Scope ของ Membership ต้องบันทึก Membership, Audit และ Outbox ภายใน MongoDB transaction เดียวตาม [ความถูกต้องของข้อมูลข้ามฐาน](system.md#ความถูกต้องของข้อมูลข้ามฐาน) หลัง commit ต้องทำให้การตัดสินสิทธิ์ พื้นที่ทำงาน และ Cache ที่ได้รับผลหมดอายุ โดย Request ถัดไปต้องตรวจสิทธิ์ปัจจุบันใหม่แบบ Fail closed ส่วนกลไก Version, ขอบเขต และ Field ให้ยึด MongoModel MCP
- การออก เพิกถอน หมดอายุ หรือตอบรับ Invitation ต้องบันทึก Invitation, Audit และ Outbox ที่เกี่ยวข้องแบบ Atomic การหมดอายุใช้ System actor และ idempotency key ของ Invitation กับ transition เพื่อให้ Request, Background job หรือ Retry ที่แข่งกันไม่สร้าง Audit/Outbox ซ้ำ การตอบรับต้องตรวจเงื่อนไข เปลี่ยน `PENDING` เป็น `ACCEPTED` และสร้าง Membership พร้อม Audit และ Outbox ภายใน MongoDB transaction เดียว

## ขั้นตอนสร้าง Holding

1. ตรวจสอบว่า User ใน Session ยัง active และผ่าน [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว)
2. ตรวจสอบข้อมูลบังคับ ต้องมีรหัสธุรกิจของ Holding และชื่อ Holding ซึ่งรองรับหลายภาษา
3. Backend สร้างรหัสถาวรของ Holding
4. ตรวจสอบรหัสธุรกิจของ Holding ซ้ำแบบ Atomic
5. ภายใน MongoDB transaction เดียว บันทึก Holding, Claim ของรหัสธุรกิจ, Membership Role OWNER ของผู้สร้าง, Audit และ Outbox
6. Commit งานทั้งหมด
7. คืนข้อมูลที่บันทึกจริงจาก Backend

ถ้าขั้นตอนใดล้มเหลว ต้องไม่เหลือ Holding, Claim, Membership, Audit หรือ Outbox ที่สร้างเพียงครึ่งเดียว

## ขั้นตอนสร้าง Company

1. ตรวจสอบ Membership และ Role ที่มีสิทธิ์สร้าง Company ใน Holding
2. ตรวจสอบว่า Holding ยัง active
3. ตรวจสอบข้อมูลบังคับ ต้องมีรหัสถาวรของ Holding รหัสธุรกิจของ Company และชื่อ Company ซึ่งรองรับหลายภาษา
4. Backend สร้างรหัสถาวรของ Company
5. ภายใน MongoDB transaction เดียว ตรวจและ Claim รหัสธุรกิจของ Company ภายใน Holding แล้วบันทึก Company, Audit และ Outbox
6. Commit งานทั้งหมด
7. คืนข้อมูลที่บันทึกจริงจาก Backend

ถ้าขั้นตอนใดล้มเหลว ต้องไม่เหลือ Company, Claim, Audit หรือ Outbox ที่สร้างเพียงครึ่งเดียว

## ขั้นตอนสร้าง Branch

1. ตรวจสอบ Membership และ Role ที่มีสิทธิ์สร้าง Branch ใน Company
2. ตรวจสอบว่า Holding และ Company ยัง active
3. ตรวจสอบข้อมูลบังคับ ต้องมีรหัสถาวรของ Holding และ Company รหัสธุรกิจของ Branch ชื่อ Branch ซึ่งรองรับหลายภาษา และเขตเวลาตาม [เขตเวลา](#เขตเวลา)
4. Backend สร้างรหัสถาวรของ Branch
5. ภายใน MongoDB transaction เดียว ตรวจและ Claim รหัสธุรกิจของ Branch ภายใน Company แล้วบันทึก Branch ใต้รหัสถาวรของ Holding และ Company ที่เป็นสายเดียวกัน พร้อม Audit และ Outbox
6. Commit งานทั้งหมด
7. คืนข้อมูลที่บันทึกจริงจาก Backend

ถ้าขั้นตอนใดล้มเหลว ต้องไม่เหลือ Branch, Claim, Audit หรือ Outbox ที่สร้างเพียงครึ่งเดียว

## สถานะการใช้งานขององค์กร

- Holding, Company และ Branch เริ่มเป็น active เมื่อสร้างสำเร็จ
- Holding OWNER เท่านั้นที่เปลี่ยนสถานะ Holding ได้ ส่วน Holding OWNER และ Holding ADMIN เปลี่ยนสถานะ Company และ Branch ภายใน Holding ได้
- พื้นที่ inactive ห้ามอ่านหรือแก้ข้อมูลธุรกิจ แต่ผู้มีสิทธิ์เปลี่ยนสถานะยังเข้าหน้าบริหารที่จำเป็น ดู Organization Audit ตาม [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว) และ activate กลับได้
- งานบริหารที่ยังทำได้ในพื้นที่ Organization effective inactive มีเฉพาะ: การเปลี่ยน Organization stored status, การดู Organization Audit ตาม [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว), การจัดการ Membership และ Invitation ภายใน Holding นั้น และการ activate กลับ Mutation อื่นในพื้นที่ Organization effective inactive ที่ไม่อยู่ในรายการนี้ ให้ปฏิเสธแบบ Fail closed จนกว่าลุงจืดจะกำหนดเพิ่ม
- Organization effective status คำนวณดังนี้: Holding effective active เมื่อ Organization stored status ของ Holding active; Company effective active เมื่อ Organization stored status ของ Company และ Holding effective active; Branch effective active เมื่อ Organization stored status ของ Branch และ Company effective active
- เมื่อ Holding inactive ให้ Company และ Branch ใต้ Holding ใช้งานไม่ได้ทันที และเมื่อ Company inactive ให้ Branch ใต้ Company ใช้งานไม่ได้ทันที โดยไม่เปลี่ยน Organization stored status ของพื้นที่ลูก
- เมื่อพื้นที่แม่กลับมา active พื้นที่ลูกจะใช้ได้เฉพาะรายการที่ Organization stored status ของตนยัง active
- ผู้มีสิทธิ์เปลี่ยนสถานะสามารถเปลี่ยน Organization stored status ของพื้นที่ลูกขณะพื้นที่แม่ inactive ได้ แต่ Organization effective status ของพื้นที่ลูกยัง inactive จนกว่าพื้นที่แม่ทุกระดับจะ active
- เมื่อเปลี่ยนเป็น inactive ต้องทำให้การตัดสินสิทธิ์ พื้นที่ทำงาน และ Cache ในขอบเขตที่ได้รับผลหมดอายุ โดย Request ถัดไปต้องตรวจสิทธิ์ปัจจุบันใหม่และใช้สิทธิ์เก่าไม่ได้ แต่คง Session Login เพื่อให้ผู้ใช้เลือกพื้นที่อื่นที่มีสิทธิ์ได้ การ activate กลับต้องเลือกพื้นที่ใหม่และห้ามนำสิทธิ์จาก Cache เดิมกลับมาใช้ ส่วนกลไก Version และ Field ให้ยึด MongoModel MCP
- ทุกการเปลี่ยนสถานะต้องบันทึก Audit ได้แก่ผู้กระทำ เวลา UTC เป้าหมาย สถานะเดิม สถานะใหม่ และเหตุผล โดยห้ามลบข้อมูลหรือ Membership

## Membership

- User มี Membership ได้ไม่เกินหนึ่งรายการต่อ Holding และมี Role เป็น OWNER, ADMIN หรือ USER
- Company และ Branch เป็น Scope แบบ Allow-list ภายใน Membership ไม่ใช่ Membership แยกอีกชุด Membership หนึ่งมี Scope ได้ศูนย์รายการขึ้นไปและผสม Company Scope กับ Branch Scope ได้ โดย Backend ต้องตรวจรายการซ้ำและตรวจว่า Company กับ Branch อยู่ใน Holding และสาย Parent ที่ถูกต้อง
- Company Scope ทำให้ผ่านเงื่อนไขด้าน Scope สำหรับข้อมูลระดับ Company นั้น และไม่ครอบคลุม Branch โดยอัตโนมัติ ยกเว้นกำหนดอย่างชัดเจนให้ครอบคลุม Branch ปัจจุบันและอนาคตทั้งหมด หากไม่ได้กำหนดเช่นนั้น ต้องมี Branch Scope ของแต่ละ Branch จึงผ่านเงื่อนไขด้าน Scope ส่วนชื่อ Field และรูปแบบจัดเก็บให้ยึด MongoModel MCP
- Branch Scope ทำให้ผ่านเงื่อนไขด้าน Scope เฉพาะ Branch ที่ระบุ การแสดง Company และ Holding ต้นสังกัดเพื่อ Navigation ไม่ได้ทำให้ผ่าน Scope ของข้อมูลระดับอื่นหรือ Branch พี่น้อง
- Scope ว่างไม่ให้เข้าถึงข้อมูลธุรกิจทุก Role รวมถึง OWNER และ ADMIN แต่ยังแสดง Holding และใช้สิทธิ์บริหารโครงสร้างตาม [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว) ได้
- การตรวจสิทธิ์แบ่ง Endpoint เป็น 4 กลุ่ม:
  1. Public Authentication Endpoint ไม่ต้องมี Session หรือ Membership เดิม แต่ต้องผ่าน Validation ตาม [กฎการเข้าสู่ระบบ](login.md)
  2. User-scoped Protected Endpoint เช่นสร้าง Holding โหลด Membership ของตน และดู Invitation ของตน ต้องมี Session และ User ที่ active แต่ไม่ต้องมี Membership เดิม แม้ข้อมูลจะอ้าง Holding การดู Invitation ของตนต้องมี Google Identity ที่ active และอีเมลตรงกับผู้รับ ส่วนการตอบรับต้องมี Invitation token ที่ใช้ได้เพิ่ม โดย Token อย่างเดียวไม่พอ
  3. Organization Administration Endpoint เช่นสร้าง Company หรือ Branch จัดการ Membership หรือ Invitation เปลี่ยน Organization stored status และดู Organization Audit ต้องตรวจ Session, User, Membership ที่ active และยังไม่หมดอายุ, สาย Parent และ Role ที่มีสิทธิ์ แต่ไม่ต้องมี Scope หรือ Domain Permission สำหรับข้อมูลธุรกิจ พื้นที่ Organization effective inactive เข้าถึงได้เฉพาะงานบริหารที่ระบุใน [สถานะการใช้งานขององค์กร](#สถานะการใช้งานขององค์กร)
  4. Business Data Endpoint ต้องตรวจ Session, User active, Membership ที่ active และยังไม่หมดอายุ, สาย Parent, Organization effective status, Scope กับ Permission ของ Domain นั้น โดยไม่ต้องมี Role ผู้บริหาร เว้นแต่ Source of Truth ของ Endpoint นั้นกำหนด Role เพิ่มโดยตรง
- Membership ที่ใช้งานได้ทำให้แสดง Holding แม้ยังไม่มี Company หรือ Branch ส่วน Membership ที่ inactive หรือหมดอายุไม่ให้เลือกเป็น Workspace
- Scope ระดับ Company ทำให้แสดง Company และ Holding ต้นสังกัดได้
- Scope ระดับ Branch ทำให้แสดง Branch พร้อม Company และ Holding ต้นสังกัดได้

## เขตเวลา

- วันเวลาที่เป็นจุดเวลาเดียวกันต้องบันทึกและส่งต่อเป็น UTC ส่วนวันที่ธุรกิจที่ไม่มีเวลาให้คงเป็นวันที่ของ Branch และห้ามแปลง timezone
- ทุก Branch ต้องมี IANA timezone ที่ถูกต้องและห้ามเป็นค่าว่างตั้งแต่สร้าง เช่น `Asia/Bangkok` โดย Frontend ใช้ timezone ที่ Browser ตรวจพบเป็นค่าเริ่มต้นเพื่อเสนอให้ผู้ใช้ตรวจหรือเปลี่ยน และ Backend ต้องตรวจสอบก่อนบันทึก
- ห้ามใช้ UTC offset เป็น timezone หลัก เพราะ offset เปลี่ยนได้ตามวันเวลา
- การแสดงวันเวลา การกำหนดวันธุรกิจ การตัดรอบ และงานตามกำหนดเวลา ต้องคำนวณจาก IANA timezone ของ Branch ที่เกี่ยวข้อง
- หาก timezone ไม่มีหรือไม่ถูกต้อง ต้องแจ้งให้ผู้มีสิทธิ์กำหนดค่า ห้าม fallback เงียบไปใช้ timezone ของ Holding, Browser, Server หรือ `Asia/Bangkok`
- OWNER และ ADMIN มีสิทธิ์เปลี่ยน timezone ของ Branch ภายใน Holding ตามสิทธิ์บริหารโครงสร้าง และทำได้เฉพาะเมื่อพื้นที่ที่เกี่ยวข้อง Organization effective active ตาม [สถานะการใช้งานขององค์กร](#สถานะการใช้งานขององค์กร) การเปลี่ยนมีผลต่อการแสดงผลและการคำนวณหลังการเปลี่ยน แต่ห้ามแก้ UTC timestamp หรือวันที่เอกสารและวันบัญชีที่ยืนยันแล้วของรายการเดิม และต้องบันทึก Audit
- หน้าหรือรายงานที่รวมหลาย Branch ต้องระบุ timezone ที่ใช้คำนวณและแสดงผลอย่างชัดเจน
