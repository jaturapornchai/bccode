// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sale_return_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SaleReturnModel _$SaleReturnModelFromJson(Map<String, dynamic> json) =>
    SaleReturnModel(
      docDate: json['doc_date'] as String,
      docno: json['docno'] as String,
      creditorNames: (json['creditornames'] as List<dynamic>)
          .map((e) => NameModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      totalValue: (json['totalvalue'] as num).toDouble(),
      detailTotalDiscount: (json['detailtotaldiscount'] as num).toDouble(),
      totalAfterDiscount: (json['totalafterdiscount'] as num).toDouble(),
      totalExceptVat: (json['totalexceptvat'] as num).toDouble(),
      totalBeforeVat: (json['totalbeforevat'] as num).toDouble(),
      totalVatValue: (json['totalvatvalue'] as num).toDouble(),
      totalAmount: (json['totalamount'] as num).toDouble(),
      transactions:
          (json['transactions'] as List<dynamic>?)
              ?.map(
                (e) => SaleReturnTransactionModel.fromJson(
                  e as Map<String, dynamic>,
                ),
              )
              .toList() ??
          const [],
    );

Map<String, dynamic> _$SaleReturnModelToJson(SaleReturnModel instance) =>
    <String, dynamic>{
      'doc_date': instance.docDate,
      'docno': instance.docno,
      'creditornames': instance.creditorNames.map((e) => e.toJson()).toList(),
      'totalvalue': instance.totalValue,
      'detailtotaldiscount': instance.detailTotalDiscount,
      'totalafterdiscount': instance.totalAfterDiscount,
      'totalexceptvat': instance.totalExceptVat,
      'totalbeforevat': instance.totalBeforeVat,
      'totalvatvalue': instance.totalVatValue,
      'totalamount': instance.totalAmount,
      'transactions': instance.transactions?.map((e) => e.toJson()).toList(),
    };

SaleReturnTransactionModel _$SaleReturnTransactionModelFromJson(
  Map<String, dynamic> json,
) => SaleReturnTransactionModel(
  shopId: json['shopid'] as String,
  docno: json['docno'] as String,
  docRef: json['docref'] as String,
  barcode: json['barcode'] as String,
  itemNames: (json['itemnames'] as List<dynamic>)
      .map((e) => NameModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitNames: (json['unitnames'] as List<dynamic>)
      .map((e) => NameModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  whNames: (json['whnames'] as List<dynamic>)
      .map((e) => NameModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  locationNames: (json['locationnames'] as List<dynamic>)
      .map((e) => NameModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  qty: (json['qty'] as num).toDouble(),
  price: (json['price'] as num).toDouble(),
  discount: (json['discount'] as num).toDouble(),
  sumAmount: (json['sumamount'] as num).toDouble(),
);

Map<String, dynamic> _$SaleReturnTransactionModelToJson(
  SaleReturnTransactionModel instance,
) => <String, dynamic>{
  'shopid': instance.shopId,
  'docno': instance.docno,
  'docref': instance.docRef,
  'barcode': instance.barcode,
  'itemnames': instance.itemNames.map((e) => e.toJson()).toList(),
  'unitnames': instance.unitNames.map((e) => e.toJson()).toList(),
  'whnames': instance.whNames.map((e) => e.toJson()).toList(),
  'locationnames': instance.locationNames.map((e) => e.toJson()).toList(),
  'qty': instance.qty,
  'price': instance.price,
  'discount': instance.discount,
  'sumamount': instance.sumAmount,
};

NameModel _$NameModelFromJson(Map<String, dynamic> json) => NameModel(
  code: json['code'] as String,
  name: json['name'] as String,
  isAuto: json['isauto'] as bool,
  isDelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$NameModelToJson(NameModel instance) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'isauto': instance.isAuto,
  'isdelete': instance.isDelete,
};

SaleReturnSummaryModel _$SaleReturnSummaryModelFromJson(
  Map<String, dynamic> json,
) => SaleReturnSummaryModel(
  fromDate: json['fromdate'] as String,
  toDate: json['todate'] as String,
  totalRecords: (json['total_records'] as num).toInt(),
  totalAmount: (json['total_amount'] as num).toDouble(),
);

Map<String, dynamic> _$SaleReturnSummaryModelToJson(
  SaleReturnSummaryModel instance,
) => <String, dynamic>{
  'fromdate': instance.fromDate,
  'todate': instance.toDate,
  'total_records': instance.totalRecords,
  'total_amount': instance.totalAmount,
};
