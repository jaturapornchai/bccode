// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_category_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductCategoryModel _$ProductCategoryModelFromJson(
  Map<String, dynamic> json,
) => ProductCategoryModel(
  guidfixed: json['guidfixed'] as String,
  parentguid: json['parentguid'] as String,
  parentguidall: json['parentguidall'] as String,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  imageuri: json['imageuri'] as String,
  childcount: (json['childcount'] as num).toInt(),
  xsorts: (json['xsorts'] as List<dynamic>?)
      ?.map((e) => SortDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  useimageorcolor: json['useimageorcolor'] as bool,
  isdisabled: json['isdisabled'] as bool,
  codelist: (json['codelist'] as List<dynamic>?)
      ?.map(
        (e) => ProductCategoryCodeListModel.fromJson(e as Map<String, dynamic>),
      )
      .toList(),
  colorselect: json['colorselect'] as String,
  colorselecthex: json['colorselecthex'] as String,
  coveruri: json['coveruri'] as String?,
  groupnumber: (json['groupnumber'] as num?)?.toInt(),
  timeforsales: (json['timeforsales'] as List<dynamic>?)
      ?.map((e) => TimeForSaleModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ProductCategoryModelToJson(
  ProductCategoryModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'parentguid': instance.parentguid,
  'parentguidall': instance.parentguidall,
  'names': instance.names?.map((e) => e.toJson()).toList(),
  'childcount': instance.childcount,
  'imageuri': instance.imageuri,
  'useimageorcolor': instance.useimageorcolor,
  'isdisabled': instance.isdisabled,
  'colorselect': instance.colorselect,
  'colorselecthex': instance.colorselecthex,
  'xsorts': instance.xsorts?.map((e) => e.toJson()).toList(),
  'codelist': instance.codelist?.map((e) => e.toJson()).toList(),
  'coveruri': instance.coveruri,
  'groupnumber': instance.groupnumber,
  'timeforsales': instance.timeforsales?.map((e) => e.toJson()).toList(),
};

ProductCategoryCodeListModel _$ProductCategoryCodeListModelFromJson(
  Map<String, dynamic> json,
) => ProductCategoryCodeListModel(
  code: json['code'] as String,
  xorder: (json['xorder'] as num).toInt(),
  barcode: json['barcode'] as String,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitnames: (json['unitnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitcode: json['unitcode'] as String,
);

Map<String, dynamic> _$ProductCategoryCodeListModelToJson(
  ProductCategoryCodeListModel instance,
) => <String, dynamic>{
  'code': instance.code,
  'barcode': instance.barcode,
  'names': instance.names?.map((e) => e.toJson()).toList(),
  'unitcode': instance.unitcode,
  'unitnames': instance.unitnames?.map((e) => e.toJson()).toList(),
  'xorder': instance.xorder,
};

TimeForSaleModel _$TimeForSaleModelFromJson(Map<String, dynamic> json) =>
    TimeForSaleModel(
      daysofweek: (json['daysofweek'] as List<dynamic>?)
          ?.map((e) => (e as num).toInt())
          .toList(),
      fromdate: json['fromdate'] as String?,
      todate: json['todate'] as String?,
      fromtime: json['fromtime'] as String?,
      totime: json['totime'] as String?,
    );

Map<String, dynamic> _$TimeForSaleModelToJson(TimeForSaleModel instance) =>
    <String, dynamic>{
      'daysofweek': instance.daysofweek,
      'fromdate': instance.fromdate,
      'todate': instance.todate,
      'fromtime': instance.fromtime,
      'totime': instance.totime,
    };
