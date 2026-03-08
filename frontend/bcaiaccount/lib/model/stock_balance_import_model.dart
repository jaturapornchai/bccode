import 'package:json_annotation/json_annotation.dart';

part 'stock_balance_import_model.g.dart';

@JsonSerializable()
class UploadSuccessModel {
  bool success;
  String id;

  UploadSuccessModel({
    required this.success,
    required this.id,
  });

  factory UploadSuccessModel.fromJson(Map<String, dynamic> json) =>
      _$UploadSuccessModelFromJson(json);
  Map<String, dynamic> toJson() => _$UploadSuccessModelToJson(this);
}

@JsonSerializable()
class StockBalanceImportModel {
  String? guidfixed;
  String? shopid;
  String? taskid;
  int? rownumber;
  String? barcode;
  String? name;
  String? unitcode;
  String? warehousecode;
  String? shelfcode;
  double? qty;
  double? price;
  double? sumamount;
  bool? isnotexist;

  StockBalanceImportModel({
    String? guidfixed,
    String? shopid,
    String? taskid,
    int? rownumber,
    String? barcode,
    String? name,
    String? unitcode,
    String? warehousecode,
    String? shelfcode,
    double? qty,
    double? price,
    double? sumamount,
    bool? isnotexist,
  })  : guidfixed = guidfixed ?? '',
        shopid = shopid ?? '',
        taskid = taskid ?? '',
        rownumber = rownumber ?? 0,
        barcode = barcode ?? '',
        name = name ?? '',
        unitcode = unitcode ?? '',
        warehousecode = warehousecode ?? '',
        shelfcode = shelfcode ?? '',
        qty = qty ?? 0,
        price = price ?? 0,
        sumamount = sumamount ?? 0,
        isnotexist = isnotexist ?? false;

  factory StockBalanceImportModel.fromJson(Map<String, dynamic> json) =>
      _$StockBalanceImportModelFromJson(json);
  Map<String, dynamic> toJson() => _$StockBalanceImportModelToJson(this);
}

@JsonSerializable()
class TotalModel {
  late final double totalitem;
  late final double totalamount;

  TotalModel({
    required this.totalitem,
    required this.totalamount,
  });

  factory TotalModel.fromJson(Map<String, dynamic> json) =>
      _$TotalModelFromJson(json);
  Map<String, dynamic> toJson() => _$TotalModelToJson(this);
}
