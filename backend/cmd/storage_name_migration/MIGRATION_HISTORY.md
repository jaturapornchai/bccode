# Storage Name Migration History

เอกสารนี้ใช้เก็บประวัติการเปลี่ยนชื่อ storage object สำหรับ migration ของ BC Ai Account

## 2026-05-20 - Standardize Storage Names

### Objective

ปรับชื่อ collection/table ของ MongoDB, PostgreSQL และ ClickHouse ให้เป็นตัวเล็กทั้งหมดและใช้ underscore ตามกฎกลางของระบบ

ตัวอย่าง:

```text
shopUserAccessLogs -> shop_user_access_logs
productbarcode -> product_barcodes
docdetail -> doc_detail
```

### Rule

- ชื่อ MongoDB collection ต้องเป็น `lowercase_no-underscore`
- ชื่อ PostgreSQL table ต้องเป็น `lowercase_no-underscore`
- ชื่อ ClickHouse table/dictionary ต้องเป็น `lowercase_no-underscore`
- ชื่อ MongoDB field ต้องเป็น `lowercase_no-underscore`
- ชื่อ PostgreSQL column ต้องเป็น `lowercase_no-underscore`
- ชื่อ ClickHouse column ต้องเป็น `lowercase_no-underscore`
- ห้ามสร้างชื่อใหม่แบบ camelCase, PascalCase, mixedCase, kebab-case, compact legacy name หรือมีช่องว่าง
- ตัวอย่าง field: `docDate` และ `docdate` ต้องเป็น `doc_date`
- ตัวอย่าง field: `custCode` และ `custcode` ต้องเป็น `cust_code`
- ถ้าเปลี่ยนชื่อใน code ต้องตรวจ data เดิมด้วย migration หรือ rename step เสมอ

### Code Changes

- เพิ่มกฎถาวรใน `AGENTS.md` หัวข้อ `Storage Naming Rule`
- เพิ่ม normalizer กลางที่ `backend/pkg/microservice/storage_name.go`
- เพิ่ม field normalizer กลาง `NormalizeStorageFieldName`
- ปรับ Mongo persister ให้ normalize ผ่าน `NormalizeMongoCollectionName`
- ปรับ ClickHouse persister ให้ normalize ผ่าน `NormalizeClickHouseTableName`
- เพิ่ม command migration ที่ `backend/cmd/storage_name_migration`
- เพิ่มเอกสารใช้งาน command ที่ `backend/cmd/storage_name_migration/README.md`
- ปรับ SQL/table/collection reference ใน backend ให้ใช้ชื่อใหม่
- ปรับ prompt ของ AI query generator ให้แนะนำ table ใหม่ เช่น `product_barcodes`, `doc_detail`
- เพิ่ม test สำหรับชื่อ legacy ที่ต้อง map เป็นชื่อใหม่

### Migration Tool Behavior

Command หลัก:

```powershell
go run ./cmd/storage_name_migration --target all
```

Apply จริง:

```powershell
go run ./cmd/storage_name_migration --target all --apply
```

พฤติกรรมของ tool:

- default เป็น dry-run
- ถ้า target name มีอยู่แล้ว จะรายงาน conflict และไม่ merge data อัตโนมัติ
- rename MongoDB ด้วย `renameCollection`
- rename PostgreSQL ด้วย `ALTER TABLE ... RENAME TO`
- rename ClickHouse ด้วย `RENAME TABLE`
- ไม่ rename column, index, view, Kafka topic, backup, dashboard หรือ external SQL file

### MongoDB Dev Migration Record

รายการที่ rename แล้วใน dev:

| Old Name | New Name |
| --- | --- |
| `aiProviderConfigs` | `ai_provider_configs` |
| `chatSessions` | `chat_sessions` |
| `kbDocumentMetadata` | `kb_document_metadata` |
| `organizationBranches` | `organization_branches` |
| `organizationBusinessTypes` | `organization_business_types` |
| `organizationCostCenters` | `organization_cost_centers` |
| `organizationJobProjects` | `organization_job_projects` |
| `productBarcodes` | `product_barcodes` |
| `productBarcodesPriceHistory` | `product_barcodes_price_history` |
| `productCategories` | `product_categories` |
| `productGroups` | `product_groups` |
| `shopUserAccessLogs` | `shop_user_access_logs` |
| `shopUsers` | `shop_users` |
| `transactionPurchaseRequisition` | `transaction_purchase_requisition` |

ผลตรวจหลัง apply:

```text
storage name migration: no rename needed
```

### PostgreSQL Dev Migration Record

ผล dry-run ล่าสุด:

```text
storage name migration: no rename needed
```

หมายเหตุ:

- connection ผ่าน `localhost:5432` ใช้ได้
- connection ตาม host ใน `bootstrap.json` เคยติด `pg_hba.conf`/SSL ในเครื่อง dev จึงใช้ local DSN สำหรับตรวจ

### ClickHouse Dev Migration Record

รายการสำคัญที่แก้:

| Old Name | New Name | Note |
| --- | --- | --- |
| `productbarcode` | `product_barcodes` | table |
| `productbarcode_dict` | `product_barcode_dict` | dictionary |

หมายเหตุ:

- ต้อง drop/recreate dictionary เพราะ `productbarcode_dict` อ้าง query ไปที่ table เดิม
- dictionary ใหม่ชี้ไปที่ `bcbidev.product_barcodes`

ผลตรวจหลัง apply:

```text
storage name migration: no rename needed
```

### Important Legacy Alias Examples

ตัวอย่าง alias ที่ต้อง map เป็นชื่อใหม่ใน normalizer:

| Legacy Name | Normalized Name |
| --- | --- |
| `productbarcode` | `product_barcodes` |
| `productbarcode_dict` | `product_barcode_dict` |
| `productbarcodeboms` | `product_barcode_boms` |
| `productbarcodeimport` | `product_barcode_import` |
| `docdetail` | `doc_detail` |
| `productunit` | `product_unit` |
| `chartofaccounts` | `chart_of_accounts` |
| `purchaserequisition` | `purchase_requisition` |
| `saleinvoice` | `sale_invoice` |
| `purchaseorder` | `purchase_order` |
| `shopUsers` | `shop_users` |
| `shopUserAccessLogs` | `shop_user_access_logs` |

### Important Field Alias Examples

ตัวอย่าง field/column legacy ที่ต้อง map เป็นชื่อใหม่:

| Legacy Field | Normalized Field |
| --- | --- |
| `docDate` | `doc_date` |
| `docdate` | `doc_date` |
| `docdatetime` | `doc_datetime` |
| `custCode` | `cust_code` |
| `custcode` | `cust_code` |
| `holdingcode` | `holdingcode` |
| `transflag` | `transflag` |
| `itemcode` | `itemcode` |
| `unitcode` | `unitcode` |
| `whcode` | `wh_code` |
| `warehousecode` | `warehouse_code` |
| `branchcode` | `branchcode` |

หมายเหตุ: รอบ 2026-05-20/2026-05-21 ยังไม่ apply rename field จริงใน database เพราะ field migration ต้องตรวจ query/API/report ทุกจุดและต้องมี backup ก่อนเสมอ

### Verification Record

Commands verified:

```powershell
go test ./cmd/storage_name_migration -count=1
go test ./run -run Test -count=1
npm run typecheck
npm run lint
npm test -- --run
curl.exe --max-time 10 -s -i http://localhost:8888/healthz
curl.exe --max-time 10 -s -i http://localhost:8888/goapi/api/language/th
```

Docker Linux verification for CGO/Kafka package:

```powershell
docker run --rm -v "${PWD}:/app" -w /app golang:1.26-bookworm go test ./pkg/microservice -run TestNormalize -count=1
```

Browser smoke:

- `http://localhost:3000` เปิดหน้า login ได้
- ปุ่มทดสอบ backend แสดงผลเชื่อมต่อสำเร็จ
- dev login `jaturapornchai@gmail.com` ไปหน้าเลือกกิจการได้

Static scan:

- ไม่พบ `Collection("CamelCase")` ใน path ที่ตรวจ
- เหลือชื่อ legacy เฉพาะ test alias และ alias map ที่ตั้งใจเก็บไว้เพื่อรองรับ migration
- เพิ่ม test สำหรับ field normalizer เช่น `docDate/docdate -> doc_date` และ `custCode/custcode -> cust_code`

### Known Risks

- มีหลาย module อ้างชื่อ storage เดิม จึงต้อง review diff ทุกครั้งก่อน merge
- production ต้อง backup database ก่อน apply
- ถ้ามีทั้งชื่อเก่าและชื่อใหม่อยู่พร้อมกัน tool จะไม่ merge data ให้เอง
- external SQL, dashboard, report, BI job หรือ script นอก repo อาจยังอ้างชื่อเดิม ต้องตรวจแยก
- Windows host อาจรัน Go package ที่ใช้ Kafka/CGO ไม่ผ่านถ้าไม่มี GCC แต่ Docker Linux Go 1.26 ใช้ตรวจแทนได้

### Future Migration Checklist

ก่อน apply:

1. ตรวจ branch และ environment ให้ชัดว่าเป็น dev, staging หรือ production
2. backup MongoDB, PostgreSQL และ ClickHouse
3. run dry-run ด้วย config เดียวกับ environment เป้าหมาย
4. ตรวจ conflict ทุกบรรทัด
5. ถ้ามี dictionary/view/materialized view ให้ตรวจ dependency ก่อน rename

หลัง apply:

1. run dry-run ซ้ำ ต้องได้ `storage name migration: no rename needed`
2. ตรวจ `healthz`
3. ตรวจ language API
4. ตรวจหน้า login/workspace ผ่าน browser
5. run focused backend/frontend tests
6. บันทึกผลในเอกสารนี้ทุกครั้ง
