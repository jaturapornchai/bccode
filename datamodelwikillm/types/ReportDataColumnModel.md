---
source: report-model.go
tags: [datamodel, general-type]
---

# ReportDataColumnModel

ค่าข้อมูลของหนึ่งคอลัมน์ในแถวข้อมูลรายงาน พร้อมรูปแบบวันเวลา จำนวนทศนิยม และสไตล์ตัวอักษร

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Value | any | value | value | ข้อมูลคอลัมน์ (ชนิดใดก็ได้) |
| DateTimeStyle | int | datetimestyle | datetimestyle | รูปแบบวันเวลา (0=none, 1=short, 2=medium, 3=long) |
| Point | int | point | point | จำนวนตำแหน่งทศนิยม |
| Style | int | style | style | สไตล์ตัวอักษร (0=normal, 1=bold, 2=italic, 3=bold-italic) |

## ความสัมพันธ์

- ถูกฝังเป็นรายการใน [[ReportDataModel]] (ฟิลด์ ColumnData)
