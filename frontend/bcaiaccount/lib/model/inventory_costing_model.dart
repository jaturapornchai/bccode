import 'package:json_annotation/json_annotation.dart';

part 'inventory_costing_model.g.dart';

// === Costing Method Constants ===

class CostingMethod {
  static const String movingAverage = 'moving_average';
  static const String periodicAverage = 'periodic_average';
  static const String fifo = 'fifo';
  static const String lifo = 'lifo';
  static const String fefo = 'fefo';
  static const String lot = 'lot';
  static const String standard = 'standard';

  static const List<String> all = [
    movingAverage,
    periodicAverage,
    fifo,
    lifo,
    fefo,
    lot,
    standard,
  ];

  static String displayName(String method) {
    switch (method) {
      case movingAverage:
        return 'Moving Average';
      case periodicAverage:
        return 'Periodic Average';
      case fifo:
        return 'FIFO';
      case lifo:
        return 'LIFO';
      case fefo:
        return 'FEFO';
      case lot:
        return 'Lot-Based';
      case standard:
        return 'Standard Cost';
      default:
        return method;
    }
  }
}

// === Product Costing Config ===

@JsonSerializable(explicitToJson: true)
class ProductCostingConfigModel {
  @JsonKey(name: 'item_code')
  String itemCode;
  @JsonKey(name: 'costing_method')
  String costingMethod;
  @JsonKey(name: 'lot_tracking_enabled')
  bool lotTrackingEnabled;
  @JsonKey(name: 'expiry_tracking_enabled')
  bool expiryTrackingEnabled;
  @JsonKey(name: 'expiry_alert_days')
  int expiryAlertDays;
  @JsonKey(name: 'auto_block_expired')
  bool autoBlockExpired;
  @JsonKey(name: 'allow_negative_stock')
  bool allowNegativeStock;
  @JsonKey(name: 'standard_cost')
  double standardCost;

  ProductCostingConfigModel({
    String? itemCode,
    String? costingMethod,
    bool? lotTrackingEnabled,
    bool? expiryTrackingEnabled,
    int? expiryAlertDays,
    bool? autoBlockExpired,
    bool? allowNegativeStock,
    double? standardCost,
  })  : itemCode = itemCode ?? '',
        costingMethod = costingMethod ?? CostingMethod.movingAverage,
        lotTrackingEnabled = lotTrackingEnabled ?? false,
        expiryTrackingEnabled = expiryTrackingEnabled ?? false,
        expiryAlertDays = expiryAlertDays ?? 30,
        autoBlockExpired = autoBlockExpired ?? false,
        allowNegativeStock = allowNegativeStock ?? false,
        standardCost = standardCost ?? 0;

  factory ProductCostingConfigModel.fromJson(Map<String, dynamic> json) =>
      _$ProductCostingConfigModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductCostingConfigModelToJson(this);
}

// === Cost Layer ===

@JsonSerializable(explicitToJson: true)
class InventoryCostLayerModel {
  int id;
  @JsonKey(name: 'shop_id')
  String shopId;
  @JsonKey(name: 'item_code')
  String itemCode;
  String barcode;
  @JsonKey(name: 'wh_code')
  String whCode;
  @JsonKey(name: 'location_code')
  String locationCode;
  @JsonKey(name: 'layer_type')
  String layerType;
  @JsonKey(name: 'ref_doc_type')
  String refDocType;
  @JsonKey(name: 'ref_doc_no')
  String refDocNo;
  @JsonKey(name: 'original_qty')
  double originalQty;
  @JsonKey(name: 'remaining_qty')
  double remainingQty;
  @JsonKey(name: 'unit_cost')
  double unitCost;
  @JsonKey(name: 'landed_cost_per_unit')
  double landedCostPerUnit;
  @JsonKey(name: 'total_unit_cost')
  double totalUnitCost;
  @JsonKey(name: 'lot_number')
  String lotNumber;
  @JsonKey(name: 'supplier_lot_number')
  String supplierLotNumber;
  @JsonKey(name: 'manufacturing_date')
  String? manufacturingDate;
  @JsonKey(name: 'expiry_date')
  String? expiryDate;
  @JsonKey(name: 'quality_status')
  String qualityStatus;
  @JsonKey(name: 'received_date')
  String receivedDate;
  @JsonKey(name: 'created_at')
  String createdAt;
  @JsonKey(name: 'updated_at')
  String updatedAt;

  InventoryCostLayerModel({
    int? id,
    String? shopId,
    String? itemCode,
    String? barcode,
    String? whCode,
    String? locationCode,
    String? layerType,
    String? refDocType,
    String? refDocNo,
    double? originalQty,
    double? remainingQty,
    double? unitCost,
    double? landedCostPerUnit,
    double? totalUnitCost,
    String? lotNumber,
    String? supplierLotNumber,
    this.manufacturingDate,
    this.expiryDate,
    String? qualityStatus,
    String? receivedDate,
    String? createdAt,
    String? updatedAt,
  })  : id = id ?? 0,
        shopId = shopId ?? '',
        itemCode = itemCode ?? '',
        barcode = barcode ?? '',
        whCode = whCode ?? '',
        locationCode = locationCode ?? '',
        layerType = layerType ?? '',
        refDocType = refDocType ?? '',
        refDocNo = refDocNo ?? '',
        originalQty = originalQty ?? 0,
        remainingQty = remainingQty ?? 0,
        unitCost = unitCost ?? 0,
        landedCostPerUnit = landedCostPerUnit ?? 0,
        totalUnitCost = totalUnitCost ?? 0,
        lotNumber = lotNumber ?? '',
        supplierLotNumber = supplierLotNumber ?? '',
        qualityStatus = qualityStatus ?? '',
        receivedDate = receivedDate ?? '',
        createdAt = createdAt ?? '',
        updatedAt = updatedAt ?? '';

  factory InventoryCostLayerModel.fromJson(Map<String, dynamic> json) =>
      _$InventoryCostLayerModelFromJson(json);

  Map<String, dynamic> toJson() => _$InventoryCostLayerModelToJson(this);
}

// === Stock Balance ===

@JsonSerializable(explicitToJson: true)
class InventoryStockBalanceModel {
  int id;
  @JsonKey(name: 'shop_id')
  String shopId;
  @JsonKey(name: 'item_code')
  String itemCode;
  String barcode;
  @JsonKey(name: 'wh_code')
  String whCode;
  @JsonKey(name: 'location_code')
  String locationCode;
  @JsonKey(name: 'current_qty')
  double currentQty;
  @JsonKey(name: 'current_avg_cost')
  double currentAvgCost;
  @JsonKey(name: 'current_total_value')
  double currentTotalValue;
  @JsonKey(name: 'last_purchase_cost')
  double lastPurchaseCost;
  @JsonKey(name: 'last_purchase_date')
  String? lastPurchaseDate;

  InventoryStockBalanceModel({
    int? id,
    String? shopId,
    String? itemCode,
    String? barcode,
    String? whCode,
    String? locationCode,
    double? currentQty,
    double? currentAvgCost,
    double? currentTotalValue,
    double? lastPurchaseCost,
    this.lastPurchaseDate,
  })  : id = id ?? 0,
        shopId = shopId ?? '',
        itemCode = itemCode ?? '',
        barcode = barcode ?? '',
        whCode = whCode ?? '',
        locationCode = locationCode ?? '',
        currentQty = currentQty ?? 0,
        currentAvgCost = currentAvgCost ?? 0,
        currentTotalValue = currentTotalValue ?? 0,
        lastPurchaseCost = lastPurchaseCost ?? 0;

  factory InventoryStockBalanceModel.fromJson(Map<String, dynamic> json) =>
      _$InventoryStockBalanceModelFromJson(json);

  Map<String, dynamic> toJson() => _$InventoryStockBalanceModelToJson(this);
}

// === Cost Transaction ===

@JsonSerializable(explicitToJson: true)
class InventoryCostTransactionModel {
  int id;
  @JsonKey(name: 'shop_id')
  String shopId;
  @JsonKey(name: 'item_code')
  String itemCode;
  String barcode;
  @JsonKey(name: 'wh_code')
  String whCode;
  @JsonKey(name: 'location_code')
  String locationCode;
  @JsonKey(name: 'transaction_type')
  String transactionType;
  @JsonKey(name: 'trans_flag')
  int transFlag;
  @JsonKey(name: 'ref_doc_type')
  String refDocType;
  @JsonKey(name: 'ref_doc_no')
  String refDocNo;
  double qty;
  @JsonKey(name: 'unit_cost')
  double unitCost;
  @JsonKey(name: 'total_cost')
  double totalCost;
  @JsonKey(name: 'lot_number')
  String lotNumber;
  @JsonKey(name: 'expiry_date')
  String? expiryDate;
  @JsonKey(name: 'balance_qty')
  double balanceQty;
  @JsonKey(name: 'balance_avg_cost')
  double balanceAvgCost;
  @JsonKey(name: 'balance_total_value')
  double balanceTotalValue;
  @JsonKey(name: 'costing_method_used')
  String costingMethodUsed;
  @JsonKey(name: 'transaction_date')
  String transactionDate;
  @JsonKey(name: 'created_by')
  String createdBy;
  @JsonKey(name: 'created_at')
  String createdAt;

  InventoryCostTransactionModel({
    int? id,
    String? shopId,
    String? itemCode,
    String? barcode,
    String? whCode,
    String? locationCode,
    String? transactionType,
    int? transFlag,
    String? refDocType,
    String? refDocNo,
    double? qty,
    double? unitCost,
    double? totalCost,
    String? lotNumber,
    this.expiryDate,
    double? balanceQty,
    double? balanceAvgCost,
    double? balanceTotalValue,
    String? costingMethodUsed,
    String? transactionDate,
    String? createdBy,
    String? createdAt,
  })  : id = id ?? 0,
        shopId = shopId ?? '',
        itemCode = itemCode ?? '',
        barcode = barcode ?? '',
        whCode = whCode ?? '',
        locationCode = locationCode ?? '',
        transactionType = transactionType ?? '',
        transFlag = transFlag ?? 0,
        refDocType = refDocType ?? '',
        refDocNo = refDocNo ?? '',
        qty = qty ?? 0,
        unitCost = unitCost ?? 0,
        totalCost = totalCost ?? 0,
        lotNumber = lotNumber ?? '',
        balanceQty = balanceQty ?? 0,
        balanceAvgCost = balanceAvgCost ?? 0,
        balanceTotalValue = balanceTotalValue ?? 0,
        costingMethodUsed = costingMethodUsed ?? '',
        transactionDate = transactionDate ?? '',
        createdBy = createdBy ?? '',
        createdAt = createdAt ?? '';

  factory InventoryCostTransactionModel.fromJson(Map<String, dynamic> json) =>
      _$InventoryCostTransactionModelFromJson(json);

  Map<String, dynamic> toJson() => _$InventoryCostTransactionModelToJson(this);
}

// === Stock Valuation ===

@JsonSerializable(explicitToJson: true)
class StockValuationModel {
  @JsonKey(name: 'item_code')
  String itemCode;
  @JsonKey(name: 'wh_code')
  String whCode;
  @JsonKey(name: 'current_qty')
  double currentQty;
  @JsonKey(name: 'average_cost')
  double averageCost;
  @JsonKey(name: 'total_value')
  double totalValue;
  @JsonKey(name: 'costing_method')
  String costingMethod;

  StockValuationModel({
    String? itemCode,
    String? whCode,
    double? currentQty,
    double? averageCost,
    double? totalValue,
    String? costingMethod,
  })  : itemCode = itemCode ?? '',
        whCode = whCode ?? '',
        currentQty = currentQty ?? 0,
        averageCost = averageCost ?? 0,
        totalValue = totalValue ?? 0,
        costingMethod = costingMethod ?? '';

  factory StockValuationModel.fromJson(Map<String, dynamic> json) =>
      _$StockValuationModelFromJson(json);

  Map<String, dynamic> toJson() => _$StockValuationModelToJson(this);
}

// === Inventory Valuation Report ===

@JsonSerializable(explicitToJson: true)
class InventoryValuationReportModel {
  @JsonKey(name: 'shop_id')
  String shopId;
  @JsonKey(name: 'as_of_date')
  String asOfDate;
  List<InventoryValuationItemModel> items;
  @JsonKey(name: 'total_value')
  double totalValue;

  InventoryValuationReportModel({
    String? shopId,
    String? asOfDate,
    List<InventoryValuationItemModel>? items,
    double? totalValue,
  })  : shopId = shopId ?? '',
        asOfDate = asOfDate ?? '',
        items = items ?? [],
        totalValue = totalValue ?? 0;

  factory InventoryValuationReportModel.fromJson(Map<String, dynamic> json) =>
      _$InventoryValuationReportModelFromJson(json);

  Map<String, dynamic> toJson() => _$InventoryValuationReportModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class InventoryValuationItemModel {
  @JsonKey(name: 'item_code')
  String itemCode;
  @JsonKey(name: 'item_name')
  String itemName;
  @JsonKey(name: 'wh_code')
  String whCode;
  @JsonKey(name: 'unit_code')
  String unitCode;
  double qty;
  @JsonKey(name: 'average_cost')
  double averageCost;
  @JsonKey(name: 'total_value')
  double totalValue;
  @JsonKey(name: 'costing_method')
  String costingMethod;

  InventoryValuationItemModel({
    String? itemCode,
    String? itemName,
    String? whCode,
    String? unitCode,
    double? qty,
    double? averageCost,
    double? totalValue,
    String? costingMethod,
  })  : itemCode = itemCode ?? '',
        itemName = itemName ?? '',
        whCode = whCode ?? '',
        unitCode = unitCode ?? '',
        qty = qty ?? 0,
        averageCost = averageCost ?? 0,
        totalValue = totalValue ?? 0,
        costingMethod = costingMethod ?? '';

  factory InventoryValuationItemModel.fromJson(Map<String, dynamic> json) =>
      _$InventoryValuationItemModelFromJson(json);

  Map<String, dynamic> toJson() => _$InventoryValuationItemModelToJson(this);
}

// === Stock Card Report ===

@JsonSerializable(explicitToJson: true)
class StockCardReportModel {
  @JsonKey(name: 'shop_id')
  String shopId;
  @JsonKey(name: 'item_code')
  String itemCode;
  @JsonKey(name: 'item_name')
  String itemName;
  @JsonKey(name: 'from_date')
  String fromDate;
  @JsonKey(name: 'to_date')
  String toDate;
  List<StockCardEntryModel> entries;

  StockCardReportModel({
    String? shopId,
    String? itemCode,
    String? itemName,
    String? fromDate,
    String? toDate,
    List<StockCardEntryModel>? entries,
  })  : shopId = shopId ?? '',
        itemCode = itemCode ?? '',
        itemName = itemName ?? '',
        fromDate = fromDate ?? '',
        toDate = toDate ?? '',
        entries = entries ?? [];

  factory StockCardReportModel.fromJson(Map<String, dynamic> json) =>
      _$StockCardReportModelFromJson(json);

  Map<String, dynamic> toJson() => _$StockCardReportModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class StockCardEntryModel {
  String date;
  @JsonKey(name: 'doc_no')
  String docNo;
  @JsonKey(name: 'transaction_type')
  String transactionType;
  @JsonKey(name: 'trans_flag')
  int transFlag;
  @JsonKey(name: 'qty_in')
  double qtyIn;
  @JsonKey(name: 'qty_out')
  double qtyOut;
  @JsonKey(name: 'unit_cost')
  double unitCost;
  @JsonKey(name: 'total_cost')
  double totalCost;
  @JsonKey(name: 'balance_qty')
  double balanceQty;
  @JsonKey(name: 'balance_value')
  double balanceValue;
  @JsonKey(name: 'average_cost')
  double averageCost;

  StockCardEntryModel({
    String? date,
    String? docNo,
    String? transactionType,
    int? transFlag,
    double? qtyIn,
    double? qtyOut,
    double? unitCost,
    double? totalCost,
    double? balanceQty,
    double? balanceValue,
    double? averageCost,
  })  : date = date ?? '',
        docNo = docNo ?? '',
        transactionType = transactionType ?? '',
        transFlag = transFlag ?? 0,
        qtyIn = qtyIn ?? 0,
        qtyOut = qtyOut ?? 0,
        unitCost = unitCost ?? 0,
        totalCost = totalCost ?? 0,
        balanceQty = balanceQty ?? 0,
        balanceValue = balanceValue ?? 0,
        averageCost = averageCost ?? 0;

  factory StockCardEntryModel.fromJson(Map<String, dynamic> json) =>
      _$StockCardEntryModelFromJson(json);

  Map<String, dynamic> toJson() => _$StockCardEntryModelToJson(this);
}

// === Cost Transaction Result ===

@JsonSerializable(explicitToJson: true)
class CostTransactionResultModel {
  InventoryCostTransactionModel? transaction;
  @JsonKey(name: 'cost_layer')
  InventoryCostLayerModel? costLayer;
  @JsonKey(name: 'balance_qty')
  double balanceQty;
  @JsonKey(name: 'balance_avg_cost')
  double balanceAvgCost;
  @JsonKey(name: 'total_value')
  double totalValue;

  CostTransactionResultModel({
    this.transaction,
    this.costLayer,
    double? balanceQty,
    double? balanceAvgCost,
    double? totalValue,
  })  : balanceQty = balanceQty ?? 0,
        balanceAvgCost = balanceAvgCost ?? 0,
        totalValue = totalValue ?? 0;

  factory CostTransactionResultModel.fromJson(Map<String, dynamic> json) =>
      _$CostTransactionResultModelFromJson(json);

  Map<String, dynamic> toJson() => _$CostTransactionResultModelToJson(this);
}
