# ชื่อเมนูตรงกับหน้าจอและสถานะรอพัฒนา (2026-09-09)

- **วัตถุประสงค์:** ป้องกันชื่อเมนูภาษาไทยเปลี่ยนความหมายหลังโหลด dictionary และแยกเมนูที่ยังไม่มีหน้าจอจริง
- **สาเหตุ:** `menuText` เดิมเลือก dictionary ก่อนชื่อในผังเมนู ทำให้แคชเก่าแสดงใบสั่งขายเป็นใบเสนอราคา/ใบแจ้งหนี้; เทสต์เดิมตรวจเพียงว่ามี key และไอคอน
- **แบบแผน:** ชื่อไทยยึด `label.th`; ภาษาอื่นยังอ่าน dictionary ก่อน fallback และต้องซิงค์คอลัมน์ไทยใน TSV ด้วย (`frontend/src/lib/menu-data.ts:485`, `frontend/src/lib/menu-data.test.ts:71`)
- **หน้าผู้ใช้งาน:** `/useraccessaudit` แสดงสิทธิ์ปัจจุบัน ไม่ใช่ประวัติ login; ใช้ชื่อ “ตรวจสอบสถานะผู้ใช้งาน” ตามข้อมูลที่หน้าจอโหลด (`frontend/src/app/system-settings/system-settings-screen.tsx:7379`, `frontend/src/lib/menu-data.ts:380`)
- **สถานะ:** คง 171 เมนู โดย 153 เมนูที่ไม่มี custom screen หรือ settings config แสดง “รอพัฒนา”; 18 เมนูมีหน้าจอเชื่อมแล้ว แต่ไม่ใช่หลักฐานว่าธุรกรรมครบวงจรผ่าน UAT (`frontend/src/lib/menu-screen-status.ts:13`) (ลบกลุ่มเอกสารและบิล 4 เมนู, กลุ่มอนุมัติ 7 เมนู, กลุ่มหน้าร้าน POS 4 เมนู, กลุ่มสมาชิก คูปอง โปรโมชัน 4 เมนู, 6 เมนูย่อยสินค้า และ 3 เมนูการเงิน/ธนาคารที่รอพัฒนาออก)
- **การเพิ่มหน้าจอ:** เพิ่ม custom route ใน dispatcher และ `CUSTOM_MENU_SCREEN_ROUTES` พร้อมกัน เทสต์เปรียบเทียบทั้งสอง; settings route ใช้ config ที่มีอยู่โดยตรง ไม่ต้องทำรายการซ้ำ
- **การแสดงผล:** ใช้ `MenuPendingBadge` ในเมนูซ้าย/บน ผลค้นหา และทางลัด ใช้ theme tokens, ตัวอักษร 0.9rem และไม่ปิดปุ่มเพราะสถานะนี้ (`frontend/src/app/menu/menu-pending-badge.tsx:6`)
- **วิธีตรวจ:** Vitest ชุดเมนู/settings/usage/status + `npx tsc --noEmit`; `cd frontend` แล้ว `npx playwright test e2e/menu-consistency.spec.ts` (ต้องมี local frontend/backend และบัญชี Demo ตัวอย่างพร้อมใช้งาน)
- **ข้อจำกัด:** E2E ตรวจการนำทางและภาพ 1600/1280/1024/768 × light/dark โดยกดสลับธีมจริง ไม่สร้าง/แก้ข้อมูลธุรกิจ; ป้ายไม่ใช่การรับรอง backend หรือบัญชี และไม่เปลี่ยนสิทธิ์การเข้าถึง
