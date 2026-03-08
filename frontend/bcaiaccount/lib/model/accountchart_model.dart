import 'package:json_annotation/json_annotation.dart';

part 'accountchart_model.g.dart';

@JsonSerializable(explicitToJson: true)
class AccountChartModel {
  String? shopid;
  String? accountcode;
  String? guidfixed;
  String? accountname;
  int? accountcategory;
  String? accountgroup;
  int? accountlevel;
  String? consolidateaccountcode;
  int? accountbalancetype;
  bool? iscenterchart;

  AccountChartModel({
    String? shopid,
    String? accountcode,
    String? guidfixed,
    String? accountname,
    int? accountcategory,
    String? accountgroup,
    int? accountlevel,
    String? consolidateaccountcode,
    int? accountbalancetype,
    bool? iscenterchart,
  })  : shopid = shopid ?? '',
        accountcode = accountcode ?? '',
        guidfixed = guidfixed ?? '',
        accountname = accountname ?? '',
        accountcategory = accountcategory ?? 0,
        accountgroup = accountgroup ?? '',
        accountlevel = accountlevel ?? 0,
        consolidateaccountcode = consolidateaccountcode ?? '',
        accountbalancetype = accountbalancetype ?? 0,
        iscenterchart = iscenterchart ?? false;

  factory AccountChartModel.fromJson(Map<String, dynamic> json) =>
      _$AccountChartModelFromJson(json);

  Map<String, dynamic> toJson() => _$AccountChartModelToJson(this);
}
