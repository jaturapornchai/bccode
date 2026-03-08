import 'package:json_annotation/json_annotation.dart';

part 'timezones_model.g.dart';

@JsonSerializable()
class TimezonesModel {
  String value;
  String abbr;
  String offset;
  @JsonKey(name: 'isdst')
  bool isDst;
  String text;
  List<String> utc;

  TimezonesModel({
    required this.value,
    required this.abbr,
    required this.offset,
    required this.isDst,
    required this.text,
    required this.utc,
  });

  factory TimezonesModel.fromJson(Map<String, dynamic> json) =>
      _$TimezonesModelFromJson(json);
  Map<String, dynamic> toJson() => _$TimezonesModelToJson(this);
}
