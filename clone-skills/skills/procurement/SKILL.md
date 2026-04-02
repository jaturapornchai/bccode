---
name: procurement
description: >
  Procurement module — PR (Purchase Requisition), RFQ, PO feature tracker and architecture.
  Triggers: "procurement", "purchase requisition", "PR feature", "RFQ", "PO feature".
---

# Procurement Module

## Update Rule
After any PR/RFQ/PO code change, update `references/feature-matrix.md` immediately.

## Document Flow
```
PR (Purchase Requisition) -> [Approve] -> RFQ (Request for Quotation) -> [Approve] -> PO (Purchase Order)
     Department request              3+ vendor quotes                  Best vendor selected
     transflag=21                    transflag=22                      transflag=6
```

## Key Files

| Area | Path |
|------|------|
| PR Backend Model | `backend/internal/transaction/purchaserequisition/models/purchaserequisition.go` |
| PR Edit Screen | `frontend/bcaiaccount/lib/screens/purchaserequisition/purchaserequisition_edit_screen.dart` |
| PR Header Widget | `frontend/bcaiaccount/lib/screens/purchaserequisition/components/pr_header_widget.dart` |
| PR Form Controller | `frontend/bcaiaccount/lib/screens/purchaserequisition/utils/pr_form_controller.dart` |
| PR Repository | `frontend/bcaiaccount/lib/repositories/trans_repository.dart` > `savePurchaseRequisition()` |
| PR List Screen | `frontend/bcaiaccount/lib/screens/purchaserequisition/purchaserequisition_list_screen.dart` |
| PO Edit Screen | `frontend/bcaiaccount/lib/screens/purchaseorder/purchaseorder_edit_screen.dart` |
| RFQ Backend Model | `backend/internal/transaction/rfq/models/rfq.go` |

## Architecture Patterns

### ExtraFields Pattern
PR/RFQ have fields not in `TransactionModel` (departmentcode, jobcode, shiptoaddress). These go through:
```
PRFormController > _buildPRExtraFields() > TransSave(extraFields: {...})
  > TransBloc > savePurchaseRequisition(trans, extraFields) > Repository: data.addAll(extraFields)
```
Fields in TransactionModel (creditdays, docrefno) set directly on screenData.

### Transient Fields Pattern
PR-specific fields returned by GoAPI map to transient fields on TransactionModel (not serialized in .g.dart). Example: `TransactionModel.prDepartmentCode <- data['departmentcode']`.

### Department Data Source
Fetch from `global.companyBranchSelectData.departments`, not from API (`/organization/department/list` returns empty). Same for Job Project and Cost Center.

### Vendor Price Auto-fill
On adding product to PR, fetch lastPrice from PurchaseHistoryRepository (12 months). Hook: `DocumentProductListWidget.onDetailAdded`. Only fills when `detail.price == 0`.

### Budget Check
Budget fields in PR header. Warning banner when totalAmount > budgetAmount. Does not block save — approver decides.

### PR Layout (2 Tabs)
Tab 1: Document header + summary (scrollable together)
Tab 2: Product list + detail items

## Competitive Position
BC Account ranks 2nd/10 (43/47 = 91%) in PR features. Price 790 THB/month vs competitors at 10,000-100,000+.

Unique features: AI Chatbot PR via LINE (MCP-based), Approval via LINE OA.

## Workflow: Add New Feature
1. Check `references/feature-matrix.md` for priority level
2. Check backend model for field support
3. Check frontend TransactionModel field mapping
4. Implement (backend + frontend)
5. Update feature-matrix.md status to done

## References
- [Feature Matrix](references/feature-matrix.md) — full feature list, status, competitive analysis
