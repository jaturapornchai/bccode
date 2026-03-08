import 'package:json_annotation/json_annotation.dart';

part 'gross_profit_by_product_summary_model.g.dart';

/// Model for Gross Profit By Product summary data
/// Contains aggregate statistics for the report period
@JsonSerializable()
class GrossProfitByProductSummary {
  /// วันที่เริ่มต้น
  final String? fromdate;

  /// วันที่สิ้นสุด
  final String? todate;

  /// จำนวนรายการทั้งหมด
  @JsonKey(name: 'total_records')
  final int? totalRecords;

  /// จำนวนสินค้าทั้งหมด
  @JsonKey(name: 'total_qty')
  final double? totalQty;

  /// ยอดรวมไม่รวม VAT ทั้งหมด
  @JsonKey(name: 'total_sumamountexcludevat')
  final double? totalSumAmountExcludeVat;

  /// ต้นทุนรวมทั้งหมด
  @JsonKey(name: 'total_totalcost')
  final double? totalTotalCost;

  /// กำไรทั้งหมด
  @JsonKey(name: 'total_pal')
  final double? totalPal;

  /// ชื่อบริษัท
  final String? companyName;

  const GrossProfitByProductSummary({
    this.fromdate,
    this.todate,
    this.totalRecords,
    this.totalQty,
    this.totalSumAmountExcludeVat,
    this.totalTotalCost,
    this.totalPal,
    this.companyName,
  });

  /// คำนวณเปอร์เซ็นต์กำไรเฉลี่ย
  /// สูตร: (กำไรรวม / ยอดขายหลังหักส่วนลด) * 100
  double get averageProfitPercentage {
    if (totalSumAmountExcludeVat == null || totalSumAmountExcludeVat == 0) {
      return 0;
    }
    if (totalPal == null) return 0;
    return (totalPal! / totalSumAmountExcludeVat!) * 100;
  }

  factory GrossProfitByProductSummary.fromJson(Map<String, dynamic> json) =>
      _$GrossProfitByProductSummaryFromJson(json);

  Map<String, dynamic> toJson() => _$GrossProfitByProductSummaryToJson(this);
}
