import 'package:smlaicloud/model/create_shop_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'shop_model.g.dart';

@JsonSerializable()
class ShopModel {
  String? guidfixed;
  List<LanguageDataModel>? address;
  String? branchcode;
  List<ImagesModel>? images;
  String? logo;
  String? name1;
  List<LanguageDataModel>? names;
  String? profilepicture;
  Settings? settings;
  String? telephone;
  bool? ismainshop;
  // ค้นหา สินค้า ตามร้านหลัก (สามารถเพิ่มได้หรือไม่ได้)
  // 0 = ไม่แสดงร้านหลัก ,  1 = แสดงร้านหลัก(แสดงร้านค้าหลัก และ ไม่สามารถเพิ่มในร้านตัวเองได้) , 2 = แสดงร้านหลัก(แสดงร้านค้าหลัก และ สามารถเพิ่มในร้านตัวเองได้)
  int? productcentertype;
  int? debtorcentertype;
  String? mainshopid;
  // ค้นหา สินค้า ตามร้านหลัก
  // 0 = ไม่แสดงร้านหลัก ,  1 = แสดงร้านหลัก
  int? posproductcentertype;
  String? apikey;

  ShopModel({
    String? guidfixed,
    List<LanguageDataModel>? address,
    String? branchcode,
    List<ImagesModel>? images,
    String? logo,
    String? name1,
    List<LanguageDataModel>? names,
    String? profilepicture,
    Settings? settings,
    String? telephone,
    bool? ismainshop,
    int? productcentertype,
    int? debtorcentertype,
    String? mainshopid,
    int? posproductcentertype,
    String? apikey,
  })  : address = address ?? [],
        branchcode = branchcode ?? '',
        images = images ?? [],
        logo = logo ?? '',
        name1 = name1 ?? '',
        names = names ?? [],
        profilepicture = profilepicture ?? '',
        settings = settings ?? Settings(),
        telephone = telephone ?? '',
        guidfixed = guidfixed ?? '',
        ismainshop = ismainshop ?? false,
        productcentertype = productcentertype ?? 0,
        debtorcentertype = debtorcentertype ?? 0,
        mainshopid = mainshopid ?? '',
        posproductcentertype = posproductcentertype ?? 0,
        apikey = apikey ?? '';

  factory ShopModel.fromJson(Map<String, dynamic> json) =>
      _$ShopModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShopModelToJson(this);
}
