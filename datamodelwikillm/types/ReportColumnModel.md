---
source: report-model.go
tags: [datamodel, general-type]
---

# ReportColumnModel

นิยามคอลัมน์หนึ่งคอลัมน์ในแถวของรายงาน กำหนดชื่อ ความกว้าง การจัดตำแหน่ง รูปแบบการแสดงผล และเส้นขอบบน/ล่าง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Name | string | name | name | ชื่อคอลัมน์ |
| Width | float64 | width | width | ความกว้างของคอลัมน์ |
| WidthCalc | float64 | widthcalc | widthcalc | ความกว้างของคอลัมน์ที่คำนวณได้ |
| Align | int | align | align | การจัดตำแหน่งข้อความ (0=left, 1=center, 2=right) |
| ColumnType | int | columntype | columntype | ประเภทของคอลัมน์ |
| Format | string | format | format | รูปแบบการแสดงผล เช่น วันที่ ตัวเลข ทศนิยม |
| Style | int | style | style | สไตล์ตัวอักษร (0=normal, 1=bold, 2=italic, 3=bold-italic) |
| TopLine | bool | topline | topline | แสดงเส้นบน |
| BottomLine | bool | bottomline | bottomline | แสดงเส้นล่าง |

## ความสัมพันธ์

- ถูกฝังเป็นรายการใน [[ReportRowModel]] (ฟิลด์ Columns)
