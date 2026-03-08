import 'package:json_annotation/json_annotation.dart';

part 'token_login_model.g.dart';

@JsonSerializable()
class TokenLoginModel {
  String token;

  TokenLoginModel({required this.token});

  factory TokenLoginModel.fromJson(Map<String, dynamic> json) =>
      _$TokenLoginModelFromJson(json);

  Map<String, dynamic> toJson() => _$TokenLoginModelToJson(this);
}
