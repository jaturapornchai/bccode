import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/select_colums_csv_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'report_list_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ReportListModel {
  String code;
  String group;
  List<LanguageDataModel> names;
  ReportEnum? type;

  ReportListModel({
    required this.code,
    required this.group,
    required this.names,
    this.type,
  });
  factory ReportListModel.fromJson(Map<String, dynamic> json) =>
      _$ReportListModelFromJson(json);
  Map<String, dynamic> toJson() => _$ReportListModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class LogDownloadParthModel {
  String? guidfixed;
  String? jobid;
  String? path;
  String? status;
  FilterrReportModel? filter;
  String? menu;
  int xorder;

  LogDownloadParthModel({
    String? guidfixed,
    String? jobid,
    String? path,
    String? status,
    FilterrReportModel? filter,
    String? menu,
    required this.xorder,
  })  : path = path ?? '',
        status = status ?? '',
        jobid = jobid ?? '',
        filter = filter ?? FilterrReportModel(),
        guidfixed = guidfixed ?? '',
        menu = menu ?? '';
  factory LogDownloadParthModel.fromJson(Map<String, dynamic> json) =>
      _$LogDownloadParthModelFromJson(json);
  Map<String, dynamic> toJson() => _$LogDownloadParthModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class FilterrReportModel {
  ReportEnum? type;
  String? fromdate;
  String? todate;
  int? showdetail;
  int? showsumbydate;
  String? search;
  String? yearnum;
  String? monthnum;
  String? fromcustcode;
  String? tocustcode;
  String? branch;
  int? iscancel;
  int? iscost;
  String? fromsalecode;
  String? tosalecode;
  String? inquirytype;
  String? ispos;
  String? frombarcode;
  String? tobarcode;
  String? fromgroup;
  String? togroup;
  String? barcode;
  String? typefile;
  List<ListColumsCsvModel>? listcolumscsv;

  FilterrReportModel({
    ReportEnum? type,
    String? fromdate,
    String? todate,
    int? showdetail,
    int? showsumbydate,
    String? search,
    String? yearnum,
    String? monthnum,
    String? fromcustcode,
    String? tocustcode,
    String? branch,
    int? iscancel,
    int? iscost,
    String? fromsalecode,
    String? tosalecode,
    String? inquirytype,
    String? ispos,
    String? frombarcode,
    String? tobarcode,
    String? fromgroup,
    String? togroup,
    String? barcode,
    String? typefile,
    List<ListColumsCsvModel>? listcolumscsv,
  })  : fromdate = fromdate ?? '',
        todate = todate ?? '',
        showdetail = showdetail ?? 0,
        showsumbydate = showsumbydate ?? 0,
        search = search ?? '',
        yearnum = yearnum ?? '',
        monthnum = monthnum ?? '',
        fromcustcode = fromcustcode ?? '',
        tocustcode = tocustcode ?? '',
        branch = branch ?? '',
        iscancel = iscancel ?? 0,
        fromsalecode = fromsalecode ?? '',
        tosalecode = tosalecode ?? '',
        inquirytype = inquirytype ?? '',
        ispos = ispos ?? '',
        frombarcode = frombarcode ?? '',
        tobarcode = tobarcode ?? '',
        fromgroup = fromgroup ?? '',
        togroup = togroup ?? '',
        iscost = iscost ?? 0,
        barcode = barcode ?? '',
        typefile = typefile ?? '',
        listcolumscsv = listcolumscsv ?? [];
  factory FilterrReportModel.fromJson(Map<String, dynamic> json) =>
      _$FilterrReportModelFromJson(json);
  Map<String, dynamic> toJson() => _$FilterrReportModelToJson(this);
}
