# หลักฐาน GL Kafka release — 2026-09-11

รวบรวมเฉพาะ metadata ที่ตรวจจาก artifact/log local ของงานนี้ ไม่เก็บ dump, session, token, secret, payload production หรือสำเนาเอกสารบัญชีจริง

- [final-ui-summary.json](final-ui-summary.json): read-only UI ผ่าน 5 reports / 35 routes, GL GET 23+82 ตอบ 200, 16 report shots + 1 FY metadata, console/write errors 0
- [frontend-patch-summary.json](frontend-patch-summary.json): frontend ปัจจุบัน r20260911-gl-demo-thai-1, 401 tests, artifact/deploy และ final UI ผ่าน
- [frontend-patch-backup-summary.json](frontend-patch-backup-summary.json): backup หลัง seed ก่อน frontend patch เก็บเฉพาะ metadata
- [deploy-summary.json](deploy-summary.json): เวลา deploy, image IDs, health และ HTTP status
- [artifact-summary.json](artifact-summary.json): image archive SHA/size, Kafka topic และผล gate สุดท้าย
- [backup-summary.json](backup-summary.json): backup size/SHA และข้อจำกัด same-server safety copy
- [seed-verification.json](seed-verification.json): seed ผ่าน business API 110 คำสั่ง, ตรวจ Mongo ทันที, PG-backed read, 5 reports และเล่นซ้ำไม่เพิ่มข้อมูล
- [production-reconciliation.json](production-reconciliation.json): ตรวจอิสระ Mongo/Kafka/PG ผ่าน, 93 records, 110 unique events, pending/lag 0 และยอด exact รายบัญชีของข้อมูลสาธิต
- [test-metrics.json](test-metrics.json): GL core/HTTP/real Kafka regression และยอด exact ของข้อมูลสังเคราะห์
- [release-gates.json](release-gates.json): release backend/outbox/projection, frontend, Docker build, CODE-MAP และ quarantine

ต้นทาง local: `tmp/gl-kafka-20260911/{deploy-output.log,artifacts-topic.log,preflight-backup-output.log,test-metrics.json,verify-summary.json,frontend-verification.log}` พร้อมไฟล์ `.exit` ของแต่ละ gate; log เต็มยังอยู่ใน tmp ไม่ใช่ backup หรือข้อมูล production ที่คัดลอกกลับมา

สถานะข้อมูลตัวอย่างธุรกิจวัสดุก่อสร้าง: **ผ่าน seed การตรวจฐานข้อมูลอิสระ และ final UI หลัง frontend patch แล้ว** ผล sample ใน seed/reconciliation แยกจาก unit/integration synthetic metrics ดู [บันทึก release](../../kms/decisions/2026-09-11-deploy-gl-kafka-demo.md)
