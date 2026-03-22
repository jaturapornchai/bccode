// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'inventory_costing_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductCostingConfigModel _$ProductCostingConfigModelFromJson(
  Map<String, dynamic> json,
) => ProductCostingConfigModel(
  itemCode: json['item_code'] as String?,
  costingMethod: json['costing_method'] as String?,
  lotTrackingEnabled: json['lot_tracking_enabled'] as bool?,
  expiryTrackingEnabled: json['expiry_tracking_enabled'] as bool?,
  expiryAlertDays: (json['expiry_alert_days'] as num?)?.toInt(),
  autoBlockExpired: json['auto_block_expired'] as bool?,
  allowNegativeStock: json['allow_negative_stock'] as bool?,
  standardCost: (json['standard_cost'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ProductCostingConfigModelToJson(
  ProductCostingConfigModel instance,
) => <String, dynamic>{
  'item_code': instance.itemCode,
  'costing_method': instance.costingMethod,
  'lot_tracking_enabled': instance.lotTrackingEnabled,
  'expiry_tracking_enabled': instance.expiryTrackingEnabled,
  'expiry_alert_days': instance.expiryAlertDays,
  'auto_block_expired': instance.autoBlockExpired,
  'allow_negative_stock': instance.allowNegativeStock,
  'standard_cost': instance.standardCost,
};

InventoryCostLayerModel _$InventoryCostLayerModelFromJson(
  Map<String, dynamic> json,
) => InventoryCostLayerModel(
  id: (json['id'] as num?)?.toInt(),
  shopId: json['shop_id'] as String?,
  itemCode: json['item_code'] as String?,
  barcode: json['barcode'] as String?,
  whCode: json['wh_code'] as String?,
  locationCode: json['location_code'] as String?,
  layerType: json['layer_type'] as String?,
  refDocType: json['ref_doc_type'] as String?,
  refDocNo: json['ref_doc_no'] as String?,
  originalQty: (json['original_qty'] as num?)?.toDouble(),
  remainingQty: (json['remaining_qty'] as num?)?.toDouble(),
  unitCost: (json['unit_cost'] as num?)?.toDouble(),
  landedCostPerUnit: (json['landed_cost_per_unit'] as num?)?.toDouble(),
  totalUnitCost: (json['total_unit_cost'] as num?)?.toDouble(),
  lotNumber: json['lot_number'] as String?,
  supplierLotNumber: json['supplier_lot_number'] as String?,
  manufacturingDate: json['manufacturing_date'] as String?,
  expiryDate: json['expiry_date'] as String?,
  qualityStatus: json['quality_status'] as String?,
  receivedDate: json['received_date'] as String?,
  createdAt: json['created_at'] as String?,
  updatedAt: json['updated_at'] as String?,
);

Map<String, dynamic> _$InventoryCostLayerModelToJson(
  InventoryCostLayerModel instance,
) => <String, dynamic>{
  'id': instance.id,
  'shop_id': instance.shopId,
  'item_code': instance.itemCode,
  'barcode': instance.barcode,
  'wh_code': instance.whCode,
  'location_code': instance.locationCode,
  'layer_type': instance.layerType,
  'ref_doc_type': instance.refDocType,
  'ref_doc_no': instance.refDocNo,
  'original_qty': instance.originalQty,
  'remaining_qty': instance.remainingQty,
  'unit_cost': instance.unitCost,
  'landed_cost_per_unit': instance.landedCostPerUnit,
  'total_unit_cost': instance.totalUnitCost,
  'lot_number': instance.lotNumber,
  'supplier_lot_number': instance.supplierLotNumber,
  'manufacturing_date': instance.manufacturingDate,
  'expiry_date': instance.expiryDate,
  'quality_status': instance.qualityStatus,
  'received_date': instance.receivedDate,
  'created_at': instance.createdAt,
  'updated_at': instance.updatedAt,
};

InventoryStockBalanceModel _$InventoryStockBalanceModelFromJson(
  Map<String, dynamic> json,
) => InventoryStockBalanceModel(
  id: (json['id'] as num?)?.toInt(),
  shopId: json['shop_id'] as String?,
  itemCode: json['item_code'] as String?,
  barcode: json['barcode'] as String?,
  whCode: json['wh_code'] as String?,
  locationCode: json['location_code'] as String?,
  currentQty: (json['current_qty'] as num?)?.toDouble(),
  currentAvgCost: (json['current_avg_cost'] as num?)?.toDouble(),
  currentTotalValue: (json['current_total_value'] as num?)?.toDouble(),
  lastPurchaseCost: (json['last_purchase_cost'] as num?)?.toDouble(),
  lastPurchaseDate: json['last_purchase_date'] as String?,
);

Map<String, dynamic> _$InventoryStockBalanceModelToJson(
  InventoryStockBalanceModel instance,
) => <String, dynamic>{
  'id': instance.id,
  'shop_id': instance.shopId,
  'item_code': instance.itemCode,
  'barcode': instance.barcode,
  'wh_code': instance.whCode,
  'location_code': instance.locationCode,
  'current_qty': instance.currentQty,
  'current_avg_cost': instance.currentAvgCost,
  'current_total_value': instance.currentTotalValue,
  'last_purchase_cost': instance.lastPurchaseCost,
  'last_purchase_date': instance.lastPurchaseDate,
};

InventoryCostTransactionModel _$InventoryCostTransactionModelFromJson(
  Map<String, dynamic> json,
) => InventoryCostTransactionModel(
  id: (json['id'] as num?)?.toInt(),
  shopId: json['shop_id'] as String?,
  itemCode: json['item_code'] as String?,
  barcode: json['barcode'] as String?,
  whCode: json['wh_code'] as String?,
  locationCode: json['location_code'] as String?,
  transactionType: json['transaction_type'] as String?,
  transFlag: (json['trans_flag'] as num?)?.toInt(),
  refDocType: json['ref_doc_type'] as String?,
  refDocNo: json['ref_doc_no'] as String?,
  qty: (json['qty'] as num?)?.toDouble(),
  unitCost: (json['unit_cost'] as num?)?.toDouble(),
  totalCost: (json['total_cost'] as num?)?.toDouble(),
  lotNumber: json['lot_number'] as String?,
  expiryDate: json['expiry_date'] as String?,
  balanceQty: (json['balance_qty'] as num?)?.toDouble(),
  balanceAvgCost: (json['balance_avg_cost'] as num?)?.toDouble(),
  balanceTotalValue: (json['balance_total_value'] as num?)?.toDouble(),
  costingMethodUsed: json['costing_method_used'] as String?,
  transactionDate: json['transaction_date'] as String?,
  createdBy: json['created_by'] as String?,
  createdAt: json['created_at'] as String?,
);

Map<String, dynamic> _$InventoryCostTransactionModelToJson(
  InventoryCostTransactionModel instance,
) => <String, dynamic>{
  'id': instance.id,
  'shop_id': instance.shopId,
  'item_code': instance.itemCode,
  'barcode': instance.barcode,
  'wh_code': instance.whCode,
  'location_code': instance.locationCode,
  'transaction_type': instance.transactionType,
  'trans_flag': instance.transFlag,
  'ref_doc_type': instance.refDocType,
  'ref_doc_no': instance.refDocNo,
  'qty': instance.qty,
  'unit_cost': instance.unitCost,
  'total_cost': instance.totalCost,
  'lot_number': instance.lotNumber,
  'expiry_date': instance.expiryDate,
  'balance_qty': instance.balanceQty,
  'balance_avg_cost': instance.balanceAvgCost,
  'balance_total_value': instance.balanceTotalValue,
  'costing_method_used': instance.costingMethodUsed,
  'transaction_date': instance.transactionDate,
  'created_by': instance.createdBy,
  'created_at': instance.createdAt,
};

StockValuationModel _$StockValuationModelFromJson(Map<String, dynamic> json) =>
    StockValuationModel(
      itemCode: json['item_code'] as String?,
      whCode: json['wh_code'] as String?,
      currentQty: (json['current_qty'] as num?)?.toDouble(),
      averageCost: (json['average_cost'] as num?)?.toDouble(),
      totalValue: (json['total_value'] as num?)?.toDouble(),
      costingMethod: json['costing_method'] as String?,
    );

Map<String, dynamic> _$StockValuationModelToJson(
  StockValuationModel instance,
) => <String, dynamic>{
  'item_code': instance.itemCode,
  'wh_code': instance.whCode,
  'current_qty': instance.currentQty,
  'average_cost': instance.averageCost,
  'total_value': instance.totalValue,
  'costing_method': instance.costingMethod,
};

InventoryValuationReportModel _$InventoryValuationReportModelFromJson(
  Map<String, dynamic> json,
) => InventoryValuationReportModel(
  shopId: json['shop_id'] as String?,
  asOfDate: json['as_of_date'] as String?,
  items: (json['items'] as List<dynamic>?)
      ?.map(
        (e) => InventoryValuationItemModel.fromJson(e as Map<String, dynamic>),
      )
      .toList(),
  totalValue: (json['total_value'] as num?)?.toDouble(),
);

Map<String, dynamic> _$InventoryValuationReportModelToJson(
  InventoryValuationReportModel instance,
) => <String, dynamic>{
  'shop_id': instance.shopId,
  'as_of_date': instance.asOfDate,
  'items': instance.items.map((e) => e.toJson()).toList(),
  'total_value': instance.totalValue,
};

InventoryValuationItemModel _$InventoryValuationItemModelFromJson(
  Map<String, dynamic> json,
) => InventoryValuationItemModel(
  itemCode: json['item_code'] as String?,
  itemName: json['item_name'] as String?,
  whCode: json['wh_code'] as String?,
  unitCode: json['unit_code'] as String?,
  qty: (json['qty'] as num?)?.toDouble(),
  averageCost: (json['average_cost'] as num?)?.toDouble(),
  totalValue: (json['total_value'] as num?)?.toDouble(),
  costingMethod: json['costing_method'] as String?,
);

Map<String, dynamic> _$InventoryValuationItemModelToJson(
  InventoryValuationItemModel instance,
) => <String, dynamic>{
  'item_code': instance.itemCode,
  'item_name': instance.itemName,
  'wh_code': instance.whCode,
  'unit_code': instance.unitCode,
  'qty': instance.qty,
  'average_cost': instance.averageCost,
  'total_value': instance.totalValue,
  'costing_method': instance.costingMethod,
};

StockCardReportModel _$StockCardReportModelFromJson(
  Map<String, dynamic> json,
) => StockCardReportModel(
  shopId: json['shop_id'] as String?,
  itemCode: json['item_code'] as String?,
  itemName: json['item_name'] as String?,
  fromDate: json['from_date'] as String?,
  toDate: json['to_date'] as String?,
  entries: (json['entries'] as List<dynamic>?)
      ?.map((e) => StockCardEntryModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$StockCardReportModelToJson(
  StockCardReportModel instance,
) => <String, dynamic>{
  'shop_id': instance.shopId,
  'item_code': instance.itemCode,
  'item_name': instance.itemName,
  'from_date': instance.fromDate,
  'to_date': instance.toDate,
  'entries': instance.entries.map((e) => e.toJson()).toList(),
};

StockCardEntryModel _$StockCardEntryModelFromJson(Map<String, dynamic> json) =>
    StockCardEntryModel(
      date: json['date'] as String?,
      docNo: json['doc_no'] as String?,
      transactionType: json['transaction_type'] as String?,
      transFlag: (json['trans_flag'] as num?)?.toInt(),
      qtyIn: (json['qty_in'] as num?)?.toDouble(),
      qtyOut: (json['qty_out'] as num?)?.toDouble(),
      unitCost: (json['unit_cost'] as num?)?.toDouble(),
      totalCost: (json['total_cost'] as num?)?.toDouble(),
      balanceQty: (json['balance_qty'] as num?)?.toDouble(),
      balanceValue: (json['balance_value'] as num?)?.toDouble(),
      averageCost: (json['average_cost'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$StockCardEntryModelToJson(
  StockCardEntryModel instance,
) => <String, dynamic>{
  'date': instance.date,
  'doc_no': instance.docNo,
  'transaction_type': instance.transactionType,
  'trans_flag': instance.transFlag,
  'qty_in': instance.qtyIn,
  'qty_out': instance.qtyOut,
  'unit_cost': instance.unitCost,
  'total_cost': instance.totalCost,
  'balance_qty': instance.balanceQty,
  'balance_value': instance.balanceValue,
  'average_cost': instance.averageCost,
};

CostTransactionResultModel _$CostTransactionResultModelFromJson(
  Map<String, dynamic> json,
) => CostTransactionResultModel(
  transaction: json['transaction'] == null
      ? null
      : InventoryCostTransactionModel.fromJson(
          json['transaction'] as Map<String, dynamic>,
        ),
  costLayer: json['cost_layer'] == null
      ? null
      : InventoryCostLayerModel.fromJson(
          json['cost_layer'] as Map<String, dynamic>,
        ),
  balanceQty: (json['balance_qty'] as num?)?.toDouble(),
  balanceAvgCost: (json['balance_avg_cost'] as num?)?.toDouble(),
  totalValue: (json['total_value'] as num?)?.toDouble(),
);

Map<String, dynamic> _$CostTransactionResultModelToJson(
  CostTransactionResultModel instance,
) => <String, dynamic>{
  'transaction': instance.transaction?.toJson(),
  'cost_layer': instance.costLayer?.toJson(),
  'balance_qty': instance.balanceQty,
  'balance_avg_cost': instance.balanceAvgCost,
  'total_value': instance.totalValue,
};
