import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'line_notify_model.g.dart';

@JsonSerializable()
class LineNotifyModel {
  String? guidfixed;
  List<BranchEvent>? branchevents;
  String? name;
  String? token;
  String? type;

  LineNotifyModel({
    String? guidfixed,
    List<BranchEvent>? branchevents,
    String? name,
    String? token,
    String? type,
  })  : guidfixed = guidfixed ?? "",
        branchevents = branchevents ?? [],
        name = name ?? "",
        token = token ?? "",
        type = type ?? "";

  factory LineNotifyModel.fromJson(Map<String, dynamic> json) =>
      _$LineNotifyModelFromJson(json);

  Map<String, dynamic> toJson() => _$LineNotifyModelToJson(this);
}

@JsonSerializable()
class BranchEvent {
  Branch? branch;
  bool? isenable;
  bool? isnearoutofstock;
  bool? isoutofstock;
  bool? issavebill;
  bool? ispreorder;

  BranchEvent({
    Branch? branch,
    bool? isenable,
    bool? isnearoutofstock,
    bool? isoutofstock,
    bool? issavebill,
    bool? ispreorder,
  })  : branch = branch ?? Branch(code: "", guidfixed: "", names: []),
        isenable = isenable ?? true,
        isnearoutofstock = isnearoutofstock ?? true,
        isoutofstock = isoutofstock ?? true,
        issavebill = issavebill ?? true,
        ispreorder = ispreorder ?? true;

  factory BranchEvent.fromJson(Map<String, dynamic> json) =>
      _$BranchEventFromJson(json);

  Map<String, dynamic> toJson() => _$BranchEventToJson(this);
}

@JsonSerializable()
class Branch {
  String? code;
  String? guidfixed;
  List<LanguageDataModel>? names;

  Branch({
    String? code,
    String? guidfixed,
    List<LanguageDataModel>? names,
  })  : code = code ?? "",
        guidfixed = guidfixed ?? "",
        names = names ?? [];

  factory Branch.fromJson(Map<String, dynamic> json) => _$BranchFromJson(json);

  Map<String, dynamic> toJson() => _$BranchToJson(this);
}
