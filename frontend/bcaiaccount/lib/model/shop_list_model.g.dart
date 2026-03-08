// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'shop_list_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShopListModel _$ShopListModelFromJson(Map<String, dynamic> json) =>
    ShopListModel(
      shopid: json['shopid'] as String,
      name: json['name'] as String,
      branchcode: json['branchcode'] as String?,
      role: (json['role'] as num).toInt(),
      isfavorite: json['isfavorite'] as bool,
      lastaccessedat: json['lastaccessedat'] as String?,
      createdby: json['createdby'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ShopListModelToJson(ShopListModel instance) =>
    <String, dynamic>{
      'shopid': instance.shopid,
      'name': instance.name,
      'names': instance.names,
      'branchcode': instance.branchcode,
      'role': instance.role,
      'isfavorite': instance.isfavorite,
      'lastaccessedat': instance.lastaccessedat,
      'createdby': instance.createdby,
    };
