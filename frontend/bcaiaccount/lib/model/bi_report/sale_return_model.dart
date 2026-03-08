import 'package:json_annotation/json_annotation.dart';

part 'sale_return_model.g.dart';

@JsonSerializable(explicitToJson: true)
class SaleReturnModel {
  @JsonKey(name: 'doc_date')
  final String docDate;
  @JsonKey(name: 'docno')
  final String docno;
  @JsonKey(name: 'creditornames')
  final List<NameModel> creditorNames;
  @JsonKey(name: 'totalvalue')
  final double totalValue;
  @JsonKey(name: 'detailtotaldiscount')
  final double detailTotalDiscount;
  @JsonKey(name: 'totalafterdiscount')
  final double totalAfterDiscount;
  @JsonKey(name: 'totalexceptvat')
  final double totalExceptVat;
  @JsonKey(name: 'totalbeforevat')
  final double totalBeforeVat;
  @JsonKey(name: 'totalvatvalue')
  final double totalVatValue;
  @JsonKey(name: 'totalamount')
  final double totalAmount;
  @JsonKey(name: 'transactions')
  final List<SaleReturnTransactionModel>? transactions;

  SaleReturnModel({
    required this.docDate,
    required this.docno,
    required this.creditorNames,
    required this.totalValue,
    required this.detailTotalDiscount,
    required this.totalAfterDiscount,
    required this.totalExceptVat,
    required this.totalBeforeVat,
    required this.totalVatValue,
    required this.totalAmount,
    this.transactions = const [],
  });

  factory SaleReturnModel.fromJson(Map<String, dynamic> json) =>
      _$SaleReturnModelFromJson(json);

  Map<String, dynamic> toJson() => _$SaleReturnModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class SaleReturnTransactionModel {
  @JsonKey(name: 'shopid')
  final String shopId;
  @JsonKey(name: 'docno')
  final String docno;
  @JsonKey(name: 'docref')
  final String docRef;
  @JsonKey(name: 'barcode')
  final String barcode;
  @JsonKey(name: 'itemnames')
  final List<NameModel> itemNames;
  @JsonKey(name: 'unitnames')
  final List<NameModel> unitNames;
  @JsonKey(name: 'whnames')
  final List<NameModel> whNames;
  @JsonKey(name: 'locationnames')
  final List<NameModel> locationNames;
  @JsonKey(name: 'qty')
  final double qty;
  @JsonKey(name: 'price')
  final double price;
  @JsonKey(name: 'discount')
  final double discount;
  @JsonKey(name: 'sumamount')
  final double sumAmount;

  SaleReturnTransactionModel({
    required this.shopId,
    required this.docno,
    required this.docRef,
    required this.barcode,
    required this.itemNames,
    required this.unitNames,
    required this.whNames,
    required this.locationNames,
    required this.qty,
    required this.price,
    required this.discount,
    required this.sumAmount,
  });

  factory SaleReturnTransactionModel.fromJson(Map<String, dynamic> json) =>
      _$SaleReturnTransactionModelFromJson(json);

  Map<String, dynamic> toJson() => _$SaleReturnTransactionModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class NameModel {
  @JsonKey(name: 'code')
  final String code;
  @JsonKey(name: 'name')
  final String name;
  @JsonKey(name: 'isauto')
  final bool isAuto;
  @JsonKey(name: 'isdelete')
  final bool isDelete;

  NameModel({
    required this.code,
    required this.name,
    required this.isAuto,
    required this.isDelete,
  });

  factory NameModel.fromJson(Map<String, dynamic> json) =>
      _$NameModelFromJson(json);

  Map<String, dynamic> toJson() => _$NameModelToJson(this);

  // Helper method to get display name (usually Thai name)
  String getDisplayName() {
    return name.isNotEmpty ? name : code;
  }
}

/// Summary model for Sale Return Report
@JsonSerializable(explicitToJson: true)
class SaleReturnSummaryModel {
  @JsonKey(name: 'fromdate')
  final String fromDate;
  @JsonKey(name: 'todate')
  final String toDate;
  @JsonKey(name: 'total_records')
  final int totalRecords;
  @JsonKey(name: 'total_amount')
  final double totalAmount;

  SaleReturnSummaryModel({
    required this.fromDate,
    required this.toDate,
    required this.totalRecords,
    required this.totalAmount,
  });

  factory SaleReturnSummaryModel.fromJson(Map<String, dynamic> json) =>
      _$SaleReturnSummaryModelFromJson(json);

  Map<String, dynamic> toJson() => _$SaleReturnSummaryModelToJson(this);
}
