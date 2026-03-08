# Party Models

ลูกหนี้ (Debtor), เจ้าหนี้ (Creditor), สมาชิก (Member)

## Debtor (ลูกหนี้ / ลูกค้า)

**Backend:** `internal/debtaccount/debtor/models/debtor.go`
**Frontend:** `lib/model/customer_model.dart` → `CustomerModel`
**MongoDB Collection:** `debtorMaster`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *PartitionIdentity | — | — | inline | ParID (ซ่อน) |
| Code | string | `code` | `code` | รหัสลูกหนี้ |
| PersonalType | int8 | `personaltype` | `personaltype` | 1=บุคคลธรรมดา, 2=นิติบุคคล |
| Names | *[]NameX | `names` | `names` | ชื่อ (หลายภาษา) |
| Images | *[]Image | `images` | `images` | รูปภาพ |
| PointBalance | float64 | `pointbalance` | `pointbalance` | แต้มสะสม |
| PointsCode | string | `pointscode` | `pointscode` | รหัสกลุ่มแต้ม |
| AddressForBilling | Address | `addressforbilling` | `addressforbilling` | ที่อยู่เรียกเก็บ |
| AddressForShipping | *[]Address | `addressforshipping` | `addressforshipping` | ที่อยู่จัดส่ง (หลายที่) |
| TaxId | string | `taxid` | `taxid` | เลขประจำตัวผู้เสียภาษี |
| Email | string | `email` | `email` | อีเมล |
| CustomerType | int | `customertype` | `customertype` | 1=สำนักงานใหญ่, 2=สาขา |
| BranchNumber | string | `branchnumber` | `branchnumber` | เลขสาขา |
| FundCode | string | `fundcode` | `fundcode` | รหัสกองทุน |
| CreditDay | int | `creditday` | `creditday` | เครดิตเทอม (วัน) |
| IsMember | bool | `ismember` | `ismember` | เป็นสมาชิก |
| GroupGUIDs | *[]string | `groups` | `groups` | กลุ่มลูกค้า |
| Auth | DebtorAuth | `auth` | `auth` | ข้อมูล auth |
| DebtorLine | DebtorLine | `line` | `line` | LINE OA info |
| PriceLevel | int | `pricelevel` | `pricelevel` | ระดับราคา (1-9) |
| IsCreditor | bool | `iscreditor` | `iscreditor` | เป็นเจ้าหนี้ด้วย |
| Telephones | *[]Telephone | `telephones` | `telephones` | เบอร์โทร |

### Address (ที่อยู่)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| Address | *[]string | `address` | `address` | บรรทัดที่อยู่ |
| CountryCode | string | `countrycode` | `countrycode` | รหัสประเทศ |
| ProvinceCode | string | `provincecode` | `provincecode` | รหัสจังหวัด |
| DistrictCode | string | `districtcode` | `districtcode` | รหัสอำเภอ |
| SubDistrictCode | string | `subdistrictcode` | `subdistrictcode` | รหัสตำบล |
| ZipCode | string | `zipcode` | `zipcode` | รหัสไปรษณีย์ |
| ContactNames | *[]NameX | `contactnames` | `contactnames` | ชื่อผู้ติดต่อ |
| PhonePrimary | string | `phoneprimary` | `phoneprimary` | เบอร์หลัก |
| PhoneSecondary | string | `phonesecondary` | `phonesecondary` | เบอร์รอง |
| Latitude | float64 | `latitude` | `latitude` | ละติจูด |
| Longitude | float64 | `longitude` | `longitude` | ลองจิจูด |

### DebtorAuth

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| Username | string | `username` | ชื่อผู้ใช้ |
| Password | string | `password` | รหัสผ่าน |

### DebtorLine (LINE OA)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| LineUID | string | `lineuid` | LINE User ID |
| LineDisplayName | string | `linedisplayname` | ชื่อแสดงใน LINE |
| LinePictureURL | string | `linepictureurl` | URL รูปโปรไฟล์ |

---

## Creditor (เจ้าหนี้ / ผู้จำหน่าย)

**Backend:** `internal/debtaccount/creditor/models/creditor.go`
**MongoDB Collection:** `creditorMaster`

โครงสร้างเหมือน Debtor แต่ไม่มี field เฉพาะลูกค้า (Point, Line, IsMember):

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *PartitionIdentity | — | — | inline | ParID |
| Code | string | `code` | `code` | รหัสเจ้าหนี้ |
| PersonalType | int8 | `personaltype` | `personaltype` | 1=บุคคล, 2=นิติบุคคล |
| Names | *[]NameX | `names` | `names` | ชื่อ (หลายภาษา) |
| Images | *[]Image | `images` | `images` | รูปภาพ |
| AddressForBilling | Address | `addressforbilling` | `addressforbilling` | ที่อยู่เรียกเก็บ |
| AddressForShipping | *[]Address | `addressforshipping` | `addressforshipping` | ที่อยู่จัดส่ง |
| TaxId | string | `taxid` | `taxid` | เลขภาษี |
| Email | string | `email` | `email` | อีเมล |
| CustomerType | int | `customertype` | `customertype` | ประเภทสาขา |
| BranchNumber | string | `branchnumber` | `branchnumber` | เลขสาขา |
| CreditDay | int | `creditday` | `creditday` | เครดิตเทอม |
| IsDebtor | bool | `isdebtor` | `isdebtor` | เป็นลูกหนี้ด้วย |
| Telephones | *[]Telephone | `telephones` | `telephones` | เบอร์โทร |
| GroupGUIDs | *[]string | `groups` | `groups` | กลุ่ม |

---

## Debtor/Creditor (ClickHouse)

ClickHouse เก็บเฉพาะข้อมูลย่อ:

### debtors table

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน |
| guidfixed | String | GUID |
| code | String | รหัสลูกหนี้ |
| name1 | String | ชื่อภาษาที่ 1 |
| name2 | String | ชื่อภาษาที่ 2 |

### creditors table

| Column | Type | หมายเหตุ |
|--------|------|----------|
| shopid | String | รหัสร้าน |
| guidfixed | String | GUID |
| code | String | รหัสเจ้าหนี้ |
| name1 | String | ชื่อภาษาที่ 1 |
| name2 | String | ชื่อภาษาที่ 2 |

---

## Member (สมาชิก)

**Backend:** `internal/member/models/member.go`
**MongoDB Collection:** `memberMaster`

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| *Identity | — | — | inline | shopid + guidfixed |
| Code | string | `code` | `code` | รหัสสมาชิก |
| Names | *[]NameX | `names` | `names` | ชื่อ (หลายภาษา) |
| MaxPoint | float64 | `maxpoint` | `maxpoint` | แต้มสูงสุด |
| PointBalance | float64 | `pointbalance` | `pointbalance` | แต้มคงเหลือ |
| Telephone | string | `telephone` | `telephone` | เบอร์โทร |
| Email | string | `email` | `email` | อีเมล |
| Birthday | string | `birthday` | `birthday` | วันเกิด |
| Gender | int8 | `gender` | `gender` | เพศ |
| Images | *[]Image | `images` | `images` | รูปภาพ |
| TierCode | string | `tiercode` | `tiercode` | ระดับสมาชิก |
| Address | Address | `address` | `address` | ที่อยู่ |

---

## Customer (Frontend — Flutter)

**Frontend:** `lib/model/customer_model.dart` → `CustomerModel`

ใช้ model เดียวกันสำหรับทั้ง Debtor และ Creditor:

| Field | Dart Type | JSON Key | หมายเหตุ |
|-------|-----------|----------|----------|
| guidfixed | String | `guidfixed` | GUID |
| code | String | `code` | รหัส |
| names | List\<LanguageDataModel\> | `names` | ชื่อ |
| personaltype | int | `personaltype` | 1=บุคคล, 2=นิติบุคคล |
| customertype | int | `customertype` | 1=สนง.ใหญ่, 2=สาขา |
| taxid | String | `taxid` | เลขภาษี |
| fundcode | String | `fundcode` | รหัสกองทุน |
| creditday | int | `creditday` | เครดิตเทอม |
| iscreditor | bool | `iscreditor` | เป็นเจ้าหนี้ |
| isdebtor | bool | `isdebtor` | เป็นลูกหนี้ |
| email | String | `email` | อีเมล |
| branchnumber | String | `branchnumber` | เลขสาขา |
| addressforbilling | CustomerAddressModel | `addressforbilling` | ที่อยู่เรียกเก็บ |
| addressforshipping | List\<CustomerAddressModel\> | `addressforshipping` | ที่อยู่จัดส่ง |
| groups | List\<CustomerGroupModel\> | `groups` | กลุ่ม |
| images | List\<ImagesModel\> | `images` | รูปภาพ |
