import 'package:json_annotation/json_annotation.dart';

part 'work_day_model.g.dart';

@JsonSerializable()
class WorkDayModel {
  String code;
  String name;
  bool isactive;
  bool fullday;
  List<WorkTimeModel> worktimes;

  WorkDayModel(
      {required this.code,
      required this.name,
      required this.isactive,
      required this.fullday,
      required this.worktimes});

  factory WorkDayModel.fromJson(Map<String, dynamic> json) =>
      _$WorkDayModelFromJson(json);

  Map<String, dynamic> toJson() => _$WorkDayModelToJson(this);
}

@JsonSerializable()
class WorkTimeModel {
  String starttime = "0";
  String endtime = "0";
  String start = "0";
  String end = "0";

  WorkTimeModel({
    required this.starttime,
    required this.endtime,
    this.start = "",
    this.end = "",
  });

  factory WorkTimeModel.fromJson(Map<String, dynamic> json) =>
      _$WorkTimeModelFromJson(json);

  Map<String, dynamic> toJson() => _$WorkTimeModelToJson(this);
}

@JsonSerializable()
class WorkDayListModel {
  List<WorkDayModel> workdays;

  WorkDayListModel({required this.workdays});

  factory WorkDayListModel.fromJson(Map<String, dynamic> json) =>
      _$WorkDayListModelFromJson(json);

  Map<String, dynamic> toJson() => _$WorkDayListModelToJson(this);
}
