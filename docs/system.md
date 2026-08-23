# ระบบ

## ฐานข้อมูล

- MongoDB เก็บข้อมูลธุรกิจและ metadata ตาม [ความถูกต้องของข้อมูลข้ามฐาน](#ความถูกต้องของข้อมูลข้ามฐาน)
- PostgreSQL รับข้อมูลจาก Kafka เพื่อเก็บ Projection สำหรับงานอ่านที่เน้นความเร็ว
- ClickHouse รับข้อมูลจาก Kafka เพื่อเก็บข้อมูลสำหรับการวิเคราะห์

## Flow

```
MongoDB transaction
├── Business data
└── Outbox -> Background publisher -> Kafka -> PostgreSQL
                                      -> ClickHouse
```

## ความถูกต้องของข้อมูลข้ามฐาน

- MongoDB เป็น Source of Truth ของข้อมูลธุรกรรม ส่วน PostgreSQL และ ClickHouse เป็น Projection ที่อาจล่าช้าได้
- Event ต้องบันทึกใน Outbox ภายในงาน Atomic เดียวกับข้อมูลธุรกิจ และ Background publisher เป็นผู้ส่งไป Kafka
- Kafka ส่งแบบ at-least-once ทุก Event ต้องมี idempotency key, aggregate identifier และ version
- Consumer ต้องรองรับ Event ซ้ำและมาผิดลำดับ พร้อม Retry, DLQ และ Reconciliation
- API ที่ต้องการข้อมูลล่าสุดห้ามใช้ Projection ที่ยังตาม MongoDB ไม่ทันโดยไม่แจ้งสถานะข้อมูล

## PostgreSQL

- แยก Database ตามรหัสถาวรของ Company ตาม [รหัสประจำตัว](organization.md#รหัสประจำตัว) เพื่อให้ PostgreSQL ทำงานได้เต็มประสิทธิภาพ ห้ามใช้รหัสธุรกิจเป็น Database routing key

## ClickHouse

- แยก Database ตามรหัสถาวรของ Holding ตาม [รหัสประจำตัว](organization.md#รหัสประจำตัว) ห้ามใช้รหัสธุรกิจเป็น Database routing key
- ตารางข้อมูลระดับ Company หรือ Branch ต้องมีมิติรหัสถาวรของระดับนั้น ส่วนรหัสธุรกิจใช้เพื่อแสดงผลหรือวิเคราะห์ได้ แต่ห้ามใช้แทนรหัสถาวร โดยโครงสร้าง Field, Type และ Index ให้ยึด [ลำดับความน่าเชื่อถือของเอกสาร](README.md#ลำดับความน่าเชื่อถือของเอกสาร)

## MinIO

- MinIO ใช้เก็บเนื้อไฟล์ รูปภาพ และวิดีโอ

## Redis

- Redis ใช้เป็น Cache และ Session Store ชั่วคราว การสูญหายของ Redis อาจทำให้ Session หมดอายุ แต่ต้องไม่ทำให้ข้อมูลธุรกิจหรือ Audit สูญหาย

## Frontend

- พัฒนาด้วย Next.js และทำงานบน Browser

## Backend

- พัฒนาด้วยภาษา Go และทำงานบน Docker หรือ Docker Desktop
