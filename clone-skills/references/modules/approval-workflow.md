---
name: approval-workflow
description: >
  Multi-level document approval system for PO, PR, RFQ, Quotation, and Sale Order with Email
  and LINE OA notifications. Use when the user mentions approval, document approval workflow,
  pending approval, LINE notification, Email approval link, or reject/withdraw document.
user-invocable: true
---

# Approval Workflow

## Overview
Multi-level document approval system: PO, PR, RFQ, Quotation, Sale Order + notifications via Email/LINE.

## When to Use
- User says "submit purchase order for approval"
- User asks how to configure approval levels or add approvers
- Troubleshooting Email/LINE approval notifications (e.g. "LINE not sending notification")
- User wants to approve or reject a document from an email link or LINE LIFF
- User asks to withdraw a submitted document

## System Structure

### Configuration Screens
| File | Purpose |
|------|---------|
| `screens/config/approval/po_approval_setting_screen.dart` | PO approval settings |
| `screens/config/approval/qt_approval_setting_screen.dart` | Quotation approval settings |
| `screens/config/approval/so_approval_setting_screen.dart` | Sale Order approval settings |
| `screens/config/approval/purchase_type_screen.dart` | Purchase types |

### Document Components
| File | Purpose |
|------|---------|
| `purchaseorder/utils/po_approval_helper.dart` | Check approval permissions |
| `purchaseorder/utils/po_approval_handler.dart` | Execute approval actions |
| `purchaseorder/utils/po_approval_utils.dart` | Utility functions |
| `purchaseorder/components/po_approval_actions_widget.dart` | Approve/reject buttons |
| `purchaseorder/components/po_approval_status_badge.dart` | Status display |

## API Endpoints

### Approval Settings
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/approval/po-settings` | View all PO approval settings |
| POST | `/api/approval/po-setting/save` | Save settings |
| POST | `/api/approval/po-setting/delete` | Delete settings |

### Approval Actions
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/approval/po-status/submit` | Submit document for approval |
| POST | `/api/approval/po-status/approve` | Approve |
| POST | `/api/approval/po-status/reject` | Reject |
| POST | `/api/approval/po-status/withdraw` | Withdraw document |
| POST | `/api/approval/po-status/pending` | List pending approvals |
| POST | `/api/approval/po-status/rejected` | List rejected items |
| POST | `/api/approval/po-status/get` | View status |
| POST | `/api/approval/po-status/batch` | View multiple statuses |

> **Note**: Replace `po` with `pr` or `rfq` — same endpoints apply.

### Notifications
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/approval/notification/send` | Send notification |
| POST | `/api/approval/notification/process` | Process pending notifications |
| POST | `/api/approval/notification/logs` | View notification history |
| POST | `/api/approval/notification/mark-opened` | Mark notification as read |
| POST | `/api/approval/notification/resend` | Resend notification |

### Approval via Email/LINE
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/approval/action` | Approve via link (token) |
| POST | `/api/approval/action` | Approve via form |
| GET | `/api/approval/token-info` | View token information |
| POST | `/api/approval/liff-approve` | Approve via LINE LIFF |
| POST | `/api/approval/po-details` | View PO details for LIFF |
| POST | `/api/approval/test-email` | Test send email |
| POST | `/api/approval/test-line-push` | Test send LINE |
| GET | `/api/approval/smtp-status` | SMTP status |
| GET | `/api/approval/lineoa-config-status` | LINE OA status |

### Timeline
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/approval/timeline` | View approval timeline |

## Approval Flow
```
Create document (Draft)
  | Submit
Pending Approval → Notification via Email/LINE
  |                    |
Approved              Rejected
  |                    |
Proceed              Edit → Resubmit
```

## Document Statuses
| Status | Meaning |
|--------|---------|
| Draft | Draft — not yet submitted for approval |
| Pending | Awaiting approval |
| Approved | Approved |
| Rejected | Rejected |
| Withdrawn | Withdrawn by creator |

## Notifications
- **Email**: Send approval link via email (requires SMTP configuration)
- **LINE OA**: Send via LINE Official Account (requires LINE OA integration)
- **LINE LIFF**: Approval page directly within LINE app
- Can approve from the link directly — **no need to open the app**

## Important Notes
- PR/RFQ/PO share the same approval system — differentiated by prefix (`po-`, `pr-`, `rfq-`)
- Approval reuse pattern from `po_approval_helper.dart` — differentiated by `purchase_type_code`
- Token-based approval has an expiry — expired tokens require resending the link
- Multi-level approval: configurable multiple levels (e.g., Team Lead → Manager → MD)

## Anti-Patterns
- ❌ Never hardcode approval status — use enum from `/enum-list` (Draft, Pending, Approved, Rejected, Withdrawn)
- ❌ Always check token expiry before approving via link — expired tokens must be resent
- ❌ Never bypass approval workflow even in dev mode — every flow must go through submit → approve
- ❌ Never create a new approval handler from scratch — reuse pattern from `po_approval_helper.dart`
- ❌ Always configure SMTP / LINE OA before testing notifications — check status endpoint first

## Related Skills
- `/sales-transaction` — PO, PR, RFQ, SO that require approval
- `/quotation` — Quotation approval settings and workflow
- `/enum-list` — for approval status codes and purchase_type_code
- `/master-data` — for approver user data and notification settings
