---
name: restaurant-pos
description: Restaurant POS system — tables, kitchen, printers, shifts
user-invocable: true
---

# Restaurant & POS (Point of Sale)

## Overview
Full-featured restaurant management system: tables, kitchen, staff, printers, POS

## System Structure

### Zone & Table
| Section | File | Purpose |
|---------|------|---------|
| Backend Zone | `backend/internal/restaurant/zone/` | Manage zones (Floor 1, Floor 2, Terrace) |
| Backend Table | `backend/internal/restaurant/table/` | Manage tables (table number, seat count, status) |
| BLoC | `bloc/zone/` | Zone state management |
| Screen | `screens/config/` | Table/zone configuration |
| Design | `backend/internal/shopdesign/zonedesign/` | Restaurant floor plan design |

### Kitchen
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/restaurant/kitchen/` | Kitchen management |
| BLoC | `bloc/kitchen/` | Kitchen state management |
| Printer | `backend/internal/restaurant/printer/` | Kitchen printer |
| BLoC Printer | `bloc/kitchen_printer/` | Kitchen printer state management |

### Staff
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/restaurant/staff/` | Staff management |
| Settings | `backend/internal/restaurant/settings/` | Restaurant settings |

### Device & Printer
| Section | File | Purpose |
|---------|------|---------|
| Backend Device | `backend/internal/restaurant/device/` | POS device pairing |
| Backend Notifier | `backend/internal/restaurant/notifier/` | Notifications |
| BLoC Device | `bloc/devices/` | Device state management |
| BLoC Printer | `bloc/deviceprint/` | Printer state management |

### POS (Point of Sale)
| Section | File | Purpose |
|---------|------|---------|
| Backend Shift | `backend/internal/pos/shift/` | Open/close shift |
| Backend Setting | `backend/internal/pos/setting/` | POS settings |
| BLoC POS | `bloc/pos_setting/` | POS state management |
| Cash Drawer | `screens/enhanced_cash_drawer_screen.dart` | Cash drawer |

## Order Flow
```
Customer sits at table -> Staff takes order (POS/mobile)
                          |
                  Send to kitchen (Kitchen Printer)
                          |
                  Kitchen completes -> Notify staff (Notifier)
                          |
                  Serve food -> Check bill -> Payment
                          |
                  Close table -> Ready for next customer
```

## Table Status
| Status | Meaning |
|--------|---------|
| Available | Ready for customers |
| Occupied | Currently in use |
| Awaiting Bill | Customer requested the bill |
| Cleaning | Table is being cleared |

## Shift
- Open shift: Count opening cash drawer balance
- During shift: Sell, receive payment, give change
- Close shift: Summarize totals, count closing cash drawer balance, verify

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `zone_bloc` | Restaurant zones |
| `table_bloc` | Tables |
| `kitchen_bloc` | Kitchen |
| `kitchen_printer_bloc` | Kitchen printer |
| `printer_bloc` | Receipt printer |
| `deviceprint_bloc` | Printer pairing |
| `devices_bloc` | POS devices |
| `pos_setting_bloc` | POS settings |
| `pos_media_bloc` | POS media |
| `staff_bloc` | Staff |
| `cash_in_drawer_bloc` | Cash drawer |

### Repositories
`zone_repository`, `table_repository`, `kitchen_repository`, `printer_repository`, `pos_setting_repository`, `pos_media_repository`, `devices_repository`

### Models
`zone_data_model`, `table_model`, `kitchen_model`, `kitchen_printer_model`, `kitchen_product_model`, `printer_model`, `device_printer_model`, `devices_model`, `pos_setting_model`, `pos_media_model`, `shift_detail_model`, `cash_in_drawer_model`

## Important Notes
- The restaurant module is separate from general transaction modules
- Notifier supports sending alerts to specific devices (`notifierdevice`)
- Floor plan design uses `zonedesign` (drag-and-drop tables)
- Kitchen printer and receipt printer are separate devices
- POS has temp data (`backend/internal/pos/temp/`) for incomplete bills
