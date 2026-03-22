# Auto Packing System Architecture

## Data Models

### Pick & Pack Document (MongoDB: `transactionPickandpack`)

```go
type Pickandpack struct {
    Transaction              // inline: docno, docdate, details[], etc.
    WhCode         string    // warehouse code
    WhNames        *[]NameX
    LocationCode   string    // stock location
    LocationNames  *[]NameX
    Sendtype       string    // shipping type
    Email          string
    Phone          string
    Address        string
    RefSaleInvoice string    // links to sale invoice docno
    PackStatus     int8      // 0=PENDING, 1=PROCESSING, 2=COMPLETED, 3=CANCELLED
    IsPrint        bool      // has been printed
    PrintAt        *time.Time
    PrintBy        string
    IsConfirm      bool      // has been confirmed
    ConfirmAt      *time.Time
    ConfirmBy      string
}
```

### Auto Packing Cache (PostgreSQL: `productbarcode`)

```go
type ProductBarcodePackingStruct struct {
    UnitName              string
    BarcodeRefUnitStand   float64  // numerator (e.g., 24 for BOX)
    BarcodeRefUnitDivide  float64  // denominator (e.g., 1 for base unit)
}
```

### BOM (MongoDB: `productBarcodeBOMs`)

```go
type ProductBarcodeBOMView struct {
    BarcodeGuidFixed string
    Level            int        // hierarchy depth
    Names            *[]NameX
    ItemUnitCode     string
    Barcode          string
    Condition        bool       // conditional BOM
    DivideValue      float64
    StandValue       float64
    Qty              float64
    ImageURI         string
    BOM              *[]ProductBarcodeBOMView  // recursive children
}
```

### Pick & Pack Device (MongoDB: `pickandpackDevices`)

```go
type PickandpackDevice struct {
    Code          string
    DeviceNumber  string
    Name          string
    DeviceType    int8           // 1=Admin, 2=Warehouse, 3=Display
    ActivePin     string
    Employees     []PPEmployee
    WhCodes       []string
    Locationcodes []string
}
```

### Warehouse Hierarchy (bclms)

```
WarehouseModel
├── code, names[]
├── LocationModel[]
│   ├── code, names[]
│   └── ShelfModel[]
│       ├── code, names[]
│       └── ShelfProduct[]
│           ├── barcode, itemcode
│           ├── qty, minqty, maxqty
│           └── unitcode
```

## Data Flow

### Auto Packing Calculation

```
1. Stock balance request comes in
2. Query productbarcode for all units of each item
3. Calculate unit_ratio = stand / divide
4. Sort by ratio DESC (largest units first)
5. Return IsAutoPacking flag + packing options
6. UI shows packing suggestion to warehouse staff
```

### Pick & Pack Workflow

```
Sale Invoice (SI) created
  → SI appears in "available-saleinvoice" list
  → Warehouse staff creates Pick & Pack document
  → Status: PENDING (0)
  → Staff prints packing list → UpdatePrint() → PROCESSING (1)
  → Staff picks items, scans barcodes
  → Staff confirms → ConfirmPickandpack() → COMPLETED (2)
  OR → CancelPickandpack() → CANCELLED (3)
```

### BOM Data Sync

```
BOM created/updated via HTTP API
  → Save to MongoDB (productBarcodeBOMs)
  → Publish to Kafka topic
  → BOM Consumer picks up message
  → Sync to PostgreSQL (productbarcodeboms)
  → Update SaleInvoiceBOMPrice if applicable
```

## Kafka Topics

| Topic Pattern | Purpose |
|--------------|---------|
| `pickandpack.created` | New pick & pack doc |
| `pickandpack.updated` | Updated pick & pack doc |
| `pickandpack.deleted` | Deleted pick & pack doc |
| `bom.created/updated/deleted` | BOM sync to PG |

## TransFlag Reference

| Flag | Type | Stock Effect |
|------|------|-------------|
| 44 | Sale Invoice | -qty |
| 48 | Sale Return | +qty |
| 56 | Stock Pickup | -qty |
| 58 | Stock Return | +qty |
| 60 | Stock Receive | +qty |
| 66 | Stock Adjust + | +qty |
| 68 | Stock Adjust - | -qty |
| 72 | Stock Transfer | ±qty |

## Frontend BOM Model (Dart)

```dart
class ProductBomModel {
  String guidfixed;
  List<LanguageDataModel> names;
  String barcode;
  String itemunitcode;
  List<LanguageDataModel> itemunitnames;
  bool condition;
  int dividevalue;
  int standvalue;
  double qty;
  String imageuri;
  List<ProductBomModel> bom;  // recursive
}
```
