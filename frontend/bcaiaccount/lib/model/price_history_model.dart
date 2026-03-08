import 'package:json_annotation/json_annotation.dart';

part 'price_history_model.g.dart';

@JsonSerializable()
class PriceHistoryModel {
  @JsonKey(name: "guidfixed")
  final String? guidfixed;

  @JsonKey(name: "shopid")
  final String? shopid;

  @JsonKey(name: "id")
  final String? id;

  @JsonKey(name: "productbarcodeguid")
  final String? productbarcodeguid;

  @JsonKey(name: "barcode")
  final String? barcode;

  @JsonKey(name: "productname")
  final String? productname;

  @JsonKey(name: "pricetype")
  final String? pricetype;

  @JsonKey(name: "keynumber")
  final int? keynumber;

  @JsonKey(name: "oldprice")
  final double? oldprice;

  @JsonKey(name: "newprice")
  final double? newprice;

  @JsonKey(name: "pricedifference")
  final double? pricedifference;

  @JsonKey(name: "action")
  final String? action;

  @JsonKey(name: "createdby")
  final String? createdby;

  @JsonKey(name: "createdat")
  final String? createdat;

  @JsonKey(name: "remark")
  final String? remark;

  const PriceHistoryModel({
    this.guidfixed,
    this.shopid,
    this.id,
    this.productbarcodeguid,
    this.barcode,
    this.productname,
    this.pricetype,
    this.keynumber,
    this.oldprice,
    this.newprice,
    this.pricedifference,
    this.action,
    this.createdby,
    this.createdat,
    this.remark,
  });

  factory PriceHistoryModel.fromJson(Map<String, dynamic> json) =>
      _$PriceHistoryModelFromJson(json);

  Map<String, dynamic> toJson() => _$PriceHistoryModelToJson(this);
}

@JsonSerializable()
class PaginationModel {
  @JsonKey(name: "total")
  final int? total;

  @JsonKey(name: "page")
  final int? page;

  @JsonKey(name: "perPage")
  final int? perPage;

  @JsonKey(name: "prev")
  final int? prev;

  @JsonKey(name: "next")
  final int? next;

  @JsonKey(name: "totalPage")
  final int? totalPage;

  const PaginationModel({
    this.total,
    this.page,
    this.perPage,
    this.prev,
    this.next,
    this.totalPage,
  });

  factory PaginationModel.fromJson(Map<String, dynamic> json) =>
      _$PaginationModelFromJson(json);

  Map<String, dynamic> toJson() => _$PaginationModelToJson(this);
}

@JsonSerializable()
class PriceHistoryResponseModel {
  @JsonKey(name: "success")
  final bool? success;

  @JsonKey(name: "data")
  final List<PriceHistoryModel>? data;

  @JsonKey(name: "pagination")
  final PaginationModel? pagination;

  const PriceHistoryResponseModel({
    this.success,
    this.data,
    this.pagination,
  });

  factory PriceHistoryResponseModel.fromJson(Map<String, dynamic> json) =>
      _$PriceHistoryResponseModelFromJson(json);

  Map<String, dynamic> toJson() => _$PriceHistoryResponseModelToJson(this);
}
