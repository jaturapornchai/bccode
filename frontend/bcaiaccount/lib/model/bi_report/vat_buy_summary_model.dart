import 'package:json_annotation/json_annotation.dart';

part 'vat_buy_summary_model.g.dart';

/// Model สำหรับสรุปรายงานภาษีซื้อ (VAT Buy Summary)
/// แสดงยอดรวมของภาษีซื้อทั้งหมดในช่วงเวลาที่กำหนด
@JsonSerializable()
class VatBuySummary {
  /// วันที่เริ่มต้น
  final String fromdate;

  /// วันที่สิ้นสุด
  final String todate;

  /// จำนวนรายการทั้งหมด
  @JsonKey(name: 'total_records')
  final int totalRecords;

  /// ยอดรวมมูลค่าสินค้า/บริการก่อนภาษี (ฐานภาษีรวม)
  @JsonKey(name: 'total_totalbeforevat')
  final double totalTotalBeforeVat;

  /// ยอดรวมจำนวนเงินภาษีมูลค่าเพิ่ม (ภาษีรวม)
  @JsonKey(name: 'total_totalvatvalue')
  final double totalTotalVatValue;

  /// ยอดรวมยกเว้นภาษี
  @JsonKey(name: 'total_totalexceptvat')
  final double totalTotalExceptVat;

  /// ยอดรวมทั้งสิ้น
  @JsonKey(name: 'total_totalamount')
  final double totalTotalAmount;

  /// ชื่อบริษัท (จาก storage)
  final String companyName;

  /// ผู้พิมพ์รายงาน (จาก storage)
  final String printedBy;

  const VatBuySummary({
    required this.fromdate,
    required this.todate,
    required this.totalRecords,
    required this.totalTotalBeforeVat,
    required this.totalTotalVatValue,
    required this.totalTotalExceptVat,
    required this.totalTotalAmount,
    required this.companyName,
    required this.printedBy,
  });

  factory VatBuySummary.fromJson(Map<String, dynamic> json) =>
      _$VatBuySummaryFromJson(json);

  Map<String, dynamic> toJson() => _$VatBuySummaryToJson(this);
}
