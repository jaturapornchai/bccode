// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sale_report_summary.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SaleReportSummary _$SaleReportSummaryFromJson(Map<String, dynamic> json) =>
    SaleReportSummary(
      fromdate: json['fromdate'] as String?,
      todate: json['todate'] as String?,
      totalRecords: (json['total_records'] as num?)?.toInt(),
      totalAmount: (json['total_amount'] as num?)?.toDouble(),
      totalValue: (json['total_value'] as num?)?.toDouble(),
      totalAfterDiscount: (json['total_after_discount'] as num?)?.toDouble(),
      totalExceptVat: (json['total_except_vat'] as num?)?.toDouble(),
      totalBeforeVat: (json['total_before_vat'] as num?)?.toDouble(),
      totalVatValue: (json['total_vat_value'] as num?)?.toDouble(),
      totalDetailDiscount: (json['total_detail_discount'] as num?)?.toDouble(),
      totalByBranch: (json['total_by_branch'] as List<dynamic>?)
          ?.map(
            (e) =>
                SaleReportSummaryByBranch.fromJson(e as Map<String, dynamic>),
          )
          .toList(),
    );

Map<String, dynamic> _$SaleReportSummaryToJson(
  SaleReportSummary instance,
) => <String, dynamic>{
  'fromdate': instance.fromdate,
  'todate': instance.todate,
  'total_records': instance.totalRecords,
  'total_amount': instance.totalAmount,
  'total_value': instance.totalValue,
  'total_after_discount': instance.totalAfterDiscount,
  'total_except_vat': instance.totalExceptVat,
  'total_before_vat': instance.totalBeforeVat,
  'total_vat_value': instance.totalVatValue,
  'total_detail_discount': instance.totalDetailDiscount,
  'total_by_branch': instance.totalByBranch?.map((e) => e.toJson()).toList(),
};

SaleReportSummaryByBranch _$SaleReportSummaryByBranchFromJson(
  Map<String, dynamic> json,
) => SaleReportSummaryByBranch(
  branchcode: json['branchcode'] as String?,
  totalRecords: (json['total_records'] as num?)?.toInt(),
  totalAmount: (json['total_amount'] as num?)?.toDouble(),
  totalValue: (json['total_value'] as num?)?.toDouble(),
  totalAfterDiscount: (json['total_after_discount'] as num?)?.toDouble(),
  totalExceptVat: (json['total_except_vat'] as num?)?.toDouble(),
  totalBeforeVat: (json['total_before_vat'] as num?)?.toDouble(),
  totalVatValue: (json['total_vat_value'] as num?)?.toDouble(),
  totalDetailDiscount: (json['total_detail_discount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$SaleReportSummaryByBranchToJson(
  SaleReportSummaryByBranch instance,
) => <String, dynamic>{
  'branchcode': instance.branchcode,
  'total_records': instance.totalRecords,
  'total_amount': instance.totalAmount,
  'total_value': instance.totalValue,
  'total_after_discount': instance.totalAfterDiscount,
  'total_except_vat': instance.totalExceptVat,
  'total_before_vat': instance.totalBeforeVat,
  'total_vat_value': instance.totalVatValue,
  'total_detail_discount': instance.totalDetailDiscount,
};
