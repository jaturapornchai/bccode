---
date: 2026-09-11
status: deployed
tags: [bc-account, deployment, general-ledger, financial-statement-designer]
---

# Deploy ระบบออกแบบงบการเงิน (Financial Statement Designer) สู่ Production

## วัตถุประสงค์และผลการปล่อยระบบ

ปล่อยรีลีส `r20260911-gl-statement-1` สู่ Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) วันที่ 11 กันยายน 2026 เวลา **20:38:19 น. ไทย** (`2026-09-11T13:38:19Z`) ตามคำสั่งลุงจืด โดยครอบคลุมทั้ง Go Backend (`mainapi` / `worker`) และ Next.js Frontend

## สรุปรายการฟังก์ชันที่ปล่อยขึ้น Production

1. **เมนูใหม่ออกแบบงบการเงิน**:
   - เพิ่มเมนู `/gl/statement-designer` (`Financial Statement Designer`) ในกลุ่มรายงานการเงินและงบบัญชีของระบบ GL (จำนวนเมนูรวม 224 รายการ, GL 34 รายการ)
   - รองรับทุกภาษาใน `languages.tsv`
2. **ปรับแต่งแบบอักษรและรูปแบบงบอิสระ**:
   - รองรับฟอนต์ Sarabun, Prompt, Kanit, Noto Sans Thai, Inter และ Monospace พร้อม dynamic Google Fonts loader
   - ตัวปรับขนาดอักษร 13–18px, Line Height, และความเข้มของเส้นตาราง
3. **โครงสร้างแถวและการผูกบัญชี**:
   - แถวประเภท Header, Account, Total, Text, Blank พร้อมเลขแถวอ้างอิงสูตร (RowNo)
   - โมดอลค้นหาและเลือกผังบัญชี (Account Picker Dialog)
   - สไตล์เฉพาะแถว: หนา, เอียง, ชิดซ้าย/กลาง/ขวา, เยื้องบรรทัด 0–5 ระดับ, ขีดเส้นใต้เดี่ยว/คู่
   - แม่แบบตั้งต้น (Starter Templates): งบดุล DBD, งบกำไรขาดทุน, งบต้นทุนผลิต, งบกระแสเงินสด
4. **เครื่องคำนวณยอดจริงและสูตรสด (Live Calculation Engine)**:
   - ดึงยอดจริงจากงบทดลองตามปีบัญชีที่เลือก
   - ประเมินผลสูตรตัวแปรแถว (`R10 + R20`), ฟังก์ชันช่วงรวม `SUM(R10:R50)`
   - คำนวณด้วย Recursive Descent Parser บน Fixed-Point BigInt (สเกล $10^8$)
   - โหมด Live Preview พร้อมปุ่มพิมพ์งบการเงิน (Print CSS) และส่งออก CSV
5. **Backend CQRS & Master Storage**:
   - บันทึกใน MongoDB `gl_statement_templates` ผ่าน Event Sourcing
   - โปรเจกต์ลง PostgreSQL `gl_records`

## Production Artifacts & Release Details

| บริการ | Image | Image ID (Docker inspect) |
|---|---|---|
| mainapi / worker | `bcai-account-mainapi:r20260911-gl-statement-1` | `sha256:b7a4f64e5e10ca55dc1fcb793a6d64bfed5948694adb9d09d4f441f1128b9bdc` |
| frontend | `bcai-account-frontend:r20260911-gl-statement-1` | `sha256:4e5580cf10f0f94c4820f570a298d0bed5cea23d4459ff452819aa74d8798d88` |

- **Release Directory**: `/opt/bcai-account/releases/r20260911-gl-statement-1`
- **Images Archive**: `images.tar` (597,476,864 bytes), SHA-256: `3bdf65842135200b67efa37a2c4e0ae5f66f8230c6e35a8040293a89c90cfc76`
- **Preflight Backups**:
  - `mongo.archive.gz`: 69,784 bytes, SHA-256: `b16064bd52a3dc3f23b4653f9c6bf7fce478760426305f5184374e2eb9398f47`
  - `postgres-all.sql`: 716,213 bytes, SHA-256: `cb65dcf74b745166b91d71ce89b418b2abd99778171e6927587b864d36b45df4`
  - `runtime-config.tar.gz`: 7,264 bytes, SHA-256: `291414cedbb132186b62902bc749a5e3e76ac058d4496ed0e80745b7d69408aa`
  - `release.env.before`: บันทึกสถานะก่อน deploy ไว้อย่างปลอดภัย
- **Rollback Procedure**: คืนค่า `/etc/bcai-account/release.env` จาก `release.env.before` แล้วสั่ง `docker compose up -d --no-deps mainapi worker frontend`

## ผลการทดสอบหลัง Deploy (Post-deploy Verification)

1. **HTTP Status Checks**:
   - `https://account.bcaicloud.com/`: 200 OK
   - `https://account.bcaicloud.com/api/gl/accounts`: 401 Unauthorized (Auth Guard ทำงานถูกต้อง)
   - `http://127.0.0.1:8888/healthz`: 200 OK
2. **Live Smoke Tests**:
   - ล็อกอิน `POST /api/auth/demo-login` $\rightarrow$ 200 OK
   - เลือกขอบเขต `demo / C01` $\rightarrow$ 200 OK
   - ดึงผังบัญชี `GET /api/gl/accounts` $\rightarrow$ 200 OK (38 รายการ)
   - ดึงสมุดรายวัน `GET /api/gl/journals` $\rightarrow$ 200 OK (18 รายการ)
   - ดึงงบทดลอง `GET /api/gl/reports/trialbalance` $\rightarrow$ 200 OK (31 แถว)
   - ดึงรายการแม่แบบงบ `GET /api/gl/statement-templates` $\rightarrow$ 200 OK
3. **Live MongoDB & PostgreSQL CRUD Cycle**:
   - **Create**: บันทึกแม่แบบ `STMT-TEST-001` (font Sarabun 15px, 3 แถวพร้อมสูตร `R20`) $\rightarrow$ 200 OK (`id: 5778392577fcf1919d2d7fccff3a9191`, version 1, sequence 111)
   - **Read**: ค้นหาและดึงข้อมูลกลับมา $\rightarrow$ 200 OK ได้รับข้อมูลครบถ้วนทั้งแถวและฟอนต์
   - **Update**: แก้ไขชื่อและสไตล์เป็น Prompt 16px $\rightarrow$ 200 OK (version 2, sequence 112)
   - **Delete**: ลบรายการทดสอบ $\rightarrow$ 200 OK (version 3, sequence 113)
   - **Clean Verify**: ตรวจสอบซ้ำ $\rightarrow$ 0 รายการ ไม่เหลือขยะในฐานข้อมูลจริง
