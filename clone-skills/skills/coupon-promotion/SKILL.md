---
name: coupon-promotion
description: Coupons, promotions, discounts, loyalty points — Coupon & Promotion
user-invocable: true
---

# Coupon & Promotion System

## Overview
Create and manage discount coupons, promotions, and points (loyalty) system.

## System Structure

### Coupon
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/coupon/coupon_screen.dart` | Create/edit/manage coupons |
| BLoC | `bloc/coupon/` | State management |
| Backend | `backend/internal/coupon/` | Main coupon CRUD |
| Shop Coupon | `backend/internal/shopcoupon/` | Shop-specific coupons |

### Coupon Components (7 widgets)
| Widget | Purpose |
|--------|---------|
| `coupon_basic_info_widget.dart` | Name, code, duration |
| `coupon_type_value_widget.dart` | Discount type (fixed amount / percentage) |
| `coupon_conditions_widget.dart` | Usage conditions (minimum amount, usage limit) |
| `coupon_product_condition_widget.dart` | Applicable to specific products |
| `coupon_customer_selection_widget.dart` | Applicable to specific customers |
| `coupon_status_widget.dart` | Enable/disable |
| `coupon_remark_widget.dart` | Remarks |

### Promotion
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/config/promotion_screen.dart` | Manage promotions |
| Backend | `backend/internal/product/promotion/` | Promotion CRUD |

### Points System
| Section | File | Purpose |
|---------|------|---------|
| Settings | `screens/config/point_setting_screen.dart` | Set point accumulation rates |
| History | `screens/config/point_transaction_screen.dart` | View point history |
| BLoC | `bloc/point_transaction/` | State management |

## Coupon Types
| Type | Example |
|------|---------|
| Fixed discount | 100 THB off |
| Percentage discount | 10% off |
| Free product | Buy 2 get 1 free |
| Product-specific | Only for "Beverages" category |
| Customer-specific | VIP members only |

## Using Coupons in Documents
```
Create sales bill → Add products
  |
Apply coupon (coupon_widget.dart)
  |
System validates conditions:
  - Minimum amount met?
  - Products match conditions?
  - Customer matches conditions?
  - Not expired?
  - Usage limit not exceeded?
  |
Apply discount → Calculate total
```

## Important Notes
- Coupon has a validator: `coupon_validator.dart` + error handler: `coupon_error_handler.dart`
- Coupons have 2 levels: `coupon` (global) + `shopcoupon` (shop-specific)
- Coupons in documents use `coupon_widget.dart` in transactions
- Promotions and coupons are separate systems — promotions use automatic rules, coupons require entering a code
