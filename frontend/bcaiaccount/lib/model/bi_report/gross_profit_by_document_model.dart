import 'package:json_annotation/json_annotation.dart';

part 'gross_profit_by_document_model.g.dart';

/// Model สำหรับข้อมูลรายงานกำไรขั้นต้นตามเอกสาร (Gross Profit By Document)
/// ใช้แสดงข้อมูลกำไรแยกตามเอกสารขาย
@JsonSerializable()
class GrossProfitByDocumentModel {
  /// วันที่เอกสาร
  final String? docdate;

  /// เลขที่เอกสาร
  final String? docno;

  /// รหัสลูกค้า
  final String? creditorcode;

  /// ชื่อลูกค้า (รองรับหลายภาษา)
  final List<CreditorName>? creditornames;

  /// ยอดรวมก่อน VAT
  final double? totalbeforevat;

  /// ส่วนลดรวม
  final double? detailtotaldiscount;

  /// ยอดรวมหลังหักส่วนลด
  final double? totalafterdiscount;

  /// ต้นทุนรวม
  final double? totalcost;

  /// กำไร (ยอดขาย - ต้นทุน)
  final double? pal;

  /// เปอร์เซ็นต์กำไร
  @JsonKey(name: 'per_pal')
  final double? perPal;

  const GrossProfitByDocumentModel({
    this.docdate,
    this.docno,
    this.creditorcode,
    this.creditornames,
    this.totalbeforevat,
    this.detailtotaldiscount,
    this.totalafterdiscount,
    this.totalcost,
    this.pal,
    this.perPal,
  });

  factory GrossProfitByDocumentModel.fromJson(Map<String, dynamic> json) =>
      _$GrossProfitByDocumentModelFromJson(json);

  Map<String, dynamic> toJson() => _$GrossProfitByDocumentModelToJson(this);
}

/// Model สำหรับชื่อลูกค้า (รองรับหลายภาษา)
@JsonSerializable()
class CreditorName {
  /// รหัสภาษา (th, en, etc.)
  final String? code;

  /// ชื่อลูกค้า
  final String? name;

  /// สร้างอัตโนมัติหรือไม่
  final bool? isauto;

  /// ถูกลบหรือไม่
  final bool? isdelete;

  const CreditorName({this.code, this.name, this.isauto, this.isdelete});

  factory CreditorName.fromJson(Map<String, dynamic> json) =>
      _$CreditorNameFromJson(json);

  Map<String, dynamic> toJson() => _$CreditorNameToJson(this);
}
