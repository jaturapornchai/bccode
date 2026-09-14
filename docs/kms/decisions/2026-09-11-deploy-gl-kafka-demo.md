---
date: 2026-09-11
status: deployed
tags: [bc-account, deployment, general-ledger, kafka]
---

# Deploy บัญชีแยกประเภทผ่าน Kafka

## ผลและขอบเขต

ปล่อย `r20260911-gl-kafka-1` ไป [account.bcaicloud.com](https://account.bcaicloud.com/) วันที่ 11 กันยายน 2026 เวลา **18:11:35 น. ไทย** (`2026-09-11T11:11:35Z`) ตามคำสั่งลุงจืดให้ MongoDB เก็บข้อมูลต้นฉบับทั้งหมด และใช้ Kafka ส่งต่อไป PostgreSQL สำหรับประมวลผล

Main API, worker และ frontend healthy; หน้าเว็บและ API health ตอบ 200 ส่วน GL API ที่ไม่ส่ง session ตอบ 401 ตามการป้องกันสิทธิ์ ผลนี้ยืนยันการปล่อยบริการและการป้องกัน endpoint ยังไม่ใช่ผล CRUD ของข้อมูลตัวอย่างบน production ([deploy summary](../../evidence/2026-09-11-gl-kafka/deploy-summary.json))

**ข้อมูลตัวอย่างธุรกิจวัสดุก่อสร้าง: ผ่าน seed การตรวจฐานข้อมูลอิสระ และ final UI แล้ว** สร้างผ่าน business API เสร็จเวลา 18:18:15 น. ไทย และตรวจ MongoDB → Kafka → PostgreSQL แบบอ่านอย่างเดียวผ่านเวลา 18:25:59 น. ไทย ([seed proof](../../evidence/2026-09-11-gl-kafka/seed-verification.json), [production reconciliation](../../evidence/2026-09-11-gl-kafka/production-reconciliation.json)) การตรวจหน้าจอพบปัญหาการแสดงผลและปล่อย frontend patch `r20260911-gl-demo-thai-1` แล้วเวลา 18:29:44 น. ไทย และ final UI ผ่านพร้อม smoke ครบ 35 routes ([ผลตรวจ UI](../../evidence/2026-09-11-gl-kafka/final-ui-summary.json))

## ข้อมูลตัวอย่างและผลกระทบยอด

- Scope `demo / C01 / 00000`; 93 records ประกอบด้วยผังบัญชี 38 บัญชี, ปีบัญชี 1, กลุ่มบัญชี 5, กลุ่มบัญชีสินค้า 3, งวด 12, งบประมาณ 11, ประมาณการเงินสด 5 และสมุดรายวัน 18 ใบ (ผ่านรายการ 17 / ร่าง 1)
- 110 command events; Mongo/PG มีลำดับ 1–110 ไม่ซ้ำ, pending 0; เล่นซ้ำ 2 คำสั่งแล้วจำนวนข้อมูลไม่เปลี่ยน และจำนวน scope อื่นคงเดิม
- PG มี 77 ledger lines; เดบิตและเครดิตรวมตรงกัน `1953100.00000000` กระทบกับยอด exact รายบัญชีจาก Mongo แล้วไม่พบเอกสารไม่สมดุลหรือจำนวนเงินผิดชนิด
- รายงาน 5 แบบตรงค่าที่กำหนดไว้ใน fixture: งบทดลอง, งบกำไรขาดทุน, งบแสดงฐานะการเงิน, งบกระแสเงินสด และประมาณการกระแสเงินสด ผลอยู่ใน `seed-verification.json` ไม่ใช่การรับรองยอดธุรกิจจริงหรือแบบยื่น DBD
- Kafka group `bcai-account-01-gl-v2`: partition 0 current=end=111, unique events ใน Mongo/PG=110 และไม่เกิดรายการบัญชีซ้ำ; อีก 5 partitions end=0, lag ทุก partition=0 ณ เวลาตรวจ สอดคล้องกับการรับข้อความซ้ำอย่างปลอดภัย
- หมายเหตุเครื่องมือตรวจ: รอบแรกอ่านค่า async mongosh ก่อน await จึงยังไม่ได้ค่าผลลัพธ์ แก้ให้ await/assign ก่อนอ่านค่าแล้วผลสุดท้ายผ่าน ปัญหานี้ไม่ใช่ runtime หรือข้อมูลบัญชีเสีย

## Frontend patch ล่าสุด — r20260911-gl-demo-thai-1

ปล่อยเฉพาะ frontend วันที่ 11 กันยายน 2026 เวลา **18:29:44 น. ไทย**; mainapi/worker คง `r20260911-gl-kafka-1` และเวลาเริ่มเดิม 18:11:19 น. ไทย ทุกบริการ healthy หน้าเว็บ/health 200 และ GL ที่ไม่มี session 401 ([frontend patch summary](../../evidence/2026-09-11-gl-kafka/frontend-patch-summary.json))

- Frontend ปัจจุบัน: `bcai-account-frontend:r20260911-gl-demo-thai-1`; image ID `sha256:5466cc156475bd4bef76e967f0f26410faff966de4895709a13a90cb1df578d9`
- Archive 295,362,048 bytes; SHA-256 `c85858d0b5e8616894c26840ff08b0aa0fbdf0bbc593d288b19fcb9528447e88` ตรวจ artifact ก่อนปล่อยผ่าน
- ข้อมูลตัวอย่างทำให้พบ enum และรหัสแถวคำนวณภายในบนรายงาน จึงแปล `accounttype/bookcode/status/direction/category` เฉพาะตอน render, เติมชื่อยอดกำไรขาดทุนที่ยังไม่ปิด/เงินสด/จำนวนบรรทัด และซ่อน `__current_earnings__` จากคอลัมน์รหัสบัญชี โดยคง payload/จำนวนเงิน/API/CSV เดิม (`frontend/src/app/gl/gl-reports.tsx:16`)
- เพิ่ม render tests 8 cases ตรวจภาษาไทย, unknown values, business codes และ decimal/CSV เดิม (`frontend/src/app/gl/gl-reports.test.ts:12`); frontend รวม 54 files / 401 tests, lint, TypeScript, CODE-MAP 45 files และ Docker production build ผ่าน
- สะท้อนแบบแผน/ข้อห้าม/เหตุผล/วิธีตรวจใน skill `docs/skills/ui-scale-polish/SKILL.md:649` แล้ว
- สำรองหลัง seed ก่อน patch เวลา 18:24:40 น. ไทย: GL 38 บัญชี / 1 ปี / 18 สมุดรายวัน / 110 events เก็บ Mongo/PG/config เฉพาะบน server release นี้ ยังไม่ใช่ restore drill ([backup หลัง seed](../../evidence/2026-09-11-gl-kafka/frontend-patch-backup-summary.json))
- Rollback patch นี้: คืน `release.env.before` แล้ว recreate **frontend เท่านั้น** คง API/worker/ข้อมูล/audit ไว้; final UI หลัง patch ผ่านตามหลักฐานด้านล่าง

## Kafka release และ artifact

| บริการ | Image | Image ID จาก Docker inspect |
|---|---|---|
| mainapi / worker | `bcai-account-mainapi:r20260911-gl-kafka-1` | `sha256:bda3098928dc1858afac194899815041d119de372222b7532984933d85ab4ecb` |
| frontend | `bcai-account-frontend:r20260911-gl-kafka-1` | `sha256:12dfb40e6052eb75ea9b0cd76735b06c28896dfc153578d9ac0d96f223a3a562` |

- Release directory: `/opt/bcai-account/releases/r20260911-gl-kafka-1`; ใช้ Dockerfile production เดิม (`backend/Dockerfile`, `frontend/Dockerfile`)
- Archive 596,661,760 bytes; SHA-256 `d7ab18a00e7d5457ed1825e6be4f14cc77d4f2c98abcfa84eb2a0ca33e15f912` ยืนยันตรงก่อนปล่อย
- Topic `bc-gl-projection-v1` สร้างแล้ว: 6 partitions, replication factor 1, `min.insync.replicas=1` ทุก partition มี leader/ISR 1 ณ จุดตรวจ
- การตรวจ artifact ครั้งแรกเปรียบเทียบ frontend config digest กับ OCI image index จึงไม่ผ่าน ผู้ดูแลตรวจ `docker image inspect` ทั้ง local/server ใหม่ได้ ID `12df…` ตรงกันพร้อม archive SHA เดิม แล้ว gate สุดท้ายผ่าน exit 0 ก่อน deploy; ไม่ใช่การข้ามการตรวจ image
- รายละเอียดเครื่องตรวจรับ: [artifact summary](../../evidence/2026-09-11-gl-kafka/artifact-summary.json)

## Workflow และ runtime

1. คำสั่งบัญชี commit ข้อมูลและ `gl_events` ใน Mongo transaction ก่อนส่งหัวคิวของแต่ละ holding/company; ส่งผิดพลาดคงข้อมูลและสถานะ pending ใน Mongo ไม่มีเส้นทางลัดไปเขียน PG (`backend/internal/generalledger/store.go:194`, `:223`)
2. Main API mode 2 เป็น outbox relay ใช้ synchronous Kafka writer, RequireAll, Hash partition key และ timeout 5 วินาที ส่งเฉพาะ reference 6 ฟิลด์ ไม่มี payload การเงิน; ไม่สร้าง topic อัตโนมัติ (`httpapi/http.go:46`, `kafkatransport/transport.go:54`, `:127`)
3. Worker mode 1 ตรวจรูปแบบ/key และโหลด immutable event จาก Mongo เพื่อยืนยัน hash จากนั้น PG Project/Rebuild → Mongo delivered → Kafka offset commit; ความล้มเหลวไม่ข้ามรายการและ retry reader กลุ่มเดิม (`backend/main.go:667`, `httpapi/http.go:74`, `store.go:252`, `kafkatransport/transport.go:209`)
4. รายงานรอ Mongo/PG สอดคล้องกัน หากยังไม่พร้อมตอบ 409 `GL_PROJECTION_PENDING`; frontend retry เฉพาะ GET ภายในเวลาที่กำหนด ไม่ส่งคำสั่ง POST ซ้ำ (`httpapi/http.go:293`, `frontend/src/lib/general-ledger-api.ts:17`)
5. Shutdown ปิด consumer ก่อน writer/PG; microservice cancel workers และรอเสร็จก่อน cleanup ทรัพยากรร่วม (`kafkatransport/transport.go:228`, `httpapi/http.go:82`, `backend/pkg/microservice/microservice.go:244`)

เส้นทาง source ในย่อหน้าที่ไม่ระบุ prefix อยู่ใต้ `backend/internal/generalledger/` ดู [architecture และคู่มือ 35 เมนู](../architecture/2026-09-11-general-ledger-v2.md) สำหรับสัญญาจำนวนเงินและสิทธิ์

## Config, dependencies และตัวอย่างตรวจ

- ต้องมี Mongo replica set สำหรับ transaction, PG database ต่อ holding พร้อมสิทธิ์สร้าง/อ่าน schema และ Kafka topic ที่เตรียมแล้ว รอบนี้ไม่ได้สร้าง database เพิ่ม
- `KAFKA_SERVER_URL` ใช้ broker host:port; `CONSUMER_GROUP_NAME` ต่อท้าย `-gl-v2` (ถ้าว่างใช้ `03-gl-v2`)
- รองรับ `KAFKA_SECURITY_PROTOCOL` ว่าง/PLAINTEXT หรือ SSL จาก local CA/cert/key และ TLS ขั้นต่ำ 1.2; SASL หรือ broker config ไม่ถูกต้องทำให้ GL ไม่เริ่ม transport ไม่มี fallback เขียน PG ตรง
- `ENABLE_KAFKA=false` ใช้กับ GL runtime นี้ไม่ได้; ตรวจค่าจริงผ่าน config source ก่อนเปลี่ยน (`backend/internal/config/config_mq.go:21`)
- ตัวอย่าง verification ที่ใช้จริง: `tools/verify.sh backend outbox projection` และ Linux `go test -json -tags=integration ./internal/generalledger/... -count=1 -timeout=240s` โดยใช้ Mongo/PG/Kafka ทดสอบแยกตาม [metrics](../../evidence/2026-09-11-gl-kafka/test-metrics.json)

## ผลตรวจ

| ชุดตรวจ | ผล |
|---|---|
| GL core + HTTP + Kafka integration | 24 top-level tests + 63 subtests; 0 fail / 0 skip |
| Real Kafka flow ภายในชุด GL | 2 top-level tests + 3 subtests; ตรวจ Mongo ก่อน ACK, ACK ไม่เขียน PG, commit ordering, restart/replay และ forged reference |
| จำนวนเงินจริงในฐานทดสอบ | เดบิต = เครดิต `0.30000003`; ผลต่าง `0.00000000`; 6 ledger lines / 8 immutable events หลัง replay |
| Release backend | PASS; compile 107 packages ที่มี tests, execute tests 85 packages |
| Release outbox / barcode / projection | PASS รวม 18 tests/subtests; 0 fail / 0 skip |
| Frontend lint / TypeScript / Vitest | PASS; 53 test files / 393 tests |
| Docker production build และ CODE-MAP | PASS ทั้ง backend/frontend; CODE-MAP 45 files ตรง source |

[Release gates](../../evidence/2026-09-11-gl-kafka/release-gates.json) และ [GL metrics](../../evidence/2026-09-11-gl-kafka/test-metrics.json) เป็นหลักฐานรอบ Kafka นี้ ผล UAT 35 เมนูของ `r20260911-gl-v2-1` เป็นหลักฐานรุ่นก่อน ห้ามนำมาแทนการตรวจ runtime Kafka รอบใหม่

## Backup และ rollback

สำรอง Mongo archive, PostgreSQL dump และ runtime config ก่อน deploy เมื่อ `2026-09-11T11:03:42Z` เก็บบนเซิร์ฟเวอร์ release เดียวกัน ไม่มีการนำข้อมูลฐานจริงหรือ secret ลง repo ดูขนาดและ SHA-256 ที่ [backup summary](../../evidence/2026-09-11-gl-kafka/backup-summary.json)

Backup รอบนี้เป็น safety copy บน server เดิม **ไม่ใช่ offsite backup และยังไม่มีผล restore drill สำหรับ archive ใหม่นี้** ผล restore ของ release ก่อนหน้านี้อ้างได้เฉพาะ backup รุ่นนั้น

Rollback ตามแผน deploy: คืน `release.env.before` แล้ว recreate เฉพาะ mainapi/worker/frontend โดยคงฐานข้อมูล, topic, offset และประวัติ audit ทั้งหมด รุ่นก่อนคือ `r20260911-gl-v2-1` ซึ่ง transport ต่างกัน จึงต้องตรวจ pending events และผลกระทบก่อนย้อน binary; ห้าม restore dump ทับรายการที่เกิดหลัง backup หรือแปลงจำนวนเงินกลับเป็น float

## ผลตรวจ UI production รอบสุดท้าย

ตรวจบน frontend `r20260911-gl-demo-thai-1` หลัง deploy เสร็จ รายงานจบเวลา 18:31:04 น. ไทย และ smoke เมนูจบ 18:32:20 น. ไทย ([final UI summary](../../evidence/2026-09-11-gl-kafka/final-ui-summary.json))

- รายงาน 5 แบบตรง expected totals ของ fixture; GL GET 23 ครั้งตอบ 200 ทั้งหมด, console errors 0, blocked writes 0
- ตรวจภาพรายงานจริงครบ 16 ภาพ light/dark โดยกดปุ่มสลับธีมจริง งบทดลองครอบคลุม 1600/1280/1024/768 และภาพปีบัญชีอีก 1 ภาพเสร็จเวลา 18:36:11 น. ไทย ไม่พบ issue ใหม่ ยืนยันป้ายประเภทบัญชีภาษาไทย, กำไรขาดทุนที่ยังไม่ปิด 52,000 บาท และรหัสแถวคำนวณแสดง “—” หลัง patch
- เปิดครบ 35 routes: 33 เมนูมีหน้าจอใช้งาน และ 2 เมนูแสดงสถานะรอข้อมูลตามข้อจำกัดเดิม; GL GET 82 ครั้งตอบ 200 ทั้งหมด, console errors 0, blocked writes 0
- ทั้งสองชุดเป็น read-only และออกจากระบบหลังตรวจแล้ว ไม่มีการเพิ่ม/แก้รายการบัญชีจากการตรวจ UI
- Repo เก็บเฉพาะ summary ที่คัดเลือก ไม่คัดลอก PNG จาก production หรือ auth/session

## ข้อจำกัด

- 15 packages ใน `backend/.ci/test-quarantine.txt` ตรวจ compile อย่างเดียว ผล release gate ไม่รับรอง business contract ของกลุ่มนี้
- ยังมี 2 เมนูรอข้อมูล: ประมวลผลเอกสารซื้อขายเดิมต้องมีบริษัทและยอด exact ที่ยืนยันได้; XBRL ต้องมีแม่แบบ DBD ตรงกิจการ/มาตรฐานและผ่านตัวตรวจรับ ยังไม่มีไฟล์พร้อมยื่น
- Production topic เป็น single broker / RF1; RequireAll ในรูปแบบนี้ไม่ใช่หลักฐานรองรับ broker เสียทั้งเครื่อง การทดสอบ local ยังไม่ครอบคลุม multi-broker failover, production TLS/ACL หรือ throughput
- ข้อมูลตัวอย่างผ่านฐานข้อมูลและ final UI ใน scope demo/C01/00000; ผลนี้ไม่รับรองข้อมูลจริงของ holding อื่นหรือ workflow/DBD ของ 2 เมนูที่ยังรอข้อมูล
