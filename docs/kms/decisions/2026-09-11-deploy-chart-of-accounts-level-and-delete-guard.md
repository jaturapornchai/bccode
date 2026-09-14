---
date: 2026-09-11
status: deployed
tags: [bc-account, deployment, general-ledger, chart-of-accounts]
---

# Deploy ระบบผังบัญชีรองรับระดับ 1–12 และการป้องกันการลบเมื่อมีข้อมูลอ้างอิงจากสมุดรายวันเด็ดขาด

## วัตถุประสงค์และผลการปล่อยระบบ

ปล่อยรีลีส `r20260911-gl-level-1` สู่ Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) วันที่ 11 กันยายน 2026 เวลา **19:47:50 น. ไทย** (`2026-09-11T12:47:50Z`) ตามคำสั่งลุงจืด โดยครอบคลุมทั้ง Go Backend (`mainapi` / `worker`) และ Next.js Frontend

## สรุปรายการเปลี่ยนแปลงที่ปล่อยขึ้น Production

1. **ผังบัญชีรองรับระดับบัญชี (Level 1–12)**:
   - บันทึก `level` ลงในเอกสาร MongoDB `chart_of_accounts` และโปรเจกต์ลง PostgreSQL `gl_records`
   - แนะนำระดับบัญชีลูกตามบัญชีแม่ให้อัตโนมัติ (`parent.level + 1`)
   - แสดงผลในตารางผังบัญชีด้วย Badge ระดับ + กิ่งไม้เยื้องลำดับชั้น (`└─`) สำหรับ Level > 1
   - แสดงการเยื้องลำดับชั้นใน Dropdown บัญชี (`AccountSelect`) ทุกจุดของระบบ
2. **การป้องกันการลบผังบัญชีเด็ดขาด (Zero Deletion on Referenced Accounts)**:
   - ตัดเงื่อนไข `isdeleted: false` ออกจากการค้นหารายการอ้างอิงใน `gl_journals` เพื่อป้องกันการลบผังบัญชีที่มีประวัติธุรกรรม (รวมถึงรายการที่ถูก Void หรือยกเลิกแล้ว เพื่อรักษา Audit Trail)
   - ปฏิเสธการลบด้วยข้อความภาษาไทย: `"บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ"`
   - แสดงป้ายแจ้งเตือนล่วงหน้าในแบบฟอร์มผังบัญชี และเพิ่มรายละเอียดในกล่องยืนยันการลบ

## Production Artifacts & Release Details

| บริการ | Image | Image ID (Docker inspect) |
|---|---|---|
| mainapi / worker | `bcai-account-mainapi:r20260911-gl-level-1` | `sha256:52c3cb94784282faaa69a2b2edb46b8e62b54092ec8ee7599704d333b4820e85` |
| frontend | `bcai-account-frontend:r20260911-gl-level-1` | `sha256:5968a4e7230a6407bb3d6fd7b057c93b39839da2dd6c2d08a2582361847e9260` |

- **Release Directory**: `/opt/bcai-account/releases/r20260911-gl-level-1`
- **Images Archive**: `images.tar` (596,744,704 bytes), SHA-256: `42bf26725362c504b86f99a88404b0bbc86c6b4c9ffd3895e7a886299a33965b`
- **Preflight Backups**:
  - `mongo.archive.gz`: 70,171 bytes, SHA-256: `0ffef84712b6964c4b01c4618a8518a06cd40ea7841934942ee85128e6b340db`
  - `postgres-all.sql`: 716,213 bytes, SHA-256: `ba9892533760f1716bf3453831e88107ac19c1af36cee50ad451217a12ae2aa1`
  - `runtime-config.tar.gz`: 7,269 bytes, SHA-256: `4c412551a78b89b904ef83aaadd60e43e2577b406eb78b134566de8dfa9b24bc`
  - `release.env.before`: บันทึกสถานะก่อน deploy ไว้อย่างปลอดภัย
- **Rollback Procedure**: คืนค่า `/etc/bcai-account/release.env` จาก `release.env.before` แล้วสั่ง `docker compose up -d --no-deps mainapi worker frontend`

## ผลการทดสอบหลัง Deploy (Post-deploy Verification)

1. **HTTP Status Checks**:
   - `https://account.bcaicloud.com/`: 200 OK
   - `https://account.bcaicloud.com/api/gl/accounts`: 401 Unauthorized (API Gateway / Auth Guard ทำงานถูกต้อง)
   - `http://127.0.0.1:8888/healthz`: 200 OK
2. **Live Smoke Tests (Demo Session)**:
   - ทดสอบล็อกอินเข้าสู่ระบบ `POST /api/auth/demo-login` $\rightarrow$ ได้รับ Bearer Token สำเร็จ
   - เลือกขอบเขต `demo / C01` สำเร็จ
   - ดึงข้อมูลผังบัญชี `GET /api/gl/accounts` $\rightarrow$ ตอบ 200 OK ครบ 38 รายการ
   - ดึงข้อมูลสมุดรายวัน `GET /api/gl/journals` $\rightarrow$ ตอบ 200 OK ครบ 18 ใบ
   - เรียกดูงบทดลอง `GET /api/gl/reports/trialbalance` $\rightarrow$ ตอบ 200 OK ครบ 31 แถว ยอดดุลตรง
