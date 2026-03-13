# Data Binding Fields Reference

64 predefined fields across 8 categories, defined in `models/data_binding.dart`.

## Document (10 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| doc_number | เลขที่เอกสาร | string | INV-2024-001 |
| doc_date | วันที่เอกสาร | date | 2024-01-15 |
| doc_date_be | วันที่ (พ.ศ.) | date | 15/01/2567 |
| doc_type_name | ประเภทเอกสาร | string | ใบกำกับภาษี |
| doc_ref | เอกสารอ้างอิง | string | QT-2024-001 |
| doc_remark | หมายเหตุ | string | - |
| credit_days | เครดิต (วัน) | number | 30 |
| due_date | วันครบกำหนด | date | 2024-02-14 |
| page_number | หน้าที่ | string | 1/1 |
| copy_label | สำเนา | string | ต้นฉบับ |

## Company/Branch (10 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| company_name | ชื่อบริษัท | string | บริษัท ตัวอย่าง จำกัด |
| company_tax_id | เลขประจำตัวผู้เสียภาษี | string | 0-1234-56789-01-2 |
| company_address | ที่อยู่บริษัท | string | 123 ถ.สุขุมวิท |
| company_phone | โทรศัพท์ | string | 02-123-4567 |
| company_fax | แฟกซ์ | string | 02-123-4568 |
| company_email | อีเมล | string | info@example.com |
| company_branch | สาขา | string | สำนักงานใหญ่ |
| company_branch_code | รหัสสาขา | string | 00000 |
| company_website | เว็บไซต์ | string | www.example.com |
| company_logo | โลโก้ | image | - |

## Contact/Customer (8 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| contact_code | รหัสลูกค้า | string | C001 |
| contact_name | ชื่อลูกค้า | string | นายสมชาย ใจดี |
| contact_tax_id | เลขผู้เสียภาษี (ลูกค้า) | string | 1-2345-67890-12-3 |
| contact_address | ที่อยู่ลูกค้า | string | 456 ถ.พหลโยธิน |
| contact_phone | โทรลูกค้า | string | 081-234-5678 |
| contact_branch | สาขาลูกค้า | string | สำนักงานใหญ่ |
| contact_branch_code | รหัสสาขาลูกค้า | string | 00000 |
| contact_email | อีเมลลูกค้า | string | customer@example.com |

## Detail Line Items (10 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| line_number | ลำดับ | number | 1 |
| item_code | รหัสสินค้า | string | P001 |
| item_barcode | บาร์โค้ด | string | 8850000000001 |
| item_name | รายการ | string | สินค้าตัวอย่าง |
| item_qty | จำนวน | number | 10 |
| item_unit | หน่วย | string | ชิ้น |
| item_price | ราคา/หน่วย | currency | 100.00 |
| item_discount | ส่วนลด | currency | 0.00 |
| item_amount | จำนวนเงิน | currency | 1,000.00 |
| item_remark | หมายเหตุ | string | - |

## Financial Summary (12 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| subtotal | รวมเงิน | currency | 10,000.00 |
| discount | ส่วนลด | currency | 500.00 |
| amount_before_vat | มูลค่าก่อน VAT | currency | 9,500.00 |
| vat_amount | ภาษีมูลค่าเพิ่ม 7% | currency | 665.00 |
| grand_total | จำนวนเงินรวมทั้งสิ้น | currency | 10,165.00 |
| total_text | จำนวนเงิน (ตัวอักษร) | string | หนึ่งหมื่นหนึ่งร้อยหกสิบห้าบาทถ้วน |
| withholding_tax | ภาษีหัก ณ ที่จ่าย | currency | 300.00 |
| net_amount | ยอดสุทธิ | currency | 9,865.00 |
| deposit_amount | เงินมัดจำ | currency | 5,000.00 |
| original_amount | ยอดเดิม | currency | 10,000.00 |
| adjustment_amount | ยอดปรับปรุง | currency | -500.00 |
| new_total | ยอดใหม่ | currency | 9,500.00 |

## Signature (5 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| prepared_by | ผู้จัดทำ | string | - |
| approved_by | ผู้อนุมัติ | string | - |
| received_by | ผู้รับ | string | - |
| delivered_by | ผู้ส่ง | string | - |
| signature_date | วันที่ลงนาม | date | ____/____/____ |

## Payment (5 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| payment_method | วิธีชำระ | string | โอนเงิน |
| bank_name | ธนาคาร | string | กสิกรไทย |
| bank_account | เลขบัญชี | string | 123-4-56789-0 |
| bank_account_name | ชื่อบัญชี | string | บจก.ตัวอย่าง |
| payment_qr | QR ชำระเงิน | image | - |

## Delivery (4 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| delivery_address | ที่อยู่จัดส่ง | string | 789 ถ.รัชดา |
| delivery_date | วันที่จัดส่ง | date | 2024-01-20 |
| shipping_method | วิธีจัดส่ง | string | Kerry Express |
| tracking_number | เลขพัสดุ | string | TH123456789 |
