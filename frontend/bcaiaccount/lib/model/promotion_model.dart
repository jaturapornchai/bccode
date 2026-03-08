import 'package:smlaicloud/model/global_model.dart';
// import 'package:smlaicloud/model/product_model.dart'; // Not used
// import 'package:smlaicloud/model/promotion_detail_model.dart'; // Not used
import 'package:json_annotation/json_annotation.dart';

part 'promotion_model.g.dart';

@JsonSerializable()
class PromotionModel {
  // ประเภท Promotion พร้อมตัวอย่าง
  // 1 = แถมสินค้า เมื่อซื้อสินค้าครบตามจำนวน
  //    ตัวอย่าง 1: ซื้อปูนซีเมนต์ครบ 50 ถุง แถมฟรี 5 ถุง
  //    ตัวอย่าง 2: ซื้อสีน้ำภายนอกครบ 3 ถัง แถมฟรี 1 ถัง
  //    ตัวอย่าง 3: ซื้อกระเบื้องปูพื้นครบ 100 แผ่น แถมฟรี 10 แผ่น

  // 2 = ส่วนลดเงินสด หรือเปอร์เซ็นต์ เมื่อซื้อครบตามจำนวน
  //    ตัวอย่าง 1: ซื้อวัสดุก่อสร้างครบ 50,000 บาท รับส่วนลด 15% หรือ 7,500 บาท
  //    ตัวอย่าง 2: ซื้อสุขภัณฑ์ครบ 20,000 บาท รับส่วนลด 2,000 บาท หรือ 10%
  //    ตัวอย่าง 3: ซื้อเหล็กเส้นครบ 100,000 บาท รับส่วนลด 20% หรือ 20,000 บาท

  // 3 = ซื้อสินค้าตาม List แล้วแถมสินค้าตาม List
  //    ตัวอย่าง 1: ซื้อโถสุขภัณฑ์ 1 ชุด แถมฝารองนั่ง 1 อัน
  //    ตัวอย่าง 2: ซื้อสีรองพื้นปูน 1 ถัง แถมลูกกลิ้งทาสี 1 ชุด
  //    ตัวอย่าง 3: ซื้อก๊อกน้ำอ่างล้างหน้า 1 ชุด แถมสายน้ำดี 1 เส้น

  // 4 = ราคาพิเศษ เมื่อซื้อครบตามจำนวน
  //    ตัวอย่าง 1: ซื้อปูนซีเมนต์ 100 ถุง ราคาพิเศษเหลือถุงละ 120 บาท (ปกติ 150 บาท)
  //    ตัวอย่าง 2: ซื้อกระเบื้องมุงหลังคา 1,000 แผ่น ราคาพิเศษเหลือแผ่นละ 45 บาท (ปกติ 55 บาท)
  //    ตัวอย่าง 3: ซื้อเหล็กเส้น 100 เส้น ราคาพิเศษเหลือเส้นละ 180 บาท (ปกติ 220 บาท)

  // 5 = ซื้อครบตามจำนวน ลดเปอร์เซ็นต์หรือจำนวนเงิน
  //    ตัวอย่าง 1: ซื้อวัสดุก่อสร้างครบ 30,000 บาท ลดเพิ่มอีก 8% หรือ 2,400 บาท
  //    ตัวอย่าง 2: ซื้อสุขภัณฑ์ครบ 50,000 บาท ลดเพิ่มอีก 12% หรือ 6,000 บาท
  //    ตัวอย่าง 3: ซื้อวัสดุมุงหลังคาครบ 100,000 บาท ลดเพิ่มอีก 15% หรือ 15,000 บาท

  // 6 = ซื้อครบ xxx บาท แถมสินค้า xx ชิ้น
  //    ตัวอย่าง 1: ซื้อสินค้าครบ 10,000 บาท แถมชุดเครื่องมือช่าง 1 ชุด
  //    ตัวอย่าง 2: ซื้อสินค้าครบ 30,000 บาท แถมปั๊มน้ำอัตโนมัติ 1 เครื่อง
  //    ตัวอย่าง 3: ซื้อสินค้าครบ 50,000 บาท แถมเครื่องฉีดน้ำแรงดันสูง 1 เครื่อง

  // 7 = ซื้อบิลครบ xxx บาท ส่วนลดเพิ่มอีก xxx บาท หรือ x%
  //    ตัวอย่าง 1: ซื้อสินค้าครบ 25,000 บาท รับส่วนลดเพิ่ม 3,000 บาท หรือ 12%
  //    ตัวอย่าง 2: ซื้อสินค้าครบ 40,000 บาท รับส่วนลดเพิ่ม 6,000 บาท หรือ 15%
  //    ตัวอย่าง 3: ซื้อสินค้าครบ 80,000 บาท รับส่วนลดเพิ่ม 12,000 บาท หรือ 18%

  // 8 = ซื้อบิลครบ xxx บาท ได้สินค้ารางวัล (Bonus)
  //    ตัวอย่าง 1: ซื้อสินค้าครบ 20,000 บาท รับฟรี! เครื่องเจียร์ไฟฟ้า มูลค่า 2,590 บาท
  //    ตัวอย่าง 2: ซื้อสินค้าครบ 35,000 บาท รับฟรี! สว่านไร้สาย 18V มูลค่า 4,990 บาท
  //    ตัวอย่าง 3: ซื้อสินค้าครบ 50,000 บาท รับฟรี! เครื่องปั๊มน้ำอัตโนมัติ มูลค่า 8,900 บาท
  String guidfixed;

  int promotiontype;
  // ลำดับ Promotion
  int index;
  // รหัส Promotion
  String code;
  // วันที่เริ่มต้น
  String datebegin;
  // วันที่สิ้นสุด
  String dateend;

  DateTime fromDate;
  // วันที่สิ้นสุด
  DateTime toDate;
  // ชื่อ Promotion
  List<LanguageDataModel> name;
  // สำหรับสมาชิกเท่านั้น
  int customeronly;
  // ข้อความส่วนลด
  String discounttext;
  // รายการ Promotion
  List<PromotionBarcodeIncludeModel> promotionbarcodeinclude;
  // จำนวนที่ต้องซื้อ
  double limitqty;
  // จำนวนที่แถม
  double promotionqty;
  // มูลค่าที่ต้องซื้อ
  double limitamount;

  PromotionModel(
      {String? guidfixed,
      String? code,
      int? promotiontype,
      int? index,
      String? datebegin,
      String? dateend,
      DateTime? fromDate,
      DateTime? toDate,
      List<LanguageDataModel>? name,
      String? discounttext,
      List<PromotionBarcodeIncludeModel>? promotionbarcodeinclude,
      double? limitqty,
      double? promotionqty,
      double? limitamount,
      int? customeronly})
      : guidfixed = guidfixed ?? "",
        limitqty = limitqty ?? 0,
        promotionqty = promotionqty ?? 0,
        limitamount = limitamount ?? 0,
        customeronly = customeronly ?? 0,
        code = code ?? "",
        promotiontype = promotiontype ?? 1,
        index = index ?? 0,
        datebegin = datebegin ?? "",
        dateend = dateend ?? "",
        fromDate = fromDate ?? DateTime.now(),
        toDate = toDate ?? DateTime.now(),
        name = name ?? [],
        discounttext = discounttext ?? "",
        promotionbarcodeinclude = promotionbarcodeinclude ??
            [
              PromotionBarcodeIncludeModel(
                  promotionproduct: [], includeproduct: [])
            ];

  factory PromotionModel.fromJson(Map<String, dynamic> json) =>
      _$PromotionModelFromJson(json);
  Map<String, dynamic> toJson() => _$PromotionModelToJson(this);
}

@JsonSerializable()
class PromotionBarcodeModel {
  String barcode;
  List<LanguageDataModel> name;
  String unitcode;
  List<LanguageDataModel> unitname;
  double qty; // จำนวน (แปลงหน่วยแล้ว)
  double price; // ราคา
  String discounttext;

  PromotionBarcodeModel(
      {String? barcode,
      List<LanguageDataModel>? name,
      String? unitcode,
      List<LanguageDataModel>? unitname,
      double? qty,
      double? price,
      String? discountText})
      : barcode = barcode ?? "",
        name = name ?? [],
        unitcode = unitcode ?? "",
        discounttext = discountText ?? "",
        unitname = unitname ?? [],
        qty = qty ?? 0,
        price = price ?? 0;

  factory PromotionBarcodeModel.fromJson(Map<String, dynamic> json) =>
      _$PromotionBarcodeModelFromJson(json);
  Map<String, dynamic> toJson() => _$PromotionBarcodeModelToJson(this);
}

@JsonSerializable()
class PromotionBarcodeIncludeModel {
  // ซื้อ xxx แถม xxx ชิ้น
  final List<PromotionBarcodeModel> promotionproduct; // เงื่อนไข
  final List<PromotionBarcodeModel> includeproduct; // รายการที่แถม

  PromotionBarcodeIncludeModel(
      {List<PromotionBarcodeModel>? promotionproduct,
      List<PromotionBarcodeModel>? includeproduct})
      : promotionproduct = promotionproduct ?? [],
        includeproduct = includeproduct ?? [];

  factory PromotionBarcodeIncludeModel.fromJson(Map<String, dynamic> json) =>
      _$PromotionBarcodeIncludeModelFromJson(json);
  Map<String, dynamic> toJson() => _$PromotionBarcodeIncludeModelToJson(this);
}
