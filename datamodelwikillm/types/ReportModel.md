---
source: report-model.go
tags: [datamodel, general-type]
---

# ReportModel

โครงสร้างหลักของรายงานหนึ่งฉบับ ประกอบด้วยชื่อรายงาน ส่วนหัว แถวหัวรายงาน แถวท้ายรายงาน และแถวข้อมูล

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Name | string | name | name | ชื่อรายงาน |
| Header | string | headers | headers | ส่วนหัวรายงาน |
| HeaderRows | `[]`[[ReportRowModel]] | headerrows | headerrows | แถวส่วนหัวรายงาน |
| FooterRows | `[]`[[ReportRowModel]] | footerrows | footerrows | แถวส่วนท้ายรายงาน |
| DataRows | `[]`[[ReportDataModel]] | datarows | datarows | แถวข้อมูลรายงาน |

## ความสัมพันธ์

- ฝัง [[ReportRowModel]] (HeaderRows, FooterRows) ซึ่งฝัง [[ReportColumnModel]] อีกชั้น
- ฝัง [[ReportDataModel]] (DataRows) ซึ่งฝัง [[ReportDataColumnModel]] อีกชั้น
