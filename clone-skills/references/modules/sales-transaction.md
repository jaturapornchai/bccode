---
name: sales-transaction
description: >
  Sales and purchase document system covering 40+ document types including Sale Invoice,
  Purchase Order, Sale Return, and all transaction flows. Use when the user mentions sales
  or purchase documents, invoices, billing, receiving goods, or document workflow.
user-invocable: true
---

# Sales & Transactions System

## Overview
Manage sales, sale orders, return documents, and all document types (40+ types).

## When to Use
- User says "create purchase order" / "PO for supplier"
- User asks about sale invoice, billing, or issuing a bill to a customer
- User mentions goods receipt
- User wants to do a sale return or purchase return
- User asks how document flow works: PR → PO → Purchase → Payment

## System Structure

### Main Screens
| File | Purpose |
|------|---------|
| `screens/transaction/transaction_edit.dart` | Create/edit all document types |
| `screens/transaction/transaction_paid.dart` | Payment processing |
| `screens/transaction/transaction_stock_balance.dart` | View stock balance |
| `screens/transaction/paidpayment_screen.dart` | Confirm payment |
| `screens/transaction/ai_analysis_screen.dart` | AI document analysis |

### Key Components (19 widgets)
| Widget | Purpose |
|--------|---------|
| `document_header_widget.dart` | Document header (number, date, customer) |
| `document_product_list_widget.dart` | Product list in document |
| `document_total_widget.dart` | Total summary (subtotal, VAT, discount) |
| `document_preview_widget.dart` | Document preview before saving |
| `document_references_widget.dart` | Related document references |
| `doc_flow_widget.dart` | Document workflow display |
| `payment_summary_widget.dart` | Payment summary |
| `payment_method_item_widget.dart` | Payment method selection |
| `coupon_widget.dart` | Coupon usage |
| `credit_terms_section_widget.dart` | Credit terms |
| `wht_section_widget.dart` | Withholding tax |
| `advance_payment_widget.dart` | Advance payment |
| `deposit_payment_widget.dart` | Deposit payment |

## Document Types (40 types)

### Purchase Side
| Type | Backend | Purpose |
|------|---------|---------|
| Purchase Order (PO) | `transaction/purchaseorder/` | Order goods from supplier |
| Purchase Requisition (PR) | `transaction/purchaserequisition/` | Internal purchase request |
| Request for Quotation (RFQ) | `transaction/rfq/` | Request price from supplier |
| Purchase | `transaction/purchase/` | Record purchase |
| Purchase Return | `transaction/purchasereturn/` | Return goods to supplier |
| Partial Purchase | `transaction/purchasepartial/` | Partial goods receipt |
| Purchase Debit Note | `transaction/purchasedebitnote/` | Debit note on purchase side |
| Accrual Receive | `transaction/accrualreceive/` | Record accrued liabilities |

### Sales Side
| Type | Backend | Purpose |
|------|---------|---------|
| Sale Order (SO) | `transaction/saleorder/` | Order goods for sale |
| Sale Invoice | `transaction/saleinvoice/` | Issue sales bill |
| Sale Return | `transaction/saleinvoicereturn/` | Receive return from customer |
| Sale Debit Note | `transaction/saledebitnote/` | Debit note on sales side |
| Quotation | `transaction/quotation/` | Quote price to customer |

## API Endpoints
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/transaction/calculate` | Calculate document totals |
| POST | `/api/transaction/quick-calc` | Quick calculation |
| POST | `/api/transaction/validate-payment` | Validate payment |
| POST | `/api/transaction/purchase-history` | Purchase history |
| POST | `/api/purchase-order/manual-close` | Manually close PO |
| POST | `/genpdf` | Generate document PDF |
| GET | `/genpdf/reprint/:id` | Reprint |
| POST | `/getdoc` | Retrieve document |

## Sales Document Flow
```
Quotation
  ↓ Approve
Sale Order
  ↓ Deliver
Sale Invoice → Payment
  ↓ Customer returns
Sale Return
```

## Purchase Document Flow
```
Purchase Requisition (PR)
  ↓ Approve
Request for Quotation (RFQ) → Compare prices
  ↓ Select supplier
Purchase Order (PO) → Approve → Send to supplier
  ↓ Receive goods
Purchase → Payment
```

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `trans_bloc` | CRUD for all document types |
| `transaction_paidpay_bloc` | Payment processing |
| `purchase_bloc` (in `bloc/purchase/`) | Purchase documents |
| `purchaseorder_bloc` (in `bloc/purchaseorder/`) | Purchase orders |
| `purchaserequisition_bloc` (in `bloc/purchaserequisition/`) | Purchase requisitions |
| `rfq_bloc` (in `bloc/rfq/`) | Request for quotation |
| `quotation_bloc` (in `bloc/quotation/`) | Quotations |

### Repositories
`trans_repository`, `transaction_paidpay_repository`, `purchase_history_repository`, `docref_repository`

### Services
`transaction_calculator_service`, `transaction_permission_service`, `pdf_service`, `approval_api_service`

### Models
`transaction_model`, `doc_format_model`, `doc_payload_model`, `approval_model`

## Important Notes
- `transaction_edit.dart` is a single screen used for **all document types** — differentiated by `transFlag`
- Every transaction must sync: MongoDB → Kafka → PG + ClickHouse
- VAT, discount, and withholding tax calculations use `TransactionCalculator`
- PDF generation uses the `genpdf` handler — colors are fixed (PdfColors) and do not change with theme
- Documents requiring approval must go through the approval system first → see skill `/approval-workflow`

## Anti-Patterns
- ❌ Never hardcode `transFlag` value — always use enum from `/enum-list` skill
- ❌ Never calculate VAT / discount manually — use `TransactionCalculator` service only
- ❌ Never forget to sync MongoDB → Kafka → PG after save — never save to MongoDB alone
- ❌ Never use `showDatePicker` — always use `CustomDatePicker` (Buddhist Era)
- ❌ Never bypass approval workflow for documents configured to require approval

## Related Skills
- `/approval-workflow` — for PO/PR/RFQ/SO that require approval before proceeding
- `/quotation` — quotation documents before converting to Sale Order
- `/payment` — payment processing after issuing a Sale Invoice
- `/creditor-debtor` — supplier/customer data used in documents
- `/enum-list` — for transFlag, document status, VAT type codes
