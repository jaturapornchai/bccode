import 'package:json_annotation/json_annotation.dart';

part 'responses_model.g.dart';

@JsonSerializable()
class ResponsesModel {
  final String id;
  final bool success;

  ResponsesModel({
    required this.id,
    required this.success,
  });

  factory ResponsesModel.fromJson(Map<String, dynamic> json) =>
      _$ResponsesModelFromJson(json);
  Map<String, dynamic> toJson() => _$ResponsesModelToJson(this);
}
