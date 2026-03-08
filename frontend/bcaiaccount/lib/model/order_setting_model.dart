import 'package:json_annotation/json_annotation.dart';

part 'order_setting_model.g.dart';

@JsonSerializable()
class OrderSettingModel {
  String? guidfixed;
  String? activepin;

  /// รหัสรเครื่อง
  String? code;

  /// หมายเลขเครื่อง
  String? devicenumber;

  /// 0 = เครื่องพนัักงาน , 1 = เครื่องลูกค้า
  int? devicetype;

  /// รูปแบบเลขที่เอกสาร
  String? docformat;

  bool? isposactive;

  String? settingcode;
  String? settingname;

  OrderSettingModel({
    String? guidfixed,
    String? activepin,
    String? code,
    String? devicenumber,
    int? devicetype,
    String? docformat,
    bool? isposactive,
    String? settingcode,
    String? settingname,
  })  : guidfixed = guidfixed ?? '',
        activepin = activepin ?? '',
        code = code ?? '',
        devicenumber = devicenumber ?? '',
        devicetype = devicetype ?? 0,
        docformat = docformat ?? '',
        isposactive = isposactive ?? false,
        settingcode = settingcode ?? '',
        settingname = settingname ?? '';

  factory OrderSettingModel.fromJson(Map<String, dynamic> json) =>
      _$OrderSettingModelFromJson(json);
  Map<String, dynamic> toJson() => _$OrderSettingModelToJson(this);
}
