# Template Row Patterns

Standard patterns for building form template JSON files.

## Header Patterns

### Company + Title Row (3 cells)
```json
{
  "id": "row_0",
  "cells": [
    { "id": "logo", "type": "image", "fixedWidth": 60, "minHeight": 50, "imagePath": "company_logo" },
    { "id": "company_name", "type": "dataField", "flex": 1, "dataBindingKey": "company_name", "fontSize": 10, "fontBold": true },
    { "id": "title", "type": "text", "fixedWidth": 140, "text": "ใบกำกับภาษี / TAX INVOICE", "fontSize": 12, "fontBold": true, "textAlign": 1 }
  ]
}
```

### Label + Data Pair (2 cells)
```json
{
  "id": "row_addr",
  "cells": [
    { "id": "addr_label", "type": "text", "fixedWidth": 80, "text": "ที่อยู่:", "fontSize": 7 },
    { "id": "addr_data", "type": "dataField", "flex": 1, "dataBindingKey": "company_address", "fontSize": 7 }
  ]
}
```

### Dual Label + Data (4 cells) - MAX RECOMMENDED
```json
{
  "id": "row_info",
  "cells": [
    { "id": "buyer_label", "type": "text", "fixedWidth": 80, "text": "ผู้ซื้อ:", "fontSize": 7 },
    { "id": "buyer_name", "type": "dataField", "flex": 1, "dataBindingKey": "contact_name", "fontSize": 7 },
    { "id": "docno_label", "type": "text", "fixedWidth": 60, "text": "เลขที่:", "fontSize": 7 },
    { "id": "doc_number", "type": "dataField", "flex": 0.6, "dataBindingKey": "doc_number", "fontSize": 7 }
  ]
}
```

### Separator Row (1 cell)
```json
{
  "id": "row_sep",
  "paddingTop": 4,
  "paddingBottom": 4,
  "cells": [
    { "id": "sep", "type": "separator", "flex": 1, "borderColor": 4278190080, "borderWidth": 0.5 }
  ]
}
```

## Detail Pattern

### Items Table (1 cell spanning full width)
```json
{
  "id": "row_table",
  "cells": [
    {
      "id": "items_table",
      "type": "table",
      "flex": 1,
      "columns": 6,
      "rows": 5,
      "tableHasHeader": true,
      "columnAligns": [1, 0, 2, 1, 2, 2],
      "borderColor": 4278190080,
      "borderWidth": 0.5,
      "fontSize": 7,
      "tableData": [
        ["ลำดับ", "รายการ", "จำนวน", "หน่วย", "ราคา/หน่วย", "จำนวนเงิน"],
        ["1", "สินค้าตัวอย่าง", "10", "ชิ้น", "100.00", "1,000.00"],
        ["", "", "", "", "", ""],
        ["", "", "", "", "", ""],
        ["", "", "", "", "", ""]
      ]
    }
  ]
}
```

### Column Align Convention
| Column | Content | Align |
|--------|---------|-------|
| ลำดับ (No.) | number | 1 (center) |
| รหัส (Code) | text | 0 (left) |
| รายการ (Desc) | text | 0 (left) |
| จำนวน (Qty) | number | 2 (right) |
| หน่วย (Unit) | text | 1 (center) |
| ราคา/หน่วย (Price) | money | 2 (right) |
| ส่วนลด (Discount) | money | 2 (right) |
| จำนวนเงิน (Amount) | money | 2 (right) |

## Footer Patterns

### Total Row (right-aligned amount)
```json
{
  "id": "row_subtotal",
  "cells": [
    { "id": "subtotal_label", "type": "text", "fixedWidth": 100, "text": "รวมเงิน", "fontSize": 7, "textAlign": 2, "fontBold": true },
    { "id": "subtotal", "type": "dataField", "flex": 1, "dataBindingKey": "subtotal", "fontSize": 7, "textAlign": 2 }
  ]
}
```

### Grand Total Row (bold, larger)
```json
{
  "id": "row_grand",
  "cells": [
    { "id": "grand_label", "type": "text", "fixedWidth": 100, "text": "จำนวนเงินรวมทั้งสิ้น", "fontSize": 9, "fontBold": true, "textAlign": 2 },
    { "id": "grand_total", "type": "dataField", "flex": 1, "dataBindingKey": "grand_total", "fontSize": 9, "fontBold": true, "textAlign": 2 }
  ]
}
```

### Signature Section (3 signatures with gap)
```json
{
  "id": "f_sign_sep",
  "gap": 30,
  "cells": [
    { "id": "sep1", "type": "separator", "flex": 1, "borderColor": 4278190080, "borderWidth": 0.5 },
    { "id": "sep2", "type": "separator", "flex": 1, "borderColor": 4278190080, "borderWidth": 0.5 },
    { "id": "sep3", "type": "separator", "flex": 1, "borderColor": 4278190080, "borderWidth": 0.5 }
  ]
},
{
  "id": "f_sign_label",
  "gap": 30,
  "cells": [
    { "id": "sign1", "type": "text", "flex": 1, "text": "ผู้จัดทำ", "fontSize": 7, "textAlign": 1 },
    { "id": "sign2", "type": "text", "flex": 1, "text": "ผู้อนุมัติ", "fontSize": 7, "textAlign": 1 },
    { "id": "sign3", "type": "text", "flex": 1, "text": "ผู้รับสินค้า", "fontSize": 7, "textAlign": 1 }
  ]
}
```

### 2-Signature Version (with gap)
Use same pattern but with 2 cells and `"gap": 60`.

## 12 Template Types

| docType | Thai Name | Signatures | Special Fields |
|---------|-----------|------------|----------------|
| tax_invoice | ใบกำกับภาษี | 3: จัดทำ/อนุมัติ/รับสินค้า | copy_label, credit_days |
| abbreviated_tax_invoice | ใบกำกับภาษีอย่างย่อ | 2: ผู้ขาย/ผู้ซื้อ | VAT included |
| receipt | ใบเสร็จรับเงิน | 3: รับเงิน/จ่ายเงิน/อนุมัติ | payment method |
| cash_bill | บิลเงินสด | 2: ผู้ขาย/ผู้ซื้อ | simpler header |
| quotation | ใบเสนอราคา | 3: เสนอราคา/อนุมัติ/สั่งซื้อ | valid_until |
| purchase_order | ใบสั่งซื้อ | 3: จัดทำ/อนุมัติ/ผู้ขาย | delivery_date, shipping |
| delivery_note | ใบส่งของ | 2: ผู้ส่ง/ผู้รับ | no pricing columns |
| invoice | ใบแจ้งหนี้ | 3: จัดทำ/อนุมัติ/รับเอกสาร | due_date, credit_days |
| credit_note | ใบลดหนี้ | 3: จัดทำ/อนุมัติ/รับเอกสาร | original_ref, reason |
| debit_note | ใบเพิ่มหนี้ | 3: จัดทำ/อนุมัติ/รับเอกสาร | original_ref, reason |
| booking_form | ใบรับจอง | 3: ผู้จอง/รับจอง/อนุมัติ | deposit_amount |
| payment_voucher | ใบสำคัญจ่าย | 2: ผู้จ่าย/ผู้รับ | withholding_tax |

## Color Values (ARGB32)
- Black: `4278190080` (0xFF000000)
- White: `4294967295` (0xFFFFFFFF)
- Gray: `4288585374` (0xFF9E9E9E)
- Use `Color(value).toARGB32()` in Dart to generate
