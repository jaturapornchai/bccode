import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'dimension_model.g.dart';

@JsonSerializable()
class DimensionModel {
  String? guidfixed;
  bool? isdisabled;
  List<ItemDimension>? items;
  List<LanguageDataModel>? names;

  DimensionModel({
    String? guidfixed,
    bool? isdisabled,
    List<ItemDimension>? items,
    List<LanguageDataModel>? names,
  })  : isdisabled = isdisabled ?? false,
        items = items ?? [],
        names = names ?? [],
        guidfixed = guidfixed ?? '';

  factory DimensionModel.fromJson(Map<String, dynamic> json) =>
      _$DimensionModelFromJson(json);
  Map<String, dynamic> toJson() => _$DimensionModelToJson(this);
}

@JsonSerializable()
class DimensionProductModel {
  String? guidfixed;
  bool? isdisabled;
  ItemDimension? item;
  List<LanguageDataModel>? names;

  DimensionProductModel({
    String? guidfixed,
    bool? isdisabled,
    ItemDimension? item,
    List<LanguageDataModel>? names,
  })  : isdisabled = isdisabled ?? false,
        names = names ?? [],
        guidfixed = guidfixed ?? '',
        item = item ?? ItemDimension();

  factory DimensionProductModel.fromJson(Map<String, dynamic> json) =>
      _$DimensionProductModelFromJson(json);
  Map<String, dynamic> toJson() => _$DimensionProductModelToJson(this);
}

@JsonSerializable()
class ItemDimension {
  String? guidfixed;
  bool? isdisabled;
  List<LanguageDataModel>? names;

  ItemDimension({
    String? guidfixed,
    bool? isdisabled,
    List<LanguageDataModel>? names,
  })  : guidfixed = guidfixed ?? '',
        isdisabled = isdisabled ?? false,
        names = names ?? [];

  factory ItemDimension.fromJson(Map<String, dynamic> json) =>
      _$ItemDimensionFromJson(json);
  Map<String, dynamic> toJson() => _$ItemDimensionToJson(this);
}
