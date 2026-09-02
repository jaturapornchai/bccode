# ระบบ

## ขอบเขตของไฟล์นี้

- ขอบเขตและลำดับความน่าเชื่อถือให้ยึด [ลำดับความน่าเชื่อถือของเอกสาร](README.md#ลำดับความน่าเชื่อถือของเอกสาร) ส่วนเนื้อหาในไฟล์นี้กำหนด Architectural Invariant, หน้าที่ของ Data Store, Consistency, Delivery Guarantee และข้อจำกัดด้าน Runtime
- คำว่า System of Record ในไฟล์นี้หมายถึงฐานข้อมูลหลักของข้อมูลที่ Runtime ใช้งาน ไม่ใช่ลำดับความน่าเชื่อถือของ Requirement หรือ Design

## ฐานข้อมูล

- MongoDB เป็น Runtime System of Record สำหรับ structured business data รวมถึง file metadata และ reference ตาม [ความถูกต้องของข้อมูลข้ามฐาน](#ความถูกต้องของข้อมูลข้ามฐาน)
- PostgreSQL รับข้อมูลจาก Kafka เพื่อเก็บ Projection สำหรับงานอ่านที่เน้นความเร็ว
- ClickHouse รับข้อมูลจาก Kafka เพื่อเก็บข้อมูลสำหรับการวิเคราะห์

## Flow สำหรับ Mutation ที่มี Event

```
MongoDB transaction
├── Business data
├── Audit ตามที่ Domain กำหนด
└── Outbox ตามที่ Domain กำหนด -> Background publisher -> Kafka -> PostgreSQL ตาม Event/Projection Contract
                                                          -> ClickHouse ตาม Event/Projection Contract
```

Projection target และ Consumer ของแต่ละ Event ยึดตาม Event/Projection Contract ของ Event ชนิดนั้น ภาพด้านบนไม่ได้หมายความว่าทุก Event ถูกส่งเข้าทั้ง PostgreSQL และ ClickHouse

## ความถูกต้องของข้อมูลข้ามฐาน

- MongoDB เป็น Runtime System of Record ส่วน PostgreSQL และ ClickHouse เป็น Projection ที่อาจล่าช้าได้และห้ามย้อนกลับมาเขียนทับข้อมูลธุรกิจใน MongoDB
- Mutation ที่ Domain กำหนดให้มี Audit และ/หรือ Event ต้องบันทึก Business data พร้อม Audit และ/หรือ Outbox ที่กฎนั้นกำหนดภายใน MongoDB transaction เดียวกัน ส่วนการส่ง Kafka, การอัปเดต Projection และการล้าง Redis อยู่นอก transaction และทำหลัง commit
- กฎ Outbox ครอบคลุม Domain Event หรือ Integration Event ที่เกิดจากการเปลี่ยน Business state ส่วน Event สำหรับ Session, Cache หรือ Telemetry ไม่รวม เว้นแต่เอกสารเจ้าของ Domain ระบุ
- Kafka ส่งแบบ at-least-once ทุก Event ต้องมี idempotency key, aggregate identifier และ version โดยขอบเขตความไม่ซ้ำ Partition key, Payload, Schema version, Evolution rule, Consumer และ Projection target ต้องอยู่ใน Event/Projection Contract ที่มี Source of Truth ชัดเจนก่อนสร้าง Event Event แต่ละชนิดส่งเฉพาะ Consumer และ Projection target ที่ Contract ระบุ ไม่ได้ส่งเข้าทั้ง PostgreSQL และ ClickHouse โดยอัตโนมัติ
- Consumer ต้องรองรับ Event ซ้ำและมาผิดลำดับ พร้อม Retry, DLQ และ Reconciliation โดย Reconciliation ต้องเทียบ Projection version หรือ checkpoint กับ MongoDB Runtime System of Record และห้ามถือ Projection เป็นข้อมูลธุรกิจหลัก
- API แบบ Current หรือ Strong consistency ต้องอ่าน MongoDB หรือรอจน Projection ถึง required version ห้ามคืนข้อมูลเก่าในฐานะข้อมูลล่าสุด ส่วน API แบบ Eventual consistency ใช้ Projection ได้ แต่ Contract ต้องระบุ `as-of`, version หรือ checkpoint ที่บอกความสดของข้อมูล

## Authorization และ Database routing

- Database routing ไม่ใช่ Authorization ก่อนเลือก Database หรือ Query ต้องจำแนก Endpoint และบังคับเฉพาะเงื่อนไขของกลุ่มนั้นตาม [ความปลอดภัยของ Session](login.md#ความปลอดภัยของ-session) และ [Membership](organization.md#membership) โดย Session, Membership, Role, Organization effective status, Scope กับ Permission ไม่ใช่เงื่อนไขร่วมของทุกกลุ่ม ห้ามเพิ่มเงื่อนไขที่เอกสารยกเว้นไว้
- ห้ามใช้ PostgreSQL หรือ ClickHouse Projection ที่อาจล่าช้าเพื่อตัดสินสิทธิ์
- Query ใน ClickHouse Database ระดับ Holding ต้องกรอง Company และ Branch ตาม Scope และ Permission ที่อนุญาต ห้ามถือว่าผู้ที่เห็น Holding เห็นข้อมูลธุรกรรมทุก Company โดยอัตโนมัติ

## PostgreSQL

- แยก Database ตามรหัสถาวรของ Company ตาม [รหัสประจำตัว](organization.md#รหัสประจำตัว) เพื่อให้ PostgreSQL ทำงานได้เต็มประสิทธิภาพ ห้ามใช้รหัสธุรกิจเป็น Database routing key
- การเลือก Database ระดับ Company ไม่ได้ให้สิทธิ์ทุก Branch Query หรือผลลัพธ์ระดับ Branch ต้องตรวจว่า Branch นั้นผ่านเงื่อนไขด้าน Scope จาก Company Scope ที่ครอบคลุมทุก Branch หรือจาก Branch Scope โดยตรง และผ่าน Domain Permission ตาม [Membership](organization.md#membership)

## ClickHouse

- แยก Database ตามรหัสถาวรของ Holding ตาม [รหัสประจำตัว](organization.md#รหัสประจำตัว) ห้ามใช้รหัสธุรกิจเป็น Database routing key
- ตารางข้อมูลระดับ Company หรือ Branch ต้องมีมิติรหัสถาวรของระดับนั้น ส่วนรหัสธุรกิจใช้เพื่อแสดงผลหรือวิเคราะห์ได้ แต่ห้ามใช้แทนรหัสถาวร ไฟล์นี้ไม่กำหนด Schema, Field, Type หรือ Index รายตาราง เจ้าของรายละเอียดเหล่านี้ให้ยึด [ลำดับความน่าเชื่อถือของเอกสาร](README.md#ลำดับความน่าเชื่อถือของเอกสาร) หากยังไม่มีให้หยุดและรายงาน

## MinIO

- MinIO เป็น Runtime System of Record สำหรับเนื้อไฟล์ รูปภาพ และวิดีโอ ส่วน MongoDB เก็บ metadata กับ reference เท่านั้น ห้ามเก็บ file bytes ซ้ำใน MongoDB หรือถือ metadata ว่าเป็นหลักฐานว่าเนื้อไฟล์ถูกเก็บสำเร็จ หาก Domain ยังไม่มี Workflow เรื่อง upload, commit, delete, compensation และ orphan cleanup ให้หยุดส่วนนั้นและรายงาน

## Redis

- Redis ใช้เป็น Cache และ Session Store ชั่วคราว โดยอำนาจของ Session และพฤติกรรมเมื่อหา Session ไม่พบให้ยึด [ความปลอดภัยของ Session](login.md#ความปลอดภัยของ-session)
- การสูญหายของ Redis ต้องไม่ทำให้ข้อมูลธุรกิจหรือ Audit สูญหาย ส่วนผลต่อ Session และ Permission cache ให้ยึด [ความปลอดภัยของ Session](login.md#ความปลอดภัยของ-session) และ [Membership](organization.md#membership)

## Frontend

- พัฒนาด้วย Next.js โดย UI และ Client Component ทำงานบน Browser ส่วน Server Component, Route Handler หรือ BFF ที่ใช้ต้องทำงานฝั่ง Server อย่างชัดเจน Secret, Credential และ Server control token ห้ามอยู่ใน Client bundle

## Backend

- พัฒนาด้วยภาษา Go และทำงานบน Docker หรือ Docker Desktop
