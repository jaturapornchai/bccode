part of 'report_bloc.dart';

abstract class ReportEvent extends Equatable {
  const ReportEvent();

  @override
  List<Object> get props => [];
}

class GetReport extends ReportEvent {
  final ReportEnum type;
  final String fromdate;
  final String todate;
  final int page;
  final int perpage;

  const GetReport(
      {required this.type,
      required this.fromdate,
      required this.todate,
      required this.page,
      required this.perpage});

  @override
  List<Object> get props => [type, fromdate, todate, page, perpage];
}

class DownloadReport extends ReportEvent {
  final String guid;
  final ReportEnum type;
  final String fromdate;
  final String todate;
  final int showDetail;
  final int showSumByDate;
  final String search;
  final String yearnum;
  final String monthnum;
  final String fromcustcode;
  final String tocustcode;
  final String branch;
  final int iscancel;
  final int iscost;
  final String fromsalecode;
  final String tosalecode;
  final String inquirytype;
  final String ispos;
  final String barcode;
  final String frombarcode;
  final String tobarcode;
  final String fromgroup;
  final String togroup;
  final int xorder;
  final String typefile;
  final List<ListColumsCsvModel> listcolumscsv;

  DownloadReport({
    String? guid,
    required this.type,
    required this.fromdate,
    required this.todate,
    int? showDetail,
    int? showSumByDate,
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
    String? barcode,
    String? frombarcode,
    String? tobarcode,
    String? fromgroup,
    String? togroup,
    required this.xorder,
    String? typefile,
    List<ListColumsCsvModel>? listcolumscsv,
  })  : guid = guid ?? '',
        showDetail = showDetail ?? 0,
        showSumByDate = showSumByDate ?? 0,
        search = search ?? '',
        yearnum = yearnum ?? '',
        monthnum = monthnum ?? '',
        fromcustcode = fromcustcode ?? '',
        tocustcode = tocustcode ?? '',
        branch = branch ?? '',
        iscancel = iscancel ?? 0,
        iscost = iscost ?? 0,
        fromsalecode = fromsalecode ?? '',
        tosalecode = tosalecode ?? '',
        inquirytype = inquirytype ?? '',
        ispos = ispos ?? '',
        barcode = barcode ?? '',
        frombarcode = frombarcode ?? '',
        tobarcode = tobarcode ?? '',
        fromgroup = fromgroup ?? '',
        togroup = togroup ?? '',
        typefile = typefile ?? '',
        listcolumscsv = listcolumscsv ?? [];

  @override
  List<Object> get props =>
      [guid, type, fromdate, todate, xorder, listcolumscsv];
}

class FileStatusGetList extends ReportEvent {
  final int limit;
  final int offset;
  final String menu;

  const FileStatusGetList(
      {required this.offset, required this.limit, required this.menu});

  @override
  List<Object> get props => [];
}

class FileStatusDeleteById extends ReportEvent {
  final String guid;

  const FileStatusDeleteById({required this.guid});

  @override
  List<Object> get props => [guid];
}

class FileStatusDeleteByMenu extends ReportEvent {
  final String menu;

  const FileStatusDeleteByMenu({required this.menu});

  @override
  List<Object> get props => [menu];
}

class FileStatusSave extends ReportEvent {
  final LogDownloadParthModel logDownloadParthModel;

  const FileStatusSave({required this.logDownloadParthModel});

  @override
  List<Object> get props => [logDownloadParthModel];
}

class FileStatusUpdate extends ReportEvent {
  final LogDownloadParthModel logDownloadParthModel;
  final String guid;

  const FileStatusUpdate(
      {required this.logDownloadParthModel, required this.guid});

  @override
  List<Object> get props => [logDownloadParthModel, guid];
}
