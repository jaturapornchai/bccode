# Enums & Constants

ค่าคงที่ที่ใช้ทั้ง system — Backend (Go) / Frontend (Dart) / Database

## TransFlag (ประเภทเอกสาร)

**สำคัญที่สุด** — field `transflag` ระบุประเภทเอกสารใน doc/docdetail ทุกตาราง

### เอกสารหลัก (ใช้บ่อย)

| TransFlag | ประเภท | CalcFlag | หมายเหตุ |
|-----------|--------|----------|----------|
| **44** | Sale Invoice (ขาย) | -1 (stock ออก) | ใบขาย |
| **48** | Sale Return (รับคืนจากลูกค้า) | 1 (stock เข้า) | ใบรับคืน |
| **12** | Purchase (ซื้อ) | 1 (stock เข้า) | ใบซื้อ |
| **16** | Purchase Return (ส่งคืนผู้จำหน่าย) | -1 (stock ออก) | ใบส่งคืน |
| **54** | Opening Balance (ยอดยกมา) | 1 (stock เข้า) | ยอดเปิด |
| **56** | Stock Pickup (เบิก) | -1 (stock ออก) | ใบเบิกสินค้า |
| **58** | Stock Return (รับคืนเบิก) | 1 (stock เข้า) | ใบรับคืนจากเบิก |
| **60** | Stock Receive (รับ) | 1 (stock เข้า) | ใบรับสินค้า |
| **66** | Adjust Increase (ปรับเพิ่ม) | 1 (stock เข้า) | ปรับยอด stock เพิ่ม |
| **68** | Adjust Decrease (ปรับลด) | -1 (stock ออก) | ปรับยอด stock ลด |
| **72** | Stock Transfer (โอน) | -1/1 (ออก/เข้า) | โอนสินค้าระหว่างคลัง |

### TransFlag ใน ClickHouse (Stock Calculation)

stock balance query ใช้ transflag เหล่านี้:
```sql
WHERE transflag IN (1,3,5,7,9,11,13,16,18,20,30,31,32,33,34,35,36)
```

**สูตรคำนวณ stock balance:**
```sql
SUM((totalqty * calcflag) * unitstand / NULLIF(unitdivide, 0))
```

### CalcFlag (ทิศทาง Stock)

| CalcFlag | ความหมาย | ตัวอย่าง |
|----------|----------|----------|
| **1** | Stock เข้า (เพิ่ม) | ซื้อ, รับคืน, รับ, ยกมา, ปรับเพิ่ม |
| **-1** | Stock ออก (ลด) | ขาย, ส่งคืน, เบิก, ปรับลด |

## VatType (ประเภท VAT)

| ค่า | ความหมาย |
|-----|----------|
| **0** | ไม่มี VAT |
| **1** | รวม VAT แล้ว (Inclusive) |
| **2** | อัตราศูนย์ (Zero-rated) |
| **3** | ยกเว้น VAT (Exempt) |

## InquiryType (ประเภทการขาย)

| ค่า | ความหมาย |
|-----|----------|
| **0** | เงินเชื่อ (Credit) |
| **1** | เงินสด (Cash) |

## DocType (ประเภทเอกสาร)

| ค่า | ความหมาย |
|-----|----------|
| **0** | ขาย (Sale) |
| **1** | ซื้อ (Purchase) |

## PersonalType (ประเภทบุคคล)

| ค่า | ความหมาย |
|-----|----------|
| **1** | บุคคลธรรมดา (Individual) |
| **2** | นิติบุคคล (Business/Corporate) |

## CustomerType (ประเภทสาขา)

| ค่า | ความหมาย |
|-----|----------|
| **1** | สำนักงานใหญ่ (Head Office) |
| **2** | สาขา (Branch) |

## ItemType (ประเภทสินค้า)

| ค่า | ความหมาย |
|-----|----------|
| **0** | สินค้าทั่วไป (General) |
| **1** | อาหาร (Food) |
| **2** | เครื่องดื่ม (Beverage) |

## ItemStockType (ประเภท Stock)

| ค่า | ความหมาย |
|-----|----------|
| **0** | สินค้า stock (Stock item) |
| **1** | บริการ (Service / Non-stock) |

## CouponType (ประเภทคูปอง)

| ค่า | ความหมาย |
|-----|----------|
| **0** | จำนวนเงิน (Amount) |
| **1** | เปอร์เซ็นต์ (Percentage) |
| **2** | เงินสด (Cash) |

## PaymentStruct TransFlag (ประเภทการชำระ)

| trans_flag | ความหมาย |
|-----------|----------|
| **1** | บัตรเครดิต (Credit Card) |
| **2** | โอนเงิน (Bank Transfer) |
| **3** | เช็ค (Cheque) |
| **4** | คูปอง (Coupon) |
| **5** | QR Code |

## WHT Rate (อัตราภาษีหัก ณ ที่จ่าย)

| อัตรา (%) | ประเภทเงินได้ (Description) |
|-----------|---------------------------|
| 1.0 | ค่าขนส่ง (Transport) |
| 1.0 | ค่าเบี้ยประกันภัย (Insurance) |
| 2.0 | ค่าโฆษณา (Advertising) |
| 3.0 | ค่าบริการ (Service) |
| 3.0 | ค่าจ้างก่อสร้าง (Construction) |
| 3.0 | ค่าลิขสิทธิ์ (Royalty) |
| 5.0 | ค่าเช่า (Rental) |
| 10.0 | เงินปันผล (Dividend) |

## Language Codes (รหัสภาษา)

| Code | ภาษา |
|------|------|
| `th` | ไทย |
| `en` | English |
| `vi` | Tiếng Việt |
| `lo` | ลาว |
| `km` | ខ្មែរ (Khmer) |
| `my` | မြန်မာ (Myanmar) |
| `cn` | 中文 (Chinese) |
| `ja` | 日本語 (Japanese) |
| `ko` | 한국어 (Korean) |

## Approval Status

| ค่า | ความหมาย |
|-----|----------|
| `""` (empty) | ไม่มี approval |
| `pending` | รออนุมัติ |
| `approved` | อนุมัติแล้ว |
| `rejected` | ปฏิเสธ |

## YearType (ประเภทปฎิทิน)

| ค่า | ความหมาย |
|-----|----------|
| `christian` | ค.ศ. (2026) |
| `buddhist` | พ.ศ. (2569) |

## Date Format

| ค่า | ตัวอย่าง |
|-----|----------|
| `dd/MM/yyyy` | 28/02/2026 |
| `yyyy-MM-dd` | 2026-02-28 |
| `MM/dd/yyyy` | 02/28/2026 |
