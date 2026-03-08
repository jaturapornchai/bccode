import 'package:json_annotation/json_annotation.dart';

part 'bulk_points_request_model.g.dart';

// Model สำหรับ Bulk Add Points Request
@JsonSerializable()
class BulkPointsRequestModel {
  final String pointscode;
  final int pointamount;
  final String description;

  BulkPointsRequestModel({
    required this.pointscode,
    required this.pointamount,
    required this.description,
  });

  factory BulkPointsRequestModel.fromJson(Map<String, dynamic> json) =>
      _$BulkPointsRequestModelFromJson(json);

  Map<String, dynamic> toJson() => _$BulkPointsRequestModelToJson(this);
}

// Model สำหรับ Bulk Add Points Response
@JsonSerializable()
class BulkPointsResponseModel {
  final bool success;
  final String message;
  final BulkPointsDataModel? data;

  BulkPointsResponseModel({
    required this.success,
    required this.message,
    this.data,
  });

  factory BulkPointsResponseModel.fromJson(Map<String, dynamic> json) =>
      _$BulkPointsResponseModelFromJson(json);

  Map<String, dynamic> toJson() => _$BulkPointsResponseModelToJson(this);
}

@JsonSerializable()
class BulkPointsDataModel {
  final int success;
  final int failed;
  final int total;
  @JsonKey(name: 'success_items')
  final List<String>? successItems;
  @JsonKey(name: 'failed_items')
  final List<Map<String, dynamic>>? failedItems;

  BulkPointsDataModel({
    required this.success,
    required this.failed,
    required this.total,
    this.successItems,
    this.failedItems,
  });

  factory BulkPointsDataModel.fromJson(Map<String, dynamic> json) =>
      _$BulkPointsDataModelFromJson(json);

  Map<String, dynamic> toJson() => _$BulkPointsDataModelToJson(this);
}
