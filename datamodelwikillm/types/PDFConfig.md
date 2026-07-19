---
source: query-result-model.go
tags: [datamodel, general-type]
---

# PDFConfig

การตั้งค่า PDF (ชื่อรายงาน, แนวกระดาษ, ขนาดกระดาษ) ใช้ฝังใน [[ResultToPDFRequest]]

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Title | string | - | title | ชื่อ PDF (default: "รายงาน") |
| Orientation | string | - | orientation | แนวกระดาษ (L=แนวนอน, P=แนวตั้ง) |
| PageSize | string | - | pagesize | ขนาดกระดาษ (A4, Letter ฯลฯ) |

## ความสัมพันธ์

- ถูกฝังเป็น field `PDFConfig` ใน [[ResultToPDFRequest]]
