---
name: coupon-promotion
description: >
  Coupons, promotions, discounts, and loyalty points. Use when the user mentions discount
  codes, promotional pricing, buy-X-get-Y rules, or points redemption.
user-invocable: true
---

# Coupon & Promotion System

## Overview
Create and manage discount coupons, promotions, and points (loyalty) system.

## When to Use
- User wants to create a discount coupon (fixed amount, percentage, or free product)
- User needs to set a promotion that applies automatically without a code
- User wants to restrict a coupon to specific products, categories, or VIP customers
- User asks about loyalty points: how to accumulate, redeem, or view point history
- User wants to check why a coupon was rejected (validation errors)

## Anti-Patterns
- Do NOT confuse promotions (automatic rules) with coupons (require code entry) — they are separate systems
- Do NOT set a coupon without an expiry date or usage limit — unlimited coupons cause financial risk
- Do NOT apply shop-specific (`shopcoupon`) coupons globally — they are shop-scoped
- Do NOT delete a coupon that has already been used in transactions — mark as inactive instead
- Do NOT bypass `coupon_validator.dart` — all validation must go through it, not manual checks

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

## Related Skills
- `/product` — product-specific coupon conditions (applicable products/categories)
- `/master-data` — customer (debtor) data for customer-specific coupons
- `/stock-inventory` — stock availability affects "free product" coupon fulfillment
