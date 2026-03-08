import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'shop_list_model.g.dart';

@JsonSerializable()
class ShopListModel {
  String shopid;
  String name;
  List<LanguageDataModel>? names;
  String? branchcode;
  int role;
  bool isfavorite;
  String? lastaccessedat;
  String? createdby;

  ShopListModel({
    required this.shopid,
    required this.name,
    this.branchcode,
    required this.role,
    required this.isfavorite,
    this.lastaccessedat,
    this.createdby,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory ShopListModel.fromJson(Map<String, dynamic> json) =>
      _$ShopListModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShopListModelToJson(this);
}
