# Procurement Feature Matrix

Last updated: 2026-03-17 (session 6 — Multi-currency + DocumentTotalWidget + Department fix + Internal Note + Purchasing Group + Tolerance + Expiry + Vendor Auto-fill + Budget Check)

## Overall: BC Account PR = 43 done / 47 target (91%)

## Legend
- **done** = Complete and usable (verified from code)
- **partial** = Partially implemented
- **in-progress** = Currently being worked on
- **pending** = Not started but planned
- **planned** = Roadmap item, no code yet
- **future** = Enterprise feature for later

---

## Level 1: Must Have (7+ competitors have it)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Auto document number | **done** | "PR" prefix + auto-generate |
| 2 | Document date | **done** | CustomDatePicker (BE/CE) |
| 3 | Requester name | **done** | Auto-fill from profileData |
| 4 | Requested delivery date | **done** | CustomDatePicker + validation |
| 5 | Urgency (3 levels) | **done** | 1=Normal, 2=Urgent, 3=Critical |
| 6 | Vendor/Supplier | **done** | Supplier search (optional) |
| 7 | Purpose/Remark | **done** | Description field (required) |
| 8 | Warehouse/Branch | **done** | Warehouse selection |
| 9 | Barcode/Item code | **done** | Barcode search + scan |
| 10 | Product name | **done** | Multi-language |
| 11 | Unit + Quantity | **done** | Unit + qty fields |
| 12 | Unit price | **done** | Price column + _getDataText wired |
| 13 | Total value | **done** | sum_amount column + calculation |
| 14 | Approval workflow | **done** | Reuses PO: submit/approve/reject/withdraw |
| 15 | Document status + Dashboard | **done** | List badge + status dashboard |
| 16 | Payment terms | **done** | creditdays UI + save + load |

## Level 2: Should Have (4-6 competitors)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 17 | Department | **done** | extraFields save + transient load |
| 18 | Mobile approval | **done** | LINE OA LIFF + push notification + in-app |
| 19 | Reference document | **done** | docrefno save + load |
| 20 | Ship-to address | **done** | extraFields save + transient load |
| 21 | Attachments | **done** | AttachmentCountIconButton in AppBar |
| 22 | Line item discount | **done** | Discount column in headerTableDetail |
| 23 | VAT | **done** | VAT type selector (exclude/include/zero/none) + rate |
| 24 | Reorder items | **done** | Reorder column + up/down arrows |

## Level 3: Nice to Have (2-3 competitors)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 25 | Job/Project | **done** | 2-step dialog: project > job, cached |
| 26 | Cost Center | **done** | Searchable datalist dialog, cached |
| 27 | Rejection comment | **done** | Reject dialog + comment field (PO+PR) |
| 28 | Email/LINE notification | **done** | Push approval + resend reminder + email |
| 29 | Purchase history | **done** | Stats (avg/min/max/last), 6-month detail, per-item button |
| 30 | Copy document | **done** | AppBar copy + confirm, deep-copy header+details |
| 31 | Serial/Lot number | **done** | Optional columns, extrajson, click-to-edit dialog |
| 32 | Estimated landed cost | **done** | 3 fields: unit cost, freight, duty via extraFields |
| 33 | Vendor preferences | **done** | Dynamic list: vendor + reason, JSON array |
| 34 | Approval deadline | **done** | Date picker, extraFields |
| 35 | Conversion tracking | **done** | Read-only: status + ref RFQ/PO doc numbers |
| 36a | Internal note | **done** | Not shown in print — for purchasing team |
| 36b | Purchasing group | **done** | Team classification (raw materials/IT/construction) |
| 36c | Over/Under delivery tolerance | **done** | % tolerance for agriculture/raw materials |
| 36d | Expiry tracking | **done** | In detail item extrajson — food/pharma |
| 36e | Vendor price auto-fill | **done** | Last price from purchase history on add |
| 36f | Budget check warning | **done** | Warning banner when over budget (non-blocking) |

## Level 4: Enterprise (Future)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 36 | Budget check | **done** | budgetcode/budgetamount in model |
| 37 | RFQ comparison | **done** | RFQ edit + list + vendor widget + routes |
| 38 | Auto PR from MRP/Reorder | **future** | Requires MRP module |
| 39 | Multi-currency | **done** | DocumentTotalWidget + CurrencyApiService + exchange rate |
| 40 | Vendor price catalog | **future** | Auto-fill price by vendor + product |
| 41 | Purchase agreement/contract | **future** | Blanket PO + contract reference |

---

## PO Integration (Document Conversion)

| Flow | Status | Notes |
|------|--------|-------|
| PR > RFQ | **done** | AppBar button + pre-fill supplier/details |
| PR > PO | **done** | AppBar button + pre-fill supplier/details |
| RFQ > PO | **done** | AppBar button + pre-fill selected vendor |
| Button shown only when approved | **done** | `_poApprovalStatus?.isCompleted == true` |
| Pre-fill docrefno | **done** | Source doc number > docrefno field |
| Copy document | **done** | AppBar icon + confirm + deep-copy details |

---

## Backend Data Pipeline

| Stage | PR (flag=21) | RFQ (flag=22) | PO (flag=6) |
|-------|-------------|---------------|-------------|
| MongoDB Save | **done** | **done** | **done** |
| Kafka Publish | **done** | **done** | **done** |
| PostgreSQL Sync | **done** | **done** | **done** |
| GoAPI Consumer | **done** | **done** | **done** |
| ClickHouse Sync | **done** | **done** | **done** |
| Approval System | **done** | **done** | **done** |

## Backend Model Fields (PR)

```
PurchaseRequisition struct:
  Transaction           (inline)
  RequesterCode         string
  RequesterName         string
  DepartmentCode        string
  DepartmentNames       *[]NameX
  Purpose               string
  BudgetCode            string
  BudgetAmount          float64
  Urgency               int8        // 1=Normal, 2=Urgent, 3=Critical
  RequestedDeliveryDate string
  RefPODocNo            string
  RefRFQDocNo           string
  ConversionStatus      string      // none/partial/full
  CreditDays            int
  JobCode               string
  ShipToAddress         string
  EstimatedUnitCost     string
  EstimatedFreight      string
  EstimatedDuty         string
  PreferredVendor       string      // JSON array
  ApprovalDeadline      string
```

## Competitive Landscape

| Software | Score | Price/Month (THB) | Strength |
|----------|-------|-------------------|----------|
| SAP Business One | 38/41 | 100,000+ | Most complete + custom fields |
| Prosoft WINSpeed | 37/41 | 20,000+ | Full Thai procurement cycle |
| Dynamics 365 | 36/41 | 50,000+ | Mobile approval + spending limit |
| **BC Account** | **43/47** | **790** | **AI LINE + LINE approval + multi-currency** |
| AccCloud | 34/41 | 15,000+ | Site-based workflow + auto PR |
| Formula | 33/41 | 10,000+ | Budget check + cost center |
| Nested | 30/41 | 8,000+ | BOQ reference + approval |
| Treesoft | 26/41 | 5,000+ | - |
| PEAK Account | 20/41 | 1,500+ | PR/IR/GR in contact overview |
| Express/SMEMOVE | 8/41 | 500+ | No dedicated PR |
