# สรุปการทำงานของ Alert Agent Screen
AlertAgentScreen เป็นหน้าจอ Flutter สำหรับจัดการระบบแจ้งเตือนอัตโนมัติ (Alert Agent) ในแอปพลิเคชัน SML AI Cloud หน้าจอนี้ช่วยให้ผู้ใช้สามารถสร้าง แก้ไข ลบ และจัดการการแจ้งเตือนที่ส่งรายงานผ่านช่องทางต่างๆ เช่น Email และ LINE Notify

## Components หลัก
# Left Panel (_buildLeftPanel())
แสดงรายการ Alert ทั้งหมด
ปุ่มเพิ่ม Alert ใหม่
สถานะ ON/OFF แสดงเป็น badge

# Right Panel (_buildRightPanel())
Mode ฟอร์มแก้ไข: แสดง AlertAgentEditForm
Mode รายละเอียด: แสดง _buildDetailView()
Mode ว่าง: แสดงข้อความแนะนำ
Alert List Item (_buildAlertListItem())

ชื่องาน (taskName)
ประเภทรายงาน
สถานะ ON/OFF
ไอคอนสถานะ

# การตรวจสอบความถูกต้อง
_hasValidChannel(): ตรวจสอบว่ามีช่องทางแจ้งเตือนที่ถูกต้องอย่างน้อย 1 ช่องทาง
Validation ก่อนบันทึก: ต้องมีช่องทางแจ้งเตือนที่กำหนดค่าแล้ว

## โหมดการส่ง (Schedule Modes)
# Interval Mode: ส่งซ้ำทุก X นาที
ตัวเลือก: 1, 5, 15, 30, 60 นาที
เหมาะสำหรับติดตามข้อมูลแบบ real-time

# Weekly Mode: กำหนดวันในสัปดาห์และเวลา
เลือกวัน (อา.-ส.) และเวลา
เหมาะสำหรับรายงานประจำวัน/สัปดาห์

# Monthly Mode: กำหนดวันที่ในเดือนและเวลา
เลือกวันที่ (1-31) หรือ "วันสิ้นเดือน"
เหมาะสำหรับรายงานสิ้นเดือน/ประจำเดือน

# การจัดการช่องทางแจ้งเตือน
แต่ละช่องทางมี Checkbox สำหรับเปิด/ปิดใช้งาน
เพิ่มรายการผ่าน TextField + ปุ่ม Add
แสดงรายการเป็น Chip ที่สามารถลบได้
Validation: ต้องมีอย่างน้อย 1 ช่องทางที่เปิดใช้งานและมีข้อมูล

# Endpoints ที่ใช้
POST /atlas/get - ดึงรายการ Alert
POST /atlas/update - บันทึก/อัปเดต Alert
POST /atlas/delete - ลบ Alert

# Request Structure

{
  "collection": "email",
  "shopid": "shop_id",
  "email": "task_name",
  "cartid": "task_name",
  "data": { /* AlertAgentModel */ },
  "upsert": true // สำหรับการสร้างใหม่
}

# # สรุปการทำงานของ Alert Agent Screen

## ภาพรวม
`AlertAgentScreen` เป็นหน้าจอ Flutter สำหรับจัดการระบบแจ้งเตือนอัตโนมัติ (Alert Agent) ในแอปพลิเคชัน SML AI Cloud หน้าจอนี้ช่วยให้ผู้ใช้สามารถสร้าง แก้ไข ลบ และจัดการการแจ้งเตือนที่ส่งรายงานผ่านช่องทางต่างๆ เช่น Email และ LINE Notify

### 3. การตรวจสอบความถูกต้อง
- `_hasValidChannel()`: ตรวจสอบว่ามีช่องทางแจ้งเตือนที่ถูกต้องอย่างน้อย 1 ช่องทาง
- Validation ก่อนบันทึก: ต้องมีช่องทางแจ้งเตือนที่กำหนดค่าแล้ว

### 4. การแสดงผลข้อมูล
- `_getReportTypeLabel()`: แปลงประเภทรายงานเป็นภาษาไทย
- `_formatScheduleMode()`: จัดรูปแบบการแสดงผลตารางเวลา
- `_formatDateTime()`: จัดรูปแบบวันที่/เวลา

### โหมดการส่ง (Schedule Modes)
1. **Interval Mode**: ส่งซ้ำทุก X นาที
   - ตัวเลือก: 1, 5, 15, 30, 60 นาที
   - เหมาะสำหรับติดตามข้อมูลแบบ real-time

2. **Weekly Mode**: กำหนดวันในสัปดาห์และเวลา
   - เลือกวัน (อา.-ส.) และเวลา
   - เหมาะสำหรับรายงานประจำวัน/สัปดาห์

3. **Monthly Mode**: กำหนดวันที่ในเดือนและเวลา
   - เลือกวันที่ (1-31) หรือ "วันสิ้นเดือน"
   - เหมาะสำหรับรายงานสิ้นเดือน/ประจำเดือน

## API Integration
### Endpoints ที่ใช้
1. `POST /atlas/get` - ดึงรายการ Alert
2. `POST /atlas/update` - บันทึก/อัปเดต Alert
3. `POST /atlas/delete` - ลบ Alert

### Request Structure
```json
{
  "collection": "email",
  "shopid": "shop_id",
  "email": "task_name",
  "cartid": "task_name",
  "data": { /* AlertAgentModel */ },
  "upsert": true // สำหรับการสร้างใหม่
}
```

## สรุป
`AlertAgentScreen` เป็นหน้าจอที่ออกแบบมาอย่างดีสำหรับจัดการระบบแจ้งเตือนอัตโนมัติ มีฟังก์ชันการทำงานครบถ้วนทั้งการสร้าง แก้ไข ลบ และจัดการการแจ้งเตือนผ่านช่องทางต่างๆ การออกแบบใช้หลักการ separation of concerns โดยแยกฟอร์มแก้ไขเป็นคลาสย่อย (`AlertAgentEditForm`) และมี validation ที่ครอบคลุมทั้งด้านข้อมูลและช่องทางการส่ง
