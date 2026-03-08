# Organization Models

สาขา, แผนก, คลังสินค้า, POS, พนักงาน

## Branch (สาขา)

**Backend:** `internal/organization/branch/models/branch.go`
**Frontend:** `lib/model/company_branch_model.dart` → `CompanyBranchModel`
**MongoDB Collection:** `branchMaster`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัสสาขา |
| Names | *[]NameX | `names` | `names` | ชื่อสาขา (หลายภาษา) |
| CompanyNames | *[]NameX | `companynames` | `companynames` | ชื่อบริษัท |
| Languages | *[]string | `languages` | `languages` | ภาษาที่รองรับ |
| Contact | BranchContact | `contact` | `contact` | ข้อมูลติดต่อ |
| LogoUri | string | `logouri` | `logouri` | URL โลโก้ |
| ImageUri | string | `imageuri` | `imageuri` | URL รูปภาพ |

### BranchContact

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| Address | *[]string | `address` | ที่อยู่ |
| Telephones | *[]Telephone | `telephones` | เบอร์โทร |
| Emails | *[]string | `emails` | อีเมล |

### Branch Settings (Frontend)

Frontend เก็บ settings เพิ่มเติมสำหรับ Branch:

| Field | Dart Type | JSON Key | หมายเหตุ |
|-------|-----------|----------|----------|
| base_currency | String | `base_currency` | สกุลเงินหลัก (THB, USD, VND, ...) |
| language | String | `language` | ภาษาหลัก (th, en, vi, ...) |
| timezone | String | `timezone` | Timezone (Asia/Bangkok, ...) |
| date_format | String | `date_format` | รูปแบบวันที่ (dd/MM/yyyy) |
| yeartype | String | `yeartype` | christian / buddhist |
| timezonelabel | String | `timezonelabel` | ป้าย timezone |
| timezoneoffset | String | `timezoneoffset` | +07:00 |
| decimal_quantity | int | `decimal_quantity` | ทศนิยมจำนวน (2-8) |
| decimal_price | int | `decimal_price` | ทศนิยมราคา (2-8) |
| decimal_document | int | `decimal_document` | ทศนิยมยอดเงิน (2-8) |

---

## Department (แผนก)

**Backend:** `internal/organization/department/models/department.go`
**MongoDB Collection:** `departmentMaster`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัสแผนก |
| Names | *[]NameX | `names` | `names` | ชื่อแผนก (หลายภาษา) |

---

## Warehouse (คลังสินค้า)

**Backend:** `internal/warehouse/models/warehouse.go`
**Frontend:** ใช้ใน transaction detail (whcode, whnames)
**MongoDB Collection:** `warehouseMaster`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัสคลัง |
| Names | *[]NameX | `names` | `names` | ชื่อคลัง (หลายภาษา) |
| Locations | *[]Location | `locations` | `locations` | ตำแหน่งในคลัง |

### Location (ตำแหน่งในคลัง)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| Code | string | `code` | รหัสตำแหน่ง |
| Names | *[]NameX | `names` | ชื่อตำแหน่ง |

### Warehouse (ClickHouse)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน |
| guidfixed | String | GUID |
| code | String | รหัสคลัง |
| name1 | String | ชื่อภาษาที่ 1 |
| name2 | String | ชื่อภาษาที่ 2 |

### Location (ClickHouse)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน |
| guidfixed | String | GUID |
| locationcode | String | รหัสตำแหน่ง |
| warehousecode | String | รหัสคลัง (parent) |
| name1 | String | ชื่อภาษาที่ 1 |
| name2 | String | ชื่อภาษาที่ 2 |

---

## Employee (พนักงาน)

**Backend:** `internal/models/employee.go` + `internal/employee/models/`
**MongoDB Collection:** `employeeMaster`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัสพนักงาน |
| Names | *[]NameX | `names` | `names` | ชื่อ (หลายภาษา) |
| DepartmentCode | string | `departmentcode` | `departmentcode` | รหัสแผนก |
| DepartmentNames | *[]NameX | `departmentnames` | `departmentnames` | ชื่อแผนก |
| PositionCode | string | `positioncode` | `positioncode` | รหัสตำแหน่ง |
| PositionNames | *[]NameX | `positionnames` | `positionnames` | ชื่อตำแหน่ง |
| BranchCode | string | `branchcode` | `branchcode` | รหัสสาขา |
| Telephones | *[]Telephone | `telephones` | `telephones` | เบอร์โทร |
| Email | string | `email` | `email` | อีเมล |
| ImageUri | string | `imageuri` | `imageuri` | URL รูป |

---

## Shop (ร้านค้า)

**ClickHouse Table:** `shop` (ReplacingMergeTree)

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน (UUID) |
| name | String | ชื่อร้าน |
| checksum | String | checksum |
| mongodbname | String | ชื่อ MongoDB database |

**ความสัมพันธ์:** `shopid` เป็น key หลักของทุกตาราง ใช้เป็น:
- ชื่อ database ใน MongoDB
- ชื่อ database ใน PostgreSQL
- Partition key ใน ClickHouse

---

## SaleChannel (ช่องทางขาย)

**MongoDB Collection:** `saleChannelMaster`
**ClickHouse Table:** `salechannel`

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| *Identity | — | — | shopid + guidfixed |
| Code | string | `code` | รหัสช่องทาง |
| Name | string | `name` | ชื่อช่องทาง |

---

## Coupon (คูปอง — Master)

**Backend:** `internal/coupon/models/`
**Frontend:** `lib/model/coupon_model.dart` → `CouponModel`
**MongoDB Collection:** `couponMaster`

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| GuidFixed | string | `guidfixed` | GUID |
| CouponCode | string | `couponcode` | รหัสคูปอง |
| Names | *[]NameX | `names` | ชื่อคูปอง |
| CouponValue | float64 | `couponvalue` | มูลค่าส่วนลด |
| CouponType | int | `coupontype` | 0=จำนวนเงิน, 1=เปอร์เซ็นต์, 2=เงินสด |
| IssuedDate | string | `issueddate` | วันที่ออก |
| ExpiryDate | string | `expirydate` | วันหมดอายุ |
| CustomerCodes | []string | `customercodes` | ลูกค้าที่ใช้ได้ |
| Status | int | `status` | 0=active, 1=inactive |
| IsOneTimeUse | bool | `isonetimeuse` | ใช้ได้ครั้งเดียว |
| MaxUsageCount | int | `maxusagecount` | จำนวนครั้งสูงสุด |
| MaxUsageCountPerCustomer | int | `maxusagecountpercustomer` | สูงสุดต่อลูกค้า |
| Remark | string | `remark` | หมายเหตุ |
