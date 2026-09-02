# กฎการเข้าสู่ระบบ

## เป้าหมาย

ให้ผู้ใช้เข้าสู่ระบบได้อย่างปลอดภัย และให้ระบบทราบว่าเป็นผู้ใช้คนใด มีสิทธิ์ระดับใด และเข้าถึง Holding, Company หรือ Branch ใดได้บ้าง

## คำศัพท์

- Authentication: การยืนยันตัวตนด้วย Google หรือ usercode+password
- Dev Login: กลไกทดสอบแยกจาก Authentication ของผู้ใช้จริง ใช้ได้เฉพาะ Local/DEV ตาม [วิธีเข้าสู่ระบบ](#วิธีเข้าสู่ระบบ)
- Session: การ Login หนึ่งครั้งต่อ Browser หรือ Device พร้อม Token family และสถานะฝั่ง Backend
- Protected Endpoint: Endpoint ที่ต้องมี Session ที่ Backend ตรวจว่ายังใช้งานได้
- Membership: ให้ยึด [Membership](organization.md#membership)
- พื้นที่ทำงาน: Holding, Company หรือ Branch ที่ผู้ใช้เลือกได้ตาม [Membership](organization.md#membership) และ [สถานะการใช้งานขององค์กร](organization.md#สถานะการใช้งานขององค์กร)

## เงื่อนไข

- เมื่อเลือกถึง Branch แล้วจึงเข้าสู่ Dashboard ที่ต้องใช้ข้อมูลระดับ Branch
- User ที่ Login สำเร็จอาจมี Membership เป็นศูนย์ หนึ่ง หรือหลายรายการ
- การสร้าง Holding, Company และ Branch รวมถึงการแสดงปุ่มเพิ่ม ต้องเป็นไปตาม [สิทธิ์ที่ยืนยันแล้ว](organization.md#สิทธิ์ที่ยืนยันแล้ว) และ Backend ต้องตรวจสิทธิ์ซ้ำทุกครั้ง

## วิธีเข้าสู่ระบบ

- วิธีเข้าสู่ระบบสำหรับผู้ใช้จริงมี 2 แบบ คือ usercode+password และ Google
- Dev Login ใช้บัญชีทดสอบ `jaturapornchai@gmail.com` และเปิดได้เฉพาะ Browser loopback กับ Frontend server ที่ bind เฉพาะ loopback พร้อม Backend Environment `dev`, explicit enable flag และ server-side control token ที่ตั้งค่าครบ ห้ามใช้ Host header, hostname หรือค่าจาก Client เพียงอย่างเดียวเป็นหลักฐานว่าเป็น Local/DEV
- Frontend แสดงปุ่ม Dev Login เฉพาะเมื่อเงื่อนไข Local/DEV ฝั่ง Server ครบ Backend ต้องปฏิเสธ Endpoint นี้แบบ Fail closed ใน Environment อื่น และห้ามส่งหรือบันทึก control token ลง Client bundle, Response หรือ Log
- Dev Login ต้องใช้ stable User identifier ที่กำหนดฝั่ง Server ตรวจว่า User ยัง active และเป็นบัญชีทดสอบที่กำหนด ห้ามสร้าง User, Google Identity, Membership, Role หรือ Scope และห้ามยกเว้นเงื่อนไข Google Identity สำหรับการสร้างองค์กร ใช้ Session, Authorization และ Audit ชุดเดียวกับ Login ปกติ
- บัญชี Demo (ปุ่ม "ทดลองใช้ระบบ (Demo)" บนหน้า Login ตั้งโดยลุงจืด 2026-09-02 ใช้ได้ทั้ง Local และ Public server): บัญชีทดลองสาธารณะ 1 บัญชีที่มีเฉพาะข้อมูลตัวอย่าง (SME ไทย เจ้าของคนเดียวหลายกิจการ) เปิดด้วย `BCAI_DEMO_LOGIN_ENABLED=true` และ `BCAI_DEMO_USERNAME` (ค่าเริ่มต้น `demo`) ฝั่ง Backend เท่านั้น Endpoint `POST /demo-login` เป็น Public ไม่มี secret และออก Session/Audit (`DEMO_LOGIN`) ชุดเดียวกับ Login ปกติ บัญชี Demo เป็น User ที่สมัครด้วย usercode ไม่มี Google Identity จึงได้รับข้อยกเว้นเงื่อนไข Google Identity เฉพาะการสร้างองค์กร (`requireOrganizationCreator`) เมื่อ Demo เปิดอยู่ ข้อมูลของ Demo ถือเป็นข้อมูลทิ้งได้ สร้างซ้ำได้ด้วย `scripts/seed-demo.mjs` และห้ามนำบัญชีนี้ไปใช้เป็น Dev Login หรือเก็บข้อมูลจริง Environment ที่ไม่เปิด flag ต้องตอบ 404 แบบ Fail closed
- Google Login ต้องตรวจลายเซ็น issuer, audience, expiry และ verified email ของ token ฝั่ง Backend และผูก User ด้วยคู่ `issuer + subject` ที่ไม่ซ้ำ ไม่ใช่อีเมลอย่างเดียว
- Google Identity ถือว่ายังใช้งานได้เมื่อ Backend เคยตรวจสอบแล้วและสถานะการเชื่อมโยงในระบบยัง active ไม่ถูกเพิกถอนหรือยกเลิก โดยไม่ต้องยืนยัน Google ใหม่ทุกครั้งและไม่ขึ้นกับวิธี Login ของ Session ปัจจุบัน
- Google Identity หนึ่งรายการผูกได้กับ User เดียว เมื่อไม่พบคู่ `issuer + subject` ให้สร้าง User ใหม่กับ Google Identity ภายใน MongoDB transaction เดียว ห้ามผูก User เดิมอัตโนมัติด้วยอีเมล แม้อีเมลผ่านการยืนยันแล้ว หากอีเมลชนกับ User เดิมให้ rollback และตอบข้อความกลาง การผูกกับ User เดิมต้องเป็น Account-linking Workflow แยกซึ่ง User ต้องมี Session ที่ active และยืนยันทั้งบัญชีเดิมกับ Google ใหม่อีกครั้ง ปัจจุบันยังไม่มี Workflow นี้จึงต้องปฏิเสธการผูก และตอบกลับด้วยข้อความตาม [ข้อความเมื่อเกิดข้อผิดพลาด](#ข้อความเมื่อเกิดข้อผิดพลาด)
- Google Login อาจสร้าง User ได้แต่ห้ามสร้าง Membership อัตโนมัติ ส่วนการสร้างหรือเปิด Membership ให้ยึด [สิทธิ์ที่ยืนยันแล้ว](organization.md#สิทธิ์ที่ยืนยันแล้ว)

## รหัสผ่านและการกู้บัญชี

- ห้ามใช้รหัสผ่านเริ่มต้นหรือรหัส Reset ร่วมกัน เช่น `12345`
- User ส่วนกลางใหม่ต้องเริ่มจาก Google Login ที่ Backend ยืนยันสำเร็จ OWNER และ ADMIN ห้ามสร้าง User ส่วนกลางแทนผู้อื่น ส่วนการเพิ่มสิทธิ์ให้ยึด [สิทธิ์ที่ยืนยันแล้ว](organization.md#สิทธิ์ที่ยืนยันแล้ว)
- หลังสร้าง User แล้ว User ตั้ง usercode และ password ได้ผ่านลิงก์ `SET_PASSWORD` แบบใช้ครั้งเดียว โดยกำหนด usercode ได้เฉพาะเมื่อ User ยังไม่มี usercode ส่วนลิงก์ `RESET_PASSWORD` เปลี่ยนได้เฉพาะ password และห้ามเปลี่ยน usercode
- Backend ต้อง normalize usercode ตามลำดับ trim ช่องว่างหัวท้าย แปลงเป็นพิมพ์เล็ก แล้วตรวจด้วยรูปแบบ `^[a-z0-9._-]{3,64}$` ต้องไม่ซ้ำทั้งระบบ และเมื่อกำหนดแล้วห้ามเปลี่ยน
- User เป็นผู้ขอลิงก์ตั้งหรือ Reset รหัสผ่านด้วย verified email ที่ผูกกับ Google Identity ซึ่งยัง active และ User ต้อง active ทั้งตอนขอและตอนใช้ลิงก์ Backend ต้องส่งลิงก์ไปยังอีเมลนั้นและตอบ HTTP status กับข้อความกลางเหมือนกันไม่ว่าจะพบบัญชีหรือไม่ User inactive ต้องไม่ได้ Token ใหม่และ Token เดิมของ User นั้นใช้ไม่ได้
- ลิงก์ตั้งหรือ Reset รหัสผ่านต้องสุ่ม เก็บเฉพาะค่า hash ใช้ได้ครั้งเดียว และใช้ได้เฉพาะเมื่อ User กับ Google Identity ยัง active, Token ยังไม่ถูกใช้ ไม่ถูกเพิกถอน และเวลาปัจจุบัน UTC ยังไม่ถึงเวลาหมดอายุ ซึ่งต้องเท่ากับ 30 นาทีหลังสร้าง ลิงก์ใหม่ต้องเพิกถอนลิงก์เดิมที่ยังใช้ได้ของ User ก่อนแบบ Atomic
- วิธีเปลี่ยน password ที่รองรับในปัจจุบันมีเฉพาะ `SET_PASSWORD` สำหรับการตั้งครั้งแรกและ `RESET_PASSWORD` สำหรับเปลี่ยน password เดิมผ่านลิงก์ ห้ามสร้าง Authenticated `CHANGE_PASSWORD` Endpoint เองจนกว่าจะมี Workflow ที่กำหนดการยืนยัน password เดิม การ Audit และผลต่อ Session
- เมื่อ User ตั้งหรือ Reset รหัสผ่านสำเร็จ ต้องบันทึก Audit และจัดการ Session ตาม [ความปลอดภัยของ Session](#ความปลอดภัยของ-session)
- User ที่ใช้ Google อย่างเดียวไม่จำเป็นต้องมีรหัสผ่าน
- Password แบบ single-factor ต้องยาวอย่างน้อย 15 ตัว ระบบต้องยอมรับความยาวอย่างน้อย 64 ตัว ห้ามตัดค่าเงียบ และต้องปฏิเสธรหัสที่พบบ่อยหรือเคยรั่วไหล ความยาวสูงสุดที่รับและข้อความเมื่อยาวกว่าความยาวสูงสุด ให้กำหนดใน MongoModel Workflow `ตั้งและ Reset รหัสผ่าน` ก่อนเปิด Endpoint โดยเมื่อยาวกว่าความยาวสูงสุดต้องปฏิเสธด้วยข้อความชัดเจนและห้ามตัดค่าทุกกรณี
- Endpoint Login และขอลิงก์ตั้งหรือ Reset รหัสผ่านต้องมี Rate limit ฝั่ง Backend โดยห้ามใช้ข้อความหรือผลตอบกลับที่เปิดเผยว่ามีบัญชีหรือไม่
- ก่อนเปิด Endpoint ในทุก Environment รายละเอียด Rate limit ได้แก่ key, window, threshold และ failure policy ต้องถูกกำหนดใน MongoModel Workflow `เข้าสู่ระบบ` และ `ตั้งและ Reset รหัสผ่าน` ตาม Endpoint ที่เกี่ยวข้อง ส่วนแหล่งตรวจรหัสผ่านรั่วไหล, cache, timeout กับพฤติกรรมเมื่อผู้ให้บริการล้มเหลวต้องอยู่ใน Workflow `ตั้งและ Reset รหัสผ่าน` Local/UAT ใช้ Test policy แยกได้เมื่อระบุชื่อและค่าชัดเจน แต่ห้ามปิด Control หาก Workflow ยังไม่ครบให้หยุด ห้ามเลือกค่าเองหรือเปิด Endpoint

## สถานะบัญชี

- `User active`: เข้าสู่ระบบได้เมื่อเงื่อนไขของวิธี Login ผ่าน
- `User inactive`: ห้ามเข้าสู่ระบบและต้องยกเลิก Session ทุกอุปกรณ์ การทำ User inactive ต้องไม่ทำให้ Holding ขาด OWNER ที่ใช้งานได้ตาม [สิทธิ์ที่ยืนยันแล้ว](organization.md#สิทธิ์ที่ยืนยันแล้ว)
- OWNER และ ADMIN ระดับ Holding ไม่มีสิทธิ์เปลี่ยนสถานะ User ส่วนกลางหรือ Google Identity ปัจจุบันยังไม่มี Platform Security Workflow ที่ระบุผู้กระทำ การ inactive/reactivate/revoke, Audit, ผลต่อ Session และ Owner recovery จึงห้ามเปิด Mutation เหล่านี้จนกว่าลุงจืดจะยืนยัน Workflow

## ความปลอดภัยของ Session

- Access token มีอายุ 15 นาที Session หมดอายุเมื่อไม่ใช้งาน 12 ชั่วโมง และมีอายุสูงสุด 12 ชั่วโมงนับจาก Login โดยการ Refresh หรือ Activity ห้ามเลื่อนอายุสูงสุด 12 ชั่วโมง (ปรับจาก 30 นาที/8 ชั่วโมง ตามคำสั่งลุงจืด 2026-09-02 เพื่อไม่ต้อง Login ใหม่บ่อย)
- Activity ที่ต่ออายุ idle timeout คือ Protected Request ที่ Backend ตรวจ Session และ User สำเร็จสำหรับ Session นั้น Public Endpoint, Health check, CORS preflight และเวลาจาก Client ไม่ถือเป็น Activity
- Access token ต้องอ้าง Session ฝั่ง Backend และ Token ที่ตรวจลายเซ็นผ่านอย่างเดียวไม่เพียงพอ ทุก Protected Request ต้องพบ Session ที่ active และ User ที่ active ใน Backend หาก Session Store สูญหายหรือหา Session ไม่พบให้ Fail closed และ Login ใหม่
- Refresh token ใช้หมุนค่าได้ครั้งเดียวภายใน Session family เมื่อออกค่าใหม่ต้องเพิกถอนค่าเดิมทันที การนำค่าเดิมกลับมาใช้ต้องยกเลิก Session family นั้นทั้งหมด
- Browser ต้องเก็บ Access token ใน memory เท่านั้น และเก็บ Refresh token ใน HttpOnly Secure SameSite=Lax cookie หลัง Reload ให้ใช้ Refresh token ขอ Access token ใหม่ ห้ามเก็บ Token ทั้งสองชนิดใน localStorage และห้ามลด `Secure` เพื่อแก้ปัญหา Local Runtime
- Endpoint ที่อาศัย Cookie และเปลี่ยนสถานะ เช่น Refresh และ Logout ต้องใช้ Method ที่ไม่ใช่ GET ตรวจ Origin ที่ Backend และบังคับ CSRF protection ฝั่ง Backend
- Logout ปกติเพิกถอน Access token, Refresh token และ Session family ปัจจุบันที่ Backend พร้อมล้าง Cookie ส่วนการ Logout ทุกอุปกรณ์ต้องเป็น Workflow แยกที่ระบุชัด
- การตั้งหรือ Reset รหัสผ่าน และการทำ User เป็น inactive ต้องยกเลิก Session ทุกอุปกรณ์
- ผลต่อพื้นที่ทำงานและ Cache เมื่อ Membership หรือสถานะองค์กรเปลี่ยน ให้ยึด [Membership](organization.md#membership) และ [สถานะการใช้งานขององค์กร](organization.md#สถานะการใช้งานขององค์กร)
- Public Authentication Endpoint ที่ไม่ต้องมี Session เดิมมีเฉพาะ Password Login, Google Login/Callback, การขอหรือใช้ลิงก์ `SET_PASSWORD`/`RESET_PASSWORD` และ Dev Login ที่ผ่านเงื่อนไขพิเศษตาม [วิธีเข้าสู่ระบบ](#วิธีเข้าสู่ระบบ) ส่วน Refresh ไม่ใช่ Public Endpoint และต้องมี Refresh token พร้อม Session family ที่ยัง active ฝั่ง Backend
- ทุก Protected Request ต้องตรวจ Session และ User จาก Backend ห้ามเชื่อตัวตน Role หรือพื้นที่จาก Client ส่วนการแบ่ง Endpoint ที่ต้องมี Membership และการตรวจ Scope, Permission กับสถานะองค์กรให้ยึด [Membership](organization.md#membership)
- กลไก CSRF, allow-list ของ Origin และ Local HTTPS ที่ทำให้ Cookie `Secure` ใช้งานได้ต้องระบุใน MongoModel Workflow `Refresh Logout และเพิกถอน Session` ก่อนเปิด Session Endpoint ในทุก Environment ห้ามลด `Secure` หรือข้าม Origin/CSRF เพื่อให้ Local Runtime ทำงาน

## ขั้นตอน Login

1. ผู้ใช้เลือกวิธีตาม [วิธีเข้าสู่ระบบ](#วิธีเข้าสู่ระบบ)
2. Backend ตรวจ credential และสถานะบัญชีตามวิธีที่เลือก
3. เมื่อสำเร็จให้สร้าง Session ตาม [ความปลอดภัยของ Session](#ความปลอดภัยของ-session)
4. โหลด Membership ตาม [Membership](organization.md#membership)
5. ถ้าไม่มี Membership ให้ห้ามเข้าถึงข้อมูลธุรกิจ และแสดงทางเลือกสร้าง Holding เฉพาะเมื่อผ่าน [สิทธิ์ที่ยืนยันแล้ว](organization.md#สิทธิ์ที่ยืนยันแล้ว)
6. ถ้ามี Membership ให้แสดงพื้นที่ตามลำดับ Holding -> Company -> Branch เท่าที่มีข้อมูลและมีสิทธิ์
7. ตรวจทุก Protected Request ตาม [ความปลอดภัยของ Session](#ความปลอดภัยของ-session) และตรวจพื้นที่ตาม [Membership](organization.md#membership)

## ข้อความเมื่อเกิดข้อผิดพลาด

| กรณี | ข้อความสำหรับผู้ใช้ |
|---|---|
| ชื่อผู้ใช้หรือรหัสผ่านผิด หรือ User inactive | ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง |
| ไม่มีสิทธิ์ในพื้นที่ | คุณไม่มีสิทธิ์ใช้งาน Holding บริษัท หรือสาขานี้ |
| Session หมดอายุ | กรุณาเข้าสู่ระบบอีกครั้ง |
| Google Login ที่อีเมลชนกับ User เดิมแต่ไม่พบคู่ `issuer + subject` | ไม่สามารถใช้ Google Account นี้เข้าสู่ระบบได้ |

## ตัวอย่างข้อมูล UAT

> ชื่อบัญชีต่อไปนี้เป็นตัวอย่างสำหรับทดสอบ ไม่ใช่รหัสบัญชีบังคับและห้ามใช้เป็นรหัสผ่านจริง

บัญชี Fixture สำหรับ UAT เป็นข้อยกเว้นด้านการเตรียมข้อมูล สร้างได้เฉพาะ Local/UAT ที่เปิดด้วย Backend configuration ชัดเจนและ Backend ต้องปฏิเสธการ Seed ใน Production Fixture Seed สร้างได้เฉพาะ User, credential และข้อมูลสิทธิ์ที่จำเป็นต่อกรณีทดสอบที่ระบุ ห้ามสร้าง Google Identity active ปลอม และห้ามใช้ Fixture เป็นบัญชี Dev Login

บัญชี Dev Login ต้องมี User และ Google Identity active จาก Google Login จริงใน Environment นั้นอยู่ก่อนแล้ว Dev Login ห้าม Seed หรือสร้างสิ่งเหล่านี้เอง

| ผู้ใช้ | หน้าที่ |
|---|---|
| uat_holding_owner | ทดสอบ Role OWNER |
| uat_holding_admin | ทดสอบ Role ADMIN |
| uat_company_user | ทดสอบ Role USER ที่มี Company Scope |
| uat_branch_user | ทดสอบ Role USER ที่มี Branch Scope |
| uat_no_access | ทดสอบ User ที่ไม่มี Membership |

รหัสผ่านของบัญชี UAT ต้องส่งผ่านช่องทางที่ปลอดภัย ห้ามบันทึกค่าที่ใช้งานจริงไว้ในเอกสารนี้

## กรณีทดสอบ UAT

1. Login ด้วยข้อมูลถูกต้องต้องสำเร็จ
2. รหัสผ่านผิดต้องไม่สร้าง Session
3. User inactive ต้องเข้าสู่ระบบไม่ได้
4. ทดสอบกรณีสำเร็จ หมดอายุ ใช้ซ้ำ และถูกเพิกถอนตาม [รหัสผ่านและการกู้บัญชี](#รหัสผ่านและการกู้บัญชี)
5. ทดสอบทั้งกรณีอนุญาตและปฏิเสธการสร้างองค์กรด้วย User ที่มี Google Identity active ตาม [สิทธิ์ที่ยืนยันแล้ว](organization.md#สิทธิ์ที่ยืนยันแล้ว)
6. ผู้ใช้ต้องเปิดข้อมูลข้าม Holding ที่ไม่มีสิทธิ์ไม่ได้
7. Refresh หน้าแล้ว Session และพื้นที่ทำงานต้องถูกต้อง
8. Logout แล้ว Session เดิมต้องใช้งานไม่ได้
9. เปิด URL โดยตรงโดยไม่มีสิทธิ์ต้องถูกปฏิเสธจาก Backend
10. Error ต้องไม่เปิดเผย password hash, token หรือข้อมูลภายในระบบ
11. Given User A ที่ active มี Google Identity active และไม่มี Membership ต้องสร้าง Holding ของตนเองได้พร้อม Role OWNER และตอบรับ Invitation ที่ระบุอีเมลของ A ซึ่งออกโดย Holding อื่นได้ แต่ถ้า Membership ที่ได้รับมี Scope เริ่มต้นว่าง ต้องเปิดข้อมูลธุรกิจของ Holding ที่รับเชิญไม่ได้ ส่วน Fixture ที่ไม่มี Google Identity ใช้ทดสอบได้เฉพาะพฤติกรรมที่ไม่ต้องใช้ Identity นี้
12. Refresh token ที่ถูกใช้ซ้ำต้องยกเลิก Session family และ Refresh ต้องไม่เลื่อนอายุสูงสุด 12 ชั่วโมง
13. Redis ไม่มี Session record ต้อง Fail closed และให้ Login ใหม่
14. Dev Login ต้องถูกปฏิเสธเมื่อไม่ใช่ loopback, Environment ไม่ใช่ dev, flag ปิด หรือ server-side control token ไม่ถูกต้อง
15. Endpoint ที่ใช้ Cookie เปลี่ยนสถานะต้องปฏิเสธ Method หรือ Origin ที่ไม่อนุญาต
16. Google Login ที่อีเมลชน User เดิมแต่ไม่พบคู่ `issuer + subject` ต้องไม่ผูกบัญชีอัตโนมัติและต้องไม่เหลือ User หรือ Google Identity ที่สร้างครึ่งเดียว
17. User inactive ต้องขอหรือใช้ลิงก์ตั้งหรือ Reset รหัสผ่านไม่ได้ โดยผลตอบกลับคำขอยังคงเป็นข้อความกลาง
18. Local/UAT ต้องใช้ Rate limit Test policy ที่ระบุชัด ส่วน Production ต้องถูกปฏิเสธการเปิด Endpoint เมื่อ Security control ที่กำหนดยังไม่ครบ
19. Fixture Seed ต้องถูกปฏิเสธใน Production และต้องสร้าง Google Identity active หรือใช้ Dev Login ไม่ได้
