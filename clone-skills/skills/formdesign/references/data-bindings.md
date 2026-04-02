# Data Binding Fields Reference

64 predefined fields across 8 categories, defined in `models/data_binding.dart`.

## Document (10 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| doc_number | Document Number | string | INV-2024-001 |
| doc_date | Document Date | date | 2024-01-15 |
| doc_date_be | Date (Buddhist Era) | date | 15/01/2567 |
| doc_type_name | Document Type | string | Tax Invoice |
| doc_ref | Reference Document | string | QT-2024-001 |
| doc_remark | Remark | string | - |
| credit_days | Credit (Days) | number | 30 |
| due_date | Due Date | date | 2024-02-14 |
| page_number | Page Number | string | 1/1 |
| copy_label | Copy Label | string | Original |

## Company/Branch (10 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| company_name | Company Name | string | Example Co., Ltd. |
| company_tax_id | Tax ID | string | 0-1234-56789-01-2 |
| company_address | Company Address | string | 123 Sukhumvit Rd. |
| company_phone | Phone | string | 02-123-4567 |
| company_fax | Fax | string | 02-123-4568 |
| company_email | Email | string | info@example.com |
| company_branch | Branch | string | Head Office |
| company_branch_code | Branch Code | string | 00000 |
| company_website | Website | string | www.example.com |
| company_logo | Logo | image | - |

## Contact/Customer (8 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| contact_code | Customer Code | string | C001 |
| contact_name | Customer Name | string | John Doe |
| contact_tax_id | Customer Tax ID | string | 1-2345-67890-12-3 |
| contact_address | Customer Address | string | 456 Pahonyothin Rd. |
| contact_phone | Customer Phone | string | 081-234-5678 |
| contact_branch | Customer Branch | string | Head Office |
| contact_branch_code | Customer Branch Code | string | 00000 |
| contact_email | Customer Email | string | customer@example.com |

## Detail Line Items (10 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| line_number | Line No. | number | 1 |
| item_code | Item Code | string | P001 |
| item_barcode | Barcode | string | 8850000000001 |
| item_name | Description | string | Sample Product |
| item_qty | Quantity | number | 10 |
| item_unit | Unit | string | PCS |
| item_price | Unit Price | currency | 100.00 |
| item_discount | Discount | currency | 0.00 |
| item_amount | Amount | currency | 1,000.00 |
| item_remark | Remark | string | - |

## Financial Summary (12 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| subtotal | Subtotal | currency | 10,000.00 |
| discount | Discount | currency | 500.00 |
| amount_before_vat | Amount Before VAT | currency | 9,500.00 |
| vat_amount | VAT 7% | currency | 665.00 |
| grand_total | Grand Total | currency | 10,165.00 |
| total_text | Total (Text) | string | Ten thousand one hundred sixty-five baht |
| withholding_tax | Withholding Tax | currency | 300.00 |
| net_amount | Net Amount | currency | 9,865.00 |
| deposit_amount | Deposit | currency | 5,000.00 |
| original_amount | Original Amount | currency | 10,000.00 |
| adjustment_amount | Adjustment | currency | -500.00 |
| new_total | New Total | currency | 9,500.00 |

## Signature (5 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| prepared_by | Prepared By | string | - |
| approved_by | Approved By | string | - |
| received_by | Received By | string | - |
| delivered_by | Delivered By | string | - |
| signature_date | Signature Date | date | ____/____/____ |

## Payment (5 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| payment_method | Payment Method | string | Bank Transfer |
| bank_name | Bank Name | string | Kasikorn Bank |
| bank_account | Account No. | string | 123-4-56789-0 |
| bank_account_name | Account Name | string | Example Co., Ltd. |
| payment_qr | Payment QR | image | - |

## Delivery (4 fields)
| Key | Label | Type | Sample |
|-----|-------|------|--------|
| delivery_address | Delivery Address | string | 789 Ratchada Rd. |
| delivery_date | Delivery Date | date | 2024-01-20 |
| shipping_method | Shipping Method | string | Kerry Express |
| tracking_number | Tracking Number | string | TH123456789 |
