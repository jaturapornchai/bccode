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
- Holding Owner: User ที่มี Membership Role OWNER ผู้สร้าง Holding ได้ Role นี้โดยอัตโนมัติ และ Holding หนึ่งมี OWNER ได้มากกว่าหนึ่งคน
- Holding Admin: User ที่มี Membership ระดับ Holding และ Role เป็น ADMIN

## รหัสประจำตัว

- Holding, Company และ Branch ต้องมีรหัสถาวรที่ Backend สร้าง ห้ามเปลี่ยน และใช้สำหรับ Relation กับ Database routing ส่วนโครงสร้างจัดเก็บให้ยึด [ลำดับความน่าเชื่อถือของเอกสาร](README.md#ลำดับความน่าเชื่อถือของเอกสาร)
- แต่ละระดับมีรหัสธุรกิจ โดย Holding ห้ามซ้ำทั้งระบบ Company ห้ามซ้ำภายใน Holding และ Branch ห้ามซ้ำภายใน Company
- รหัสธุรกิจต้องตัดช่องว่างหัวท้ายและเปรียบเทียบแบบไม่แยกตัวพิมพ์เล็กใหญ่ การเปลี่ยนรหัสต้องเป็นไปตาม [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว) ตรวจความซ้ำแบบ Atomic บันทึก Audit และห้ามนำรหัสที่เคยใช้กลับมาใช้ซ้ำภายในขอบเขตเดิม

## สิทธิ์ที่ยืนยันแล้ว

- เงื่อนไขด้านตัวตนสำหรับการสร้าง Holding, Company และ Branch คือ User ต้อง active และมี Google Identity ที่ใช้งานได้ตาม [วิธีเข้าสู่ระบบ](login.md#วิธีเข้าสู่ระบบ)
- User ที่ผ่านเงื่อนไขด้านตัวตนสร้าง Holding ได้ และผู้สร้างได้รับ Role OWNER ภายในงาน Atomic เดียวกับการสร้าง Holding
- OWNER และ ADMIN สร้าง Company และ Branch รวมถึงบริหารโครงสร้างภายใน Holding ที่เลือกและยัง active ได้
- OWNER และ ADMIN ไม่ได้รับสิทธิ์อ่านหรือแก้ข้อมูลธุรกรรมโดยอัตโนมัติ การเข้าถึงธุรกรรมให้ยึด [Membership](#membership) และ Permission ตามเอกสาร Domain ที่เกี่ยวข้อง
- Membership ทุก Role ยกเว้น OWNER ผู้สร้าง Holding ต้องเริ่มจาก Invitation
- OWNER และ ADMIN เชิญ USER รวมถึงเปลี่ยนสถานะ วันหมดอายุ และ Scope Company/Branch ของ USER ได้
- OWNER เท่านั้นที่เชิญ แต่งตั้ง ลดสิทธิ์ หรือถอน OWNER และ ADMIN
- Invitation ต้องระบุ Holding และ Role โดย Scope เริ่มต้นอาจว่างได้ตาม [Membership](#membership) ผู้รับต้องตอบรับด้วย Google Account ซึ่ง verified email ตรงกับอีเมลผู้รับหลังตัดช่องว่างและเปรียบเทียบแบบไม่แยกตัวพิมพ์เล็กใหญ่ ห้ามตีความ dot หรือ plus alias เอง แล้วจึงสร้าง Membership ที่ active แบบ Atomic
- Invitation ใช้ได้ครั้งเดียว หมดอายุภายใน 72 ชั่วโมง และเพิกถอนได้ การออก Invitation ใหม่ให้ผู้รับและ Holding เดิมต้องเพิกถอนรายการเดิม
- ห้ามถอน ลดสิทธิ์ ทำ inactive หรือกำหนดวันหมดอายุที่ทำให้ Holding ไม่มี OWNER ที่ active อย่างน้อยหนึ่งคน
- ทุกการเปลี่ยน Role, Membership, Invitation หรือ Scope ต้องทำแบบ Atomic บันทึก Audit และยกเลิกพื้นที่ทำงานกับ Cache สิทธิ์ที่ได้รับผลทันที

## ขั้นตอนสร้าง Holding

1. ตรวจสอบว่า User Login บัญชียัง active และผ่าน [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว)
2. ตรวจสอบข้อมูลบังคับ ต้องมีรหัสธุรกิจของ Holding และชื่อ Holding ซึ่งรองรับหลายภาษา
3. Backend สร้างรหัสถาวรของ Holding
4. ตรวจสอบรหัสธุรกิจของ Holding ซ้ำแบบ Atomic
5. บันทึก Holding
6. Backend สร้าง Membership ระดับ Holding และ Role OWNER ให้ผู้สร้างโดยอัตโนมัติ ภายในงานเดียวกับการสร้าง Holding
7. บันทึก Audit
8. คืนข้อมูลที่บันทึกจริงจาก Backend

ถ้าขั้นตอนใดล้มเหลว ต้องไม่เหลือ Holding หรือ Membership ที่สร้างเพียงครึ่งเดียว

## ขั้นตอนสร้าง Company

1. ตรวจสอบ Membership และ Role ที่มีสิทธิ์สร้าง Company ใน Holding
2. ตรวจสอบว่า Holding ยัง active
3. ตรวจสอบข้อมูลบังคับ ต้องมีรหัสถาวรของ Holding รหัสธุรกิจของ Company และชื่อ Company ซึ่งรองรับหลายภาษา
4. Backend สร้างรหัสถาวรของ Company
5. ตรวจสอบรหัสธุรกิจของ Company ซ้ำภายใน Holding เดียวกันแบบ Atomic
6. บันทึก Company ใต้รหัสถาวรของ Holding ที่ถูกต้อง
7. บันทึก Audit
8. คืนข้อมูลที่บันทึกจริงจาก Backend

## ขั้นตอนสร้าง Branch

1. ตรวจสอบ Membership และ Role ที่มีสิทธิ์สร้าง Branch ใน Company
2. ตรวจสอบว่า Holding และ Company ยัง active
3. ตรวจสอบข้อมูลบังคับ ต้องมีรหัสถาวรของ Holding และ Company รหัสธุรกิจของ Branch ชื่อ Branch ซึ่งรองรับหลายภาษา และเขตเวลาตาม [เขตเวลา](#เขตเวลา)
4. Backend สร้างรหัสถาวรของ Branch
5. ตรวจสอบรหัสธุรกิจของ Branch ซ้ำภายใน Company เดียวกันแบบ Atomic
6. บันทึก Branch ใต้รหัสถาวรของ Holding และ Company ที่เป็นสายเดียวกัน
7. บันทึก Audit
8. คืนข้อมูลที่บันทึกจริงจาก Backend

## สถานะการใช้งานขององค์กร

- Holding, Company และ Branch เริ่มเป็น active เมื่อสร้างสำเร็จ
- Holding OWNER เท่านั้นที่เปลี่ยนสถานะ Holding ได้ ส่วน Holding OWNER และ Holding ADMIN เปลี่ยนสถานะ Company และ Branch ภายใน Holding ได้
- พื้นที่ inactive ห้ามอ่านหรือแก้ข้อมูลธุรกิจ แต่ผู้มีสิทธิ์เปลี่ยนสถานะยังเข้าหน้าบริหารที่จำเป็น ดู Audit และ activate กลับได้
- เมื่อ Holding inactive ให้ Company และ Branch ใต้ Holding ใช้งานไม่ได้ทันที และเมื่อ Company inactive ให้ Branch ใต้ Company ใช้งานไม่ได้ทันที โดยไม่เปลี่ยนสถานะที่บันทึกไว้ของพื้นที่ลูก
- เมื่อพื้นที่แม่กลับมา active พื้นที่ลูกจะใช้ได้เฉพาะรายการที่สถานะของตนยัง active
- เมื่อเปลี่ยนเป็น inactive ต้องยกเลิกพื้นที่ทำงานและ Cache สิทธิ์ในขอบเขตที่ได้รับผลทันที แต่คง Session Login เพื่อให้ผู้ใช้เลือกพื้นที่อื่นที่มีสิทธิ์ได้ การ activate กลับต้องเลือกพื้นที่ใหม่และห้ามนำสิทธิ์จาก Cache เดิมกลับมาใช้
- ทุกการเปลี่ยนสถานะต้องบันทึก Audit ได้แก่ผู้กระทำ เวลา UTC เป้าหมาย สถานะเดิม สถานะใหม่ และเหตุผล โดยห้ามลบข้อมูลหรือ Membership

## Membership

- User มี Membership ได้ไม่เกินหนึ่งรายการต่อ Holding และมี Role เป็น OWNER, ADMIN หรือ USER
- Company และ Branch เป็น Scope ภายใน Membership ไม่ใช่ Membership แยกอีกชุด
- Backend ต้องตรวจ Membership ที่ active และยังไม่หมดอายุก่อนทุก Request และเมื่อ Request เข้าถึง Company, Branch หรือข้อมูลธุรกรรม ต้องตรวจ Scope และ Permission ที่เกี่ยวข้องด้วย
- การตีความ Role และสิทธิ์บริหารโครงสร้างให้ยึด [สิทธิ์ที่ยืนยันแล้ว](#สิทธิ์ที่ยืนยันแล้ว) โดย Scope ว่างต้องไม่หมายถึงเข้าถึงธุรกรรมทั้งหมด
- Membership ทำให้แสดง Holding ได้ แม้ยังไม่มี Company หรือ Branch
- Scope ระดับ Company ทำให้แสดง Company และ Holding ต้นสังกัดได้
- Scope ระดับ Branch ทำให้แสดง Branch พร้อม Company และ Holding ต้นสังกัดได้

## เขตเวลา

- วันเวลาที่เป็นจุดเวลาเดียวกันต้องบันทึกและส่งต่อเป็น UTC ส่วนวันที่ธุรกิจที่ไม่มีเวลาให้คงเป็นวันที่ของ Branch และห้ามแปลง timezone
- ทุก Branch ต้องกำหนด IANA timezone เช่น `Asia/Bangkok` ก่อนใช้งาน โดย Frontend ใช้ timezone ที่ Browser ตรวจพบเป็นค่าเริ่มต้นเพื่อเสนอให้ผู้ใช้ตรวจหรือเปลี่ยน และ Backend ต้องตรวจสอบก่อนบันทึก
- ห้ามใช้ UTC offset เป็น timezone หลัก เพราะ offset เปลี่ยนได้ตามวันเวลา
- การแสดงวันเวลา การกำหนดวันธุรกิจ การตัดรอบ และงานตามกำหนดเวลา ต้องคำนวณจาก IANA timezone ของ Branch ที่เกี่ยวข้อง
- หาก timezone ไม่มีหรือไม่ถูกต้อง ต้องแจ้งให้ผู้มีสิทธิ์กำหนดค่า ห้าม fallback เงียบไปใช้ timezone ของ Holding, Browser, Server หรือ `Asia/Bangkok`
- การเปลี่ยน timezone มีผลต่อการแสดงผลและการคำนวณหลังการเปลี่ยน แต่ห้ามแก้ UTC timestamp หรือวันที่เอกสารและวันบัญชีที่ยืนยันแล้วของรายการเดิม และต้องบันทึก Audit
- หน้าหรือรายงานที่รวมหลาย Branch ต้องระบุ timezone ที่ใช้คำนวณและแสดงผลอย่างชัดเจน
