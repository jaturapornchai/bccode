import 'package:json_annotation/json_annotation.dart';

part 'import_product_model.g.dart';

@JsonSerializable()
class UploadSuccessModel {
  bool success;
  String id;

  UploadSuccessModel({required this.success, required this.id});

  factory UploadSuccessModel.fromJson(Map<String, dynamic> json) =>
      _$UploadSuccessModelFromJson(json);
  Map<String, dynamic> toJson() => _$UploadSuccessModelToJson(this);
}

@JsonSerializable()
class ImportProductModel {
  String? guidfixed;
  String? shopid;
  String? taskid;
  int? rownumber;
  String? barcode;
  String? name;
  String? unitcode;
  double? price;
  double? pricemember;
  double? pricedelivery;
  bool? isduplicate;
  bool? isexist;
  bool? isunitnotexist;

  ImportProductModel({
    String? guidfixed,
    String? shopid,
    String? taskid,
    int? rownumber,
    String? barcode,
    String? name,
    String? unitcode,
    double? price,
    double? pricemember,
    double? pricedelivery,
    bool? isduplicate,
    bool? isexist,
    bool? isunitnotexist,
  }) : guidfixed = guidfixed ?? '',
       shopid = shopid ?? '',
       taskid = taskid ?? '',
       rownumber = rownumber ?? 0,
       barcode = barcode ?? '',
       name = name ?? '',
       unitcode = unitcode ?? '',
       price = price ?? 0,
       pricemember = pricemember ?? 0,
       pricedelivery = pricedelivery ?? 0,
       isduplicate = isduplicate ?? false,
       isexist = isexist ?? false,
       isunitnotexist = isunitnotexist ?? false;

  factory ImportProductModel.fromJson(Map<String, dynamic> json) =>
      _$ImportProductModelFromJson(json);
  Map<String, dynamic> toJson() => _$ImportProductModelToJson(this);
}

// get status model
@JsonSerializable()
class ImportProductStatusModel {
  @JsonKey(name: 'task_id')
  String taskId;
  @JsonKey(name: 'shop_id')
  String shopId;
  String status;
  int progress;
  @JsonKey(name: 'created_at')
  String createdAt;
  @JsonKey(name: 'updated_at')
  String updatedAt;

  ImportProductStatusModel({
    required this.taskId,
    required this.shopId,
    required this.status,
    required this.progress,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ImportProductStatusModel.fromJson(Map<String, dynamic> json) =>
      _$ImportProductStatusModelFromJson(json);
  Map<String, dynamic> toJson() => _$ImportProductStatusModelToJson(this);
}

// Compare Detail Models
@JsonSerializable()
class ImportDataModel {
  String? barcode;
  String? code;
  String? name;
  String? unitcode;
  double? price;
  double? pricemember;
  double? pricedelivery;
  double? priceone;
  double? pricetwo;
  double? pricethree;
  double? pricefour;
  double? pricefive;
  double? pricesix;
  double? priseseven;
  double? priceeight;
  double? pricenine;
  bool? isduplicate;
  bool? isexist;
  bool? isunitnotexist;
  String? groupcode;
  String? groupsubonecode;
  String? groupsubtwocode;
  String? brandcode;
  String? designcode;
  String? modelcode;
  String? patterncode;
  String? gradecode;
  String? categorycode;
  String? classcode;
  String? barcoderef;
  double? standvalue;
  double? dividevalue;
  bool? issumpoint;

  ImportDataModel({
    this.barcode,
    this.code,
    this.name,
    this.unitcode,
    this.price,
    this.pricemember,
    this.pricedelivery,
    this.priceone,
    this.pricetwo,
    this.pricethree,
    this.pricefour,
    this.pricefive,
    this.pricesix,
    this.priseseven,
    this.priceeight,
    this.pricenine,
    this.isduplicate,
    this.isexist,
    this.isunitnotexist,
    this.groupcode,
    this.groupsubonecode,
    this.groupsubtwocode,
    this.brandcode,
    this.designcode,
    this.modelcode,
    this.patterncode,
    this.gradecode,
    this.categorycode,
    this.classcode,
    this.barcoderef,
    this.standvalue,
    this.dividevalue,
    this.issumpoint,
  });

  factory ImportDataModel.fromJson(Map<String, dynamic> json) =>
      _$ImportDataModelFromJson(json);
  Map<String, dynamic> toJson() => _$ImportDataModelToJson(this);
}

@JsonSerializable()
class CompareItemModel {
  @JsonKey(name: 'import_data')
  ImportDataModel? importData;
  String? status;
  @JsonKey(name: 'can_update')
  bool? canUpdate;

  CompareItemModel({this.importData, this.status, this.canUpdate});

  factory CompareItemModel.fromJson(Map<String, dynamic> json) =>
      _$CompareItemModelFromJson(json);
  Map<String, dynamic> toJson() => _$CompareItemModelToJson(this);
}

@JsonSerializable()
class CompareSummaryModel {
  @JsonKey(name: 'import_mode')
  String? importMode;
  @JsonKey(name: 'total_changes')
  int? totalChanges;
  @JsonKey(name: 'price_changes')
  int? priceChanges;
  @JsonKey(name: 'name_changes')
  int? nameChanges;
  @JsonKey(name: 'unit_changes')
  int? unitChanges;
  @JsonKey(name: 'master_data_changes')
  int? masterDataChanges;
  @JsonKey(name: 'ref_barcode_changes')
  int? refBarcodeChanges;
  @JsonKey(name: 'total_conflicts')
  int? totalConflicts;
  @JsonKey(name: 'high_conflicts')
  int? highConflicts;
  @JsonKey(name: 'medium_conflicts')
  int? mediumConflicts;
  @JsonKey(name: 'low_conflicts')
  int? lowConflicts;
  @JsonKey(name: 'estimated_time')
  String? estimatedTime;
  @JsonKey(name: 'created_at')
  String? createdAt;

  CompareSummaryModel({
    this.importMode,
    this.totalChanges,
    this.priceChanges,
    this.nameChanges,
    this.unitChanges,
    this.masterDataChanges,
    this.refBarcodeChanges,
    this.totalConflicts,
    this.highConflicts,
    this.mediumConflicts,
    this.lowConflicts,
    this.estimatedTime,
    this.createdAt,
  });

  factory CompareSummaryModel.fromJson(Map<String, dynamic> json) =>
      _$CompareSummaryModelFromJson(json);
  Map<String, dynamic> toJson() => _$CompareSummaryModelToJson(this);
}

@JsonSerializable()
class CompareDetailModel {
  @JsonKey(name: 'task_id')
  String? taskId;
  @JsonKey(name: 'total_records')
  int? totalRecords;
  @JsonKey(name: 'new_records')
  int? newRecords;
  @JsonKey(name: 'existing_records')
  int? existingRecords;
  @JsonKey(name: 'updated_records')
  int? updatedRecords;
  @JsonKey(name: 'conflict_records')
  int? conflictRecords;
  List<CompareItemModel>? items;
  CompareSummaryModel? summary;
  @JsonKey(name: 'created_at')
  String? createdAt;

  CompareDetailModel({
    this.taskId,
    this.totalRecords,
    this.newRecords,
    this.existingRecords,
    this.updatedRecords,
    this.conflictRecords,
    this.items,
    this.summary,
    this.createdAt,
  });

  factory CompareDetailModel.fromJson(Map<String, dynamic> json) =>
      _$CompareDetailModelFromJson(json);
  Map<String, dynamic> toJson() => _$CompareDetailModelToJson(this);
}

// Apply Import Models
@JsonSerializable()
class ApplyImportDataModel {
  @JsonKey(name: 'batch_size')
  int? batchSize;
  @JsonKey(name: 'import_mode')
  String? importMode;
  String? mode;
  @JsonKey(name: 'original_task')
  String? originalTask;
  @JsonKey(name: 'process_task_id')
  String? processTaskId;
  @JsonKey(name: 'status_url')
  String? statusUrl;

  ApplyImportDataModel({
    this.batchSize,
    this.importMode,
    this.mode,
    this.originalTask,
    this.processTaskId,
    this.statusUrl,
  });

  factory ApplyImportDataModel.fromJson(Map<String, dynamic> json) =>
      _$ApplyImportDataModelFromJson(json);
  Map<String, dynamic> toJson() => _$ApplyImportDataModelToJson(this);
}

@JsonSerializable()
class ApplyImportResponseModel {
  bool? success;
  String? message;
  String? id;
  ApplyImportDataModel? data;

  ApplyImportResponseModel({this.success, this.message, this.id, this.data});

  factory ApplyImportResponseModel.fromJson(Map<String, dynamic> json) =>
      _$ApplyImportResponseModelFromJson(json);
  Map<String, dynamic> toJson() => _$ApplyImportResponseModelToJson(this);
}
