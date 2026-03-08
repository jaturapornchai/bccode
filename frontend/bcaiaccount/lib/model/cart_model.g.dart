// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'cart_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CartModel _$CartModelFromJson(Map<String, dynamic> json) => CartModel(
  cartId: json['cartId'] as String?,
  cartName: json['cartName'] as String?,
  cartTransFlag: (json['cartTransFlag'] as num?)?.toInt(),
  cartStatus: json['cartStatus'] as String?,
  statistics: json['statistics'] == null
      ? null
      : CartStatistics.fromJson(json['statistics'] as Map<String, dynamic>),
  branch: json['branch'] == null
      ? null
      : CartLocation.fromJson(json['branch'] as Map<String, dynamic>),
  warehouse: json['warehouse'] == null
      ? null
      : CartLocation.fromJson(json['warehouse'] as Map<String, dynamic>),
  location: json['location'] == null
      ? null
      : CartLocation.fromJson(json['location'] as Map<String, dynamic>),
  destWarehouse: json['destWarehouse'] == null
      ? null
      : CartLocation.fromJson(json['destWarehouse'] as Map<String, dynamic>),
  destLocation: json['destLocation'] == null
      ? null
      : CartLocation.fromJson(json['destLocation'] as Map<String, dynamic>),
  items: (json['items'] as List<dynamic>?)
      ?.map((e) => CartItem.fromJson(e as Map<String, dynamic>))
      .toList(),
  createdAt: json['createdAt'] as String?,
);

Map<String, dynamic> _$CartModelToJson(CartModel instance) => <String, dynamic>{
  'cartId': instance.cartId,
  'cartName': instance.cartName,
  'cartTransFlag': instance.cartTransFlag,
  'cartStatus': instance.cartStatus,
  'statistics': instance.statistics?.toJson(),
  'branch': instance.branch?.toJson(),
  'warehouse': instance.warehouse?.toJson(),
  'location': instance.location?.toJson(),
  'destWarehouse': instance.destWarehouse?.toJson(),
  'destLocation': instance.destLocation?.toJson(),
  'items': instance.items?.map((e) => e.toJson()).toList(),
  'createdAt': instance.createdAt,
};

CartStatistics _$CartStatisticsFromJson(Map<String, dynamic> json) =>
    CartStatistics(
      totalQuantity: (json['totalQuantity'] as num).toInt(),
      uniqueItems: (json['uniqueItems'] as num).toInt(),
    );

Map<String, dynamic> _$CartStatisticsToJson(CartStatistics instance) =>
    <String, dynamic>{
      'totalQuantity': instance.totalQuantity,
      'uniqueItems': instance.uniqueItems,
    };

CartLocation _$CartLocationFromJson(Map<String, dynamic> json) => CartLocation(
  code: json['code'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$CartLocationToJson(CartLocation instance) =>
    <String, dynamic>{
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };

CartItem _$CartItemFromJson(Map<String, dynamic> json) => CartItem(
  barcode: json['barcode'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  itemunitcode: json['itemunitcode'] as String,
  itemunitnames: (json['itemunitnames'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  quantity: (json['quantity'] as num).toDouble(),
  description: json['description'] as String?,
  imageuri: json['imageuri'] as String?,
);

Map<String, dynamic> _$CartItemToJson(CartItem instance) => <String, dynamic>{
  'barcode': instance.barcode,
  'names': instance.names.map((e) => e.toJson()).toList(),
  'itemunitcode': instance.itemunitcode,
  'itemunitnames': instance.itemunitnames.map((e) => e.toJson()).toList(),
  'quantity': instance.quantity,
  'description': instance.description,
  'imageuri': instance.imageuri,
};
