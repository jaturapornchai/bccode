---
date: 2026-09-12
status: deployed
tags: [bc-account, deployment, ui, full-screen-search, copy-dom, general-ledger]
---

# Deploy ระบบค้นหาผังบัญชีแบบเต็มจอและ Copy DOM สู่ Production (account.bcaicloud.com)

## วัตถุประสงค์และผลการปล่อยระบบ

ปล่อยรีลีส `r20260912-search-dom-1` สู่ Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) วันที่ 12 กันยายน 2026 ตามคำสั่งลุงจืด โดยครอบคลุมทั้งระบบค้นหาผังบัญชีแบบเต็มจอ และการเปิดใช้งานวิดเจ็ต Copy DOM บน Production

## สรุปรายการฟังก์ชันที่ปล่อยขึ้น Production

1. **ระบบค้นหาผังบัญชีแบบเต็มจอ (Full-Screen Chart of Accounts Search Dialog)**:
   - เปิดผ่านปุ่ม `[ 🔍 ]`, การคลิกกล่อง `AccountSelect`, หรือกดคีย์ลัด `F2` / `Space`
   - หน้าต่างขนาดใหญ่ (94vw / 88vh) พร้อมปุ่ม Toggle ขยายเต็มจอ (Fullscreen 100vw / 100vh)
   - ช่องค้นหา Hero Search รองรับรหัสบัญชี, ชื่อบัญชีภาษาไทย, ชื่อภาษาอังกฤษ และหมวดบัญชี
   - แท็บกรอง 5 หมวดบัญชีมาตรฐาน พร้อม Badge ตัวเลขนับจำนวนแบบเรียลไทม์
   - แสดงตารางลำดับชั้น มีการเยื้องตาม Level ด้วยสัญลักษณ์ `└─ `, Badge บันทึกรายการได้/บัญชีคุม, และด้านปกติ เดบิต/เครดิต
   - การนำทางด้วยคีย์บอร์ด: `↑` / `↓` เลื่อนแถว, `Enter` เลือก, ดับเบิ้ลคลิกเลือก, `Esc` ปิด
   - รองรับทั้ง Single-Select และ Multi-Select (ใช้ใน Statement Designer)
   - Zero Regression: คง `<select aria-label={label}>` ไว้ใน DOM เพื่อความเข้ากันได้ 100% กับ Playwright E2E
2. **วิดเจ็ต Copy DOM บน Production (`DevDomInspector`)**:
   - ปลดล็อคเงื่อนไข `NODE_ENV !== 'development'` ให้ทำงานบน Production ตามคำสั่งลุงจืด
   - ปุ่มลอยมุมล่างซ้าย `Copy DOM [Alt+คลิก]`
   - โหมด Toggle และโหมดกดปุ่ม `Alt` ค้าง
   - ดักจับ Capture Phase (`useCapture: true`) ป้องกันการกดโดนลิงก์/ปุ่มระหว่างคัดลอก
   - คัดลอก `outerHTML` ลง Clipboard พร้อมแจ้งเตือน Toast และ Tooltip แสดง Tag/ID/Class

## Production Artifacts & Release Details

| บริการ | Image |
|---|---|
| mainapi / worker | `bcai-account-mainapi:r20260912-search-dom-1` |
| frontend | `bcai-account-frontend:r20260912-search-dom-1` |

- **Release Directory**: `/opt/bcai-account/releases/r20260912-search-dom-1`
- **Images Archive**: `images.tar` (597,855,232 bytes, SHA-256: `e56f014081d99c6e4ec63f61f4589a32e85f2ada0fa7707fcc1f16ed223438f4`)
- **Preflight Backups**:
  - `mongo.archive.gz`: 71,945 bytes (SHA-256: `583bd3391183dd1675e6e9398443124991c62b004eba283686ae257a8010fd4b`)
  - `postgres-all.sql`: 723,955 bytes (SHA-256: `edcc4b6883a26cada5df44e6379bb316a1d23ebffe90dc8c172b2798624a0be2`)
  - `runtime-config.tar.gz`: 7,266 bytes (SHA-256: `cb24d11d0957d423e176113d34a8506de034187cce0dfbb1c977a57fb746545a`)
  - `release.env.before`: บันทึกสถานะก่อน deploy ไว้อย่างปลอดภัย
- **Rollback Procedure**: คืนค่า `/etc/bcai-account/release.env` จาก `release.env.before` แล้วสั่ง `docker compose up -d --no-deps mainapi worker frontend`
