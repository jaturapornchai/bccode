import 'package:smlaicloud/model/accountchart_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'doc_format_model.g.dart';

@JsonSerializable()
class DocFormatModel {
  late String? guidfixed;
  late String dateformate;
  late List<DetailModel> details;
  late String doccode;
  late int docnumber;
  late String module;
  late String? description;

  late String docformat;
  late bool isautoformat;
  late int yeartype;
  String? accountgroup;
  String? bookcode;

  /// use gl prosess screen
  bool? ischeck;
  int? totaldocument;
  double? successdocument;

  DocFormatModel({
    String? guidfixed,
    required this.dateformate,
    required this.details,
    required this.doccode,
    required this.docnumber,
    required this.module,
    String? description,
    required this.docformat,
    required this.isautoformat,
    required this.yeartype,
    String? accountgroup,
    String? bookcode,
    bool? ischeck,
    int? totaldocument,
    double? successdocument,
  })  : guidfixed = guidfixed ?? '',
        description = description ?? '',
        accountgroup = accountgroup ?? '',
        bookcode = bookcode ?? '',
        ischeck = ischeck ?? false,
        totaldocument = totaldocument ?? 0,
        successdocument = successdocument ?? 0;

  factory DocFormatModel.fromJson(Map<String, dynamic> json) =>
      _$DocFormatModelFromJson(json);
  Map<String, dynamic> toJson() => _$DocFormatModelToJson(this);
}

@JsonSerializable()
class DetailModel {
  late String actioncode;
  late String credit;
  late String debit;
  late String detail;
  late bool isentryselfaccount;
  AccountChartModel? accountcredit;
  AccountChartModel? accountdebit;

  DetailModel({
    required this.actioncode,
    required this.credit,
    required this.debit,
    required this.detail,
    required this.isentryselfaccount,
    AccountChartModel? accountcredit,
    AccountChartModel? accountdebit,
  })  : accountcredit = accountcredit ?? AccountChartModel(),
        accountdebit = accountdebit ?? AccountChartModel();

  factory DetailModel.fromJson(Map<String, dynamic> json) =>
      _$DetailModelFromJson(json);
  Map<String, dynamic> toJson() => _$DetailModelToJson(this);
}

@JsonSerializable()
class DefaultDocFormatModel {
  late String dateformate;
  late String doccode;
  late int docnumber;
  late String name;

  DefaultDocFormatModel({
    required this.dateformate,
    required this.doccode,
    required this.docnumber,
    required this.name,
  });

  factory DefaultDocFormatModel.fromJson(Map<String, dynamic> json) =>
      _$DefaultDocFormatModelFromJson(json);
  Map<String, dynamic> toJson() => _$DefaultDocFormatModelToJson(this);
}
