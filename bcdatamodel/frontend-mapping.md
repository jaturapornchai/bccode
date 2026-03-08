# Frontend ↔ Backend Mapping

การเชื่อมต่อระหว่าง Flutter Frontend (bcaiaccount) และ Go Backend

## MCP Configuration

Frontend เชื่อมต่อ backend ผ่าน MCP (Model Context Protocol):

**File:** `D:\bcdev\bcaiaccount\.mcp.json`
```json
{
  "mcpServers": {
    "bcai-backend": {
      "type": "sse",
      "url": "http://localhost:8888/goapi/mcp/sse",
      "headers": { "X-API-Key": "..." }
    }
  }
}
```

**กฏ:** Frontend AI ห้ามอ่าน backend code โดยตรง — ต้องผ่าน MCP เท่านั้น

---

## Multi-Language Field Mapping

| Backend (Go) | Frontend (Dart) | JSON | ตัวอย่าง |
|-------------|----------------|------|----------|
| `[]models.NameX` | `List<LanguageDataModel>` | `[{"code":"th","name":"..."}]` | names, custnames |
| `*[]models.NameX` | `List<LanguageDataModel>` | เหมือนกัน (nullable) | itemnames, whnames |

**LanguageDataModel (Dart):**
```dart
class LanguageDataModel {
  String code;  // "th", "en", "vi", ...
  String name;  // ชื่อในภาษานั้น
}
```

---

## Transaction Model Mapping

### Header Fields

| Backend Field (Go) | Frontend Field (Dart) | JSON Key | หมายเหตุ |
|--------------------|-----------------------|----------|----------|
| `DocNo` | `docno` | `docno` | ตรงกัน |
| `DocDatetime` | `docdatetime` | `docdatetime` | Go=time.Time, Dart=String |
| `TransFlag` | `transflag` | `transflag` | ตรงกัน |
| `CustCode` | `custcode` | `custcode` | ตรงกัน |
| `CustNames` | `custnames` | `custnames` | NameX ↔ LanguageDataModel |
| `TotalAmount` | `totalamount` | `totalamount` | ตรงกัน |
| `TotalDiscount` | `totaldiscount` | `totaldiscount` | ตรงกัน |
| `VatType` | `vattype` | `vattype` | ตรงกัน |
| `IsCancel` | `iscancel` | `iscancel` | ตรงกัน |
| `PaymentDetail` | `paymentdetail` | `paymentdetail` | PaymentDetail ↔ TransactionPayModel |
| `Branch` | — | `branch` | TransactionBranch struct |
| `CreatorCode` | `creator_code` | `creator_code` | **BSON≠JSON:** bson=creatorcode |
| `CreatorName` | `creator_name` | `creator_name` | **BSON≠JSON:** bson=creatorname |
| `CreatedAt` | `created_at` | `created_at` | **BSON≠JSON:** bson=createdat |
| `UpdaterCode` | `modifier_code` | `modifier_code` | **BSON≠JSON:** bson=updatercode |
| `UpdaterName` | `modifier_name` | `modifier_name` | **BSON≠JSON:** bson=updatername |
| `UpdatedAt` | `modified_at` | `modified_at` | **BSON≠JSON:** bson=updatedat |

> **สำคัญ:** Audit fields ใช้ **snake_case** ใน JSON/Dart (`creator_code`) แต่ **camelCase** ใน BSON/MongoDB (`creatorcode`)

### Multi-Currency Fields

| Backend (Go) | Frontend (Dart) | JSON Key |
|-------------|----------------|----------|
| `Currency` | `currency` | `currency` |
| `CurrencySymbol` | `currencysymbol` | `currencysymbol` |
| `DocCurrency` | `doc_currency` | `doc_currency` |
| `DocCurrencySymbol` | `doc_currencysymbol` | `doc_currencysymbol` |
| `ExchangeRate` | `exchangerate` | `exchangerate` |
| `TotalAmountDoc` | `totalamount_doc` | `totalamount_doc` |

### Detail Fields

| Backend Field (Go) | Frontend Field (Dart) | JSON Key |
|--------------------|-----------------------|----------|
| `Barcode` | `barcode` | `barcode` |
| `ItemCode` | `itemcode` | `itemcode` |
| `ItemNames` | `itemnames` | `itemnames` |
| `Qty` | `qty` | `qty` |
| `Price` | `price` | `price` |
| `SumAmount` | `sumamount` | `sumamount` |
| `Discount` | `discount` | `discount` |
| `DiscountAmount` | `discountamount` | `discountamount` |
| `CalcFlag` | `calcflag` | `calcflag` |
| `WhCode` | `whcode` | `whcode` |
| `LocationCode` | `locationcode` | `locationcode` |
| `StandValue` | `stand` (in unit) | `standvalue` |
| `DivideValue` | `divider` (in unit) | `dividevalue` |
| `PriceDoc` | `price_doc` | `price_doc` |
| `SumAmountDoc` | `sumamount_doc` | `sumamount_doc` |

### Payment Mapping

| Backend (Go) | Frontend (Dart) | JSON Key |
|-------------|----------------|----------|
| `PaymentDetail.CashAmount` | `paymentdetail.cashamount` | `cashamount` |
| `PaymentDetail.PaymentCreditCards` | `paymentdetail.paymentcreditcards` | `paymentcreditcards` |
| `PaymentDetail.PaymentTransfers` | `paymentdetail.paymenttransfers` | `paymenttransfers` |
| — | `paymentStructs` (BillPayStruct list) | `paymentStructs` |

> Frontend มี `BillPayStruct` เพิ่มเติมที่ Backend ไม่มี — ใช้สำหรับ UI ฝั่ง frontend เท่านั้น

---

## Customer/Party Mapping

| Backend (Go) | Frontend (Dart) | หมายเหตุ |
|-------------|----------------|----------|
| `Debtor` struct | `CustomerModel` | Frontend ใช้ model เดียว |
| `Creditor` struct | `CustomerModel` | ดูที่ `iscreditor`/`isdebtor` |
| `Debtor.AddressForBilling` | `CustomerAddressModel` | ตรงกัน |
| `Debtor.AddressForShipping` | `List<CustomerAddressModel>` | array |

---

## Product Mapping

| Backend (Go) | Frontend (Dart) | หมายเหตุ |
|-------------|----------------|----------|
| `Product` struct | `ProductMasterModel` | Master product |
| `ProductBarcode` struct | `ProductModel` | Product + barcode |
| `Product.Barcodes[].Prices` | `ProductUnitModel` | Unit ↔ Barcode |
| `StandValue` | `stand` | ชื่อต่าง JSON เดียวกัน |
| `DivideValue` | `divider` | ชื่อต่าง JSON เดียวกัน |

---

## Branch/Organization Mapping

| Backend (Go) | Frontend (Dart) | หมายเหตุ |
|-------------|----------------|----------|
| Branch struct | `CompanyBranchModel` | Branch + company info |
| Branch struct (simple) | `BranchModel` | Code + names |
| BranchContact | `ContactModel` | ข้อมูลติดต่อ |

---

## API Endpoints

### MainAPI Routes (port 8888)

| Method | Path | หมายเหตุ |
|--------|------|----------|
| GET | `/transaction/sale-invoice/list` | รายการใบขาย |
| GET | `/transaction/sale-return/list` | รายการรับคืน |
| GET | `/transaction/purchase/list` | รายการใบซื้อ |
| GET | `/transaction/purchase-return/list` | รายการส่งคืน |
| GET | `/transaction/stock-pickup/list` | รายการเบิก |
| GET | `/transaction/stock-return/list` | รายการรับคืนเบิก |
| GET | `/transaction/stock-receive/list` | รายการรับ |
| GET | `/transaction/transfer/list` | รายการโอน |
| GET | `/transaction/adjust/list` | รายการปรับยอด |
| POST | `/transaction/{type}` | สร้างเอกสาร |
| PUT | `/transaction/{type}/{guid}` | แก้ไขเอกสาร |
| DELETE | `/transaction/{type}/{guid}` | ลบเอกสาร |

### GoAPI Routes (port 8888, prefix /goapi)

| Method | Path | หมายเหตุ |
|--------|------|----------|
| GET | `/goapi/version` | เวอร์ชัน |
| GET | `/goapi/api/health` | Health check |
| POST | `/goapi/api/report/sales/by-document` | Report ยอดขายตามเอกสาร |
| GET | `/goapi/mcp/sse` | MCP SSE endpoint |
| GET | `/goapi/mcp/health` | MCP health |
| GET | `/goapi/mcp/tools` | MCP tools list |

### MCP Tools (35 tools)

AI เข้าถึงผ่าน MCP protocol — ดูรายละเอียดที่ backend `internal/goapi/mcp/tools/`

| Tool Category | File | Tools |
|--------------|------|-------|
| Sales | sales.go | GetSalesSummary, GetSalesByProduct, GetSalesByCustomer, GetSalesTimeSeries, GetHourlySalesPattern, GetSalesByChannel |
| Dashboard | dashboard.go | GetDashboardKPI |
| Financial | financial.go | GetProfitAnalysis, GetAccountsReceivable, GetAccountsPayable, GetCashFlow |
| Inventory | inventory.go | GetInventoryValue, GetInventoryTurnover, GetLowStockAlerts |
| Customers | customers.go | GetCustomerSegmentation, GetCustomerGrowth, GetTopCustomers |
| Products | products.go | GetProductPerformance, GetPriceAnalysis |
| Comparison | comparison.go | GetYoYComparison, GetMoMComparison |
| Database | database.go | QueryDatabase |
| ClickHouse | clickhouse_query.go | QueryClickHouse |
| Model Schema | model_schema.go | GetModelSchema |
| Enum Catalog | enum_catalog.go | GetEnumCatalog |
| API Catalog | api_catalog.go | GetAPICatalog |
| API Spec | api_spec.go | GetAPISpec |
| Token Export | token_export.go | ExportShopTokens |

---

## Data Type Mapping

| Go Type | Dart Type | JSON | หมายเหตุ |
|---------|-----------|------|----------|
| string | String | string | — |
| int / int8 / int16 | int | number | — |
| float64 | double | number | — |
| bool | bool | boolean | — |
| time.Time | String | string | Go=RFC3339, Dart=ISO8601 string |
| []NameX | List\<LanguageDataModel\> | array | Multi-language |
| *[]string | List\<String\> | array/null | Nullable array |
| interface{} | dynamic | any | Flexible type |

---

## Naming Conventions

| Context | Style | ตัวอย่าง |
|---------|-------|----------|
| Go struct fields | PascalCase | `TotalAmount`, `DocNo` |
| JSON keys | lowercase/snake_case | `totalamount`, `creator_code` |
| BSON keys | lowercase | `totalamount`, `creatorcode` |
| Dart fields | camelCase | `totalAmount`, `creatorCode` |
| ClickHouse columns | lowercase | `totalamount`, `docdatetime` |
| PostgreSQL columns | lowercase | `totalamount`, `docdatetime` |

> **ข้อยกเว้น:** Audit fields (creator/modifier) ใช้ snake_case ใน JSON (`creator_code`) แต่ camelCase ใน BSON (`creatorcode`) เพื่อ backward compatibility
