---
name: quotation
description: Quotation creation, approval, conversion to sale order — Quotation System
user-invocable: true
---

# Quotation System

## Overview
Create quotations for customers with approval workflow + conversion to sale orders.

## System Structure

### Screens
| File | Purpose |
|------|---------|
| `screens/quotation/quotation_edit_screen.dart` | Create/edit quotation |
| `screens/quotation/quotation_list_screen.dart` | Quotation list |

### Components
| File | Purpose |
|------|---------|
| `qt_appbar_actions.dart` | Top bar action buttons |
| `qt_list_card.dart` | List item card |
| `qt_responsive_layout.dart` | Responsive layout |

### Utils (9 files)
| File | Purpose |
|------|---------|
| `qt_form_controller.dart` | Form control |
| `qt_save_helper.dart` | Save data |
| `qt_validation_utils.dart` | Data validation |
| `qt_workflow_manager.dart` | Workflow management |
| `qt_cart_integration_handler.dart` | Cart system integration |
| `qt_currency_handler.dart` | Currency handling |
| `qt_product_search_handler.dart` | Product search |
| `qt_detail_command_handler.dart` | Detail management |
| `qt_bloc_effect_handler.dart` | BLoC side effects handling |

### Backend
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/transaction/quotation/` | Quotation CRUD |
| BLoC | `bloc/quotation/` | State management |
| Approval | `screens/config/approval/qt_approval_setting_screen.dart` | Approval settings |

## Quotation Flow
```
Create quotation
  |
Add products + prices + conditions
  |
Submit for approval (if configured)
  | Approved
Send to customer
  | Customer accepts
Convert to Sale Order
  |
Convert to Invoice
```

## Quotation Data
| Section | Details |
|---------|---------|
| Document Header | Number, date, customer, address |
| Line Items | Name, quantity, unit price, discount |
| Totals | Subtotal, VAT, bill discount, net total |
| Conditions | Payment terms, delivery method, remarks |
| Currency | Multi-currency support |
| References | Related documents |

## Important Notes
- Quotations have an expiry — configurable expiration date
- Multi-currency support — uses `qt_currency_handler.dart`
- Cart integration — connects with the shopping cart system
- Quotation type configured at `quotation_type_screen.dart`
- Quotation approval uses the same system as PO → see skill `/approval-workflow`
