import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/dimension_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:smlaicloud/model/product_type_model.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'product_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ProductMasterModel {
  String shopid;
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  String groupcode;
  String groupguid;
  List<LanguageDataModel> groupname;
  List<LanguageDataModel> manufacturername;
  String unitguid;
  int itemtype;
  String manufacturercode;
  String manufacturerguid;
  List<DimensionModel> dimensions;
  List<ProductMasterBarcode> barcodes;

  ProductMasterModel({
    String? shopid,
    List<LanguageDataModel>? groupname,
    List<LanguageDataModel>? manufacturername,
    String? manufacturercode,
    String? guidfixed,
    String? groupcode,
    String? code,
    List<LanguageDataModel>? names,
    String? groupguid,
    String? unitguid,
    int? itemtype,
    String? manufacturerguid,
    List<DimensionModel>? dimensions,
    List<ProductMasterBarcode>? barcodes,
  }) : shopid = shopid ?? '',
       manufacturercode = manufacturercode ?? '',
       groupcode = groupcode ?? '',
       guidfixed = guidfixed ?? '',
       code = code ?? '',
       names = names ?? <LanguageDataModel>[],
       groupguid = groupguid ?? '',
       unitguid = unitguid ?? '',
       itemtype = itemtype ?? 0,
       manufacturerguid = manufacturerguid ?? '',
       dimensions = dimensions ?? <DimensionModel>[],
       groupname = groupname ?? <LanguageDataModel>[],
       manufacturername = manufacturername ?? <LanguageDataModel>[],
       barcodes = barcodes ?? <ProductMasterBarcode>[];

  factory ProductMasterModel.fromJson(Map<String, dynamic> json) {
    // ทำความสะอาดข้อมูล dimensions และ barcodes ก่อน parse
    // แก้ปัญหากรณีที่ API ส่งมาเป็น string ว่าง หรือ null แทนที่จะเป็น array
    try {
      // ตรวจสอบและแก้ไข dimensions field
      if (json['dimensions'] == null || json['dimensions'] == '' || json['dimensions'] is! List) {
        json['dimensions'] = [];
      }

      // ตรวจสอบและแก้ไข barcodes field
      if (json['barcodes'] == null || json['barcodes'] == '' || json['barcodes'] is! List) {
        json['barcodes'] = [];
      }

      // ตรวจสอบและแก้ไข names field
      if (json['names'] == null || json['names'] == '' || json['names'] is! List) {
        json['names'] = [];
      }

      // ตรวจสอบและแก้ไข groupname field
      if (json['groupname'] == null || json['groupname'] == '' || json['groupname'] is! List) {
        json['groupname'] = [];
      }

      // ตรวจสอบและแก้ไข manufacturername field
      if (json['manufacturername'] == null || json['manufacturername'] == '' || json['manufacturername'] is! List) {
        json['manufacturername'] = [];
      }

      return _$ProductMasterModelFromJson(json);
    } catch (e) {
      AppLogger.error('❌ ERROR in ProductMasterModel.fromJson: $e');
      AppLogger.error('JSON data: $json');
      rethrow;
    }
  }

  Map<String, dynamic> toJson() => _$ProductMasterModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductMasterBarcode {
  String guidfixed;
  String barcode;
  String itemunitcode;
  List<LanguageDataModel> itemunitnames;
  List<PriceDataModel> prices;
  bool condition;
  double standvalue;
  double dividevalue;
  double qty;
  bool ismainbarcode;

  ProductMasterBarcode({
    String? guidfixed,
    String? barcode,
    String? itemunitcode,
    List<LanguageDataModel>? itemunitnames,
    List<PriceDataModel>? prices,
    bool? condition,
    double? standvalue,
    double? dividevalue,
    double? qty,
    bool? ismainbarcode,
  }) : guidfixed = guidfixed ?? '',
       barcode = barcode ?? '',
       itemunitcode = itemunitcode ?? '',
       itemunitnames = itemunitnames ?? <LanguageDataModel>[],
       prices = prices ?? <PriceDataModel>[],
       condition = condition ?? false,
       standvalue = standvalue ?? 0,
       dividevalue = dividevalue ?? 0,
       qty = qty ?? 0,
       ismainbarcode = ismainbarcode ?? true;

  factory ProductMasterBarcode.fromJson(Map<String, dynamic> json) {
    // ทำความสะอาดข้อมูล list fields ก่อน parse
    try {
      // ตรวจสอบและแก้ไข itemunitnames field
      if (json['itemunitnames'] == null || json['itemunitnames'] == '' || json['itemunitnames'] is! List) {
        json['itemunitnames'] = [];
      }

      // ตรวจสอบและแก้ไข prices field
      if (json['prices'] == null || json['prices'] == '' || json['prices'] is! List) {
        json['prices'] = [];
      }

      return _$ProductMasterBarcodeFromJson(json);
    } catch (e) {
      AppLogger.error('❌ ERROR in ProductMasterBarcode.fromJson: $e');
      AppLogger.error('JSON data: $json');
      rethrow;
    }
  }

  Map<String, dynamic> toJson() => _$ProductMasterBarcodeToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductModel {
  /// guid ฐานข้อมูล
  String guidfixed;

  /// รหัสสินค้า
  String itemcode;

  /// guid กลุ่มสินค้า
  String? groupcode;

  List<LanguageDataModel>? groupnames;

  /// Barcode (หลายตัว)
  List<String>? barcodes;

  /// ชื่อสินค้า (หลายภาษา)
  List<LanguageDataModel>? names;

  /// True=หลายหน่วยนับ,​ False=หน่วยนับเดียว
  bool? multiunit;

  /// True=ใช้ Serial Number,​ False=ไม่ใช้ Serial Number
  bool? useserialnumber;
  // หน่วยนับ (หลายหน่วยนับ)
  List<ProductUnitModel>? units;

  ProductUnitModel? unit;

  /// uri รูปภาพ (หลายรูป)
  List<ImagesModel>? images;
  // หน่วยต้นทุน สำหรับจัดทำบัญชี (เปลี่ยนไม่ได้)
  String? unitcost;

  List<LanguageDataModel>? unitcostnames;

  /// หน่วยมาตราฐาน สำหรับแสดงในรายการค้นหา (เปลี่ยนไปมาได้)
  String? unitstandard;

  List<LanguageDataModel>? unitstandardnames;

  /// ประเภทสินค้า 0=สินค้า, (มีสต๊อก)​ 1=บริการ (ไม่มีสต๊อก)
  int? itemstocktype;

  /// ประเภทสินค้า 0=สินค้าทั่วไป,​ 1=อาหาร,​ 2=เครื่องดื่ม
  int? itemtype;

  /// ประเภทภาษี 0=ภาษีมูลค่าเพิ่ม 1=ยกเว้นภาษี
  int? taxtype;

  /// ภาษี 1=คิดภาษีมูลค่าเพิ่ม, 2=ยกเว้นภาษีมูลค่าเพิ่ม
  int? vattype;

  /// คิดคะแนน False=ไม่คิด,​ True=คิด (สะสมแต้ม)
  bool? issumpoint;

  /// เป็นหน่วยต้นทุน
  @JsonKey(includeFromJson: false, includeToJson: false)
  int idxunitcost = 0;

  /// เป็นหหน่วยมาตราฐาน
  @JsonKey(includeFromJson: false, includeToJson: false)
  int idxunitstandard = 0;

  ProductModel({
    required this.guidfixed,
    required this.itemcode,
    String? groupcode,
    List<LanguageDataModel>? groupnames,
    List<String>? barcodes,
    List<LanguageDataModel>? names,
    bool? useserialnumber,
    List<ProductUnitModel>? units,
    ProductUnitModel? unit,
    List<ImagesModel>? images,
    String? unitcost,
    List<LanguageDataModel>? unitcostnames,
    String? unitstandard,
    List<LanguageDataModel>? unitstandardnames,
    bool? multiunit,
    int? itemstocktype,
    int? vattype,
    int? taxtype,
    bool? issumpoint,
    int? itemtype,
    List<PriceDataModel>? prices,
  }) : groupcode = groupcode ?? '',
       groupnames = groupnames ?? <LanguageDataModel>[],
       barcodes = barcodes ?? <String>[],
       names = names ?? <LanguageDataModel>[],
       units = units ?? <ProductUnitModel>[],
       unit = unit ?? ProductUnitModel(unitcode: "", divider: 0, stand: 0, xorder: 0, stockcount: false),
       images = images ?? <ImagesModel>[],
       unitcost = unitcost ?? '',
       unitcostnames = unitcostnames ?? <LanguageDataModel>[],
       unitstandardnames = unitstandardnames ?? <LanguageDataModel>[],
       unitstandard = unitstandard ?? '',
       multiunit = multiunit ?? false,
       itemstocktype = itemstocktype ?? 0,
       vattype = vattype ?? 0,
       taxtype = taxtype ?? 0,
       issumpoint = issumpoint ?? false,
       itemtype = itemtype ?? 0,
       useserialnumber = useserialnumber ?? false;

  factory ProductModel.fromJson(Map<String, dynamic> json) => _$ProductModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductUnitModel {
  /// รหัสหน่วยนับ
  String unitcode;

  /// ชื่อหน่วยนับ
  List<LanguageDataModel>? names;

  /// ตัวหาร
  double divider;

  /// ตัวตั้ง
  double stand;

  /// ลำดับ
  int xorder;

  /// นับสต๊อก
  bool stockcount;

  @JsonKey(includeFromJson: false, includeToJson: false)
  int smallerOrBigger = 0;

  ProductUnitModel({required this.unitcode, List<LanguageDataModel>? names, required this.divider, required this.stand, required this.xorder, required this.stockcount})
    : names = names ?? <LanguageDataModel>[];

  factory ProductUnitModel.fromJson(Map<String, dynamic> json) => _$ProductUnitModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductUnitModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductUnitPriceModel {
  String unitcode; // รหัสหน่วยนับ
  List<PriceDataModel> prices; // ราคาขาย

  ProductUnitPriceModel({required this.unitcode, required this.prices});

  factory ProductUnitPriceModel.fromJson(Map<String, dynamic> json) => _$ProductUnitPriceModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductUnitPriceModelToJson(this);
}

/// คลังสินค้า
@JsonSerializable(explicitToJson: true)
class ProductWareHouseModel {
  String code; // รหัสคลังสินค้า
  List<ProductLocationModel> locatecodes; // รหัสตำแหน่ง

  ProductWareHouseModel({required this.code, required this.locatecodes});

  factory ProductWareHouseModel.fromJson(Map<String, dynamic> json) => _$ProductWareHouseModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductWareHouseModelToJson(this);
}

/// บริเวณที่เก็บสินค้า
@JsonSerializable(explicitToJson: true)
class ProductLocationModel {
  String code; // รหัสบริเวณ
  List<String> recommendshelfs; // แนะนำที่วาง

  ProductLocationModel({required this.code, required this.recommendshelfs});

  factory ProductLocationModel.fromJson(Map<String, dynamic> json) => _$ProductLocationModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductLocationModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductBarcodeRefModel {
  /// รหัสบาร์โค้ด อ้างอิง
  String? refbarcode;

  /// ชื่อหน่วยนับ Barcode อ้างอิง
  List<LanguageDataModel>? refunitnames;

  /// Barcode ตัดสต๊อก
  String? stockbarcode;

  /// จำนวน Barcode อ้างอิง
  double? qty;

  /// ตัวหาร Barcode อ้างอิง
  double? dividevalue;

  /// ตัวตั้ง Barcode อ้างอิง
  double? standvalue;

  ProductBarcodeRefModel({String? refbarcode, double? qty, List<LanguageDataModel>? unitnames, String? stockbarcode, double? dividevalue, double? standvalue})
    : refbarcode = refbarcode ?? '',
      stockbarcode = stockbarcode ?? '',
      refunitnames = unitnames ?? <LanguageDataModel>[],
      qty = qty ?? 0,
      dividevalue = dividevalue ?? 0,
      standvalue = standvalue ?? 0;
  factory ProductBarcodeRefModel.fromJson(Map<String, dynamic> json) => _$ProductBarcodeRefModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductBarcodeRefModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductBarcodeModel {
  /// guid ฐานข้อมูล
  String guidfixed;

  /// รหัสสินค้า
  String? itemcode;
  String itemguid;

  /// รหัสหน่วยนับ
  String itemunitcode;

  String itemguidfixed;

  bool ismainbarcode;

  /// ชื่อหน่วยนับ
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? itemunitnames;

  /// ประเภทรายการ true=สินค้าหลัก,false=สินค้าเขืื่อม
  bool ismainitem;

  /// guid กลุ่มสินค้า

  String groupguid;
  String? groupcode;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? groupnames;

  /// Barcode
  String? barcode;

  /// ชื่อสินค้า
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? names;

  /// ราคาขาย
  @JsonKey(defaultValue: <PriceDataModel>[])
  List<PriceDataModel>? prices;

  /// uri รูปภาพ
  String? imageuri;

  /// ใช้รูปหรือสี True=Image,False=Color
  bool? useimageorcolor;

  /// สีที่เลือก
  String? colorselect;

  /// สีที่เลือก (Hex)
  String? colorselecthex;

  /// ข้อเลือกสินค้า (เช่น เพิ่มไข่)
  @JsonKey(defaultValue: <ProductOptionModel>[])
  List<ProductOptionModel>? options;

  //บาร์โค้ดหลัก
  String? parentguid;

  /// ประเภทสินค้า 0=สินค้ามีสต็อก,​ 1=สินค้าบริการ,​ 2=สินค้าชุด ,3=สินค้าวัตถุดิบ , 4=สินค้ากึ่งสำเร็จรูป , 5=สินค้าไม่นับสต็อก
  int? itemtype;

  /// ประเภทภาษี 0=ภาษีมูลค่าเพิ่ม 1=ยกเว้นภาษี
  int? taxtype;

  /// ภาษี 0=คิดภาษีมูลค่าเพิ่ม, 1=ยกเว้นภาษีมูลค่าเพิ่ม
  int? vattype;

  /// คิดคะแนน False=ไม่คิด,​ True=คิด (สะสมแต้ม)
  bool? issumpoint;

  //ส่วนลดสูงสุด
  String? maxdiscount;

  //สินค้าร่วมรายการสะสมยอดขายเพื่อรับเงินปันผล
  bool? isdividend;

  // บาร์โค้ดย่อย
  @JsonKey(defaultValue: <BarCodeSubModel>[])
  List<BarCodeSubModel>? barcodes;

  bool? isusesubbarcodes;

  bool? condition;

  /// ตัวตั้ง
  double? standvalue;

  /// ตัวหาร
  double? dividevalue;

  int? vatcal;

  /// บาร์โค้ดอ้างอิง
  @JsonKey(defaultValue: <BarCodeSubModel>[])
  List<BarCodeSubModel>? refbarcodes;

  /// สูตรผลิต
  @JsonKey(defaultValue: <BomModel>[])
  List<BomModel>? bom;

  /// อลาคาร์ท
  bool? isalacarte;

  /// ประเภทสั่งอาหาร
  @JsonKey(defaultValue: <OrderTypeProductBarcode>[])
  List<OrderTypeProductBarcode>? ordertypes;

  /// พิมพ์ใบจัดอาหารแบบแยก
  bool? issplitunitprint;

  /// แสดงเฉพาะพนักงาน
  bool? isonlystaff;

  /// ประเภทสินค้า
  ProductTypeModel? producttype;

  /// ประเภทอาหาร
  int? foodtype;

  /// ใช้ใน impart_product
  int? showisdividend;

  /// นำเข้าข้อมูลสินค้า ลำดับ
  int? rownumber;

  /// ส่วนลด
  String? discount;

  /// สต๊อกสำหรับร้านอาหาร
  bool? isstockforrestaurant;

  /// ผู้ผลิต save to service
  String? manufacturerguid;

  String? manufacturercode;

  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? manufacturernames;

  /// มิติ
  @JsonKey(defaultValue: <DimensionProductModel>[])
  List<DimensionProductModel>? dimensions;

  /// ประเภทธุรกิจ
  @JsonKey(defaultValue: <BusinessTypeModel>[])
  List<BusinessTypeModel>? businesstypes;

  /// สาขาที่ไม่ใช้
  @JsonKey(defaultValue: <BranchModel>[])
  List<BranchModel>? ignorebranches;

  /// สาขา ใช้ใน ui
  @JsonKey(defaultValue: <CompanyBranchModel>[])
  List<CompanyBranchModel>? branches;

  /// ลดราคา ณ จุดขาย
  bool? isdiscountpointofpurchase;

  /// ร้านอาหาร
  ProductBarcodeRestaurantModel? restaurant;

  /// ใช้ระบบแจ้งเตือน
  bool? isalert;

  /// รายละเอียดแจ้งเตือน
  String? alertdescription;

  /// รายละเอียดสินค้า
  String? description;

  /// isdisble
  bool? isdisable;

  /// categorys
  @JsonKey(defaultValue: <ProductCategoryModel>[])
  List<ProductCategoryModel>? categorys;

  /// เวลาขายสินค้า
  @JsonKey(defaultValue: <TimeForSaleModel>[])
  List<TimeForSaleModel>? timeforsales;

  /// ต้นทุนมาตราฐาน
  @JsonKey(defaultValue: <FiexdCostModel>[])
  List<FiexdCostModel>? fixedcost;

  /// ประเภทวัตถุดิบ
  /// 0 = ทั่วไป, 1 = วัตถุดิบ , 2 = วัตถุดิบกึ่งสำเร็จรูป
  int? materialtype;

  /// Brand fields
  String? brandcode;
  String? brandguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? brandnames;

  /// Category fields
  String? categorycode;
  String? categoryguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? categorynames;

  /// Class fields
  String? classcode;
  String? classguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? classnames;

  /// Design fields
  String? designcode;
  String? designguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? designnames;

  /// Grade fields
  String? gradecode;
  String? gradeguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? gradenames;

  /// Model fields
  String? modelcode;
  String? modelguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? modelnames;

  /// Group Sub One fields
  String? groupsubonecode;
  String? groupsuboneguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? groupsubonenames;

  /// Group Sub Two fields
  String? groupsubtwocode;
  String? groupsubtwoguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? groupsubtwonames;

  /// Pattern fields
  String? patterncode;
  String? patternguid;
  @JsonKey(defaultValue: <LanguageDataModel>[])
  List<LanguageDataModel>? patternnames;

  /// Shelf fields - เพิ่มสำหรับข้อมูลชั้นวาง
  String? shelfCode;
  String? shelfName;

  /// จำนวนหน่วยนับ (จาก PG list) — 1=หน่วยเดียว, >1=หลายหน่วย
  @JsonKey(name: 'unit_count', defaultValue: 0)
  int unitCount;

  /// รายชื่อหน่วยนับทั้งหมด คั่นด้วย ", " (จาก PG list)
  @JsonKey(name: 'all_unit_names', defaultValue: '')
  String allUnitNames;

  /// ยอดคงเหลือ (จาก PG list)
  @JsonKey(name: 'balance_qty', defaultValue: 0.0)
  double balanceQty;

  /// ยอดคงเหลือ formatted ด้วย auto packing (เช่น "2 ลัง/5 โหล/3 ชิ้น")
  @JsonKey(name: 'balance_formatted', defaultValue: '')
  String balanceFormatted;

  ProductBarcodeModel({
    required this.guidfixed,
    String? groupguid,
    String? itemguid,
    bool? ismainbarcode,
    String? itemguidfixed,

    /// รหัสสินค้า สามารถซ้ำได้ (กรณีมี Barcode หลายตัว)
    String? itemcode,

    /// รหัสบาร์โค้ด อ้างอิง
    String? refbarcode,

    /// ชื่อหน่วยนับ Barcode อ้างอิง
    List<LanguageDataModel>? refunitnames,

    /// 1=ใหญ่กว่า,2=เล็กกว่า
    int? refpackstyle,

    /// guid กลุ่มสินค้า
    String? groupcode,
    bool? ismainitem,

    /// ชื่อกลุ่มสินค้า
    List<LanguageDataModel>? groupnames,

    /// Barcode
    String? barcode,

    /// ชื่อสินค้า
    List<LanguageDataModel>? names,

    /// รหัสหน่วยนับ
    String? itemunitcode,

    /// ชื่อหน่วยนับ
    List<LanguageDataModel>? itemunitnames,

    /// ราคาขาย
    List<PriceDataModel>? prices,

    /// uri รูปภาพ
    String? imageuri,

    /// ใช้รูปหรือสี True=Image,False=Color
    bool? useimageorcolor,

    /// สีที่เลือก
    String? colorselect,

    /// สีที่เลือก (Hex)
    String? colorselecthex,

    /// ข้อเลือกสินค้า (เช่น เพิ่มไข่)
    List<ProductOptionModel>? options,

    /// บาร์โค้ดหลัก
    String? parentguid,

    ///
    List<BarCodeSubModel>? barcodes,

    /// ใช้บาร์โค้ดย่อย
    bool? isusesubbarcodes,

    /// ตัวตั้ง
    double? standvalue,

    /// ตัวหาร
    double? dividevalue,

    /// เงื่อนไข
    bool? condition,

    /// ประเภทสินค้า 0=สินค้าทั่วไป,​ 1=อาหาร,​ 2=เครื่องดื่ม
    int? itemtype,

    /// ประเภทภาษี 0=ภาษีมูลค่าเพิ่ม 1=ยกเว้นภาษี
    int? taxtype,

    /// ภาษี 0=คิดภาษีมูลค่าเพิ่ม, 1=ยกเว้นภาษีมูลค่าเพิ่ม
    int? vattype,

    /// คิดคะแนน False=ไม่คิด,​ True=คิด (สะสมแต้ม)
    bool? issumpoint,

    /// ส่วนลดสูงสุด
    String? maxdiscount,

    ///
    bool? isdividend,
    int? vatcal,

    /// จำนวนในหน่วยนับ

    /// บาร์โค้ดอ้างอิง
    List<BarCodeSubModel>? refbarcodes,

    /// สูตรผลิต
    List<BomModel>? bom,

    /// อลาคาร์ท
    bool? isalacarte,

    /// ประเภทสั่งอาหาร
    List<OrderTypeProductBarcode>? ordertypes,

    /// พิมพ์ใบจัดอาหารแบบแยก
    bool? issplitunitprint,

    /// แสดงเฉพาะพนักงาน
    bool? isonlystaff,

    /// ประเภทสินค้า
    ProductTypeModel? producttype,

    /// ประเภทอาหาร
    int? foodtype,

    /// ใช้ใน impart_product
    int? showisdividend,

    /// นำเข้าข้อมูลสินค้า ลำดับ
    int? rownumber,

    /// ส่วนลด
    String? discount,

    /// สต๊อกสำหรับร้านอาหาร
    bool? isstockforrestaurant,

    /// ผู้ผลิต save to service
    String? manufacturerguid,
    String? manufacturercode,
    List<LanguageDataModel>? manufacturernames,

    /// มิติ
    List<DimensionProductModel>? dimensions,

    /// ประเภทธุรกิจ
    List<BusinessTypeModel>? businesstypes,

    /// สาขา
    List<BranchModel>? ignorebranches,

    /// สาขา ใช้ใน ui
    List<CompanyBranchModel>? branches,

    /// ลดราคา ณ จุดขาย
    bool? isdiscountpointofpurchase,

    /// ร้านอาหาร
    ProductBarcodeRestaurantModel? restaurant,

    /// ใช้ระบบแจ้งเตือน
    bool? isalert,

    /// รายละเอียดแจ้งเตือน
    String? alertdescription,

    /// รายละเอียดสินค้า
    String? description,

    /// disable
    bool? isdisable,

    /// categorys
    List<ProductCategoryModel>? categorys,

    /// เวลาขายสินค้า
    List<TimeForSaleModel>? timeforsales,

    /// ต้นทุนมาตราฐาน
    List<FiexdCostModel>? fixedcost,

    /// ประเภทวัตถุดิบ
    int? materialtype,

    /// Brand parameters
    String? brandcode,
    String? brandguid,
    List<LanguageDataModel>? brandnames,

    /// Category parameters
    String? categorycode,
    String? categoryguid,
    List<LanguageDataModel>? categorynames,

    /// Class parameters
    String? classcode,
    String? classguid,
    List<LanguageDataModel>? classnames,

    /// Design parameters
    String? designcode,
    String? designguid,
    List<LanguageDataModel>? designnames,

    /// Grade parameters
    String? gradecode,
    String? gradeguid,
    List<LanguageDataModel>? gradenames,

    /// Model parameters
    String? modelcode,
    String? modelguid,
    List<LanguageDataModel>? modelnames,

    /// Group Sub One parameters
    String? groupsubonecode,
    String? groupsuboneguid,
    List<LanguageDataModel>? groupsubonenames,

    /// Group Sub Two parameters
    String? groupsubtwocode,
    String? groupsubtwoguid,
    List<LanguageDataModel>? groupsubtwonames,

    /// Pattern parameters
    String? patterncode,
    String? patternguid,
    List<LanguageDataModel>? patternnames,

    /// Shelf parameters - เพิ่มสำหรับข้อมูลชั้นวาง
    this.shelfCode,
    this.shelfName,

    /// Unit info (จาก PG list)
    int? unitCount,
    String? allUnitNames,

    /// Balance info (จาก PG list)
    double? balanceQty,
    String? balanceFormatted,
  }) : groupcode = groupcode ?? "",
       ismainbarcode = ismainbarcode ?? true,
       itemguid = itemguid ?? "",
       groupguid = groupguid ?? "",
       itemguidfixed = itemguidfixed ?? "",
       itemcode = itemcode ?? "",
       groupnames = groupnames ?? <LanguageDataModel>[],
       barcode = barcode ?? "",
       names = names ?? <LanguageDataModel>[],
       itemunitcode = itemunitcode ?? "",
       itemunitnames = itemunitnames ?? <LanguageDataModel>[],
       prices = prices ?? <PriceDataModel>[],
       imageuri = imageuri ?? "",
       useimageorcolor = useimageorcolor ?? false,
       isusesubbarcodes = isusesubbarcodes ?? false,
       colorselect = colorselect ?? "",
       colorselecthex = colorselecthex ?? "",
       options = options ?? <ProductOptionModel>[],
       parentguid = parentguid ?? "",
       barcodes = barcodes ?? <BarCodeSubModel>[],
       condition = condition ?? false,
       itemtype = itemtype ?? 0,
       taxtype = taxtype ?? 0,
       vattype = vattype ?? 0,
       issumpoint = issumpoint ?? true,
       maxdiscount = maxdiscount ?? '',
       ismainitem = ismainitem ?? false,
       refbarcodes = refbarcodes ?? <BarCodeSubModel>[],
       bom = bom ?? <BomModel>[],
       isdividend = isdividend ?? false,
       standvalue = standvalue ?? 0,
       dividevalue = dividevalue ?? 0,
       vatcal = vatcal ?? 0,
       isalacarte = isalacarte ?? true,
       ordertypes = ordertypes ?? <OrderTypeProductBarcode>[],
       issplitunitprint = issplitunitprint ?? true,
       isonlystaff = isonlystaff ?? false,
       producttype = producttype ?? ProductTypeModel(code: "", names: <LanguageDataModel>[], guidfixed: ''),
       foodtype = foodtype ?? 0,
       showisdividend = showisdividend ?? 0,
       rownumber = rownumber ?? 0,
       discount = discount ?? '',
       isstockforrestaurant = isstockforrestaurant ?? false,
       manufacturerguid = manufacturerguid ?? '',
       manufacturercode = manufacturercode ?? '',
       manufacturernames = manufacturernames ?? <LanguageDataModel>[],
       dimensions = dimensions ?? <DimensionProductModel>[],
       businesstypes = businesstypes ?? <BusinessTypeModel>[],
       ignorebranches = ignorebranches ?? <BranchModel>[],
       branches = branches ?? <CompanyBranchModel>[],
       isdiscountpointofpurchase = isdiscountpointofpurchase ?? false,
       restaurant = restaurant ?? ProductBarcodeRestaurantModel(),
       isalert = isalert ?? false,
       alertdescription = alertdescription ?? '',
       description = description ?? '',
       isdisable = isdisable ?? false,
       categorys = categorys ?? <ProductCategoryModel>[],
       timeforsales = timeforsales ?? <TimeForSaleModel>[],
       fixedcost = fixedcost ?? <FiexdCostModel>[],
       materialtype = materialtype ?? 0,
       brandcode = brandcode ?? '',
       brandguid = brandguid ?? '',
       brandnames = brandnames ?? <LanguageDataModel>[],
       categorycode = categorycode ?? '',
       categoryguid = categoryguid ?? '',
       categorynames = categorynames ?? <LanguageDataModel>[],
       classcode = classcode ?? '',
       classguid = classguid ?? '',
       classnames = classnames ?? <LanguageDataModel>[],
       designcode = designcode ?? '',
       designguid = designguid ?? '',
       designnames = designnames ?? <LanguageDataModel>[],
       gradecode = gradecode ?? '',
       gradeguid = gradeguid ?? '',
       gradenames = gradenames ?? <LanguageDataModel>[],
       modelcode = modelcode ?? '',
       modelguid = modelguid ?? '',
       modelnames = modelnames ?? <LanguageDataModel>[],
       groupsubonecode = groupsubonecode ?? '',
       groupsuboneguid = groupsuboneguid ?? '',
       groupsubonenames = groupsubonenames ?? <LanguageDataModel>[],
       groupsubtwocode = groupsubtwocode ?? '',
       groupsubtwoguid = groupsubtwoguid ?? '',
       groupsubtwonames = groupsubtwonames ?? <LanguageDataModel>[],
       patterncode = patterncode ?? '',
       patternguid = patternguid ?? '',
       patternnames = patternnames ?? <LanguageDataModel>[],
       unitCount = unitCount ?? 0,
       allUnitNames = allUnitNames ?? '',
       balanceQty = balanceQty ?? 0.0,
       balanceFormatted = balanceFormatted ?? '';

  factory ProductBarcodeModel.fromJson(Map<String, dynamic> json) {
    // Clean up null values in list fields before parsing
    // แก้ปัญหากรณีที่ API ส่ง null แทนที่จะเป็น array ว่าง
    try {
      final cleanedJson = Map<String, dynamic>.from(json);

      // List of fields that should be arrays, not null
      const listFields = [
        'itemunitnames',
        'groupnames',
        'names',
        'prices',
        'options',
        'images', // ← เพิ่ม images เพื่อแก้ปัญหา null
        'barcodes',
        'refbarcodes',
        'bom',
        'ordertypes',
        'manufacturernames',
        'dimensions',
        'businesstypes',
        'ignorebranches',
        'branches',
        'categorys',
        'timeforsales',
        'fixedcost',
        'brandnames',
        'categorynames',
        'classnames',
        'designnames',
        'gradenames',
        'modelnames',
        'groupsubonenames',
        'groupsubtwonames',
        'patternnames',
        'xsorts',
      ];

      // Replace null with empty array for all list fields
      int nullFieldsCount = 0;
      for (final field in listFields) {
        if (cleanedJson[field] == null || cleanedJson[field] == '') {
          cleanedJson[field] = [];
          nullFieldsCount++;
        }
      }

      if (nullFieldsCount > 0) {
        AppLogger.debug('🔧 ProductBarcodeModel: แปลง $nullFieldsCount null fields → []');
      }

      final model = _$ProductBarcodeModelFromJson(cleanedJson);

      // Unit info + Balance จาก PG list (ไม่อยู่ใน .g.dart)
      model.unitCount = cleanedJson['unit_count'] as int? ?? 0;
      model.allUnitNames = cleanedJson['all_unit_names'] as String? ?? '';
      model.balanceQty = (cleanedJson['balance_qty'] as num?)?.toDouble() ?? 0.0;
      model.balanceFormatted = cleanedJson['balance_formatted'] as String? ?? '';

      return model;
    } catch (e) {
      AppLogger.error('❌ ERROR in ProductBarcodeModel.fromJson: $e');
      AppLogger.error('JSON data: $json');
      rethrow;
    }
  }

  Map<String, dynamic> toJson() => _$ProductBarcodeModelToJson(this);
}

//barcodeย่อย
@JsonSerializable(explicitToJson: true)
class BarCodeSubModel {
  String? guidfixed;
  String barcode;
  List<LanguageDataModel> names;
  List<LanguageDataModel> itemunitnames;
  String itemunitcode;
  bool condition;
  double standvalue;
  double dividevalue;
  double qty;

  BarCodeSubModel({
    String? guidfixed,
    required this.barcode,
    List<LanguageDataModel>? names,
    required this.itemunitcode,
    List<LanguageDataModel>? itemunitnames,
    required this.condition,
    required this.standvalue,
    required this.dividevalue,
    required this.qty,
  }) : guidfixed = guidfixed ?? "",
       names = names ?? <LanguageDataModel>[],
       itemunitnames = itemunitnames ?? <LanguageDataModel>[];

  factory BarCodeSubModel.fromJson(Map<String, dynamic> json) {
    // Clean up null values in list fields
    final cleanedJson = Map<String, dynamic>.from(json);
    if (cleanedJson['names'] == null) cleanedJson['names'] = [];
    if (cleanedJson['itemunitnames'] == null) cleanedJson['itemunitnames'] = [];

    return _$BarCodeSubModelFromJson(cleanedJson);
  }

  Map<String, dynamic> toJson() => _$BarCodeSubModelToJson(this);
}

//BOOM
@JsonSerializable(explicitToJson: true)
class BomModel {
  String? guidfixed;
  String barcode;
  List<LanguageDataModel> names;
  List<LanguageDataModel> itemunitnames;
  String itemunitcode;
  bool condition;
  double standvalue;
  double dividevalue;
  double qty;

  BomModel({
    String? guidfixed,
    required this.barcode,
    List<LanguageDataModel>? names,
    required this.itemunitcode,
    List<LanguageDataModel>? itemunitnames,
    required this.condition,
    required this.standvalue,
    required this.dividevalue,
    required this.qty,
  }) : guidfixed = guidfixed ?? "",
       names = names ?? <LanguageDataModel>[],
       itemunitnames = itemunitnames ?? <LanguageDataModel>[];

  factory BomModel.fromJson(Map<String, dynamic> json) {
    // Clean up null values in list fields
    final cleanedJson = Map<String, dynamic>.from(json);
    if (cleanedJson['names'] == null) cleanedJson['names'] = [];
    if (cleanedJson['itemunitnames'] == null) cleanedJson['itemunitnames'] = [];

    return _$BomModelFromJson(cleanedJson);
  }

  Map<String, dynamic> toJson() => _$BomModelToJson(this);
}

/// ข้อเลือกย่อยสินค้า
@JsonSerializable(explicitToJson: true)
class ProductChoiceModel {
  String guid;

  /// ชื่อข้อเลือกย่อย
  List<LanguageDataModel> names;

  /// ราคาข้อเลือกย่อย (เพิ่ม) บาท/เปอร์เซ็นต์
  String price;

  /// ตัดสต๊อกสินค้า
  bool isstock;

  /// อ้างอิง Barcode
  String refbarcode;

  /// อ้างอิงสินค้า เพื่อตัดสต๊อก
  String refproductcode;

  /// อ้างอิงหน่วยนับ เพื่อตัดสต๊อก
  String refunitcode;

  List<LanguageDataModel>? refunitnames = [];

  /// จำนวนเพื่อตัดสต๊อก
  double qty;

  String? imageuri;

  /// เลือกให้เลย
  bool? isdefault;

  /// อ้างอิงชื่อสินค้า
  List<LanguageDataModel>? refbarcodenames;

  /// ประเภทภาษี
  /// 0 = ภาษีมูลค่าเพิ่ม , 1 = ยกเว้นภาษี
  int? vatcal;

  ProductChoiceModel({
    required this.guid,
    required this.refbarcode,
    required this.refproductcode,
    required this.refunitcode,
    required this.names,
    required this.isstock,
    required this.price,
    required this.qty,
    String? imageuri,
    bool? isdefault,
    List<LanguageDataModel>? refbarcodenames,
    List<LanguageDataModel>? refunitnames,
    int? vatcal,
  }) : imageuri = imageuri ?? '',
       isdefault = isdefault ?? false,
       refbarcodenames = refbarcodenames ?? <LanguageDataModel>[],
       refunitnames = refunitnames ?? <LanguageDataModel>[],
       vatcal = vatcal ?? 0;

  factory ProductChoiceModel.fromJson(Map<String, dynamic> json) => _$ProductChoiceModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductChoiceModelToJson(this);
}

/// ข้อเลือกพิเศษ
@JsonSerializable(explicitToJson: true)
class ProductOptionModel {
  String guid;

  /// ประเภทข้อเลือก (0=Check Box,1=Radio Button)
  int choicetype = 0;

  /// เลือกได้น้อยสุด
  int minselect = 0;

  /// เลือกได้สูงสุด
  int maxselect = 0;

  /// ชื่อข้อเลือกพิเศษ
  List<LanguageDataModel> names;

  /// รายการข้อเลือกย่อย
  List<ProductChoiceModel> choices;

  ProductOptionModel({required this.guid, required this.choicetype, required this.maxselect, required this.names, required this.minselect, required this.choices});

  factory ProductOptionModel.fromJson(Map<String, dynamic> json) => _$ProductOptionModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductOptionModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class UnitModel {
  // รหัสหน่วยนับทัทั้งหมด
  String unitcode;
  String guidfixed;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  UnitModel({String? unitcode, String? guidfixed, List<LanguageDataModel>? names}) : names = names ?? <LanguageDataModel>[], unitcode = unitcode ?? '', guidfixed = guidfixed ?? '';

  factory UnitModel.fromJson(Map<String, dynamic> json) => _$UnitModelFromJson(json);

  Map<String, dynamic> toJson() => _$UnitModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class OrderTypeProductBarcode {
  String guidfixed;
  String code;
  List<LanguageDataModel> names;
  double price;

  OrderTypeProductBarcode({required this.guidfixed, required this.code, required this.names, required this.price});

  factory OrderTypeProductBarcode.fromJson(Map<String, dynamic> json) => _$OrderTypeProductBarcodeFromJson(json);

  Map<String, dynamic> toJson() => _$OrderTypeProductBarcodeToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ProductBarcodeRestaurantModel {
  /// ทานที่ร้าน
  bool? isforrestaurant;

  /// ทานกลับบ้าน
  bool? isfortakeaway;

  /// จัดส่ง
  bool? isfordelivery;

  /// สำหรับลูกค้าสามารถสั่งได้
  bool? isforcustomer;

  /// ลูกค้าสามารถ preorder ได้
  bool? isforcustomerpreorder;

  ProductBarcodeRestaurantModel({bool? isforrestaurant, bool? isfortakeaway, bool? isfordelivery, bool? isforcustomer, bool? isforcustomerpreorder})
    : isforrestaurant = isforrestaurant ?? true,
      isfortakeaway = isfortakeaway ?? true,
      isfordelivery = isfordelivery ?? true,
      isforcustomer = isforcustomer ?? true,
      isforcustomerpreorder = isforcustomerpreorder ?? true;

  factory ProductBarcodeRestaurantModel.fromJson(Map<String, dynamic> json) => _$ProductBarcodeRestaurantModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductBarcodeRestaurantModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class FiexdCostModel {
  /// ต้นทุน
  double? amount;

  /// วันที่
  String? effectdate;

  FiexdCostModel({double? amount, String? effectdate}) : amount = amount ?? 0, effectdate = effectdate ?? '';

  factory FiexdCostModel.fromJson(Map<String, dynamic> json) => _$FiexdCostModelFromJson(json);

  Map<String, dynamic> toJson() => _$FiexdCostModelToJson(this);

  static FiexdCostModel from(FiexdCostModel fiexdCost) {
    // Add your logic here
    return fiexdCost;
  }
}
