import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'product_category_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ProductCategoryModel {
  String guidfixed; //อ้างอิง
  String parentguid; // อ้างอิงตก่อนหน้า (ตัวแม่)
  String parentguidall; // อ้างอิงทั้งหมด (มีคอมม่าขั้น)
  List<LanguageDataModel>? names; // ชื่อ (หลายภาษา)
  int childcount; // จำนวนลูก
  String imageuri; // รูปภาพ
  bool useimageorcolor; // True=Image,False=Color
  bool isdisabled; // True=ปิด,False=เปิด
  String colorselect; // สีที่เลือก
  String colorselecthex; // สีที่เลือก (Hex)
  List<SortDataModel>? xsorts; // ลำดับการเรียง
  List<ProductCategoryCodeListModel>? codelist; // บาร์โค้ด
  String? coveruri;

  /// หมายเลขหมวด
  int? groupnumber;

  List<TimeForSaleModel>? timeforsales;

  ProductCategoryModel({
    required this.guidfixed,
    required this.parentguid,
    required this.parentguidall,
    required this.names,
    required this.imageuri,
    required this.childcount,
    required this.xsorts,
    required this.useimageorcolor,
    required this.isdisabled,
    required this.codelist,
    required this.colorselect,
    required this.colorselecthex,
    String? coveruri,
    int? groupnumber,
    List<TimeForSaleModel>? timeforsales,
  })  : coveruri = coveruri ?? "",
        groupnumber = groupnumber ?? 1,
        timeforsales = timeforsales ?? [];

  factory ProductCategoryModel.fromJson(Map<String, dynamic> json) =>
      _$ProductCategoryModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductCategoryModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductCategoryCodeListModel {
  String code;
  String barcode;
  List<LanguageDataModel>? names;
  String unitcode;
  List<LanguageDataModel>? unitnames;

  int xorder;

  ProductCategoryCodeListModel(
      {required this.code,
      required this.xorder,
      required this.barcode,
      required this.names,
      required this.unitnames,
      required this.unitcode});

  factory ProductCategoryCodeListModel.fromJson(Map<String, dynamic> json) =>
      _$ProductCategoryCodeListModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductCategoryCodeListModelToJson(this);
}

@JsonSerializable()
class TimeForSaleModel {
  List<int>? daysofweek;
  String? fromdate;
  String? todate;
  String? fromtime;
  String? totime;

  TimeForSaleModel({
    List<int>? daysofweek,
    String? fromdate,
    String? todate,
    String? fromtime,
    String? totime,
  })  : daysofweek = daysofweek ?? [],
        fromdate = fromdate ?? "",
        todate = todate ?? "",
        fromtime = fromtime ?? "",
        totime = totime ?? "";

  factory TimeForSaleModel.fromJson(Map<String, dynamic> json) =>
      _$TimeForSaleModelFromJson(json);
  Map<String, dynamic> toJson() => _$TimeForSaleModelToJson(this);
}
