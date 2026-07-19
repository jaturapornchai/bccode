---
source: report-model.go
tags: [datamodel, general-type]
---

# ReportRowModel

นิยามแถวหนึ่งแถวของรายงาน (ใช้ใน header/footer) กำหนดระยะขอบซ้าย ขนาดตัวอักษร เส้นบน/ล่าง และรายการคอลัมน์ในแถว

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| LeftMarginPercent | float64 | leftmarginpercent | leftmarginpercent | ระยะห่างด้านซ้ายของแถว (เปอร์เซ็นต์) |
| FontSize | float64 | fontsize | fontsize | ขนาดตัวอักษร |
| TopLine | bool | topline | topline | แสดงเส้นบน |
| BottomLine | bool | bottomline | bottomline | แสดงเส้นล่าง |
| Columns | `[]`[[ReportColumnModel]] | columns | columns | คอลัมน์ต่างๆ ในแถว |

## ความสัมพันธ์

- ฝัง [[ReportColumnModel]] เป็นรายการคอลัมน์
- ถูกฝังใน [[ReportModel]] (ฟิลด์ HeaderRows, FooterRows)
