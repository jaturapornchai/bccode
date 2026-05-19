# BC Ai Account Data Model Reference

เอกสาร Data Model สำหรับ BC Ai Account Platform — ใช้ร่วมกันระหว่าง **Backend** (Go) และ **Frontend** (Flutter)

## สารบัญ

| ไฟล์ | เนื้อหา |
|------|--------|
| [architecture.md](architecture.md) | สถาปัตยกรรม Database (MongoDB / PostgreSQL / ClickHouse) |
| [base-models.md](base-models.md) | Base Models ที่ใช้ร่วมกัน (Identity, NameX, Activity) |
| [transactions.md](transactions.md) | Transaction Models (เอกสารซื้อ/ขาย/โอน/ปรับ) |
| [products.md](products.md) | Product Models (สินค้า, Barcode, หน่วย, BOM) |
| [parties.md](parties.md) | Party Models (ลูกหนี้, เจ้าหนี้, สมาชิก) |
| [organization.md](organization.md) | Organization Models (สาขา, แผนก, คลัง, POS) |
| [clickhouse.md](clickhouse.md) | ClickHouse Tables (OLAP/Analytics — 29 ตาราง) |
| [enums.md](enums.md) | Enum / Constants (TransFlag, VatType, PayType, etc.) |
| [frontend-mapping.md](frontend-mapping.md) | Frontend ↔ Backend Field Mapping |

## ข้อตกลง

### Multi-Language
ทุก field ที่เป็นชื่อใช้ `[]NameX` (backend Go) / `List<LanguageDataModel>` (frontend Dart):
```
{ "code": "th", "name": "ชื่อภาษาไทย" }
{ "code": "en", "name": "English Name" }
```
ภาษาที่รองรับ: `th`, `en`, `vi`, `lo`, `km`, `my`, `cn`, `ja`, `ko`

### Multi-Currency
เอกสารทุกประเภทรองรับ 2 สกุลเงิน:
- **Base Currency** (`currency`) — สกุลเงินหลักสำหรับลงบัญชี (ค่าเริ่มต้น THB)
- **Document Currency** (`doc_currency`) — สกุลเงินของเอกสาร (USD, JPY, EUR, ...)
- **Exchange Rate** (`exchangerate`) — 1 Document Currency = ? Base Currency
- Field ที่ลงท้าย `_doc` = ยอดในสกุลเงินเอกสาร

### Per-Shop Database
- MongoDB: 1 database ต่อ 1 shop (ตั้งชื่อตาม shopid)
- PostgreSQL: 1 database ต่อ 1 shop (ตั้งชื่อตาม shopid)
- ClickHouse: 1 database กลาง, partition by shopid

### Soft Delete
- MongoDB: ใช้ `deletedat` field (zero time = ยังไม่ลบ)
- ClickHouse: ใช้ `isdelete` Bool field
- Frontend: ใช้ `isdelete` Bool field

## Source Code Locations

### Backend (Go)
```
D:\bcdev\backend\
├── internal\models\              # Base models (Identity, NameX, Activity)
├── internal\transaction\models\  # Transaction structs
├── internal\product\             # Product structs
├── internal\debtaccount\         # Debtor/Creditor structs
├── internal\member\              # Member structs
├── internal\organization\        # Branch/Department structs
├── internal\warehouse\           # Warehouse structs
├── internal\goapi\models\        # GoAPI/BI models
└── internal\goapi\myclickhouse\  # ClickHouse table definitions
```

### Frontend (Flutter/Dart)
```
D:\bcdev\bcaiaccount\
├── lib\model\                    # Dart model classes (250+ files)
├── lib\services\                 # API service files
└── lib\repositories\             # Repository files (60+)
```
