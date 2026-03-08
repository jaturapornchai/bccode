import 'package:json_annotation/json_annotation.dart';
import '../global.dart' as global;

part 'wht_model.g.dart';

/// WHT Entry Model - รายการภาษีหัก ณ ที่จ่าย
/// เอกสาร 1 ใบอาจมีหลายอัตราภาษีหัก ณ ที่จ่าย
@JsonSerializable(explicitToJson: true)
class WHTEntryModel {
  /// ประเภทเงินได้ (เลือกจาก radio button)
  String description;

  /// อัตราภาษี % (0.5, 1, 2, 3, 5, 10, 15)
  double rate;

  /// ยอดที่คำนวณได้ (ฐานภาษี)
  double taxbase;

  /// จำนวนเงินหัก ณ ที่จ่าย (taxbase × rate%)
  double amount;

  /// คำอธิบายเพิ่มเติม / หมายเหตุ
  String? note;

  WHTEntryModel({
    String? description,
    double? rate,
    double? taxbase,
    double? amount,
    this.note,
  })  : description = description ?? "",
        rate = rate ?? 0.0,
        taxbase = taxbase ?? 0.0,
        amount = amount ?? 0.0;

  factory WHTEntryModel.fromJson(Map<String, dynamic> json) =>
      _$WHTEntryModelFromJson(json);

  Map<String, dynamic> toJson() => _$WHTEntryModelToJson(this);
}

/// ประเภทเงินได้ (สำหรับ WHT) - ใช้แสดงใน radio button
class WHTIncomeType {
  final String code;
  final String nameTH;
  final String nameEN;
  final double defaultRate; // อัตราภาษีที่แนะนำ

  const WHTIncomeType({
    required this.code,
    required this.nameTH,
    required this.nameEN,
    required this.defaultRate,
  });

  static final List<WHTIncomeType> values = [
    WHTIncomeType(
      code: 'transport',
      nameTH: 'ค่าขนส่ง (จดทะเบียน)',
      nameEN: 'Transport (Registered)',
      defaultRate: 1.0,
    ),
    WHTIncomeType(
      code: 'insurance',
      nameTH: 'ค่าเบี้ยประกันภัย',
      nameEN: 'Insurance Premium',
      defaultRate: 1.0,
    ),
    WHTIncomeType(
      code: 'advertising',
      nameTH: 'ค่าโฆษณา',
      nameEN: 'Advertising',
      defaultRate: 2.0,
    ),
    WHTIncomeType(
      code: 'service',
      nameTH: 'ค่าบริการ',
      nameEN: 'Service',
      defaultRate: 3.0,
    ),
    WHTIncomeType(
      code: 'construction',
      nameTH: 'ค่าจ้างก่อสร้าง / พิมพ์สิ่งพิมพ์',
      nameEN: 'Construction / Printing',
      defaultRate: 3.0,
    ),
    WHTIncomeType(
      code: 'royalty',
      nameTH: 'ค่าลิขสิทธิ์',
      nameEN: 'Royalty',
      defaultRate: 3.0,
    ),
    WHTIncomeType(
      code: 'rental',
      nameTH: 'ค่าเช่า',
      nameEN: 'Rental',
      defaultRate: 5.0,
    ),
    WHTIncomeType(
      code: 'dividend',
      nameTH: global.language('isdividend'),
      nameEN: 'Dividend',
      defaultRate: 10.0,
    ),
  ];

  static WHTIncomeType? findByCode(String code) {
    try {
      return values.firstWhere((type) => type.code == code);
    } catch (e) {
      return null;
    }
  }
}
