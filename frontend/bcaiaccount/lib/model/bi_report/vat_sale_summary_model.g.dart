// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vat_sale_summary_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

VatSaleSummary _$VatSaleSummaryFromJson(Map<String, dynamic> json) =>
    VatSaleSummary(
      fromdate: json['fromdate'] as String,
      todate: json['todate'] as String,
      totalRecords: (json['total_records'] as num).toInt(),
      totalTotalBeforeVat: (json['total_totalbeforevat'] as num).toDouble(),
      totalTotalVatValue: (json['total_totalvatvalue'] as num).toDouble(),
      totalTotalExceptVat: (json['total_totalexceptvat'] as num).toDouble(),
      totalTotalAmount: (json['total_totalamount'] as num).toDouble(),
      companyName: json['companyName'] as String,
      printedBy: json['printedBy'] as String,
    );

Map<String, dynamic> _$VatSaleSummaryToJson(VatSaleSummary instance) =>
    <String, dynamic>{
      'fromdate': instance.fromdate,
      'todate': instance.todate,
      'total_records': instance.totalRecords,
      'total_totalbeforevat': instance.totalTotalBeforeVat,
      'total_totalvatvalue': instance.totalTotalVatValue,
      'total_totalexceptvat': instance.totalTotalExceptVat,
      'total_totalamount': instance.totalTotalAmount,
      'companyName': instance.companyName,
      'printedBy': instance.printedBy,
    };
