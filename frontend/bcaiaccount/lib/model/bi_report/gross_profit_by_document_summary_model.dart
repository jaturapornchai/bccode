import 'package:json_annotation/json_annotation.dart';

part 'gross_profit_by_document_summary_model.g.dart';

/// Model สำหรับข้อมูลสรุปรายงานกำไรขั้นต้นตามเอกสาร
/// ใช้แสดงยอดรวมของข้อมูลกำไรทั้งหมด
@JsonSerializable()
class GrossProfitByDocumentSummary {
  /// วันที่เริ่มต้น
  final String? fromdate;

  /// วันที่สิ้นสุด
  final String? todate;

  /// จำนวนเอกสารทั้งหมด
  @JsonKey(name: 'total_records')
  final int? totalRecords;

  /// ยอดรวมก่อน VAT ทั้งหมด
  @JsonKey(name: 'total_totalbeforevat')
  final double? totalTotalBeforeVat;

  /// ส่วนลดรวมทั้งหมด
  @JsonKey(name: 'total_detailtotaldiscount')
  final double? totalDetailTotalDiscount;

  /// ยอดรวมหลังหักส่วนลดทั้งหมด
  @JsonKey(name: 'total_totalafterdiscount')
  final double? totalTotalAfterDiscount;

  /// ต้นทุนรวมทั้งหมด
  @JsonKey(name: 'total_totalcost')
  final double? totalTotalCost;

  /// กำไรรวมทั้งหมด (PAL)
  @JsonKey(name: 'total_pal')
  final double? totalPal;

  /// ชื่อบริษัท
  final String? companyName;

  const GrossProfitByDocumentSummary({
    this.fromdate,
    this.todate,
    this.totalRecords,
    this.totalTotalBeforeVat,
    this.totalDetailTotalDiscount,
    this.totalTotalAfterDiscount,
    this.totalTotalCost,
    this.totalPal,
    this.companyName,
  });

  factory GrossProfitByDocumentSummary.fromJson(Map<String, dynamic> json) =>
      _$GrossProfitByDocumentSummaryFromJson(json);

  Map<String, dynamic> toJson() => _$GrossProfitByDocumentSummaryToJson(this);

  /// คำนวณเปอร์เซ็นต์กำไรเฉลี่ย
  double get averageProfitPercentage {
    if (totalTotalAfterDiscount == null || totalTotalAfterDiscount == 0) {
      return 0;
    }
    if (totalPal == null) return 0;
    return (totalPal! / totalTotalAfterDiscount!) * 100;
  }
}
