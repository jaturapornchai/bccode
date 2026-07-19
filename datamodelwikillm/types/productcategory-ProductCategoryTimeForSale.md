---
source: backend/internal/product/productcategory/models/productcategory.go
tags: [datamodel, general-type]
---

# ProductCategoryTimeForSale

โครงสร้าง `ProductCategoryTimeForSale` จากโมดูล mainapi `productcategory` มีฟิลด์ตามซอร์ส `productcategory.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DaysOfWeek | []int8 | daysofweek | daysofweek | ค่าของ DaysOfWeek ตามฟิลด์ `daysofweek` ในซอร์ส |
| FromDate | string | fromdate | fromdate | วันที่เริ่มช่วง |
| ToDate | string | todate | todate | วันที่สิ้นสุดช่วง |
| FromTime | string | fromtime | fromtime | ค่าของ FromTime ตามฟิลด์ `fromtime` ในซอร์ส |
| ToTime | string | totime | totime | ค่าของ ToTime ตามฟิลด์ `totime` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
