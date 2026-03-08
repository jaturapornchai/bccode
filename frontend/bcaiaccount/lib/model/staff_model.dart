import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'staff_model.g.dart';

@JsonSerializable()
class StaffModel {
  late String guidfixed;
  late String code;
  late String email;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  late bool cashier;
  late bool order;

  StaffModel({
    required this.guidfixed,
    required this.code,
    required this.email,
    required this.names,
    required this.cashier,
    required this.order,
  });

  factory StaffModel.fromJson(Map<String, dynamic> json) =>
      _$StaffModelFromJson(json);
  Map<String, dynamic> toJson() => _$StaffModelToJson(this);
}
