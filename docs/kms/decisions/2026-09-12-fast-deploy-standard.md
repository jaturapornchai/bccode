---
date: 2026-09-12
status: implemented
tags: [bc-account, deployment, fast-deploy, streaming, devops, zero-disk]
---

# มาตรฐานการ Deploy แบบเร็วที่สุด (Fast Streamed Zero-Disk Deploy) และกฎเสร็จแล้ว Deploy ทันที

## บริบทและโจทย์ความต้องการ (Context & Requirement)

ลุงจืดได้ตั้งกฎใหม่:
> **"ตั้งกฏใหม่ ต่อไปเสร็จแล้ว deploy ได้เลย หาวิธี deploy ที่เร็วที่สุดด้วย"**

เดิมทีการ deploy มีคอขวด 4 ประการ:
1. ต้องหยุดรอถามความเห็นลุงจืดก่อน deploy ทำให้รอบการส่งมอบช้าลง
2. กระบวนการเดิมบันทึกไฟล์อิมเมจ `images.tar` ขนาดใหญ่ (600MB – 1.2GB) ลงบน SSD ของเครื่อง Windows จากนั้นอัปโหลดผ่าน `scp` แล้วจึงสั่ง `docker load` บนเซิร์ฟเวอร์ กินเวลา I/O และ Bandwidth รวม 2–3 นาที
3. การ rebuild และบันทึก `mainapi` ทุกครั้งแม้จะไม่มีการแก้ไขโค้ด backend (เสียเวลาและเน็ตเพิ่มขึ้น 300MB โดยไม่จำเป็น)
4. การรีสตาร์ต service ทั้งหมดทำให้เสียเวลาเช็ก health ของ container ที่ไม่ได้เปลี่ยนแปลง

## สถาปัตยกรรมและกระบวนการ Deploy ที่เร็วที่สุด (Fast Streamed Zero-Disk Architecture)

เครื่องมือกลาง: `tools/fast-deploy.py`

### 1. กฎการทำงานอัตโนมัติ (Auto-Deploy on Done)
- เพิ่มกฎบังคับใน `AGENTS.md`: เมื่อทำงานใดๆ ผ่านการ Verify (Unit tests / TypeScript / Build) ครบ 100% แล้ว ให้ดำเนินการ Deploy ขึ้น Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) ได้เลยทันทีโดยไม่ต้องหยุดถาม

### 2. Frontend-Only Fast Path (Zero-Rebuild Backend)
- สคริปต์ตรวจจับการเปลี่ยนแปลงของโค้ด Go ใน `backend/` อัตโนมัติ
- หากมีเฉพาะการเปลี่ยนแปลงใน `frontend/`:
  - ข้ามการ build `mainapi` บนเครื่อง local
  - ข้ามการเซฟและอัปโหลด image `mainapi` (ประหยัดแบนด์วิดท์ไปทันที ~300MB)
  - ทำการ Re-tag อิมเมจ `mainapi` เดิมบนเซิร์ฟเวอร์โดยตรงผ่าน `docker tag <prev_mainapi> <new_mainapi>` (ใช้เวลา 0 วินาที)
  - สั่ง restart เฉพาะ container `frontend` เท่านั้น

### 3. In-Memory Streaming over Compressed SSH Pipe (Zero-Disk)
- ไม่เขียนไฟล์ `images.tar` ลงฮาร์ดดิสก์ของเครื่อง local และเซิร์ฟเวอร์
- สตรีมไบนารีจาก `docker save` ส่งตรงเข้า `stdin` ของ `ssh -C root@159.223.43.229 "docker load"`
- อาศัยการบีบอัดข้อมูลแบบเรียลไทม์ผ่าน SSH Transport Layer (`-C`)
- ถ่ายโอนอิมเมจ Next.js ขนาด 1.2GB เข้าสู่เซิร์ฟเวอร์ได้เสร็จสิ้นภายใน **~35-40 วินาที** (เร็วกว่าเดิมกว่า 3 เท่า)

### 4. ความปลอดภัยระดับ Production (Zero Data Loss Guarantee)
- ยังคงรัน **Preflight Backups** เสมอ:
  - สำรอง MongoDB ด้วย `mongodump --archive --gzip --oplog`
  - สำรอง PostgreSQL ด้วย `pg_dumpall`
  - สำรอง Runtime Config `/etc/bcai-account`
  - บันทึก `/etc/bcai-account/release.env.before`
- สลับเวอร์ชันใน `/etc/bcai-account/release.env` แบบ Atomic (`os.replace`)
- ตรวจสอบความพร้อมของ Container (`State.Health.Status == healthy`) ก่อนยืนยันผล
- รัน Smoke test และตรวจ HTTP 200/401 auth guard บน Public URL จริง
