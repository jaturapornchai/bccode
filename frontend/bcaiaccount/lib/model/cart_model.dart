// cart_model.dart
import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'cart_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CartModel {
  final String? cartId;
  final String? cartName;
  final int? cartTransFlag;
  final String? cartStatus;
  final CartStatistics? statistics;
  final CartLocation? branch;
  final CartLocation? warehouse;
  final CartLocation? location;
  final CartLocation? destWarehouse;
  final CartLocation? destLocation;
  final List<CartItem>? items;
  final String? createdAt;

  CartModel({
    String? cartId,
    String? cartName,
    int? cartTransFlag,
    String? cartStatus,
    CartStatistics? statistics,
    CartLocation? branch,
    CartLocation? warehouse,
    CartLocation? location,
    CartLocation? destWarehouse,
    CartLocation? destLocation,
    List<CartItem>? items,
    String? createdAt,
  })  : cartId = cartId ?? '',
        cartName = cartName ?? '',
        cartTransFlag = cartTransFlag ?? 0,
        cartStatus = cartStatus ?? '',
        statistics =
            statistics ?? CartStatistics(totalQuantity: 0, uniqueItems: 0),
        branch = branch ?? CartLocation(code: '', names: []),
        warehouse = warehouse ?? CartLocation(code: '', names: []),
        location = location ?? CartLocation(code: '', names: []),
        destWarehouse = destWarehouse ?? CartLocation(code: '', names: []),
        destLocation = destLocation ?? CartLocation(code: '', names: []),
        items = items ?? [],
        createdAt = createdAt ?? '';

  factory CartModel.fromJson(Map<String, dynamic> json) =>
      _$CartModelFromJson(json);
  Map<String, dynamic> toJson() => _$CartModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class CartStatistics {
  final int totalQuantity;
  final int uniqueItems;

  CartStatistics({
    required this.totalQuantity,
    required this.uniqueItems,
  });

  factory CartStatistics.fromJson(Map<String, dynamic> json) =>
      _$CartStatisticsFromJson(json);
  Map<String, dynamic> toJson() => _$CartStatisticsToJson(this);
}

@JsonSerializable(explicitToJson: true)
class CartLocation {
  final String code;
  final List<LanguageDataModel> names;

  CartLocation({
    required this.code,
    required this.names,
  });

  factory CartLocation.fromJson(Map<String, dynamic> json) =>
      _$CartLocationFromJson(json);
  Map<String, dynamic> toJson() => _$CartLocationToJson(this);
}

@JsonSerializable(explicitToJson: true)
class CartItem {
  final String barcode;
  final List<LanguageDataModel> names;
  final String itemunitcode;
  final List<LanguageDataModel> itemunitnames;
  final double quantity;
  final String? description;
  final String? imageuri;

  CartItem({
    required this.barcode,
    required this.names,
    required this.itemunitcode,
    required this.itemunitnames,
    required this.quantity,
    String? description,
    String? imageuri,
  })  : description = description ?? '',
        imageuri = imageuri ?? '';

  factory CartItem.fromJson(Map<String, dynamic> json) =>
      _$CartItemFromJson(json);
  Map<String, dynamic> toJson() => _$CartItemToJson(this);
}
