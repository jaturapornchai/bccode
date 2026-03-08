// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'doc_payload_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DocPayLoadModel _$DocPayLoadModelFromJson(Map<String, dynamic> json) =>
    DocPayLoadModel(
      shopid: json['shopid'] as String,
      system: json['system'] as String,
      offset: (json['offset'] as num).toInt(),
      limit: (json['limit'] as num).toInt(),
      search: json['search'] as String,
      custcode: json['custcode'] as String,
      dateorder: (json['dateorder'] as num).toInt(),
      fromdate: json['fromdate'] as String?,
      todate: json['todate'] as String?,
      minamount: (json['minamount'] as num?)?.toDouble(),
      maxamount: (json['maxamount'] as num?)?.toDouble(),
      custcodes: (json['custcodes'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$DocPayLoadModelToJson(DocPayLoadModel instance) =>
    <String, dynamic>{
      'shopid': instance.shopid,
      'system': instance.system,
      'offset': instance.offset,
      'limit': instance.limit,
      'search': instance.search,
      'custcode': instance.custcode,
      'dateorder': instance.dateorder,
      'fromdate': ?instance.fromdate,
      'todate': ?instance.todate,
      'minamount': ?instance.minamount,
      'maxamount': ?instance.maxamount,
      'custcodes': ?instance.custcodes,
    };

MongoGetDataPayloadModel _$MongoGetDataPayloadModelFromJson(
  Map<String, dynamic> json,
) => MongoGetDataPayloadModel(
  shopid: json['shopid'] as String,
  collection: json['collection'] as String,
  guidfixed: json['guidfixed'] as String,
);

Map<String, dynamic> _$MongoGetDataPayloadModelToJson(
  MongoGetDataPayloadModel instance,
) => <String, dynamic>{
  'shopid': instance.shopid,
  'collection': instance.collection,
  'guidfixed': instance.guidfixed,
};
