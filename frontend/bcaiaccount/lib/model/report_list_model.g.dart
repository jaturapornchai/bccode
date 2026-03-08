// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'report_list_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReportListModel _$ReportListModelFromJson(Map<String, dynamic> json) =>
    ReportListModel(
      code: json['code'] as String,
      group: json['group'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      type: $enumDecodeNullable(_$ReportEnumEnumMap, json['type']),
    );

Map<String, dynamic> _$ReportListModelToJson(ReportListModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'group': instance.group,
      'names': instance.names.map((e) => e.toJson()).toList(),
      'type': _$ReportEnumEnumMap[instance.type],
    };

const _$ReportEnumEnumMap = {
  ReportEnum.product: 'product',
  ReportEnum.saleinvoice: 'saleinvoice',
  ReportEnum.saleinvoicedetail: 'saleinvoicedetail',
  ReportEnum.debtor: 'debtor',
  ReportEnum.creditor: 'creditor',
  ReportEnum.bookbank: 'bookbank',
  ReportEnum.purchase: 'purchase',
  ReportEnum.purchasereturn: 'purchasereturn',
  ReportEnum.saleinvoicereturn: 'saleinvoicereturn',
  ReportEnum.transfer: 'transfer',
  ReportEnum.receive: 'receive',
  ReportEnum.pickup: 'pickup',
  ReportEnum.returnproduct: 'returnproduct',
  ReportEnum.stockadjustment: 'stockadjustment',
  ReportEnum.paid: 'paid',
  ReportEnum.pay: 'pay',
  ReportEnum.getpaid: 'getpaid',
  ReportEnum.getpay: 'getpay',
  ReportEnum.vatsale: 'vatsale',
  ReportEnum.vatpurchase: 'vatpurchase',
  ReportEnum.salebydebtor: 'salebydebtor',
  ReportEnum.salebydate: 'salebydate',
  ReportEnum.receivemoney: 'receivemoney',
  ReportEnum.salebyproduct: 'salebyproduct',
  ReportEnum.productmovement: 'productmovement',
  ReportEnum.stockbalance: 'stockbalance',
  ReportEnum.stockcard: 'stockcard',
  ReportEnum.csvsaledetail: 'csvsaledetail',
};

LogDownloadParthModel _$LogDownloadParthModelFromJson(
  Map<String, dynamic> json,
) => LogDownloadParthModel(
  guidfixed: json['guidfixed'] as String?,
  jobid: json['jobid'] as String?,
  path: json['path'] as String?,
  status: json['status'] as String?,
  filter: json['filter'] == null
      ? null
      : FilterrReportModel.fromJson(json['filter'] as Map<String, dynamic>),
  menu: json['menu'] as String?,
  xorder: (json['xorder'] as num).toInt(),
);

Map<String, dynamic> _$LogDownloadParthModelToJson(
  LogDownloadParthModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'jobid': instance.jobid,
  'path': instance.path,
  'status': instance.status,
  'filter': instance.filter?.toJson(),
  'menu': instance.menu,
  'xorder': instance.xorder,
};

FilterrReportModel _$FilterrReportModelFromJson(Map<String, dynamic> json) =>
    FilterrReportModel(
      type: $enumDecodeNullable(_$ReportEnumEnumMap, json['type']),
      fromdate: json['fromdate'] as String?,
      todate: json['todate'] as String?,
      showdetail: (json['showdetail'] as num?)?.toInt(),
      showsumbydate: (json['showsumbydate'] as num?)?.toInt(),
      search: json['search'] as String?,
      yearnum: json['yearnum'] as String?,
      monthnum: json['monthnum'] as String?,
      fromcustcode: json['fromcustcode'] as String?,
      tocustcode: json['tocustcode'] as String?,
      branch: json['branch'] as String?,
      iscancel: (json['iscancel'] as num?)?.toInt(),
      iscost: (json['iscost'] as num?)?.toInt(),
      fromsalecode: json['fromsalecode'] as String?,
      tosalecode: json['tosalecode'] as String?,
      inquirytype: json['inquirytype'] as String?,
      ispos: json['ispos'] as String?,
      frombarcode: json['frombarcode'] as String?,
      tobarcode: json['tobarcode'] as String?,
      fromgroup: json['fromgroup'] as String?,
      togroup: json['togroup'] as String?,
      barcode: json['barcode'] as String?,
      typefile: json['typefile'] as String?,
      listcolumscsv: (json['listcolumscsv'] as List<dynamic>?)
          ?.map((e) => ListColumsCsvModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$FilterrReportModelToJson(FilterrReportModel instance) =>
    <String, dynamic>{
      'type': _$ReportEnumEnumMap[instance.type],
      'fromdate': instance.fromdate,
      'todate': instance.todate,
      'showdetail': instance.showdetail,
      'showsumbydate': instance.showsumbydate,
      'search': instance.search,
      'yearnum': instance.yearnum,
      'monthnum': instance.monthnum,
      'fromcustcode': instance.fromcustcode,
      'tocustcode': instance.tocustcode,
      'branch': instance.branch,
      'iscancel': instance.iscancel,
      'iscost': instance.iscost,
      'fromsalecode': instance.fromsalecode,
      'tosalecode': instance.tosalecode,
      'inquirytype': instance.inquirytype,
      'ispos': instance.ispos,
      'frombarcode': instance.frombarcode,
      'tobarcode': instance.tobarcode,
      'fromgroup': instance.fromgroup,
      'togroup': instance.togroup,
      'barcode': instance.barcode,
      'typefile': instance.typefile,
      'listcolumscsv': instance.listcolumscsv?.map((e) => e.toJson()).toList(),
    };
