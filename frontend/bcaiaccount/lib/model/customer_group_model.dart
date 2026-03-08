import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'customer_group_model.g.dart';

/// test
@JsonSerializable(explicitToJson: true)
class CustomerGroupModel {
  String guidfixed;

  /// รหัสกลุ่มลูกค้า
  String groupcode;

  /// ชื่อกลุ่มลูกค้า
  List<LanguageDataModel> names = <LanguageDataModel>[];

  CustomerGroupModel({
    required this.guidfixed,
    required this.groupcode,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory CustomerGroupModel.fromJson(Map<String, dynamic> json) =>
      _$CustomerGroupModelFromJson(json);

  Map<String, dynamic> toJson() => _$CustomerGroupModelToJson(this);
}
