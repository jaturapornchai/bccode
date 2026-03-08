# Base Models

Model พื้นฐานที่ใช้ร่วมกันทั้ง project — embed ใน struct อื่นๆ

## Identity Models

### PartitionIdentity
ใช้สำหรับ MongoDB partitioning (ไม่แสดงใน JSON)

| Field | Go Type | BSON | JSON | หมายเหตุ |
|-------|---------|------|------|----------|
| ParID | string | `parid` | `-` | Partition ID (ซ่อนจาก API) |

**Backend:** `internal/models/identity.go`

### Identity (ShopID + GuidFixed)
ใช้เป็น primary key สำหรับ MongoDB documents

| Field | Go Type | BSON | JSON | หมายเหตุ |
|-------|---------|------|------|----------|
| ShopID | string | `shopid` | `shopid` | รหัสร้าน (เป็น database name ด้วย) |
| GuidFixed | string | `guidfixed` | `guidfixed` | UUID ของเอกสาร (primary key) |

### ShopIdentity
ใช้เมื่อต้องการแค่ ShopID

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| ShopID | string | `shopid` | `shopid` |

### DocIdentity
ใช้เมื่อต้องการแค่ GuidFixed

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| GuidFixed | string | `guidfixed` | `guidfixed` |

## Name Models

### NameX (Multi-language name)
ใช้สำหรับชื่อที่รองรับหลายภาษา — **model สำคัญที่สุดที่ใช้ทั่วทั้ง system**

| Field | Go Type | BSON | JSON | หมายเหตุ |
|-------|---------|------|------|----------|
| Code | *string | `code` | `code` | รหัสภาษา: `th`, `en`, `vi`, `lo`, `km`, `my`, `cn`, `ja`, `ko` |
| Name | *string | `name` | `name` | ชื่อในภาษานั้น |
| IsAuto | bool | `isauto` | `isauto` | สร้างอัตโนมัติหรือไม่ |
| IsDelete | bool | — | `isdelete` | ลบแล้วหรือไม่ |

**Backend:** `internal/models/name.go` → `NameX`
**Frontend:** `lib/model/global_model.dart` → `LanguageDataModel`

**ตัวอย่าง JSON:**
```json
[
  { "code": "th", "name": "กาแฟดำ", "isauto": false },
  { "code": "en", "name": "Black Coffee", "isauto": false }
]
```

### Name (Legacy — 5 ภาษาคงที่)
Model เก่า ใช้ 5 ช่อง name คงที่ (บาง collection ยังใช้อยู่)

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| Name1 | string | `name1` | `name1` |
| Name2 | *string | `name2` | `name2` |
| Name3 | *string | `name3` | `name3` |
| Name4 | *string | `name4` | `name4` |
| Name5 | *string | `name5` | `name5` |

### UnitName (Legacy — 5 ภาษาคงที่)

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| UnitName1 | string | `unitname1` | `unitname1` |
| UnitName2-5 | *string | `unitname2-5` | `unitname2-5` |

### Description (Legacy)

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| Description1-5 | *string | `description1-5` | `description1-5` |

## Activity Models

### Activity (Audit Trail)
ใช้ติดตามว่าใครสร้าง/แก้ไข/ลบ document

| Field | Go Type | BSON | JSON | หมายเหตุ |
|-------|---------|------|------|----------|
| CreatedBy | string | `createdby` | `createdby` | ผู้สร้าง |
| CreatedAt | time.Time | `createdat` | `createdat` | วันเวลาสร้าง |
| UpdatedBy | string | `updatedby` | `updatedby` | ผู้แก้ไขล่าสุด |
| UpdatedAt | time.Time | `updatedat` | `updatedat` | วันเวลาแก้ไข |
| DeletedBy | string | `deletedby` | `deletedby` | ผู้ลบ |
| DeletedAt | time.Time | `deletedat` | `deletedat` | วันเวลาลบ (zero = ยังไม่ลบ) |

### ActivityDoc (ซ่อนจาก JSON)
เหมือน Activity แต่ JSON tag = `"-"` (ไม่ส่งออก API)

### ActivityTime (ไม่มี user info)
เก็บแค่เวลา ไม่มีข้อมูลผู้ใช้

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| CreatedAt | time.Time | `createdat` | `createdat` |
| UpdatedAt | time.Time | `updatedat` | `updatedat` |
| DeletedAt | time.Time | `deletedat` | `deletedat` |

### LastUpdate

| Field | Go Type | BSON | JSON |
|-------|---------|------|------|
| LastUpdatedAt | time.Time | `lastupdatedat` | `lastupdatedat` |

## API Response Model

### ApiResponse
Response wrapper มาตรฐานของ API ทุกตัว

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| Success | bool | `success` | สำเร็จหรือไม่ |
| Message | string | `message` | ข้อความ |
| DocNo | string | `docno` | เลขที่เอกสาร (ถ้ามี) |
| ID | interface{} | `id` | ID ที่สร้าง/แก้ไข |
| Data | interface{} | `data` | ข้อมูล response |
| Pagination | interface{} | `pagination` | ข้อมูลหน้า (ถ้ามี) |
| Total | interface{} | `total` | จำนวนทั้งหมด |

**Frontend:** `ApiResponse<T>` ใน `lib/repositories/client.dart`
