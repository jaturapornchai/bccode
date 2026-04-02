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
    { "id": "title", "type": "text", "fixedWidth": 140, "text": "TAX INVOICE", "fontSize": 12, "fontBold": true, "textAlign": 1 }
  ]
}
```

### Label + Data Pair (2 cells)
```json
{
  "id": "row_addr",
  "cells": [
    { "id": "addr_label", "type": "text", "fixedWidth": 80, "text": "Address:", "fontSize": 7 },
    { "id": "addr_data", "type": "dataField", "flex": 1, "dataBindingKey": "company_address", "fontSize": 7 }
  ]
}
```

### Dual Label + Data (4 cells) — MAX RECOMMENDED
```json
{
  "id": "row_info",
  "cells": [
    { "id": "buyer_label", "type": "text", "fixedWidth": 80, "text": "Buyer:", "fontSize": 7 },
    { "id": "buyer_name", "type": "dataField", "flex": 1, "dataBindingKey": "contact_name", "fontSize": 7 },
    { "id": "docno_label", "type": "text", "fixedWidth": 60, "text": "Doc No.:", "fontSize": 7 },
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
        ["No.", "Description", "Qty", "Unit", "Unit Price", "Amount"],
        ["1", "Sample Product", "10", "PCS", "100.00", "1,000.00"],
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
| No. | number | 1 (center) |
| Code | text | 0 (left) |
| Description | text | 0 (left) |
| Qty | number | 2 (right) |
| Unit | text | 1 (center) |
| Unit Price | money | 2 (right) |
| Discount | money | 2 (right) |
| Amount | money | 2 (right) |

## Footer Patterns

### Total Row (right-aligned)
```json
{
  "id": "row_subtotal",
  "cells": [
    { "id": "subtotal_label", "type": "text", "fixedWidth": 100, "text": "Subtotal", "fontSize": 7, "textAlign": 2, "fontBold": true },
    { "id": "subtotal", "type": "dataField", "flex": 1, "dataBindingKey": "subtotal", "fontSize": 7, "textAlign": 2 }
  ]
}
```

### Grand Total Row (bold, larger)
```json
{
  "id": "row_grand",
  "cells": [
    { "id": "grand_label", "type": "text", "fixedWidth": 100, "text": "Grand Total", "fontSize": 9, "fontBold": true, "textAlign": 2 },
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
    { "id": "sign1", "type": "text", "flex": 1, "text": "Prepared By", "fontSize": 7, "textAlign": 1 },
    { "id": "sign2", "type": "text", "flex": 1, "text": "Approved By", "fontSize": 7, "textAlign": 1 },
    { "id": "sign3", "type": "text", "flex": 1, "text": "Received By", "fontSize": 7, "textAlign": 1 }
  ]
}
```

For 2-signature version, use 2 cells with `"gap": 60`.

## 12 Template Types

| docType | Name | Signatures | Special Fields |
|---------|------|------------|----------------|
| tax_invoice | Tax Invoice | 3: Prepared/Approved/Received | copy_label, credit_days |
| abbreviated_tax_invoice | Abbreviated Tax Invoice | 2: Seller/Buyer | VAT included |
| receipt | Receipt | 3: Received/Paid/Approved | payment method |
| cash_bill | Cash Bill | 2: Seller/Buyer | simpler header |
| quotation | Quotation | 3: Quoted/Approved/Ordered | valid_until |
| purchase_order | Purchase Order | 3: Prepared/Approved/Seller | delivery_date, shipping |
| delivery_note | Delivery Note | 2: Sender/Receiver | no pricing columns |
| invoice | Invoice | 3: Prepared/Approved/Received | due_date, credit_days |
| credit_note | Credit Note | 3: Prepared/Approved/Received | original_ref, reason |
| debit_note | Debit Note | 3: Prepared/Approved/Received | original_ref, reason |
| booking_form | Booking Form | 3: Booker/Accepted/Approved | deposit_amount |
| payment_voucher | Payment Voucher | 2: Payer/Payee | withholding_tax |

## Color Values (ARGB32)
- Black: `4278190080` (0xFF000000)
- White: `4294967295` (0xFFFFFFFF)
- Gray: `4288585374` (0xFF9E9E9E)
- Use `Color(value).toARGB32()` in Dart to generate
