// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sale_daily_report_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SaleDailyTransaction _$SaleDailyTransactionFromJson(
  Map<String, dynamic> json,
) => SaleDailyTransaction(
  docDate: json['doc_date'] as String,
  docno: json['docno'] as String,
  totalValue: (json['total_value'] as num).toDouble(),
  detailTotalDiscount: (json['detail_total_discount'] as num).toDouble(),
  totalAfterDiscount: (json['total_after_discount'] as num).toDouble(),
  totalExceptVat: (json['total_except_vat'] as num).toDouble(),
  totalBeforeVat: (json['total_before_vat'] as num).toDouble(),
  totalVatValue: (json['total_vat_value'] as num).toDouble(),
  detailTotalAmount: (json['detail_total_amount'] as num).toDouble(),
  totalDiscount: (json['total_discount'] as num).toDouble(),
  totalAmount: (json['total_amount'] as num).toDouble(),
);

Map<String, dynamic> _$SaleDailyTransactionToJson(
  SaleDailyTransaction instance,
) => <String, dynamic>{
  'doc_date': instance.docDate,
  'docno': instance.docno,
  'total_value': instance.totalValue,
  'detail_total_discount': instance.detailTotalDiscount,
  'total_after_discount': instance.totalAfterDiscount,
  'total_except_vat': instance.totalExceptVat,
  'total_before_vat': instance.totalBeforeVat,
  'total_vat_value': instance.totalVatValue,
  'detail_total_amount': instance.detailTotalAmount,
  'total_discount': instance.totalDiscount,
  'total_amount': instance.totalAmount,
};

SaleDailyReportData _$SaleDailyReportDataFromJson(Map<String, dynamic> json) =>
    SaleDailyReportData(
      docDate: json['doc_date'] as String,
      totalValue: (json['total_value'] as num).toDouble(),
      detailTotalDiscount: (json['detail_total_discount'] as num).toDouble(),
      totalAfterDiscount: (json['total_after_discount'] as num).toDouble(),
      totalExceptVat: (json['total_except_vat'] as num).toDouble(),
      totalBeforeVat: (json['total_before_vat'] as num).toDouble(),
      totalVatValue: (json['total_vat_value'] as num).toDouble(),
      detailTotalAmount: (json['detail_total_amount'] as num).toDouble(),
      totalDiscount: (json['total_discount'] as num).toDouble(),
      totalAmount: (json['total_amount'] as num).toDouble(),
      transactions:
          (json['transactions'] as List<dynamic>?)
              ?.map(
                (e) => SaleDailyTransaction.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          const [],
    );

Map<String, dynamic> _$SaleDailyReportDataToJson(
  SaleDailyReportData instance,
) => <String, dynamic>{
  'doc_date': instance.docDate,
  'total_value': instance.totalValue,
  'detail_total_discount': instance.detailTotalDiscount,
  'total_after_discount': instance.totalAfterDiscount,
  'total_except_vat': instance.totalExceptVat,
  'total_before_vat': instance.totalBeforeVat,
  'total_vat_value': instance.totalVatValue,
  'detail_total_amount': instance.detailTotalAmount,
  'total_discount': instance.totalDiscount,
  'total_amount': instance.totalAmount,
  'transactions': ?instance.transactions?.map((e) => e.toJson()).toList(),
};

SaleDailyReportSummaryResponse _$SaleDailyReportSummaryResponseFromJson(
  Map<String, dynamic> json,
) => SaleDailyReportSummaryResponse(
  status: json['status'] as String,
  data: SaleDailyReportSummary.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$SaleDailyReportSummaryResponseToJson(
  SaleDailyReportSummaryResponse instance,
) => <String, dynamic>{
  'status': instance.status,
  'data': instance.data.toJson(),
};
