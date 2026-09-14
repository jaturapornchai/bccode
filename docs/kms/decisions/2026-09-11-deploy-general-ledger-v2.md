---
date: 2026-09-11
status: deployed
tags: [bc-account, deployment, general-ledger, champ]
---

# Deploy ระบบบัญชีแยกประเภทใหม่

## วัตถุประสงค์และผลที่ปล่อย

ตามคำสั่งลุงจืด “deploy เลย” ปล่อยระบบบัญชีแยกประเภทที่อิงลำดับงาน Champ ไป [account.bcaicloud.com](https://account.bcaicloud.com/) วันที่ 11 กันยายน 2026 เวลา 17:25 น. ไทย

มีหน้าจอและ API 33 เมนู; ประมวลผลเอกสารซื้อขายเดิมยังรอต้นทางยอดที่ยืนยัน และ XBRL ยังรอแม่แบบ DBD ของบริษัท ทั้งสองเมนูคงสถานะรอพัฒนา รายละเอียด [ระบบและ Markdown 35 เมนู](../architecture/2026-09-11-general-ledger-v2.md)

## Release และขอบเขต

- Host `159.223.43.229`; compose project `bcai-account`
- Main API และ worker: `bcai-account-mainapi:r20260911-gl-v2-1`, image ID `sha256:519dde57b408115eec5160af7949f20e48f9d9b508f9a3d520d585f59e48ce91`
- Frontend: `bcai-account-frontend:r20260911-gl-v2-1`, image ID `sha256:9211304d9890dcc8cd5dd4e15a4b7e19ee6644d23c678ec68f65c7aceca41960`
- รุ่นเดิม mainapi/worker `r20260902-11`, frontend `r20260911-menu-top-1`; เก็บ image เดิมไว้
- Source base `a04bfb57` พร้อม working tree ที่ผ่านการทดสอบ ไม่ได้ commit หรือ push; source/public/config frontend 321 ไฟล์ตรงกับชุด UAT
- สร้าง image ด้วย Dockerfile production เดิม; backend URL ภายใน `http://mainapi:8888` และ public Google client ID เดิม ไม่เปลี่ยน secret, DNS หรือ Caddy
- Archive image ทั้งสองฝั่ง SHA-256 ตรงกัน: `65a98f63dc4e21aeecfd8d3c060f2448d138d9277b4357567d8b79714ed7b5b3`

## Workflow และการเตรียมฐานข้อมูล

1. ตรวจ MongoDB ฐานจริง `bcai_account`: replica set `rs0`, PRIMARY, สมาชิก `mongo:27017`; ผังบัญชี/ปีบัญชี/สมุดรายวัน/เหตุการณ์ GL ยังไม่มีข้อมูลเดิม จึงไม่มีการแปลงจำนวนเงินหรือเขียนทับข้อมูลเก่า
2. ตรวจ PostgreSQL และสิทธิ์แอป พบ holding `demo` มีฐานพร้อมแล้ว แต่ holding `test` ยังไม่มีฐาน จึงสำรองก่อนแล้วสร้างเฉพาะฐาน `test` เจ้าของ `bcai` และตรวจ CONNECT/USAGE/CREATE ผ่าน ไม่คัดลอกธุรกรรมหรือสร้างปีบัญชีแทนผู้ใช้
3. สำรอง MongoDB แบบ archive/gzip/oplog, PostgreSQL ด้วย pg_dumpall และ config ไว้เฉพาะเซิร์ฟเวอร์ใน `/opt/bcai-account/releases/r20260911-gl-v2-1/backup/` สิทธิ์ 600; สำรอง release.env เดิมแยกไว้
4. กู้ MongoDB ในคอนเทนเนอร์ใหม่ที่ไม่มี network/ports และไม่ใช้ volume production: สำเร็จ 462 เอกสาร, 0 failures; จำนวนตรงต้นทางที่ตรวจภายหลัง 23 จาก 25 collections ส่วน auth/access logs เพิ่มหลังสำรองและบันทึกความต่างไว้ ล้างคอนเทนเนอร์และ volumes ทดสอบแล้ว หลักฐานฉบับเต็มอยู่บนเซิร์ฟเวอร์ที่ `mongo-restore-proof/`
5. เปลี่ยนเฉพาะตัวแปร image ใน release.env แบบ atomic แล้ว compose `up -d --no-deps mainapi worker` ตรวจ health/API ก่อนสลับ frontend; script บนเซิร์ฟเวอร์ `deploy-release.py` ย้อน image เดิมอัตโนมัติถ้า health หรือ HTTP ตรวจไม่ผ่าน

สำเนาสำรองรอบนี้อยู่บนเซิร์ฟเวอร์เดียวกัน ไม่ใช่ offsite backup; ทดสอบ restore จริงเฉพาะ MongoDB ไม่ได้อ้างว่าทดสอบการกู้ทุกระบบครบ

## ผลตรวจ

- Frontend: Vitest 53 ไฟล์ / 384 tests ผ่าน, TypeScript ผ่าน, lint 0 errors / 219 warnings เดิม, production Docker build ผ่าน
- Backend/outbox/projection ตาม `tools/verify.sh` ผ่าน; GL Linux core + HTTP ผ่าน 17 top-level tests และ 38 subtests ไม่ข้ามทดสอบ
- Projection สองรอบแรกติด Kafka NotLeaderForPartition ในขั้นเตรียมข้อมูลทันทีหลังสร้าง topic แก้เฉพาะเทสต์ให้โหลด metadata ใหม่และรอ leader ไม่เกิน 10 วินาที รับ retry เฉพาะ leadership errors; เงื่อนไขตรวจ consumer เดิมไม่เปลี่ยน รอบหลังผ่าน 5 tests + 4 subtests (`backend/internal/product/projection/consumer_kafka_integration_test.go:70`)
- เทสต์เมนูเดิมปรับ Account Mapping ให้คาดหวังสถานะพร้อมใช้ตรงระบบใหม่ การแก้เทสต์ทั้งสองจุดไม่เปลี่ยน runtime image
- ใช้ frontend checks ผ่าน PowerShell และ Docker build ประกอบกับ backend/outbox/projection runner แทนการอ้างว่า wrapper `npm run verify:all` บน Git Bash ผ่าน; quarantine เดิม 15 backend packages ยัง compile-only ตาม runner
- หลังปล่อย mainapi/worker/frontend healthy, restart 0, ไม่มี OOM; HTTPS หน้าแรก 200, main API health 200, GL API/BFF ที่ไม่ login ตอบ 401 ตามสิทธิ์
- Browser smoke บน HTTPS จริงผ่าน: เปิดครบ 35 เมนู / 33 พร้อมใช้ / 2 รอข้อมูล; GL GET 82 ครั้งตอบ 200 ทั้งหมด, console errors 0, ไม่มีคำขอเขียนธุรกิจ และออกจากระบบสำเร็จ ตัวทดสอบอนุญาตเฉพาะ auth/session กับ backend health probe; ไม่สร้างธุรกรรมบัญชี

หลักฐานที่ตรวจเนื้อหาแล้ว: [ผล deploy](../../evidence/2026-09-11-general-ledger-release/deploy-result.json), [frontend](../../evidence/2026-09-11-general-ledger-release/frontend-verification.json), [backend](../../evidence/2026-09-11-general-ledger-release/backend-verification-notes.md), [ตัวเลข tests](../../evidence/2026-09-11-general-ledger-release/test-metrics.json), [preflight/backup metadata](../../evidence/2026-09-11-general-ledger-release/preflight-backup.json)

## การย้อนกลับ

ต้องตรวจว่าไม่มี release ใหม่กว่าก่อนคืน `/etc/bcai-account/release.env` จาก `/opt/bcai-account/releases/r20260911-gl-v2-1/release.env.before` แล้วใช้ compose เดิม `up -d --no-deps mainapi worker frontend` และตรวจ health/HTTPS ซ้ำ เก็บฐาน `test`, GL collections/tables และ audit history ไว้ ห้ามลบหรือคืนฐานทับรายการที่อาจถูกสร้างหลัง deploy

## ข้อจำกัด

ไม่ย้ายเอกสารเดิมและไม่สร้างรายการตัวอย่างใน production ผู้ใช้ต้องกำหนดผังบัญชี ปีบัญชีและงวดก่อนบันทึกงานจริง การสำรอง/restore JSON ในเมนูและรูปแบบงบยังมีข้อจำกัดตามคู่มือ GL; สองเมนูที่รอข้อมูลยังไม่เปิดดำเนินการจริง

ผล browser ที่ตัดเหลือเฉพาะสถานะและจำนวน: [production smoke summary](../../evidence/2026-09-11-general-ledger-release/production-smoke-summary.json)
