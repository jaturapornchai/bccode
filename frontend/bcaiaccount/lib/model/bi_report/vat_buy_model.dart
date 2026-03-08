import 'package:json_annotation/json_annotation.dart';

part 'vat_buy_model.g.dart';

/// Model สำหรับรายงานภาษีซื้อ (VAT Buy Report)
/// แสดงรายละเอียดการซื้อสินค้า/บริการที่มีภาษีมูลค่าเพิ่ม
@JsonSerializable()
class VatBuyModel {
  /// วันที่ใบกำกับภาษี
  final String? taxdocdate;

  /// เลขที่ใบกำกับภาษี
  final String? taxdocno;

  /// วันที่เอกสาร
  final String? docdate;

  /// เลขที่เอกสาร
  final String? docno;

  /// รหัสเจ้าหนี้ (ผู้ขาย/ให้บริการ)
  final String? creditorcode;

  /// ชื่อเจ้าหนี้ (รองรับหลายภาษา)
  final List<CreditorName>? creditorname;

  /// รหัสสาขา
  final String? branchcode;

  /// เลขประจำตัวผู้เสียภาษี
  final String? taxid;

  /// มูลค่าสินค้า/บริการก่อนภาษี (ฐานภาษี)
  final double? totalbeforevat;

  /// จำนวนเงินภาษีมูลค่าเพิ่ม
  final double? totalvatvalue;

  /// ยอดยกเว้นภาษี
  final double? totalexceptvat;

  /// รวมหลังหักภาษี
  final double? totalaftervat;

  /// รวมทั้งสิ้น
  final double? totalamount;

  VatBuyModel({
    this.taxdocdate,
    this.taxdocno,
    this.docdate,
    this.docno,
    this.creditorcode,
    this.creditorname,
    this.branchcode,
    this.taxid,
    this.totalbeforevat,
    this.totalvatvalue,
    this.totalexceptvat,
    this.totalaftervat,
    this.totalamount,
  });

  factory VatBuyModel.fromJson(Map<String, dynamic> json) =>
      _$VatBuyModelFromJson(json);

  Map<String, dynamic> toJson() => _$VatBuyModelToJson(this);
}

/// Model สำหรับชื่อเจ้าหนี้ (รองรับหลายภาษา)
@JsonSerializable()
class CreditorName {
  /// รหัสภาษา (th, en, etc.)
  final String? code;

  /// ชื่อเจ้าหนี้
  final String? name;

  /// สร้างอัตโนมัติ
  final bool? isauto;

  /// ถูกลบแล้ว
  final bool? isdelete;

  CreditorName({this.code, this.name, this.isauto, this.isdelete});

  factory CreditorName.fromJson(Map<String, dynamic> json) =>
      _$CreditorNameFromJson(json);

  Map<String, dynamic> toJson() => _$CreditorNameToJson(this);
}
