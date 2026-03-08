import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'location_model.g.dart';

@JsonSerializable(explicitToJson: true)
class LocationModel {
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  List<ShelfModel> shelf = <ShelfModel>[];

  LocationModel({
    required this.code,
    List<LanguageDataModel>? names,
    List<ShelfModel>? shelf,
  })  : names = names ?? <LanguageDataModel>[],
        shelf = shelf ?? <ShelfModel>[];
  factory LocationModel.fromJson(Map<String, dynamic> json) =>
      _$LocationModelFromJson(json);
  Map<String, dynamic> toJson() => _$LocationModelToJson(this);
}
