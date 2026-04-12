---
name: quotation
description: >
  Quotation system for creating price quotes for customers, managing approval workflow, and
  converting approved quotations to sale orders. Use when the user mentions quotations, price
  quotes, customer quotes, multi-currency quotes, or convert quotation to sale order.
user-invocable: true
---

# Quotation System

## Overview
Create quotations for customers with approval workflow + conversion to sale orders.

## When to Use
- User says "create quotation for customer ABC"
- User wants to add products or change pricing on a quotation
- User asks about quotation expiry issues
- User needs to approve a quotation or configure quotation approval levels
- User says "convert quotation to sale order"

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

## Anti-Patterns
- ❌ Never hardcode quotation status — use enum from `/enum-list` skill
- ❌ Never forget to validate currency rate before calculating multi-currency quotation
- ❌ Never use `showDatePicker` — use `CustomDatePicker` for expiry date
- ❌ Never convert quotation to SO without going through approval workflow (if configured)
- ❌ Never forget to set quotation expiry date — a quotation without an expiry date may cause issues

## Related Skills
- `/approval-workflow` — for quotation approval levels and notifications
- `/sales-transaction` — after converting QT → SO → Invoice
- `/creditor-debtor` — customer data used in quotations
- `/master-data` — for product lookup and pricing
- `/enum-list` — for quotation status codes and quotation type
