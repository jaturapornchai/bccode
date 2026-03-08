import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/option_model.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'inventory_model.g.dart';

@JsonSerializable()
class InventoryModel {
  String groupcode;
  String guidfixed;
  String itemguid;
  String barcode;
  String name1;
  String name2;
  String name3;
  String name4;
  String name5;
  String description1;
  String description2;
  String description3;
  String description4;
  String description5;
  String itemcode;
  String itemunitcode;
  double itemunitstd = 0.0;
  double itemunitdiv = 0.0;
  String unitname1;
  String unitname2;
  String unitname3;
  String unitname4;
  String unitname5;
  double price = 0.0;
  List<OptionModel> options = <OptionModel>[];
  List<ImageUpload> images = <ImageUpload>[];

  InventoryModel({
    String? groupcode,
    String? guidfixed,
    String? itemguid,
    required this.barcode,
    required this.name1,
    String? name2,
    String? name3,
    String? name4,
    String? name5,
    String? description1,
    String? description2,
    String? description3,
    String? description4,
    String? description5,
    String? itemcode,
    String? itemunitcode,
    double itemunitstd = 0.0,
    double itemunitdiv = 0.0,
    String? unitname1,
    String? unitname2,
    String? unitname3,
    String? unitname4,
    String? unitname5,
    double price = 0.0,
    List<OptionModel>? options,
    List<ImageUpload>? images,
  })  : guidfixed = guidfixed ?? '',
        groupcode = groupcode ?? '',
        itemguid = itemguid ?? '',
        name2 = name2 ?? '',
        name3 = name3 ?? '',
        name4 = name4 ?? '',
        name5 = name5 ?? '',
        description1 = description1 ?? '',
        description2 = description2 ?? '',
        description3 = description3 ?? '',
        description4 = description4 ?? '',
        description5 = description5 ?? '',
        itemcode = itemcode ?? '',
        itemunitcode = itemunitcode ?? '',
        unitname1 = unitname1 ?? '',
        unitname2 = unitname2 ?? '',
        unitname3 = unitname3 ?? '',
        unitname4 = unitname4 ?? '',
        unitname5 = unitname5 ?? '',
        options = options ?? <OptionModel>[],
        images = images ?? <ImageUpload>[];

  factory InventoryModel.fromJson(Map<String, dynamic> json) {
    // ทำความสะอาดข้อมูล options และ images ก่อน parse
    // แก้ปัญหากรณีที่ API ส่งมาเป็น string ว่าง หรือ null แทนที่จะเป็น array
    try {
      // ตรวจสอบและแก้ไข options field
      if (json['options'] == null || json['options'] == '' || json['options'] is! List) {
        json['options'] = [];
      }

      // ตรวจสอบและแก้ไข images field
      if (json['images'] == null || json['images'] == '' || json['images'] is! List) {
        json['images'] = [];
      }

      return _$InventoryModelFromJson(json);
    } catch (e) {
      AppLogger.error('❌ ERROR in InventoryModel.fromJson: $e');
      AppLogger.error('JSON data: $json');
      rethrow;
    }
  }

  Map<String, dynamic> toJson() => _$InventoryModelToJson(this);
}
