// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'import_product_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

UploadSuccessModel _$UploadSuccessModelFromJson(Map<String, dynamic> json) =>
    UploadSuccessModel(
      success: json['success'] as bool,
      id: json['id'] as String,
    );

Map<String, dynamic> _$UploadSuccessModelToJson(UploadSuccessModel instance) =>
    <String, dynamic>{'success': instance.success, 'id': instance.id};

ImportProductModel _$ImportProductModelFromJson(Map<String, dynamic> json) =>
    ImportProductModel(
      guidfixed: json['guidfixed'] as String?,
      shopid: json['shopid'] as String?,
      taskid: json['taskid'] as String?,
      rownumber: (json['rownumber'] as num?)?.toInt(),
      barcode: json['barcode'] as String?,
      name: json['name'] as String?,
      unitcode: json['unitcode'] as String?,
      price: (json['price'] as num?)?.toDouble(),
      pricemember: (json['pricemember'] as num?)?.toDouble(),
      pricedelivery: (json['pricedelivery'] as num?)?.toDouble(),
      isduplicate: json['isduplicate'] as bool?,
      isexist: json['isexist'] as bool?,
      isunitnotexist: json['isunitnotexist'] as bool?,
    );

Map<String, dynamic> _$ImportProductModelToJson(ImportProductModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'shopid': instance.shopid,
      'taskid': instance.taskid,
      'rownumber': instance.rownumber,
      'barcode': instance.barcode,
      'name': instance.name,
      'unitcode': instance.unitcode,
      'price': instance.price,
      'pricemember': instance.pricemember,
      'pricedelivery': instance.pricedelivery,
      'isduplicate': instance.isduplicate,
      'isexist': instance.isexist,
      'isunitnotexist': instance.isunitnotexist,
    };

ImportProductStatusModel _$ImportProductStatusModelFromJson(
  Map<String, dynamic> json,
) => ImportProductStatusModel(
  taskId: json['task_id'] as String,
  shopId: json['shop_id'] as String,
  status: json['status'] as String,
  progress: (json['progress'] as num).toInt(),
  createdAt: json['created_at'] as String,
  updatedAt: json['updated_at'] as String,
);

Map<String, dynamic> _$ImportProductStatusModelToJson(
  ImportProductStatusModel instance,
) => <String, dynamic>{
  'task_id': instance.taskId,
  'shop_id': instance.shopId,
  'status': instance.status,
  'progress': instance.progress,
  'created_at': instance.createdAt,
  'updated_at': instance.updatedAt,
};

ImportDataModel _$ImportDataModelFromJson(Map<String, dynamic> json) =>
    ImportDataModel(
      barcode: json['barcode'] as String?,
      code: json['code'] as String?,
      name: json['name'] as String?,
      unitcode: json['unitcode'] as String?,
      price: (json['price'] as num?)?.toDouble(),
      pricemember: (json['pricemember'] as num?)?.toDouble(),
      pricedelivery: (json['pricedelivery'] as num?)?.toDouble(),
      priceone: (json['priceone'] as num?)?.toDouble(),
      pricetwo: (json['pricetwo'] as num?)?.toDouble(),
      pricethree: (json['pricethree'] as num?)?.toDouble(),
      pricefour: (json['pricefour'] as num?)?.toDouble(),
      pricefive: (json['pricefive'] as num?)?.toDouble(),
      pricesix: (json['pricesix'] as num?)?.toDouble(),
      priseseven: (json['priseseven'] as num?)?.toDouble(),
      priceeight: (json['priceeight'] as num?)?.toDouble(),
      pricenine: (json['pricenine'] as num?)?.toDouble(),
      isduplicate: json['isduplicate'] as bool?,
      isexist: json['isexist'] as bool?,
      isunitnotexist: json['isunitnotexist'] as bool?,
      groupcode: json['groupcode'] as String?,
      groupsubonecode: json['groupsubonecode'] as String?,
      groupsubtwocode: json['groupsubtwocode'] as String?,
      brandcode: json['brandcode'] as String?,
      designcode: json['designcode'] as String?,
      modelcode: json['modelcode'] as String?,
      patterncode: json['patterncode'] as String?,
      gradecode: json['gradecode'] as String?,
      categorycode: json['categorycode'] as String?,
      classcode: json['classcode'] as String?,
      barcoderef: json['barcoderef'] as String?,
      standvalue: (json['standvalue'] as num?)?.toDouble(),
      dividevalue: (json['dividevalue'] as num?)?.toDouble(),
      issumpoint: json['issumpoint'] as bool?,
    );

Map<String, dynamic> _$ImportDataModelToJson(ImportDataModel instance) =>
    <String, dynamic>{
      'barcode': instance.barcode,
      'code': instance.code,
      'name': instance.name,
      'unitcode': instance.unitcode,
      'price': instance.price,
      'pricemember': instance.pricemember,
      'pricedelivery': instance.pricedelivery,
      'priceone': instance.priceone,
      'pricetwo': instance.pricetwo,
      'pricethree': instance.pricethree,
      'pricefour': instance.pricefour,
      'pricefive': instance.pricefive,
      'pricesix': instance.pricesix,
      'priseseven': instance.priseseven,
      'priceeight': instance.priceeight,
      'pricenine': instance.pricenine,
      'isduplicate': instance.isduplicate,
      'isexist': instance.isexist,
      'isunitnotexist': instance.isunitnotexist,
      'groupcode': instance.groupcode,
      'groupsubonecode': instance.groupsubonecode,
      'groupsubtwocode': instance.groupsubtwocode,
      'brandcode': instance.brandcode,
      'designcode': instance.designcode,
      'modelcode': instance.modelcode,
      'patterncode': instance.patterncode,
      'gradecode': instance.gradecode,
      'categorycode': instance.categorycode,
      'classcode': instance.classcode,
      'barcoderef': instance.barcoderef,
      'standvalue': instance.standvalue,
      'dividevalue': instance.dividevalue,
      'issumpoint': instance.issumpoint,
    };

CompareItemModel _$CompareItemModelFromJson(Map<String, dynamic> json) =>
    CompareItemModel(
      importData: json['import_data'] == null
          ? null
          : ImportDataModel.fromJson(
              json['import_data'] as Map<String, dynamic>,
            ),
      status: json['status'] as String?,
      canUpdate: json['can_update'] as bool?,
    );

Map<String, dynamic> _$CompareItemModelToJson(CompareItemModel instance) =>
    <String, dynamic>{
      'import_data': instance.importData,
      'status': instance.status,
      'can_update': instance.canUpdate,
    };

CompareSummaryModel _$CompareSummaryModelFromJson(Map<String, dynamic> json) =>
    CompareSummaryModel(
      importMode: json['import_mode'] as String?,
      totalChanges: (json['total_changes'] as num?)?.toInt(),
      priceChanges: (json['price_changes'] as num?)?.toInt(),
      nameChanges: (json['name_changes'] as num?)?.toInt(),
      unitChanges: (json['unit_changes'] as num?)?.toInt(),
      masterDataChanges: (json['master_data_changes'] as num?)?.toInt(),
      refBarcodeChanges: (json['ref_barcode_changes'] as num?)?.toInt(),
      totalConflicts: (json['total_conflicts'] as num?)?.toInt(),
      highConflicts: (json['high_conflicts'] as num?)?.toInt(),
      mediumConflicts: (json['medium_conflicts'] as num?)?.toInt(),
      lowConflicts: (json['low_conflicts'] as num?)?.toInt(),
      estimatedTime: json['estimated_time'] as String?,
      createdAt: json['created_at'] as String?,
    );

Map<String, dynamic> _$CompareSummaryModelToJson(
  CompareSummaryModel instance,
) => <String, dynamic>{
  'import_mode': instance.importMode,
  'total_changes': instance.totalChanges,
  'price_changes': instance.priceChanges,
  'name_changes': instance.nameChanges,
  'unit_changes': instance.unitChanges,
  'master_data_changes': instance.masterDataChanges,
  'ref_barcode_changes': instance.refBarcodeChanges,
  'total_conflicts': instance.totalConflicts,
  'high_conflicts': instance.highConflicts,
  'medium_conflicts': instance.mediumConflicts,
  'low_conflicts': instance.lowConflicts,
  'estimated_time': instance.estimatedTime,
  'created_at': instance.createdAt,
};

CompareDetailModel _$CompareDetailModelFromJson(Map<String, dynamic> json) =>
    CompareDetailModel(
      taskId: json['task_id'] as String?,
      totalRecords: (json['total_records'] as num?)?.toInt(),
      newRecords: (json['new_records'] as num?)?.toInt(),
      existingRecords: (json['existing_records'] as num?)?.toInt(),
      updatedRecords: (json['updated_records'] as num?)?.toInt(),
      conflictRecords: (json['conflict_records'] as num?)?.toInt(),
      items: (json['items'] as List<dynamic>?)
          ?.map((e) => CompareItemModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      summary: json['summary'] == null
          ? null
          : CompareSummaryModel.fromJson(
              json['summary'] as Map<String, dynamic>,
            ),
      createdAt: json['created_at'] as String?,
    );

Map<String, dynamic> _$CompareDetailModelToJson(CompareDetailModel instance) =>
    <String, dynamic>{
      'task_id': instance.taskId,
      'total_records': instance.totalRecords,
      'new_records': instance.newRecords,
      'existing_records': instance.existingRecords,
      'updated_records': instance.updatedRecords,
      'conflict_records': instance.conflictRecords,
      'items': instance.items,
      'summary': instance.summary,
      'created_at': instance.createdAt,
    };

ApplyImportDataModel _$ApplyImportDataModelFromJson(
  Map<String, dynamic> json,
) => ApplyImportDataModel(
  batchSize: (json['batch_size'] as num?)?.toInt(),
  importMode: json['import_mode'] as String?,
  mode: json['mode'] as String?,
  originalTask: json['original_task'] as String?,
  processTaskId: json['process_task_id'] as String?,
  statusUrl: json['status_url'] as String?,
);

Map<String, dynamic> _$ApplyImportDataModelToJson(
  ApplyImportDataModel instance,
) => <String, dynamic>{
  'batch_size': instance.batchSize,
  'import_mode': instance.importMode,
  'mode': instance.mode,
  'original_task': instance.originalTask,
  'process_task_id': instance.processTaskId,
  'status_url': instance.statusUrl,
};

ApplyImportResponseModel _$ApplyImportResponseModelFromJson(
  Map<String, dynamic> json,
) => ApplyImportResponseModel(
  success: json['success'] as bool?,
  message: json['message'] as String?,
  id: json['id'] as String?,
  data: json['data'] == null
      ? null
      : ApplyImportDataModel.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$ApplyImportResponseModelToJson(
  ApplyImportResponseModel instance,
) => <String, dynamic>{
  'success': instance.success,
  'message': instance.message,
  'id': instance.id,
  'data': instance.data,
};
