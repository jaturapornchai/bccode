// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'bi_sale_report_data.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SaleReportData _$SaleReportDataFromJson(Map<String, dynamic> json) =>
    SaleReportData(
      docdate: json['docdate'] as String,
      docTime: json['doc_time'] as String,
      docno: json['docno'] as String,
      creditorcode: json['creditorcode'] as String,
      creditornames:
          (json['creditornames'] as List<dynamic>?)
              ?.map((e) => SaleCreditorName.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      totalvalue: const DoubleConverter().fromJson(json['totalvalue']),
      detailtotaldiscount: const DoubleConverter().fromJson(
        json['detailtotaldiscount'],
      ),
      totalafterdiscount: const DoubleConverter().fromJson(
        json['totalafterdiscount'],
      ),
      totalexceptvat: const DoubleConverter().fromJson(json['totalexceptvat']),
      totalbeforevat: const DoubleConverter().fromJson(json['totalbeforevat']),
      totalvatvalue: const DoubleConverter().fromJson(json['totalvatvalue']),
      detailtotalamount: const DoubleConverter().fromJson(
        json['detailtotalamount'],
      ),
      totaldiscount: const DoubleConverter().fromJson(json['totaldiscount']),
      totalamount: const DoubleConverter().fromJson(json['totalamount']),
      salecode: json['salecode'] as String? ?? '',
      salename: json['salename'] as String? ?? '',
      inquirytype: json['inquirytype'] as String? ?? '',
      iscancel: json['iscancel'] as bool,
      ispos: json['ispos'] as bool,
      branchcode: json['branchcode'] as String,
      branchnames:
          (json['branchnames'] as List<dynamic>?)
              ?.map((e) => BranchName.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      transactions:
          (json['transactions'] as List<dynamic>?)
              ?.map((e) => SaleTransaction.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );

Map<String, dynamic> _$SaleReportDataToJson(SaleReportData instance) =>
    <String, dynamic>{
      'docdate': instance.docdate,
      'doc_time': instance.docTime,
      'docno': instance.docno,
      'creditorcode': instance.creditorcode,
      'creditornames': instance.creditornames,
      'totalvalue': const DoubleConverter().toJson(instance.totalvalue),
      'detailtotaldiscount': const DoubleConverter().toJson(
        instance.detailtotaldiscount,
      ),
      'totalafterdiscount': const DoubleConverter().toJson(
        instance.totalafterdiscount,
      ),
      'totalexceptvat': const DoubleConverter().toJson(instance.totalexceptvat),
      'totalbeforevat': const DoubleConverter().toJson(instance.totalbeforevat),
      'totalvatvalue': const DoubleConverter().toJson(instance.totalvatvalue),
      'detailtotalamount': const DoubleConverter().toJson(
        instance.detailtotalamount,
      ),
      'totaldiscount': const DoubleConverter().toJson(instance.totaldiscount),
      'totalamount': const DoubleConverter().toJson(instance.totalamount),
      'salecode': instance.salecode,
      'salename': instance.salename,
      'inquirytype': instance.inquirytype,
      'iscancel': instance.iscancel,
      'ispos': instance.ispos,
      'branchcode': instance.branchcode,
      'branchnames': instance.branchnames,
      'transactions': instance.transactions,
    };

SaleCreditorName _$SaleCreditorNameFromJson(Map<String, dynamic> json) =>
    SaleCreditorName(
      code: json['code'] as String,
      name: json['name'] as String,
      isauto: json['isauto'] as bool,
      isdelete: json['isdelete'] as bool,
    );

Map<String, dynamic> _$SaleCreditorNameToJson(SaleCreditorName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };

BranchName _$BranchNameFromJson(Map<String, dynamic> json) => BranchName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$BranchNameToJson(BranchName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };

SaleTransaction _$SaleTransactionFromJson(Map<String, dynamic> json) =>
    SaleTransaction(
      shopid: json['shopid'] as String,
      docdate: json['docdate'] as String,
      docno: json['docno'] as String,
      linenumber: const IntConverter().fromJson(json['linenumber']),
      barcode: json['barcode'] as String? ?? '',
      itemnames:
          (json['itemnames'] as List<dynamic>?)
              ?.map((e) => ItemName.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      unitnames:
          (json['unitnames'] as List<dynamic>?)
              ?.map((e) => UnitName.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      whnames:
          (json['whnames'] as List<dynamic>?)
              ?.map((e) => WarehouseName.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      locationnames:
          (json['locationnames'] as List<dynamic>?)
              ?.map((e) => LocationName.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      qty: const DoubleConverter().fromJson(json['qty']),
      price: const DoubleConverter().fromJson(json['price']),
      discount: const DoubleConverter().fromJson(json['discount']),
      sumamount: const DoubleConverter().fromJson(json['sumamount']),
    );

Map<String, dynamic> _$SaleTransactionToJson(SaleTransaction instance) =>
    <String, dynamic>{
      'shopid': instance.shopid,
      'docdate': instance.docdate,
      'docno': instance.docno,
      'linenumber': const IntConverter().toJson(instance.linenumber),
      'barcode': instance.barcode,
      'itemnames': instance.itemnames,
      'unitnames': instance.unitnames,
      'whnames': instance.whnames,
      'locationnames': instance.locationnames,
      'qty': const DoubleConverter().toJson(instance.qty),
      'price': const DoubleConverter().toJson(instance.price),
      'discount': const DoubleConverter().toJson(instance.discount),
      'sumamount': const DoubleConverter().toJson(instance.sumamount),
    };

ItemName _$ItemNameFromJson(Map<String, dynamic> json) => ItemName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$ItemNameToJson(ItemName instance) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'isauto': instance.isauto,
  'isdelete': instance.isdelete,
};

UnitName _$UnitNameFromJson(Map<String, dynamic> json) => UnitName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$UnitNameToJson(UnitName instance) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'isauto': instance.isauto,
  'isdelete': instance.isdelete,
};

WarehouseName _$WarehouseNameFromJson(Map<String, dynamic> json) =>
    WarehouseName(
      code: json['code'] as String,
      name: json['name'] as String,
      isauto: json['isauto'] as bool,
      isdelete: json['isdelete'] as bool,
    );

Map<String, dynamic> _$WarehouseNameToJson(WarehouseName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };

LocationName _$LocationNameFromJson(Map<String, dynamic> json) => LocationName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$LocationNameToJson(LocationName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };
