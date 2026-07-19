---
source: report-model.go
tags: [datamodel, general-type]
---

# ReportDataModel

แถวข้อมูลหนึ่งแถวในส่วนเนื้อหาของรายงาน ระบุลำดับแถว เส้นบน/ล่าง และข้อมูลรายคอลัมน์

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| RowIndex | int | rowindex | rowindex | ลำดับของแถว |
| TopLine | bool | topline | topline | แสดงเส้นบน |
| BottomLine | bool | bottomline | bottomline | แสดงเส้นล่าง |
| ColumnData | `[]`[[ReportDataColumnModel]] | nextrowdata | nextrowdata | ข้อมูลรายคอลัมน์ของแถว (tag ใช้ชื่อ nextrowdata) |

## ความสัมพันธ์

- ฝัง [[ReportDataColumnModel]] เป็นรายการข้อมูลคอลัมน์
- ถูกฝังใน [[ReportModel]] (ฟิลด์ DataRows)
