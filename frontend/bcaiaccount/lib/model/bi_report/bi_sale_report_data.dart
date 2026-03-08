import 'package:json_annotation/json_annotation.dart';

part 'bi_sale_report_data.g.dart';

// Custom converter for handling String/int/double to double conversion
class DoubleConverter implements JsonConverter<double, dynamic> {
  const DoubleConverter();

  @override
  double fromJson(dynamic value) {
    if (value is double) return value;
    if (value is int) return value.toDouble();
    if (value is String) {
      final parsed = double.tryParse(value);
      if (parsed != null) return parsed;
    }
    return 0.0; // Default value if parsing fails
  }

  @override
  dynamic toJson(double value) => value;
}

// Custom converter for handling String/int to int conversion
class IntConverter implements JsonConverter<int, dynamic> {
  const IntConverter();

  @override
  int fromJson(dynamic value) {
    if (value is int) return value;
    if (value is double) return value.toInt();
    if (value is String) {
      final parsed = int.tryParse(value);
      if (parsed != null) return parsed;
    }
    return 0; // Default value if parsing fails
  }

  @override
  dynamic toJson(int value) => value;
}

@JsonSerializable()
class SaleReportData {
  final String docdate;
  @JsonKey(name: 'doc_time')
  final String docTime;
  final String docno;
  final String creditorcode;
  @JsonKey(defaultValue: <SaleCreditorName>[])
  final List<SaleCreditorName> creditornames;
  @DoubleConverter()
  final double totalvalue;
  @DoubleConverter()
  final double detailtotaldiscount;
  @DoubleConverter()
  final double totalafterdiscount;
  @DoubleConverter()
  final double totalexceptvat;
  @DoubleConverter()
  final double totalbeforevat;
  @DoubleConverter()
  final double totalvatvalue;
  @DoubleConverter()
  final double detailtotalamount;
  @DoubleConverter()
  final double totaldiscount;
  @DoubleConverter()
  final double totalamount;
  @JsonKey(defaultValue: '')
  final String salecode;
  @JsonKey(defaultValue: '')
  final String salename;
  @JsonKey(defaultValue: '')
  final String
      inquirytype; // เปลี่ยนเป็น String: "" = ทั้งหมด, "1" = ขาย, "2" = คืน
  final bool iscancel;
  final bool ispos;
  final String branchcode;
  @JsonKey(defaultValue: <BranchName>[])
  final List<BranchName> branchnames;
  @JsonKey(defaultValue: <SaleTransaction>[])
  final List<SaleTransaction> transactions;

  const SaleReportData({
    required this.docdate,
    required this.docTime,
    required this.docno,
    required this.creditorcode,
    required this.creditornames,
    required this.totalvalue,
    required this.detailtotaldiscount,
    required this.totalafterdiscount,
    required this.totalexceptvat,
    required this.totalbeforevat,
    required this.totalvatvalue,
    required this.detailtotalamount,
    required this.totaldiscount,
    required this.totalamount,
    required this.salecode,
    required this.salename,
    required this.inquirytype,
    required this.iscancel,
    required this.ispos,
    required this.branchcode,
    required this.branchnames,
    required this.transactions,
  });

  factory SaleReportData.fromJson(Map<String, dynamic> json) =>
      _$SaleReportDataFromJson(json);

  Map<String, dynamic> toJson() => _$SaleReportDataToJson(this);
}

@JsonSerializable()
class SaleCreditorName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const SaleCreditorName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory SaleCreditorName.fromJson(Map<String, dynamic> json) =>
      _$SaleCreditorNameFromJson(json);

  Map<String, dynamic> toJson() => _$SaleCreditorNameToJson(this);
}

@JsonSerializable()
class BranchName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const BranchName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory BranchName.fromJson(Map<String, dynamic> json) =>
      _$BranchNameFromJson(json);

  Map<String, dynamic> toJson() => _$BranchNameToJson(this);
}

// Transaction Models for detailed report
@JsonSerializable()
class SaleTransaction {
  final String shopid;
  final String docdate;
  final String docno;
  @IntConverter()
  final int linenumber;
  @JsonKey(defaultValue: '')
  final String barcode;
  @JsonKey(defaultValue: <ItemName>[])
  final List<ItemName> itemnames;
  @JsonKey(defaultValue: <UnitName>[])
  final List<UnitName> unitnames;
  @JsonKey(defaultValue: <WarehouseName>[])
  final List<WarehouseName> whnames;
  @JsonKey(defaultValue: <LocationName>[])
  final List<LocationName> locationnames;
  @DoubleConverter()
  final double qty;
  @DoubleConverter()
  final double price;
  @DoubleConverter()
  final double discount;
  @DoubleConverter()
  final double sumamount;

  const SaleTransaction({
    required this.shopid,
    required this.docdate,
    required this.docno,
    required this.linenumber,
    required this.barcode,
    required this.itemnames,
    required this.unitnames,
    required this.whnames,
    required this.locationnames,
    required this.qty,
    required this.price,
    required this.discount,
    required this.sumamount,
  });

  factory SaleTransaction.fromJson(Map<String, dynamic> json) =>
      _$SaleTransactionFromJson(json);

  Map<String, dynamic> toJson() => _$SaleTransactionToJson(this);
}

@JsonSerializable()
class ItemName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const ItemName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory ItemName.fromJson(Map<String, dynamic> json) =>
      _$ItemNameFromJson(json);

  Map<String, dynamic> toJson() => _$ItemNameToJson(this);
}

@JsonSerializable()
class UnitName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const UnitName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory UnitName.fromJson(Map<String, dynamic> json) =>
      _$UnitNameFromJson(json);

  Map<String, dynamic> toJson() => _$UnitNameToJson(this);
}

@JsonSerializable()
class WarehouseName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const WarehouseName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory WarehouseName.fromJson(Map<String, dynamic> json) =>
      _$WarehouseNameFromJson(json);

  Map<String, dynamic> toJson() => _$WarehouseNameToJson(this);
}

@JsonSerializable()
class LocationName {
  final String code;
  final String name;
  final bool isauto;
  final bool isdelete;

  const LocationName({
    required this.code,
    required this.name,
    required this.isauto,
    required this.isdelete,
  });

  factory LocationName.fromJson(Map<String, dynamic> json) =>
      _$LocationNameFromJson(json);

  Map<String, dynamic> toJson() => _$LocationNameToJson(this);
}
